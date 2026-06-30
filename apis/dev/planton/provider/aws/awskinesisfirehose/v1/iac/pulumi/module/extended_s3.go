package module

import (
	awskinesisfirehose "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awskinesisfirehose/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/kinesis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildExtendedS3Args constructs the Extended S3 destination configuration.
func buildExtendedS3Args(dest *awskinesisfirehose.AwsKinesisFirehoseExtendedS3Destination, locals *Locals) (*kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationArgs, error) {
	args := &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationArgs{
		BucketArn: pulumi.String(dest.BucketArn.GetValue()),
		RoleArn:   pulumi.String(dest.RoleArn.GetValue()),
	}

	if dest.Prefix != "" {
		args.Prefix = pulumi.StringPtr(dest.Prefix)
	}
	if dest.ErrorOutputPrefix != "" {
		args.ErrorOutputPrefix = pulumi.StringPtr(dest.ErrorOutputPrefix)
	}
	if dest.CompressionFormat != "" {
		args.CompressionFormat = pulumi.StringPtr(dest.CompressionFormat)
	}
	if dest.KmsKeyArn != nil {
		args.KmsKeyArn = pulumi.StringPtr(dest.KmsKeyArn.GetValue())
	}
	if dest.CustomTimeZone != "" {
		args.CustomTimeZone = pulumi.StringPtr(dest.CustomTimeZone)
	}
	if dest.FileExtension != "" {
		args.FileExtension = pulumi.StringPtr(dest.FileExtension)
	}

	// Buffering hints
	if b := dest.Buffering; b != nil {
		if b.IntervalInSeconds > 0 {
			args.BufferingInterval = pulumi.IntPtr(int(b.IntervalInSeconds))
		}
		if b.SizeInMbs > 0 {
			args.BufferingSize = pulumi.IntPtr(int(b.SizeInMbs))
		}
	}

	// S3 backup mode
	if dest.S3BackupMode != "" {
		args.S3BackupMode = pulumi.StringPtr(dest.S3BackupMode)
	}
	if dest.S3Backup != nil {
		args.S3BackupConfiguration = buildS3BackupConfig(dest.S3Backup)
	}

	// Processing
	if proc := buildProcessingConfig(dest.Processing); proc != nil {
		args.ProcessingConfiguration = proc
	}

	// CloudWatch logging
	if log := buildCloudwatchLogging(dest.Logging); log != nil {
		args.CloudwatchLoggingOptions = log
	}

	// Dynamic partitioning
	if dp := dest.DynamicPartitioning; dp != nil && dp.Enabled {
		dpArgs := &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDynamicPartitioningConfigurationArgs{
			Enabled: pulumi.BoolPtr(true),
		}
		if dp.RetryDurationInSeconds > 0 {
			dpArgs.RetryDuration = pulumi.IntPtr(int(dp.RetryDurationInSeconds))
		}
		args.DynamicPartitioningConfiguration = dpArgs
	}

	// Data format conversion
	if dfc := dest.DataFormatConversion; dfc != nil && dfc.Enabled {
		args.DataFormatConversionConfiguration = buildDataFormatConversion(dfc)
	}

	return args, nil
}

// buildS3BackupConfig constructs the S3 backup configuration from the shared
// AwsKinesisFirehoseS3Config proto message.
func buildS3BackupConfig(cfg *awskinesisfirehose.AwsKinesisFirehoseS3Config) *kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationS3BackupConfigurationArgs {
	args := &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationS3BackupConfigurationArgs{
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

// buildDataFormatConversion constructs the data format conversion configuration
// for Extended S3 (Parquet/ORC via Glue catalog).
func buildDataFormatConversion(dfc *awskinesisfirehose.AwsKinesisFirehoseDataFormatConversion) *kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationArgs {
	args := &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationArgs{
		Enabled: pulumi.BoolPtr(true),
	}

	// Input format (deserializer)
	inputFormat := dfc.InputFormat
	if inputFormat == "" {
		inputFormat = "OPENX_JSON"
	}

	switch inputFormat {
	case "HIVE_JSON":
		args.InputFormatConfiguration = &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationInputFormatConfigurationArgs{
			Deserializer: &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationInputFormatConfigurationDeserializerArgs{
				HiveJsonSerDe: &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationInputFormatConfigurationDeserializerHiveJsonSerDeArgs{},
			},
		}
	default: // OPENX_JSON
		args.InputFormatConfiguration = &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationInputFormatConfigurationArgs{
			Deserializer: &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationInputFormatConfigurationDeserializerArgs{
				OpenXJsonSerDe: &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationInputFormatConfigurationDeserializerOpenXJsonSerDeArgs{},
			},
		}
	}

	// Output format (serializer)
	switch dfc.OutputFormat {
	case "ORC":
		orcArgs := &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationOutputFormatConfigurationSerializerOrcSerDeArgs{}
		if dfc.OrcCompression != "" {
			orcArgs.Compression = pulumi.StringPtr(dfc.OrcCompression)
		}
		args.OutputFormatConfiguration = &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationOutputFormatConfigurationArgs{
			Serializer: &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationOutputFormatConfigurationSerializerArgs{
				OrcSerDe: orcArgs,
			},
		}
	default: // PARQUET
		parquetArgs := &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationOutputFormatConfigurationSerializerParquetSerDeArgs{}
		if dfc.ParquetCompression != "" {
			parquetArgs.Compression = pulumi.StringPtr(dfc.ParquetCompression)
		}
		args.OutputFormatConfiguration = &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationOutputFormatConfigurationArgs{
			Serializer: &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationOutputFormatConfigurationSerializerArgs{
				ParquetSerDe: parquetArgs,
			},
		}
	}

	// Schema configuration (Glue catalog)
	if schema := dfc.Schema; schema != nil {
		schemaArgs := &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationSchemaConfigurationArgs{
			DatabaseName: pulumi.String(schema.DatabaseName),
			TableName:    pulumi.String(schema.TableName),
			RoleArn:      pulumi.String(schema.RoleArn.GetValue()),
		}
		if schema.CatalogId != "" {
			schemaArgs.CatalogId = pulumi.StringPtr(schema.CatalogId)
		}
		if schema.Region != "" {
			schemaArgs.Region = pulumi.StringPtr(schema.Region)
		}
		if schema.VersionId != "" {
			schemaArgs.VersionId = pulumi.StringPtr(schema.VersionId)
		}
		args.SchemaConfiguration = schemaArgs
	}

	return args
}
