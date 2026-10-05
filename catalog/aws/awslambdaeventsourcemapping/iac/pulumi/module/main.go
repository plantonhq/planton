package module

import (
	"github.com/pkg/errors"
	awslambdaeventsourcemappingv1alpha1 "github.com/plantonhq/planton/catalog/aws/awslambdaeventsourcemapping/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the Lambda event source mapping and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awslambdaeventsourcemappingv1alpha1.AwsLambdaEventSourceMappingIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web identity,
	// or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsLambdaEventSourceMapping.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdMapping, err := mapping(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create event source mapping")
	}

	ctx.Export(OpUuid, createdMapping.Uuid)
	ctx.Export(OpMappingArn, createdMapping.Arn)
	ctx.Export(OpFunctionArn, createdMapping.FunctionArn)
	ctx.Export(OpState, createdMapping.State)

	return nil
}
