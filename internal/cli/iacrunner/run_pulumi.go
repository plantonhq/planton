package iacrunner

import (
	"os"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/iac/pulumi"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumistack"
	"github.com/spf13/cobra"
)

// RunPulumi executes a Pulumi operation using the resolved context.
func RunPulumi(ctx *Context, cmd *cobra.Command, operation pulumi.PulumiOperationType, isPreview bool) error {
	// Stack can be provided via flag or extracted from manifest
	stackFqdn, err := cmd.Flags().GetString(string(flag.Stack))
	if err != nil {
		return errors.Wrap(err, "failed to get stack flag")
	}

	// Get auto-approve behavior
	isAutoApprove := !isPreview
	if yes, _ := cmd.Flags().GetBool(string(flag.Yes)); yes {
		isAutoApprove = true
	}

	err = pulumistack.Run(
		ctx.ModuleDir,
		stackFqdn,
		ctx.ManifestPath,
		operation,
		isPreview,
		isAutoApprove,
		ctx.ValueOverrides,
		ctx.ShowDiff,
		ctx.ModuleVersion,
		ctx.NoCleanup,
		ctx.KubeContext,
		ctx.StackInputFilePath,
		ctx.ProviderConfig,
	)
	if err != nil {
		cliprint.PrintPulumiFailure()
		os.Exit(1)
	}
	cliprint.PrintPulumiSuccess()
	return nil
}
