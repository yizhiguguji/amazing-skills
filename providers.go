package amazing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func openFile(p string) (*os.File, error) { return os.Open(p) }
func ProviderEntries(skills string) (map[string]M, error) {
	v, e := read(filepath.Join(skills, "skill-router/references/providers.lock.json"))
	if e != nil {
		return nil, e
	}
	out := map[string]M{}
	for _, v := range arr(obj(v)["providers"]) {
		p := obj(v)
		out[str(p["id"])] = p
	}
	return out, nil
}
func Doctor(skills, name string) ([]M, error) {
	entries, e := ProviderEntries(skills)
	if e != nil {
		return nil, e
	}
	names := []string{}
	if name != "" {
		if entries[name] == nil {
			return nil, fmt.Errorf("Unknown provider")
		}
		names = append(names, name)
	} else {
		for n := range entries {
			names = append(names, n)
		}
		names = dedup(names)
	}
	out := []M{}
	for _, n := range names {
		p := entries[n]
		errors := ProviderIntegrity(os.DirFS(skills), "", p)
		status := "passed"
		if len(errors) > 0 {
			status = "failed"
		}
		out = append(out, M{"id": n, "revision": p["revision"], "skill": filepath.Join(skills, str(p["installed_directory"]), "SKILL.md"), "integrity": status, "errors": errors})
	}
	return out, nil
}
func engineLocation(skills, project string) (M, string, string, error) {
	entries, e := ProviderEntries(skills)
	if e != nil {
		return nil, "", "", e
	}
	engine := obj(entries["impeccable"]["engine"])
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x64"
	}
	if runtime.GOOS == "windows" && arch == "arm64" {
		arch = "x64"
	}
	asset := "impeccable-" + runtime.GOOS + "-" + arch
	if runtime.GOOS == "windows" {
		asset += ".exe"
	}
	if !present(obj(engine["digests"])[asset]) {
		return nil, "", "", fmt.Errorf("Unsupported engine platform")
	}
	dest, e := Inside(project, ".product/tool-cache/impeccable/"+str(engine["version"])+"/"+asset)
	return engine, asset, dest, e
}
func appendEvent(root string, v M) error {
	p, e := Inside(root, ".product/provider-events.jsonl")
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		return e
	}
	f, e := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if e != nil {
		return e
	}
	_, e = f.Write(append(canonical(v), '\n'))
	ce := f.Close()
	if e != nil {
		return e
	}
	return ce
}
func SetupEngine(skills, project string) (M, error) {
	engine, asset, dest, e := engineLocation(skills, project)
	if e != nil {
		return nil, e
	}
	expected := str(obj(engine["digests"])[asset])
	if exists(dest) {
		h, e := fileHash(dest)
		if e != nil || h != expected {
			return nil, fmt.Errorf("Cached engine checksum mismatch")
		}
	} else {
		client := http.Client{Timeout: 60 * time.Second}
		url := "https://github.com/pbakaus/impeccable/releases/download/engine-v" + str(engine["version"]) + "/" + asset
		res, e := client.Get(url)
		if e != nil {
			return nil, e
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("Engine download HTTP %d", res.StatusCode)
		}
		data, e := io.ReadAll(io.LimitReader(res.Body, 200*1024*1024))
		if e != nil {
			return nil, e
		}
		if hash(data) != expected {
			return nil, fmt.Errorf("Downloaded engine checksum mismatch")
		}
		if e = writeBytes(dest+".part", data, true); e != nil {
			return nil, e
		}
		defer os.Remove(dest + ".part")
		if e = os.Chmod(dest+".part", 0755); e != nil {
			return nil, e
		}
		if e = os.Rename(dest+".part", dest); e != nil {
			return nil, e
		}
	}
	rel, _ := filepath.Rel(project, dest)
	event := M{"at": now(), "provider": "impeccable", "kind": "engine-setup", "version": engine["version"], "asset": asset, "sha256": expected, "status": "verified", "evidence": filepath.ToSlash(rel)}
	e = appendEvent(project, event)
	return event, e
}
func RunProvider(skills, project, provider string, args []string) (M, string, int, error) {
	items, e := Doctor(skills, provider)
	if e != nil {
		return nil, "", 1, e
	}
	if items[0]["integrity"] != "passed" {
		return nil, "", 1, fmt.Errorf("Provider integrity check failed")
	}
	var output string
	code := 0
	implementation := "pinned-engine"
	if provider == "ui-ux-pro-max" {
		implementation = "amazing-skills-go-promax"
		r, e := NativeSearchCLI(skills, args)
		if e != nil {
			return nil, "", 1, e
		}
		o, _ := parse(args)
		if o.Values["json"] == "true" {
			if o.Values["design-system"] == "true" {
				delete(r, "text")
			}
			output = string(canonical(r)) + "\n"
		} else if o.Values["design-system"] == "true" {
			output = str(r["text"]) + "\n"
			if r["persistence"] != nil {
				output += "Persistence: " + str(obj(r["persistence"])["status"]) + "\n"
			}
		} else {
			output = formatSearchResult(r, o.Values["full"] == "true")
		}
	} else if provider == "impeccable" {
		if len(args) == 0 || !member(args[0], "context", "detect", "engine-probe") {
			return nil, "", 1, fmt.Errorf("Supported engine verbs: context, detect, engine-probe")
		}
		engine, asset, dest, e := engineLocation(skills, project)
		if e != nil {
			return nil, "", 1, e
		}
		h, e := fileHash(dest)
		if e != nil {
			return nil, "", 1, fmt.Errorf("Run providers setup-engine explicitly before invoking the engine")
		}
		if h != str(obj(engine["digests"])[asset]) {
			return nil, "", 1, fmt.Errorf("Cached engine checksum mismatch")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, dest, args...)
		cmd.Dir = project
		cmd.Env = append(os.Environ(), "IMPECCABLE_HOME="+filepath.Join(project, ".product/tool-cache/impeccable"))
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		e = cmd.Run()
		if e != nil {
			if ex, ok := e.(*exec.ExitError); ok {
				code = ex.ExitCode()
			} else {
				return nil, "", 1, e
			}
		}
		output = stdout.String()
		if stderr.Len() > 0 {
			output += "\nSTDERR:\n" + stderr.String()
		}
	} else {
		return nil, "", 1, fmt.Errorf("finesse-brief has no executable; read its Skill and references")
	}
	stamp := time.Now().UTC().Format("20060102T150405.000000000Z")
	rel := ".product/provider-evidence/" + stamp + "-" + provider + ".txt"
	redacted := strings.ReplaceAll(output, project, "<PROJECT_ROOT>")
	redacted = strings.ReplaceAll(redacted, skills, "<INSTALLED_SKILLS_ROOT>")
	if home, err := os.UserHomeDir(); err == nil {
		redacted = strings.ReplaceAll(redacted, home, "<HOME>")
	}
	redacted = "# Path-redacted tool output; local root paths replaced with placeholders.\n" + redacted
	p, e := Inside(project, rel)
	if e != nil {
		return nil, "", 1, e
	}
	if e = writeBytes(p, []byte(redacted), true); e != nil {
		return nil, "", 1, e
	}
	event := M{"at": now(), "provider": provider, "revision": items[0]["revision"], "kind": "tool-execution", "implementation": implementation, "arguments": args, "exit_code": code, "evidence": rel}
	e = appendEvent(project, event)
	return event, redacted, code, e
}
