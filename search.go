package amazing

// Search behavior is derived from UI UX Pro Max core.py, pinned in providers.lock.json.
// Copyright upstream contributors; MIT, see skills/ui-ux-pro-max/LICENSE.
import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

var proContract M

func init() {
	if e := json.Unmarshal(mustReadBundle("promax/contract.json"), &proContract); e != nil {
		panic(e)
	}
}
func mustReadBundle(p string) []byte {
	b, e := Bundle.ReadFile(p)
	if e != nil {
		panic(e)
	}
	return b
}
func word(r rune) bool        { return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' }
func phrase(s, p string) bool { return len(phrasePositions(s, p)) > 0 }
func phrasePositions(s, p string) []int {
	a, b := []rune(s), []rune(p)
	out := []int{}
	for i := 0; i+len(b) <= len(a); i++ {
		if string(a[i:i+len(b)]) == p && (i == 0 || !word(a[i-1])) && (i+len(b) == len(a) || !word(a[i+len(b)])) {
			out = append(out, i)
		}
	}
	return out
}
func normalize(s string) string {
	for _, pair := range arr(proContract["synonyms"]) {
		v := arr(pair)
		p := str(v[0])
		a := []rune(s)
		positions := phrasePositions(strings.ToLower(s), p)
		for i := len(positions) - 1; i >= 0; i-- {
			n := positions[i]
			a = append(append(append([]rune{}, a[:n]...), []rune(str(v[1]))...), a[n+len([]rune(p)):]...)
		}
		s = string(a)
	}
	return s
}
func rawTokens(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return !word(r) })
}
func tokens(s string) []string {
	stop := promaxStrings(proContract["stopwords"])
	out := []string{}
	for _, t := range rawTokens(normalize(strings.ToLower(s))) {
		if len([]rune(t)) >= 2 && !member(t, stop...) {
			out = append(out, t)
		}
	}
	return out
}

type csvSnapshot struct {
	size     int64
	modified int64
	rows     []M
}

var csvSnapshots = map[string]csvSnapshot{}
var searchCacheMu sync.Mutex
var indexSnapshots = map[[32]byte]*bmIndex{}

