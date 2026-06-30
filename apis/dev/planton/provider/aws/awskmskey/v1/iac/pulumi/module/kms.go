package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/kms"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type KmsKeyResult struct {
	KeyId           pulumi.StringOutput
	KeyArn          pulumi.StringOutput
	AliasName       pulumi.StringOutput
	RotationEnabled pulumi.BoolPtrOutput
}

func kmsKey(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*KmsKeyResult, error) {
	spec := locals.AwsKmsKey.Spec

	var keySpec pulumi.StringPtrInput
	switch spec.KeySpec {
	case 0:
		keySpec = pulumi.StringPtr("SYMMETRIC_DEFAULT")
	case 1:
		keySpec = pulumi.StringPtr("RSA_2048")
	case 2:
		keySpec = pulumi.StringPtr("RSA_4096")
	case 3:
		keySpec = pulumi.StringPtr("ECC_NIST_P256")
	default:
		keySpec = pulumi.StringPtr("SYMMETRIC_DEFAULT")
	}

	createdKey, err := kms.NewKey(ctx, locals.AwsKmsKey.Metadata.Name, &kms.KeyArgs{
		Description:           pulumi.StringPtr(spec.Description),
		DeletionWindowInDays:  pulumi.IntPtr(int(spec.DeletionWindowDays)),
		EnableKeyRotation:     pulumi.BoolPtr(!spec.DisableKeyRotation),
		CustomerMasterKeySpec: keySpec,
		Tags:                  pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kms key")
	}

	var aliasNameOutput pulumi.StringOutput
	if spec.AliasName != "" {
		_, err := kms.NewAlias(ctx, locals.AwsKmsKey.Metadata.Name+"-alias", &kms.AliasArgs{
			Name:        pulumi.String(spec.AliasName),
			TargetKeyId: createdKey.KeyId,
		}, pulumi.Provider(provider))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create kms alias")
		}
		aliasNameOutput = pulumi.String(spec.AliasName).ToStringOutput()
	} else {
		aliasNameOutput = pulumi.String("").ToStringOutput()
	}

	return &KmsKeyResult{
		KeyId:           createdKey.KeyId,
		KeyArn:          createdKey.Arn,
		AliasName:       aliasNameOutput,
		RotationEnabled: createdKey.EnableKeyRotation,
	}, nil
}
