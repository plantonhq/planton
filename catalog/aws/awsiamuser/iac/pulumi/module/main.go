package module

import (
	"github.com/pkg/errors"
	awsiamuserv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsiamuser/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awsiamuserv1alpha1.AwsIamUserIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsIamUser.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Create IAM user and related resources
	results, err := iamUser(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create iam user")
	}

	// Export outputs
	ctx.Export(OpUserArn, results.UserArn)
	ctx.Export(OpUserName, results.UserName)
	ctx.Export(OpUserId, results.UserId)
	ctx.Export(OpConsoleUrl, results.ConsoleUrl)
	ctx.Export(OpAccessKeyId, results.AccessKeyId)
	ctx.Export(OpSecretAccessKey, pulumi.ToSecret(results.SecretAccessKey))
	ctx.Export(OpAccessKeyStatus, results.AccessKeyStatus)

	return nil
}
