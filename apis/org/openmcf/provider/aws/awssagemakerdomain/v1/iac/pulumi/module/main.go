package module

import (
	"github.com/pkg/errors"
	awssagemakerdomainv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awssagemakerdomain/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS SageMaker Domain related resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awssagemakerdomainv1.AwsSagemakerDomainStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsSagemakerDomain.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsSagemakerDomain.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	// SageMaker Domain
	createdDomain, err := domain(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "sagemaker domain")
	}

	// Export outputs
	outputs(ctx, createdDomain)

	return nil
}
