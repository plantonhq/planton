package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrusttunnelv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrusttunnel/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—kept small to mirror a Terraform module's main.tf.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarezerotrusttunnelv1alpha1.CloudflareZeroTrustTunnelIacInput,
) error {
	locals := initializeLocals(ctx, iacInput)

	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	if err := tunnel(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare tunnel")
	}

	return nil
}
