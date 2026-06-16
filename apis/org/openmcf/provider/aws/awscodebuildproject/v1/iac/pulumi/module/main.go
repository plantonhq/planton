package module

import (
	"github.com/pkg/errors"
	awscodebuildprojectv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscodebuildproject/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS CodeBuild resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awscodebuildprojectv1.AwsCodeBuildProjectStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsCodeBuildProject.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// 1. CodeBuild project (primary resource)
	createdProject, err := project(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "codebuild project")
	}

	// 2. Webhook (optional, depends on project)
	if locals.AwsCodeBuildProject.Spec.Webhook != nil {
		err := webhook(ctx, locals, provider, createdProject)
		if err != nil {
			return errors.Wrap(err, "codebuild webhook")
		}
	}

	// Export outputs
	ctx.Export(OpProjectArn, createdProject.Arn)
	ctx.Export(OpProjectName, createdProject.Name)
	ctx.Export(OpServiceRoleArn, pulumi.String(locals.AwsCodeBuildProject.Spec.ServiceRole.GetValue()))

	return nil
}
