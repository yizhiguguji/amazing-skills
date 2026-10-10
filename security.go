package amazing

import "fmt"

// Optional security metadata validates traceability, not real authorization or security correctness.
func SecurityCheck(plan M) []string {
	errs := []string{}
	reqs := index(plan["requirements"], "requirements", &errs)
	cases := index(plan["cases"], "cases", &errs)
	gaps := index(plan["gaps"], "gaps", &errs)
	sources := index(plan["sources"], "sources", &errs)
	raw, enabled := plan["security"]
	if !enabled {
		for id, c := range cases {
			if c["security"] != nil && c["status"] != "retired" {
				errs = append(errs, "Security case needs scope ledger: "+id)
			}
		}
		return dedup(errs)
	}
	s := obj(raw)
	if !member(s["mode"], "functional", "specialist") {
		errs = append(errs, "Security mode must be functional or specialist")
	}
	scope := arr(s["scope_requirement_ids"])
	if !stringList(s["scope_requirement_ids"]) {
		errs = append(errs, "Security scope needs requirement IDs")
	}
	for _, id := range scope {
		if reqs[str(id)]["status"] != "active" {
			errs = append(errs, "Security scope needs active requirement: "+str(id))
		}
	}
	targets := index(s["targets"], "security targets", &errs)
	checks := index(s["checks"], "security checks", &errs)
	if len(checks) == 0 {
		errs = append(errs, "Security review needs candidate decisions")
	}
	bounded := func(ids any, label string) {
		if !stringList(ids) {
			errs = append(errs, "Security scope trace missing: "+label)
		}
		for _, id := range arr(ids) {
			if !has(scope, str(id)) {
				errs = append(errs, "Security outside selected scope: "+label+"/"+str(id))
			}
		}
	}
	for id, target := range targets {
		bounded(target["requirement_ids"], id)
		if !present(target["platform"]) || !present(target["locator"]) {
			errs = append(errs, "Security target needs platform and locator: "+id)
		}
	}
	for id, check := range checks {
		decision := str(check["decision"])
		if !member(decision, "include", "not-applicable", "clarify", "deferred") || !present(check["reason"]) || !present(check["family"]) {
			errs = append(errs, "Security candidate needs decision, family and reason: "+id)
		}
		bounded(check["requirement_ids"], id)
		for _, tid := range arr(check["target_ids"]) {
			if targets[str(tid)] == nil {
				errs = append(errs, "Unknown security target: "+id+"/"+str(tid))
			}
		}
		if decision == "include" && (!stringList(check["target_ids"]) || !stringList(check["case_ids"])) {
			errs = append(errs, "Included security check needs targets and cases: "+id)
		}
		if decision == "not-applicable" && len(arr(check["case_ids"])) > 0 {
			errs = append(errs, "Not-applicable security check cannot include cases: "+id)
		}
		if member(decision, "clarify", "deferred") {
			if !stringList(check["gap_ids"]) {
				errs = append(errs, "Security candidate needs gap: "+id)
			}
			for _, gid := range arr(check["gap_ids"]) {
				if gaps[str(gid)] == nil {
					errs = append(errs, "Security candidate needs gap: "+id)
				}
			}
		}
		for _, cid := range arr(check["case_ids"]) {
			c := cases[str(cid)]
			meta := obj(c["security"])
			if c == nil || c["status"] == "retired" || meta["check_id"] != id || !has(check["target_ids"], str(meta["target_id"])) {
				errs = append(errs, "Security ledger and case mismatch: "+id+"/"+str(cid))
			}
		}
	}
	for id, c := range cases {
		if s["mode"] == "specialist" && c["status"] == "ready" && c["security"] == nil {
			errs = append(errs, "Specialist executable case needs security scope metadata: "+id)
		}
		if c["status"] == "retired" || c["security"] == nil {
			continue
		}
		meta := obj(c["security"])
		check := checks[str(meta["check_id"])]
		target := targets[str(meta["target_id"])]
		if check == nil || target == nil || !has(check["case_ids"], id) || !has(check["target_ids"], str(meta["target_id"])) {
			errs = append(errs, "Security case missing from scope ledger: "+id)
		}
		caseReqs := []any{}
		for _, v := range arr(c["requirement_refs"]) {
			caseReqs = append(caseReqs, obj(v)["id"])
		}
		bounded(caseReqs, id)
		related := false
		for _, rid := range arr(check["requirement_ids"]) {
			if has(caseReqs, str(rid)) && has(target["requirement_ids"], str(rid)) {
				related = true
			}
		}
		if !related {
			errs = append(errs, "Security case lacks related target/requirement: "+id)
		}
		if !member(meta["action"], "observe", "request-modification", "scan", "load", "device-change") {
			errs = append(errs, "Security case needs action: "+id)
		}
		oracle := str(meta["oracle"])
		if !member(oracle, "confirmed", "unknown") {
			errs = append(errs, "Security oracle must be confirmed or unknown: "+id)
		}
		if oracle == "unknown" && (c["status"] == "ready" || !member(c["status"], "proposed", "deferred") || len(arr(c["expected"])) > 0) {
			errs = append(errs, "Unknown security expectation cannot be executable or fixed: "+id)
		}
		if c["status"] == "ready" && (oracle != "confirmed" || check["decision"] != "include") {
			errs = append(errs, "Ready security case needs included check and confirmed oracle: "+id)
		}
		if oracle == "confirmed" {
			if len(arr(meta["oracle_refs"])) == 0 {
				errs = append(errs, "Confirmed security expectation needs source: "+id)
			}
			for _, v := range arr(meta["oracle_refs"]) {
				ref := obj(v)
				r := reqs[str(ref["id"])]
				matched := false
				for _, cr := range arr(c["requirement_refs"]) {
					if obj(cr)["id"] == ref["id"] && num(obj(cr)["revision"]) == num(ref["revision"]) {
						matched = true
					}
				}
				if !matched || r["status"] != "active" || num(r["revision"]) != num(ref["revision"]) {
					errs = append(errs, "Security expectation uses missing, stale or unrelated requirement: "+id)
				}
			}
		}
	}
	if value, ok := s["execution"]; ok {
		ex := obj(value)
		if !member(ex["status"], "authorized", "pending") {
			errs = append(errs, "Security execution status must be authorized or pending")
		}
		if ex["status"] == "authorized" {
			for _, key := range []string{"target_ids", "allowed_actions", "stop_conditions"} {
				if !stringList(ex[key]) {
					errs = append(errs, "Security execution needs "+key)
				}
			}
			if !present(ex["budget"]) {
				errs = append(errs, "Security execution needs budget")
			}
			for _, tid := range arr(ex["target_ids"]) {
				if targets[str(tid)] == nil {
					errs = append(errs, "Security execution outside target scope: "+str(tid))
				}
			}
			if len(arr(ex["source_refs"])) == 0 {
				errs = append(errs, "Security execution needs authorization source")
			}
			for _, v := range arr(ex["source_refs"]) {
				ref := obj(v)
				if sources[str(ref["source_id"])]["status"] != "accepted" || !present(ref["locator"]) {
					errs = append(errs, "Security execution needs accepted authorization source")
				}
			}
		}
	}
	return dedup(errs)
}
func SecurityRecordAllowed(plan, c, observation M) error {
	if c["security"] == nil || !member(observation["status"], "passed", "failed") {
		return nil
	}
	meta := obj(c["security"])
	if obj(plan["security"])["mode"] != "specialist" && meta["action"] == "observe" {
		return nil
	}
	ex := obj(obj(plan["security"])["execution"])
	if ex["status"] != "authorized" || !has(ex["target_ids"], str(meta["target_id"])) || !has(ex["allowed_actions"], str(meta["action"])) {
		return fmt.Errorf("Security execution lacks confirmed target/action authorization; design readiness is not execution authorization")
	}
	return nil
}

// A changed target is a changed test context even if the stable target ID remains the same.
func SecurityTargetHistory(plan, oldCase, oldTarget M) []string {
	if oldCase["security"] == nil || len(oldTarget) == 0 {
		return nil
	}
	var current M
	for _, v := range arr(plan["cases"]) {
		if obj(v)["id"] == oldCase["id"] {
			current = obj(v)
			break
		}
	}
	if current == nil || current["status"] == "retired" || num(current["revision"]) > num(oldCase["revision"]) {
		return nil
	}
	meta := obj(current["security"])
	for _, v := range arr(obj(plan["security"])["targets"]) {
		target := obj(v)
		if target["id"] == meta["target_id"] && string(canonical(securityTargetContext(target))) != string(canonical(securityTargetContext(oldTarget))) {
			return []string{"Security target changed without case revision: " + str(current["id"])}
		}
	}
	return nil
}

func securityTargetContext(target M) M {
	m := obj(clone(target))
	delete(m, "requirement_ids")
	return m
}
