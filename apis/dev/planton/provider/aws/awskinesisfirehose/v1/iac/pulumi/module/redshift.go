package module

import (
	awskinesisfirehose "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awskinesisfirehose/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/kinesis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildRedshiftArgs constructs the Redshift destination configuration.
func buildRedshiftArgs(dest *awskinesisfirehose.AwsKinesisFirehoseRedshiftDestination, locals *Locals) (*kinesis.FirehoseDeliveryStreamRedshiftConfigurationArgs, error) {
	args := &kinesis.FirehoseDeliveryStreamRedshiftConfigurationArgs{
		ClusterJdbcurl: pulumi.String(dest.ClusterJdbcurl),
		RoleArn:        pulumi.String(dest.RoleArn.GetValue()),
		DataTableName:  pulumi.String(dest.DataTableName),
	}

	if dest.DataTableColumns != "" {
		args.DataTableColumns = pulumi.StringPtr(dest.DataTableColumns)
	}
	if dest.CopyOptions != "" {
		args.CopyOptions = pulumi.StringPtr(dest.CopyOptions)
	}
	if dest.Username != "" {
		args.Username = pulumi.StringPtr(dest.Username)
	}
	if dest.Password != nil {
		args.Password = pulumi.StringPtr(dest.Password.GetValue())
	}

	if dest.RetryDurationInSeconds > 0 {
		args.RetryDuration = pulumi.IntPtr(int(dest.RetryDurationInSeconds))
	}

	// S3 intermediate config (required for Redshift COPY)
	if cfg := dest.S3Config; cfg != nil {
		args.S3Configuration = buildRedshiftS3Config(cfg)
	}

	// S3 backup
	if dest.S3BackupMode != "" {
		args.S3BackupMode = pulumi.StringPtr(dest.S3BackupMode)
	}
	if dest.S3Backup != nil {
		args.S3BackupConfiguration = buildRedshiftS3BackupConfig(dest.S3Backup)
	}

	// Processing
	if dest.Processing != nil && dest.Processing.Enabled {
		args.ProcessingConfiguration = buildRedshiftProcessingConfig(dest.Processing)
	}

	// CloudWatch logging
	if dest.Logging != nil && dest.Logging.Enabled {
		args.CloudwatchLoggingOptions = buildRedshiftCloudwatchLogging(dest.Logging)
	}

	return args, nil
}

// buildRedshiftS3Config builds the S3 intermediate staging configuration for Redshift.
func buildRedshiftS3Config(cfg *awskinesisfirehose.AwsKinesisFirehoseS3Config) *kinesis.FirehoseDeliveryStreamRedshiftConfigurationS3ConfigurationArgs {
	args := &kinesis.FirehoseDeliveryStreamRedshiftConfigurationS3ConfigurationArgs{
		BucketArn: pulumi.String(cfg.BucketArn.GetValue()),
		RoleArn:   pulumi.String(cfg.RoleArn.GetValue()),
	}
	if cfg.Prefix != "" {
		args.Prefix = pulumi.StringPtr(cfg.Prefix)
	}
	if cfg.ErrorOutputPrefix != "" {
		args.ErrorOutputPrefix = pulumi.StringPtr(cfg.ErrorOutputPrefix)
	}
	if cfg.CompressionFormat != "" {
		args.CompressionFormat = pulumi.StringPtr(cfg.CompressionFormat)
	}
	if cfg.KmsKeyArn != nil {
		args.KmsKeyArn = pulumi.StringPtr(cfg.KmsKeyArn.GetValue())
	}
	if b := cfg.Buffering; b != nil {
		if b.IntervalInSeconds > 0 {
			args.BufferingInterval = pulumi.IntPtr(int(b.IntervalInSeconds))
		}
		if b.SizeInMbs > 0 {
			args.BufferingSize = pulumi.IntPtr(int(b.SizeInMbs))
		}
	}
	return args
}

// buildRedshiftS3BackupConfig builds the S3 backup configuration for Redshift source records.
func buildRedshiftS3BackupConfig(cfg *awskinesisfirehose.AwsKinesisFirehoseS3Config) *kinesis.FirehoseDeliveryStreamRedshiftConfigurationS3BackupConfigurationArgs {
	args := &kinesis.FirehoseDeliveryStreamRedshiftConfigurationS3BackupConfigurationArgs{
		BucketArn: pulumi.String(cfg.BucketArn.GetValue()),
		RoleArn:   pulumi.String(cfg.RoleArn.GetValue()),
	}
	if cfg.Prefix != "" {
		args.Prefix = pulumi.StringPtr(cfg.Prefix)
	}
	if cfg.ErrorOutputPrefix != "" {
		args.ErrorOutputPrefix = pulumi.StringPtr(cfg.ErrorOutputPrefix)
	}
	if cfg.CompressionFormat != "" {
		args.CompressionFormat = pulumi.StringPtr(cfg.CompressionFormat)
	}
	if cfg.KmsKeyArn != nil {
		args.KmsKeyArn = pulumi.StringPtr(cfg.KmsKeyArn.GetValue())
	}
	if b := cfg.Buffering; b != nil {
		if b.IntervalInSeconds > 0 {
			args.BufferingInterval = pulumi.IntPtr(int(b.IntervalInSeconds))
		}
		if b.SizeInMbs > 0 {
			args.BufferingSize = pulumi.IntPtr(int(b.SizeInMbs))
		}
	}
	return args
}

// buildRedshiftProcessingConfig constructs Lambda processing for Redshift.
func buildRedshiftProcessingConfig(proc *awskinesisfirehose.AwsKinesisFirehoseLambdaProcessing) *kinesis.FirehoseDeliveryStreamRedshiftConfigurationProcessingConfigurationArgs {
	if proc == nil || !proc.Enabled {
		return nil
	}
	params := kinesis.FirehoseDeliveryStreamRedshiftConfigurationProcessingConfigurationProcessorParameterArray{
		&kinesis.FirehoseDeliveryStreamRedshiftConfigurationProcessingConfigurationProcessorParameterArgs{
			ParameterName:  pulumi.String("LambdaArn"),
			ParameterValue: pulumi.String(proc.LambdaArn.GetValue()),
		},
	}
	return &kinesis.FirehoseDeliveryStreamRedshiftConfigurationProcessingConfigurationArgs{
		Enabled: pulumi.Bool(true),
		Processors: kinesis.FirehoseDeliveryStreamRedshiftConfigurationProcessingConfigurationProcessorArray{
			&kinesis.FirehoseDeliveryStreamRedshiftConfigurationProcessingConfigurationProcessorArgs{
				Type:       pulumi.String("Lambda"),
				Parameters: params,
			},
		},
	}
}

// buildRedshiftCloudwatchLogging constructs CloudWatch logging for Redshift.
func buildRedshiftCloudwatchLogging(logging *awskinesisfirehose.AwsKinesisFirehoseCloudwatchLogging) *kinesis.FirehoseDeliveryStreamRedshiftConfigurationCloudwatchLoggingOptionsArgs {
	if logging == nil || !logging.Enabled {
		return nil
	}
	return &kinesis.FirehoseDeliveryStreamRedshiftConfigurationCloudwatchLoggingOptionsArgs{
		Enabled:       pulumi.Bool(true),
		LogGroupName:  pulumi.StringPtr(logging.LogGroupName),
		LogStreamName: pulumi.StringPtr(logging.LogStreamName),
	}
}
