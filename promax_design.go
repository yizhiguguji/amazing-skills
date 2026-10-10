package amazing

// Derived from pinned UI UX Pro Max design_system.py and reasoning_contract.py (MIT).
import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func get(r M, k, def string) string {
	if v, ok := r[k]; ok {
		return str(v)
	}
	return def
}
func first(rs []any) M {
	if len(rs) > 0 {
		return obj(rs[0])
	}
	return M{}
}
func nullable(r M, k string) any {
	if v, ok := r[k]; ok {
		return v
	}
	return nil
}
func lookup(rows []M, kind string) map[string]M {
	m := map[string]M{}
	for _, r := range rows {
		for _, id := range identities(r, kind) {
			m[strings.ToLower(id)] = r
		}
	}
	return m
}
func resolveStyle(m map[string]M, id string) M {
	r := m[strings.ToLower(strings.TrimSpace(id))]
	seen := map[string]bool{}
	for r != nil && str(r["Status"]) == "deprecated" {
		id = str(r["Style ID"])
		parent := str(r["Parent Style ID"])
		if parent == "" || seen[id] {
			return M{}
		}
		seen[id] = true
		r = m[strings.ToLower(parent)]
	}
	if r == nil {
		return M{}
	}
	return r
}

// Preserve JSON rule order and reject duplicate conditions; no executable expression is evaluated.
func applyRules(raw, q string) (M, []any, []string, []string, string, string, error) {
	rules := M{}
	activated := []any{}
	styles, constraints := []string{}, []string{}
	pattern, mode := "", ""
	d := json.NewDecoder(strings.NewReader(raw))
	t, e := d.Token()
	if e != nil || t != json.Delim('{') {
		return nil, nil, nil, nil, "", "", fmt.Errorf("decision rules must be a JSON object")
	}
	for d.More() {
		tok, e := d.Token()
		if e != nil {
			return nil, nil, nil, nil, "", "", e
		}
		key := str(tok)
		if _, exists := rules[key]; exists {
			return nil, nil, nil, nil, "", "", fmt.Errorf("duplicate decision-rule key: %s", key)
		}
		signals := promaxStrings(obj(proContract["signals"])[key])
		if key != "must_have" && len(signals) == 0 {
			return nil, nil, nil, nil, "", "", fmt.Errorf("unknown decision-rule condition: %s", key)
		}
		actions := []string{}
		if e = d.Decode(&actions); e != nil || len(actions) == 0 {
			return nil, nil, nil, nil, "", "", fmt.Errorf("invalid action array: %s", key)
		}
		seen := map[string]bool{}
		for _, a := range actions {
			p := strings.SplitN(a, ":", 2)
			if len(p) != 2 || !member(p[0], "constraint", "style", "pattern", "mode") || seen[a] {
				return nil, nil, nil, nil, "", "", fmt.Errorf("invalid decision-rule action: %s", a)
			}
			seen[a] = true
			if member(p[0], "constraint", "style") && !regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(p[1]) || p[0] == "mode" && !member(p[1], "dark", "light") || p[0] == "pattern" && strings.TrimSpace(p[1]) == "" {
				return nil, nil, nil, nil, "", "", fmt.Errorf("invalid decision-rule action: %s", a)
			}
		}
		rules[key] = actions
		active := key == "must_have"
		for _, s := range signals {
			if phrase(strings.ToLower(q), s) {
				active = true
			}
		}
		if !active {
			continue
		}
		activated = append(activated, M{"condition": key, "actions": actions})
		for _, a := range actions {
			p := strings.SplitN(a, ":", 2)
			switch p[0] {
			case "style":
				if !member(p[1], styles...) {
					styles = append(styles, p[1])
				}
			case "constraint":
				if !member(p[1], constraints...) {
					constraints = append(constraints, p[1])
				}
			case "pattern":
				pattern = p[1]
			case "mode":
				mode = p[1]
			}
		}
	}
	if _, e = d.Token(); e != nil {
		return nil, nil, nil, nil, "", "", e
	}
	if _, e = d.Token(); e != io.EOF {
		return nil, nil, nil, nil, "", "", fmt.Errorf("trailing decision-rule JSON")
	}
	return rules, activated, styles, constraints, pattern, mode, nil
}
func dial(name string, v any) M {
	if v == nil {
		return nil
	}
	n := int(num(v))
	if n < 1 {
		n = 1
	}
	if n > 10 {
		n = 10
	}
	for _, tier := range arr(obj(proContract["dials"])[name]) {
		t := arr(tier)
		if n >= int(float64num(t[0])) && n <= int(float64num(t[1])) {
			m := M{}
			for k, v := range obj(t[2]) {
				m[k] = v
			}
			m["value"] = n
			return m
		}
	}
	return nil
}
func luminance(s string) (float64, bool) {
	s = strings.TrimLeft(strings.TrimSpace(s), "#")
	if len(s) == 3 {
		s = string(s[0]) + string(s[0]) + string(s[1]) + string(s[1]) + string(s[2]) + string(s[2])
	}
	if len(s) != 6 {
		return 0, false
	}
	weights := []float64{.2126, .7152, .0722}
	sum := 0.
	for i := 0; i < 3; i++ {
		n, e := strconv.ParseUint(s[i*2:i*2+2], 16, 8)
		if e != nil {
			return 0, false
		}
		v := float64(n) / 255
		if v <= .04045 {
			v /= 12.92
		} else {
			v = math.Pow((v+.055)/1.055, 2.4)
		}
		sum += weights[i] * v
	}
	return sum, true
}
func contrast(a, b string) float64 {
	x, ok := luminance(a)
	if !ok {
		return 0
	}
	y, ok := luminance(b)
	if !ok {
		return 0
	}
	return (math.Max(x, y) + .05) / (math.Min(x, y) + .05)
}
func darkPalette(r M) bool { v, ok := luminance(str(r["Background"])); return ok && v < .18 }
func deriveDark(r M) M {
	m := M{}
	for k, v := range r {
		m[k] = v
	}
	ring := "#60A5FA"
	for _, k := range []string{"Ring", "Accent", "Primary"} {
		if contrast(str(r[k]), "#0F172A") >= 3 {
			ring = str(r[k])
			break
		}
	}
	for k, v := range (M{"Background": "#0F172A", "Foreground": "#F8FAFC", "Card": "#111827", "Card Foreground": "#F8FAFC", "Muted": "#1E293B", "Muted Foreground": "#CBD5E1", "Border": "#334155", "Ring": ring, "_mode_derivation": "derived-dark"}) {
		m[k] = v
	}
	return m
}
func containsAny(s string, terms []string) bool {
	for _, t := range terms {
		if strings.Contains(s, t) {
			return true
		}
	}
	return false
}
func chooseStyle(results []any, priority []string, styles map[string]M) M {
	if len(results) == 0 {
		return M{}
	}
	for _, p := range priority {
		r := resolveStyle(styles, p)
		if len(r) > 0 {
			return r
		}
	}
	best, high := first(results), 0
	ascii := regexp.MustCompile(`[a-z0-9]+`)
	for _, v := range results {
		r := obj(v)
		score := 0
		text := strings.ToLower(fmt.Sprint(r))
		name := set(ascii.FindAllString(strings.ToLower(str(r["Style Category"])), -1))
		keywords := set(ascii.FindAllString(strings.ToLower(str(r["Keywords"])), -1))
		for _, p := range priority {
			ts := set(ascii.FindAllString(strings.ToLower(p), -1))
			if len(ts) > 0 && subset(ts, name) {
				score += 10
			} else if intersects(ts, keywords) {
				score += 3
			} else {
				for t := range ts {
					if strings.Contains(text, t) {
						score++
						break
					}
				}
			}
		}
		if score > high {
			best, high = r, score
		}
	}
	return best
}
func GenerateDesignSystem(root, q, name string, dials M) (M, error) {
	styles, e := readCSV(root, "styles.csv")
	if e != nil {
		return nil, e
	}
	rules, e := readCSV(root, "ui-reasoning.csv")
	if e != nil {
		return nil, e
	}
	landing, e := readCSV(root, "landing.csv")
	if e != nil {
		return nil, e
	}
	styleLookup := lookup(styles, "style")
	landingLookup := lookup(landing, "landing")
	run := func(query, domain string, n int) ([]any, error) {
		r, e := Search(root, query, domain, "", n)
		return arr(r["results"]), e
	}
	products, e := run(q, "product", 1)
	if e != nil {
		return nil, e
	}
	category := get(first(products), "Product Type", "General")
	var rule M
	for _, r := range rules {
		if strings.EqualFold(strings.TrimSpace(str(r["UI_Category"])), strings.TrimSpace(category)) {
			rule = r
			break
		}
	}
	reason := M{"pattern": "Hero + Features + CTA", "color_mood": "Professional", "typography_mood": "Clean", "key_effects": "Subtle hover transitions", "anti_patterns": "", "severity": "MEDIUM"}
	priority := []string{"Minimalism", "Flat Design"}
	decision := M{}
	activated := []any{}
	constraints := []string{}
	mode := ""
	if rule != nil {
		var ids []string
		var pattern string
		decision, activated, ids, constraints, pattern, mode, e = applyRules(get(rule, "Decision_Rules", "{}"), q)
		if e != nil {
			return nil, e
		}
		priority = []string{}
		for _, id := range ids {
			priority = append(priority, get(resolveStyle(styleLookup, id), "Style Category", id))
		}
		for _, p := range strings.Split(str(rule["Style_Priority"]), "+") {
			priority = append(priority, strings.TrimSpace(p))
		}
		for k, field := range map[string]string{"pattern": "Recommended_Pattern", "color_mood": "Color_Mood", "typography_mood": "Typography_Mood", "key_effects": "Key_Effects", "anti_patterns": "Anti_Patterns", "severity": "Severity"} {
			reason[k] = get(rule, field, "")
		}
		if pattern != "" {
			reason["pattern"] = pattern
		}
	}
	variance, motion, density := dial("variance", dials["variance"]), dial("motion", dials["motion"]), dial("density", dials["density"])
	if variance != nil {
		priority = append(promaxStrings(variance["style_keywords"]), priority...)
	}
	ct := []string{}
	for _, c := range constraints {
		ct = append(ct, strings.ReplaceAll(c, "-", " "))
	}
	parts := []string{}
	for _, s := range []string{q, category, strings.Join(ct, " ")} {
		if s != "" {
			parts = append(parts, s)
		}
	}
	resolved := strings.Join(parts, " ")
	styleQ := resolved
	if len(priority) > 0 {
		limit := len(priority)
		if limit > 2 {
			limit = 2
		}
		styleQ += " " + strings.Join(priority[:limit], " ")
	}
	styleRows, e := run(styleQ, "style", 3)
	if e != nil {
		return nil, e
	}
	bestStyle := chooseStyle(styleRows, priority, styleLookup)
	colorRows, e := run(str(reason["color_mood"])+" "+resolved, "color", 5)
	if e != nil {
		return nil, e
	}
	typoRows, e := run(str(reason["typography_mood"])+" "+resolved, "typography", 2)
	if e != nil {
		return nil, e
	}
	patternQ := str(reason["pattern"])
	if landingLookup[strings.ToLower(patternQ)] == nil {
		patternQ += " " + resolved
	}
	if patternQ == "" {
		patternQ = q
	}
	landingRows, e := run(patternQ, "landing", 2)
	if e != nil {
		return nil, e
	}
	bestLanding, bestTypo := first(landingRows), first(typoRows)
	for _, v := range landingRows {
		r := obj(v)
		if str(r["Pattern Name"]) == str(reason["pattern"]) {
			bestLanding = r
			break
		}
	}
	if mode == "" {
		mode = "light"
		preferred := strings.ToLower(strings.TrimSpace(str(bestStyle["Preferred Mode"])))
		styleDark := preferred == "dark"
		if preferred != "dark" && preferred != "light" {
			styleDark = str(bestStyle["Light Mode ✓"]) == "not-recommended" && str(bestStyle["Dark Mode ✓"]) == "supported" || containsAny(strings.ToLower(str(bestStyle["Light Mode ✓"])+" "+str(bestStyle["Dark Mode ✓"])), promaxStrings(proContract["dark_primary"]))
		}
		if styleDark || containsAny(strings.ToLower(q), promaxStrings(proContract["dark_query"])) {
			mode = "dark"
		}
	}
	bestColor := first(colorRows)
	categoryFound := false
	for _, v := range colorRows {
		r := obj(v)
		if str(r["Product Type"]) == category {
			bestColor = r
			categoryFound = true
			break
		}
	}
	if mode == "dark" && categoryFound && !darkPalette(bestColor) {
		bestColor = deriveDark(bestColor)
	} else if mode == "dark" && !categoryFound {
		for _, v := range colorRows {
			if darkPalette(obj(v)) {
				bestColor = obj(v)
				break
			}
		}
	}
	motionSnippet := M{}
	if motion != nil {
		rs, e := run(q+" "+str(motion["tier"]), "gsap", 5)
		if e != nil {
			return nil, e
		}
		motionSnippet = first(rs)
		for _, v := range rs {
			if str(obj(v)["Intensity Tier"]) == str(motion["tier"]) {
				motionSnippet = obj(v)
				break
			}
		}
	}
	anti := str(reason["anti_patterns"])
	if mode == "dark" {
		kept := []string{}
		for _, clause := range strings.Split(anti, "+") {
			if !containsAny(strings.ToLower(clause), []string{"dark mode", "dark modes", "dark theme"}) && strings.TrimSpace(clause) != "" {
				kept = append(kept, strings.TrimSpace(clause))
			}
		}
		anti = strings.Join(kept, " + ")
	}
	effects := str(bestStyle["Effects & Animation"])
	combined := effects
	if combined == "" {
		combined = str(reason["key_effects"])
	}
	if name == "" {
		name = strings.ToUpper(q)
	}
	colors := M{}
	defaults := map[string]string{"primary": "#2563EB", "secondary": "#3B82F6", "accent": "#F97316", "background": "#F8FAFC", "foreground": "#1E293B"}
	for _, entry := range arr(proContract["semantic_colors"]) {
		v := arr(entry)
		k := str(v[1])
		field := str(v[0])
		if k == "accent" {
			field = "Accent"
		}
		if k == "on_accent" {
			field = "On Accent"
		}
		colors[k] = get(bestColor, field, defaults[k])
	}
	colors["notes"] = get(bestColor, "Notes", "")
	colors["cta"] = colors["accent"]
	colors["text"] = colors["foreground"]
	colors["on_cta"] = colors["on_accent"]
	dialOutput := M{}
	for k, v := range map[string]M{"variance": variance, "motion": motion, "density": density} {
		dialOutput[k] = nullable(v, "value")
		dialOutput[k+"_label"] = nullable(v, "label")
	}
	var productSource, reasonSource any
	if len(products) > 0 {
		productSource = category
	}
	if rule != nil {
		reasonSource = category
	}
	styleSource := nullable(bestStyle, "Style ID")
	if styleSource == nil || styleSource == "" {
		styleSource = nullable(bestStyle, "Style Category")
	}
	return M{"project_name": name, "category": category, "pattern": M{"name": get(bestLanding, "Pattern Name", str(reason["pattern"])), "sections": get(bestLanding, "Section Order", "Hero > Features > CTA"), "cta_placement": get(bestLanding, "Primary CTA Placement", "Above fold"), "color_strategy": get(bestLanding, "Color Strategy", ""), "conversion": get(bestLanding, "Conversion Optimization", "")}, "style": M{"id": get(bestStyle, "Style ID", "minimalism-and-swiss-style"), "name": get(bestStyle, "Style Category", "Minimalism"), "type": get(bestStyle, "Type", "General"), "effects": effects, "keywords": get(bestStyle, "Keywords", ""), "best_for": get(bestStyle, "Best For", ""), "performance": get(bestStyle, "Performance", ""), "accessibility": get(bestStyle, "Accessibility", ""), "light_mode": get(bestStyle, "Light Mode ✓", ""), "dark_mode": get(bestStyle, "Dark Mode ✓", "")}, "colors": colors, "typography": M{"heading": get(bestTypo, "Heading Font", "Inter"), "body": get(bestTypo, "Body Font", "Inter"), "mood": get(bestTypo, "Mood/Style Keywords", str(reason["typography_mood"])), "best_for": get(bestTypo, "Best For", ""), "google_fonts_url": get(bestTypo, "Google Fonts URL", ""), "css_import": get(bestTypo, "CSS Import", "")}, "key_effects": combined, "anti_patterns": anti, "decision_rules": decision, "activated_rules": activated, "constraints": constraints, "reasoning_default": rule == nil, "source_identities": M{"product": productSource, "reasoning": reasonSource, "style": styleSource, "color": nullable(bestColor, "Product Type"), "typography": nullable(bestTypo, "Font Pairing Name"), "landing": nullable(bestLanding, "Pattern Name")}, "source_derivations": M{"color_mode": nullable(bestColor, "_mode_derivation")}, "severity": str(reason["severity"]), "dials": dialOutput, "motion_snippet": motionSnippet, "spacing_scale": nullable(density, "spacing")}, nil
}
func DesignSuggestions(root, q string, dials M) (M, error) {
	return DesignSystemResult(root, q, "", dials, "markdown", false, "", "", false)
}
func DesignSystemResult(root, q, name string, dials M, format string, persist bool, page, output string, force bool) (M, error) {
	ds, e := GenerateDesignSystem(root, q, name, dials)
	if e != nil {
		return nil, e
	}
	var p any
	if persist {
		p, e = PersistDesignSystem(root, ds, page, output, q, force)
		if e != nil {
			return nil, e
		}
	}
	return M{"text": formatDesign(ds, format), "design_system": ds, "persistence": p}, nil
}
func safeSlug(s, fallback string) string {
	s = regexp.MustCompile(`[^a-z0-9_-]+`).ReplaceAllString(strings.ToLower(s), "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = fallback
	}
	return s
}
func atomicPersist(path, content string, force bool) error {
	f, e := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if e != nil {
		return e
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, e = f.WriteString(content); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	if force {
		return os.Rename(tmp, path)
	}
	return os.Link(tmp, path)
}
func PageOverrides(root, page, q string) (M, error) {
	context := strings.ToLower(page) + " " + strings.ToLower(q)
	sr, e := Search(root, context, "style", "", 1)
	if e != nil {
		return nil, e
	}
	ur, e := Search(root, context, "ux", "", 3)
	if e != nil {
		return nil, e
	}
	lr, e := Search(root, context, "landing", "", 1)
	if e != nil {
		return nil, e
	}
	s := first(arr(sr["results"]))
	l := first(arr(lr["results"]))
	layout, spacing, colors := M{}, M{}, M{}
	components, recs := []string{}, []string{}
	if len(s) > 0 {
		k := strings.ToLower(str(s["Keywords"]))
		if containsAny(k, []string{"data", "dense", "dashboard", "grid"}) {
			layout["Max Width"] = "1400px or full-width"
			layout["Grid"] = "12-column grid for data flexibility"
			spacing["Content Density"] = "High — optimize for information display"
		} else if containsAny(k, []string{"minimal", "simple", "clean", "single"}) {
			layout["Max Width"] = "800px (narrow, focused)"
			layout["Layout"] = "Single column, centered"
			spacing["Content Density"] = "Low — focus on clarity"
		} else {
			layout["Max Width"] = "1200px (standard)"
			layout["Layout"] = "Full-width sections, centered content"
		}
		if str(s["Effects & Animation"]) != "" {
			recs = append(recs, "Effects: "+str(s["Effects & Animation"]))
		}
	}
	for _, v := range arr(ur["results"]) {
		u := obj(v)
		if str(u["Do"]) != "" {
			recs = append(recs, str(u["Category"])+": "+str(u["Do"]))
		}
		if str(u["Don't"]) != "" {
			components = append(components, "Avoid: "+str(u["Don't"]))
		}
	}
	if str(l["Section Order"]) != "" {
		layout["Sections"] = l["Section Order"]
	}
	if str(l["Primary CTA Placement"]) != "" {
		recs = append(recs, "CTA Placement: "+str(l["Primary CTA Placement"]))
	}
	if str(l["Color Strategy"]) != "" {
		colors["Strategy"] = l["Color Strategy"]
	}
	if len(layout) == 0 {
		layout = M{"Max Width": "1200px", "Layout": "Responsive grid"}
	}
	if len(recs) == 0 {
		recs = []string{"Refer to MASTER.md for all design rules", "Add specific overrides as needed for this page"}
	}
	patterns := []struct {
		terms []string
		name  string
	}{{[]string{"dashboard", "admin", "analytics", "data", "metrics", "stats", "monitor", "overview"}, "Dashboard / Data View"}, {[]string{"checkout", "payment", "cart", "purchase", "order", "billing"}, "Checkout / Payment"}, {[]string{"settings", "profile", "account", "preferences", "config"}, "Settings / Profile"}, {[]string{"landing", "marketing", "homepage", "hero", "home", "promo"}, "Landing / Marketing"}, {[]string{"login", "signin", "signup", "register", "auth", "password"}, "Authentication"}, {[]string{"pricing", "plans", "subscription", "tiers", "packages"}, "Pricing / Plans"}, {[]string{"blog", "article", "post", "news", "content", "story"}, "Blog / Article"}, {[]string{"product", "item", "detail", "pdp", "shop", "store"}, "Product Detail"}, {[]string{"search", "results", "browse", "filter", "catalog", "list"}, "Search Results"}, {[]string{"empty", "404", "error", "not found", "zero"}, "Empty State"}}
	pageType := "General"
	for _, p := range patterns {
		if containsAny(context, p.terms) {
			pageType = p.name
			break
		}
	}
	if pageType == "General" {
		best := strings.ToLower(str(s["Best For"]))
		if containsAny(best, []string{"dashboard", "data"}) {
			pageType = "Dashboard / Data View"
		} else if containsAny(best, []string{"landing", "marketing"}) {
			pageType = "Landing / Marketing"
		}
	}
	return M{"page_type": pageType, "layout": layout, "spacing": spacing, "typography": M{}, "colors": colors, "components": components, "unique_components": []string{}, "recommendations": recs}, nil
}
func PersistDesignSystem(root string, ds M, page, output, q string, force bool) (M, error) {
	if output == "" {
		output = "."
	}
	base, e := rootPath(output)
	if e != nil {
		return nil, e
	}
	dirRel := filepath.ToSlash(filepath.Join("design-system", safeSlug(str(ds["project_name"]), "default")))
	dir, e := Inside(base, dirRel)
	if e != nil {
		return nil, e
	}
	pages, e := Inside(base, dirRel+"/pages")
	if e != nil {
		return nil, e
	}
	if e = os.MkdirAll(pages, 0755); e != nil {
		return nil, e
	}
	master, e := Inside(base, dirRel+"/MASTER.md")
	if e != nil {
		return nil, e
	}
	created := []string{}
	result := M{"status": "success", "design_system_dir": dir, "master_file": master, "created_files": created}
	content := formatMaster(ds)
	e = atomicPersist(master, content, force)
	if e == nil {
		created = append(created, master)
	} else if os.IsExist(e) {
		if page == "" {
			result["status"] = "skipped_exists"
			result["message"] = master + " already exists and was not modified. Read it first; use --force only for an approved replacement."
			return result, nil
		}
	} else {
		return nil, e
	}
	if page != "" {
		p, e := Inside(base, dirRel+"/pages/"+safeSlug(page, "page")+".md")
		if e != nil {
			return nil, e
		}
		overrides, e := PageOverrides(root, page, q)
		if e != nil {
			return nil, e
		}
		raw, _ := json.MarshalIndent(overrides, "", "  ")
		pageContent := fmt.Sprintf("# %s — Page Overrides\n\nRead ../MASTER.md first. Only the explicit overrides below replace master rules; all other rules inherit MASTER.md.\n\n```json\n%s\n```\n", page, string(raw))
		e = atomicPersist(p, pageContent, force)
		if e == nil {
			created = append(created, p)
		} else if os.IsExist(e) {
			if len(created) == 0 {
				result["status"] = "skipped_exists"
				result["message"] = p + " already exists and was not modified."
			}
		} else {
			return nil, e
		}
	}
	result["created_files"] = created
	return result, nil
}

// Keep deterministic key rendering available for human documentation.
func sortedKeys(m M) []string {
	k := []string{}
	for s := range m {
		k = append(k, s)
	}
	sort.Strings(k)
	return k
}
