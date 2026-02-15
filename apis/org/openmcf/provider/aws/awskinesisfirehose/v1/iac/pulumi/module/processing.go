package module

import (
	"fmt"

	awskinesisfirehose "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awskinesisfirehose/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/kinesis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildProcessingConfig constructs the Pulumi processing configuration from the
// proto Lambda processing spec. Returns nil if processing is not configured or
// not enabled.
func buildProcessingConfig(proc *awskinesisfirehose.AwsKinesisFirehoseLambdaProcessing) *kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationArgs {
	if proc == nil || !proc.Enabled {
		return nil
	}

	params := kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorParameterArray{
		&kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorParameterArgs{
			ParameterName:  pulumi.String("LambdaArn"),
			ParameterValue: pulumi.String(proc.LambdaArn.GetValue()),
		},
	}

	if proc.BufferSizeInMbs > 0 {
		params = append(params, &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorParameterArgs{
			ParameterName:  pulumi.String("BufferSizeInMBs"),
			ParameterValue: pulumi.String(fmt.Sprintf("%d", proc.BufferSizeInMbs)),
		})
	}

	if proc.BufferIntervalInSeconds > 0 {
		params = append(params, &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorParameterArgs{
			ParameterName:  pulumi.String("BufferIntervalInSeconds"),
			ParameterValue: pulumi.String(fmt.Sprintf("%d", proc.BufferIntervalInSeconds)),
		})
	}

	if proc.NumberOfRetries > 0 {
		params = append(params, &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorParameterArgs{
			ParameterName:  pulumi.String("NumberOfRetries"),
			ParameterValue: pulumi.String(fmt.Sprintf("%d", proc.NumberOfRetries)),
		})
	}

	return &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationArgs{
		Enabled: pulumi.Bool(true),
		Processors: kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorArray{
			&kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorArgs{
				Type:       pulumi.String("Lambda"),
				Parameters: params,
			},
		},
	}
}

// buildCloudwatchLogging constructs CloudWatch logging options. Returns nil
// if not configured or not enabled.
func buildCloudwatchLogging(logging *awskinesisfirehose.AwsKinesisFirehoseCloudwatchLogging) *kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationCloudwatchLoggingOptionsArgs {
	if logging == nil || !logging.Enabled {
		return nil
	}

	return &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationCloudwatchLoggingOptionsArgs{
		Enabled:       pulumi.Bool(true),
		LogGroupName:  pulumi.StringPtr(logging.LogGroupName),
		LogStreamName: pulumi.StringPtr(logging.LogStreamName),
	}
}
