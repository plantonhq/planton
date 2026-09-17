//go:build !codegen
// +build !codegen

package root

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/providerparity"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/spf13/cobra"
)

// ProviderParity is a developer tool (registered by the standalone binary
// beside e2e, never in the engine set): it reads committed schema artifacts
// and the catalog source tree, which exist only in a repo checkout --
// embedding it in the platform CLI would ship a command that always fails.
var ProviderParity = &cobra.Command{
	Use:   "provider-parity",
	Short: "Total-accounting check of the catalog against its pinned Terraform provider",
	Long: `Measure one cloud provider's catalog against the Terraform provider it deploys
through, at the exact pinned version, and check TOTAL accounting:

  depth   -- every configurable, non-deprecated argument of every consumed
             resource is exact-matched to a spec field, mapped by the kind's
             iac/provider-parity.yaml, or excluded there with a reason; and
             every spec field reaches provider surface (reverse drift check).
  breadth -- every GA resource carries exactly one disposition: modeled and
             iam-covered are computed; composed/model-planned/deferred/
             excluded-deprecated are recorded in the dispositions ledger.

This is PROVIDER parity -- a different axis from the cross-engine parity the
component audit's --parity focus checks (one kind's two IaC modules
implementing the same contract).

Run from the repository root. Default output is a per-kind accounting
summary; --kind details one kind (what the component audit invokes);
--output json emits the full accounting (what the public parity report
renders from); --check gates against the burn-down baseline for CI;
--write-baseline regenerates the baseline after judged work;
--write-report renders the provider's public parity page to
catalog/<provider>/terraform-parity.md (drift-gated by CI once committed).`,
	Example: `  planton provider-parity --provider gcp --ga-schema google
  planton provider-parity --provider gcp --ga-schema google --kind GcpGcsBucket
  planton provider-parity --provider gcp --ga-schema google --check
  planton provider-parity --provider gcp --ga-schema google --write-report`,
	Run: providerParityHandler,
}

func init() {
	ProviderParity.Flags().String("provider", "", "cloud provider whose catalog to account (e.g. gcp)")
	ProviderParity.Flags().String("ga-schema", "", "the parity-baseline Terraform provider schema (e.g. google)")
	ProviderParity.Flags().Bool("check", false, "exit non-zero if the parity gate fails (for CI; always gates every enrolled provider's findings)")
	ProviderParity.Flags().Bool("write-baseline", false, "regenerate the accepted-gap baseline file (always from every enrolled provider's findings)")
	ProviderParity.Flags().Bool("write-report", false, "render the public parity page from the accounting")
	ProviderParity.Flags().String("report-path", "", "public parity page path (default catalog/<provider>/terraform-parity.md)")
	ProviderParity.Flags().String("baseline", providerparity.DefaultBaselinePath, "path to the baseline file")
	ProviderParity.Flags().String("schema-dir", providerparity.DefaultSchemaDir, "directory of committed provider schema artifacts")
	ProviderParity.Flags().String("dispositions", "", "path to the dispositions ledger (default <dispositions dir>/<ga-schema>.yaml)")
	ProviderParity.Flags().String("admissions-dir", "", "directory of secondary-channel admission lists (default "+providerparity.DefaultAdmissionsDir+")")
	ProviderParity.Flags().String("kind", "", "detail one kind's accounting")
	ProviderParity.Flags().String("output", "text", "output format: text | json")
	_ = ProviderParity.MarkFlagRequired("provider")
	_ = ProviderParity.MarkFlagRequired("ga-schema")
}

