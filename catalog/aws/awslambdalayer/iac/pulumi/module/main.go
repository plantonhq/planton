package module

import (
	"github.com/pkg/errors"
	awslambdalayerv1alpha1 "github.com/plantonhq/planton/catalog/aws/awslambdalayer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates publishing the layer version with its share
// grants, and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awslambdalayerv1alpha1.AwsLambdaLayerIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := layer(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "layer")
	}

	return nil
}
