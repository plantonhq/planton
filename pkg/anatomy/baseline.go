// Baseline of accepted anatomy gaps and the gate that compares live findings
// against it. The baseline is a burn-down list (kinds that predate a
// requirement and have not been brought up to shape yet -- gap-fill belongs
// to the provider parity programs); it is never a permanent exemption. The
// shape mirrors pkg/secretcoverage's baseline so a reader who knows one
// knows both.
package anatomy

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const baselineHeader = `# Kind-anatomy baseline -- the accepted backlog of anatomy gaps: kinds
# that predate a requirement (spec tests, module READMEs, presets, catalog pages)
# and have not been brought up to the canonical shape yet. Gap-fill is routed to
# the provider parity programs; this list trends to 0 and never grows.
#
# The CI guardrail (go test ./pkg/anatomy/...) fails when:
#   - a violation appears that is NOT listed here (new anatomy drift shipped), or
#   - a listed entry is no longer a violation (it was fixed -- remove it here).
#
# Entry format: <repo-relative path>:<rule>   (see pkg/anatomy rule constants)
# Regenerate with:  PLANTON_REGEN_ANATOMY_BASELINE=1 go test ./pkg/anatomy/
`

type baselineDoc struct {
	Violations []string `yaml:"violations"`
}

// LoadBaseline reads the accepted-gap set from a baseline file.
func LoadBaseline(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc baselineDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse baseline %s: %w", path, err)
	}
	set := make(map[string]bool, len(doc.Violations))
	for _, v := range doc.Violations {
		set[v] = true
	}
	return set, nil
}

// WriteBaseline writes the current violations as the accepted baseline.
func WriteBaseline(path string, violations []Violation) error {
	ids := make([]string, 0, len(violations))
	for _, v := range violations {
		ids = append(ids, v.ID())
	}
	sort.Strings(ids)
	var b strings.Builder
	b.WriteString(baselineHeader)
	if len(ids) == 0 {
		b.WriteString("violations: []\n")
	} else {
		b.WriteString("violations:\n")
		for _, id := range ids {
			b.WriteString("  - " + id + "\n")
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// GateResult is the verdict of comparing live violations to the baseline.
type GateResult struct {
	NewViolations []Violation // drift not in the baseline -- new anatomy drift shipped
	StaleEntries  []string    // baseline ids no longer violated -- remove them
}

// OK reports whether the gate passes.
func (g GateResult) OK() bool {
	return len(g.NewViolations) == 0 && len(g.StaleEntries) == 0
}

// Gate compares violations against the accepted baseline. It is the single
// source of truth for the CI test (and any future doctor command), so the
// two can never disagree.
func Gate(violations []Violation, baseline map[string]bool) GateResult {
	var res GateResult
	current := map[string]bool{}
	for _, v := range violations {
		id := v.ID()
		current[id] = true
		if !baseline[id] {
			res.NewViolations = append(res.NewViolations, v)
		}
	}
	for id := range baseline {
		if !current[id] {
			res.StaleEntries = append(res.StaleEntries, id)
		}
	}
	sort.Slice(res.NewViolations, func(i, j int) bool {
		return res.NewViolations[i].ID() < res.NewViolations[j].ID()
	})
	sort.Strings(res.StaleEntries)
	return res
}
