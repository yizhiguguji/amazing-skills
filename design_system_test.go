package amazing

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func designFixture(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "package")
	_, e := DesignSystemInit(root, "calm-paper")
	need(t, e)
	return root
}
func mutateDesignJSON(t *testing.T, root, file string, fn func(M)) {
	t.Helper()
	m := obj(mustRead(filepath.Join(root, file)))
	fn(m)
	need(t, write(filepath.Join(root, file), m, false))
}
func TestDesignSystemLifecycleAndExistingLock(t *testing.T) {
	for _, preset := range []string{"calm-paper", "precise-console"} {
		t.Run(preset, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "synthetic package")
			result, e := DesignSystemInit(root, preset)
			need(t, e)
			if result["reference_extraction"] != "not performed" || result["visual_verified"] != false {
				t.Fatal(result)
			}
			if es := DesignSystemCheck(os.DirFS(root), ""); len(es) > 0 {
				t.Fatal(es)
			}
			project := filepath.Join(t.TempDir(), "project")
			applied, e := DesignSystemApply(root, project)
			need(t, e)
			if applied["implementation_verified"] != false {
				t.Fatal(applied)
			}
			_, e = ProjectInit(project)
			need(t, e)
			if es := DesignSystemCheck(os.DirFS(project), ""); len(es) > 0 {
				t.Fatal(es)
			}
			_, e = LockDesign(project, "tokens.css", false)
			need(t, e)
			if es := CheckDesign(project); len(es) > 0 {
				t.Fatal(es)
			}
			css, e := os.ReadFile(filepath.Join(project, "tokens.css"))
			need(t, e)
			need(t, writeBytes(filepath.Join(project, "tokens.css"), bytes.ReplaceAll(css, []byte("--ds-space: "), []byte("--ds-space: 1")), false))
			if len(CheckDesign(project)) == 0 || len(DesignSystemCheck(os.DirFS(project), "")) == 0 {
				t.Fatal("drift or stale preview accepted")
			}
			state := obj(mustRead(filepath.Join(project, ".product/state.json")))
			if state["phase"] == "complete" || state["status"] == "passed" {
				t.Fatal("package promoted project acceptance")
			}
			if _, e = DesignSystemApply(root, project); e == nil {
				t.Fatal("existing lock overwritten")
			}
		})
	}
}
func TestDesignSystemPreservesExistingArtifacts(t *testing.T) {
	root := designFixture(t)
	if _, e := DesignSystemInit(root, "precise-console"); e == nil {
		t.Fatal("init overwrote package")
	}
	project := t.TempDir()
	need(t, writeBytes(filepath.Join(project, "DESIGN.md"), []byte("Existing visual authority"), true))
	if _, e := DesignSystemApply(root, project); e == nil {
		t.Fatal("existing design overwritten")
	}
	files, e := os.ReadDir(project)
	need(t, e)
	if len(files) != 1 {
		t.Fatal("partial application on preflight failure", files)
	}
	b, e := os.ReadFile(filepath.Join(project, "DESIGN.md"))
	need(t, e)
	if string(b) != "Existing visual authority" {
		t.Fatal("existing design changed")
	}
	if _, e = DesignSystemPreview(root, false); e == nil {
		t.Fatal("preview replaced implicitly")
	}
}
func TestDesignSystemEvidenceIntegrityAndPaths(t *testing.T) {
	root := designFixture(t)
	need(t, writeBytes(filepath.Join(root, "sources/measurement.txt"), []byte("Synthetic measured background: parchment"), true))
	h, e := fileHash(filepath.Join(root, "sources/measurement.txt"))
	need(t, e)
	mutateDesignJSON(t, root, "design-evidence.json", func(m M) {
		m["claims"] = []any{M{"id": "measurement", "kind": "observed", "reference": "synthetic browser fixture", "locator": "page / main / background", "claim": "Background measured in synthetic fixture", "decision": "Use measured background", "token": "--ds-bg", "file": "sources/measurement.txt", "sha256": h}}
	})
	_, e = DesignSystemPreview(root, true)
	need(t, e)
	project := filepath.Join(t.TempDir(), "applied")
	_, e = DesignSystemApply(root, project)
	need(t, e)
	if es := DesignSystemCheck(os.DirFS(project), ""); len(es) > 0 {
		t.Fatal(es)
	}
	need(t, writeBytes(filepath.Join(root, "sources/measurement.txt"), []byte("changed"), false))
	if !strings.Contains(strings.Join(DesignSystemCheck(os.DirFS(root), ""), ";"), "digest mismatch") {
		t.Fatal("evidence tamper accepted")
	}
	for _, rel := range []string{"../outside", "sources/../../outside", "/absolute", "sources/C:\\private"} {
		mutateDesignJSON(t, root, "design-evidence.json", func(m M) { obj(arr(m["claims"])[0])["file"] = rel })
		if len(DesignSystemCheck(os.DirFS(root), "")) == 0 {
			t.Fatal("unsafe path accepted", rel)
		}
	}
}
func TestDesignSystemRejectsFalseObservationAndSymlinks(t *testing.T) {
	root := designFixture(t)
	mutateDesignJSON(t, root, "design-evidence.json", func(m M) { obj(arr(m["claims"])[0])["kind"] = "observed" })
	if len(DesignSystemCheck(os.DirFS(root), "")) == 0 {
		t.Fatal("observation without evidence accepted")
	}
	root = designFixture(t)
	outside := filepath.Join(t.TempDir(), "tokens.css")
	b, e := os.ReadFile(filepath.Join(root, "tokens.css"))
	need(t, e)
	need(t, os.WriteFile(outside, b, 0644))
	need(t, os.Remove(filepath.Join(root, "tokens.css")))
	need(t, os.Symlink(outside, filepath.Join(root, "tokens.css")))
	if len(DesignSystemCheck(os.DirFS(root), "")) == 0 {
		t.Fatal("linked source accepted")
	}
	root = designFixture(t)
	need(t, os.Remove(filepath.Join(root, "design-preview.html")))
	need(t, os.Symlink(outside, filepath.Join(root, "design-preview.html")))
	if _, e = DesignSystemPreview(root, true); e == nil {
		t.Fatal("linked preview replaced")
	}
}