func providerParityHandler(cmd *cobra.Command, _ []string) {
	providerName, _ := cmd.Flags().GetString("provider")
	gaSchema, _ := cmd.Flags().GetString("ga-schema")
	schemaDir, _ := cmd.Flags().GetString("schema-dir")
	dispositionsPath, _ := cmd.Flags().GetString("dispositions")
	admissionsDir, _ := cmd.Flags().GetString("admissions-dir")
	baselinePath, _ := cmd.Flags().GetString("baseline")

	// Accepts the catalog directory name ("gcp", "digitalocean") as well as
	// the registry enum name ("digital_ocean") -- the same resolution the
	// committed pages' embedded parameters go through.
	provider := crkreflect.ProviderFromString(providerName)
	if provider == cloudresourcekind.CloudResourceProvider_cloud_resource_provider_unspecified {
		cliprint.PrintError(fmt.Sprintf("unknown cloud provider %q", providerName))
		os.Exit(1)
	}
	schemas, err := providerparity.LoadSchemas(schemaDir)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("failed to load schema artifacts: %v (run from the repository root)", err))
		os.Exit(1)
	}

	if write, _ := cmd.Flags().GetBool("write-baseline"); write {
		// The baseline is ONE file for every enrolled provider, so a write
		// is always computed from all enrollments' merged findings -- writing
		// this run's single-provider findings would silently drop every
		// other provider's entries.
		accountings, err := providerparity.EnrolledAccountings(".", schemas)
		if err != nil {
			cliprint.PrintError(fmt.Sprintf("failed to account enrolled providers: %v", err))
			os.Exit(1)
		}
		if err := providerparity.WriteBaseline(baselinePath, providerparity.MergeFindings(accountings)); err != nil {
			cliprint.PrintError(fmt.Sprintf("failed to write baseline: %v", err))
			os.Exit(1)
		}
		cliprint.PrintSuccess(fmt.Sprintf("baseline written to %s from all enrolled providers -- review the diff before committing", baselinePath))
		return
	}

	if check, _ := cmd.Flags().GetBool("check"); check {
		// The gate reads the same one-file-for-every-provider baseline the
		// write above regenerates, so it too must consume the merged
		// findings of ALL enrollments -- gating this run's single-provider
		// findings against the shared baseline would misreport every other
		// provider's entries as stale.
		accountings, err := providerparity.EnrolledAccountings(".", schemas)
		if err != nil {
			cliprint.PrintError(fmt.Sprintf("failed to account enrolled providers: %v", err))
			os.Exit(1)
		}
		runProviderParityCheck(providerparity.MergeFindings(accountings), baselinePath)
		return
	}

	if write, _ := cmd.Flags().GetBool("write-report"); write {
		reportPath, _ := cmd.Flags().GetString("report-path")
		if reportPath == "" {
			reportPath = providerparity.PublicReportPath(provider)
		}
		page, err := providerparity.GeneratePublicReport(".",
			provider, schemas, gaSchema, dispositionsPath, admissionsDir)
		if err != nil {
			cliprint.PrintError(fmt.Sprintf("failed to render the parity page: %v", err))
			os.Exit(1)
		}
		if err := os.WriteFile(reportPath, []byte(page), 0o644); err != nil {
			cliprint.PrintError(fmt.Sprintf("failed to write the parity page: %v", err))
			os.Exit(1)
		}
		cliprint.PrintSuccess(fmt.Sprintf("parity page written to %s -- drift-gated by CI once committed", reportPath))
		return
	}

	acc, err := providerparity.BuildAccounting(".",
		provider, schemas, gaSchema, dispositionsPath, admissionsDir)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("accounting failed: %v", err))
		os.Exit(1)
	}

	if kind, _ := cmd.Flags().GetString("kind"); kind != "" {
		printKindAccounting(acc, kind)
		return
	}

	if output, _ := cmd.Flags().GetString("output"); output == "json" {
		encoded, err := json.MarshalIndent(acc, "", "  ")
		if err != nil {
			cliprint.PrintError(fmt.Sprintf("failed to marshal accounting: %v", err))
			os.Exit(1)
		}
		fmt.Println(string(encoded))
		return
	}

	printAccountingSummary(acc)
}

func runProviderParityCheck(findings []providerparity.Finding, baselinePath string) {
	baseline, err := providerparity.LoadBaseline(baselinePath)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("failed to load baseline: %v", err))
		os.Exit(1)
	}
	res := providerparity.Gate(findings, baseline)
	if res.OK() {
		cliprint.PrintSuccess("provider-parity gate passed (all enrolled providers)")
		return
	}
	for _, f := range res.NewFindings {
		cliprint.PrintError(fmt.Sprintf("new parity gap [%s]: %s", f.BaselineKey, f.Detail))
	}
	for _, key := range res.StaleEntries {
		cliprint.PrintError(fmt.Sprintf("stale baseline entry (closed -- remove it): %s", key))
	}
	os.Exit(1)
}

