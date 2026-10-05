// Package vocabulary holds the words Planton uses for its own concepts and
// the spellings it does not use, and finds those spellings in a repository.
//
// One idea has one word, the same in the console, the CLI, the API, the docs
// and the code. vocabulary.yaml is the single statement of those words: the
// current ones with their meanings, the spellings Planton does not use with
// what to write instead, and the narrow allowances for words that belong to
// someone else (a vendor's API or type, a vendor's CLI).
//
// A retired spelling is matched case-insensitively in every case form and
// anywhere in a line, so the entry CloudResource catches CloudResourceKind,
// cloud_resource_id, CLOUD_RESOURCE, cloud-resource, cloudresourcekind and
// "Cloud Resource" alike. A match inside an allowance span on the same line
// is excused; every other match is a Finding.
//
// The CI lane is .github/workflows/lint.vocabulary.yaml, which runs this
// package's tests; TestRetiredVocabularyGate walks every tracked file.
package vocabulary

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

//go:embed vocabulary.yaml
var vocabularyYAML []byte

// Vocabulary is vocabulary.yaml, parsed.
type Vocabulary struct {
	Words   []Word      `yaml:"words"`
	Retired []Retired   `yaml:"retired"`
	Phrases []Phrase    `yaml:"phrases"`
	Allow   []Allowance `yaml:"allow"`
	Exclude []string    `yaml:"exclude"`
}

// Word is a current word and what it means.
type Word struct {
	Name  string `yaml:"name"`
	Means string `yaml:"means"`
}

// Retired is a spelling Planton does not use, in its PascalCase form (or a
// single lowercase token), and what to write instead.
type Retired struct {
	Spelling string `yaml:"spelling"`
	Use      string `yaml:"use"`
}

// Phrase is a retired name or shape, as an RE2 pattern. Anchors are
// lowercase literals of which every match contains at least one; the scan
// reads a line with the pattern only when the lowercased line contains an
// anchor. A space in an anchor stands for an optional separator (none, a
// space, _ or -).
type Phrase struct {
	Pattern string   `yaml:"pattern"`
	Anchors []string `yaml:"anchors"`
	Use     string   `yaml:"use"`
}

// Allowance is a span that contains a retired spelling but is not a Planton
// word. Paths, when set, scopes it to repository paths matching those globs.
type Allowance struct {
	Pattern string   `yaml:"pattern"`
	Paths   []string `yaml:"paths"`
	Reason  string   `yaml:"reason"`
}

// Finding is one retired spelling at one place.
type Finding struct {
	Path  string
	Line  int
	Match string
	Use   string
}

func (f Finding) String() string {
	return fmt.Sprintf("%s:%d: %q -- use %s", f.Path, f.Line, f.Match, f.Use)
}

// Load parses the embedded vocabulary.yaml.
func Load() (*Vocabulary, error) {
	return Parse(vocabularyYAML)
}

// Parse parses a vocabulary document.
func Parse(data []byte) (*Vocabulary, error) {
	var v Vocabulary
	if err := yaml.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("parse vocabulary: %w", err)
	}
	return &v, nil
}

// SpellingPattern turns a retired spelling into the case-insensitive pattern
// that matches it in every case form: CloudResource becomes
// (?i)cloud[ _-]?resource, which matches CloudResource, cloudResource,
// cloud_resource, CLOUD_RESOURCE, cloud-resource, cloudresource and
// "cloud resource".
func SpellingPattern(spelling string) string {
	words := splitWords(spelling)
	quoted := make([]string, len(words))
	for i, w := range words {
		quoted[i] = regexp.QuoteMeta(strings.ToLower(w))
	}
	return "(?i)" + strings.Join(quoted, "[ _-]?")
}

// splitWords splits a PascalCase spelling at each upper-case letter that
// starts a new word: CloudResource -> [Cloud Resource]. A spelling with no
// upper-case letter is one word.
func splitWords(s string) []string {
	var words []string
	start := 0
	runes := []rune(s)
	for i := 1; i < len(runes); i++ {
		if unicode.IsUpper(runes[i]) && !unicode.IsUpper(runes[i-1]) {
			words = append(words, string(runes[start:i]))
			start = i
		}
	}
	return append(words, string(runes[start:]))
}

type rule struct {
	re  *regexp.Regexp
	use string
}

type allowance struct {
	re    *regexp.Regexp
	paths []*regexp.Regexp
}

// Scanner finds retired spellings. Build one with NewScanner.
type Scanner struct {
	rules []rule
	// squeezed holds each retired spelling, and each phrase anchor written
	// with spaces, lowercased with no separators (cloudresource); a line can
	// match one only if its own squeezed form contains it. anchors holds the
	// other phrase anchors, looked for in the lowercased line as written.
	// Checking literals first keeps the scan at memory speed: the patterns
	// run only on the few lines that could match.
	squeezed [][]byte
	anchors  [][]byte
	allow    []allowance
	exclude  []*regexp.Regexp
}

