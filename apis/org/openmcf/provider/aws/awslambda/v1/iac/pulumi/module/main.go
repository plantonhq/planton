package module

import (
	"github.com/pkg/errors"
	awslambdav1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awslambda/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point that prepares locals, initialises the AWS provider,
// orchestrates the Lambda function creation, and exports outputs as defined in AwsLambdaStackOutputs.
func Resources(ctx *pulumi.Context, stackInput *awslambdav1.AwsLambdaStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsLambda.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsLambda.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	// Create the Lambda function (and supporting log group)
	if _, err := lambdaFunction(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create lambda function")
	}

	return nil
}
