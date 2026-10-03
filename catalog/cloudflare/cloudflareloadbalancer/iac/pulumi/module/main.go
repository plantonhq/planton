package module

import (
	"github.com/pkg/errors"
	cloudflareloadbalancerv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareloadbalancer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry‑point. It prepares locals, sets up the provider,
// then provisions the load balancer.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflareloadbalancerv1alpha1.CloudflareLoadBalancerIacInput,
) error {
	// 1. Gather handy references.
	locals := initializeLocals(ctx, iacInput)

	// 2. Create a Cloudflare provider from the supplied credential.
	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	// 3. Provision the zone-scoped load balancer (pools/monitors are separate kinds).
	if _, err := load_balancer(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare load balancer")
	}

	return nil
}
