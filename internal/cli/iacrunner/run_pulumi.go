package iacrunner

import (
	"os"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/internal/cli/ui"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumistack"
	"github.com/plantonhq/planton/pkg/outputs"
	"github.com/plantonhq/planton/shared/iac/pulumi"
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

	// The backend-url flag joins the resolution chain inside Run
	// (flag > manifest annotation > PLANTON_BACKEND_URL env).
	backendUrl, _ := cmd.Flags().GetString(string(flag.BackendUrl))

	// A real update captures the stack's outputs so they can be shown
	// (masked) after success — previews and the other operations do not.
	runOpts := []pulumistack.RunOption{pulumistack.WithBackendURL(backendUrl)}
	var captured *outputs.CaptureResult
	if operation == pulumi.PulumiOperationType_update && !isPreview {
		captured = &outputs.CaptureResult{}
		runOpts = append(runOpts, pulumistack.WithOutputCapture(captured))
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
		ctx.IacInputFilePath,
		ctx.ProviderConfig,
		runOpts...,
	)
	if err != nil {
		ui.EngineExecutionFailed("Pulumi", err)
		os.Exit(1)
	}
	cliprint.PrintPulumiSuccess()
	ui.OutputsSummary(captured)
	return nil
}
