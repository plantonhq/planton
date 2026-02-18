package module

import (
	"github.com/pkg/errors"
	awscloudwatchalarmv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscloudwatchalarm/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awscloudwatchalarmv1.AwsCloudwatchAlarmStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsCloudwatchAlarm.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsCloudwatchAlarm.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	result, err := alarm(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudwatch metric alarm")
	}

	ctx.Export(OpAlarmArn, result.AlarmArn)
	ctx.Export(OpAlarmName, result.AlarmName)

	return nil
}
