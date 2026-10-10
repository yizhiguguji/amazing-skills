package amazing

import (
	"fmt"
	"html"
	"io/fs"
	"strings"
	"unicode/utf8"
)

type securityFixture struct {
	nickname map[string]string
	max      int
	broken   bool
}

func (m *securityFixture) save(actor, owner, value string) bool {
	if actor != owner && !m.broken {
		return false
	}
	if n := utf8.RuneCountInString(value); n < 2 || n > m.max {
		return false
	}
	m.nickname[owner] = value
	return true
}
func (m *securityFixture) read(actor, owner string) (string, bool) {
	if actor != owner && !m.broken {
		return "", false
	}
	return m.nickname[owner], true
}
func securityPlans(root string) ([]M, error) {
	out := []M{}
	for _, v := range []string{"r1", "r2"} {
		p, e := load(Bundle, "examples/qa-security-scope/fixtures/"+v+".json")
		if e != nil {
			return nil, e
		}
		plan := obj(p)
		for _, s := range arr(plan["sources"]) {
			row := obj(s)
			h, e := fileHash(path(root, str(row["locator"])))
			if e != nil {
				return nil, e
			}
			row["sha256"] = h
		}
		out = append(out, plan)
	}
	return out, nil
}

// SecurityReplay is a bounded synthetic model, not a scan, browser test, or independent LLM routing evaluation.
func SecurityReplay(root string) (M, error) {
	if exists(root) {
		return nil, fmt.Errorf("Replay needs a fresh directory")
	}
	if e := fs.WalkDir(Bundle, "examples/qa-security-scope/sources", func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		return writeBytes(path(root, "sources/"+d.Name()), mustReadBundle(p), true)
	}); e != nil {
		return nil, e
	}
	if _, e := QAInit(root); e != nil {
		return nil, e
	}
	plans, e := securityPlans(root)
	if e != nil {
		return nil, e
	}
	if e = write(path(root, ".qa/plan.json"), plans[0], false); e != nil {
		return nil, e
	}
	if e = requireQA(root); e != nil {
		return nil, e
	}
	if _, e = Snapshot(root, "security-r1"); e != nil {
		return nil, e
	}
	model := securityFixture{map[string]string{"test-user-a": "原昵称甲", "test-user-b": "原昵称乙"}, 12, true}
	checks := M{}
	observe := func(id, run, status, actual string, revision int) error {
		file := "evidence/" + run + "-" + id + "-" + status + ".txt"
		if e := writeBytes(path(root, file), []byte("Synthetic model only. "+actual), true); e != nil {
			return e
		}
		_, e := Record(root, M{"execution_id": run + "-" + id + "-" + status, "run_id": run, "case_id": id, "case_revision": revision, "status": status, "mode": "synthetic-fixture", "environment": "bounded in-memory profile fixture; no network/device", "actual": actual, "evidence": []any{file}, "started_at": "2026-10-10T00:00:00Z", "finished_at": "2026-10-10T00:00:01Z"})
		return e
	}
	if !model.save("test-user-a", "test-user-a", "示例昵称") {
		return nil, fmt.Errorf("happy flow failed")
	}
	if e = observe("TC-001", "r1", "passed", "Own valid nickname saved", 1); e != nil {
		return nil, e
	}
	before := model.nickname["test-user-a"]
	if model.save("test-user-a", "test-user-a", strings.Repeat("名", 13)) || model.nickname["test-user-a"] != before {
		return nil, fmt.Errorf("boundary failed")
	}
	if e = observe("TC-002", "r1", "passed", "13 codepoints rejected at max 12; data unchanged", 1); e != nil {
		return nil, e
	}
	if _, ok := model.read("test-user-a", "test-user-b"); !ok {
		return nil, fmt.Errorf("expected seeded access bug missing")
	}
	if e = observe("TC-003", "r1", "failed", "Seeded fixture bug returns another test user's nickname", 1); e != nil {
		return nil, e
	}
	model.broken = false
	if _, ok := model.read("test-user-a", "test-user-b"); ok || model.save("test-user-a", "test-user-b", "错误昵称") {
		return nil, fmt.Errorf("fix failed")
	}
	if e = observe("TC-003", "r1", "passed", "Fixed fixture refuses cross-user read/write; retry kept separate", 1); e != nil {
		return nil, e
	}
	if html.EscapeString("<em>示例</em>") != "&lt;em&gt;示例&lt;/em&gt;" {
		return nil, fmt.Errorf("text fixture failed")
	}
	if e = observe("TC-004", "r1", "passed", "Model HTML encoding returns escaped text; no browser execution proven", 1); e != nil {
		return nil, e
	}
	r1, e := Summary(root, "r1")
	if e != nil {
		return nil, e
	}
	if num(obj(r1["counts"])["failed"]) != 1 {
		return nil, fmt.Errorf("first failure was hidden")
	}
	checks["same_run_failure_retained"] = "passed"
	// An uncertain security policy must not become a fixed executable expectation.
	bad := obj(clone(plans[0]))
	for _, v := range arr(bad["cases"]) {
		c := obj(v)
		if c["id"] == "TC-005" {
			c["status"] = "ready"
			c["expected"] = []string{"Session expires in 30 minutes with 401"}
		}
	}
	if len(SecurityCheck(bad)) == 0 {
		return nil, fmt.Errorf("invented expectation accepted")
	}
	checks["unknown_expectation_rejected"] = "passed"
	bad = obj(clone(plans[0]))
	obj(arr(obj(bad["security"])["targets"])[0])["requirement_ids"] = []string{"unrelated-upload"}
	if len(SecurityCheck(bad)) == 0 {
		return nil, fmt.Errorf("expanded scope accepted")
	}
	checks["scope_expansion_rejected"] = "passed"
	if e = write(path(root, ".qa/plan.json"), plans[1], false); e != nil {
		return nil, e
	}
	if e = requireQA(root); e != nil {
		return nil, e
	}
	impact, e := Impact(root, "security-r1")
	if e != nil {
		return nil, e
	}
	checks["withdrawal_and_changed_revision"] = "passed"
	r2, e := Summary(root, "r2")
	if e != nil {
		return nil, e
	}
	if num(obj(r2["counts"])["passed"]) != 0 {
		return nil, fmt.Errorf("old pass leaked into new run")
	}
	checks["new_revision_not_inheriting_pass"] = "passed"
	// Specialist readiness does not grant execution permission.
	specialist := obj(clone(plans[1]))
	obj(specialist["security"])["mode"] = "specialist"
	if e = write(path(root, ".qa/plan.json"), specialist, false); e != nil {
		return nil, e
	}
	case3 := obj(arr(specialist["cases"])[2])
	if SecurityRecordAllowed(specialist, case3, M{"status": "passed"}) == nil {
		return nil, fmt.Errorf("specialist execution permission fabricated")
	}
	checks["specialist_without_authorization_rejected"] = "passed"
	obj(specialist["security"])["execution"] = M{"status": "authorized", "target_ids": []string{"T1"}, "allowed_actions": []string{"observe", "request-modification"}, "source_refs": []M{{"source_id": "S1", "locator": "AUTH-1"}}, "stop_conditions": []string{"Target escape, unexpected data, or model failure"}, "budget": "At most 10 isolated model calls; no network"}
	if e = write(path(root, ".qa/plan.json"), specialist, false); e != nil {
		return nil, e
	}
	if e = requireQA(root); e != nil {
		return nil, e
	}
	if _, ok := model.read("test-user-a", "test-user-b"); ok || model.save("test-user-a", "test-user-b", "错误昵称") {
		return nil, fmt.Errorf("authorized retest failed")
	}
	if e = observe("TC-003", "r2", "passed", "Authorized synthetic retest refuses cross-user read/write; no real scan", 1); e != nil {
		return nil, e
	}
	if _, e = Snapshot(root, "security-r2"); e != nil {
		return nil, e
	}
	final, e := Summary(root, "r2")
	if e != nil {
		return nil, e
	}
	out := M{"mode": "synthetic-fixture", "checks": checks, "first_run": r1, "impact": impact, "current_run": final, "not_verified": []string{"Natural-language routing by a separate AI", "Real application/browser/native device", "Real security vulnerabilities or compliance", "External Jira submission"}}
	if e = write(path(root, "replay-results.json"), out, false); e != nil {
		return nil, e
	}
	return out, nil
}
