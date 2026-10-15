package amazing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const designAssetRoot = "skills/design-reference-extraction/assets/"
const designSchema = "amazing-design-system/v1"

var designFiles = []string{"design-system.json", "DESIGN.md", "tokens.css", "design-evidence.json", "design-preview.html", "design-system.LICENSE"}
var designTokens = []string{"--ds-bg", "--ds-surface", "--ds-fg", "--ds-muted", "--ds-border", "--ds-accent", "--ds-accent-on", "--ds-danger", "--ds-font-body", "--ds-font-display", "--ds-space", "--ds-radius", "--ds-content", "--ds-text", "--ds-leading", "--ds-control"}
var designID = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
var designVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

type DesignManifest struct {
	Schema      string            `json:"schema"`
	ID          string            `json:"id"`
	Version     string            `json:"version"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Source      DesignSource      `json:"source"`
	Files       map[string]string `json:"files"`
}
type DesignSource struct {
	Origin   string `json:"origin"`
	License  string `json:"license"`
	Revision string `json:"revision"`
}
type DesignClaim struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Reference string `json:"reference"`
	Locator   string `json:"locator"`
	Claim     string `json:"claim"`
	Decision  string `json:"decision"`
	Token     string `json:"token,omitempty"`
	File      string `json:"file,omitempty"`
	SHA256    string `json:"sha256,omitempty"`
}
type DesignEvidence struct {
	SchemaVersion int           `json:"schema_version"`
	Claims        []DesignClaim `json:"claims"`
	Unknowns      []string      `json:"unknowns"`
}
type designPackage struct {
	Manifest    DesignManifest
	Evidence    DesignEvidence
	Tokens      map[string]string
	CSS, Design string
	Files       map[string][]byte
}

func strictDesignJSON(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return e
	}
	var extra any
	if e := d.Decode(&extra); e != io.EOF {
		return fmt.Errorf("Expected one JSON value")
	}
	return nil
}

// The portable token profile is deliberately a flat, declarative CSS subset.
// App CSS, media queries and external resources belong in the implementation.
func parseDesignTokens(css string) (map[string]string, error) {
	css = regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(css, "")
	m := regexp.MustCompile(`(?s)^\s*:root\s*\{([^{}]*)\}\s*$`).FindStringSubmatch(css)
	if m == nil {
		return nil, fmt.Errorf("tokens.css requires one flat :root block")
	}
	tokens := map[string]string{}
	for _, declaration := range strings.Split(m[1], ";") {
		if strings.TrimSpace(declaration) == "" {
			continue
		}
		p := strings.SplitN(declaration, ":", 2)
		if len(p) != 2 {
			return nil, fmt.Errorf("Invalid token declaration")
		}
		key, value := strings.TrimSpace(p[0]), strings.TrimSpace(p[1])
		if !regexp.MustCompile(`^--ds-[a-z][a-z0-9-]*$`).MatchString(key) || !regexp.MustCompile(`^[a-zA-Z0-9#(),.%'" _+\-]+$`).MatchString(value) || strings.Contains(strings.ToLower(value), "url(") || strings.Contains(strings.ToLower(value), "expression(") {
			return nil, fmt.Errorf("Unsupported token syntax: %s", key)
		}
		if _, exists := tokens[key]; exists {
			return nil, fmt.Errorf("Duplicate token: %s", key)
		}
		tokens[key] = value
		withoutVars := regexp.MustCompile(`var\(\s*--ds-[a-z0-9-]+\s*\)`).ReplaceAllString(value, "")
		if strings.Contains(strings.ToLower(withoutVars), "var(") {
			return nil, fmt.Errorf("Only simple var(--ds-name) references are supported: %s", key)
		}
	}
	for _, key := range designTokens {
		if tokens[key] == "" {
			return nil, fmt.Errorf("Missing token: %s", key)
		}
	}
	var visit func(string, map[string]bool) error
	visit = func(k string, stack map[string]bool) error {
		if tokens[k] == "" {
			return fmt.Errorf("Unresolved token: %s", k)
		}
		if stack[k] {
			return fmt.Errorf("Cyclic token: %s", k)
		}
		stack[k] = true
		for _, ref := range regexp.MustCompile(`var\(\s*(--ds-[a-z0-9-]+)\s*\)`).FindAllStringSubmatch(tokens[k], -1) {
			if e := visit(ref[1], stack); e != nil {
				return e
			}
		}
		delete(stack, k)
		return nil
	}
	for k := range tokens {
		if e := visit(k, map[string]bool{}); e != nil {
			return nil, e
		}
	}
	return tokens, nil
}
func readDesignPackage(f fs.FS, base string) (*designPackage, error) {
	p := &designPackage{Files: map[string][]byte{}}
	readFile := func(name string) ([]byte, error) {
		if !fs.ValidPath(name) || strings.ContainsAny(name, `\:`) {
			return nil, fmt.Errorf("Unsafe package path")
		}
		// Lstat every segment: portable packages never contain links, even internal links.
		parts := strings.Split(name, "/")
		for i := range parts {
			if member(parts[i], ".git", "node_modules", "__pycache__", ".DS_Store", "tool-cache") {
				return nil, fmt.Errorf("Excluded package content: %s", parts[i])
			}
			info, err := fs.Lstat(f, base+strings.Join(parts[:i+1], "/"))
			if err != nil || info.Mode()&fs.ModeSymlink != 0 {
				return nil, fmt.Errorf("Missing or linked package file: %s", name)
			}
		}
		info, err := fs.Stat(f, base+name)
		if err != nil || !info.Mode().IsRegular() || info.Size() > 2<<20 {
			return nil, fmt.Errorf("Invalid package file: %s", name)
		}
		b, err := fs.ReadFile(f, base+name)
		if err == nil {
			p.Files[name] = b
		}
		return b, err
	}
	for _, name := range designFiles {
		if name == "design-preview.html" {
			continue
		}
		if _, e := readFile(name); e != nil {
			return nil, e
		}
	}
	if e := strictDesignJSON(p.Files["design-system.json"], &p.Manifest); e != nil {
		return nil, fmt.Errorf("design-system.json: %w", e)
	}
	m := p.Manifest
	if m.Schema != designSchema || !designID.MatchString(m.ID) || !designVersion.MatchString(m.Version) || strings.TrimSpace(m.Name) == "" || strings.TrimSpace(m.Description) == "" || strings.TrimSpace(m.Source.Origin) == "" || strings.TrimSpace(m.Source.License) == "" || strings.TrimSpace(m.Source.Revision) == "" {
		return nil, fmt.Errorf("Invalid design system metadata/source")
	}
	want := map[string]string{"design": "DESIGN.md", "tokens": "tokens.css", "evidence": "design-evidence.json", "preview": "design-preview.html", "license": "design-system.LICENSE"}
	if len(m.Files) != len(want) {
		return nil, fmt.Errorf("Invalid design system file map")
	}
	for k, v := range want {
		if m.Files[k] != v {
			return nil, fmt.Errorf("Invalid package path for %s", k)
		}
	}
	p.Design = string(p.Files["DESIGN.md"])
	if len(strings.TrimSpace(p.Design)) < 40 || len(bytes.TrimSpace(p.Files["design-system.LICENSE"])) == 0 {
		return nil, fmt.Errorf("Missing design prose or license")
	}
	var e error
	// Only inject the exact CSS subset we validate. Comments are inert in a CSS
	// file but can close an HTML <style> element before the CSS parser sees them.
	p.CSS = regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(string(p.Files["tokens.css"]), "")
	p.Tokens, e = parseDesignTokens(p.CSS)
	if e != nil {
		return nil, e
	}
	if e = strictDesignJSON(p.Files["design-evidence.json"], &p.Evidence); e != nil {
		return nil, fmt.Errorf("design-evidence.json: %w", e)
	}
	if p.Evidence.SchemaVersion != 1 || len(p.Evidence.Claims) == 0 || p.Evidence.Unknowns == nil {
		return nil, fmt.Errorf("Evidence needs claims and explicit unknowns list")
	}
	seen := map[string]bool{}
	for _, c := range p.Evidence.Claims {
		if !designID.MatchString(c.ID) || seen[c.ID] || !member(c.Kind, "observed", "provided", "inferred") || strings.TrimSpace(c.Reference) == "" || strings.TrimSpace(c.Locator) == "" || strings.TrimSpace(c.Claim) == "" || strings.TrimSpace(c.Decision) == "" {
			return nil, fmt.Errorf("Invalid/duplicate evidence claim: %s", c.ID)
		}
		seen[c.ID] = true
		if c.Token != "" && p.Tokens[c.Token] == "" {
			return nil, fmt.Errorf("Evidence refers to missing token: %s", c.Token)
		}
		if c.Kind == "observed" && c.File == "" {
			return nil, fmt.Errorf("Observed claim requires local evidence: %s", c.ID)
		}
		if c.File != "" {
			if !strings.HasPrefix(c.File, "sources/") {
				return nil, fmt.Errorf("Evidence files must live under sources/")
			}
			b, err := readFile(c.File)
			if err != nil {
				return nil, err
			}
			if hash(b) != c.SHA256 {
				return nil, fmt.Errorf("Evidence digest mismatch: %s", c.ID)
			}
		} else if c.SHA256 != "" {
			return nil, fmt.Errorf("Evidence hash without file")
		}
	}
	return p, nil
}
func designPreview(p *designPackage) ([]byte, error) {
	b, e := fs.ReadFile(Bundle, designAssetRoot+"preview.gohtml")
	if e != nil {
		return nil, e
	}
	t, e := template.New("preview").Parse(string(b))
	if e != nil {
		return nil, e
	}
	type swatch struct {
		Name, Value string
		CSS         template.CSS
	}
	swatches := []swatch{}
	for _, k := range []string{"--ds-bg", "--ds-surface", "--ds-fg", "--ds-muted", "--ds-accent", "--ds-danger"} {
		swatches = append(swatches, swatch{k, p.Tokens[k], template.CSS("background:var(" + k + ")")})
	}
	digest := M{}
	for name, data := range p.Files {
		digest[name] = hash(data)
	}
	var out bytes.Buffer
	e = t.Execute(&out, struct {
		Manifest       DesignManifest
		Evidence       DesignEvidence
		CSS            template.CSS
		Design, Digest string
		Swatches       []swatch
	}{p.Manifest, p.Evidence, template.CSS(p.CSS), p.Design, hash(canonical(digest)), swatches})
	return out.Bytes(), e
}
func DesignSystemCheck(f fs.FS, base string) []string {
	p, e := readDesignPackage(f, base)
	if e != nil {
		return []string{e.Error()}
	}
	want, e := designPreview(p)
	if e != nil {
		return []string{e.Error()}
	}
	info, e := fs.Lstat(f, base+"design-preview.html")
	if e != nil || !info.Mode().IsRegular() {
		return []string{"Missing or linked design-preview.html; run design-system preview"}
	}
	got, e := fs.ReadFile(f, base+"design-preview.html")
	if e != nil || !bytes.Equal(got, want) {
		return []string{"Stale component preview; regenerate after package changes"}
	}
	return []string{}
}
func DesignSystemList() ([]M, error) {
	entries, e := fs.ReadDir(Bundle, designAssetRoot+"systems")
	if e != nil {
		return nil, e
	}
	out := []M{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		p, e := readDesignPackage(Bundle, designAssetRoot+"systems/"+entry.Name()+"/")
		if e != nil {
			return nil, e
		}
		out = append(out, M{"id": p.Manifest.ID, "name": p.Manifest.Name, "description": p.Manifest.Description, "source": p.Manifest.Source, "synthetic": true})
	}
	return out, nil
}
func DesignSystemInit(output, preset string) (M, error) {
	if !member(preset, "calm-paper", "precise-console") {
		return nil, fmt.Errorf("Unknown synthetic preset")
	}
	p, e := readDesignPackage(Bundle, designAssetRoot+"systems/"+preset+"/")
	if e != nil {
		return nil, e
	}
	preview, e := designPreview(p)
	if e != nil {
		return nil, e
	}
	p.Files["design-preview.html"] = preview
	output, e = rootPath(output)
	if e != nil {
		return nil, e
	}
	if e = os.MkdirAll(filepath.Dir(output), 0755); e != nil {
		return nil, e
	}
	if e = os.Mkdir(output, 0755); e != nil {
		return nil, fmt.Errorf("Output must be a fresh directory")
	}
	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(output)
		}
	}()
	for name, data := range p.Files {
		if e = writeBytes(filepath.Join(output, name), data, true); e != nil {
			return nil, e
		}
	}
	success = true
	return M{"created": designFiles, "preset": preset, "reference_extraction": "not performed", "visual_verified": false}, nil
}
func DesignSystemPreview(root string, replace bool) (M, error) {
	p, e := readDesignPackage(os.DirFS(root), "")
	if e != nil {
		return nil, e
	}
	b, e := designPreview(p)
	if e != nil {
		return nil, e
	}
	dest := filepath.Join(root, "design-preview.html")
	info, statErr := os.Lstat(dest)
	if statErr == nil && (!replace || !info.Mode().IsRegular()) {
		return nil, fmt.Errorf("Preview exists or is not a regular file; --replace only regenerates a regular preview")
	}
	if statErr != nil && !os.IsNotExist(statErr) {
		return nil, statErr
	}
	// Atomic regeneration avoids truncating the previous preview on write failure.
	tmp, e := os.CreateTemp(root, ".design-preview-*")
	if e != nil {
		return nil, e
	}
	defer os.Remove(tmp.Name())
	if _, e = tmp.Write(b); e != nil {
		_ = tmp.Close()
		return nil, e
	}
	if e = tmp.Close(); e != nil {
		return nil, e
	}
	if e = os.Chmod(tmp.Name(), 0644); e != nil {
		return nil, e
	}
	if !replace { // exclusive creation when the target did not exist
		if e = os.Link(tmp.Name(), dest); e != nil {
			return nil, e
		}
	} else if e = os.Rename(tmp.Name(), dest); e != nil {
		return nil, e
	}
	return M{"preview": "design-preview.html", "visual_verified": false, "note": "Open in a browser; structure and freshness are not visual acceptance."}, nil
}
func DesignSystemApply(source, project string) (M, error) {
	if errors := DesignSystemCheck(os.DirFS(source), ""); len(errors) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	p, e := readDesignPackage(os.DirFS(source), "")
	if e != nil {
		return nil, e
	}
	preview, e := os.ReadFile(filepath.Join(source, "design-preview.html"))
	if e != nil {
		return nil, e
	}
	p.Files["design-preview.html"] = preview
	if _, e = os.Lstat(filepath.Join(project, ".product/design-lock.json")); !os.IsNotExist(e) {
		return nil, fmt.Errorf("Existing design lock; reconcile through project decisions")
	}
	receipt := ".product/design-system-selection.json"
	p.Files[receipt] = append(canonical(M{"schema_version": 1, "id": p.Manifest.ID, "version": p.Manifest.Version, "manifest_sha256": hash(p.Files["design-system.json"]), "note": "Package applied; browser and implementation acceptance pending."}), '\n')
	names := []string{}
	for name := range p.Files {
		dest, err := Inside(project, name)
		if err != nil {
			return nil, err
		}
		if _, err = os.Lstat(dest); !os.IsNotExist(err) {
			return nil, fmt.Errorf("Refusing to overwrite project artifact: %s", name)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	created := []string{}
	for _, name := range names {
		dest, err := Inside(project, name)
		if err == nil {
			err = writeBytes(dest, p.Files[name], true)
		}
		if err != nil {
			for _, rel := range created {
				_ = os.Remove(filepath.Join(project, rel))
			}
			return nil, err
		}
		created = append(created, name)
	}
	return M{"applied": names, "id": p.Manifest.ID, "version": p.Manifest.Version, "visual_verified": false, "implementation_verified": false, "next": "Integrate tokens, inspect actual UI, then project lock-design --tokens tokens.css"}, nil
}
