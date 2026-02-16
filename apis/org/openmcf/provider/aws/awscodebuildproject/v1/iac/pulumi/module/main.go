package module

import (
	"github.com/pkg/errors"
	awscodebuildprojectv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscodebuildproject/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS CodeBuild resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awscodebuildprojectv1.AwsCodeBuildProjectStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(awsProviderConfig.GetRegion()),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
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
