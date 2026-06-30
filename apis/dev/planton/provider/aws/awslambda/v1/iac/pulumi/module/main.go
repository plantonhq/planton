package module

import (
	"github.com/pkg/errors"
	awslambdav1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awslambda/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point that prepares locals, initialises the AWS provider,
// orchestrates the Lambda function creation, and exports outputs as defined in AwsLambdaStackOutputs.
func Resources(ctx *pulumi.Context, stackInput *awslambdav1.AwsLambdaStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsLambda.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Create the Lambda function (and supporting log group)
	if _, err := lambdaFunction(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create lambda function")
	}

	return nil
}
