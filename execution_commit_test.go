package amazing

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRecordCheckpointWriteFailureRemainsRetryable(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("requires Unix file permissions for fault injection")
	}
	root, _ := qaFixture(t)
	need(t, os.WriteFile(path(root, ".qa/executions.jsonl"), nil, 0644))
	directory := path(root, ".qa")
	need(t, os.Chmod(directory, 0555))
	t.Cleanup(func() { _ = os.Chmod(directory, 0755) })
	_, err := Record(root, event("E1", "passed", "run1"))
	if err == nil {
		t.Fatal("expected an unwritable commit/checkpoint directory to fail")
	}
	need(t, os.Chmod(directory, 0755))
	events, err := ReadEvents(root)
	need(t, err)
	if len(events) != 0 {
		t.Fatal("failed commit changed execution history")
	}
	_, err = Record(root, event("E1", "passed", "run1"))
	need(t, err)
	events, err = ReadEvents(root)
	need(t, err)
	if len(events) != 1 {
		t.Fatal("retry did not commit exactly once")
	}
}

func TestExecutionCommitRecoversEveryPublicationBoundary(t *testing.T) {
	for _, artifact := range []string{"execution-pending.json", "executions.jsonl", "execution-checkpoint.json"} {
		for _, published := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/published=%t", artifact, published), func(t *testing.T) {
				root, _ := qaFixture(t)
				_, err := Record(root, event("E0", "failed", "run0"))
				need(t, err)
				before, err := executionJournalBytes(root)
				need(t, err)
				injected := false
				writer := func(name string, data []byte) error {
					if !injected && filepath.Base(name) == artifact {
						injected = true
						if published {
							need(t, atomicExecutionWrite(name, data))
						}
						return fmt.Errorf("injected publication interruption")
					}
					return atomicExecutionWrite(name, data)
				}
				observation := event("E1", "passed", "run1")
				if _, err = recordWithExecutionWriter(root, observation, writer); err == nil || !injected {
					t.Fatal("fault was not exercised", err)
				}
				events, err := ReadEvents(root)
				need(t, err)
				want := 2
				if artifact == "execution-pending.json" && !published {
					want = 1
				}
				if len(events) != want || exists(path(root, executionPending)) {
					t.Fatal("commit did not recover to its precise boundary", len(events))
				}
				after, err := executionJournalBytes(root)
				need(t, err)
				if !bytes.HasPrefix(after, before) {
					t.Fatal("existing history bytes changed")
				}
				result, err := Record(root, observation)
				need(t, err)
				if want == 2 && result["idempotent"] != true {
					t.Fatal("recovery duplicated a committed execution", result)
				}
				events, err = ReadEvents(root)
				need(t, err)
				if len(events) != 2 {
					t.Fatal("retry did not preserve exact execution identity")
				}
				entries, err := os.ReadDir(path(root, ".qa"))
				need(t, err)
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), ".execution-") {
						t.Fatal("temporary commit file leaked")
					}
				}
			})
		}
	}
}

func TestExecutionCommitRejectsChangedHistoryCheckpointEvidenceOrIntent(t *testing.T) {
	for _, damage := range []string{"prefix", "tail", "checkpoint", "evidence", "intent"} {
		t.Run(damage, func(t *testing.T) {
			root, _ := qaFixture(t)
			_, err := Record(root, event("E0", "failed", "run0"))
			need(t, err)
			writer := func(name string, data []byte) error {
				if filepath.Base(name) == "execution-checkpoint.json" {
					return fmt.Errorf("injected checkpoint failure")
				}
				return atomicExecutionWrite(name, data)
			}
			if _, err := recordWithExecutionWriter(root, event("E1", "passed", "run1"), writer); err == nil {
				t.Fatal("fault was not exercised")
			}
			journal := path(root, ".qa/executions.jsonl")
			switch damage {
			case "prefix":
				data, err := os.ReadFile(journal)
				need(t, err)
				data[0] = '['
				need(t, os.WriteFile(journal, data, 0644))
			case "tail":
				data, err := os.ReadFile(journal)
				need(t, err)
				need(t, os.WriteFile(journal, append(data, []byte("unexpected\n")...), 0644))
			case "checkpoint":
				need(t, write(path(root, ".qa/execution-checkpoint.json"), M{"count": 99, "hash": "different"}, false))
			case "evidence":
				need(t, os.WriteFile(path(root, "evidence/observation.txt"), []byte("changed evidence"), 0644))
			case "intent":
				data, err := os.ReadFile(path(root, executionPending))
				need(t, err)
				need(t, os.WriteFile(path(root, executionPending), append(data, []byte("{}")...), 0644))
			}
			names := []string{journal, path(root, executionPending), path(root, ".qa/execution-checkpoint.json")}
			original := map[string][]byte{}
			for _, name := range names {
				original[name], err = os.ReadFile(name)
				need(t, err)
			}
			if _, err := ReadEvents(root); err == nil {
				t.Fatal("changed execution artifacts were silently repaired")
			}
			for name, expected := range original {
				got, err := os.ReadFile(name)
				need(t, err)
				if !bytes.Equal(got, expected) {
					t.Fatal("failed recovery changed original artifacts", filepath.Base(name))
				}
			}
		})
	}
}

func TestFirstExecutionCommitRecoversWithoutExistingCheckpoint(t *testing.T) {
	root, _ := qaFixture(t)
	writer := func(name string, data []byte) error {
		if filepath.Base(name) == "execution-checkpoint.json" {
			return fmt.Errorf("injected first checkpoint failure")
		}
		return atomicExecutionWrite(name, data)
	}
	observation := event("E1", "passed", "run1")
	if _, err := recordWithExecutionWriter(root, observation, writer); err == nil {
		t.Fatal("fault was not exercised")
	}
	events, err := ReadEvents(root)
	need(t, err)
	if len(events) != 1 {
		t.Fatal("first execution was lost")
	}
	result, err := Record(root, observation)
	need(t, err)
	if result["idempotent"] != true {
		t.Fatal("first execution duplicated")
	}
}

func TestExecutionAppendPreservesHistoryWithoutFinalNewline(t *testing.T) {
	root, _ := qaFixture(t)
	_, err := Record(root, event("E0", "failed", "run0"))
	need(t, err)
	before, err := executionJournalBytes(root)
	need(t, err)
	before = bytes.TrimSuffix(before, []byte("\n"))
	need(t, os.WriteFile(path(root, ".qa/executions.jsonl"), before, 0644))
	_, err = Record(root, event("E1", "passed", "run1"))
	need(t, err)
	after, err := executionJournalBytes(root)
	need(t, err)
	if !bytes.HasPrefix(after, append(append([]byte{}, before...), '\n')) {
		t.Fatal("legacy history bytes changed")
	}
	events, err := ReadEvents(root)
	need(t, err)
	if len(events) != 2 {
		t.Fatal("legacy record or new append lost")
	}
}

func TestOversizedExecutionCannotCreateUnrecoverableIntent(t *testing.T) {
	root, _ := qaFixture(t)
	observation := event("large", "passed", "run1")
	observation["actual"] = strings.Repeat("x", 8*1024*1024)
	if _, err := Record(root, observation); err == nil {
		t.Fatal("unreadable oversized event accepted")
	}
	if exists(path(root, executionPending)) {
		t.Fatal("invalid event became a durable intent")
	}
	events, err := ReadEvents(root)
	need(t, err)
	if len(events) != 0 {
		t.Fatal("rejected oversized event changed history")
	}
	_, err = Record(root, event("normal", "passed", "run1"))
	need(t, err)
}
