package module

import (
	"github.com/pkg/errors"
	cloudflareemailroutingrulev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareemailroutingrule/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—kept small to mirror a Terraform module's main.tf.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflareemailroutingrulev1alpha1.CloudflareEmailRoutingRuleIacInput,
) error {
	locals := initializeLocals(ctx, iacInput)

	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	if _, err := emailRoutingRule(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare email routing rule")
	}

	return nil
}
