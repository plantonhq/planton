package module

import (
	"github.com/pkg/errors"
	awsegressonlyinternetgatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsegressonlyinternetgateway/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awsegressonlyinternetgatewayv1.AwsEgressOnlyInternetGatewayStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsEgressOnlyInternetGateway.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if _, err := egressOnlyInternetGateway(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create egress-only internet gateway")
	}

	return nil
}
