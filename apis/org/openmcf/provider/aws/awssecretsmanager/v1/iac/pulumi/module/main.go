package module

import (
	"fmt"

	"github.com/pkg/errors"
	awssecretsmanagerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awssecretsmanager/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	PlaceholderSecretValue = "placeholder"
)

func Resources(ctx *pulumi.Context, stackInput *awssecretsmanagerv1.AwsSecretsManagerStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsSecretsManager.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsSecretsManager.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	secretArnMap := pulumi.StringMap{}

	// For each secret in the input spec, create a secret in AWS Secrets Manager
	for _, secretName := range locals.AwsSecretsManager.Spec.SecretNames {
		if secretName == "" {
			continue
		}

		// Construct the secret ID to make it unique within the AWS account
		secretId := fmt.Sprintf("%s-%s", locals.AwsSecretsManager.Metadata.Id, secretName)

		createdSecret, err := createSecret(ctx, locals, provider, secretName, secretId)
		if err != nil {
			return errors.Wrapf(err, "secret %s", secretName)
		}

		secretArnMap[secretName] = createdSecret.Arn
	}

	ctx.Export(OpSecretArnMap, secretArnMap)

	return nil
}
