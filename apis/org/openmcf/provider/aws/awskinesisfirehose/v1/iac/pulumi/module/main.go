package module

import (
	"github.com/pkg/errors"
	awskinesisfirehose "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awskinesisfirehose/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates Kinesis Data Firehose delivery stream creation and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awskinesisfirehose.AwsKinesisFirehoseStackInput) error {
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

	stream, err := deliveryStream(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "firehose delivery stream")
	}

	if err := outputs(ctx, stream); err != nil {
		return errors.Wrap(err, "outputs")
	}

	return nil
}
