package amazing

import (
	"path/filepath"
	"strings"
	"testing"
)

func securityFixturePlan(t *testing.T) (string, M) {
	root, plans := qaFixture(t)
	p := plans[1]
	var c M
	for _, v := range arr(p["cases"]) {
		if obj(v)["id"] == "TC-001" {
			c = obj(v)
		}
	}
	r := obj(arr(c["requirement_refs"])[0])
	rid := str(r["id"])
	c["security"] = M{"check_id": "SC1", "target_id": "T1", "oracle": "confirmed", "oracle_refs": []any{clone(r)}, "action": "observe"}
	p["security"] = M{"mode": "functional", "scope_requirement_ids": []string{rid}, "targets": []M{{"id": "T1", "platform": "model", "locator": "fixture:note", "requirement_ids": []string{rid}}}, "checks": []M{{"id": "SC1", "family": "authorization", "decision": "include", "reason": "Explicit in-scope rule", "requirement_ids": []string{rid}, "target_ids": []string{"T1"}, "case_ids": []string{"TC-001"}}}}
	need(t, write(path(root, ".qa/plan.json"), p, false))
	if errs := QACheck(root); len(errs) > 0 {
		t.Fatal(errs)
	}
	return root, p
}
func TestSecurityScopeAndOracleGuards(t *testing.T) {
	_, base := securityFixturePlan(t)
	tests := []struct {
		name, expect string
		mutate       func(M)
	}{
		{"scope", "outside selected scope", func(p M) { obj(arr(obj(p["security"])["targets"])[0])["requirement_ids"] = []string{"unrelated"} }},
		{"unknown", "Unknown security expectation", func(p M) { obj(obj(arr(p["cases"])[0])["security"])["oracle"] = "unknown" }},
		{"source", "expectation needs source", func(p M) { obj(obj(arr(p["cases"])[0])["security"])["oracle_refs"] = []any{} }},
		{"stale", "stale or unrelated", func(p M) { obj(arr(obj(obj(arr(p["cases"])[0])["security"])["oracle_refs"])[0])["revision"] = 999 }},
		{"target", "missing from scope ledger", func(p M) { obj(obj(arr(p["cases"])[0])["security"])["target_id"] = "unrelated" }},
		{"ledger", "ledger and case mismatch", func(p M) { obj(arr(obj(p["security"])["checks"])[0])["case_ids"] = []string{"TC-missing"} }},
		{"gap", "needs gap", func(p M) { obj(arr(obj(p["security"])["checks"])[0])["decision"] = "clarify" }},
		{"no-scope", "needs scope ledger", func(p M) { delete(p, "security") }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := obj(clone(base))
			test.mutate(p)
			errs := SecurityCheck(p)
			if !strings.Contains(strings.Join(errs, ";"), test.expect) {
				t.Fatal(errs)
			}
		})
	}
}
func TestSecuritySpecialistExecutionNeedsTargetAndAction(t *testing.T) {
	root, p := securityFixturePlan(t)
	s := obj(p["security"])
	s["mode"] = "specialist"
	for _, v := range arr(p["cases"]) {
		c := obj(v)
		if c["id"] != "TC-001" && c["status"] == "ready" {
			c["status"] = "deferred"
		}
	}
	need(t, write(path(root, ".qa/plan.json"), p, false))
	if _, e := Record(root, event("security-no-auth", "passed", "r1")); e == nil {
		t.Fatal("readiness fabricated execution authorization")
	}
	s["execution"] = M{"status": "authorized", "target_ids": []string{"T1"}, "allowed_actions": []string{"observe"}, "source_refs": []M{{"source_id": str(obj(arr(p["sources"])[0])["id"]), "locator": "simulated allowed scope"}}, "stop_conditions": []string{"Unexpected target or error"}, "budget": "one isolated mock observation"}
	need(t, write(path(root, ".qa/plan.json"), p, false))
	_, e := Record(root, event("security-authorized", "passed", "r1"))
	need(t, e)
	c := obj(arr(p["cases"])[0])
	if c["id"] != "TC-001" {
		t.Fatal("fixture identity")
	}
	obj(c["security"])["action"] = "request-modification"
	if SecurityRecordAllowed(p, c, M{"status": "passed"}) == nil {
		t.Fatal("action expansion accepted")
	}
	obj(c["security"])["target_id"] = "T2"
	if SecurityRecordAllowed(p, c, M{"status": "failed"}) == nil {
		t.Fatal("target expansion accepted")
	}
}
func TestSecurityChangesRequireRevision(t *testing.T) {
	root, p := securityFixturePlan(t)
	_, e := Snapshot(root, "security-base")
	need(t, e)
	obj(obj(arr(p["cases"])[0])["security"])["action"] = "request-modification"
	need(t, write(path(root, ".qa/plan.json"), p, false))
	checkHas(t, root, "Case changed without revision")
}
func TestSecurityReplayBoundsUnknownAndHistory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "replay")
	r, e := SecurityReplay(root)
	need(t, e)
	if r["mode"] != "synthetic-fixture" || len(obj(r["checks"])) != 6 {
		t.Fatal(r)
	}
	counts := obj(obj(r["current_run"])["counts"])
	if num(counts["passed"]) != 1 || num(counts["not-run"]) != 2 || len(arr(obj(r["current_run"])["retired_cases"])) != 1 || len(arr(obj(r["current_run"])["proposed_cases"])) != 1 {
		t.Fatal(r)
	}
	if _, e = SecurityReplay(root); e == nil {
		t.Fatal("existing replay overwritten")
	}
}
func TestFunctionalSecurityCannotBypassActiveActionAuthorization(t *testing.T) {
	root, p := securityFixturePlan(t)
	obj(obj(arr(p["cases"])[0])["security"])["action"] = "scan"
	need(t, write(path(root, ".qa/plan.json"), p, false))
	if _, e := Record(root, event("functional-active", "passed", "r1")); e == nil {
		t.Fatal("functional label bypassed scan authorization")
	}
}
func TestUnknownSecurityPolicyRemainsProposed(t *testing.T) {
	root, p := securityFixturePlan(t)
	c := obj(arr(p["cases"])[0])
	c["status"] = "proposed"
	c["expected"] = []any{}
	m := obj(c["security"])
	m["oracle"] = "unknown"
	m["oracle_refs"] = []any{}
	check := obj(arr(obj(p["security"])["checks"])[0])
	check["decision"] = "clarify"
	check["gap_ids"] = []string{"SEC-G1"}
	p["gaps"] = append(arr(p["gaps"]), M{"id": "SEC-G1", "requirement_ids": obj(p["security"])["scope_requirement_ids"], "question": "Policy not defined", "close_condition": "Project confirms rule and current source"})
	need(t, write(path(root, ".qa/plan.json"), p, false))
	if errs := QACheck(root); len(errs) > 0 {
		t.Fatal(errs)
	}
	if _, e := Record(root, event("unknown-pass", "passed", "r1")); e == nil {
		t.Fatal("unknown policy passed")
	}
	_, e := ExportCases(root)
	need(t, e)
	report, e := Summary(root, "")
	need(t, e)
	if report["security_review"] == nil {
		t.Fatal("security scope missing in report")
	}
}
func TestSecurityTargetChangeInvalidatesOldContext(t *testing.T) {
	for _, history := range []string{"snapshot", "record"} {
		t.Run(history, func(t *testing.T) {
			root, p := securityFixturePlan(t)
			if history == "snapshot" {
				_, e := Snapshot(root, "target-base")
				need(t, e)
			} else {
				_, e := Record(root, event("target-base", "passed", "r1"))
				need(t, e)
			}
			obj(arr(obj(p["security"])["targets"])[0])["locator"] = "fixture:different-entry"
			need(t, write(path(root, ".qa/plan.json"), p, false))
			checkHas(t, root, "Security target changed without case revision")
		})
	}
}
