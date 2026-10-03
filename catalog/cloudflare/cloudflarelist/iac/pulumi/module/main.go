package module

import (
	"github.com/pkg/errors"
	cloudflarelistv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarelist/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—kept small to mirror a Terraform module's main.tf.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarelistv1alpha1.CloudflareListIacInput,
) error {
	locals := initializeLocals(ctx, iacInput)

	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	if _, err := list(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare list")
	}

	return nil
}
