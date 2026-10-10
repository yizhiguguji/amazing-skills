package amazing

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var Dimensions = []string{"interaction", "happy", "negative", "boundary", "state", "app-switch", "background", "lock", "offline", "weak-network", "thermal"}
var Statuses = []string{"passed", "failed", "blocked", "deferred", "not-run"}
var Modes = []string{"synthetic-fixture", "mock", "browser-emulation", "real-device", "manual"}

func QAInit(root string) (M, error) {
	created := []string{}
	e := safeCall(func() error {
		config, e := load(Bundle, "skills/qa-test-design/assets/config.json")
		if e != nil {
			return e
		}
		files := map[string]any{"config.json": config, "plan.json": M{"schema_version": 1, "sources": []any{}, "requirements": []any{}, "cases": []any{}, "dimensions": []any{}, "gaps": []any{}}}
		for name, v := range files {
			p := path(root, ".qa/"+name)
			if !exists(p) {
				if e = write(p, v, true); e != nil {
					return e
				}
				created = append(created, name)
			}
		}
		return nil
	})
	return M{"created": created, "note": "Draft only; no test execution or approval."}, e
}
func QACheck(root string) []string {
	errs := []string{}
	e := safeCall(func() error {
		plan := obj(mustRead(path(root, ".qa/plan.json")))
		config := obj(mustRead(path(root, ".qa/config.json")))
		if num(plan["schema_version"]) != 1 || num(config["schema_version"]) != 1 {
			errs = append(errs, "Unsupported schema")
		}
		sources := index(plan["sources"], "sources", &errs)
		reqs := index(plan["requirements"], "requirements", &errs)
		cases := index(plan["cases"], "cases", &errs)
		dims := index(plan["dimensions"], "dimensions", &errs)
		gaps := index(plan["gaps"], "gaps", &errs)
		if len(reqs) == 0 || len(cases) == 0 {
			errs = append(errs, "Plan needs requirements and cases")
		}
		if len(dims) != len(Dimensions) {
			errs = append(errs, "Account for each functional dimension")
		}
		for _, d := range Dimensions {
			if _, ok := dims[d]; !ok {
				errs = append(errs, "Missing functional dimension: "+d)
			}
		}
		for sid, s := range sources {
			if !member(s["status"], "accepted", "unverified", "withdrawn") {
				errs = append(errs, "Source status: "+sid)
			}
			for _, k := range []string{"version", "kind", "locator"} {
				if !present(s[k]) {
					errs = append(errs, "Source field: "+sid+"/"+k)
				}
			}
			loc := str(s["locator"])
			if !strings.HasPrefix(loc, "http://") && !strings.HasPrefix(loc, "https://") {
				h, e := fileHash(path(root, loc))
				if e != nil || h != str(s["sha256"]) {
					errs = append(errs, "Source missing or changed: "+sid)
				}
			}
		}
		for rid, r := range reqs {
			if !positiveRevision(r["revision"]) {
				errs = append(errs, "Requirement revision: "+rid)
			}
			if !member(r["status"], "active", "uncertain", "superseded", "withdrawn") {
				errs = append(errs, "Requirement status: "+rid)
			}
			if !present(r["text"]) {
				errs = append(errs, "Requirement text: "+rid)
			}
			refs := arr(r["source_refs"])
			if len(refs) == 0 {
				errs = append(errs, "Requirement source: "+rid)
			}
			for _, v := range refs {
				ref := obj(v)
				s := sources[str(ref["source_id"])]
				if s == nil || !present(ref["locator"]) {
					errs = append(errs, "Missing source locator: "+rid)
				} else if r["status"] == "active" && s["status"] != "accepted" {
					errs = append(errs, "Active requirement uses unaccepted source: "+rid)
				}
			}
		}
		for cid, c := range cases {
			if !positiveRevision(c["revision"]) {
				errs = append(errs, "Case revision: "+cid)
			}
			if !member(c["status"], "ready", "proposed", "deferred", "retired") {
				errs = append(errs, "Case status: "+cid)
			}
			for _, k := range []string{"objective", "module", "subfunction", "category", "impact", "method"} {
				if !present(c[k]) {
					errs = append(errs, "Case field: "+cid+"/"+k)
				}
			}
			if !stringList(c["dimensions"]) {
				errs = append(errs, "Case dimensions: "+cid)
			}
			for _, d := range arr(c["dimensions"]) {
				if !member(d, Dimensions...) {
					errs = append(errs, "Case dimensions: "+cid)
				}
			}
			refs := arr(c["requirement_refs"])
			if len(refs) == 0 {
				errs = append(errs, "Case requirement trace: "+cid)
			}
			for _, v := range refs {
				ref := obj(v)
				r := reqs[str(ref["id"])]
				if r == nil {
					errs = append(errs, "Unknown requirement: "+cid)
				} else if c["status"] == "ready" && (r["status"] != "active" || num(r["revision"]) != num(ref["revision"])) {
					errs = append(errs, "Ready case uses inactive or stale requirement: "+cid)
				}
			}
			if c["status"] == "ready" {
				for _, k := range []string{"preconditions", "steps", "expected"} {
					if !stringList(c[k]) {
						errs = append(errs, "Executable case missing "+k+": "+cid)
					}
				}
				if !present(c["data"]) {
					errs = append(errs, "Case data: "+cid)
				}
				for _, v := range arr(c["dimensions"]) {
					d := str(v)
					if dims[d]["decision"] != "include" {
						errs = append(errs, "Ready case in unselected dimension: "+cid+"/"+d)
					} else if !has(dims[d]["case_ids"], cid) {
						errs = append(errs, "Ready case missing from dimension ledger: "+cid+"/"+d)
					}
				}
			}
		}
		for rid, r := range reqs {
			if r["status"] != "active" {
				continue
			}
			covered := false
			for _, c := range cases {
				if member(c["status"], "ready", "proposed", "deferred") {
					for _, ref := range arr(c["requirement_refs"]) {
						if obj(ref)["id"] == rid {
							covered = true
						}
					}
				}
			}
			for _, g := range gaps {
				if has(g["requirement_ids"], rid) {
					covered = true
				}
			}
			if !covered {
				errs = append(errs, "Active requirement has no case or explicit coverage gap: "+rid)
			}
		}
		for did, d := range dims {
			if !member(d["decision"], "include", "not-applicable", "clarify", "deferred") || !present(d["reason"]) {
				errs = append(errs, "Dimension decision/reason: "+did)
			}
			if d["decision"] == "include" {
				if len(arr(d["requirement_ids"])) == 0 {
					errs = append(errs, "Dimension needs active scope: "+did)
				}
				for _, id := range arr(d["requirement_ids"]) {
					if reqs[str(id)]["status"] != "active" {
						errs = append(errs, "Dimension needs active scope: "+did)
					}
				}
				if len(arr(d["case_ids"])) == 0 {
					errs = append(errs, "Dimension needs matching cases: "+did)
				}
				for _, id := range arr(d["case_ids"]) {
					c := cases[str(id)]
					if !member(c["status"], "ready", "proposed", "deferred") || !has(c["dimensions"], did) {
						errs = append(errs, "Dimension needs matching cases: "+did)
					}
				}
			}
			if member(d["decision"], "clarify", "deferred") {
				if len(arr(d["gap_ids"])) == 0 {
					errs = append(errs, "Dimension needs gap: "+did)
				}
				for _, id := range arr(d["gap_ids"]) {
					if gaps[str(id)] == nil {
						errs = append(errs, "Dimension needs gap: "+did)
					}
				}
			}
		}
		for gid, g := range gaps {
			if !present(g["question"]) || !present(g["close_condition"]) {
				errs = append(errs, "Gap missing question/close condition: "+gid)
			}
		}
		history := []any{}
		reqHistory := []any{}
		snapshotRoot := path(root, ".qa/snapshots")
		if exists(snapshotRoot) {
			folders, e := os.ReadDir(snapshotRoot)
			if e != nil {
				return e
			}
			for _, folder := range folders {
				if !folder.IsDir() {
					continue
				}
				base := ".qa/snapshots/" + folder.Name() + "/"
				m := obj(mustRead(path(root, base+"manifest.json")))
				for rel, h := range obj(m["files"]) {
					actual, e := fileHash(path(root, base+rel))
					if e != nil || actual != str(h) {
						errs = append(errs, "Frozen snapshot drift: "+folder.Name()+"/"+rel)
					}
				}
				p := obj(mustRead(path(root, base+"plan.json")))
				history = append(history, arr(p["cases"])...)
				reqHistory = append(reqHistory, arr(p["requirements"])...)
			}
		}
		events, e := ReadEvents(root)
		if e != nil {
			return e
		}
		observations := map[string]M{}
		for _, event := range events {
			history = append(history, event["case_snapshot"])
			observations[str(event["execution_id"])] = event
		}
		dp := path(root, ".qa/defects.json")
		if exists(dp) {
			defects := index(mustRead(dp), "defects", &errs)
			for did, d := range defects {
				if !member(d["classification"], "confirmed-bug", "question", "suggestion", "coverage-gap") {
					errs = append(errs, "Defect classification: "+did)
				}
				if !member(d["status"], "open", "resolved", "withdrawn") {
					errs = append(errs, "Defect status: "+did)
				}
				if d["classification"] == "confirmed-bug" {
					found := observations[str(d["found_execution_id"])]
					if found["status"] != "failed" {
						errs = append(errs, "Confirmed bug needs recorded failure: "+did)
					}
					if d["status"] == "resolved" {
						fixed := observations[str(d["resolution_execution_id"])]
						if fixed["status"] != "passed" || found == nil || fixed["case_id"] != found["case_id"] {
							errs = append(errs, "Resolved bug needs corresponding retest: "+did)
						}
					}
				}
			}
		}
		for _, group := range []struct {
			current map[string]M
			history []any
			label   string
		}{{cases, history, "Case"}, {reqs, reqHistory, "Requirement"}} {
			for _, v := range group.history {
				old := obj(v)
				id := str(old["id"])
				cur := group.current[id]
				if cur == nil {
					errs = append(errs, "Frozen "+strings.ToLower(group.label)+" removed; retain as retired/withdrawn")
					continue
				}
				if num(cur["revision"]) < num(old["revision"]) {
					errs = append(errs, group.label+" revision moved backwards: "+id)
				}
				if num(cur["revision"]) == num(old["revision"]) && string(canonical(withoutStatus(cur))) != string(canonical(withoutStatus(old))) {
					errs = append(errs, group.label+" changed without revision: "+id)
				}
			}
		}
		return nil
	})
	if e != nil {
		errs = append(errs, e.Error())
	}
	return dedup(errs)
}
func positiveRevision(v any) bool { return num(v) > 0 && string(canonical(v)) == fmt.Sprint(num(v)) }
func requireQA(root string) error {
	if errs := QACheck(root); len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

var identityPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,63}$`)

func Snapshot(root, label string) (M, error) {
	result := M{}
	e := safeCall(func() error {
		if !identityPattern.MatchString(label) {
			return fmt.Errorf("Invalid snapshot label")
		}
		if e := requireQA(root); e != nil {
			return e
		}
		folder := path(root, ".qa/snapshots/"+label)
		if exists(folder) {
			return fmt.Errorf("Snapshot exists; refusing replacement")
		}
		if e := os.MkdirAll(filepath.Dir(folder), 0755); e != nil {
			return e
		}
		if e := os.Mkdir(folder, 0755); e != nil {
			return e
		}
		files := M{}
		for _, name := range []string{"plan.json", "config.json"} {
			b, e := os.ReadFile(path(root, ".qa/"+name))
			if e != nil {
				return e
			}
			if e = writeBytes(filepath.Join(folder, name), b, true); e != nil {
				return e
			}
			files[name] = hash(b)
		}
		p := obj(mustRead(filepath.Join(folder, "plan.json")))
		for _, v := range arr(p["sources"]) {
			s := obj(v)
			loc := str(s["locator"])
			if strings.HasPrefix(loc, "http://") || strings.HasPrefix(loc, "https://") {
				continue
			}
			name := "sources/" + hash([]byte(str(s["id"]))) + ".bin"
			b, e := os.ReadFile(path(root, loc))
			if e != nil {
				return e
			}
			if e = writeBytes(filepath.Join(folder, name), b, true); e != nil {
				return e
			}
			files[name] = hash(b)
		}
		if e := write(filepath.Join(folder, "manifest.json"), M{"created_at": now(), "files": files}, true); e != nil {
			return e
		}
		result = M{"snapshot": ".qa/snapshots/" + label, "approval": "not implied"}
		return nil
	})
	return result, e
}
func Impact(root, before string) (M, error) {
	result := M{}
	e := safeCall(func() error {
		if !identityPattern.MatchString(before) {
			return fmt.Errorf("Invalid snapshot label")
		}
		old := obj(mustRead(path(root, ".qa/snapshots/"+before+"/plan.json")))
		cur := obj(mustRead(path(root, ".qa/plan.json")))
		errs := []string{}
		a := index(old["requirements"], "requirements", &errs)
		b := index(cur["requirements"], "requirements", &errs)
		sa := index(old["sources"], "sources", &errs)
		sb := index(cur["sources"], "sources", &errs)
		changedSources := []string{}
		changed := []string{}
		for id, s := range sa {
			if string(canonical(s)) != string(canonical(sb[id])) {
				changedSources = append(changedSources, id)
			}
		}
		for id, s := range sb {
			if string(canonical(s)) != string(canonical(sa[id])) {
				changedSources = append(changedSources, id)
			}
		}
		ids := map[string]bool{}
		for id := range a {
			ids[id] = true
		}
		for id := range b {
			ids[id] = true
		}
		for id := range ids {
			different := string(canonical(a[id])) != string(canonical(b[id]))
			for _, r := range append(arr(a[id]["source_refs"]), arr(b[id]["source_refs"])...) {
				if member(obj(r)["source_id"], changedSources...) {
					different = true
				}
			}
			if different {
				changed = append(changed, id)
			}
		}
		candidates := []string{}
		for _, c := range append(arr(old["cases"]), arr(cur["cases"])...) {
			for _, r := range arr(obj(c)["requirement_refs"]) {
				if member(obj(r)["id"], changed...) {
					candidates = append(candidates, str(obj(c)["id"]))
				}
			}
		}
		unchanged := []string{}
		for _, c := range arr(cur["cases"]) {
			id := str(obj(c)["id"])
			if !member(id, candidates...) {
				unchanged = append(unchanged, id)
			}
		}
		result = M{"changed_sources": dedup(changedSources), "changed_requirements": dedup(changed), "candidate_cases": dedup(candidates), "unchanged_cases": dedup(unchanged), "note": "Deterministic candidates only; review source, UI and dependency changes separately."}
		return nil
	})
	return result, e
}
func ReadEvents(root string) ([]M, error) {
	events := []M{}
	err := safeCall(func() error {
		p := path(root, ".qa/executions.jsonl")
		previous := ""
		ids := map[string]bool{}
		if exists(p) {
			f, e := os.Open(p)
			if e != nil {
				return e
			}
			defer f.Close()
			scanner := bufio.NewScanner(f)
			scanner.Buffer(make([]byte, 4096), 16*1024*1024)
			for scanner.Scan() {
				v, e := decode(scanner.Bytes())
				if e != nil {
					return e
				}
				m := obj(v)
				body := obj(clone(m))
				delete(body, "hash")
				h := str(m["hash"])
				if m["previous_hash"] != previous || hash(canonical(body)) != h {
					return fmt.Errorf("Execution history drift")
				}
				id := str(m["execution_id"])
				if ids[id] {
					return fmt.Errorf("Duplicate execution identity")
				}
				ids[id] = true
				previous = h
				events = append(events, m)
			}
			if e = scanner.Err(); e != nil {
				return e
			}
		}
		cp := path(root, ".qa/execution-checkpoint.json")
		if exists(cp) {
			checkpoint := obj(mustRead(cp))
			if num(checkpoint["count"]) != len(events) || checkpoint["hash"] != previous || len(checkpoint) != 2 {
				return fmt.Errorf("Execution checkpoint mismatch")
			}
		} else if len(events) > 0 {
			return fmt.Errorf("Missing execution checkpoint")
		}
		for _, ev := range events {
			for _, v := range arr(ev["evidence"]) {
				rel := str(v)
				h, e := fileHash(path(root, rel))
				if e != nil || h != str(obj(ev["evidence_hashes"])[rel]) {
					return fmt.Errorf("Execution evidence drift")
				}
			}
		}
		return nil
	})
	return events, err
}
func Record(root string, observation M) (M, error) {
	result := M{}
	e := safeCall(func() error {
		if e := requireQA(root); e != nil {
			return e
		}
		plan := obj(mustRead(path(root, ".qa/plan.json")))
		var c M
		for _, v := range arr(plan["cases"]) {
			if obj(v)["id"] == observation["case_id"] {
				c = obj(v)
			}
		}
		if c == nil || c["status"] != "ready" || num(c["revision"]) != num(observation["case_revision"]) {
			return fmt.Errorf("Execution needs a ready current case revision")
		}
		for _, key := range []string{"execution_id", "run_id", "environment", "actual"} {
			if !present(observation[key]) {
				return fmt.Errorf("Execution missing %s", key)
			}
		}
		if !member(observation["status"], Statuses...) || !member(observation["mode"], Modes...) {
			return fmt.Errorf("Invalid execution status/mode")
		}
		start, e := time.Parse(time.RFC3339Nano, str(observation["started_at"]))
		if e != nil {
			return fmt.Errorf("Invalid execution start/timezone")
		}
		end, e := time.Parse(time.RFC3339Nano, str(observation["finished_at"]))
		if e != nil || end.Before(start) {
			return fmt.Errorf("Execution times need timezone and valid ordering")
		}
		evidence, ok := observation["evidence"].([]any)
		if !ok {
			return fmt.Errorf("Invalid evidence list")
		}
		if member(observation["status"], "passed", "failed") && len(evidence) == 0 {
			return fmt.Errorf("Passed/failed execution needs evidence")
		}
		digests := M{}
		for _, v := range evidence {
			if !present(v) {
				return fmt.Errorf("Invalid evidence list")
			}
			h, e := fileHash(path(root, str(v)))
			if e != nil {
				return e
			}
			digests[str(v)] = h
		}
		events, e := ReadEvents(root)
		if e != nil {
			return e
		}
		for _, ev := range events {
			if ev["execution_id"] == observation["execution_id"] {
				if string(canonical(ev["observation"])) != string(canonical(observation)) {
					return fmt.Errorf("Execution identity collision; do not overwrite history")
				}
				result = M{"recorded": false, "idempotent": true}
				return nil
			}
		}
		ev := obj(clone(observation))
		ev["observation"] = clone(observation)
		ev["case_snapshot"] = c
		ev["evidence_hashes"] = digests
		previous := ""
		if len(events) > 0 {
			previous = str(events[len(events)-1]["hash"])
		}
		ev["previous_hash"] = previous
		ev["hash"] = hash(canonical(ev))
		p := path(root, ".qa/executions.jsonl")
		if e = os.MkdirAll(filepath.Dir(p), 0755); e != nil {
			return e
		}
		f, e := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if e != nil {
			return e
		}
		_, e = f.Write(append(canonical(ev), '\n'))
		ce := f.Close()
		if e != nil {
			return e
		}
		if ce != nil {
			return ce
		}
		if e = write(path(root, ".qa/execution-checkpoint.json"), M{"count": len(events) + 1, "hash": ev["hash"]}, false); e != nil {
			return e
		}
		result = M{"recorded": true, "execution_id": ev["execution_id"], "mode": ev["mode"]}
		return nil
	})
	return result, e
}
func Summary(root, run string) (M, error) {
	out := M{}
	e := safeCall(func() error {
		if e := requireQA(root); e != nil {
			return e
		}
		plan := obj(mustRead(path(root, ".qa/plan.json")))
		events, e := ReadEvents(root)
		if e != nil {
			return e
		}
		counts := M{}
		for _, s := range Statuses {
			counts[s] = 0
		}
		details := []any{}
		modes := M{}
		proposed := []string{}
		retired := []string{}
		for _, v := range arr(plan["cases"]) {
			c := obj(v)
			if c["status"] == "proposed" {
				proposed = append(proposed, str(c["id"]))
			}
			if c["status"] == "retired" {
				retired = append(retired, str(c["id"]))
			}
			if !member(c["status"], "ready", "deferred") {
				continue
			}
			formal := []M{}
			for _, ev := range events {
				if ev["case_id"] == c["id"] && num(ev["case_revision"]) == num(c["revision"]) && (run == "" || ev["run_id"] == run) && ev["status"] != "not-run" {
					formal = append(formal, ev)
				}
			}
			var selected M
			if len(formal) > 0 {
				selected = formal[len(formal)-1]
				for _, ev := range formal {
					if ev["run_id"] == selected["run_id"] && ev["status"] == "failed" {
						selected = ev
						break
					}
				}
			}
			status := "not-run"
			if c["status"] == "deferred" {
				status = "deferred"
			}
			var eid, mode any
			if selected != nil {
				status = str(selected["status"])
				eid = selected["execution_id"]
				mode = selected["mode"]
				modes[str(mode)] = num(modes[str(mode)]) + 1
			}
			counts[status] = num(counts[status]) + 1
			details = append(details, M{"case_id": c["id"], "revision": c["revision"], "status": status, "execution_id": eid, "mode": mode})
		}
		denominator := num(counts["passed"]) + num(counts["failed"])
		var rate any
		if denominator > 0 {
			rate = float64(num(counts["passed"])) / float64(denominator)
		}
		total := 0
		for _, v := range counts {
			total += num(v)
		}
		synthetic := len(modes) > 0
		for mode := range modes {
			if !member(mode, "synthetic-fixture", "mock") {
				synthetic = false
			}
		}
		quality := "recorded-results-require-scope-review"
		if denominator == 0 {
			quality = "insufficient-evidence"
		}
		if synthetic {
			quality = "synthetic-only"
		}
		scope := run
		if scope == "" {
			scope = "latest-per-current-revision"
		}
		failures := []string{}
		for _, ev := range events {
			if ev["status"] == "failed" {
				failures = append(failures, str(ev["execution_id"]))
			}
		}
		out = M{"schema_version": 1, "generated_at": now(), "plan_sha256": hash(canonical(plan)), "run": scope, "counts": counts, "total": total, "executed_definition": "passed + failed", "pass_rate": rate, "modes": modes, "quality_status": quality, "release_decision": "unverified; never inferred from counts alone", "cases": details, "historical_failures": failures, "proposed_cases": proposed, "retired_cases": retired, "gaps": plan["gaps"]}
		dp := path(root, ".qa/defects.json")
		defects := M{"source": "not-provided", "confirmed_total": nil, "current_open": nil}
		if exists(dp) {
			ds := arr(mustRead(dp))
			n, open := 0, 0
			for _, v := range ds {
				d := obj(v)
				if d["classification"] == "confirmed-bug" {
					n++
					if d["status"] == "open" {
						open++
					}
				}
			}
			defects = M{"source": "saved-local-records", "confirmed_total": n, "current_open": open}
		}
		out["defect_summary"] = defects
		if e = write(path(root, ".qa/report.json"), out, false); e != nil {
			return e
		}
		lines := []string{"# QA result summary", "", "Status: " + quality, "Scope: " + scope, "Release decision: " + str(out["release_decision"]), "", "| Status | Count |", "| --- | --- |"}
		for _, s := range Statuses {
			lines = append(lines, fmt.Sprintf("| %s | %d |", s, num(counts[s])))
		}
		lines = append(lines, "", fmt.Sprintf("Total: %d", total), fmt.Sprintf("Pass-rate denominator: passed + failed; rate: %v", rate), "Evidence modes: "+string(canonical(modes)), "Historical failures: "+strings.Join(failures, ", "), "Defects (separate from failed cases): "+string(canonical(defects)), "", "## Current cases", "")
		for _, v := range details {
			c := obj(v)
			lines = append(lines, fmt.Sprintf("- %s@%d: %s (%v)", str(c["case_id"]), num(c["revision"]), str(c["status"]), c["mode"]))
		}
		lines = append(lines, "", "## Gaps", "")
		for _, v := range arr(plan["gaps"]) {
			g := obj(v)
			lines = append(lines, "- "+str(g["id"])+": "+str(g["question"]))
		}
		return writeBytes(path(root, ".qa/report.md"), []byte(strings.Join(lines, "\n")+"\n"), false)
	})
	return out, e
}
func cell(v any) string {
	var s string
	if a, ok := v.([]any); ok {
		parts := []string{}
		for _, x := range a {
			parts = append(parts, fmt.Sprint(x))
		}
		s = strings.Join(parts, "；")
	} else {
		s = fmt.Sprint(v)
	}
	return strings.ReplaceAll(strings.ReplaceAll(s, "|", `\|`), "\n", "<br>")
}
func ExportCases(root string) (M, error) {
	e := safeCall(func() error {
		if e := requireQA(root); e != nil {
			return e
		}
		p := obj(mustRead(path(root, ".qa/plan.json")))
		errs := []string{}
		reqs := index(p["requirements"], "requirements", &errs)
		lines := []string{"# 当前功能测试用例", "", "此稿是设计产物，不是审批或执行通过。proposed/deferred/retired单列状态，只有ready可承诺执行。", "", "| ID@修订 | 状态 | 目标 | 模块/子功能 | 方法/影响 | 前置/数据 | 步骤 | 预期 | 需求与依据 |", "| --- | --- | --- | --- | --- | --- | --- | --- | --- |"}
		for _, v := range arr(p["cases"]) {
			c := obj(v)
			refs := []any{}
			for _, r := range arr(c["requirement_refs"]) {
				ref := obj(r)
				src := []string{}
				for _, s := range arr(reqs[str(ref["id"])]["source_refs"]) {
					m := obj(s)
					src = append(src, str(m["source_id"])+"/"+str(m["locator"]))
				}
				refs = append(refs, fmt.Sprintf("%s@%d: %s", str(ref["id"]), num(ref["revision"]), strings.Join(src, ", ")))
			}
			values := []any{fmt.Sprintf("%s@%d", str(c["id"]), num(c["revision"])), c["status"], c["objective"], str(c["module"]) + "/" + str(c["subfunction"]), str(c["method"]) + "/" + str(c["impact"]), append(arr(c["preconditions"]), c["data"]), c["steps"], c["expected"], refs}
			cells := []string{}
			for _, v := range values {
				cells = append(cells, cell(v))
			}
			lines = append(lines, "| "+strings.Join(cells, " | ")+" |")
		}
		if e := writeBytes(path(root, ".qa/cases.md"), []byte(strings.Join(lines, "\n")+"\n"), false); e != nil {
			return e
		}
		lines = []string{"# 范围、维度与追踪", "", "| 验收点@修订 | 状态 | 用例 |", "| --- | --- | --- |"}
		for _, v := range arr(p["requirements"]) {
			r := obj(v)
			ids := []any{}
			for _, v := range arr(p["cases"]) {
				c := obj(v)
				for _, ref := range arr(c["requirement_refs"]) {
					if obj(ref)["id"] == r["id"] {
						ids = append(ids, fmt.Sprintf("%s@%d:%s", str(c["id"]), num(c["revision"]), str(c["status"])))
						break
					}
				}
			}
			lines = append(lines, fmt.Sprintf("| %s@%d | %s | %s |", cell(r["id"]), num(r["revision"]), cell(r["status"]), cell(ids)))
		}
		lines = append(lines, "", "| 维度 | 决策 | 理由 | 场景/缺口 |", "| --- | --- | --- | --- |")
		for _, v := range arr(p["dimensions"]) {
			d := obj(v)
			lines = append(lines, "| "+cell(d["id"])+" | "+cell(d["decision"])+" | "+cell(d["reason"])+" | "+cell(append(arr(d["case_ids"]), arr(d["gap_ids"])...))+" |")
		}
		lines = append(lines, "", "## 待明确", "")
		for _, v := range arr(p["gaps"]) {
			g := obj(v)
			lines = append(lines, "- "+str(g["id"])+": "+str(g["question"])+" 关闭条件："+str(g["close_condition"]))
		}
		return writeBytes(path(root, ".qa/coverage.md"), []byte(strings.Join(lines, "\n")+"\n"), false)
	})
	return M{"exports": []string{".qa/cases.md", ".qa/coverage.md"}, "execution_verified": false}, e
}
func osFS(root string) fs.FS { return os.DirFS(root) }