func readCSV(root, file string) ([]M, error) {
	p, e := Inside(root, "ui-ux-pro-max/data/"+file)
	if e != nil {
		return nil, e
	}
	info, e := os.Stat(p)
	if e != nil {
		return nil, e
	}
	searchCacheMu.Lock()
	snapshot, cached := csvSnapshots[p]
	searchCacheMu.Unlock()
	if cached && snapshot.size == info.Size() && snapshot.modified == info.ModTime().UnixNano() {
		return snapshot.rows, nil
	}
	f, e := openFile(p)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	header, e := r.Read()
	if e != nil {
		return nil, e
	}
	rows := []M{}
	for {
		v, e := r.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		if len(v) != len(header) {
			return nil, fmt.Errorf("invalid CSV width")
		}
		row := M{}
		for i, k := range header {
			row[k] = v[i]
		}
		rows = append(rows, row)
	}
	searchCacheMu.Lock()
	csvSnapshots[p] = csvSnapshot{info.Size(), info.ModTime().UnixNano(), rows}
	searchCacheMu.Unlock()
	return rows, nil
}
func projectRow(row M, cols []string) M {
	m := M{}
	for _, k := range cols {
		if v, ok := row[k]; ok {
			m[k] = v
		}
	}
	return m
}
func identities(r M, kind string) []string {
	fields := []string{"Style ID", "Style Category", "Aliases"}
	if kind == "landing" {
		fields = []string{"Pattern ID", "Pattern Name", "Aliases"}
	}
	out := []string{}
	for _, k := range fields {
		v := []string{str(r[k])}
		if k == "Aliases" {
			v = strings.Split(str(r[k]), "|")
		}
		for _, s := range v {
			if s = strings.TrimSpace(s); s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}
func set(ts []string) map[string]bool {
	m := map[string]bool{}
	for _, t := range ts {
		m[t] = true
	}
	return m
}
func subset(a, b map[string]bool) bool {
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}
func intersects(a, b map[string]bool) bool {
	for k := range a {
		if b[k] {
			return true
		}
	}
	return false
}
func styleIdentity(rows []M, q string, contained bool) M {
	fold := strings.ToLower(strings.TrimSpace(q))
	qt := set(rawTokens(normalize(fold)))
	type candidate struct {
		a, b, c int
		r       M
	}
	cs := []candidate{}
	generic := set([]string{"app", "design", "interface", "style", "system", "ui"})
	for _, r := range rows {
		for _, id := range identities(r, "style") {
			if strings.ToLower(id) == fold {
				return r
			}
			if !contained {
				continue
			}
			ts := set(rawTokens(normalize(strings.ToLower(id))))
			long := false
			distinct := 0
			for t := range ts {
				if len([]rune(t)) >= 4 {
					long = true
				}
				if !generic[t] {
					distinct++
				}
			}
			if len(ts) > 0 && long && subset(ts, qt) {
				cs = append(cs, candidate{distinct, len(ts), len([]rune(id)), r})
			}
		}
	}
	if len(cs) == 0 {
		return nil
	}
	sort.SliceStable(cs, func(i, j int) bool {
		if cs[i].a != cs[j].a {
			return cs[i].a > cs[j].a
		}
		return cs[i].b > cs[j].b
	})
	best := cs[0]
	ids := map[string]M{}
	for _, c := range cs {
		if c.a == best.a && c.b == best.b && c.c == best.c {
			ids[str(c.r["Style ID"])] = c.r
		}
	}
	if len(ids) == 1 {
		for _, r := range ids {
			return r
		}
	}
	return nil
}
func styleDestination(rows []M, r M) (M, M) {
	if r == nil || str(r["Status"]) != "deprecated" {
		return r, nil
	}
	id := strings.TrimSpace(str(r["Parent Style ID"]))
	domain := strings.TrimSpace(str(r["Replacement Domain"]))
	if id == "" && domain == "style" {
		id = strings.TrimSpace(str(r["Replacement ID"]))
	}
	if id != "" {
		for _, row := range rows {
			if str(row["Style ID"]) == id {
				return row, nil
			}
		}
		return nil, nil
	}
	if domain != "" && str(r["Replacement ID"]) != "" {
		return nil, M{"domain": domain, "id": strings.TrimSpace(str(r["Replacement ID"]))}
	}
	return nil, nil
}
func detectDomain(q string) (string, string) {
	q = normalize(strings.ToLower(q))
	type ds struct {
		d string
		s float64
	}
	scores := []ds{}
	for _, d := range promaxStrings(proContract["order"]) {
		score := 0.
		for _, k := range promaxStrings(obj(proContract["keywords"])[d]) {
			if phrase(q, k) {
				v := float64(len(strings.Fields(k)))
				if d != "product" {
					v *= 2
				}
				score += v
			}
		}
		if d == "color" {
			for _, hex := range regexp.MustCompile(`#[0-9a-f]{3,8}`).FindAllString(q, -1) {
				if phrase(q, hex) {
					score += 2
					break
				}
			}
		}
		scores = append(scores, ds{d, score})
	}
	sort.SliceStable(scores, func(i, j int) bool { return scores[i].s > scores[j].s })
	d, runner := "style", ""
	if scores[0].s > 0 {
		d = scores[0].d
	}
	if scores[1].s > 0 {
		runner = scores[1].d
	}
	return d, runner
}

type scored struct {
	I int
	S float64
}
type bmIndex struct {
	rows []M
	docs []map[string]int
	lens []int
	df   map[string]int
	avg  float64
}

func newIndex(rows []M, cols []string) *bmIndex {
	raw, _ := json.Marshal([]any{rows, cols})
	key := sha256.Sum256(raw)
	searchCacheMu.Lock()
	cached := indexSnapshots[key]
	searchCacheMu.Unlock()
	if cached != nil {
		return cached
	}
	b := &bmIndex{rows: rows, df: map[string]int{}}
	total := 0
	for _, r := range rows {
		text := []string{}
		for _, k := range cols {
			text = append(text, str(r[k]))
		}
		ts := tokens(strings.Join(text, " "))
		tf := map[string]int{}
		for _, t := range ts {
			tf[t]++
		}
		for t := range tf {
			b.df[t]++
		}
		b.docs = append(b.docs, tf)
		b.lens = append(b.lens, len(ts))
		total += len(ts)
	}
	b.avg = 1
	if len(rows) > 0 && total > 0 {
		b.avg = float64(total) / float64(len(rows))
	}
	searchCacheMu.Lock()
	indexSnapshots[key] = b
	searchCacheMu.Unlock()
	return b
}
func (b *bmIndex) rank(q string) []scored {
	out := []scored{}
	N := float64(len(b.rows))
	for i, tf := range b.docs {
		s := 0.
		for _, t := range tokens(q) {
			if df, ok := b.df[t]; ok {
				n := float64(tf[t])
				idf := math.Log((N-float64(df)+.5)/(float64(df)+.5) + 1)
				s += idf * (n * 2.5) / (n + 1.5*(.25+.75*float64(b.lens[i])/b.avg))
			}
		}
		out = append(out, scored{i, s})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].S > out[j].S })
	return out
}
func (b *bmIndex) coverage(q string) float64 {
	ts := set(tokens(q))
	if len(ts) == 0 {
		return 0
	}
	n := 0
	for t := range ts {
		if b.df[t] > 0 {
			n++
		}
	}
	return float64(n) / float64(len(ts))
}
func (b *bmIndex) passes(q string, threshold M) bool {
	rank := b.rank(q)
	if len(rank) == 0 {
		return false
	}
	return rank[0].S > float64num(threshold["min_score"]) && b.coverage(q) >= float64num(threshold["min_coverage"])
}
func float64num(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case json.Number:
		x, _ := n.Float64()
		return x
	}
	return 0
}

