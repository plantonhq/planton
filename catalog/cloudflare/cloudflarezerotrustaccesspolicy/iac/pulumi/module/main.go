package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrustaccesspolicyv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustaccesspolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—kept small to mirror a Terraform module's main.tf.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarezerotrustaccesspolicyv1alpha1.CloudflareZeroTrustAccessPolicyIacInput,
) error {
	locals := initializeLocals(ctx, iacInput)

	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	if _, err := policy(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare zero trust access policy")
	}

	return nil
}
