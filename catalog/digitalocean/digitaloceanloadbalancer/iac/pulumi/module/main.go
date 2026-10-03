package module

import (
	"github.com/pkg/errors"
	digitaloceanloadbalancerv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanloadbalancer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—mirrors the pattern used in the VPC module.
func Resources(
	ctx *pulumi.Context,
	iacInput *digitaloceanloadbalancerv1alpha1.DigitalOceanLoadBalancerIacInput,
) error {
	// 1. Prepare locals (metadata, labels, credentials, etc.).
	locals := initializeLocals(ctx, iacInput)

	// 2. Create a DigitalOcean provider from the supplied credential.
	digitalOceanProvider, err := pulumidigitaloceanprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup digitalocean provider")
	}

	// 3. Create the Load Balancer.
	if _, err := loadBalancer(ctx, locals, digitalOceanProvider); err != nil {
		return errors.Wrap(err, "failed to create load balancer")
	}

	return nil
}
