package amazing

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func Catalog(f fs.FS) (M, error) { v, e := load(f, "catalog.json"); return obj(v), e }
func Closure(c M, chosen []string) ([]string, error) {
	entries := map[string]M{}
	for _, v := range arr(c["skills"]) {
		e := obj(v)
		entries[str(e["id"])] = e
	}
	result := []string{}
	seen := map[string]bool{}
	visiting := map[string]bool{}
	var visit func(string) error
	visit = func(n string) error {
		e, ok := entries[n]
		if !ok {
			return fmt.Errorf("Unknown skill: %s", n)
		}
		if visiting[n] {
			return fmt.Errorf("Dependency cycle: %s", n)
		}
		if seen[n] {
			return nil
		}
		visiting[n] = true
		for _, d := range arr(e["requires"]) {
			if err := visit(str(d)); err != nil {
				return err
			}
		}
		delete(visiting, n)
		seen[n] = true
		result = append(result, n)
		return nil
	}
	for _, n := range chosen {
		if e := visit(n); e != nil {
			return nil, e
		}
	}
	return result, nil
}
func ProviderIntegrity(f fs.FS, skillsPrefix string, p M) []string {
	errs := []string{}
	folder := skillsPrefix + str(p["installed_directory"])
	files := obj(p["files"])
	for rel, expected := range files {
		b, e := fs.ReadFile(f, folder+"/"+rel)
		if e != nil || hash(b) != str(expected) {
			errs = append(errs, "Provider integrity: "+str(p["id"])+"/"+rel)
		}
	}
	return dedup(errs)
}
func Install(f fs.FS, target string, chosen []string, executable string) ([]string, error) {
	c, e := Catalog(f)
	if e != nil {
		return nil, e
	}
	if len(chosen) == 0 {
		for _, v := range arr(c["skills"]) {
			chosen = append(chosen, str(obj(v)["id"]))
		}
	}
	names, e := Closure(c, chosen)
	if e != nil {
		return nil, e
	}
	l, e := load(f, "providers.lock.json")
	if e != nil {
		return nil, e
	}
	for _, v := range arr(obj(l)["providers"]) {
		p := obj(v)
		for _, n := range names {
			if n == str(p["id"]) {
				if es := ProviderIntegrity(f, "skills/", p); len(es) > 0 {
					return nil, fmt.Errorf("%s", strings.Join(es, "; "))
				}
			}
		}
	}
	target, e = rootPath(target)
	if e != nil {
		return nil, e
	}
	for _, n := range names {
		if _, err := os.Lstat(filepath.Join(target, n)); !os.IsNotExist(err) {
			return nil, fmt.Errorf("Refusing to overwrite: %s", n)
		}
	}
	runtimeDir := filepath.Join(target, ".amazing-skills")
	if exists(runtimeDir) {
		return nil, fmt.Errorf("Runtime already exists; use a fresh target or explicitly manage existing installation")
	}
	if !isFile(executable) {
		return nil, fmt.Errorf("Runtime executable missing")
	}
	if e = os.MkdirAll(target, 0755); e != nil {
		return nil, e
	}
	stage, e := os.MkdirTemp(target, ".amazing-stage-")
	if e != nil {
		return nil, e
	}
	defer os.RemoveAll(stage)
	for _, n := range names {
		e = fs.WalkDir(f, "skills/"+n, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Name() == "__pycache__" {
				return fs.SkipDir
			}
			if d.Name() == ".DS_Store" {
				return nil
			}
			rel := strings.TrimPrefix(p, "skills/")
			dest := filepath.Join(stage, filepath.FromSlash(rel))
			if d.IsDir() {
				return os.MkdirAll(dest, 0755)
			}
			b, e := fs.ReadFile(f, p)
			if e != nil {
				return e
			}
			if e = writeBytes(dest, b, true); e != nil {
				return e
			}
			info, e := d.Info()
			if (e == nil && info.Mode()&0111 != 0) || (n == "impeccable" && strings.HasSuffix(p, "/scripts/impeccable")) {
				_ = os.Chmod(dest, 0755)
			}
			return nil
		})
		if e != nil {
			return nil, e
		}
	}
	data, e := os.ReadFile(executable)
	if e != nil {
		return nil, e
	}
	runtimeName := filepath.Base(executable)
	if strings.HasSuffix(strings.ToLower(runtimeName), ".exe") {
		runtimeName = "amazing-skills.exe"
	} else {
		runtimeName = "amazing-skills"
	}
	if e = writeBytes(filepath.Join(stage, ".amazing-skills", runtimeName), data, true); e != nil {
		return nil, e
	}
	_ = os.Chmod(filepath.Join(stage, ".amazing-skills", runtimeName), 0755)
	published := []string{}
	for _, n := range append(names, ".amazing-skills") {
		dest := filepath.Join(target, n)
		if _, err := os.Lstat(dest); !os.IsNotExist(err) {
			e = fmt.Errorf("Target appeared: %s", n)
		} else {
			e = os.Rename(filepath.Join(stage, n), dest)
		}
		if e != nil {
			for _, p := range published {
				_ = os.RemoveAll(p)
			}
			return nil, e
		}
		published = append(published, dest)
	}
	return names, nil
}
func Validate(f fs.FS) []string {
	errs := []string{}
	e := safeCall(func() error {
		c, err := Catalog(f)
		if err != nil {
			return err
		}
		for _, id := range []string{"calm-paper", "precise-console"} {
			for _, message := range DesignSystemCheck(f, designAssetRoot+"systems/"+id+"/") {
				errs = append(errs, "Design system "+id+": "+message)
			}
		}
		entries := index(c["skills"], "skills", &errs)
		for _, n := range []string{"skill-router", "product-workflow"} {
			v, err := load(f, "skills/"+n+"/references/catalog.json")
			if err != nil || string(canonical(v)) != string(canonical(c)) {
				errs = append(errs, n+": catalog snapshot stale")
			}
		}
		routing := obj(c["entrypoints"])
		defaultID := str(routing["default"])
		defaults := []string{}
		domains := map[string]bool{}
		for _, v := range obj(routing["domains"]) {
			domains[str(v)] = true
		}
		for id, s := range entries {
			level := str(s["level"])
			if level == "entry" {
				defaults = append(defaults, id)
			}
			if level == "domain-workflow" && !domains[id] {
				errs = append(errs, "Unregistered domain workflow: "+id)
			}
			if !member(level, "entry", "domain-workflow", "capability", "provider") || !present(s["display_name"]) {
				errs = append(errs, "Invalid display metadata: "+id)
			}
			if !member(s["kind"], "workflow", "capability") || !has(c["categories"], str(s["category"])) {
				errs = append(errs, "Invalid category/kind: "+id)
			}
			for _, key := range []string{"tags", "inputs", "outputs", "requires", "side_effects"} {
				if _, ok := s[key].([]any); !ok {
					errs = append(errs, "Invalid list: "+id+"/"+key)
				}
			}
			if str(s["path"]) != id+"/SKILL.md" {
				errs = append(errs, "Invalid path: "+id)
			}
			b, err := fs.ReadFile(f, "skills/"+id+"/SKILL.md")
			if err != nil || !strings.HasPrefix(string(b), "---\n") || !strings.Contains(string(b), "name: "+id+"\n") || !regexp.MustCompile(`(?m)^description: .+`).Match(b) {
				errs = append(errs, "Invalid frontmatter: "+id)
			}
			if _, err = Closure(c, []string{id}); err != nil {
				errs = append(errs, err.Error())
			}
		}
		if len(defaults) != 1 || defaults[0] != defaultID {
			errs = append(errs, "Invalid default entrypoint")
		}
		for id := range domains {
			if entries[id]["level"] != "domain-workflow" || entries[id]["kind"] != "workflow" {
				errs = append(errs, "Invalid domain entrypoint: "+id)
			}
		}
		dirs, err := fs.ReadDir(f, "skills")
		if err != nil {
			return err
		}
		for _, d := range dirs {
			if d.IsDir() {
				if _, err := fs.Stat(f, "skills/"+d.Name()+"/SKILL.md"); err == nil {
					if _, ok := entries[d.Name()]; !ok {
						errs = append(errs, "Unregistered skill: "+d.Name())
					}
				}
			}
		}
		lock, err := load(f, "providers.lock.json")
		if err != nil {
			return err
		}
		snapshot, err := load(f, "skills/skill-router/references/providers.lock.json")
		if err != nil || string(canonical(lock)) != string(canonical(snapshot)) {
			errs = append(errs, "Provider snapshot stale")
		}
		upstream := map[string]bool{}
		for _, v := range arr(obj(lock)["providers"]) {
			p := obj(v)
			id := str(p["id"])
			upstream[id] = true
			if entries[id]["revision"] != p["revision"] {
				errs = append(errs, "Provider revision mismatch: "+id)
			}
			errs = append(errs, ProviderIntegrity(f, "skills/", p)...)
			files := obj(p["files"])
			folder := "skills/" + str(p["installed_directory"])
			_ = fs.WalkDir(f, folder, func(name string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.Name() == "__pycache__" {
					return fs.SkipDir
				}
				if !d.IsDir() {
					rel := strings.TrimPrefix(name, folder+"/")
					if _, ok := files[rel]; !ok {
						errs = append(errs, "Unexpected provider file: "+name)
					}
				}
				return nil
			})
		}
		for id, s := range entries {
			if (s["origin"] == "upstream") != upstream[id] {
				errs = append(errs, "Provider catalog/lock differ: "+id)
			}
		}
		re := regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)
		_ = fs.WalkDir(f, "skills", func(name string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(name, ".md") {
				parts := strings.Split(name, "/")
				if upstream[parts[1]] {
					return nil
				}
				b, _ := fs.ReadFile(f, name)
				for _, m := range re.FindAllStringSubmatch(string(b), -1) {
					link := m[1]
					if strings.Contains(link, "://") || strings.HasPrefix(link, "#") {
						continue
					}
					link = strings.Split(link, "#")[0]
					if link != "" {
						rel := filepath.ToSlash(filepath.Clean(filepath.Join(filepath.Dir(name), link)))
						if _, err := fs.Stat(f, rel); err != nil {
							errs = append(errs, "Broken reference: "+name+" -> "+link)
						}
					}
				}
			}
			return nil
		})
		for _, v := range arr(c["external_capabilities"]) {
			if _, err := fs.Stat(f, "skills/"+str(obj(v)["adapter"])); err != nil {
				errs = append(errs, "Missing adapter")
			}
		}
		sources, err := load(f, "qa-sources.lock.json")
		if err != nil {
			return err
		}
		snap, err := load(f, "skills/qa-test-design/references/sources.lock.json")
		if err != nil || string(canonical(sources)) != string(canonical(snap)) {
			errs = append(errs, "QA sources snapshot stale")
		}
		for _, v := range arr(obj(sources)["sources"]) {
			s := obj(v)
			if !member(s["license"], "MIT", "Apache-2.0") || !regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(str(s["revision"])) {
				errs = append(errs, "Invalid QA license/revision")
			}
			for _, name := range []string{str(s["license_file"]), "skills/qa-test-design/references/licenses/" + filepath.Base(str(s["license_file"]))} {
				b, e := fs.ReadFile(f, name)
				if e != nil || hash(b) != str(s["license_sha256"]) {
					errs = append(errs, "QA license drift: "+name)
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
func Sync(root string) error {
	for src, dests := range map[string][]string{"catalog.json": {"skills/skill-router/references/catalog.json", "skills/product-workflow/references/catalog.json"}, "providers.lock.json": {"skills/skill-router/references/providers.lock.json"}, "qa-sources.lock.json": {"skills/qa-test-design/references/sources.lock.json"}} {
		b, e := os.ReadFile(filepath.Join(root, src))
		if e != nil {
			return e
		}
		for _, dest := range dests {
			if e = writeBytes(filepath.Join(root, dest), b, false); e != nil {
				return e
			}
		}
	}
	v, e := read(filepath.Join(root, "qa-sources.lock.json"))
	if e != nil {
		return e
	}
	for _, x := range arr(obj(v)["sources"]) {
		p := str(obj(x)["license_file"])
		b, e := os.ReadFile(filepath.Join(root, p))
		if e != nil {
			return e
		}
		if e = writeBytes(filepath.Join(root, "skills/qa-test-design/references/licenses", filepath.Base(p)), b, false); e != nil {
			return e
		}
	}
	return nil
}
