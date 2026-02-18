package module

import (
	"github.com/pkg/errors"
	awscloudfrontv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscloudfront/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awscloudfrontv1.AwsCloudFrontStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.Target.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.Target.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	dist, err := createDistribution(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "create cloudfront distribution")
	}

	// Export outputs mapped to AwsCloudFrontStackOutputs
	ctx.Export(OpDistributionId, dist.ID())
	ctx.Export(OpDomainName, dist.DomainName)
	ctx.Export(OpHostedZoneId, pulumi.String("Z2FDTNDATAQYW2"))

	return nil
}
