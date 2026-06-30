package module

import (
	"github.com/pkg/errors"
	awsnlbv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsnetworkloadbalancer/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the primary entry point for the AwsNetworkLoadBalancer Pulumi
// module. It creates the NLB, listeners with target groups, and optional DNS.
func Resources(ctx *pulumi.Context, stackInput *awsnlbv1.AwsNetworkLoadBalancerStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.Nlb.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	nlbResource, err := nlb(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create Network Load Balancer")
	}

	if err := listeners(ctx, locals, provider, nlbResource); err != nil {
		return errors.Wrap(err, "failed to create listeners and target groups")
	}

	if locals.Nlb.Spec.Dns != nil && locals.Nlb.Spec.Dns.Enabled {
		if err := dns(ctx, locals, provider, nlbResource); err != nil {
			return errors.Wrap(err, "failed to configure DNS")
		}
	}

	return nil
}