// NewScanner compiles every pattern in the vocabulary.
func (v *Vocabulary) NewScanner() (*Scanner, error) {
	s := &Scanner{}
	for _, r := range v.Retired {
		re, err := regexp.Compile(SpellingPattern(r.Spelling))
		if err != nil {
			return nil, fmt.Errorf("retired %q: %w", r.Spelling, err)
		}
		s.rules = append(s.rules, rule{re: re, use: r.Use})
		s.squeezed = append(s.squeezed, squeeze(bytes.ToLower([]byte(r.Spelling))))
	}
	for _, p := range v.Phrases {
		if len(p.Anchors) == 0 {
			return nil, fmt.Errorf("phrase %q: no anchors", p.Pattern)
		}
		for _, a := range p.Anchors {
			if a != strings.ToLower(a) || a == "" {
				return nil, fmt.Errorf("phrase %q: anchor %q must be a non-empty lowercase literal", p.Pattern, a)
			}
			if strings.Contains(a, " ") {
				s.squeezed = append(s.squeezed, squeeze([]byte(a)))
			} else {
				s.anchors = append(s.anchors, []byte(a))
			}
		}
		re, err := regexp.Compile(p.Pattern)
		if err != nil {
			return nil, fmt.Errorf("phrase %q: %w", p.Pattern, err)
		}
		s.rules = append(s.rules, rule{re: re, use: p.Use})
	}
	for _, a := range v.Allow {
		re, err := regexp.Compile(a.Pattern)
		if err != nil {
			return nil, fmt.Errorf("allow %q: %w", a.Pattern, err)
		}
		al := allowance{re: re}
		for _, g := range a.Paths {
			al.paths = append(al.paths, globRegexp(g))
		}
		s.allow = append(s.allow, al)
	}
	for _, g := range v.Exclude {
		s.exclude = append(s.exclude, globRegexp(g))
	}
	return s, nil
}

// squeeze drops the separators a retired spelling may carry between its
// words, so cloud_resource, cloud-resource and "cloud resource" all read
// cloudresource.
func squeeze(b []byte) []byte {
	out := make([]byte, 0, len(b))
	for _, c := range b {
		if c != ' ' && c != '_' && c != '-' {
			out = append(out, c)
		}
	}
	return out
}

// mayMatch reports whether lowercased text contains any retired spelling's
// or phrase's literal: a cheap necessary condition for a finding.
func (s *Scanner) mayMatch(lower []byte) bool {
	for _, a := range s.anchors {
		if bytes.Contains(lower, a) {
			return true
		}
	}
	sq := squeeze(lower)
	for _, w := range s.squeezed {
		if bytes.Contains(sq, w) {
			return true
		}
	}
	return false
}

