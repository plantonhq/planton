package module

import (
	awskinesisfirehose "github.com/plantonhq/planton/catalog/aws/awskinesisfirehose/v1alpha1"
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

	// Processing pipeline (normalized typed processors)
	if proc := buildExtendedS3Processing(dest.Processing); proc != nil {
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
	// CloudWatch logging for the backup S3 leg -- distinct from the
	// destination-level logging options.
	if log := cfg.Logging; log != nil && log.Enabled {
		args.CloudwatchLoggingOptions = &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationS3BackupConfigurationCloudwatchLoggingOptionsArgs{
			Enabled:       pulumi.BoolPtr(true),
			LogGroupName:  pulumi.StringPtr(log.LogGroupName),
			LogStreamName: pulumi.StringPtr(log.LogStreamName),
		}
	}
	return args
}

// buildDataFormatConversion constructs the data format conversion configuration
// for Extended S3 (Parquet/ORC via Glue catalog). The spec enforces exactly one
// deserializer arm and exactly one serializer arm when conversion is enabled;
// unset optional leaves are omitted so the AWS defaults win.
func buildDataFormatConversion(dfc *awskinesisfirehose.AwsKinesisFirehoseDataFormatConversion) *kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationArgs {
	args := &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationArgs{
		Enabled: pulumi.BoolPtr(true),
	}

	// Input format (deserializer)
	if hj := dfc.GetHiveJson(); hj != nil {
		hjArgs := &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationInputFormatConfigurationDeserializerHiveJsonSerDeArgs{}
		if len(hj.TimestampFormats) > 0 {
			hjArgs.TimestampFormats = pulumi.ToStringArray(hj.TimestampFormats)
		}
		args.InputFormatConfiguration = &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationInputFormatConfigurationArgs{
			Deserializer: &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationInputFormatConfigurationDeserializerArgs{
				HiveJsonSerDe: hjArgs,
			},
		}
	} else if oxj := dfc.GetOpenXJson(); oxj != nil {
		oxjArgs := &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationInputFormatConfigurationDeserializerOpenXJsonSerDeArgs{}
		// AWS defaults case_insensitive to true; only an explicit choice is
		// sent (the spec field carries presence).
		if oxj.CaseInsensitive != nil {
			oxjArgs.CaseInsensitive = pulumi.BoolPtr(oxj.GetCaseInsensitive())
		}
		if len(oxj.ColumnToJsonKeyMappings) > 0 {
			oxjArgs.ColumnToJsonKeyMappings = pulumi.ToStringMap(oxj.ColumnToJsonKeyMappings)
		}
		if oxj.ConvertDotsInJsonKeysToUnderscores {
			oxjArgs.ConvertDotsInJsonKeysToUnderscores = pulumi.BoolPtr(true)
		}
		args.InputFormatConfiguration = &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationInputFormatConfigurationArgs{
			Deserializer: &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationInputFormatConfigurationDeserializerArgs{
				OpenXJsonSerDe: oxjArgs,
			},
		}
	}

	// Output format (serializer)
	if orc := dfc.GetOrc(); orc != nil {
		orcArgs := &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationOutputFormatConfigurationSerializerOrcSerDeArgs{}
		if orc.Compression != "" {
			orcArgs.Compression = pulumi.StringPtr(orc.Compression)
		}
		if orc.BlockSizeBytes > 0 {
			orcArgs.BlockSizeBytes = pulumi.IntPtr(int(orc.BlockSizeBytes))
		}
		if orc.StripeSizeBytes > 0 {
			orcArgs.StripeSizeBytes = pulumi.IntPtr(int(orc.StripeSizeBytes))
		}
		if len(orc.BloomFilterColumns) > 0 {
			orcArgs.BloomFilterColumns = pulumi.ToStringArray(orc.BloomFilterColumns)
		}
		// Optional-with-presence in the spec: explicit 0 is AWS-legal here,
		// distinct from "unset -> AWS default".
		if orc.BloomFilterFalsePositiveProbability != nil {
			orcArgs.BloomFilterFalsePositiveProbability = pulumi.Float64Ptr(orc.GetBloomFilterFalsePositiveProbability())
		}
		if orc.DictionaryKeyThreshold > 0 {
			orcArgs.DictionaryKeyThreshold = pulumi.Float64Ptr(orc.DictionaryKeyThreshold)
		}
		if orc.EnablePadding {
			orcArgs.EnablePadding = pulumi.BoolPtr(true)
		}
		if orc.PaddingTolerance != nil {
			orcArgs.PaddingTolerance = pulumi.Float64Ptr(orc.GetPaddingTolerance())
		}
		if orc.FormatVersion != "" {
			orcArgs.FormatVersion = pulumi.StringPtr(orc.FormatVersion)
		}
		if orc.RowIndexStride > 0 {
			orcArgs.RowIndexStride = pulumi.IntPtr(int(orc.RowIndexStride))
		}
		args.OutputFormatConfiguration = &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationOutputFormatConfigurationArgs{
			Serializer: &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationOutputFormatConfigurationSerializerArgs{
				OrcSerDe: orcArgs,
			},
		}
	} else if pq := dfc.GetParquet(); pq != nil {
		parquetArgs := &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDataFormatConversionConfigurationOutputFormatConfigurationSerializerParquetSerDeArgs{}
		if pq.Compression != "" {
			parquetArgs.Compression = pulumi.StringPtr(pq.Compression)
		}
		if pq.BlockSizeBytes > 0 {
			parquetArgs.BlockSizeBytes = pulumi.IntPtr(int(pq.BlockSizeBytes))
		}
		if pq.PageSizeBytes > 0 {
			parquetArgs.PageSizeBytes = pulumi.IntPtr(int(pq.PageSizeBytes))
		}
		if pq.MaxPaddingBytes > 0 {
			parquetArgs.MaxPaddingBytes = pulumi.IntPtr(int(pq.MaxPaddingBytes))
		}
		if pq.EnableDictionaryCompression {
			parquetArgs.EnableDictionaryCompression = pulumi.BoolPtr(true)
		}
		if pq.WriterVersion != "" {
			parquetArgs.WriterVersion = pulumi.StringPtr(pq.WriterVersion)
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
