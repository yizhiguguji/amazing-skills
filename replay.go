package amazing

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type noteModel struct {
	Limit                    int
	Stay, Bug, CancelEnabled bool
	Draft, Saved, Status     string
	Writes                   int
	Events                   []any
}

func (m *noteModel) enter(v string) {
	m.Draft = v
	m.Status = "editing"
	m.Events = append(m.Events, M{"event": "enter", "value": v})
}
func (m *noteModel) begin() bool {
	if m.Status == "pending" {
		m.Events = append(m.Events, M{"event": "duplicate-submit", "accepted": false})
		return false
	}
	limit := m.Limit
	if m.Bug {
		limit++
	}
	accepted := strings.TrimSpace(m.Draft) != "" && (limit == 0 || utf8.RuneCountInString(m.Draft) <= limit)
	if accepted {
		m.Status = "pending"
	} else {
		m.Status = "invalid"
	}
	m.Events = append(m.Events, M{"event": "submit", "accepted": accepted})
	return accepted
}
func (m *noteModel) response(ok bool) {
	if m.Status != "pending" {
		return
	}
	if !ok {
		m.Status = "error"
	} else {
		m.Saved = m.Draft
		m.Writes++
		if m.Stay {
			m.Status = "saved"
		} else {
			m.Status = "list"
			m.Draft = ""
		}
	}
	m.Events = append(m.Events, M{"event": "response", "success": ok})
}
func (m *noteModel) observe() M {
	return M{"draft": m.Draft, "saved": m.Saved, "status": m.Status, "writes": m.Writes, "events": m.Events}
}
func ExecuteFixture(c M, stage int, bug bool) M {
	limit := 0
	if stage == 2 {
		limit = 20
	}
	if stage >= 3 {
		limit = 30
	}
	m := noteModel{Limit: limit, Stay: stage >= 2, Bug: bug, CancelEnabled: stage < 3, Status: "list", Events: []any{}}
	kind := str(c["fixture_kind"])
	failure := ""
	assert := func(ok bool, msg string) {
		if !ok && failure == "" {
			failure = msg
		}
	}
	switch kind {
	case "save", "min", "max", "over", "unicode", "empty", "spaces":
		n := limit
		if n == 0 {
			n = 1
		}
		values := map[string]string{"save": "记录下一步", "min": "中", "max": strings.Repeat("中", n), "over": strings.Repeat("中", n+1), "unicode": strings.Repeat("🙂", n), "empty": "", "spaces": "   "}
		value := values[kind]
		m.enter(value)
		accepted := m.begin()
		wanted := !member(kind, "over", "empty", "spaces")
		assert(accepted == wanted, "提交有效性不符合长度/空白规则")
		if failure == "" && wanted {
			m.response(true)
			assert(m.Saved == value && m.Writes == 1, "保存值或写入次数不符")
			if kind == "save" {
				status := "list"
				draft := ""
				if stage >= 2 {
					status = "saved"
					draft = value
				}
				assert(m.Status == status && m.Draft == draft, "保存后的页面/草稿不符")
			}
		} else if failure == "" {
			assert(m.Writes == 0 && m.Saved == "", "非法输入产生写入")
		}
	case "cancel":
		m.enter("未保存")
		if m.CancelEnabled {
			m.Status = "list"
			m.Draft = ""
		}
		m.Events = append(m.Events, M{"event": "cancel"})
		assert(m.Saved == "" && m.Status == "list", "取消后状态不符")
	case "offline":
		m.enter("保留")
		assert(m.begin(), "未进入提交")
		m.response(false)
		assert(m.Draft == "保留" && m.Writes == 0 && m.Status == "error", "失败时数据不符")
		assert(m.begin(), "无法手动重试")
		m.response(true)
		assert(m.Saved == "保留" && m.Writes == 1, "重试写入不符")
	case "delay":
		m.enter("等待")
		assert(m.begin(), "未进入提交")
		assert(!m.begin(), "重复提交未被阻止")
		assert(m.Writes == 0 && m.Status == "pending", "等待状态不符")
		m.response(true)
		assert(m.Writes == 1, "重复写入")
	case "app-switch", "background", "lock":
		m.enter("生命周期草稿")
		m.Events = append(m.Events, M{"event": kind, "scope": "same-process-model"})
		assert(m.Draft == "生命周期草稿" && m.Writes == 0, "模拟恢复后草稿变化")
	default:
		failure = "Unknown fixture kind"
	}
	status := "passed"
	var f any
	if failure != "" {
		status = "failed"
		f = failure
	}
	return M{"status": status, "failure": f, "observation": m.observe(), "scope": "Synthetic Go model only; no native app, OS event, radio network or temperature tested."}
}
func fixturePlans(root string) ([]M, error) {
	plans := []M{}
	for _, version := range []string{"r1", "r2", "r3"} {
		v, e := load(Bundle, "examples/qa-continuous-change/fixtures/"+version+".json")
		if e != nil {
			return nil, e
		}
		p := obj(v)
		for _, v := range arr(p["sources"]) {
			s := obj(v)
			h, e := fileHash(path(root, str(s["locator"])))
			if e != nil {
				return nil, e
			}
			s["sha256"] = h
		}
		plans = append(plans, p)
	}
	return plans, nil
}
func executePlan(root string, plan M, stage int, run string, bug bool, only string) error {
	for _, v := range arr(plan["cases"]) {
		c := obj(v)
		if c["status"] != "ready" || (only != "" && c["id"] != only) {
			continue
		}
		t := now()
		result := ExecuteFixture(c, stage, bug)
		id := run + "-" + str(c["id"])
		rel := "evidence/" + id + ".json"
		if e := write(path(root, rel), result, true); e != nil {
			return e
		}
		o := M{"execution_id": id, "run_id": run, "case_id": c["id"], "case_revision": c["revision"], "status": result["status"], "mode": "synthetic-fixture", "environment": "Go model, no device or external services", "actual": string(canonical(result["observation"])), "evidence": []any{rel}, "started_at": t, "finished_at": now()}
		if _, e := Record(root, o); e != nil {
			return e
		}
	}
	return nil
}
func mockDelivery(root string) error {
	d := M{"id": "FD-001", "summary": "超出长度上限仍进入提交态", "environment": "Synthetic Go fixture r2, configured boundary defect", "steps": []any{"打开详情输入21个Unicode码点", "点击保存并观察提交有效性"}, "actual": "模拟模型允许提交21个码点", "expected": "上限20，21应拒绝且不写入", "basis": "合成r2 AC-LIMIT；实际失败r2-bug-TC-007", "evidence": []any{"evidence/r2-bug-TC-007.json"}}
	if e := write(path(root, "evidence/defect-input.json"), d, true); e != nil {
		return e
	}
	draft, e := Draft(root, d)
	if e != nil {
		return e
	}
	store := map[string]M{}
	operations := []any{}
	create := func(id string) M {
		if store[id] == nil {
			store[id] = M{"issue": "MOCK-1", "body_hash": draft["payload_hash"]}
		}
		operations = append(operations, M{"action": "create-or-read", "operation_id": id, "issue": store[id]["issue"]})
		return store[id]
	}
	first := create("create-FD-001")
	receipt := M{"mode": "mock", "operation_id": "create-FD-001", "operation": "create", "issue_url": "mock://issues/MOCK-1", "write_outcome": "saved", "readback_confirmed": true, "expected_payload_hash": draft["payload_hash"], "readback_payload_hash": first["body_hash"], "evidence_entries": []any{M{"local": "evidence/r2-bug-TC-007.json", "uploaded": true, "visible": false, "remote_id": "mock-attachment-1"}}}
	incomplete := Receipt(root, receipt)
	if incomplete["receipt_check"] != "incomplete" {
		return fmt.Errorf("Incomplete media incorrectly accepted")
	}
	if e = write(path(root, "evidence/jira-partial-receipt.json"), receipt, true); e != nil {
		return e
	}
	if e = write(path(root, "evidence/jira-partial-check.json"), incomplete, true); e != nil {
		return e
	}
	retry := create("create-FD-001")
	if retry["issue"] != first["issue"] || len(store) != 1 {
		return fmt.Errorf("Duplicate mock issue")
	}
	obj(arr(receipt["evidence_entries"])[0])["visible"] = true
	complete := Receipt(root, receipt)
	if complete["receipt_check"] != "complete" || complete["real_jira_verified"] != false {
		return fmt.Errorf("Invalid completed mock receipt")
	}
	if e = write(path(root, "evidence/jira-final-receipt.json"), receipt, true); e != nil {
		return e
	}
	if e = write(path(root, "evidence/jira-final-check.json"), complete, true); e != nil {
		return e
	}
	return write(path(root, "evidence/mock-jira-operations.json"), M{"mode": "mock", "operations": operations, "issue_count": len(store), "attachment_count": 1}, true)
}
func Replay(root string) (M, error) {
	var final M
	e := safeCall(func() error {
		if exists(path(root, ".qa")) {
			return fmt.Errorf("Refusing to overwrite previous evaluation; use a fresh output directory")
		}
		if e := fs.WalkDir(Bundle, "examples/qa-continuous-change/sources", func(p string, d fs.DirEntry, e error) error {
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
			return writeBytes(path(root, "sources/"+filepath.Base(p)), b, true)
		}); e != nil {
			return e
		}
		if _, e := QAInit(root); e != nil {
			return e
		}
		config := obj(mustRead(path(root, ".qa/config.json")))
		config["project_id"] = "synthetic-note-feature"
		config["source_policy"] = "Explicit r2/r3 decisions supersede matching r1 behavior"
		config["review_policy"] = "No real approval; model-authored evaluation only."
		if e := write(path(root, ".qa/config.json"), config, false); e != nil {
			return e
		}
		plans, e := fixturePlans(root)
		if e != nil {
			return e
		}
		for i, p := range plans {
			label := fmt.Sprintf("r%d", i+1)
			if e = write(path(root, ".qa/plan.json"), p, false); e != nil {
				return e
			}
			if i > 0 {
				impact, e := Impact(root, fmt.Sprintf("r%d", i))
				if e != nil {
					return e
				}
				if e = write(path(root, "evidence/impact-"+fmt.Sprintf("r%d", i)+"-"+label+".json"), impact, true); e != nil {
					return e
				}
				before, e := Summary(root, "")
				if e != nil {
					return e
				}
				if e = write(path(root, "evidence/"+label+"-before-execution.json"), before, true); e != nil {
					return e
				}
				if i == 1 {
					for _, v := range arr(before["cases"]) {
						c := obj(v)
						if c["case_id"] == "TC-001" && c["status"] != "not-run" {
							return fmt.Errorf("Changed assertion inherited old pass")
						}
					}
				}
				if i == 2 && num(obj(before["counts"])["not-run"]) != 4 {
					return fmt.Errorf("Changed revision not isolated")
				}
			}
			if _, e = Snapshot(root, label); e != nil {
				return e
			}
			if _, e = ExportCases(root); e != nil {
				return e
			}
			if i == 1 {
				if e = executePlan(root, p, 2, "r2-bug", true, ""); e != nil {
					return e
				}
				if e = executePlan(root, p, 2, "r2-retry", false, "TC-007"); e != nil {
					return e
				}
				events, e := ReadEvents(root)
				if e != nil {
					return e
				}
				obs := obj(clone(events[len(events)-1]["observation"]))
				obs["run_id"] = "r2-bug"
				obs["execution_id"] = "r2-bug-retry-TC-007"
				if _, e = Record(root, obs); e != nil {
					return e
				}
				same, e := Summary(root, "r2-bug")
				if e != nil {
					return e
				}
				if num(obj(same["counts"])["failed"]) != 1 {
					return fmt.Errorf("Same-round failure overwritten")
				}
				if e = write(path(root, "evidence/r2-first-failure-summary.json"), same, true); e != nil {
					return e
				}
				if e = write(path(root, ".qa/defects.json"), []any{M{"id": "FD-001", "classification": "confirmed-bug", "status": "open", "found_execution_id": "r2-bug-TC-007"}}, false); e != nil {
					return e
				}
				if e = mockDelivery(root); e != nil {
					return e
				}
				if e = executePlan(root, p, 2, "r2-fixed", false, ""); e != nil {
					return e
				}
				if e = write(path(root, ".qa/defects.json"), []any{M{"id": "FD-001", "classification": "confirmed-bug", "status": "resolved", "found_execution_id": "r2-bug-TC-007", "resolution_execution_id": "r2-fixed-TC-007"}}, false); e != nil {
					return e
				}
				r, e := Summary(root, "r2-fixed")
				if e != nil {
					return e
				}
				if e = write(path(root, "evidence/r2-fixed-summary.json"), r, true); e != nil {
					return e
				}
			} else {
				if e = executePlan(root, p, i+1, label, false, ""); e != nil {
					return e
				}
				final, e = Summary(root, label)
				if e != nil {
					return e
				}
				if e = write(path(root, "evidence/"+label+"-summary.json"), final, true); e != nil {
					return e
				}
			}
		}
		if num(final["total"]) != 12 || num(obj(final["counts"])["failed"]) != 0 || final["quality_status"] != "synthetic-only" || len(arr(final["historical_failures"])) != 1 {
			return fmt.Errorf("Invalid final synthetic report")
		}
		if e = requireQA(root); e != nil {
			return e
		}
		return write(path(root, "evidence/replay-results.json"), M{"status": "passed", "implementation": "Go", "assertions": []string{"Unclear rules remain proposed", "Changed expectation has no inherited pass", "Same-round first failure preserved", "Mock media failure incomplete", "Mock retry one issue", "Withdrawn case retained outside denominator", "Four changed cases require new execution", "Report remains synthetic-only"}, "real_device_verified": false, "real_jira_submitted": false, "final_counts": final["counts"]}, true)
	})
	if e != nil {
		return nil, e
	}
	return M{"workflow_replay": "passed", "quality_status": final["quality_status"], "cases": final["total"], "output": root}, nil
}