func TestDesignSystemDoesNotDistributeGitOrToolCaches(t *testing.T) {
	for _, rel := range []string{"sources/.git/config", "sources/tool-cache/session.txt", "sources/node_modules/cache.txt"} {
		root := designFixture(t)
		need(t, writeBytes(filepath.Join(root, rel), []byte("excluded fixture"), true))
		h, e := fileHash(filepath.Join(root, rel))
		need(t, e)
		mutateDesignJSON(t, root, "design-evidence.json", func(m M) { c := obj(arr(m["claims"])[0]); c["file"] = rel; c["sha256"] = h })
		if es := DesignSystemCheck(os.DirFS(root), ""); !strings.Contains(strings.Join(es, ";"), "Excluded package content") {
			t.Fatal(es)
		}
	}
}
func TestDesignSystemRejectsUnsafeOrBrokenTokens(t *testing.T) {
	for _, bad := range []string{
		":root {--ds-bg:red;} body {display:none}",
		":root {--ds-bg:url(https://example.invalid/x);}",
		":root {--ds-bg:</style><script>alert(1)</script>;}",
	} {
		if _, e := parseDesignTokens(bad); e == nil {
			t.Fatal("unsafe CSS accepted")
		}
	}
	root := designFixture(t)
	original, e := os.ReadFile(filepath.Join(root, "tokens.css"))
	need(t, e)
	for _, appendix := range []string{"--ds-extra:var(--ds-missing);", "--ds-a:var(--ds-b);--ds-b:var(--ds-a);", "--ds-bg:red;"} {
		css := strings.TrimSuffix(strings.TrimSpace(string(original)), "}") + appendix + "}"
		if _, e = parseDesignTokens(css); e == nil {
			t.Fatal("invalid references/duplicate accepted", appendix)
		}
	}
	css := strings.TrimSuffix(strings.TrimSpace(string(original)), "}") + "--ds-extra:var(--ds-missing, red);}"
	if _, e = parseDesignTokens(css); e == nil {
		t.Fatal("unsupported fallback concealed missing reference")
	}
}
func TestDesignSystemPreviewEscapesSourceAndDetectsStaleness(t *testing.T) {
	root := designFixture(t)
	mutateDesignJSON(t, root, "design-system.json", func(m M) { m["name"] = "<script>alert('x')</script>" })
	if len(DesignSystemCheck(os.DirFS(root), "")) == 0 {
		t.Fatal("metadata changed without regeneration")
	}
	_, e := DesignSystemPreview(root, true)
	need(t, e)
	b, e := os.ReadFile(filepath.Join(root, "design-preview.html"))
	need(t, e)
	if bytes.Contains(b, []byte("<script>alert('x')</script>")) || !bytes.Contains(b, []byte("&lt;script&gt;")) {
		t.Fatal("unescaped metadata")
	}
	if es := DesignSystemCheck(os.DirFS(root), ""); len(es) > 0 {
		t.Fatal(es)
	}
}

