package amazing

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func formatDesign(ds M, format string) string {
	md := formatDesignMarkdown(ds)
	if format != "ascii" {
		return md
	}
	lines := strings.Split(md, "\n")
	for i, l := range lines {
		l = strings.ReplaceAll(l, "**", "")
		l = strings.TrimLeft(l, "#")
		lines[i] = strings.TrimSpace(l)
	}
	return strings.Join(lines, "\n")
}
func formatDesignMarkdown(ds M) string {
	var b strings.Builder
	line := func(f string, v ...any) { fmt.Fprintf(&b, f+"\n", v...) }
	line("# %s — Design System", str(ds["project_name"]))
	line("")
	line("**Category:** %s", str(ds["category"]))
	for _, section := range []string{"pattern", "style"} {
		line("\n## %s\n", strings.ToUpper(section[:1])+section[1:])
		r := obj(ds[section])
		for _, k := range sortedKeys(r) {
			if str(r[k]) != "" {
				line("- **%s:** %s", k, str(r[k]))
			}
		}
	}
	line("\n## Colors\n")
	line("| Role | Hex | CSS Variable |")
	line("|---|---|---|")
	colors := obj(ds["colors"])
	for _, v := range arr(proContract["semantic_colors"]) {
		entry := arr(v)
		if color := str(colors[str(entry[1])]); color != "" {
			line("| %s | `%s` | `%s` |", str(entry[0]), color, str(entry[2]))
		}
	}
	if str(colors["notes"]) != "" {
		line("\n**Color Notes:** %s", str(colors["notes"]))
	}
	line("\n## Typography\n")
	typo := obj(ds["typography"])
	for _, k := range []string{"heading", "body", "mood", "best_for", "google_fonts_url"} {
		if str(typo[k]) != "" {
			line("- **%s:** %s", k, str(typo[k]))
		}
	}
	if str(typo["css_import"]) != "" {
		line("\n```css\n%s\n```", str(typo["css_import"]))
	}
	line("\n## Key Effects\n\n%s", str(ds["key_effects"]))
	line("\n## Anti-Patterns\n")
	for _, c := range strings.Split(str(ds["anti_patterns"]), "+") {
		if strings.TrimSpace(c) != "" {
			line("- %s", strings.TrimSpace(c))
		}
	}
	motion := obj(ds["motion_snippet"])
	if len(motion) > 0 {
		line("\n## Motion\n")
		for _, k := range sortedKeys(motion) {
			if k == "GSAP Snippet" {
				line("\n```js\n%s\n```", str(motion[k]))
			} else {
				line("- **%s:** %s", k, str(motion[k]))
			}
		}
	}
	line("\n## Reasoning and Provenance\n")
	audit := M{}
	for _, k := range []string{"decision_rules", "activated_rules", "constraints", "reasoning_default", "source_identities", "source_derivations", "severity", "dials", "spacing_scale"} {
		audit[k] = ds[k]
	}
	raw, _ := json.MarshalIndent(audit, "", "  ")
	line("```json\n%s\n```", string(raw))
	return b.String()
}
func formatMaster(ds M) string {
	var b strings.Builder
	b.WriteString("# Design System Master File\n\n> Read this master before UI changes. Explicit rules in pages/[page-name].md override this master; all unspecified rules inherit it. Review change requests before replacing established decisions.\n\n")
	b.WriteString(formatDesignMarkdown(ds))
	b.WriteString("\n## Spacing Variables\n\n| Token | Value | Usage |\n|---|---|---|\n")
	scale := obj(ds["spacing_scale"])
	if len(scale) == 0 {
		tier := arr(arr(obj(proContract["dials"])["density"])[1])
		scale = obj(obj(tier[2])["spacing"])
	}
	keys := []string{"xs", "sm", "md", "lg", "xl", "2xl", "3xl"}
	usages := []string{"Tight gaps", "Icon gaps, inline spacing", "Standard padding", "Section padding", "Large gaps", "Section margins", "Hero padding"}
	for i, k := range keys {
		px := str(scale[k])
		n, _ := strconv.Atoi(strings.TrimSuffix(px, "px"))
		fmt.Fprintf(&b, "| `--space-%s` | `%s` / `%grem` | %s |\n", k, px, float64(n)/16, usages[i])
	}
	b.WriteString("\n")
	b.Write(mustReadBundle("promax/shadows.md"))
	colors := obj(ds["colors"])
	outline := str(colors["primary"])
	if contrast(outline, str(colors["background"])) < 4.5 && str(colors["foreground"]) != "" {
		outline = str(colors["foreground"])
	}
	components := string(mustReadBundle("promax/components.md"))
	onCTA := str(colors["on_cta"])
	if onCTA == "" {
		onCTA = "white"
	}
	components = strings.NewReplacer("{{PRIMARY}}", str(colors["primary"]), "{{BACKGROUND}}", str(colors["background"]), "{{CTA}}", str(colors["cta"]), "{{ON_CTA}}", onCTA).Replace(components)
	components = strings.Replace(components, "color: "+str(colors["primary"])+";", "color: "+outline+";", 1)
	b.WriteString(components)
	b.Write(mustReadBundle("promax/checklist.md"))
	return b.String()
}
func formatSearchResult(r M, full bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## UI Pro Max %s\n", map[bool]string{true: "Stack Guidelines", false: "Search Results"}[str(r["stack"]) != ""])
	if str(r["stack"]) != "" {
		fmt.Fprintf(&b, "**Stack:** %s | **Query:** %s\n", r["stack"], r["query"])
	} else {
		fmt.Fprintf(&b, "**Domain:** %s | **Query:** %s\n", r["domain"], r["query"])
	}
	fmt.Fprintf(&b, "**Source:** %s | **Found:** %v results\n\n", r["file"], r["count"])
	if num(r["count"]) == 0 {
		if redirect := obj(r["redirect"]); len(redirect) > 0 {
			fmt.Fprintf(&b, "Legacy style redirects to domain `%s`, identity `%s`. Search that domain.\n", redirect["domain"], redirect["id"])
		} else {
			b.WriteString("No database match found. Retry broader/different keywords before using defaults; disclose any fallback.\n")
			if suggestions := promaxStrings(r["suggestions"]); len(suggestions) > 0 {
				fmt.Fprintf(&b, "**Closest known terms:** %s\n", strings.Join(suggestions, ", "))
			}
		}
		return b.String()
	}
	cols := promaxStrings(obj(obj(proContract["csv"])[str(r["domain"])])["output_cols"])
	if str(r["stack"]) != "" {
		cols = promaxStrings(obj(proContract["stack_cols"])["output_cols"])
	}
	untruncated := []string{"Code Example Good", "Code Example Bad", "Code Good", "Code Bad", "Implementation Checklist", "Design System Variables", "CSS Import", "Tailwind Config", "GSAP Snippet"}
	for i, v := range arr(r["results"]) {
		fmt.Fprintf(&b, "### Result %d\n", i+1)
		row := obj(v)
		for _, k := range cols {
			if val, exists := row[k]; exists {
				s := str(val)
				runes := []rune(s)
				if !full && !member(k, untruncated...) && len(runes) > 300 {
					s = string(runes[:300]) + "..."
				}
				fmt.Fprintf(&b, "- **%s:** %s\n", k, s)
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}
