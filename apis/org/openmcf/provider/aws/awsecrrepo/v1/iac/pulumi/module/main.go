package module

import (
	"github.com/pkg/errors"
	awsecrrepov1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsecrrepo/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point for the aws_ecr_repo Pulumi module.
// It initializes locals, configures a provider (default or custom), then calls ecrRepo.
func Resources(ctx *pulumi.Context, stackInput *awsecrrepov1.AwsEcrRepoStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsEcrRepo.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsEcrRepo.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	if err := ecrRepo(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create aws_ecr_repo resource")
	}

	return nil
}
