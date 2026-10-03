package module

import (
	"github.com/pkg/errors"
	awsinternetgatewayv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsinternetgateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awsinternetgatewayv1alpha1.AwsInternetGatewayIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsInternetGateway.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if _, err := internetGateway(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create internet gateway")
	}

	return nil
}
