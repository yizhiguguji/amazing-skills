package amazing

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"time"
)

var projectFiles = map[string]string{"spec.md": ".workbench/spec.md", "DESIGN.md": "DESIGN.md", "plan.md": ".product/plan.md", "decisions.md": ".product/decisions.md", "verification.md": ".product/verification.md"}

func ProjectInit(root string) ([]string, error) {
	out := []string{}
	err := safeCall(func() error {
		for source, dest := range projectFiles {
			p := path(root, dest)
			if !exists(p) {
				b, e := fs.ReadFile(Bundle, "skills/product-workflow/assets/templates/"+source)
				if e != nil {
					return e
				}
				if e = writeBytes(p, b, true); e != nil {
					return e
				}
				out = append(out, dest)
			}
		}
		p := path(root, ".product/state.json")
		if !exists(p) {
			e := write(p, M{"schema_version": 1, "phase": "definition", "status": "draft", "selected": []any{}, "open_questions": []any{}, "evidence": []any{}, "acceptance": []any{}, "updated_at": now()}, true)
			if e != nil {
				return e
			}
			out = append(out, ".product/state.json")
		}
		return nil
	})
	return out, err
}
func ProjectCheck(root string) []string {
	errs := []string{}
	err := safeCall(func() error {
		for _, dest := range projectFiles {
			if !isFile(path(root, dest)) {
				errs = append(errs, "Missing artifact: "+dest)
			}
		}
		s := obj(mustRead(path(root, ".product/state.json")))
		if num(s["schema_version"]) != 1 {
			errs = append(errs, "Unsupported state schema")
		}
		if !member(s["phase"], "definition", "design", "implement", "verify", "complete") {
			errs = append(errs, "Invalid phase")
		}
		if !member(s["status"], "draft", "active", "needs-input", "blocked", "passed") {
			errs = append(errs, "Invalid status")
		}
		for _, k := range []string{"selected", "open_questions", "evidence", "acceptance"} {
			if _, ok := s[k].([]any); !ok {
				errs = append(errs, "Expected list: "+k)
			}
		}
		if _, e := time.Parse(time.RFC3339Nano, str(s["updated_at"])); e != nil {
			errs = append(errs, "Invalid updated_at")
		}
		for _, v := range arr(s["selected"]) {
			m := obj(v)
			for _, k := range []string{"id", "provider", "reason"} {
				if !present(m[k]) {
					errs = append(errs, "Selection needs id, provider and reason")
				}
			}
		}
		for _, v := range arr(s["evidence"]) {
			if !isFile(path(root, str(v))) {
				errs = append(errs, "Missing evidence: "+str(v))
			}
		}
		ids := map[string]bool{}
		for _, v := range arr(s["acceptance"]) {
			a := obj(v)
			id := str(a["id"])
			if !present(id) || !member(a["status"], "passed", "failed", "unverified") {
				errs = append(errs, "Invalid acceptance entry")
			}
			if ids[id] {
				errs = append(errs, "Duplicate acceptance id")
			}
			ids[id] = true
			if a["status"] == "passed" {
				if len(arr(a["evidence"])) == 0 {
					errs = append(errs, "Passed acceptance needs existing evidence: "+id)
				}
				for _, v := range arr(a["evidence"]) {
					if !isFile(path(root, str(v))) {
						errs = append(errs, "Passed acceptance needs existing evidence: "+id)
					}
				}
			}
		}
		if s["phase"] == "complete" {
			if s["status"] != "passed" {
				errs = append(errs, "Complete requires passed status")
			}
			if len(arr(s["open_questions"])) > 0 {
				errs = append(errs, "Complete has unresolved questions")
			}
			if len(arr(s["evidence"])) == 0 {
				errs = append(errs, "Complete requires evidence")
			}
			if len(arr(s["acceptance"])) == 0 {
				errs = append(errs, "Complete requires nonempty passed acceptance list")
			}
			for _, v := range arr(s["acceptance"]) {
				if obj(v)["status"] != "passed" {
					errs = append(errs, "Complete requires passed acceptance")
				}
			}
			for _, p := range projectFiles {
				b, e := fs.ReadFile(osFS(root), p)
				if e == nil {
					for _, marker := range []string{"待定义", "待填写", "待规划", "状态：未验证"} {
						if strings.Contains(string(b), marker) {
							errs = append(errs, "Complete still contains template markers: "+p)
						}
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		errs = append(errs, err.Error())
	}
	return dedup(errs)
}
func artifactDigest(root string, item M) (string, error) {
	p, e := Inside(root, str(item["path"]))
	if e != nil {
		return "", e
	}
	b, e := fs.ReadFile(osFS(root), str(item["path"]))
	if e != nil {
		return "", e
	}
	_ = p
	if item["mode"] == "css-root" {
		matches := regexp.MustCompile(`(?s):root\s*\{([^{}]*)\}`).FindAllStringSubmatch(string(b), -1)
		if len(matches) == 0 {
			return "", fmt.Errorf("No flat :root token block")
		}
		parts := []string{}
		for _, m := range matches {
			parts = append(parts, m[1])
		}
		b = []byte(strings.Join(parts, "\n"))
	} else if item["mode"] != "file" {
		return "", fmt.Errorf("Unknown lock mode")
	}
	return hash(b), nil
}
func LockDesign(root, tokens string, replace bool) (M, error) {
	var result M
	e := safeCall(func() error {
		p := path(root, ".product/design-lock.json")
		if exists(p) && !replace {
			return fmt.Errorf("Design lock exists; compare with check-design")
		}
		items := []any{M{"path": "DESIGN.md", "mode": "file"}}
		if tokens != "" {
			mode := "file"
			if strings.HasSuffix(tokens, ".css") {
				mode = "css-root"
			}
			items = append(items, M{"path": tokens, "mode": mode})
		}
		for _, v := range items {
			i := obj(v)
			h, e := artifactDigest(root, i)
			if e != nil {
				return e
			}
			i["sha256"] = h
		}
		result = M{"schema_version": 1, "created_at": now(), "artifacts": items}
		return write(p, result, !replace)
	})
	return result, e
}
func CheckDesign(root string) []string {
	out := []string{}
	e := safeCall(func() error {
		s := obj(mustRead(path(root, ".product/design-lock.json")))
		if num(s["schema_version"]) != 1 || len(arr(s["artifacts"])) == 0 {
			return fmt.Errorf("Invalid design lock")
		}
		for _, v := range arr(s["artifacts"]) {
			m := obj(v)
			h, e := artifactDigest(root, m)
			if e != nil {
				return e
			}
			if h != str(m["sha256"]) {
				out = append(out, "Design drift: "+str(m["path"]))
			}
		}
		return nil
	})
	if e != nil {
		out = append(out, e.Error())
	}
	return out
}
