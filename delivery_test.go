package amazing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func defectFixture() M {
	return M{"id": "FD-NEW", "summary": "Synthetic observation", "environment": "Synthetic fixture, not a real device", "steps": []any{"Open synthetic record", "Submit"}, "actual": "Observed rejected save", "expected": "Save succeeds under confirmed rule", "basis": "Synthetic requirement R1", "evidence": []any{}}
}
func TestDraftNoEvidenceOmitsBodyPlaceholderAndRetainsTrace(t *testing.T) {
	root := t.TempDir()
	d := defectFixture()
	r, e := Draft(root, d)
	need(t, e)
	b, e := os.ReadFile(filepath.Join(root, str(r["draft"])))
	need(t, e)
	for _, unwanted := range []string{"## 证据", "未提供媒体", "暂无证据", "<span", "color:"} {
		if strings.Contains(string(b), unwanted) {
			t.Fatal("project preference/placeholder leaked", unwanted)
		}
	}
	if !strings.Contains(string(b), str(d["environment"])) || r["evidence_status"] != "not-attached" || num(r["evidence_count"]) != 0 || r["remote_submission"] != "not performed" {
		t.Fatal(r)
	}
	if r["payload_hash"] != hash(b) {
		t.Fatal("hash does not describe stored body")
	}
}
func TestDraftStructuredExpectedAndObservations(t *testing.T) {
	root := t.TempDir()
	d := defectFixture()
	d["expected"] = []any{"Record one retains its amount\nCompare with the before-save value", "Record two is unchanged"}
	d["observations"] = []any{"Retry succeeded in a separate run"}
	d["steps"] = []any{"Open\nRecord current value", "Submit"}
	r, e := Draft(root, d)
	need(t, e)
	b, e := os.ReadFile(filepath.Join(root, str(r["draft"])))
	need(t, e)
	body := string(b)
	for _, want := range []string{"1. Record one retains its amount\n   Compare with the before-save value\n2. Record two is unchanged", "## 补充观察\n\n1. Retry succeeded in a separate run", "1. Open\n   Record current value\n2. Submit"} {
		if !strings.Contains(body, want) {
			t.Fatal(body)
		}
	}
	if _, e = Draft(root, d); e != nil {
		t.Fatal("idempotent list draft rejected", e)
	}
	d["observations"] = []any{"A different observation"}
	if _, e = Draft(root, d); e == nil {
		t.Fatal("changed observations overwritten")
	}
}
func TestDraftLegacyBodiesAndHashesArePreserved(t *testing.T) {
	for _, withEvidence := range []bool{false, true} {
		root := t.TempDir()
		d := defectFixture()
		if withEvidence {
			need(t, writeBytes(filepath.Join(root, "evidence/observation.txt"), []byte("Synthetic observation"), true))
			d["evidence"] = []any{"evidence/observation.txt"}
		}
		// Historical layout intentionally uses the previous exact renderer, including
		// unindented multiline steps and a placeholder for missing media.
		d["steps"] = []any{"Open\nRecord current state", "Submit"}
		old := "# Synthetic observation\n\n环境：Synthetic fixture, not a real device\n\n## 复现步骤\n\n1. Open\nRecord current state\n2. Submit\n\n## 实际结果\n\nObserved rejected save\n\n## 预期结果\n\nSave succeeds under confirmed rule\n\n## 依据\n\nSynthetic requirement R1\n\n## 证据\n\n"
		if withEvidence {
			old += "本地证据已保存，外部可访问引用由实际连接器上传并回读确认。\n"
		} else {
			old += "未提供媒体证据。\n"
		}
		p := filepath.Join(root, ".qa/defect-drafts/FD-NEW.md")
		need(t, writeBytes(p, []byte(old), true))
		r, e := Draft(root, d)
		need(t, e)
		b, e := os.ReadFile(p)
		need(t, e)
		if string(b) != old || r["payload_hash"] != hash([]byte(old)) || r["format"] != "legacy-preserved" {
			t.Fatal(r)
		}
		d["actual"] = "Changed observation"
		if _, e = Draft(root, d); e == nil {
			t.Fatal("legacy content collision accepted")
		}
	}
}
func TestDraftInvalidOptionalFieldsCannotWrite(t *testing.T) {
	for _, value := range []any{[]any{}, []any{""}, []any{"valid", 5}, 12, nil} {
		d := defectFixture()
		d["expected"] = value
		root := t.TempDir()
		if _, e := Draft(root, d); e == nil {
			t.Fatal("invalid expected accepted", value)
		}
		if exists(filepath.Join(root, ".qa")) {
			t.Fatal("invalid draft wrote artifacts")
		}
	}
	for _, value := range []any{[]any{}, "unstructured", nil, []any{""}} {
		d := defectFixture()
		d["observations"] = value
		if _, e := Draft(t.TempDir(), d); e == nil {
			t.Fatal("invalid observations accepted", value)
		}
	}
}

func TestDraftMultilineDoubleDigitStepsStayInTheirItem(t *testing.T) {
	root := t.TempDir()
	d := defectFixture()
	steps := make([]any, 10)
	for i := range steps {
		steps[i] = "Perform action"
	}
	steps[9] = "Read value\r\nCompare with baseline"
	d["steps"] = steps
	d["expected"] = "1. Preserve caller formatting\n\n2. Do not renumber this string"
	r, e := Draft(root, d)
	need(t, e)
	b, e := os.ReadFile(filepath.Join(root, str(r["draft"])))
	need(t, e)
	if !strings.Contains(string(b), "10. Read value\n    Compare with baseline") || !strings.Contains(string(b), str(d["expected"])) {
		t.Fatal(string(b))
	}
}
func TestEditorialRevisionKeepsHistoryAndDoesNotInheritPass(t *testing.T) {
	root, plans := qaFixture(t)
	_, e := Snapshot(root, "editorial-baseline")
	need(t, e)
	_, e = Record(root, event("editorial-before", "passed", "before"))
	need(t, e)
	plan := obj(clone(plans[1]))
	c := obj(arr(plan["cases"])[0])
	c["objective"] = str(c["objective"]) + " (wording clarified)"
	need(t, write(filepath.Join(root, ".qa/plan.json"), plan, false))
	checkHas(t, root, "Case changed without revision")
	c["revision"] = num(c["revision"]) + 1
	need(t, write(filepath.Join(root, ".qa/plan.json"), plan, false))
	if es := QACheck(root); len(es) > 0 {
		t.Fatal(es)
	}
	summary, e := Summary(root, "")
	need(t, e)
	if num(obj(summary["counts"])["passed"]) != 0 {
		t.Fatal("editorial revision inherited old pass")
	}
	events, e := ReadEvents(root)
	need(t, e)
	if len(events) != 1 || num(events[0]["case_revision"]) == num(c["revision"]) {
		t.Fatal("historical execution rewritten")
	}
}
