// Package presetvalidity is the machine-enforced truth of the catalog's
// presets: every preset YAML under catalog/<provider>/<kind>/presets/ must
// load as a manifest of its own kind and pass that kind's complete
// validation rules, so a preset that dead-ends the user who copies it is
// unshippable instead of auditable.
//
// A preset is a promise -- the catalog's own authored starting point for a
// kind. Two placeholder grammars keep that promise while staying visibly
// replace-me:
//
//   - Free-string fields may carry angle-bracket tokens (`<aws-region>`):
//     nothing fences them, consoles recognize and clear them, and they read
//     as obviously-not-real.
//   - Pattern- or CEL-fenced fields need pattern-VALID placeholders (a
//     documentation-account id like 123456789012, `cdn.replaceme.example.com`,
//     `arn:aws:iam::123456789012:role/replace-me`), and proto-sensitive
//     fields carry reference-shaped placeholders (`$secret/replace-with-...`)
//     that teach the managed-secret contract instead of plaintext habits.
//
// This gate is what makes the second grammar mandatory: an angle-bracket
// token seeded into a fenced field fails the kind's own validation right
// here, at authoring time, instead of in the user's first apply.
//
// The division of labor with the sibling catalog gates is deliberate and
// closed:
//
//   - pkg/anatomy proves the FILES EXIST (a preset and its .md sidecar are
//     anatomy).
//   - pkg/catalogpage proves the catalog PAGE has the right shape and its
//     embedded manifests validate.
//   - pkg/presetvalidity (this package) proves the PRESETS THEMSELVES
//     validate.
//
// The descriptor of accepted gaps lives in baseline.yaml -- a burn-down
// list mirroring pkg/anatomy's and pkg/catalogpage's baselines (a reader
// who knows one knows all three). The CI lane is
// .github/workflows/lint.preset-validity.yaml.
package presetvalidity

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"buf.build/go/protovalidate"

	"github.com/plantonhq/planton/internal/manifest"
)

// Violation is one preset-validity rule broken at one preset.
type Violation struct {
	// Path is the repo-root-relative preset YAML path.
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
	// RuleInvalidPreset fires when a preset fails to load as a manifest of
	// its declared kind or fails that kind's validation rules -- the exact
	// rejection the user who copies the preset would hit.
	RuleInvalidPreset = "invalid-preset"

	// RulePhantomPlaceholder fires when the preset's guide (its .md
	// sidecar) lists, in its placeholder table, an angle-bracket
	// placeholder the preset YAML does not contain. It accepts no
	// baseline entries: a guide is fixed in the same change that
	// renames or drops its manifest's placeholder.
	RulePhantomPlaceholder = "phantom-placeholder"
)

// unexemptableRules are the rules the baseline may never accept: Gate fails
// their violations whether or not baseline.yaml lists them, and
// WriteBaseline never writes them.
var unexemptableRules = map[string]bool{RulePhantomPlaceholder: true}

// CheckPreset validates one preset's bytes as a manifest of its own kind.
// path is the repo-root-relative preset path used in violation IDs.
func CheckPreset(path string, content []byte) []Violation {
	v, err := protovalidate.New()
	if err != nil {
		return []Violation{{Path: path, Rule: RuleInvalidPreset,
			Detail: fmt.Sprintf("validator init: %v", err)}}
	}
	return checkPresetWith(v, path, content)
}

