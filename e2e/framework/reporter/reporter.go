// Package reporter generates JSON and Markdown reports from E2E test results.
package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/e2e/framework/runner"
)

// Report is the top-level E2E test report structure.
type Report struct {
	Timestamp  time.Time             `json:"timestamp"`
	Duration   time.Duration         `json:"duration"`
	TotalTests int                   `json:"total_tests"`
	Passed     int                   `json:"passed"`
	Failed     int                   `json:"failed"`
	Results    []ComponentReport     `json:"results"`
}

// ComponentReport holds the outcome of one component's E2E test.
type ComponentReport struct {
	Component string        `json:"component"`
	Engine    string        `json:"engine"`
	Passed    bool          `json:"passed"`
	Duration  time.Duration `json:"duration"`
	Phases    []PhaseReport `json:"phases"`
	Error     string        `json:"error,omitempty"`
}

// PhaseReport holds the outcome of a single test phase.
type PhaseReport struct {
	Phase    string        `json:"phase"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
}

// NewReport creates a report from a list of test results.
func NewReport(results []*runner.TestResult) *Report {
	report := &Report{
		Timestamp:  time.Now(),
		TotalTests: len(results),
	}

	var totalDuration time.Duration
	for _, r := range results {
		cr := ComponentReport{
			Component: r.Component,
			Engine:    r.Engine,
			Passed:    r.Passed,
			Duration:  r.Duration,
		}

		for _, p := range r.Phases {
			pr := PhaseReport{
				Phase:    string(p.Phase),
				Passed:   p.Passed,
				Duration: p.Duration,
			}
			if p.Error != nil {
				pr.Error = p.Error.Error()
				cr.Error = p.Error.Error()
			}
			cr.Phases = append(cr.Phases, pr)
		}

		if r.Passed {
			report.Passed++
		} else {
			report.Failed++
		}

		report.Results = append(report.Results, cr)
		totalDuration += r.Duration
	}

	report.Duration = totalDuration
	return report
}

// WriteJSON writes the report as JSON to the given path.
func (r *Report) WriteJSON(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal report to JSON")
	}
	return os.WriteFile(path, data, 0644)
}

// WriteMarkdown writes the report as a Markdown summary to the given path.
func (r *Report) WriteMarkdown(path string) error {
	var sb strings.Builder

	sb.WriteString("# OpenMCF E2E Test Report\n\n")
	sb.WriteString(fmt.Sprintf("**Date:** %s\n\n", r.Timestamp.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("**Total:** %d | **Passed:** %d | **Failed:** %d\n\n", r.TotalTests, r.Passed, r.Failed))

	if r.Failed > 0 {
		sb.WriteString("## Failed Tests\n\n")
		for _, cr := range r.Results {
			if !cr.Passed {
				sb.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", cr.Component, cr.Engine, cr.Error))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Results\n\n")
	sb.WriteString("| Component | Engine | Status | Duration |\n")
	sb.WriteString("|-----------|--------|--------|----------|\n")
	for _, cr := range r.Results {
		status := "PASS"
		if !cr.Passed {
			status = "FAIL"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", cr.Component, cr.Engine, status, cr.Duration.Round(time.Millisecond)))
	}
	sb.WriteString("\n")

	sb.WriteString("## Phase Details\n\n")
	for _, cr := range r.Results {
		sb.WriteString(fmt.Sprintf("### %s (%s)\n\n", cr.Component, cr.Engine))
		for _, p := range cr.Phases {
			icon := "OK"
			if !p.Passed {
				icon = "FAIL"
			}
			sb.WriteString(fmt.Sprintf("- %s: %s (%s)", p.Phase, icon, p.Duration.Round(time.Millisecond)))
			if p.Error != "" {
				sb.WriteString(fmt.Sprintf(" -- %s", p.Error))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}
