package module

import (
	"github.com/pkg/errors"
	awsegressonlyinternetgatewayv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsegressonlyinternetgateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awsegressonlyinternetgatewayv1alpha1.AwsEgressOnlyInternetGatewayIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsEgressOnlyInternetGateway.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if _, err := egressOnlyInternetGateway(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create egress-only internet gateway")
	}

	return nil
}
