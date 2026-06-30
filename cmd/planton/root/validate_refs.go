//go:build !codegen
// +build !codegen

package root

import (
	"fmt"
	"os"

	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/pkg/refcheck"
	"github.com/spf13/cobra"
)

var ValidateRefs = &cobra.Command{
	Use:   "validate-refs",
	Short: "Validate that every foreign-key reference resolves to a real field on the referenced kind",
	Long: `Walk every production cloud-resource kind and check each field annotated with
(dev.planton.shared.foreignkey.v1.default_kind_field_path): the path must resolve
against the referenced kind's resolved target -- its status.outputs message for
"status.outputs.*" paths, or its spec for "spec.*" paths.

A dangling path is a composition that silently fails to resolve at deploy time (the
orchestrator reads the referenced output and finds nothing), so this is a hard
invariant. Use --check in CI to fail the build on any dangling reference.`,
	Run: validateRefsHandler,
}

func init() {
	ValidateRefs.Flags().Bool("check", false, "exit non-zero if any foreign-key reference does not resolve (for CI)")
}

func validateRefsHandler(cmd *cobra.Command, _ []string) {
	findings := refcheck.Analyze()

	if len(findings) == 0 {
		cliprint.PrintSuccess("foreign-key references: all resolve")
		return
	}

	for _, f := range findings {
		cliprint.PrintError(fmt.Sprintf("%s:%s -> %s %q -- %s", f.Kind, f.FieldPath, f.TargetKind, f.RefPath, f.Reason))
	}
	fmt.Printf("\n%d dangling foreign-key reference(s)\n", len(findings))

	if check, _ := cmd.Flags().GetBool("check"); check {
		os.Exit(1)
	}
}
