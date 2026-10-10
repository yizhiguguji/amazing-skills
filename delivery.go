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
		for _, key := range []string{"id", "summary", "environment", "actual", "expected", "basis"} {
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
		content := "# " + str(d["summary"]) + "\n\n环境：" + str(d["environment"]) + "\n\n## 复现步骤\n\n"
		for i, v := range arr(d["steps"]) {
			if i > 0 {
				content += "\n"
			}
			content += fmt.Sprintf("%d. %s", i+1, str(v))
		}
		content += "\n\n## 实际结果\n\n" + str(d["actual"]) + "\n\n## 预期结果\n\n" + str(d["expected"]) + "\n\n## 依据\n\n" + str(d["basis"]) + "\n\n## 证据\n\n"
		if len(evidence) > 0 {
			content += "本地证据已保存，外部可访问引用由实际连接器上传并回读确认。"
		} else {
			content += "未提供媒体证据。"
		}
		content += "\n"
		rel := ".qa/defect-drafts/" + str(d["id"]) + ".md"
		p := path(root, rel)
		if exists(p) {
			b, e := os.ReadFile(p)
			if e != nil {
				return e
			}
			if string(b) != content {
				return fmt.Errorf("Draft exists with different content; retain revision instead of overwrite")
			}
		} else if e := writeBytes(p, []byte(content), true); e != nil {
			return e
		}
		result = M{"draft": rel, "remote_submission": "not performed", "payload_hash": hash([]byte(content))}
		return nil
	})
	return result, e
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
