package module

import (
	"github.com/pkg/errors"
	awskmskeyv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awskmskey/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awskmskeyv1.AwsKmsKeyStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsKmsKey.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Create KMS key and optional alias
	result, err := kmsKey(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create kms key")
	}

	// Export outputs
	ctx.Export(OpKeyId, result.KeyId)
	ctx.Export(OpKeyArn, result.KeyArn)
	ctx.Export(OpAliasName, result.AliasName)
	ctx.Export(OpRotationEnabled, result.RotationEnabled)

	return nil
}
