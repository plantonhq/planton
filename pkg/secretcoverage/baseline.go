// Baseline of accepted secret-coverage gaps and the gate that compares live findings
// against it. The baseline is the annotation-sweep backlog (fields that ARE secrets
// but are not annotated yet); it is deliberately distinct from the permanent proto
// `sensitive_exempt_reason` escape hatch (fields that are intentionally NOT secrets).

//go:build !codegen
// +build !codegen

package secretcoverage

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultBaselinePath is repo-root-relative (where the CLI runs). The test reads the
// file by its bare name because `go test` runs with the package directory as cwd.
const DefaultBaselinePath = "pkg/secretcoverage/baseline.yaml"

const baselineHeader = `# Secret-coverage baseline -- the accepted backlog of cloud-resource fields that
# LOOK sensitive by name (the secret heuristic) but are not yet annotated with the
# Planton ` + "`sensitive`" + ` option. This is the annotation-sweep TODO list.
#
# It is NOT a permanent exemption. A field that is intentionally not a secret (a
# public key, an access-key id, a resource name) must use the proto
# ` + "`sensitive_exempt_reason`" + ` option instead, which documents WHY in the proto itself.
#
# The CI guardrail (go test ./pkg/secretcoverage/...) fails when:
#   - a gap appears that is NOT listed here (a new unannotated secret field shipped), or
#   - a listed entry is no longer a gap (it was annotated/exempted -- remove it here).
# As the sweep proceeds, annotate fields and delete their lines; this list trends to 0.
#
# Regenerate with:  planton secret-coverage --write-baseline
`

type baselineDoc struct {
	Gaps []string `yaml:"gaps"`
}

// GapID is the stable identifier for a gap: "<Kind>:<fieldPath>".
func GapID(f Finding) string {
	return f.Kind + ":" + f.Path
}

// GapIDs returns the sorted gap identifiers among findings.
func GapIDs(findings []Finding) []string {
	var ids []string
	for _, f := range findings {
		if f.Class == Gap {
			ids = append(ids, GapID(f))
		}
	}
	sort.Strings(ids)
	return ids
}

func LoadBaseline(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc baselineDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse baseline %s: %w", path, err)
	}
	set := make(map[string]bool, len(doc.Gaps))
	for _, g := range doc.Gaps {
		set[g] = true
	}
	return set, nil
}

func WriteBaseline(path string, findings []Finding) error {
	ids := GapIDs(findings)
	var b strings.Builder
	b.WriteString(baselineHeader)
	if len(ids) == 0 {
		b.WriteString("gaps: []\n")
	} else {
		b.WriteString("gaps:\n")
		for _, id := range ids {
			b.WriteString("  - " + id + "\n")
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// GateResult is the verdict of comparing live findings to the checked-in baseline.
type GateResult struct {
	NewGaps              []string  // gaps not in the baseline -- new unannotated secret fields
	StaleEntries         []string  // baseline ids that are no longer gaps -- remove them
	AnnotationViolations []Finding // findings whose annotation is self-contradictory or pointless
}

func (g GateResult) OK() bool {
	return len(g.NewGaps) == 0 && len(g.StaleEntries) == 0 && len(g.AnnotationViolations) == 0
}

// Gate compares findings against the accepted baseline and reports every reason the
// guardrail should fail. It is the single source of truth for both the CLI `--check`
// and the CI test, so they can never disagree.
func Gate(findings []Finding, baseline map[string]bool) GateResult {
	var res GateResult
	currentGaps := map[string]bool{}
	for _, f := range findings {
		if len(f.Violations) > 0 {
			res.AnnotationViolations = append(res.AnnotationViolations, f)
		}
		if f.Class == Gap {
			id := GapID(f)
			currentGaps[id] = true
			if !baseline[id] {
				res.NewGaps = append(res.NewGaps, id)
			}
		}
	}
	for id := range baseline {
		if !currentGaps[id] {
			res.StaleEntries = append(res.StaleEntries, id)
		}
	}
	sort.Strings(res.NewGaps)
	sort.Strings(res.StaleEntries)
	return res
}
