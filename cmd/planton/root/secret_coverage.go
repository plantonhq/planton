//go:build !codegen
// +build !codegen

package root

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/pkg/secretcoverage"
	"github.com/spf13/cobra"
)

var SecretCoverage = &cobra.Command{
	Use:   "secret-coverage",
	Short: "Report secure-by-default coverage of `sensitive` fields across all resource kinds",
	Long: `Walk every production cloud-resource kind and classify each string-bearing field as:
  covered  -- annotated (dev.planton.shared.options.sensitive) = true
  exempt   -- annotated (dev.planton.shared.options.sensitive_exempt_reason) = "..."
  gap      -- looks like a secret by name but is not annotated

Default output is a human-readable coverage report. Use --output json to emit a
machine-readable report (overall + per-provider summary + gap list) for downstream
surfaces such as the Planton OS secret-coverage audit tile. Use --check in CI to
fail on any new gap (one not in the baseline), any stale baseline entry, or a
self-contradictory annotation. Use --write-baseline to record the current accepted
gaps after an annotation pass.`,
	Run: secretCoverageHandler,
}

func init() {
	SecretCoverage.Flags().Bool("check", false, "exit non-zero if the coverage gate fails (for CI)")
	SecretCoverage.Flags().Bool("write-baseline", false, "regenerate the accepted-gap baseline file")
	SecretCoverage.Flags().String("baseline", secretcoverage.DefaultBaselinePath, "path to the baseline file")
	SecretCoverage.Flags().String("output", "text", "output format: text | json")
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

	if output, _ := cmd.Flags().GetString("output"); output == "json" {
		printJSONReport(findings)
		return
	}

	printReport(findings)
}

func printJSONReport(findings []secretcoverage.Finding) {
	report := secretcoverage.BuildReport(findings)
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("failed to marshal coverage report: %v", err))
		os.Exit(1)
	}
	fmt.Println(string(encoded))
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
	report := secretcoverage.BuildReport(findings)
	fmt.Printf("Secure-by-default coverage: %.1f%%  (covered=%d exempt=%d gap=%d)\n",
		report.CoveragePercent, report.Covered, report.Exempt, report.Gap)
	if report.Violations > 0 {
		fmt.Printf("annotation violations: %d (run with --check for detail)\n", report.Violations)
	}

	fmt.Println("\nBy provider:")
	for _, p := range report.Providers {
		fmt.Printf("  %-16s %.0f%%  covered=%d exempt=%d gap=%d\n", p.Provider, p.CoveragePercent, p.Covered, p.Exempt, p.Gap)
	}

	if len(report.Gaps) > 0 {
		fmt.Printf("\nGaps (%d) -- unannotated fields that look sensitive:\n", len(report.Gaps))
		for _, id := range report.Gaps {
			fmt.Printf("  %s\n", id)
		}
	}
}