// globRegexp compiles a path glob: ** crosses directories (**/ also matches
// none, so **/go.mod matches the root go.mod), * and ? stay within one path
// segment.
func globRegexp(glob string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(glob); i++ {
		switch c := glob[i]; c {
		case '*':
			if strings.HasPrefix(glob[i:], "**/") {
				b.WriteString("(?:.*/)?")
				i += 2
			} else if i+1 < len(glob) && glob[i+1] == '*' {
				b.WriteString(".*")
				i++
			} else {
				b.WriteString("[^/]*")
			}
		case '?':
			b.WriteString("[^/]")
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String())
}

// Excluded reports whether a repository-relative path is never scanned.
func (s *Scanner) Excluded(path string) bool {
	for _, re := range s.exclude {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}

// ScanText returns the findings in one file's content. path is
// repository-relative and slash-separated; it scopes path allowances.
func (s *Scanner) ScanText(path string, content []byte) []Finding {
	content = []byte(lookalikes.Replace(string(content)))
	if !s.mayMatch(bytes.ToLower(wrapJoined.ReplaceAll(content, []byte(" ")))) {
		return nil
	}
	var allow []*regexp.Regexp
	for _, a := range s.allow {
		if len(a.paths) == 0 || matchesAny(a.paths, path) {
			allow = append(allow, a.re)
		}
	}
	var findings []Finding
	lines := strings.Split(string(content), "\n")
	code := isSourceCode(path)
	wraps := func(above, below string) bool {
		if keyLine.MatchString(below) {
			return false
		}
		return !code || (commentLine.MatchString(above) && commentLine.MatchString(below))
	}
	for i, line := range lines {
		// Each line is read in the window of its neighbours, the way a reader
		// reads wrapped prose: a two-word name broken across a line break
		// ("deployment" ending one comment line, "component" opening the
		// next) is found on the line it starts on, and a vendor's phrase
		// wrapped the same way still excuses its words on both lines.
		prev, cur, next := "", line, ""
		if i > 0 && wraps(lines[i-1], line) {
			prev = lines[i-1] + " "
			cur = commentMarker.ReplaceAllString(line, "")
		}
		if i+1 < len(lines) && wraps(line, lines[i+1]) {
			next = " " + commentMarker.ReplaceAllString(lines[i+1], "")
		}
		findings = append(findings, s.scanWindow(path, i+1, prev+cur+next, len(prev), len(prev)+len(cur), allow)...)
	}
	return findings
}

// commentMarker is the indentation and comment leader a wrapped line opens with.
// lookalikes maps the Unicode hyphens and the no-break space to their ASCII
// forms, so a name written with them reads the way the rules spell it.
var lookalikes = strings.NewReplacer("\u2010", "-", "\u2011", "-", "\u2012", "-", "\u00a0", " ")

var commentMarker = regexp.MustCompile(`^\s*(//+|#+|\*|--|;+)?\s*`)

// sourceSuffixes are the source-code files, where a line break separates
// tokens: there only a comment wraps, so two neighbouring lines are read
// together only when both are comments.
var sourceSuffixes = []string{
	".go", ".java", ".kt", ".scala", ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".py", ".rs",
	".dart", ".swift", ".proto", ".sh", ".bash", ".sql", ".cypher", ".bzl", ".bazel", ".tf", ".hcl",
	".c", ".h", ".cc", ".cpp",
}

func isSourceCode(path string) bool {
	for _, s := range sourceSuffixes {
		if strings.HasSuffix(path, s) {
			return true
		}
	}
	return false
}

var commentLine = regexp.MustCompile(`^\s*(//|#|\*|/\*|--|;)`)

// keyLine is a line that opens with a key (YAML, a struct literal): it
// starts a new statement, so it never continues the line above.
var keyLine = regexp.MustCompile(`^\s*(- )?["']?[A-Za-z_][\w.-]*["']?\s*:(\s|$)`)

// wrapJoined is a line break with the next line's comment leader: the file
// pre-check reads it as one space, so a wrapped name still reaches the scan.
var wrapJoined = regexp.MustCompile(`\s*\n\s*(//+|#+|\*|--|;+)?\s*`)

// scanWindow reports the retired spellings that START inside [lo, hi) of
// view -- the current line within its neighbours -- and are not excused by
// an allowance anywhere in the window.
func (s *Scanner) scanWindow(path string, line int, view string, lo, hi int, allow []*regexp.Regexp) []Finding {
	if !s.mayMatch(bytes.ToLower([]byte(view))) {
		return nil
	}
	var excused [][]int
	for _, re := range allow {
		excused = append(excused, re.FindAllStringIndex(view, -1)...)
	}
	var findings []Finding
	seen := map[[2]int]bool{}
	for _, r := range s.rules {
		for _, m := range r.re.FindAllStringIndex(view, -1) {
			span := [2]int{m[0], m[1]}
			if m[0] < lo || m[0] >= hi || seen[span] || within(m, excused) {
				continue
			}
			seen[span] = true
			findings = append(findings, Finding{Path: path, Line: line, Match: view[m[0]:m[1]], Use: r.use})
		}
	}
	return findings
}

func matchesAny(res []*regexp.Regexp, path string) bool {
	for _, re := range res {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}

func within(m []int, spans [][]int) bool {
	for _, sp := range spans {
		if m[0] >= sp[0] && m[1] <= sp[1] {
			return true
		}
	}
	return false
}

// ScanTree scans every file git tracks under root, skipping excluded paths
// and binary files. Findings are sorted by path and line.
func (s *Scanner) ScanTree(root string) ([]Finding, error) {
	cmd := exec.Command("git", "ls-files", "-z")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files in %s: %w", root, err)
	}
	var findings []Finding
	for _, p := range strings.Split(strings.TrimRight(string(out), "\x00"), "\x00") {
		if p == "" || s.Excluded(p) {
			continue
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		if err != nil {
			// A tracked path deleted in the working tree has nothing to scan.
			continue
		}
		if isBinary(content) {
			continue
		}
		findings = append(findings, s.ScanText(p, content)...)
	}
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		return findings[i].Line < findings[j].Line
	})
	return findings, nil
}

// isBinary treats a NUL byte in the first 8 KiB as binary content.
func isBinary(content []byte) bool {
	head := content
	if len(head) > 8192 {
		head = head[:8192]
	}
	return bytes.IndexByte(head, 0) >= 0
}

// CountByUse groups findings by what to write instead, for a gate's summary.
func CountByUse(findings []Finding) map[string]int {
	counts := map[string]int{}
	for _, f := range findings {
		counts[f.Use]++
	}
	return counts
}
