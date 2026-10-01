package module

import (
	"github.com/pkg/errors"
	awscodebuildprojectv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscodebuildproject/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS CodeBuild resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awscodebuildprojectv1alpha1.AwsCodeBuildProjectStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AwsCodeBuildProject.Spec

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// 1. CodeBuild project (primary resource)
	createdProject, err := project(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "codebuild project")
	}

	// 2. Webhook (optional folded satellite, depends on project)
	if spec.Webhook != nil {
		createdWebhook, err := webhook(ctx, locals, provider, createdProject)
		if err != nil {
			return errors.Wrap(err, "codebuild webhook")
		}
		ctx.Export(OpWebhookUrl, createdWebhook.Url)
		ctx.Export(OpWebhookPayload, createdWebhook.PayloadUrl)
		ctx.Export(OpWebhookSecret, pulumi.ToSecret(createdWebhook.Secret))
	} else {
		// The output contract is engine-invariant: webhook outputs always
		// exist and are empty when no webhook is configured (matching the
		// Terraform module).
		ctx.Export(OpWebhookUrl, pulumi.String(""))
		ctx.Export(OpWebhookPayload, pulumi.String(""))
		ctx.Export(OpWebhookSecret, pulumi.ToSecret(pulumi.String("")))
	}

	// 3. Resource policy (optional folded satellite, depends on project)
	if spec.ResourcePolicy != nil {
		if err := resourcePolicy(ctx, locals, provider, createdProject); err != nil {
			return errors.Wrap(err, "codebuild resource policy")
		}
	}

	// Export outputs. Badge URL and public alias are computed by AWS and
	// empty in the disabled/private cases -- exported unconditionally so the
	// output shape never varies.
	ctx.Export(OpProjectArn, createdProject.Arn)
	ctx.Export(OpProjectName, createdProject.Name)
	ctx.Export(OpServiceRoleArn, pulumi.String(spec.ServiceRole.GetValue()))
	ctx.Export(OpBadgeUrl, createdProject.BadgeUrl)
	ctx.Export(OpPublicProjectAlias, createdProject.PublicProjectAlias)

	return nil
}