func TestDesignSystemPreviewCannotEscapeStyleThroughCSSComment(t *testing.T) {
	root := designFixture(t)
	css, e := os.ReadFile(filepath.Join(root, "tokens.css"))
	need(t, e)
	attack := "/* </style><script>globalThis.injected=true</script><style> */\n"
	need(t, writeBytes(filepath.Join(root, "tokens.css"), append([]byte(attack), css...), false))
	_, e = DesignSystemPreview(root, true)
	need(t, e)
	preview, e := os.ReadFile(filepath.Join(root, "design-preview.html"))
	need(t, e)
	if bytes.Contains(preview, []byte("globalThis.injected")) || bytes.Count(preview, []byte("<script>")) != 1 {
		t.Fatal("CSS comment escaped the preview style element")
	}
	if es := DesignSystemCheck(os.DirFS(root), ""); len(es) > 0 {
		t.Fatal(es)
	}
}
func TestDesignSystemCLIBoundaries(t *testing.T) {
	for _, args := range [][]string{{"design-system", "apply"}, {"design-system", "preview"}, {"design-system", "init", "--output", "x"}, {"design-system", "check", "--package", "x", "--replace"}, {"design-system", "list", "--output", "x"}, {"design-system", "list", "unexpected"}} {
		var out bytes.Buffer
		if CLI(args, &out) == 0 {
			t.Fatal("invalid flags accepted", args, out.String())
		}
	}
	var out bytes.Buffer
	if CLI([]string{"design-system", "list"}, &out) != 0 {
		t.Fatal(out.String())
	}
	var list []M
	need(t, json.Unmarshal(out.Bytes(), &list))
	if len(list) != 2 {
		t.Fatal(list)
	}
}
func TestDesignSystemStandaloneSkillContainsResources(t *testing.T) {
	target := filepath.Join(t.TempDir(), "skills")
	exe, e := os.Executable()
	need(t, e)
	names, e := Install(Bundle, target, []string{"design-reference-extraction"}, exe)
	need(t, e)
	if len(names) != 1 {
		t.Fatal(names)
	}
	for _, name := range []string{"references/package-contract.md", "assets/preview.gohtml", "assets/systems/calm-paper/tokens.css", "assets/systems/precise-console/design-preview.html"} {
		info, e := fs.Stat(os.DirFS(target), "design-reference-extraction/"+name)
		need(t, e)
		if !info.Mode().IsRegular() {
			t.Fatal(name)
		}
	}
}

func TestDesignLoopPreviewServesTokensWithCSSMIME(t *testing.T) {
	h, e := PreviewHandler(".", "design-system-loop")
	need(t, e)
	for _, route := range []string{"/styles.css", "/tokens.css"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", route, nil))
		if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/css") {
			t.Fatal(route, w.Code, w.Header())
		}
	}
	for _, route := range []string{"/design-system.json", "/sources/request.md", "/.product/design-lock.json"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", route, nil))
		if w.Code != 404 {
			t.Fatal("preview exposed metadata", route)
		}
	}
}
