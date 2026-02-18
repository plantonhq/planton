package module

import (
	"github.com/pkg/errors"
	awsecsservicev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsecsservice/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It reads the AwsEcsServiceStackInput, initializes locals, configures
// an AWS provider, then calls the service(...) function.
func Resources(ctx *pulumi.Context, stackInput *awsecsservicev1.AwsEcsServiceStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsEcsService.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsEcsService.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	// Call the service(...) function to create the AWS ECS Service resource.
	if err := ecsService(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create aws ecs service resource")
	}

	return nil
}
