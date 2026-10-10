package amazing

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type options struct {
	Values map[string]string
	Args   []string
}

func parse(args []string) (options, error) {
	o := options{map[string]string{}, []string{}}
	aliases := map[string]string{"-d": "--domain", "-s": "--stack", "-n": "--max-results", "-p": "--project-name", "-ds": "--design-system", "-f": "--format", "-o": "--output-dir"}
	booleans := map[string]bool{"json": true, "full": true, "design-system": true, "replace": true, "persist": true, "force": true, "diagnostics": true}
	for i := 0; i < len(args); i++ {
		s := args[i]
		if a, ok := aliases[s]; ok {
			s = a
		}
		if strings.HasPrefix(s, "--") {
			name := strings.TrimPrefix(s, "--")
			if name == "" {
				o.Args = append(o.Args, args[i+1:]...)
				break
			}
			if strings.Contains(name, "=") {
				parts := strings.SplitN(name, "=", 2)
				o.Values[parts[0]] = parts[1]
				continue
			}
			if booleans[name] {
				o.Values[name] = "true"
				continue
			}
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "--") {
				return o, fmt.Errorf("Missing value for %s", s)
			}
			i++
			o.Values[name] = args[i]
		} else {
			o.Args = append(o.Args, s)
		}
	}
	return o, nil
}
func NativeSearchCLI(skills string, args []string) (M, error) {
	o, e := parse(args)
	if e != nil {
		return nil, e
	}
	if len(o.Args) != 1 {
		return nil, fmt.Errorf("search requires one quoted query")
	}
	for k := range o.Values {
		if !member(k, "domain", "stack", "max-results", "json", "full", "design-system", "project-name", "format", "variance", "motion", "density", "persist", "page", "output-dir", "force", "diagnostics") {
			return nil, fmt.Errorf("Unsupported search option: %s", k)
		}
	}
	if o.Values["design-system"] == "true" {

		dials := M{}
		for _, k := range []string{"density", "motion", "variance"} {
			if s := o.Values[k]; s != "" {
				n, e := strconv.Atoi(s)
				if e != nil {
					return nil, e
				}
				dials[k] = n
			}
		}
		format := o.Values["format"]
		if format == "" {
			format = "ascii"
		}
		if format != "ascii" && format != "markdown" {
			return nil, fmt.Errorf("format must be ascii or markdown")
		}
		return DesignSystemResult(skills, o.Args[0], o.Values["project-name"], dials, format, o.Values["persist"] == "true", o.Values["page"], o.Values["output-dir"], o.Values["force"] == "true")
	}
	if o.Values["persist"] != "" || o.Values["page"] != "" || o.Values["force"] != "" || o.Values["output-dir"] != "" {
		return nil, fmt.Errorf("persistence requires --design-system")
	}
	n := 3
	if s := o.Values["max-results"]; s != "" {
		n, e = strconv.Atoi(s)
		if e != nil {
			return nil, e
		}
	}
	return SearchDetailed(skills, o.Args[0], o.Values["domain"], o.Values["stack"], n, o.Values["diagnostics"] == "true")
}
func defaultSkills() string {
	exe, _ := os.Executable()
	if filepath.Base(filepath.Dir(exe)) == ".amazing-skills" {
		return filepath.Dir(filepath.Dir(exe))
	}
	return "skills"
}
func CLI(args []string, out io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		fmt.Fprintln(out, `Amazing Skills — Go runtime (no Python required)
Usage: amazing-skills <command> [options]
  install --target <skills-directory> [--skills <comma-separated IDs>]
  validate [--repo <repository>]
  sync [--repo <repository>]
  build --output <directory> [--repo <repository>]  (requires Go compiler)
  project init|check|lock-design|check-design --project <path> [--tokens styles.css]
  qa init|check|snapshot|impact|record|summary|export-cases --project <path>
  jira draft|check-receipt --project <path> --input <json>
  providers doctor|setup-engine|run [--skills-root <directory>] [--provider <id>]
  search <quoted query> [--domain chart|style|color|ux|...] [--stack react]
  serve --example customer-followup|project-workbench [--port 8765]
  security-replay --output <fresh-directory>  (synthetic security model only)
  replay --output <fresh project directory>
Use --skills-root after installing only where auto-discovery is insufficient.
Skills are model instructions; tool success does not prove product acceptance.`)
		return 0
	}
	command := args[0]
	rest := args[1:]
	action := ""
	if member(command, "project", "qa", "jira", "providers") {
		if len(rest) == 0 {
			fmt.Fprintln(out, "Missing subcommand")
			return 1
		}
		action = rest[0]
		rest = rest[1:]
	}
	toolArgs := []string{}
	if command == "providers" && action == "run" {
		for i, s := range rest {
			if s == "--" {
				toolArgs = rest[i+1:]
				rest = rest[:i]
				break
			}
		}
	}
	o, e := parse(rest)
	if e != nil {
		fmt.Fprintln(out, e)
		return 1
	}
	allowed := map[string][]string{
		"install": {"target", "skills"}, "validate": {"repo"}, "sync": {"repo"}, "build": {"repo", "output"},
		"project": {"project", "tokens", "replace"}, "qa": {"project", "label", "before", "input", "run"}, "jira": {"project", "input"},
		"providers": {"project", "provider", "skills-root"}, "serve": {"repo", "example", "port"}, "replay": {"output"}, "security-replay": {"output"},
		"search": {"skills-root", "domain", "stack", "max-results", "json", "full", "design-system", "project-name", "format", "variance", "motion", "density", "persist", "page", "force", "output-dir", "diagnostics"},
	}
	for key := range o.Values {
		if !member(key, allowed[command]...) {
			fmt.Fprintln(out, "Unsupported option: --"+key)
			return 1
		}
	}
	var result any
	exit := 0
	e = safeCall(func() error {
		repo := o.Values["repo"]
		if repo == "" {
			repo = "."
		}
		repo, err := rootPath(repo)
		if err != nil {
			return err
		}
		skills := o.Values["skills-root"]
		if skills == "" {
			skills = defaultSkills()
		}
		skills, err = rootPath(skills)
		if err != nil {
			return err
		}
		project := o.Values["project"]
		if member(command, "project", "qa", "jira") || (command == "providers" && action != "doctor") {
			if project == "" {
				return fmt.Errorf("--project is required")
			}
			project, err = rootPath(project)
			if err != nil {
				return err
			}
		}
		switch command {
		case "install":
			if o.Values["target"] == "" {
				return fmt.Errorf("--target required")
			}
			chosen := []string{}
			if s := o.Values["skills"]; s != "" {
				chosen = strings.Split(s, ",")
			}
			chosen = append(chosen, o.Args...)
			exe, _ := os.Executable()
			names, err := Install(Bundle, o.Values["target"], chosen, exe)
			result = M{"installed": names, "runtime": ".amazing-skills/amazing-skills", "python_required": false}
			return err
		case "validate":
			errs := Validate(os.DirFS(repo))
			status := "passed"
			if len(errs) > 0 {
				status = "failed"
				exit = 1
			}
			result = M{"catalog_and_references": status, "errors": errs}
		case "sync":
			if err := Sync(repo); err != nil {
				return err
			}
			result = M{"catalog_snapshots": "synchronized"}
		case "project":
			switch action {
			case "init":
				created, err := ProjectInit(project)
				result = M{"created": created}
				return err
			case "lock-design":
				r, err := LockDesign(project, o.Values["tokens"], o.Values["replace"] == "true")
				result = r
				return err
			case "check", "check-design":
				errs := ProjectCheck(project)
				if action == "check-design" {
					errs = CheckDesign(project)
				}
				status := "passed"
				if len(errs) > 0 {
					status = "failed"
					exit = 1
				}
				result = M{"documentary_check": status, "errors": errs, "note": "Does not verify implementation or completeness of acceptance coverage."}
			default:
				return fmt.Errorf("Unknown project action")
			}
		case "qa":
			switch action {
			case "init":
				r, err := QAInit(project)
				result = r
				return err
			case "check":
				errs := QACheck(project)
				status := "passed"
				if len(errs) > 0 {
					status = "failed"
					exit = 1
				}
				result = M{"documentary_check": status, "errors": errs, "business_or_execution_verified": false}
			case "snapshot":
				r, err := Snapshot(project, o.Values["label"])
				result = r
				return err
			case "impact":
				r, err := Impact(project, o.Values["before"])
				result = r
				return err
			case "record":
				if o.Values["input"] == "" {
					return fmt.Errorf("--input required")
				}
				v, err := read(o.Values["input"])
				if err != nil {
					return err
				}
				r, err := Record(project, obj(v))
				result = r
				return err
			case "summary":
				r, err := Summary(project, o.Values["run"])
				result = r
				return err
			case "export-cases":
				r, err := ExportCases(project)
				result = r
				return err
			default:
				return fmt.Errorf("Unknown QA action")
			}
		case "jira":
			if o.Values["input"] == "" {
				return fmt.Errorf("--input required")
			}
			v, err := read(o.Values["input"])
			if err != nil {
				return err
			}
			switch action {
			case "draft":
				r, err := Draft(project, obj(v))
				result = r
				return err
			case "check-receipt":
				r := Receipt(project, obj(v))
				result = r
				if r["receipt_check"] != "complete" {
					exit = 1
				}
			default:
				return fmt.Errorf("Unknown Jira action")
			}
		case "providers":
			switch action {
			case "doctor":
				r, err := Doctor(skills, o.Values["provider"])
				result = r
				if err != nil {
					return err
				}
				for _, item := range r {
					if item["integrity"] != "passed" {
						exit = 1
					}
				}
			case "setup-engine":
				r, err := SetupEngine(skills, project)
				result = r
				return err
			case "run":
				r, text, code, err := RunProvider(skills, project, o.Values["provider"], toolArgs)
				if err != nil {
					return err
				}
				fmt.Fprintln(out, string(canonical(r)))
				fmt.Fprint(out, text)
				result = nil
				exit = code
			default:
				return fmt.Errorf("Unknown provider action")
			}
		case "search":
			searchArgs := rest
			filtered := []string{}
			for i := 0; i < len(searchArgs); i++ {
				if searchArgs[i] == "--skills-root" {
					i++
					continue
				}
				filtered = append(filtered, searchArgs[i])
			}
			r, err := NativeSearchCLI(skills, filtered)
			if err == nil && o.Values["json"] != "true" {
				if o.Values["design-system"] == "true" {
					fmt.Fprintln(out, str(r["text"]))
					if r["persistence"] != nil {
						fmt.Fprintln(out, "Persistence:", str(obj(r["persistence"])["status"]))
					}
				} else {
					fmt.Fprintln(out, formatSearchResult(r, o.Values["full"] == "true"))
				}
				return nil
			}
			if o.Values["design-system"] == "true" {
				delete(r, "text")
			}
			result = r
			return err
		case "serve":
			port := o.Values["port"]
			if port == "" {
				port = "8765"
			}
			return Serve(repo, o.Values["example"], port, out)
		case "replay":
			if o.Values["output"] == "" {
				return fmt.Errorf("--output required")
			}
			root, err := rootPath(o.Values["output"])
			if err != nil {
				return err
			}
			r, err := Replay(root)
			result = r
			return err
		case "security-replay":
			if o.Values["output"] == "" {
				return fmt.Errorf("--output required")
			}
			root, err := rootPath(o.Values["output"])
			if err != nil {
				return err
			}
			r, err := SecurityReplay(root)
			result = r
			return err
		case "build":
			if o.Values["output"] == "" {
				return fmt.Errorf("--output required")
			}
			r, err := Build(repo, o.Values["output"])
			result = r
			return err
		default:
			return fmt.Errorf("Unknown command: %s", command)
		}
		return nil
	})
	if e != nil {
		fmt.Fprintln(out, string(canonical(M{"error": e.Error()})))
		return 1
	}
	if result != nil {
		fmt.Fprintln(out, string(canonical(result)))
	}
	return exit
}
func Build(repo, output string) (M, error) {
	dir, e := rootPath(output)
	if e != nil {
		return nil, e
	}
	if e = os.MkdirAll(dir, 0755); e != nil {
		return nil, e
	}
	platforms := []struct{ OS, Arch string }{{"darwin", "arm64"}, {"darwin", "amd64"}, {"linux", "arm64"}, {"linux", "amd64"}, {"windows", "arm64"}, {"windows", "amd64"}}
	files := M{}
	sums := []string{}
	for _, p := range platforms {
		name := "amazing-skills-" + p.OS + "-" + p.Arch
		if p.OS == "windows" {
			name += ".exe"
		}
		dest := filepath.Join(dir, name)
		cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", dest, "./cmd/amazing-skills")
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+p.OS, "GOARCH="+p.Arch)
		out, e := cmd.CombinedOutput()
		if e != nil {
			return nil, fmt.Errorf("Build %s failed: %s", name, string(out))
		}
		h, e := fileHash(dest)
		if e != nil {
			return nil, e
		}
		files[name] = h
		sums = append(sums, h+"  "+name)
	}
	if e = writeBytes(filepath.Join(dir, "SHA256SUMS"), []byte(strings.Join(sums, "\n")+"\n"), false); e != nil {
		return nil, e
	}
	manifest := M{"runtime": "Go standalone binary", "platforms": files, "python_required": false, "cross_compiled_only": "Other operating systems require actual execution validation"}
	e = write(filepath.Join(dir, "manifest.json"), manifest, false)
	return manifest, e
}