// SequenceMatcher's longest contiguous block recursion, preserving its earliest-position tie break.
func similarity(x, y string) float64 {
	a, b := []rune(x), []rune(y)
	if len(a)+len(b) == 0 {
		return 1
	}
	var matches func(int, int, int, int) int
	matches = func(al, ah, bl, bh int) int {
		ai, bi, n := al, bl, 0
		prev := map[int]int{}
		for i := al; i < ah; i++ {
			next := map[int]int{}
			for j := bl; j < bh; j++ {
				if a[i] == b[j] {
					k := prev[j-1] + 1
					next[j] = k
					if k > n {
						ai, bi, n = i-k+1, j-k+1, k
					}
				}
			}
			prev = next
		}
		if n == 0 {
			return 0
		}
		return n + matches(al, ai, bl, bi) + matches(ai+n, ah, bi+n, bh)
	}
	return 2 * float64(matches(0, len(a), 0, len(b))) / float64(len(a)+len(b))
}
func suggestions(b *bmIndex, q string, threshold M) []string {
	out := []string{}
	if b == nil {
		return out
	}
	qt := set(tokens(q))
	type c struct {
		t  string
		s  float64
		df int
	}
	cs := []c{}
	for t, df := range b.df {
		if qt[t] {
			continue
		}
		s := 0.
		for k := range qt {
			s = math.Max(s, similarity(k, t))
		}
		if s >= .72 && b.passes(t, threshold) {
			cs = append(cs, c{t, s, df})
		}
	}
	sort.Slice(cs, func(i, j int) bool {
		if cs[i].s != cs[j].s {
			return cs[i].s > cs[j].s
		}
		if cs[i].df != cs[j].df {
			return cs[i].df > cs[j].df
		}
		return cs[i].t < cs[j].t
	})
	for i, c := range cs {
		if i == 6 {
			break
		}
		out = append(out, c.t)
	}
	return out
}
func identitySuggestions(rows []M, q string) []string {
	type c struct {
		s float64
		n int
		t string
	}
	cs := []c{}
	seen := map[string]bool{}
	for _, r := range rows {
		for _, id := range identities(r, "landing") {
			s := 0.
			it := set(tokens(id))
			for a := range set(tokens(q)) {
				for b := range it {
					s = math.Max(s, similarity(a, b))
				}
			}
			if s >= .72 && strings.ToLower(id) != strings.ToLower(strings.TrimSpace(q)) && !seen[id] {
				cs = append(cs, c{s, len(it), id})
				seen[id] = true
			}
		}
	}
	sort.Slice(cs, func(i, j int) bool {
		if cs[i].s != cs[j].s {
			return cs[i].s > cs[j].s
		}
		if cs[i].n != cs[j].n {
			return cs[i].n < cs[j].n
		}
		return cs[i].t < cs[j].t
	})
	out := []string{}
	for i, c := range cs {
		if i == 6 {
			break
		}
		out = append(out, c.t)
	}
	return out
}
func exactDiagnostic(q, reason string) M {
	return M{"normalized_query": normalize(q), "search_query": q, "query_rewrites": []string{}, "top_score": 0., "runner_up_score": 0., "margin": 0., "token_coverage": 1., "abstained": false, "calibration_version": "2026-08-12-v1", "reason": reason}
}
func detailedSearch(rows []M, cols, outCols []string, q, domain string, n int, threshold M) ([]any, *bmIndex, M) {
	out := []any{}
	if len(rows) == 0 {
		return out, nil, M{"reason": "empty-data"}
	}
	b := newIndex(rows, cols)
	rewrites, replacements := []string{}, []string{}
	for _, k := range promaxStrings(obj(proContract["keywords"])[domain]) {
		if !phrase(normalize(strings.ToLower(q)), k) {
			continue
		}
		known := false
		for _, t := range tokens(k) {
			if b.df[t] > 0 {
				known = true
			}
		}
		replacement := str(obj(obj(proContract["rewrites"])[domain])[k])
		if !known && replacement != "" {
			rewrites = append(rewrites, k+"->"+replacement)
			replacements = append(replacements, replacement)
		}
	}
	sq := q
	if len(replacements) > 0 {
		sq += " " + strings.Join(dedup(replacements), " ")
	}
	rank := b.rank(sq)
	top, runner := 0., 0.
	if len(rank) > 0 {
		top = rank[0].S
	}
	if len(rank) > 1 {
		runner = rank[1].S
	}
	coverage := b.coverage(sq)
	abstain := top <= float64num(threshold["min_score"]) || coverage < float64num(threshold["min_coverage"])
	if !abstain {
		for i, s := range rank {
			if i >= n {
				break
			}
			if s.S > 0 {
				out = append(out, projectRow(rows[s.I], outCols))
			}
		}
	}
	reason := "matched"
	if abstain {
		reason = "low-confidence"
	}
	return out, b, M{"normalized_query": normalize(q), "search_query": sq, "query_rewrites": dedup(rewrites), "top_score": top, "runner_up_score": runner, "margin": top - runner, "token_coverage": coverage, "abstained": abstain, "calibration_version": "2026-08-12-v1", "reason": reason}
}
func Search(root, q, domain, stack string, n int) (M, error) {
	return SearchDetailed(root, q, domain, stack, n, false)
}
func SearchDetailed(root, q, domain, stack string, n int, diagnostics bool) (M, error) {
	if n < 1 || n > 20 {
		return nil, fmt.Errorf("max-results must be 1..20")
	}
	if stack != "" {
		return searchStack(root, q, stack, n, diagnostics)
	}
	auto := domain == ""
	runner := ""
	styleRows, _ := readCSV(root, "styles.csv")
	var exact, redirect M
	if auto {
		matched := styleIdentity(styleRows, q, false)
		if matched != nil {
			domain = "style"
			exact, redirect = styleDestination(styleRows, matched)
		} else {
			domain, runner = detectDomain(q)
		}
	}
	searchDomain := domain
	config := obj(obj(proContract["csv"])[searchDomain])
	if len(config) == 0 {
		return nil, fmt.Errorf("Unknown search domain: %s", domain)
	}
	file := str(config["file"])
	rows, e := readCSV(root, file)
	if e != nil {
		return nil, e
	}
	if domain == "style" && exact == nil && redirect == nil {
		exact, redirect = styleDestination(styleRows, styleIdentity(styleRows, q, true))
	} else if domain == "landing" {
		matches := []M{}
		for _, r := range rows {
			for _, id := range identities(r, "landing") {
				if strings.EqualFold(strings.TrimSpace(q), id) {
					matches = append(matches, r)
					break
				}
			}
		}
		if len(matches) == 1 {
			exact = matches[0]
		}
	}
	outCols := promaxStrings(config["output_cols"])
	threshold := obj(obj(proContract["thresholds"])[domain])
	results := []any{}
	var b *bmIndex
	var diag M
	if exact != nil {
		results = append(results, projectRow(exact, outCols))
		diag = exactDiagnostic(q, "exact-identity")
	} else if redirect != nil {
		diag = M{"normalized_query": normalize(q), "search_query": q, "query_rewrites": []string{}, "abstained": true, "calibration_version": "2026-08-12-v1", "reason": "cross-domain-redirect"}
	} else {
		filtered := rows
		if domain == "style" {
			filtered = []M{}
			for _, r := range rows {
				if str(r["Status"]) == "active" {
					filtered = append(filtered, r)
				}
			}
		}
		results, b, diag = detailedSearch(filtered, promaxStrings(config["search_cols"]), outCols, q, domain, n, threshold)
	}
	if domain == "icons" && phrase(normalize(strings.ToLower(q)), "lucide") {
		results = []any{}
		diag["abstained"] = true
		diag["reason"] = "unsupported-library"
	}
	out := M{"domain": domain, "query": q, "file": file, "count": len(results), "results": results}
	if auto {
		out["auto_detected"] = true
		if runner != "" {
			out["runner_up_domain"] = runner
		}
	}
	if redirect != nil {
		out["redirect"] = redirect
	}
	if len(results) == 0 {
		if domain == "landing" {
			out["suggestions"] = identitySuggestions(rows, q)
		} else {
			out["suggestions"] = suggestions(b, q, threshold)
		}
	}
	if diagnostics {
		out["diagnostics"] = diag
	}
	return out, nil
}
func regexpID(s string) bool { return regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(s) }
func legacyQuery(q, stack string) bool {
	if stack == "uwp" {
		return true
	}
	q = normalize(strings.ToLower(q))
	migration := regexp.MustCompile(`\b(migrat\w*|upgrad\w*|replac\w*|instead|modern|current)\b`).MatchString(q)
	cur := arr(obj(proContract["current"])[stack])
	name := str(obj(proContract["stack_names"])[stack])
	versions := [][]int{}
	if len(cur) > 0 && name != "" {
		pattern := `\b(?:` + name + `)\s*(?:sdk|ui)?\s*(?:[@(]\s*)?(?:v(?:ersion)?\s*)?(\d+)(?:\.(\d+))?\s*\)?`
		for _, m := range regexp.MustCompile(pattern).FindAllStringSubmatch(q, -1) {
			v := []int{}
			for _, s := range m[1:] {
				if s != "" {
					i, _ := strconv.Atoi(s)
					v = append(v, i)
				}
			}
			versions = append(versions, v)
		}
		if stack == "threejs" {
			for _, m := range regexp.MustCompile(`\br(\d+)\b`).FindAllStringSubmatch(q, -1) {
				i, _ := strconv.Atoi(m[1])
				versions = append(versions, []int{0, i})
			}
		}
		if len(versions) > 0 {
			allOlder := true
			for _, v := range versions {
				older := false
				for i, n := range v {
					if i >= len(cur) {
						break
					}
					c := int(float64num(cur[i]))
					if n != c {
						older = n < c
						break
					}
				}
				if !older {
					allOlder = false
				}
			}
			if migration && !allOlder {
				return false
			}
			return allOlder
		}
	}
	if migration {
		return false
	}
	return regexp.MustCompile(`\b(legacy|deprecated)\b`).MatchString(q)
}
func searchStack(root, q, stack string, n int, diagnostics bool) (M, error) {
	config := obj(obj(proContract["stack"])[stack])
	if len(config) == 0 {
		return nil, fmt.Errorf("Unknown stack: %s", stack)
	}
	file := str(config["file"])
	rows, e := readCSV(root, file)
	if e != nil {
		return nil, e
	}
	statuses := map[string]bool{}
	for _, r := range rows {
		statuses[str(r["Status"])] = true
	}
	legacy := legacyQuery(q, stack)
	variant := "non-legacy"
	status := ""
	if legacy && statuses["deprecated"] {
		status = "deprecated"
		variant = "legacy-only"
	} else if legacy && obj(proContract["current"])[stack] != nil {
		status = "unavailable"
		variant = "legacy-unavailable"
	} else if statuses["active"] {
		status = "active"
		variant = "current-only"
	}
	base := ""
	lower := normalize(strings.ToLower(q))
	if stack == "shadcn" {
		if strings.Contains(lower, "base ui") {
			base = "base"
		} else if strings.Contains(lower, "react aria") {
			base = "aria"
		} else if strings.Contains(lower, "radix") || strings.Contains(lower, "aschild") {
			base = "radix"
		}
	}
	filtered := []M{}
	for _, r := range rows {
		st := str(r["Status"])
		if status != "" && st != status || status == "" && st == "deprecated" {
			continue
		}
		if base != "" {
			m := regexp.MustCompile(`\bbase=([^;]+)`).FindStringSubmatch(strings.ToLower(str(r["Applies To"])))
			if len(m) < 2 || !member(base, strings.Split(m[1], "|")...) {
				continue
			}
		}
		filtered = append(filtered, r)
	}
	threshold := obj(proContract["stack_threshold"])
	if variant == "legacy-only" && base == "" {
		threshold = M{}
	}
	cols := obj(proContract["stack_cols"])
	outCols := promaxStrings(cols["output_cols"])
	var exact M
	matches := []M{}
	if stack == "uwp" && regexp.MustCompile(`\b(?:brand new|new)\s+(?:app|application|project)\b`).MatchString(lower) {
		for _, r := range filtered {
			t := strings.ToLower(str(r["Guideline"]) + " " + str(r["Description"]) + " " + str(r["Do"]))
			if regexp.MustCompile(`\b(?:prefer|choose|use)\b.*\bnew (?:apps?|projects?)\b`).MatchString(t) {
				matches = append(matches, r)
			}
		}
		if len(matches) == 1 {
			exact = matches[0]
		}
	}
	id := strings.TrimSpace(q)
	if exact == nil && len([]rune(id)) >= 6 && !strings.ContainsFunc(id, unicode.IsSpace) {
		matches = []M{}
		re := regexp.MustCompile(`(?i)(^|[^A-Za-z0-9_])` + regexp.QuoteMeta(id) + `($|[^A-Za-z0-9_])`)
		for _, r := range filtered {
			for _, k := range []string{"Guideline", "Description", "Do", "Don't", "Code Good", "Code Bad"} {
				if re.MatchString(str(r[k])) {
					matches = append(matches, r)
					break
				}
			}
		}
		if len(matches) == 1 {
			exact = matches[0]
		}
	}
	results := []any{}
	var b *bmIndex
	var diag M
	if exact != nil {
		results = append(results, projectRow(exact, outCols))
		diag = exactDiagnostic(q, "exact-identifier")
	} else {
		results, b, diag = detailedSearch(filtered, promaxStrings(cols["search_cols"]), outCols, q, "", n, threshold)
	}
	out := M{"domain": "stack", "stack": stack, "query": q, "file": file, "count": len(results), "results": results}
	if len(results) == 0 {
		out["suggestions"] = suggestions(b, q, threshold)
	}
	if diagnostics {
		out["diagnostics"] = diag
	}
	return out, nil
}
func promaxStrings(v any) []string {
	out := []string{}
	for _, s := range arr(v) {
		out = append(out, str(s))
	}
	return out
}
