package module

import (
	"github.com/pkg/errors"
	awsinternetgatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsinternetgateway/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awsinternetgatewayv1.AwsInternetGatewayStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsInternetGateway.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if _, err := internetGateway(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create internet gateway")
	}

	return nil
}
