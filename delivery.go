package amazing

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func Draft(root string, d M) (M, error) {
	result := M{}
	e := safeCall(func() error {
		errs := []string{}
		for _, key := range []string{"id", "summary", "environment", "actual", "basis"} {
			if !present(d[key]) {
				errs = append(errs, "Missing "+key)
			}
		}
		if !identityPattern.MatchString(str(d["id"])) {
			errs = append(errs, "Invalid defect identity")
		}
		if !stringList(d["steps"]) {
			errs = append(errs, "Missing reproducible steps")
		}
		if !present(d["expected"]) && !stringList(d["expected"]) {
			errs = append(errs, "Expected a nonempty expected string or string list")
		}
		if observations, ok := d["observations"]; ok && !stringList(observations) {
			errs = append(errs, "observations must be a nonempty string list when provided")
		}
		evidence, ok := d["evidence"].([]any)
		if !ok {
			errs = append(errs, "Expected evidence list")
		}
		for _, rel := range evidence {
			if !isFile(path(root, str(rel))) {
				errs = append(errs, "Missing evidence")
			}
		}
		if len(errs) > 0 {
			return fmt.Errorf("%s", strings.Join(errs, "; "))
		}
		content := renderDraft(d, evidence, false)
		format := "current"
		rel := ".qa/defect-drafts/" + str(d["id"]) + ".md"
		p := path(root, rel)
		if exists(p) {
			b, e := os.ReadFile(p)
			if e != nil {
				return e
			}
			if string(b) != content {
				// Recognize only the exact previous rendering of this same input.
				// Never normalize arbitrary existing bodies or change receipt hashes.
				if present(d["expected"]) && d["observations"] == nil && string(b) == renderDraft(d, evidence, true) {
					content = string(b)
					format = "legacy-preserved"
				} else {
					return fmt.Errorf("Draft exists with different content; retain revision instead of overwrite")
				}
			}
		} else if e := writeBytes(p, []byte(content), true); e != nil {
			return e
		}
		evidenceStatus := "local-only"
		if len(evidence) == 0 {
			evidenceStatus = "not-attached"
		}
		result = M{"draft": rel, "remote_submission": "not performed", "payload_hash": hash([]byte(content)), "format": format, "evidence_count": len(evidence), "evidence_status": evidenceStatus}
		return nil
	})
	return result, e
}

func draftList(values []any) string {
	lines := make([]string, 0, len(values))
	for i, value := range values {
		// Indent continuation lines so multiline assertions stay in their item.
		prefix := fmt.Sprintf("%d. ", i+1)
		text := strings.ReplaceAll(strings.ReplaceAll(str(value), "\r\n", "\n"), "\n", "\n"+strings.Repeat(" ", len(prefix)))
		lines = append(lines, prefix+text)
	}
	return strings.Join(lines, "\n")
}

// legacy is solely an exact-match recognizer for immutable, pre-upgrade drafts.
func renderDraft(d M, evidence []any, legacy bool) string {
	steps := draftList(arr(d["steps"]))
	if legacy {
		lines := []string{}
		for i, value := range arr(d["steps"]) {
			lines = append(lines, fmt.Sprintf("%d. %s", i+1, str(value)))
		}
		steps = strings.Join(lines, "\n")
	}
	content := "# " + str(d["summary"]) + "\n\n环境：" + str(d["environment"]) + "\n\n## 复现步骤\n\n" + steps
	content += "\n\n## 实际结果\n\n" + str(d["actual"])
	if !legacy && len(arr(d["observations"])) > 0 {
		content += "\n\n## 补充观察\n\n" + draftList(arr(d["observations"]))
	}
	expected := str(d["expected"])
	if !legacy && !present(d["expected"]) {
		items := arr(d["expected"])
		if len(items) == 1 {
			expected = str(items[0])
		} else {
			expected = draftList(items)
		}
	}
	content += "\n\n## 预期结果\n\n" + expected + "\n\n## 依据\n\n" + str(d["basis"])
	if len(evidence) > 0 {
		content += "\n\n## 证据\n\n本地证据已保存，外部可访问引用由实际连接器上传并回读确认。"
	} else if legacy {
		content += "\n\n## 证据\n\n未提供媒体证据。"
	}
	return content + "\n"
}
func Receipt(root string, r M) M {
	errs := []string{}
	e := safeCall(func() error {
		if !member(r["mode"], "mock", "live") {
			errs = append(errs, "Missing mode")
		}
		if !member(r["operation"], "create", "update", "comment", "transition") {
			errs = append(errs, "Invalid operation")
		}
		if !present(r["operation_id"]) {
			errs = append(errs, "Missing operation identity")
		}
		if !present(r["issue_url"]) {
			errs = append(errs, "Missing issue link")
		} else if r["mode"] == "live" && !strings.HasPrefix(str(r["issue_url"]), "https://") {
			errs = append(errs, "Live receipt needs https issue link")
		}
		if r["write_outcome"] != "saved" {
			errs = append(errs, "Write outcome not saved; read back before retry")
		}
		if r["readback_confirmed"] != true {
			errs = append(errs, "Readback not confirmed")
		}
		a, b := str(r["expected_payload_hash"]), str(r["readback_payload_hash"])
		if !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(a) || a != b {
			errs = append(errs, "Readback payload mismatch")
		}
		ev, ok := r["evidence_entries"].([]any)
		if !ok {
			errs = append(errs, "Missing evidence receipts")
		}
		for _, v := range ev {
			item := obj(v)
			if !isFile(path(root, str(item["local"]))) {
				errs = append(errs, "Missing local evidence")
			}
			if item["uploaded"] != true || item["visible"] != true || !present(item["remote_id"]) {
				errs = append(errs, "Evidence upload/display incomplete")
			}
		}
		if r["relationship_required"] == true && r["relationship_saved"] != true {
			errs = append(errs, "Project association incomplete")
		}
		if r["operation"] == "transition" && !present(r["final_state"]) {
			errs = append(errs, "Missing final state")
		}
		return nil
	})
	if e != nil {
		errs = append(errs, e.Error())
	}
	status := "complete"
	if len(errs) > 0 {
		status = "incomplete"
	}
	return M{"receipt_check": status, "mode": r["mode"], "real_jira_verified": false, "note": "Structured receipt only; actual network readback must be evidenced separately.", "errors": errs}
}
