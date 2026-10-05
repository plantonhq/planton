package module

import (
	"github.com/pkg/errors"
	awskmskeyv1alpha1 "github.com/plantonhq/planton/catalog/aws/awskmskey/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awskmskeyv1alpha1.AwsKmsKeyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsKmsKey.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	result, err := kmsKey(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create kms key")
	}

	ctx.Export(OpKeyId, result.KeyId)
	ctx.Export(OpKeyArn, result.KeyArn)
	ctx.Export(OpAliasNames, result.AliasNames)
	ctx.Export(OpGrantIds, result.GrantIds)

	return nil
}
