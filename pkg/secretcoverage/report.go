package secretcoverage

import "sort"

// Report is the machine-readable secure-by-default coverage report. It carries
// the overall summary, the per-provider breakdown, and the accepted/observed gap
// list -- everything the human stdout report shows -- so downstream surfaces
// (e.g. the Planton OS secret-coverage audit tile) can render the numbers
// without re-deriving the heuristic. Built from the same Analyze() findings the
// CLI report and the CI gate use, so there is a single source of truth.
type Report struct {
	// CoveragePercent is (covered+exempt) / (covered+exempt+gap) * 100.
	CoveragePercent float64 `json:"coveragePercent"`
	Covered         int     `json:"covered"`
	Exempt          int     `json:"exempt"`
	Gap             int     `json:"gap"`
	// Violations is the count of self-contradictory annotations (sensitive AND
	// exempt, or similar) that the gate reports.
	Violations int `json:"violations"`
	// Providers is the per-provider breakdown, sorted by provider name.
	Providers []ProviderCoverage `json:"providers"`
	// Gaps is the sorted list of "<Kind>:<spec.field.path>" gap IDs -- fields
	// that look sensitive by name but are not yet annotated or exempted.
	Gaps []string `json:"gaps"`
}

// ProviderCoverage is one provider's slice of the coverage report.
type ProviderCoverage struct {
	Provider        string  `json:"provider"`
	CoveragePercent float64 `json:"coveragePercent"`
	Covered         int     `json:"covered"`
	Exempt          int     `json:"exempt"`
	Gap             int     `json:"gap"`
}

// BuildReport aggregates raw findings into the machine-readable Report. It is the
// single aggregation used by both the JSON emit and the human stdout report.
func BuildReport(findings []Finding) Report {
	overall := Summarize(findings)

	byProvider := map[string]*Summary{}
	for _, f := range findings {
		s := byProvider[f.Provider]
		if s == nil {
			s = &Summary{}
			byProvider[f.Provider] = s
		}
		switch f.Class {
		case Covered:
			s.Covered++
		case Exempt:
			s.Exempt++
		case Gap:
			s.Gap++
		}
	}

	providers := make([]ProviderCoverage, 0, len(byProvider))
	for name, s := range byProvider {
		providers = append(providers, ProviderCoverage{
			Provider:        name,
			CoveragePercent: s.CoveragePercent(),
			Covered:         s.Covered,
			Exempt:          s.Exempt,
			Gap:             s.Gap,
		})
	}
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Provider < providers[j].Provider
	})

	return Report{
		CoveragePercent: overall.CoveragePercent(),
		Covered:         overall.Covered,
		Exempt:          overall.Exempt,
		Gap:             overall.Gap,
		Violations:      overall.Violations,
		Providers:       providers,
		Gaps:            GapIDs(findings),
	}
}
