package module

import (
	"github.com/pkg/errors"
	awsiamuserv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsiamuser/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awsiamuserv1.AwsIamUserStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsIamUser.Spec.Region)
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
	ctx.Export(OpSecretAccessKey, results.SecretAccessKey)

	return nil
}
