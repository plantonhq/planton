//go:build !codegen
// +build !codegen

package root

import (
	"fmt"
	"os"
	"sort"

	"github.com/plantonhq/openmcf/internal/cli/cliprint"
	"github.com/plantonhq/openmcf/pkg/secretcoverage"
	"github.com/spf13/cobra"
)

var SecretCoverage = &cobra.Command{
	Use:   "secret-coverage",
	Short: "Report secure-by-default coverage of `sensitive` fields across all resource kinds",
	Long: `Walk every production cloud-resource kind and classify each string-bearing field as:
  covered  -- annotated (org.openmcf.shared.options.sensitive) = true
  exempt   -- annotated (org.openmcf.shared.options.sensitive_exempt_reason) = "..."
  gap      -- looks like a secret by name but is not annotated

Default output is a coverage report. Use --check in CI to fail on any new gap (one
not in the baseline), any stale baseline entry, or a self-contradictory annotation.
Use --write-baseline to record the current accepted gaps after an annotation pass.`,
	Run: secretCoverageHandler,
}

func init() {
	SecretCoverage.Flags().Bool("check", false, "exit non-zero if the coverage gate fails (for CI)")
	SecretCoverage.Flags().Bool("write-baseline", false, "regenerate the accepted-gap baseline file")
	SecretCoverage.Flags().String("baseline", secretcoverage.DefaultBaselinePath, "path to the baseline file")
}

func secretCoverageHandler(cmd *cobra.Command, _ []string) {
	findings := secretcoverage.Analyze()
	baselinePath, _ := cmd.Flags().GetString("baseline")

	if write, _ := cmd.Flags().GetBool("write-baseline"); write {
		if err := secretcoverage.WriteBaseline(baselinePath, findings); err != nil {
			cliprint.PrintError(fmt.Sprintf("failed to write baseline: %v", err))
			os.Exit(1)
		}
		cliprint.PrintSuccess(fmt.Sprintf("wrote %d accepted gaps to %s", len(secretcoverage.GapIDs(findings)), baselinePath))
		return
	}

	if check, _ := cmd.Flags().GetBool("check"); check {
		runCheck(findings, baselinePath)
		return
	}

	printReport(findings)
}

func runCheck(findings []secretcoverage.Finding, baselinePath string) {
	baseline, err := secretcoverage.LoadBaseline(baselinePath)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("failed to load baseline: %v", err))
		os.Exit(1)
	}
	res := secretcoverage.Gate(findings, baseline)
	if res.OK() {
		cliprint.PrintSuccess("secret-coverage gate passed")
		return
	}
	for _, id := range res.NewGaps {
		cliprint.PrintError(fmt.Sprintf("new unannotated sensitive-looking field: %s -- annotate it `sensitive`, or exempt it with `sensitive_exempt_reason`", id))
	}
	for _, id := range res.StaleEntries {
		cliprint.PrintError(fmt.Sprintf("stale baseline entry (no longer a gap): %s -- remove it from %s", id, baselinePath))
	}
	for _, f := range res.AnnotationViolations {
		for _, v := range f.Violations {
			cliprint.PrintError(fmt.Sprintf("%s:%s -- %s", f.Kind, f.Path, v))
		}
	}
	os.Exit(1)
}

func printReport(findings []secretcoverage.Finding) {
	summary := secretcoverage.Summarize(findings)
	fmt.Printf("Secure-by-default coverage: %.1f%%  (covered=%d exempt=%d gap=%d)\n",
		summary.CoveragePercent(), summary.Covered, summary.Exempt, summary.Gap)
	if summary.Violations > 0 {
		fmt.Printf("annotation violations: %d (run with --check for detail)\n", summary.Violations)
	}

	byProvider := map[string]*secretcoverage.Summary{}
	for _, f := range findings {
		s := byProvider[f.Provider]
		if s == nil {
			s = &secretcoverage.Summary{}
			byProvider[f.Provider] = s
		}
		switch f.Class {
		case secretcoverage.Covered:
			s.Covered++
		case secretcoverage.Exempt:
			s.Exempt++
		case secretcoverage.Gap:
			s.Gap++
		}
	}
	providers := make([]string, 0, len(byProvider))
	for p := range byProvider {
		providers = append(providers, p)
	}
	sort.Strings(providers)

	fmt.Println("\nBy provider:")
	for _, p := range providers {
		s := byProvider[p]
		fmt.Printf("  %-16s %.0f%%  covered=%d exempt=%d gap=%d\n", p, s.CoveragePercent(), s.Covered, s.Exempt, s.Gap)
	}

	gaps := secretcoverage.GapIDs(findings)
	if len(gaps) > 0 {
		fmt.Printf("\nGaps (%d) -- unannotated fields that look sensitive:\n", len(gaps))
		for _, id := range gaps {
			fmt.Printf("  %s\n", id)
		}
	}
}
