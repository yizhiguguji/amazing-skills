package amazing

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func equivalent(a, b any) bool {
	switch x := a.(type) {
	case map[string]any:
		y, ok := b.(map[string]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for k, v := range x {
			if !equivalent(v, y[k]) {
				return false
			}
		}
		return true
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i, v := range x {
			if !equivalent(v, y[i]) {
				return false
			}
		}
		return true
	case float64:
		y, ok := b.(float64)
		return ok && math.Abs(x-y) <= 1e-10*math.Max(1, math.Abs(x))
	default:
		return fmt.Sprint(a) == fmt.Sprint(b)
	}
}
func normalizedJSON(v any) any { b, _ := json.Marshal(v); var a any; json.Unmarshal(b, &a); return a }
func TestProMaxPinnedOriginalParity(t *testing.T) {
	b, e := os.ReadFile("testdata/promax-parity.json")
	need(t, e)
	var f M
	need(t, json.Unmarshal(b, &f))
	for i, v := range arr(f["search"]) {
		c := obj(v)
		t.Run(fmt.Sprintf("search/%d", i), func(t *testing.T) {
			r, e := SearchDetailed("skills", str(c["query"]), str(c["domain"]), str(c["stack"]), int(num(c["limit"])), true)
			need(t, e)
			if !equivalent(normalizedJSON(r), c["expected"]) {
				actual, _ := json.Marshal(r)
				expected, _ := json.Marshal(c["expected"])
				t.Fatalf("%s domain=%s stack=%s\nactual %s\nexpected %s", c["query"], c["domain"], c["stack"], actual, expected)
			}
		})
	}
	for i, v := range arr(f["design"]) {
		c := obj(v)
		t.Run(fmt.Sprintf("design/%d", i), func(t *testing.T) {
			r, e := GenerateDesignSystem("skills", str(c["query"]), "Parity Fixture", obj(c["dials"]))
			need(t, e)
			if !equivalent(normalizedJSON(r), c["expected"]) {
				actual, _ := json.Marshal(r)
				expected, _ := json.Marshal(c["expected"])
				t.Fatalf("%s dials=%v\nactual %s\nexpected %s", c["query"], c["dials"], actual, expected)
			}
		})
	}
	for i, v := range arr(f["pages"]) {
		c := obj(v)
		t.Run(fmt.Sprintf("page/%d", i), func(t *testing.T) {
			r, e := PageOverrides("skills", str(c["page"]), str(c["query"]))
			need(t, e)
			if !equivalent(normalizedJSON(r), c["expected"]) {
				t.Fatalf("actual %v expected %v", r, c["expected"])
			}
		})
	}
}
func TestProMaxPersistenceAndStyleStability(t *testing.T) {
	root := t.TempDir()
	args := []string{"SaaS dashboard dark mode", "--design-system", "--project-name", "Portable Demo", "--persist", "--output-dir", root, "--density", "8", "--motion", "2"}
	r, e := NativeSearchCLI("skills", args)
	need(t, e)
	p := obj(r["persistence"])
	if p["status"] != "success" {
		t.Fatal(p)
	}
	master := str(p["master_file"])
	before, e := os.ReadFile(master)
	need(t, e)
	for _, s := range []string{"--color-primary", "--space-md", "Component Specs", "Pre-Delivery Checklist", "source_identities", "gsap.from"} {
		if !strings.Contains(string(before), s) {
			t.Fatalf("missing %s", s)
		}
	}
	r, e = NativeSearchCLI("skills", append(args, "--page", "dashboard"))
	need(t, e)
	after, e := os.ReadFile(master)
	need(t, e)
	if string(before) != string(after) {
		t.Fatal("page persistence changed master")
	}
	if len(arr(obj(r["persistence"])["created_files"])) != 1 {
		t.Fatal(r)
	}
	args[0] = "ecommerce luxury checkout"
	r, e = NativeSearchCLI("skills", args)
	need(t, e)
	if obj(r["persistence"])["status"] != "skipped_exists" {
		t.Fatal(r)
	}
	after, e = os.ReadFile(master)
	need(t, e)
	if string(before) != string(after) {
		t.Fatal("new query replaced established style")
	}
	r, e = NativeSearchCLI("skills", append(args, "--force"))
	need(t, e)
	after, e = os.ReadFile(master)
	need(t, e)
	if string(before) == string(after) {
		t.Fatal("explicit force did not update")
	}
}
func TestProMaxClosedRuleGrammar(t *testing.T) {
	for _, raw := range []string{`{"must_have":["mode:dark"],"must_have":["mode:light"]}`, `{"if_unknown":["style:minimalism"]}`, `{"must_have":["exec:code"]}`, `{"must_have":["mode:dark","mode:dark"]}`, `{"must_have":["style:../../outside"]}`, `{"must_have":[]}`, `[]`, `{} {}`} {
		if _, _, _, _, _, _, e := applyRules(raw, "dashboard"); e == nil {
			t.Fatal("invalid rules accepted", raw)
		}
	}
	_, activated, _, _, _, mode, e := applyRules(`{"must_have":["mode:dark"],"if_light_mode_needed":["mode:light"]}`, "light mode")
	need(t, e)
	if len(activated) != 2 || mode != "light" {
		t.Fatal(mode, activated)
	}
}
func TestProMaxPersistenceRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	need(t, os.Symlink(outside, filepath.Join(root, "design-system")))
	ds, e := GenerateDesignSystem("skills", "SaaS dashboard", "demo", M{})
	need(t, e)
	if _, e = PersistDesignSystem("skills", ds, "", root, "", false); e == nil {
		t.Fatal("symlink escape accepted")
	}
}
