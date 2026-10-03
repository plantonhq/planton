package module

import (
	"github.com/pkg/errors"
	cloudflareloadbalancerpoolv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareloadbalancerpool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—kept small to mirror a Terraform module's main.tf.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflareloadbalancerpoolv1alpha1.CloudflareLoadBalancerPoolIacInput,
) error {
	locals := initializeLocals(ctx, iacInput)

	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	if _, err := pool(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare load balancer pool")
	}

	return nil
}
