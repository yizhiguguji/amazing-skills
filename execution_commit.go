package amazing

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const executionPending = ".qa/execution-pending.json"

type executionWriter func(string, []byte) error

// The durable intent authorizes exactly one append to an anchored history prefix.
// It remains until both the journal and its checkpoint have been published.
type executionCommit struct {
	Version        int    `json:"version"`
	PreviousBytes  int64  `json:"previous_bytes"`
	PreviousDigest string `json:"previous_digest"`
	PreviousCount  int    `json:"previous_count"`
	PreviousHash   string `json:"previous_hash"`
	Line           string `json:"line"`
}

func syncExecutionDirectory(directory string) error {
	// Windows does not expose directory fsync through os.File.Sync.
	if runtime.GOOS == "windows" {
		return nil
	}
	f, err := os.Open(directory)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}

func atomicExecutionWrite(name string, data []byte) error {
	mode := os.FileMode(0644)
	if info, err := os.Lstat(name); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("Execution artifact must be a regular file")
		}
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	directory := filepath.Dir(name)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(directory, ".execution-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(mode); err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Rename(f.Name(), name); err != nil {
		return err
	}
	return syncExecutionDirectory(directory)
}

func executionJournalBytes(root string) ([]byte, error) {
	b, err := os.ReadFile(path(root, ".qa/executions.jsonl"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return b, err
}

func parseExecutionLog(data []byte) ([]M, error) {
	events := []M{}
	previous := ""
	ids := map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 4096), 16*1024*1024)
	for scanner.Scan() {
		v, err := decode(scanner.Bytes())
		if err != nil {
			return nil, err
		}
		m := obj(v)
		body := obj(clone(m))
		delete(body, "hash")
		h := str(m["hash"])
		if m["previous_hash"] != previous || hash(canonical(body)) != h {
			return nil, fmt.Errorf("Execution history drift")
		}
		id := str(m["execution_id"])
		if ids[id] {
			return nil, fmt.Errorf("Duplicate execution identity")
		}
		ids[id] = true
		previous = h
		events = append(events, m)
	}
	return events, scanner.Err()
}

func checkpointMatches(checkpoint M, count int, lastHash string) bool {
	return len(checkpoint) == 2 && num(checkpoint["count"]) == count && checkpoint["hash"] == lastHash
}

func appendExecution(root string, events []M, event M, writeAtomic executionWriter) error {
	before, err := executionJournalBytes(root)
	if err != nil {
		return err
	}
	current, err := parseExecutionLog(before)
	if err != nil {
		return err
	}
	previous := ""
	if len(current) > 0 {
		previous = str(current[len(current)-1]["hash"])
	}
	if len(current) != len(events) || previous != event["previous_hash"] {
		return fmt.Errorf("Execution history changed before commit")
	}
	line := string(canonical(event)) + "\n"
	if len(before) > 0 && before[len(before)-1] != '\n' {
		line = "\n" + line
	}
	// Reject an unreadable candidate (including an oversized event) before it
	// becomes a durable intent that subsequent readers must recover.
	if _, err = parseExecutionLog(append(append([]byte{}, before...), []byte(line)...)); err != nil {
		return err
	}
	intent := executionCommit{1, int64(len(before)), hash(before), len(events), previous, line}
	if err = writeAtomic(path(root, executionPending), append(canonical(intent), '\n')); err != nil {
		return err
	}
	return recoverExecutionCommit(root, writeAtomic)
}

func recoverExecutionCommit(root string, writeAtomic executionWriter) error {
	pending := path(root, executionPending)
	data, err := os.ReadFile(pending)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var intent executionCommit
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&intent); err != nil {
		return fmt.Errorf("Invalid execution commit: %w", err)
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF || intent.Version != 1 || intent.PreviousBytes < 0 || intent.PreviousCount < 0 || !strings.HasSuffix(intent.Line, "\n") {
		return fmt.Errorf("Invalid execution commit")
	}
	journal, err := executionJournalBytes(root)
	if err != nil {
		return err
	}
	if intent.PreviousBytes > int64(len(journal)) {
		return fmt.Errorf("Execution commit history prefix missing")
	}
	before := journal[:intent.PreviousBytes]
	if hash(before) != intent.PreviousDigest {
		return fmt.Errorf("Execution commit history prefix changed")
	}
	after := append(append([]byte{}, before...), []byte(intent.Line)...)
	if len(journal) != len(before) && !bytes.Equal(journal, after) {
		return fmt.Errorf("Execution commit journal changed")
	}
	events, err := parseExecutionLog(after)
	if err != nil {
		return err
	}
	if len(events) != intent.PreviousCount+1 || events[len(events)-1]["previous_hash"] != intent.PreviousHash {
		return fmt.Errorf("Execution commit history anchor mismatch")
	}
	expectedLine := string(canonical(events[len(events)-1])) + "\n"
	if len(before) > 0 && before[len(before)-1] != '\n' {
		expectedLine = "\n" + expectedLine
	}
	if intent.Line != expectedLine {
		return fmt.Errorf("Execution commit must contain exactly one canonical event")
	}
	lastHash := str(events[len(events)-1]["hash"])
	checkpointPath := path(root, ".qa/execution-checkpoint.json")
	checkpoint := M{}
	if exists(checkpointPath) {
		v, err := read(checkpointPath)
		if err != nil {
			return err
		}
		checkpoint = obj(v)
		if !checkpointMatches(checkpoint, intent.PreviousCount, intent.PreviousHash) && !checkpointMatches(checkpoint, len(events), lastHash) {
			return fmt.Errorf("Execution commit checkpoint changed")
		}
	} else if intent.PreviousCount != 0 || intent.PreviousHash != "" {
		return fmt.Errorf("Missing execution checkpoint")
	}
	// Never complete an intent whose original evidence has since changed.
	for _, event := range events {
		for _, value := range arr(event["evidence"]) {
			rel := str(value)
			digest, err := fileHash(path(root, rel))
			if err != nil || digest != str(obj(event["evidence_hashes"])[rel]) {
				return fmt.Errorf("Execution evidence drift")
			}
		}
	}
	if !bytes.Equal(journal, after) {
		if err = writeAtomic(path(root, ".qa/executions.jsonl"), after); err != nil {
			return err
		}
	}
	if !checkpointMatches(checkpoint, len(events), lastHash) {
		if err = writeAtomic(checkpointPath, append(canonical(M{"count": len(events), "hash": lastHash}), '\n')); err != nil {
			return err
		}
	}
	if err = os.Remove(pending); err != nil {
		return err
	}
	return syncExecutionDirectory(filepath.Dir(pending))
}
