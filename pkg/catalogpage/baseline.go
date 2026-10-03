// Baseline of accepted catalog-page gaps and the gate that compares live
// findings against it. The baseline is a burn-down list (pages that predate
// the ONE standard and have not been brought to the bar yet -- the upgrade
// sweep runs provider by provider); it is never a permanent exemption. The
// shape mirrors pkg/anatomy's baseline so a reader who knows one knows both.
package catalogpage

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const baselineHeader = `# Catalog-page baseline -- the accepted backlog of pages below the ONE catalog
# page standard (_rules/docs/write-catalog-kind-catalog-md.mdc): pages that
# predate the standard and have not been brought to the bar yet. The upgrade
# sweep burns this list down provider by provider; it trends to 0 and never
# grows.
#
# The CI guardrail (go test ./pkg/catalogpage/...) fails when:
#   - a violation appears that is NOT listed here (a page shipped below the bar), or
#   - a listed entry is no longer a violation (it was fixed -- remove it here).
#
# Entry format: <repo-relative catalog.md path>:<rule>   (see pkg/catalogpage rule constants)
# Regenerate with:  PLANTON_REGEN_CATALOG_PAGE_BASELINE=1 go test ./pkg/catalogpage/
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
	seen := map[string]bool{}
	ids := make([]string, 0, len(violations))
	for _, v := range violations {
		if id := v.ID(); !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
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
	NewViolations []Violation // drift not in the baseline -- a page shipped below the bar
	StaleEntries  []string    // baseline ids no longer violated -- remove them
}

// OK reports whether the gate passes.
func (g GateResult) OK() bool {
	return len(g.NewViolations) == 0 && len(g.StaleEntries) == 0
}

// Gate compares violations against the accepted baseline. It is the single
// source of truth for the CI test, so the two can never disagree.
func Gate(violations []Violation, baseline map[string]bool) GateResult {
	var res GateResult
	current := map[string]bool{}
	for _, v := range violations {
		id := v.ID()
		if current[id] {
			continue // one rule can fire multiple times per page; the baseline entry is one
		}
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
