// Package catalogpage is the machine-enforced shape of the catalog pages:
// every kind's catalog.md follows the ONE standard structure
// (_rules/docs/write-catalog-kind-catalog-md.mdc) and embeds only
// manifests that actually validate, so a page that teaches a broken shape is
// unshippable instead of auditable.
//
// The division of labor across the three catalog packages is deliberate and
// closed:
//
//   - pkg/anatomy proves the FILE EXISTS (page presence is anatomy).
//   - pkg/catalogpage proves the PAGE HAS THE RIGHT SHAPE AND TELLS NO LIES
//     (head contract, section structure, manifest validity).
//   - pkg/catalogbundle PROJECTS the page into release artifacts and never
//     gates -- its head extraction deliberately falls back on malformed
//     input, which is exactly why the shape is gated here.
//
// The head checks guard a live machine contract: the bundle derives every
// kind's console-card title from the page's H1 and its one-line
// description from the intro's first sentence, extracted from the FIRST
// PHYSICAL LINE of the intro paragraph. A missing H1 silently ships the raw
// kind name; a hard-wrapped intro silently ships a mid-sentence fragment as
// the kind's search-result copy. Both are violations here.
//
// Structure checks fire in two tiers so the report stays signal-dense: a
// page whose H2 skeleton diverges gets ONE nonstandard-structure verdict
// (the fix is "bring the page to the standard", not section arithmetic);
// the finer checks (fixed H3 anchors, the InfraChart law) activate only
// once the skeleton conforms.
//
// The descriptor of accepted gaps lives in baseline.yaml -- a burn-down
// list mirroring pkg/anatomy's baseline (a reader who knows one knows
// both). The CI lane is .github/workflows/lint.catalog-page.yaml.
package catalogpage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/plantonhq/planton/internal/manifest"
)

// Violation is one catalog-page rule broken at one page.
type Violation struct {
	// Path is the repo-root-relative catalog.md path.
	Path string
	// Rule is the stable rule identifier (baseline entries key on it).
	Rule string
	// Detail says what the fix is, in plain language.
	Detail string
}

// ID is the stable baseline identifier: "<path>:<rule>".
func (v Violation) ID() string { return v.Path + ":" + v.Rule }

// Rule identifiers. Stable: baseline.yaml entries reference them.
const (
	RuleMissingH1            = "missing-h1"
	RuleMissingIntro         = "missing-intro"
	RuleIntroFragment        = "intro-first-line-fragment"
	RuleNonstandardStructure = "nonstandard-structure"
	RuleMissingAnchor        = "missing-h3-anchor"
	RuleInfraChartArm        = "infrachart-arm-mismatch"
	RuleInvalidManifest      = "invalid-manifest"
)

// mandatoryH2s is the exact H2 skeleton of the standard, in order. The write
// rule is the source of truth; this list mirrors it and must never drift.
var mandatoryH2s = []string{
	"What Gets Created",
	"Before You Deploy",
	"Deploy",
	"Key Configuration",
	"Outputs and Dependencies",
	"Common Patterns",
	"Works With",
}

// fixedH3Anchors are the H3 headings whose exact text is part of the
// standard (AI agents parse on them). The variable ones -- the provider
// account heading under Before You Deploy and the conditional InfraChart
// arm -- are deliberately not listed.
var fixedH3Anchors = []string{
	"Planton Setup",
	"Console",
	"CLI",
	"What This Kind Consumes",
	"What This Kind Provides",
}

var (
	yamlBlock      = regexp.MustCompile("(?s)```yaml\n(.*?)```")
	apiVersionLine = regexp.MustCompile(`(?m)^apiVersion:`)
	kindLine       = regexp.MustCompile(`(?m)^kind:`)
)

// ExtractManifests returns every COMPLETE manifest (a fenced yaml document
// declaring apiVersion + kind) embedded in a catalog page. Partial spec
// fragments (the InfraChart arm's wiring examples) are deliberately not
// manifests and are not returned.
func ExtractManifests(page []byte) []string {
	var docs []string
	for _, block := range yamlBlock.FindAllStringSubmatch(string(page), -1) {
		for _, doc := range strings.Split(block[1], "\n---\n") {
			if apiVersionLine.MatchString(doc) && kindLine.MatchString(doc) {
				docs = append(docs, doc)
			}
		}
	}
	return docs
}

// heading is one markdown heading outside fenced code (fenced "# comment"
// lines are code, not structure).
type heading struct {
	level int
	text  string
}

// parseHeadings extracts the page's heading outline, fence-aware.
func parseHeadings(page []byte) []heading {
	var hs []heading
	inFence := false
	for _, line := range strings.Split(string(page), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if inFence || !strings.HasPrefix(line, "#") {
			continue
		}
		level := 0
		for level < len(line) && line[level] == '#' {
			level++
		}
		if level >= len(line) || line[level] != ' ' {
			continue
		}
		hs = append(hs, heading{level: level, text: strings.TrimSpace(line[level+1:])})
	}
	return hs
}