// checkPresetWith validates one preset with a caller-owned validator. The
// walk hands each KIND its own validator (see Check) so a kind's presets
// share one rule compilation while the compiled programs stay collectible
// once the walk moves on -- a whole-catalog validator accumulates every
// kind's compiled CEL programs and has been observed to exhaust a CI
// runner's memory, while a per-preset validator recompiles the same kind's
// rules once per preset. The rules evaluated are byte-identical to the CLI
// validation path's.
//
// A preset may deliberately hold SEVERAL documents (a composition preset
// teaching a multi-resource pattern), so validation splits the stream and
// checks every document -- the manifest loader itself accepts exactly one
// document per call, and validating only the first would ship a preset
// whose later steps were never checked.
func checkPresetWith(v protovalidate.Validator, path string, content []byte) []Violation {
	docs, err := manifest.SplitDocuments(content)
	if err != nil {
		return []Violation{{
			Path:   path,
			Rule:   RuleInvalidPreset,
			Detail: fmt.Sprintf("the preset does not validate against its own kind's rules: %s", firstMeaningfulLine(err.Error())),
		}}
	}

	var details []string
	for i, doc := range docs {
		loaded, err := manifest.LoadManifestBytes(doc, path)
		if err == nil {
			err = v.Validate(loaded)
		}
		if err != nil {
			label := ""
			if len(docs) > 1 {
				label = fmt.Sprintf("document %d: ", i+1)
			}
			details = append(details, label+firstMeaningfulLine(err.Error()))
		}
	}
	if len(details) == 0 {
		return nil
	}
	return []Violation{{
		Path:   path,
		Rule:   RuleInvalidPreset,
		Detail: fmt.Sprintf("the preset does not validate against its own kind's rules: %s", strings.Join(details, "; ")),
	}}
}

// Check walks every preset YAML at repoRoot and returns every violation,
// sorted by ID: each preset's validity against its kind, and its guide's
// placeholder table against the preset's own placeholders. It never consults the baseline -- Gate does the comparison.
// Preset PRESENCE (and the .md sidecar) is pkg/anatomy's rule, so only
// existing presets are checked; the _test provider's fixtures are not
// product presets.
//
// Presets are checked kind by kind (the sorted walk groups a kind's presets
// naturally), each kind under its own short-lived validator -- see
// checkPresetWith for why.
func Check(repoRoot string) ([]Violation, error) {
	presets, err := filepath.Glob(filepath.Join(repoRoot, "catalog", "*", "*", "presets", "*.yaml"))
	if err != nil {
		return nil, err
	}
	sort.Strings(presets)
	var vs []Violation
	var kindDir string
	var v protovalidate.Validator
	for _, preset := range presets {
		rel, err := filepath.Rel(repoRoot, preset)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(rel, filepath.Join("catalog", "_test")+string(filepath.Separator)) {
			continue
		}
		if dir := filepath.Dir(rel); dir != kindDir {
			kindDir = dir
			if v, err = protovalidate.New(); err != nil {
				return nil, err
			}
		}
		content, err := os.ReadFile(preset)
		if err != nil {
			return nil, err
		}
		vs = append(vs, checkPresetWith(v, filepath.ToSlash(rel), content)...)
		// The guide's presence is pkg/anatomy's rule; a missing sidecar
		// has no table to check here.
		if guide, err := os.ReadFile(strings.TrimSuffix(preset, ".yaml") + ".md"); err == nil {
			vs = append(vs, checkPlaceholderTable(filepath.ToSlash(rel), guide, content)...)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	sort.Slice(vs, func(i, j int) bool { return vs[i].ID() < vs[j].ID() })
	return vs, nil
}

// firstMeaningfulLine collapses the manifest validator's multi-line, colored
// error rendering into the first line that carries an actual violation, so a
// gate failure reads as one actionable sentence per preset.
func firstMeaningfulLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(stripANSI(line))
		if line == "" {
			continue
		}
		// Skip the validator's banner/framing lines; keep the first line
		// that names a field path or a plain-language violation.
		if line == "validation error:" || line == "validation errors:" ||
			strings.ContainsAny(line, "╔╚║═") ||
			strings.HasPrefix(line, "❌") || strings.HasPrefix(line, "⚠️") ||
			strings.HasPrefix(line, "💡") || strings.HasPrefix(line, "📋") ||
			strings.HasPrefix(line, "📚") || strings.HasPrefix(line, "•") ||
			strings.HasPrefix(line, "Please review") ||
			strings.HasPrefix(line, "in your manifest") {
			continue
		}
		return strings.TrimPrefix(strings.TrimPrefix(line, "validation error: "), "- ")
	}
	return strings.TrimSpace(s)
}

// stripANSI removes terminal color escape sequences from the validator's
// user-facing rendering so baseline details stay plain text.
func stripANSI(s string) string {
	var b strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inEscape {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				inEscape = false
			}
			continue
		}
		if c == 0x1b {
			inEscape = true
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}