func printKindAccounting(acc providerparity.Accounting, kind string) {
	for _, k := range acc.Kinds {
		if k.Kind != kind {
			continue
		}
		fmt.Printf("%s against %s@%s -- manifest: %v\n", k.Kind, acc.GASchema, acc.GASchemaVersion, k.HasManifest)
		fmt.Printf("  args: %d total = %d matched + %d mapped + %d excluded + %d unaccounted\n",
			k.TotalArgs, k.MatchedArgs, k.MappedArgs, k.ExcludedArgs, len(k.UnaccountedArgs))
		for _, res := range k.InternalResources {
			fmt.Printf("  internal resource: %s\n", res)
		}
		for _, res := range k.ExternalResources {
			fmt.Printf("  external resource: %s\n", res)
		}
		for _, res := range k.AdmittedResources {
			fmt.Printf("  ADMITTED secondary-channel resource: %s\n", res)
		}
		for _, gap := range k.AdmissionGaps {
			fmt.Printf("  ADMISSION gap: %s\n", gap)
		}
		for _, arg := range k.UnaccountedArgs {
			fmt.Printf("  UNACCOUNTED arg: %s\n", arg)
		}
		for _, field := range k.UncoveredSpecFields {
			fmt.Printf("  UNCOVERED spec field: %s\n", field)
		}
		for _, stale := range k.ManifestStale {
			fmt.Printf("  STALE manifest entry: %s\n", stale)
		}
		if k.Accounted() {
			cliprint.PrintSuccess("kind is at total accounting")
		} else {
			cliprint.PrintError("kind is NOT at total accounting")
			os.Exit(1)
		}
		return
	}
	cliprint.PrintError(fmt.Sprintf("kind %q not found in the %s catalog", kind, acc.CloudProvider))
	os.Exit(1)
}

func printAccountingSummary(acc providerparity.Accounting) {
	accounted := 0
	for _, k := range acc.Kinds {
		if k.Accounted() {
			accounted++
		}
	}
	fmt.Printf("Provider parity (%s catalog vs %s@%s): %d/%d kinds at total accounting\n",
		acc.CloudProvider, acc.GASchema, acc.GASchemaVersion, accounted, len(acc.Kinds))

	if pc := acc.ProviderConfig; pc != nil {
		verdict := "AT TOTAL ACCOUNTING"
		if !pc.Accounted() {
			verdict = fmt.Sprintf("IN DEBT (unaccounted=%d uncovered=%d unjudged-module=%d stale=%d)",
				len(pc.UnaccountedArgs), len(pc.UncoveredConfigFields), len(pc.UnjudgedModuleArgs), len(pc.ManifestStale))
		}
		fmt.Printf("Provider block: %d args = %d matched + %d mapped + %d module-owned + %d excluded -- %s\n",
			pc.TotalArgs, pc.MatchedArgs, pc.MappedArgs, pc.ModuleOwnedArgs, pc.ExcludedArgs, verdict)
	}

	fmt.Println("\nBreadth dispositions:")
	classes := make([]string, 0, len(acc.DispositionTotals))
	for class := range acc.DispositionTotals {
		classes = append(classes, class)
	}
	sort.Strings(classes)
	for _, class := range classes {
		label := class
		if label == "" {
			label = "UNDISPOSITIONED"
		}
		fmt.Printf("  %-22s %d\n", label, acc.DispositionTotals[class])
	}

	fmt.Println("\nKinds in debt (run --kind <Kind> for detail):")
	for _, k := range acc.Kinds {
		if k.Accounted() {
			continue
		}
		fmt.Printf("  %-32s unaccounted=%d uncovered=%d stale=%d of %d args\n",
			k.Kind, len(k.UnaccountedArgs), len(k.UncoveredSpecFields), len(k.ManifestStale), k.TotalArgs)
	}
}
