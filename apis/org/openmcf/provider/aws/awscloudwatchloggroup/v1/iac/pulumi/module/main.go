package module

import (
	"github.com/pkg/errors"
	awscloudwatchloggroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscloudwatchloggroup/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awscloudwatchloggroupv1.AwsCloudwatchLogGroupStackInput) error {
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

	result, err := logGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudwatch log group")
	}

	ctx.Export(OpLogGroupArn, result.LogGroupArn)
	ctx.Export(OpLogGroupName, result.LogGroupName)

	return nil
}
