package module

import (
	"github.com/pkg/errors"
	awsnlbv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsnlb/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the primary entry point for the AwsNlb Pulumi
// module. It creates the NLB and optional DNS records; listeners and target
// groups are separate resources that attach to the NLB by ARN.
func Resources(ctx *pulumi.Context, iacInput *awsnlbv1alpha1.AwsNlbIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Nlb.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	nlbResource, err := nlb(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create Network Load Balancer")
	}

	if locals.Nlb.Spec.Dns != nil && locals.Nlb.Spec.Dns.Enabled {
		if err := dns(ctx, locals, provider, nlbResource); err != nil {
			return errors.Wrap(err, "failed to configure DNS")
		}
	}

	return nil
}
