package module

import (
	"fmt"

	"github.com/pkg/errors"
	awssecretsmanagerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awssecretsmanager/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	PlaceholderSecretValue = "placeholder"
)

func Resources(ctx *pulumi.Context, stackInput *awssecretsmanagerv1.AwsSecretsManagerStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsSecretsManager.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
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
