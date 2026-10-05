package module

import (
	"github.com/pkg/errors"
	awscodepipelinev1alpha1 "github.com/plantonhq/planton/catalog/aws/awscodepipeline/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS CodePipeline resources and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awscodepipelinev1alpha1.AwsCodePipelineIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsCodePipeline.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdPipeline, err := pipeline(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "codepipeline")
	}

	// Export outputs -- identical names and semantics to the Terraform
	// module (the cross-engine output contract).
	ctx.Export(OpPipelineArn, createdPipeline.Arn)
	ctx.Export(OpPipelineName, createdPipeline.Name)

	return nil
}
