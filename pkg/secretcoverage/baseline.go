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

const baselineHeader = `# Secret-coverage baseline -- the accepted backlog of infra-component fields that
# LOOK sensitive by name (the secret heuristic) but are not yet annotated with the
# Planton ` + "`sensitive`" + ` option. This is the annotation-sweep TODO list.
#
# It is NOT a permanent exemption. A field that is intentionally not a secret (a
# public key, an access-key id, a resource name) must use the proto
# ` + "`sensitive_exempt_reason`" + ` option instead, which documents WHY in the proto itself.
#
# The ` + "`violations`" + ` list is the accepted backlog of annotation violations -- today
# that means sensitive fields still carrying a value-content validation rule (a
# pattern/CEL written for the raw value, which a stored managed-secret reference can
# never satisfy, making the field impossible to fill through annotation-driven
# surfaces). The fix is always the same: delete the rule, teach the shape in the
# field's comment, and remove the entry here. Each entry names its owning program.
#
# The CI guardrail (go test ./pkg/secretcoverage/...) fails when:
#   - a gap appears that is NOT listed here (a new unannotated secret field shipped),
#   - a violation appears that is NOT listed here (a new contradictory annotation),
#   - a listed entry is no longer a gap/violation (fixed -- remove it here).
# Both lists only burn down; never add an entry for new work.
#
# Regenerate with:  planton secret-coverage --write-baseline
`

type baselineDoc struct {
	Gaps       []string `yaml:"gaps"`
	Violations []string `yaml:"violations"`
}

// Baseline is the parsed accepted backlog: gap ids and violation ids (both
// "<Kind>:<fieldPath>"). A violation entry accepts ALL violations on that field --
// one field, one owner, one fix.
type Baseline struct {
	Gaps       map[string]bool
	Violations map[string]bool
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

// ViolationIDs returns the sorted ids of findings carrying annotation violations.
func ViolationIDs(findings []Finding) []string {
	var ids []string
	for _, f := range findings {
		if len(f.Violations) > 0 {
			ids = append(ids, GapID(f))
		}
	}
	sort.Strings(ids)
	return ids
}

func LoadBaseline(path string) (Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Baseline{}, err
	}
	var doc baselineDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return Baseline{}, fmt.Errorf("parse baseline %s: %w", path, err)
	}
	b := Baseline{Gaps: map[string]bool{}, Violations: map[string]bool{}}
	for _, g := range doc.Gaps {
		b.Gaps[g] = true
	}
	for _, v := range doc.Violations {
		b.Violations[v] = true
	}
	return b, nil
}

func WriteBaseline(path string, findings []Finding) error {
	var b strings.Builder
	b.WriteString(baselineHeader)
	writeIDList(&b, "gaps", GapIDs(findings))
	writeIDList(&b, "violations", ViolationIDs(findings))
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func writeIDList(b *strings.Builder, key string, ids []string) {
	if len(ids) == 0 {
		b.WriteString(key + ": []\n")
		return
	}
	b.WriteString(key + ":\n")
	for _, id := range ids {
		b.WriteString("  - " + id + "\n")
	}
}

// GateResult is the verdict of comparing live findings to the checked-in baseline.
type GateResult struct {
	NewGaps               []string  // gaps not in the baseline -- new unannotated secret fields
	StaleEntries          []string  // baseline gap ids that are no longer gaps -- remove them
	AnnotationViolations  []Finding // non-baselined findings whose annotations contradict each other or the field's rules
	StaleViolationEntries []string  // baseline violation ids no longer violating -- remove them
}

func (g GateResult) OK() bool {
	return len(g.NewGaps) == 0 && len(g.StaleEntries) == 0 &&
		len(g.AnnotationViolations) == 0 && len(g.StaleViolationEntries) == 0
}

// Gate compares findings against the accepted baseline and reports every reason the
// guardrail should fail. It is the single source of truth for both the CLI `--check`
// and the CI test, so they can never disagree.
func Gate(findings []Finding, baseline Baseline) GateResult {
	var res GateResult
	currentGaps := map[string]bool{}
	currentViolations := map[string]bool{}
	for _, f := range findings {
		if len(f.Violations) > 0 {
			id := GapID(f)
			currentViolations[id] = true
			if !baseline.Violations[id] {
				res.AnnotationViolations = append(res.AnnotationViolations, f)
			}
		}
		if f.Class == Gap {
			id := GapID(f)
			currentGaps[id] = true
			if !baseline.Gaps[id] {
				res.NewGaps = append(res.NewGaps, id)
			}
		}
	}
	for id := range baseline.Gaps {
		if !currentGaps[id] {
			res.StaleEntries = append(res.StaleEntries, id)
		}
	}
	for id := range baseline.Violations {
		if !currentViolations[id] {
			res.StaleViolationEntries = append(res.StaleViolationEntries, id)
		}
	}
	sort.Strings(res.NewGaps)
	sort.Strings(res.StaleEntries)
	sort.Strings(res.StaleViolationEntries)
	return res
}
