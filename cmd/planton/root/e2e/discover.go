package e2e

import (
	"fmt"
	"os"

	"github.com/plantonhq/planton/internal/cli/ui/e2ediscover"
	"github.com/plantonhq/planton/pkg/e2e/profile"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	componentv1 "github.com/plantonhq/planton/apis/dev/planton/qa/componente2eprofile/v1"
	sharedpb "github.com/plantonhq/planton/apis/dev/planton/shared"
)

var Discover = &cobra.Command{
	Use:   "discover",
	Short: "Discover E2E-testable components and generate CI matrix",
	Long: `Scan a provider's E2E profiles and display component readiness.

Three output modes:
  interactive  Interactive TUI with keyboard navigation (default in terminal)
  table        Plain text table (default when piped)
  github-matrix  JSON for GitHub Actions matrix strategy`,
	Example: `  planton e2e discover --provider kubernetes
  planton e2e discover --provider kubernetes --output github-matrix
  planton e2e discover --provider kubernetes --status green --tier 1`,
	RunE: runDiscover,
}

func init() {
	Discover.Flags().String("provider", "", "cloud provider to discover (required)")
	Discover.MarkFlagRequired("provider")
	Discover.Flags().String("output", "", "output format: interactive, table, github-matrix (auto-detected if omitted)")
	Discover.Flags().String("status", "", "filter by status: green, deferred, skip, stub")
	Discover.Flags().Int32("tier", 0, "filter by tier (1-4)")
	Discover.Flags().String("provisioner", "", "filter by validated provisioner: pulumi, terraform")
}

func runDiscover(cmd *cobra.Command, args []string) error {
	providerName, _ := cmd.Flags().GetString("provider")
	outputMode, _ := cmd.Flags().GetString("output")
	statusFilter, _ := cmd.Flags().GetString("status")
	tierFilter, _ := cmd.Flags().GetInt32("tier")
	provisionerFilter, _ := cmd.Flags().GetString("provisioner")

	repoRoot, err := detectRepoRoot()
	if err != nil {
		return err
	}

	opts := profile.FilterOpts{
		Tier: tierFilter,
	}

	if statusFilter != "" {
		switch statusFilter {
		case "green":
			opts.Status = componentv1.ComponentE2EProfileSpec_green
		case "deferred":
			opts.Status = componentv1.ComponentE2EProfileSpec_deferred
		case "skip":
			opts.Status = componentv1.ComponentE2EProfileSpec_skip
		case "stub":
			opts.Status = componentv1.ComponentE2EProfileSpec_stub
		default:
			return fmt.Errorf("unknown status %q: must be green, deferred, skip, or stub", statusFilter)
		}
	}

	if provisionerFilter != "" {
		switch provisionerFilter {
		case "pulumi":
			opts.Provisioner = sharedpb.IacProvisioner_pulumi
		case "terraform":
			opts.Provisioner = sharedpb.IacProvisioner_terraform
		case "tofu":
			opts.Provisioner = sharedpb.IacProvisioner_tofu
		default:
			return fmt.Errorf("unknown provisioner %q: must be pulumi, terraform, or tofu", provisionerFilter)
		}
	}

	result, err := profile.Discover(repoRoot, providerName, opts)
	if err != nil {
		return err
	}

	if outputMode == "" {
		outputMode = autoDetectOutput()
	}

	switch outputMode {
	case "github-matrix":
		return renderGitHubMatrix(result)
	case "table":
		return e2ediscover.RenderTable(os.Stdout, result)
	case "interactive":
		return e2ediscover.RunInteractive(result)
	default:
		return fmt.Errorf("unknown output format %q: must be interactive, table, or github-matrix", outputMode)
	}
}

func renderGitHubMatrix(result *profile.DiscoverResult) error {
	matrix := profile.BuildGitHubMatrix(result)
	jsonStr, err := profile.MatrixJSON(matrix)
	if err != nil {
		return err
	}
	fmt.Println(jsonStr)
	return nil
}

func autoDetectOutput() string {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return "interactive"
	}
	return "table"
}

func detectRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk up from cwd looking for go.mod with planton module name.
	// For simplicity, just use cwd -- the user should run from the repo root
	// or the CLI can be enhanced later to walk up directories.
	if _, err := os.Stat(cwd + "/go.mod"); err == nil {
		return cwd, nil
	}

	return cwd, nil
}
