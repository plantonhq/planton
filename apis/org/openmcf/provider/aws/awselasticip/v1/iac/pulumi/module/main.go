package module

import (
	"github.com/pkg/errors"
	awselasticipv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awselasticip/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awselasticipv1.AwsElasticIpStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(awsProviderConfig.GetRegion()),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	result, err := eip(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create elastic ip")
	}

	ctx.Export(OpAllocationId, result.AllocationId)
	ctx.Export(OpPublicIp, result.PublicIp)
	ctx.Export(OpArn, result.Arn)
	ctx.Export(OpPublicDns, result.PublicDns)

	return nil
}
