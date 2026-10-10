package amazing

import (
	"bytes"
	"encoding/json"

	"io/fs"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func need(t *testing.T, e error) {
	t.Helper()
	if e != nil {
		t.Fatal(e)
	}
}
func qaFixture(t *testing.T) (string, []M) {
	t.Helper()
	r := t.TempDir()
	need(t, fs.WalkDir(Bundle, "examples/qa-continuous-change/sources", func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		b, e := fs.ReadFile(Bundle, p)
		if e != nil {
			return e
		}
		return writeBytes(path(r, "sources/"+d.Name()), b, true)
	}))
	_, e := QAInit(r)
	need(t, e)
	plans, e := fixturePlans(r)
	need(t, e)
	need(t, write(path(r, ".qa/plan.json"), plans[1], false))
	need(t, writeBytes(path(r, "evidence/observation.txt"), []byte("Observed fixture value"), true))
	return r, plans
}
func event(id, status, run string) M {
	return obj(clone(M{"execution_id": id, "run_id": run, "case_id": "TC-001", "case_revision": 2, "status": status, "mode": "synthetic-fixture", "environment": "Go fixture", "actual": "Recorded observation, no real device", "evidence": []any{"evidence/observation.txt"}, "started_at": "2026-10-10T00:00:00Z", "finished_at": "2026-10-10T00:00:01Z"}))
}
func checkHas(t *testing.T, root, text string) {
	t.Helper()
	if !strings.Contains(strings.Join(QACheck(root), "; "), text) {
		t.Fatalf("missing %q in %v", text, QACheck(root))
	}
}
func TestCatalogAndProviderIntegrity(t *testing.T) {
	if errs := Validate(Bundle); len(errs) > 0 {
		t.Fatal(errs)
	}
}
func TestClosure(t *testing.T) {
	c, e := Catalog(Bundle)
	need(t, e)
	for id, n := range map[string]int{"product-workflow": 8, "qa-workflow": 5} {
		names, e := Closure(c, []string{id})
		need(t, e)
		if len(names) != n {
			t.Fatal(names)
		}
	}
	_, e = Closure(M{"skills": []any{M{"id": "a", "requires": []any{"b"}}, M{"id": "b", "requires": []any{"a"}}}}, []string{"a"})
	if e == nil {
		t.Fatal("cycle accepted")
	}
}
func TestInstallAndRelocateRuntimeWithoutPython(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "amazing-skills")
	cmd := exec.Command("go", "build", "-trimpath", "-o", binary, "./cmd/amazing-skills")
	out, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("%v: %s", e, out)
	}
	target := filepath.Join(dir, "portable skills")
	names, e := Install(Bundle, target, []string{"qa-workflow"}, binary)
	need(t, e)
	if len(names) != 5 {
		t.Fatal(names)
	}
	newTarget := filepath.Join(dir, "moved skills")
	need(t, os.Rename(target, newTarget))
	installed := filepath.Join(newTarget, ".amazing-skills/amazing-skills")
	project := filepath.Join(dir, "Go project")
	cmd = exec.Command(installed, "replay", "--output", project)
	cmd.Dir = dir
	cmd.Env = []string{"PATH=" + filepath.Join(dir, "empty-path")}
	out, e = cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("No-Python replay: %v %s", e, out)
	}
	if !strings.Contains(string(out), `"workflow_replay":"passed"`) {
		t.Fatal(string(out))
	}
	cmd = exec.Command(installed, "qa", "check", "--project", project)
	cmd.Dir = dir
	cmd.Env = []string{"PATH=" + filepath.Join(dir, "empty-path")}
	out, e = cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("%v %s", e, out)
	}
	if strings.Contains(string(out), `"business_or_execution_verified":true`) {
		t.Fatal(string(out))
	}

	designInstall := filepath.Join(dir, "workbench skills")
	_, e = Install(Bundle, designInstall, []string{"product-workflow"}, binary)
	need(t, e)
	designMoved := filepath.Join(dir, "moved workbench")
	need(t, os.Rename(designInstall, designMoved))
	designBinary := filepath.Join(designMoved, ".amazing-skills/amazing-skills")
	designProject := filepath.Join(dir, "design project")
	cmd = exec.Command(designBinary, "search", "SaaS dashboard dark mode", "--design-system", "--project-name", "Portable Demo", "--density", "8", "--motion", "2", "--persist", "--output-dir", designProject, "--json")
	cmd.Dir = dir
	cmd.Env = []string{"PATH=" + filepath.Join(dir, "empty-path")}
	out, e = cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("No-Python design system: %v %s", e, out)
	}
	var design M
	need(t, json.Unmarshal(out, &design))
	if obj(design["persistence"])["status"] != "success" || len(obj(obj(design["design_system"])["source_identities"])) != 6 {
		t.Fatal(string(out))
	}
	master := str(obj(design["persistence"])["master_file"])
	before, e := os.ReadFile(master)
	need(t, e)
	cmd = exec.Command(designBinary, "search", "ecommerce luxury checkout", "--design-system", "--project-name", "Portable Demo", "--persist", "--output-dir", designProject, "--json")
	cmd.Dir = dir
	cmd.Env = []string{"PATH=" + filepath.Join(dir, "empty-path")}
	out, e = cmd.CombinedOutput()
	need(t, e)
	if !strings.Contains(string(out), `"status":"skipped_exists"`) {
		t.Fatal(string(out))
	}
	after, e := os.ReadFile(master)
	need(t, e)
	if string(before) != string(after) {
		t.Fatal("portable query changed master")
	}
	_, e = Install(Bundle, newTarget, []string{"qa-workflow"}, binary)
	if e == nil {
		t.Fatal("overwrite accepted")
	}
}
func TestInstalledEntryPoints(t *testing.T) {
	dir := t.TempDir()
	exe, _ := os.Executable()
	target := filepath.Join(dir, "skills")
	names, e := Install(Bundle, target, nil, exe)
	need(t, e)
	if len(names) != 13 {
		t.Fatal(names)
	}
	c := obj(mustRead(filepath.Join(target, "skill-router/references/catalog.json")))
	r := obj(c["entrypoints"])
	ids := []string{str(r["default"])}
	for _, v := range obj(r["domains"]) {
		ids = append(ids, str(v))
	}
	for _, id := range ids {
		if !isFile(filepath.Join(target, id, "SKILL.md")) {
			t.Fatal(id)
		}
	}
}
func TestProjectInitPreservesDecisions(t *testing.T) {
	r := t.TempDir()
	_, e := ProjectInit(r)
	need(t, e)
	need(t, writeBytes(path(r, "DESIGN.md"), []byte("Existing decisions"), false))
	created, e := ProjectInit(r)
	need(t, e)
	if len(created) != 0 {
		t.Fatal(created)
	}
	b, _ := os.ReadFile(path(r, "DESIGN.md"))
	if string(b) != "Existing decisions" {
		t.Fatal(string(b))
	}
}
func TestIncompleteProjectCannotClaimComplete(t *testing.T) {
	r := t.TempDir()
	_, e := ProjectInit(r)
	need(t, e)
	s := obj(mustRead(path(r, ".product/state.json")))
	s["phase"] = "complete"
	s["status"] = "passed"
	need(t, write(path(r, ".product/state.json"), s, false))
	if len(ProjectCheck(r)) == 0 {
		t.Fatal("empty complete accepted")
	}
}
func TestStyleLock(t *testing.T) {
	r := t.TempDir()
	need(t, writeBytes(path(r, "DESIGN.md"), []byte("Pinned"), true))
	need(t, writeBytes(path(r, "styles.css"), []byte(":root { --accent: green; } .button {padding:10px;}"), true))
	_, e := LockDesign(r, "styles.css", false)
	need(t, e)
	need(t, writeBytes(path(r, "styles.css"), []byte(":root { --accent: green; } .new {padding:20px;}"), false))
	if len(CheckDesign(r)) != 0 {
		t.Fatal(CheckDesign(r))
	}
	need(t, writeBytes(path(r, "styles.css"), []byte(":root { --accent: red; }"), false))
	if len(CheckDesign(r)) == 0 {
		t.Fatal("style drift accepted")
	}
	_, e = LockDesign(r, "styles.css", false)
	if e == nil {
		t.Fatal("replaced without authorization")
	}
}
func TestSourceDrift(t *testing.T) {
	r, _ := qaFixture(t)
	need(t, writeBytes(path(r, "sources/r2.md"), []byte("changed"), false))
	checkHas(t, r, "Source missing or changed")
}
func TestUnknownExpectedCannotBeReady(t *testing.T) {
	r, p := qaFixture(t)
	for _, v := range arr(p[0]["cases"]) {
		c := obj(v)
		if c["id"] == "TC-006" {
			c["status"] = "ready"
		}
	}
	need(t, write(path(r, ".qa/plan.json"), p[0], false))
	checkHas(t, r, "inactive or stale")
}
func TestDimensionCannotExpandScope(t *testing.T) {
	r, p := qaFixture(t)
	c := obj(arr(p[1]["cases"])[0])
	c["dimensions"] = append(arr(c["dimensions"]), "thermal")
	need(t, write(path(r, ".qa/plan.json"), p[1], false))
	checkHas(t, r, "unselected dimension")
}
func TestRequirementCoverage(t *testing.T) {
	r, p := qaFixture(t)
	newReq := obj(clone(arr(p[1]["requirements"])[0]))
	newReq["id"] = "R-MISSING"
	p[1]["requirements"] = append(arr(p[1]["requirements"]), newReq)
	need(t, write(path(r, ".qa/plan.json"), p[1], false))
	checkHas(t, r, "no case or explicit coverage gap")
}
func TestFrozenRevisionRequired(t *testing.T) {
	r, p := qaFixture(t)
	_, e := Snapshot(r, "baseline")
	need(t, e)
	obj(arr(p[1]["requirements"])[0])["text"] = "New unversioned rule"
	obj(arr(p[1]["cases"])[0])["expected"] = []any{"New assertion"}
	need(t, write(path(r, ".qa/plan.json"), p[1], false))
	checkHas(t, r, "Requirement changed without revision")
	checkHas(t, r, "Case changed without revision")
}
func TestSnapshotNoOverwriteOrTraversal(t *testing.T) {
	r, _ := qaFixture(t)
	_, e := Snapshot(r, "baseline")
	need(t, e)
	for _, label := range []string{"baseline", "../outside"} {
		_, e = Snapshot(r, label)
		if e == nil {
			t.Fatal("invalid snapshot accepted")
		}
	}
}
func TestSnapshotTamper(t *testing.T) {
	r, _ := qaFixture(t)
	_, e := Snapshot(r, "baseline")
	need(t, e)
	need(t, writeBytes(path(r, ".qa/snapshots/baseline/config.json"), []byte("{}"), false))
	checkHas(t, r, "Frozen snapshot drift")
}
func TestEvidenceTamper(t *testing.T) {
	r, _ := qaFixture(t)
	_, e := Record(r, event("E1", "passed", "run1"))
	need(t, e)
	need(t, writeBytes(path(r, "evidence/observation.txt"), []byte("changed"), false))
	checkHas(t, r, "Execution evidence drift")
}
func TestRecordIdempotencyAndCollision(t *testing.T) {
	r, _ := qaFixture(t)
	o := event("E1", "passed", "run1")
	_, e := Record(r, o)
	need(t, e)
	v, e := Record(r, o)
	need(t, e)
	if v["idempotent"] != true {
		t.Fatal(v)
	}
	o["actual"] = "Changed"
	_, e = Record(r, o)
	if e == nil {
		t.Fatal("identity collision accepted")
	}
	events, e := ReadEvents(r)
	need(t, e)
	if len(events) != 1 {
		t.Fatal(len(events))
	}
}
func TestFirstFailureAndIndependentRetest(t *testing.T) {
	r, _ := qaFixture(t)
	for _, o := range []M{event("fail", "failed", "run1"), event("retry", "passed", "run1")} {
		_, e := Record(r, o)
		need(t, e)
	}
	s, e := Summary(r, "run1")
	need(t, e)
	if num(obj(s["counts"])["failed"]) != 1 {
		t.Fatal(s)
	}
	_, e = Record(r, event("retest", "passed", "run2"))
	need(t, e)
	s, e = Summary(r, "")
	need(t, e)
	if num(obj(s["counts"])["passed"]) != 1 || len(arr(s["historical_failures"])) != 1 || s["quality_status"] != "synthetic-only" {
		t.Fatal(s)
	}
}
func TestChangedRevisionAndWithdrawal(t *testing.T) {
	r, p := qaFixture(t)
	_, e := Snapshot(r, "baseline")
	need(t, e)
	o := event("E1", "passed", "run1")
	o["case_id"] = "TC-006"
	_, e = Record(r, o)
	need(t, e)
	need(t, write(path(r, ".qa/plan.json"), p[2], false))
	s, e := Summary(r, "")
	need(t, e)
	if num(obj(s["counts"])["passed"]) != 0 || num(s["total"]) != 12 || s["pass_rate"] != nil || obj(s["defect_summary"])["confirmed_total"] != nil {
		t.Fatal(s)
	}
}
func TestFrozenIdentityRemoval(t *testing.T) {
	r, p := qaFixture(t)
	_, e := Snapshot(r, "baseline")
	need(t, e)
	cs := arr(p[1]["cases"])
	p[1]["cases"] = cs[:len(cs)-1]
	need(t, write(path(r, ".qa/plan.json"), p[1], false))
	checkHas(t, r, "Frozen case removed")
}
func TestTruncatedHistory(t *testing.T) {
	r, _ := qaFixture(t)
	_, e := Record(r, event("E1", "passed", "run1"))
	need(t, e)
	need(t, writeBytes(path(r, ".qa/executions.jsonl"), []byte{}, false))
	_, e = ReadEvents(r)
	if e == nil {
		t.Fatal("truncation accepted")
	}
}
func TestRecordEvidenceModeAndTime(t *testing.T) {
	r, _ := qaFixture(t)
	for _, modify := range []func(M){func(m M) { m["evidence"] = []any{} }, func(m M) { m["mode"] = "pretend-real-device" }, func(m M) { m["started_at"] = "2026-10-10T00:00:00" }, func(m M) { m["finished_at"] = "2026-10-09T00:00:00Z" }} {
		o := event("E1", "passed", "run1")
		modify(o)
		_, e := Record(r, o)
		if e == nil {
			t.Fatal("invalid execution accepted")
		}
	}
}
func TestPathAndSymlinkEscape(t *testing.T) {
	r, _ := qaFixture(t)
	outside := filepath.Join(t.TempDir(), "outside.txt")
	need(t, os.WriteFile(outside, []byte("outside"), 0644))
	need(t, os.Symlink(outside, path(r, "evidence/link.txt")))
	for _, p := range []string{"../outside", "/absolute", "evidence/link.txt", "C:\\Users\\someone\\file"} {
		o := event("E1", "passed", "run1")
		o["evidence"] = []any{p}
		_, e := Record(r, o)
		if e == nil {
			t.Fatalf("escape accepted: %s", p)
		}
	}
}
func TestReceiptIncompleteAndMock(t *testing.T) {
	r, _ := qaFixture(t)
	receipt := obj(clone(M{"mode": "mock", "operation_id": "create-1", "operation": "create", "issue_url": "mock://issues/1", "write_outcome": "saved", "readback_confirmed": true, "expected_payload_hash": strings.Repeat("a", 64), "readback_payload_hash": strings.Repeat("a", 64), "evidence_entries": []any{M{"local": "evidence/observation.txt", "uploaded": true, "visible": false, "remote_id": "A1"}}}))
	if Receipt(r, receipt)["receipt_check"] != "incomplete" {
		t.Fatal("incomplete accepted")
	}
	obj(arr(receipt["evidence_entries"])[0])["visible"] = true
	result := Receipt(r, receipt)
	if result["receipt_check"] != "complete" || result["real_jira_verified"] != false {
		t.Fatal(result)
	}
	receipt["readback_payload_hash"] = strings.Repeat("b", 64)
	if Receipt(r, receipt)["receipt_check"] != "incomplete" {
		t.Fatal("payload mismatch accepted")
	}
}
func TestDraftRevisionsAndPrivacy(t *testing.T) {
	r, _ := qaFixture(t)
	d := obj(clone(M{"id": "FD1", "summary": "Observed issue", "environment": "Synthetic", "steps": []any{"Do action"}, "actual": "Observed", "expected": "Expected", "basis": "Source", "evidence": []any{"evidence/observation.txt"}}))
	v, e := Draft(r, d)
	need(t, e)
	b, _ := os.ReadFile(path(r, str(v["draft"])))
	if strings.Contains(string(b), r) {
		t.Fatal("private path leaked")
	}
	_, e = Draft(r, d)
	need(t, e)
	d["actual"] = "Changed"
	_, e = Draft(r, d)
	if e == nil {
		t.Fatal("draft overwritten")
	}
}
func TestGoReplay(t *testing.T) {
	r := filepath.Join(t.TempDir(), "new project")
	v, e := Replay(r)
	need(t, e)
	if v["workflow_replay"] != "passed" || num(v["cases"]) != 12 {
		t.Fatal(v)
	}
	_, e = Replay(r)
	if e == nil {
		t.Fatal("replay overwrote previous run")
	}
}
func TestLegacyPythonHistoryReadable(t *testing.T) {
	r := "examples/qa-continuous-change"
	events, e := ReadEvents(r)
	need(t, e)
	if len(events) < 50 {
		t.Fatal(len(events))
	}
	if errs := QACheck(r); len(errs) > 0 {
		t.Fatal(errs)
	}
}
func TestNativeSearchUsesPinnedCSVWithoutPython(t *testing.T) {
	for _, query := range []struct{ Q, D, S string }{{"category count comparison", "chart", ""}, {"dashboard restrained", "style", ""}, {"keyboard form", "ux", ""}, {"state hooks", "style", "react"}} {
		r, e := Search("skills", query.Q, query.D, query.S, 3)
		need(t, e)
		if num(r["count"]) == 0 {
			t.Fatal(r)
		}
	}
	_, e := NativeSearchCLI("skills", []string{"dashboard", "--persist"})
	if e == nil {
		t.Fatal("unsupported persistence silently accepted")
	}
}
func TestDesignSystemHasTraceableReasoning(t *testing.T) {
	r, e := DesignSuggestions("skills", "project dashboard", M{"density": 8, "motion": 2})
	need(t, e)
	if len(obj(obj(r["design_system"])["source_identities"])) != 6 || obj(obj(r["design_system"])["dials"])["density"] != 8 || r["persistence"] != nil {
		t.Fatal(r)
	}
}
func TestPreviewDoesNotExposeSecrets(t *testing.T) {
	h, e := PreviewHandler(".", "customer-followup")
	need(t, e)
	for _, p := range []string{"/.product/state.json", "/.git/config", "/../catalog.json", "/evidence/desktop-empty.png"} {
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest("GET", p, nil))
		if r.Code != 404 {
			t.Fatal(p, r.Code)
		}
	}
	r := httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest("GET", "/", nil))
	if r.Code != 200 {
		t.Fatal(r.Code)
	}
}
func TestCLIValidationAndInvalidFlags(t *testing.T) {
	var b bytes.Buffer
	if CLI([]string{"qa", "record"}, &b) == 0 {
		t.Fatal("missing args accepted")
	}
	if CLI([]string{"validate", "--not-real", "value"}, &b) == 0 {
		t.Fatal("Unknown flag accepted")
	}
	if CLI([]string{"validate"}, &b) != 0 {
		t.Fatal(b.String())
	}
}

func TestDanglingSymlinkCannotWriteOutside(t *testing.T) {
	r := t.TempDir()
	outside := filepath.Join(t.TempDir(), "not-created.txt")
	need(t, os.Symlink(outside, filepath.Join(r, "link")))
	if _, e := Inside(r, "link"); e == nil {
		t.Fatal("Dangling symlink accepted")
	}
	if exists(outside) {
		t.Fatal("Wrote outside")
	}
}