// introFirstLine returns the first physical prose line between the H1 and
// the first H2 (fence-aware), or "" when the intro paragraph is absent.
// That exact line is what the bundle's description extraction reads.
func introFirstLine(page []byte) string {
	inFence := false
	seenH1 := false
	for _, line := range strings.Split(string(page), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if strings.HasPrefix(line, "# ") {
			seenH1 = true
			continue
		}
		if strings.HasPrefix(line, "##") {
			return "" // reached structure before any prose
		}
		if seenH1 && trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// completesSentence reports whether the bundle's first-sentence extraction
// gets a complete sentence out of this line: either a sentence boundary to
// cut at, or the line itself ends one.
func completesSentence(line string) bool {
	return strings.Contains(line, ". ") || strings.HasSuffix(line, ".")
}

// consumesHasRows reports whether the Consumes section carries a non-empty
// dependency table (data rows beyond the header and separator). A prose
// fallback ("no foreign key dependencies") counts as empty.
func consumesHasRows(page []byte) bool {
	inFence := false
	inConsumes := false
	rows := 0
	for _, line := range strings.Split(string(page), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if strings.HasPrefix(line, "#") {
			inConsumes = strings.HasPrefix(line, "### ") &&
				strings.TrimSpace(strings.TrimPrefix(line, "### ")) == "What This Kind Consumes"
			continue
		}
		if inConsumes && strings.HasPrefix(trimmed, "|") && !strings.HasPrefix(trimmed, "|-") &&
			!strings.Contains(trimmed, "---") && !strings.Contains(trimmed, "Dependency") {
			rows++
		}
	}
	return rows > 0
}

// CheckPage runs every catalog-page rule against one page's content. The
// path is used only for violation reporting.
func CheckPage(path string, page []byte) []Violation {
	var vs []Violation
	add := func(rule, detail string) {
		vs = append(vs, Violation{Path: path, Rule: rule, Detail: detail})
	}

	hs := parseHeadings(page)

	// Head contract -- always checked: the bundle projects it silently.
	if len(hs) == 0 || hs[0].level != 1 {
		add(RuleMissingH1, "the H1 is the kind's display name (console card title) and must be the first heading")
	} else {
		switch line := introFirstLine(page); {
		case line == "":
			add(RuleMissingIntro, "an intro paragraph between the H1 and the first H2 is the kind's description source")
		case !completesSentence(line):
			add(RuleIntroFragment, fmt.Sprintf(
				"the intro's first line must complete a sentence (the bundle extracts the description from it); got a fragment: %q", line))
		}
	}

	// Manifest truth -- always checked: a page whose copy-pasted manifest
	// fails `planton apply` teaches every reader a broken shape.
	for i, doc := range ExtractManifests(page) {
		loaded, err := manifest.LoadManifestBytes([]byte(doc), path)
		if err == nil {
			err = manifest.ValidateLoaded(loaded)
		}
		if err != nil {
			add(RuleInvalidManifest, fmt.Sprintf("embedded manifest #%d does not validate: %s", i+1, firstLine(err.Error())))
		}
	}

	// Structure -- one verdict for a diverging skeleton; the finer checks
	// activate only once the skeleton conforms (signal over stacking).
	var h2s []string
	h3s := map[string]bool{}
	for _, h := range hs {
		switch h.level {
		case 2:
			h2s = append(h2s, h.text)
		case 3:
			h3s[h.text] = true
		}
	}
	if diff := structureDiff(h2s); diff != "" {
		add(RuleNonstandardStructure, "the H2 skeleton diverges from the standard ("+diff+") -- bring the page to _rules/docs/write-catalog-kind-catalog-md.mdc")
		return vs
	}
	for _, anchor := range fixedH3Anchors {
		if !h3s[anchor] {
			add(RuleMissingAnchor, fmt.Sprintf("the standard's fixed H3 anchor %q is absent", anchor))
		}
	}
	hasRows := consumesHasRows(page)
	if hasRows && !h3s["InfraChart"] {
		add(RuleInfraChartArm, "the Consumes table has rows, so the Deploy section requires an InfraChart arm")
	}
	if !hasRows && h3s["InfraChart"] {
		add(RuleInfraChartArm, "the Consumes table is empty, so the InfraChart arm must be omitted")
	}
	return vs
}

// structureDiff compares a page's H2 sequence against the standard and
// describes the divergence, or returns "" when it conforms exactly.
func structureDiff(h2s []string) string {
	if len(h2s) == len(mandatoryH2s) {
		match := true
		for i := range h2s {
			if h2s[i] != mandatoryH2s[i] {
				match = false
				break
			}
		}
		if match {
			return ""
		}
	}
	want := map[string]bool{}
	for _, s := range mandatoryH2s {
		want[s] = true
	}
	got := map[string]bool{}
	for _, s := range h2s {
		got[s] = true
	}
	var missing, foreign []string
	for _, s := range mandatoryH2s {
		if !got[s] {
			missing = append(missing, s)
		}
	}
	for _, s := range h2s {
		if !want[s] {
			foreign = append(foreign, s)
		}
	}
	var parts []string
	if len(missing) > 0 {
		parts = append(parts, "missing: "+strings.Join(missing, ", "))
	}
	if len(foreign) > 0 {
		parts = append(parts, "foreign: "+strings.Join(foreign, ", "))
	}
	if len(parts) == 0 {
		parts = append(parts, "sections out of order")
	}
	return strings.Join(parts, "; ")
}

// Check walks every kind catalog page at repoRoot and returns every
// violation, sorted by ID. It never consults the baseline -- Gate does the
// comparison. Page PRESENCE is pkg/anatomy's rule, so only existing pages
// are checked; the _test provider's fixtures are not product pages.
func Check(repoRoot string) ([]Violation, error) {
	pages, err := filepath.Glob(filepath.Join(repoRoot, "catalog", "*", "*", "catalog.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(pages)
	var vs []Violation
	for _, page := range pages {
		rel, err := filepath.Rel(repoRoot, page)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(rel, filepath.Join("catalog", "_test")+string(filepath.Separator)) {
			continue
		}
		content, err := os.ReadFile(page)
		if err != nil {
			return nil, err
		}
		vs = append(vs, CheckPage(filepath.ToSlash(rel), content)...)
	}
	sort.Slice(vs, func(i, j int) bool { return vs[i].ID() < vs[j].ID() })
	return vs, nil
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
