package module

import (
	"github.com/pkg/errors"
	cloudflareworkflowv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareworkflow/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflareworkflowv1alpha1.CloudflareWorkflowIacInput,
) error {
	// 1. Prepare locals (metadata, credentials).
	locals := initializeLocals(ctx, iacInput)

	// 2. Create a Cloudflare provider from the supplied credential.
	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	// 3. Register the workflow.
	if err := workflow(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create workflow")
	}

	return nil
}
