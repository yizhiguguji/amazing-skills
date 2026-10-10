package amazing

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type M = map[string]any

func obj(v any) M {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return M{}
}
func arr(v any) []any {
	switch a := v.(type) {
	case []any:
		return a
	case []string:
		r := []any{}
		for _, s := range a {
			r = append(r, s)
		}
		return r
	case []M:
		r := []any{}
		for _, s := range a {
			r = append(r, s)
		}
		return r
	}
	return []any{}
}
func str(v any) string { s, _ := v.(string); return s }
func num(v any) int {
	switch x := v.(type) {
	case json.Number:
		n, _ := x.Int64()
		return int(n)
	case int:
		return x
	case float64:
		return int(x)
	}
	return 0
}
func present(v any) bool { return strings.TrimSpace(str(v)) != "" }
func member(v any, choices ...string) bool {
	for _, c := range choices {
		if str(v) == c {
			return true
		}
	}
	return false
}
func has(v any, s string) bool {
	for _, x := range arr(v) {
		if str(x) == s {
			return true
		}
	}
	return false
}
func stringList(v any) bool {
	a, ok := v.([]any)
	if !ok || len(a) == 0 {
		return false
	}
	for _, x := range a {
		if !present(x) {
			return false
		}
	}
	return true
}
func decode(b []byte) (any, error) {
	var v any
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	if e := d.Decode(&v); e != nil {
		return nil, e
	}
	return v, nil
}
func read(p string) (any, error) {
	b, e := os.ReadFile(p)
	if e != nil {
		return nil, e
	}
	return decode(b)
}
func load(f fs.FS, p string) (any, error) {
	b, e := fs.ReadFile(f, p)
	if e != nil {
		return nil, e
	}
	return decode(b)
}
func canonical(v any) []byte {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	_ = e.Encode(v)
	s := strings.TrimSuffix(b.String(), "\n")
	s = strings.ReplaceAll(strings.ReplaceAll(s, `\u2028`, "\u2028"), `\u2029`, "\u2029")
	return []byte(s)
}
func hash(b []byte) string              { v := sha256.Sum256(b); return hex.EncodeToString(v[:]) }
func fileHash(p string) (string, error) { b, e := os.ReadFile(p); return hash(b), e }
func now() string                       { return time.Now().UTC().Format(time.RFC3339Nano) }
func clone(v any) any                   { x, _ := decode(canonical(v)); return x }
func write(p string, v any, exclusive bool) error {
	b := canonical(v)
	var out bytes.Buffer
	_ = json.Indent(&out, b, "", "  ")
	return writeBytes(p, append(out.Bytes(), '\n'), exclusive)
}
func writeBytes(p string, b []byte, exclusive bool) error {
	if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		return e
	}
	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if exclusive {
		flags = os.O_CREATE | os.O_WRONLY | os.O_EXCL
	}
	f, e := os.OpenFile(p, flags, 0644)
	if e != nil {
		return e
	}
	_, e = f.Write(b)
	closeErr := f.Close()
	if e != nil {
		return e
	}
	return closeErr
}
func exists(p string) bool { _, e := os.Stat(p); return e == nil }
func isFile(p string) bool { s, e := os.Stat(p); return e == nil && s.Mode().IsRegular() }
func rootPath(p string) (string, error) {
	if strings.HasPrefix(p, "~/") {
		h, e := os.UserHomeDir()
		if e != nil {
			return "", e
		}
		p = filepath.Join(h, p[2:])
	}
	a, e := filepath.Abs(p)
	if e != nil {
		return "", e
	}
	return resolveMissing(a)
}
func resolveMissing(p string) (string, error) {
	r, e := filepath.EvalSymlinks(p)
	if e == nil {
		return r, nil
	}
	if !errors.Is(e, os.ErrNotExist) {
		return "", e
	}
	if info, statErr := os.Lstat(p); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("Dangling symlink is not a writable artifact: %s", filepath.Base(p))
	}
	parent := filepath.Dir(p)
	if parent == p {
		return "", e
	}
	r, e = resolveMissing(parent)
	return filepath.Join(r, filepath.Base(p)), e
}
func Inside(root, rel string) (string, error) {
	if rel == "" || filepath.IsAbs(rel) || strings.Contains(rel, "\\") || strings.Contains(rel, ":") {
		return "", fmt.Errorf("Path must be project-relative: %s", rel)
	}
	for _, s := range strings.Split(rel, "/") {
		if s == ".." {
			return "", fmt.Errorf("Path must stay inside project: %s", rel)
		}
	}
	root, e := rootPath(root)
	if e != nil {
		return "", e
	}
	p, e := resolveMissing(filepath.Join(root, filepath.FromSlash(rel)))
	if e != nil {
		return "", e
	}
	r, e := filepath.Rel(root, p)
	if e != nil || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("Path escapes project: %s", rel)
	}
	return p, nil
}
func path(root, rel string) string {
	p, e := Inside(root, rel)
	if e != nil {
		panic(e)
	}
	return p
}
func safeCall(fn func() error) (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = fmt.Errorf("%v", r)
		}
	}()
	return fn()
}
func mustRead(p string) any {
	v, e := read(p)
	if e != nil {
		panic(e)
	}
	return v
}
func dedup(a []string) []string {
	m := map[string]bool{}
	for _, s := range a {
		m[s] = true
	}
	out := []string{}
	for s := range m {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
func index(v any, label string, errs *[]string) map[string]M {
	out := map[string]M{}
	a, ok := v.([]any)
	if !ok {
		*errs = append(*errs, "Expected list: "+label)
		return out
	}
	for _, x := range a {
		m := obj(x)
		id := str(m["id"])
		if !present(id) {
			*errs = append(*errs, "Invalid identity: "+label)
			continue
		}
		if _, ok := out[id]; ok {
			*errs = append(*errs, "Duplicate identity: "+id)
		}
		out[id] = m
	}
	return out
}
func withoutStatus(v M) M { o := obj(clone(v)); delete(o, "status"); return o }
