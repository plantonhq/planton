package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/secretsmanager"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createSecret creates an AWS Secrets Manager secret and seeds it with a placeholder SecretVersion.
// Returns the created Secret resource.
func createSecret(ctx *pulumi.Context, locals *Locals, provider *aws.Provider, logicalName string, secretId string) (*secretsmanager.Secret, error) {
	if logicalName == "" {
		return nil, errors.New("logicalName cannot be empty")
	}
	if secretId == "" {
		return nil, errors.New("secretId cannot be empty")
	}

	createdSecret, err := secretsmanager.NewSecret(ctx,
		logicalName,
		&secretsmanager.SecretArgs{
			Name: pulumi.String(secretId),
			Tags: pulumi.ToStringMap(locals.AwsTags),
		}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create secret")
	}

	_, err = secretsmanager.NewSecretVersion(ctx, secretId, &secretsmanager.SecretVersionArgs{
		SecretId:     createdSecret.ID(),
		SecretString: pulumi.String(PlaceholderSecretValue),
	}, pulumi.Parent(createdSecret), pulumi.IgnoreChanges([]string{"secretString"}))
	if err != nil {
		return nil, errors.Wrap(err, "create placeholder secret version")
	}

	return createdSecret, nil
}
