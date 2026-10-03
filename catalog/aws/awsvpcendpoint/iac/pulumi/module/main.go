package module

import (
	"github.com/pkg/errors"
	awsvpcendpointv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsvpcendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the VPC endpoint. The endpoint composes onto its
// neighbors instead of embedding them: the VPC attaches by reference,
// gateway endpoints reference route tables (a subnet-owned table or the
// VPC's main/default table), and interface endpoints reference AwsSubnet
// and AwsSecurityGroup nodes. This module never modifies a resource it
// merely references -- in particular it never edits the referenced route
// tables' own routes; AWS manages the endpoint's prefix-list route as
// part of the endpoint itself.
func Resources(ctx *pulumi.Context, iacInput *awsvpcendpointv1alpha1.AwsVpcEndpointIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web identity,
	// or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsVpcEndpoint.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := vpcEndpoint(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create VPC endpoint")
	}

	return nil
}
