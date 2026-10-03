package module

import (
	"github.com/pkg/errors"
	awsnatgatewayv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsnatgateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awsnatgatewayv1alpha1.AwsNatGatewayIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsNatGateway.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if _, err := natGateway(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create nat gateway")
	}

	return nil
}
