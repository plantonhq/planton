package module

import (
	"github.com/pkg/errors"
	awscodepipelinev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awscodepipeline/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS CodePipeline resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awscodepipelinev1.AwsCodePipelineStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsCodePipeline.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdPipeline, err := pipeline(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "codepipeline")
	}

	ctx.Export(OpPipelineArn, createdPipeline.Arn)
	ctx.Export(OpPipelineName, createdPipeline.Name)

	return nil
}
