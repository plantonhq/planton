package awskinesisfirehosev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsKinesisFirehoseSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsKinesisFirehoseSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

// helper: minimal valid S3 config (backup/staging block shared by every
// non-S3 destination).
func minimalS3Config() *AwsKinesisFirehoseS3Config {
	return &AwsKinesisFirehoseS3Config{
		BucketArn: strRef("arn:aws:s3:::backup-bucket"),
		RoleArn:   strRef("arn:aws:iam::123456789012:role/firehose-s3"),
	}
}

// helper: minimal valid Extended S3 destination.
func minimalExtendedS3() *AwsKinesisFirehoseExtendedS3Destination {
	return &AwsKinesisFirehoseExtendedS3Destination{
		BucketArn: strRef("arn:aws:s3:::my-data-lake"),
		RoleArn:   strRef("arn:aws:iam::123456789012:role/firehose-s3"),
	}
}

// helper: minimal valid OpenSearch destination.
func minimalOpenSearch() *AwsKinesisFirehoseOpenSearchDestination {
	return &AwsKinesisFirehoseOpenSearchDestination{
		DomainArn: strRef("arn:aws:es:us-east-1:123456789012:domain/logs"),
		IndexName: "logs",
		RoleArn:   strRef("arn:aws:iam::123456789012:role/firehose-os"),
		S3Config:  minimalS3Config(),
	}
}

// helper: minimal valid OpenSearch Serverless destination.
func minimalOpenSearchServerless() *AwsKinesisFirehoseOpenSearchServerlessDestination {
	return &AwsKinesisFirehoseOpenSearchServerlessDestination{
		CollectionEndpoint: "https://abc123.us-east-1.aoss.amazonaws.com",
		IndexName:          "logs",
		RoleArn:            strRef("arn:aws:iam::123456789012:role/firehose-aoss"),
		S3Config:           minimalS3Config(),
	}
}

// helper: minimal valid HTTP endpoint destination.
func minimalHttpEndpoint() *AwsKinesisFirehoseHttpEndpointDestination {
	return &AwsKinesisFirehoseHttpEndpointDestination{
		Url:      "https://api.example.com/firehose",
		S3Config: minimalS3Config(),
	}
}

// helper: minimal valid Redshift destination (plaintext credentials).
func minimalRedshift() *AwsKinesisFirehoseRedshiftDestination {
	return &AwsKinesisFirehoseRedshiftDestination{
		ClusterJdbcurl: "jdbc:redshift://cluster.abc.us-east-1.redshift.amazonaws.com:5439/mydb",
		RoleArn:        strRef("arn:aws:iam::123456789012:role/firehose-rs"),
		DataTableName:  "events",
		Username:       "firehose_user",
		Password:       "SecretPass123!",
		S3Config:       minimalS3Config(),
	}
}

// helper: minimal valid Splunk destination (plaintext HEC token).
func minimalSplunk() *AwsKinesisFirehoseSplunkDestination {
	return &AwsKinesisFirehoseSplunkDestination{
		HecEndpoint: "https://http-inputs-mycompany.splunkcloud.com:443",
		HecToken:    "11111111-2222-3333-4444-555555555555",
		S3Config:    minimalS3Config(),
	}
}

// helper: minimal valid Snowflake destination (inline key-pair credentials).
func minimalSnowflake() *AwsKinesisFirehoseSnowflakeDestination {
	return &AwsKinesisFirehoseSnowflakeDestination{
		AccountUrl: "https://myaccount.snowflakecomputing.com",
		Database:   "ANALYTICS",
		Schema:     "PUBLIC",
		Table:      "EVENTS",
		RoleArn:    strRef("arn:aws:iam::123456789012:role/firehose-snowflake"),
		User:       "FIREHOSE_USER",
		PrivateKey: "MIIEvQIBADANBgkqhkiG9w0BAQEFAASC...",
		S3Config:   minimalS3Config(),
	}
}

// helper: minimal valid Iceberg destination.
func minimalIceberg() *AwsKinesisFirehoseIcebergDestination {
	return &AwsKinesisFirehoseIcebergDestination{
		CatalogArn: strRef("arn:aws:glue:us-east-1:123456789012:catalog"),
		RoleArn:    strRef("arn:aws:iam::123456789012:role/firehose-iceberg"),
		DestinationTables: []*AwsKinesisFirehoseIcebergDestinationTable{
			{DatabaseName: "lakehouse", TableName: "events"},
		},
		S3Config: minimalS3Config(),
	}
}

// helper: minimal valid secrets manager config.
func minimalSecretsManager() *AwsKinesisFirehoseSecretsManagerConfig {
	return &AwsKinesisFirehoseSecretsManagerConfig{
		SecretArn: strRef("arn:aws:secretsmanager:us-east-1:123456789012:secret:firehose-cred-AbCdEf"),
	}
}

var _ = ginkgo.Describe("AwsKinesisFirehoseSpec validations", func() {
	var spec *AwsKinesisFirehoseSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: Extended S3 destination with Direct PUT source.
		spec = &AwsKinesisFirehoseSpec{
			Region: "us-west-2",
			DestinationConfig: &AwsKinesisFirehoseSpec_ExtendedS3{
				ExtendedS3: minimalExtendedS3(),
			},
		}
	})

	// =========================================================================
	// Happy path: Extended S3 destination
	// =========================================================================

	ginkgo.It("accepts a minimal Extended S3 destination (Direct PUT)", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Extended S3 with compression and prefix", func() {
		s3 := minimalExtendedS3()
		s3.CompressionFormat = "GZIP"
		s3.Prefix = "data/year=!{timestamp:yyyy}/month=!{timestamp:MM}/"
		s3.ErrorOutputPrefix = "errors/"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Extended S3 with SSE enabled (Direct PUT)", func() {
		spec.SseEnabled = true
		spec.SseKmsKeyArn = strRef("arn:aws:kms:us-east-1:123456789012:key/abc")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Extended S3 with SSE using AWS-owned CMK", func() {
		spec.SseEnabled = true
		// No sse_kms_key_arn -> AWS_OWNED_CMK
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Extended S3 with Kinesis stream source", func() {
		spec.KinesisStreamSource = &AwsKinesisFirehoseKinesisStreamSource{
			StreamArn: strRef("arn:aws:kinesis:us-east-1:123456789012:stream/my-stream"),
			RoleArn:   strRef("arn:aws:iam::123456789012:role/firehose-kinesis"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Extended S3 with a Lambda processor", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{Lambda: &AwsKinesisFirehoseLambdaProcessor{
					LambdaArn: strRef("arn:aws:lambda:us-east-1:123456789012:function:transform"),
				}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Extended S3 with data format conversion to Parquet", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled:      true,
			Deserializer: &AwsKinesisFirehoseDataFormatConversion_OpenXJson{OpenXJson: &AwsKinesisFirehoseOpenXJsonDeserializer{}},
			Serializer:   &AwsKinesisFirehoseDataFormatConversion_Parquet{Parquet: &AwsKinesisFirehoseParquetSerializer{}},
			Schema: &AwsKinesisFirehoseGlueSchemaConfig{
				DatabaseName: "analytics",
				TableName:    "events",
				RoleArn:      strRef("arn:aws:iam::123456789012:role/firehose-glue"),
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts fully tuned Parquet and Hive JSON serde arms", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled: true,
			Deserializer: &AwsKinesisFirehoseDataFormatConversion_HiveJson{HiveJson: &AwsKinesisFirehoseHiveJsonDeserializer{
				TimestampFormats: []string{"yyyy-MM-dd'T'HH:mm:ss", "millis"},
			}},
			Serializer: &AwsKinesisFirehoseDataFormatConversion_Parquet{Parquet: &AwsKinesisFirehoseParquetSerializer{
				Compression:                 "GZIP",
				BlockSizeBytes:              134217728,
				PageSizeBytes:               2097152,
				MaxPaddingBytes:             1048576,
				EnableDictionaryCompression: true,
				WriterVersion:               "V2",
			}},
			Schema: &AwsKinesisFirehoseGlueSchemaConfig{
				DatabaseName: "analytics",
				TableName:    "events",
				RoleArn:      strRef("arn:aws:iam::123456789012:role/firehose-glue"),
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a fully tuned ORC serializer with OpenX key mappings", func() {
		s3 := minimalExtendedS3()
		caseSensitive := false
		fpp := 0.01
		tolerance := 0.0
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled: true,
			Deserializer: &AwsKinesisFirehoseDataFormatConversion_OpenXJson{OpenXJson: &AwsKinesisFirehoseOpenXJsonDeserializer{
				CaseInsensitive:                    &caseSensitive,
				ColumnToJsonKeyMappings:            map[string]string{"ts": "timestamp"},
				ConvertDotsInJsonKeysToUnderscores: true,
			}},
			Serializer: &AwsKinesisFirehoseDataFormatConversion_Orc{Orc: &AwsKinesisFirehoseOrcSerializer{
				Compression:                         "ZLIB",
				BlockSizeBytes:                      134217728,
				StripeSizeBytes:                     16777216,
				BloomFilterColumns:                  []string{"customer_id", "event_type"},
				BloomFilterFalsePositiveProbability: &fpp,
				DictionaryKeyThreshold:              0.8,
				EnablePadding:                       true,
				PaddingTolerance:                    &tolerance,
				FormatVersion:                       "V0_12",
				RowIndexStride:                      5000,
			}},
			Schema: &AwsKinesisFirehoseGlueSchemaConfig{
				DatabaseName: "analytics",
				TableName:    "events",
				RoleArn:      strRef("arn:aws:iam::123456789012:role/firehose-glue"),
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Extended S3 with dynamic partitioning driven by metadata extraction", func() {
		s3 := minimalExtendedS3()
		s3.Prefix = "data/customer=!{partitionKeyFromQuery:customer_id}/"
		s3.DynamicPartitioning = &AwsKinesisFirehoseDynamicPartitioning{
			Enabled: true,
		}
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{MetadataExtraction: &AwsKinesisFirehoseMetadataExtractionProcessor{
					Query:             "{customer_id: .customer_id}",
					JsonParsingEngine: "JQ-1.6",
				}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a multi-processor pipeline in execution order", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{Decompression: &AwsKinesisFirehoseDecompressionProcessor{
					CompressionFormat: "GZIP",
				}},
				{CloudwatchLogProcessing: &AwsKinesisFirehoseCloudwatchLogProcessingProcessor{
					DataMessageExtraction: true,
				}},
				{AppendDelimiter: &AwsKinesisFirehoseAppendDelimiterProcessor{
					Delimiter: "\\n",
				}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a record deaggregation processor with DELIMITED sub-records", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{RecordDeaggregation: &AwsKinesisFirehoseRecordDeaggregationProcessor{
					SubRecordType: "DELIMITED",
					Delimiter:     "Cg==",
				}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Extended S3 with S3 backup enabled", func() {
		s3 := minimalExtendedS3()
		s3.S3BackupMode = "Enabled"
		s3.S3Backup = minimalS3Config()
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Extended S3 with buffering hints", func() {
		s3 := minimalExtendedS3()
		s3.Buffering = &AwsKinesisFirehoseBufferingHints{
			IntervalInSeconds: 60,
			SizeInMbs:         128,
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path: MSK source
	// =========================================================================

	ginkgo.It("accepts Extended S3 with an MSK source", func() {
		spec.MskSource = &AwsKinesisFirehoseMskSource{
			MskClusterArn: strRef("arn:aws:kafka:us-east-1:123456789012:cluster/my-cluster/abc"),
			TopicName:     "events",
			Connectivity:  "PRIVATE",
			RoleArn:       strRef("arn:aws:iam::123456789012:role/firehose-msk"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an MSK source with read_from_timestamp", func() {
		spec.MskSource = &AwsKinesisFirehoseMskSource{
			MskClusterArn:     strRef("arn:aws:kafka:us-east-1:123456789012:cluster/my-cluster/abc"),
			TopicName:         "events",
			Connectivity:      "PUBLIC",
			RoleArn:           strRef("arn:aws:iam::123456789012:role/firehose-msk"),
			ReadFromTimestamp: "2026-05-01T00:00:00Z",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path: OpenSearch destination
	// =========================================================================

	ginkgo.It("accepts a minimal OpenSearch destination", func() {
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Opensearch{
			Opensearch: minimalOpenSearch(),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts OpenSearch with cluster_endpoint instead of domain_arn", func() {
		os := minimalOpenSearch()
		os.DomainArn = nil
		os.ClusterEndpoint = "https://search-domain-xxxx.us-east-1.es.amazonaws.com"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Opensearch{Opensearch: os}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts OpenSearch with VPC config", func() {
		os := minimalOpenSearch()
		os.VpcConfig = &AwsKinesisFirehoseVpcConfig{
			SubnetIds:        []*foreignkeyv1.StringValueOrRef{strRef("subnet-abc123")},
			SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{strRef("sg-abc123")},
			RoleArn:          strRef("arn:aws:iam::123456789012:role/firehose-vpc"),
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Opensearch{Opensearch: os}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts OpenSearch with index rotation, backup mode, and document ID format", func() {
		os := minimalOpenSearch()
		os.IndexRotationPeriod = "OneHour"
		os.S3BackupMode = "AllDocuments"
		os.DefaultDocumentIdFormat = "NO_DOCUMENT_ID"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Opensearch{Opensearch: os}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path: OpenSearch Serverless destination
	// =========================================================================

	ginkgo.It("accepts a minimal OpenSearch Serverless destination", func() {
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_OpensearchServerless{
			OpensearchServerless: minimalOpenSearchServerless(),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts OpenSearch Serverless with VPC config and backup mode", func() {
		aoss := minimalOpenSearchServerless()
		aoss.S3BackupMode = "AllDocuments"
		aoss.VpcConfig = &AwsKinesisFirehoseVpcConfig{
			SubnetIds:        []*foreignkeyv1.StringValueOrRef{strRef("subnet-abc123")},
			SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{strRef("sg-abc123")},
			RoleArn:          strRef("arn:aws:iam::123456789012:role/firehose-vpc"),
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_OpensearchServerless{OpensearchServerless: aoss}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path: HTTP endpoint destination
	// =========================================================================

	ginkgo.It("accepts a minimal HTTP endpoint destination", func() {
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_HttpEndpoint{
			HttpEndpoint: minimalHttpEndpoint(),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts HTTP endpoint with access key and request config", func() {
		http := minimalHttpEndpoint()
		http.Name = "Datadog"
		http.AccessKey = "my-secret-key"
		http.RequestConfig = &AwsKinesisFirehoseRequestConfig{
			ContentEncoding: "GZIP",
			CommonAttributes: []*AwsKinesisFirehoseRequestAttribute{
				{Name: "env", Value: "production"},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_HttpEndpoint{HttpEndpoint: http}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts HTTP endpoint with Secrets Manager credentials", func() {
		http := minimalHttpEndpoint()
		http.SecretsManager = minimalSecretsManager()
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_HttpEndpoint{HttpEndpoint: http}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path: Redshift destination
	// =========================================================================

	ginkgo.It("accepts a minimal Redshift destination with plaintext credentials", func() {
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Redshift{
			Redshift: minimalRedshift(),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Redshift with COPY options", func() {
		rs := minimalRedshift()
		rs.CopyOptions = "JSON 'auto'"
		rs.DataTableColumns = "id,event_type,created_at"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Redshift{Redshift: rs}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Redshift with Secrets Manager credentials", func() {
		rs := minimalRedshift()
		rs.Username = ""
		rs.Password = ""
		rs.SecretsManager = minimalSecretsManager()
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Redshift{Redshift: rs}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path: Splunk destination
	// =========================================================================

	ginkgo.It("accepts a minimal Splunk destination with a HEC token", func() {
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Splunk{
			Splunk: minimalSplunk(),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Splunk with Secrets Manager credentials", func() {
		sp := minimalSplunk()
		sp.HecToken = ""
		sp.SecretsManager = minimalSecretsManager()
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Splunk{Splunk: sp}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Splunk with endpoint type, ack timeout, and tight buffering", func() {
		sp := minimalSplunk()
		sp.HecEndpointType = "Event"
		sp.HecAcknowledgmentTimeoutInSeconds = 300
		sp.S3BackupMode = "AllEvents"
		sp.Buffering = &AwsKinesisFirehoseBufferingHints{
			IntervalInSeconds: 60,
			SizeInMbs:         5,
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Splunk{Splunk: sp}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path: Snowflake destination
	// =========================================================================

	ginkgo.It("accepts a minimal Snowflake destination with key-pair credentials", func() {
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Snowflake{
			Snowflake: minimalSnowflake(),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Snowflake with Secrets Manager credentials", func() {
		sf := minimalSnowflake()
		sf.User = ""
		sf.PrivateKey = ""
		sf.SecretsManager = minimalSecretsManager()
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Snowflake{Snowflake: sf}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Snowflake with VARIANT content mapping and a Snowflake role", func() {
		sf := minimalSnowflake()
		sf.DataLoadingOption = "VARIANT_CONTENT_AND_METADATA_MAPPING"
		sf.ContentColumnName = "RECORD_CONTENT"
		sf.MetadataColumnName = "RECORD_METADATA"
		sf.SnowflakeRole = "FIREHOSE_INGEST"
		sf.PrivateLinkVpceId = "com.amazonaws.vpce.us-east-1.vpce-svc-0123456789abcdef0"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Snowflake{Snowflake: sf}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Snowflake with an encrypted private key and passphrase", func() {
		sf := minimalSnowflake()
		sf.KeyPassphrase = "my-passphrase"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Snowflake{Snowflake: sf}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path: Iceberg destination
	// =========================================================================

	ginkgo.It("accepts a minimal Iceberg destination", func() {
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Iceberg{
			Iceberg: minimalIceberg(),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Iceberg with multiple tables, unique keys, and append-only off", func() {
		ib := minimalIceberg()
		ib.DestinationTables = []*AwsKinesisFirehoseIcebergDestinationTable{
			{DatabaseName: "lakehouse", TableName: "orders", UniqueKeys: []string{"order_id"}},
			{DatabaseName: "lakehouse", TableName: "customers", UniqueKeys: []string{"customer_id"}, S3ErrorOutputPrefix: "errors/customers/"},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Iceberg{Iceberg: ib}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Iceberg in append-only mode", func() {
		ib := minimalIceberg()
		ib.AppendOnly = true
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Iceberg{Iceberg: ib}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("rejects Iceberg with no destination tables", func() {
		// A table-less Iceberg destination cannot route records -- AWS
		// rejects it at apply time, so min_items catches it here.
		ib := minimalIceberg()
		ib.DestinationTables = nil
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Iceberg{Iceberg: ib}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).ToNot(gomega.BeNil())
	})

	// =========================================================================
	// Happy path: production-ready Extended S3 (full config)
	// =========================================================================

	ginkgo.It("accepts a production-ready Extended S3 with all features", func() {
		spec.KinesisStreamSource = &AwsKinesisFirehoseKinesisStreamSource{
			StreamArn: strRef("arn:aws:kinesis:us-east-1:123456789012:stream/events"),
			RoleArn:   strRef("arn:aws:iam::123456789012:role/firehose-kinesis-read"),
		}
		s3 := &AwsKinesisFirehoseExtendedS3Destination{
			BucketArn:         strRef("arn:aws:s3:::prod-data-lake"),
			RoleArn:           strRef("arn:aws:iam::123456789012:role/firehose-prod"),
			Prefix:            "data/customer=!{partitionKeyFromQuery:customer_id}/day=!{timestamp:dd}/",
			ErrorOutputPrefix: "errors/year=!{timestamp:yyyy}/",
			KmsKeyArn:         strRef("arn:aws:kms:us-east-1:123456789012:key/prod-key"),
			FileExtension:     ".parquet",
			Buffering: &AwsKinesisFirehoseBufferingHints{
				IntervalInSeconds: 120,
				SizeInMbs:         64,
			},
			Processing: &AwsKinesisFirehoseProcessing{
				Enabled: true,
				Processors: []*AwsKinesisFirehoseProcessor{
					{MetadataExtraction: &AwsKinesisFirehoseMetadataExtractionProcessor{
						Query: "{customer_id: .customer_id}",
					}},
					{Lambda: &AwsKinesisFirehoseLambdaProcessor{
						LambdaArn:               strRef("arn:aws:lambda:us-east-1:123456789012:function:enrich"),
						BufferSizeInMbs:         3,
						BufferIntervalInSeconds: 60,
						NumberOfRetries:         3,
					}},
				},
			},
			DynamicPartitioning: &AwsKinesisFirehoseDynamicPartitioning{
				Enabled:                true,
				RetryDurationInSeconds: 600,
			},
			DataFormatConversion: &AwsKinesisFirehoseDataFormatConversion{
				Enabled:      true,
				Deserializer: &AwsKinesisFirehoseDataFormatConversion_OpenXJson{OpenXJson: &AwsKinesisFirehoseOpenXJsonDeserializer{}},
				Serializer: &AwsKinesisFirehoseDataFormatConversion_Parquet{Parquet: &AwsKinesisFirehoseParquetSerializer{
					Compression: "SNAPPY",
				}},
				Schema: &AwsKinesisFirehoseGlueSchemaConfig{
					DatabaseName: "analytics",
					TableName:    "events_v2",
					RoleArn:      strRef("arn:aws:iam::123456789012:role/firehose-glue"),
				},
			},
			Logging: &AwsKinesisFirehoseCloudwatchLogging{
				Enabled:       true,
				LogGroupName:  "/aws/kinesisfirehose/prod-stream",
				LogStreamName: "S3Delivery",
			},
			S3BackupMode: "Enabled",
			S3Backup: &AwsKinesisFirehoseS3Config{
				BucketArn:         strRef("arn:aws:s3:::prod-backup"),
				RoleArn:           strRef("arn:aws:iam::123456789012:role/firehose-backup"),
				CompressionFormat: "GZIP",
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Failure: missing destination (oneof required)
	// =========================================================================

	ginkgo.It("fails when no destination is configured", func() {
		spec.DestinationConfig = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure: source exclusivity and SSE conflicts
	// =========================================================================

	ginkgo.It("fails when both Kinesis and MSK sources are configured", func() {
		spec.KinesisStreamSource = &AwsKinesisFirehoseKinesisStreamSource{
			StreamArn: strRef("arn:aws:kinesis:us-east-1:123456789012:stream/my-stream"),
			RoleArn:   strRef("arn:aws:iam::123456789012:role/firehose-kinesis"),
		}
		spec.MskSource = &AwsKinesisFirehoseMskSource{
			MskClusterArn: strRef("arn:aws:kafka:us-east-1:123456789012:cluster/my-cluster/abc"),
			TopicName:     "events",
			Connectivity:  "PRIVATE",
			RoleArn:       strRef("arn:aws:iam::123456789012:role/firehose-msk"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when SSE is enabled with Kinesis stream source", func() {
		spec.SseEnabled = true
		spec.KinesisStreamSource = &AwsKinesisFirehoseKinesisStreamSource{
			StreamArn: strRef("arn:aws:kinesis:us-east-1:123456789012:stream/my-stream"),
			RoleArn:   strRef("arn:aws:iam::123456789012:role/firehose-kinesis"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when SSE is enabled with MSK source", func() {
		spec.SseEnabled = true
		spec.MskSource = &AwsKinesisFirehoseMskSource{
			MskClusterArn: strRef("arn:aws:kafka:us-east-1:123456789012:cluster/my-cluster/abc"),
			TopicName:     "events",
			Connectivity:  "PRIVATE",
			RoleArn:       strRef("arn:aws:iam::123456789012:role/firehose-msk"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when sse_kms_key_arn is set without sse_enabled", func() {
		spec.SseKmsKeyArn = strRef("arn:aws:kms:us-east-1:123456789012:key/abc")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure: Kinesis source missing required fields
	// =========================================================================

	ginkgo.It("fails when Kinesis source is missing stream_arn", func() {
		spec.KinesisStreamSource = &AwsKinesisFirehoseKinesisStreamSource{
			RoleArn: strRef("arn:aws:iam::123456789012:role/firehose-kinesis"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Kinesis source is missing role_arn", func() {
		spec.KinesisStreamSource = &AwsKinesisFirehoseKinesisStreamSource{
			StreamArn: strRef("arn:aws:kinesis:us-east-1:123456789012:stream/my-stream"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure: MSK source validations
	// =========================================================================

	ginkgo.It("fails when MSK source is missing msk_cluster_arn", func() {
		spec.MskSource = &AwsKinesisFirehoseMskSource{
			TopicName:    "events",
			Connectivity: "PRIVATE",
			RoleArn:      strRef("arn:aws:iam::123456789012:role/firehose-msk"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when MSK source is missing topic_name", func() {
		spec.MskSource = &AwsKinesisFirehoseMskSource{
			MskClusterArn: strRef("arn:aws:kafka:us-east-1:123456789012:cluster/my-cluster/abc"),
			Connectivity:  "PRIVATE",
			RoleArn:       strRef("arn:aws:iam::123456789012:role/firehose-msk"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when MSK source has invalid connectivity", func() {
		spec.MskSource = &AwsKinesisFirehoseMskSource{
			MskClusterArn: strRef("arn:aws:kafka:us-east-1:123456789012:cluster/my-cluster/abc"),
			TopicName:     "events",
			Connectivity:  "VPC",
			RoleArn:       strRef("arn:aws:iam::123456789012:role/firehose-msk"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when MSK read_from_timestamp is not RFC 3339", func() {
		spec.MskSource = &AwsKinesisFirehoseMskSource{
			MskClusterArn:     strRef("arn:aws:kafka:us-east-1:123456789012:cluster/my-cluster/abc"),
			TopicName:         "events",
			Connectivity:      "PRIVATE",
			RoleArn:           strRef("arn:aws:iam::123456789012:role/firehose-msk"),
			ReadFromTimestamp: "2026-05-01 00:00:00",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure: Extended S3 validations
	// =========================================================================

	ginkgo.It("fails when Extended S3 has invalid compression_format", func() {
		s3 := minimalExtendedS3()
		s3.CompressionFormat = "BROTLI"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Extended S3 has invalid s3_backup_mode", func() {
		s3 := minimalExtendedS3()
		s3.S3BackupMode = "SomeInvalidMode"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Extended S3 s3_backup is set without s3_backup_mode Enabled", func() {
		s3 := minimalExtendedS3()
		s3.S3Backup = minimalS3Config()
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Extended S3 is missing bucket_arn", func() {
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{
			ExtendedS3: &AwsKinesisFirehoseExtendedS3Destination{
				RoleArn: strRef("arn:aws:iam::123456789012:role/firehose"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure: Data format conversion validations
	// =========================================================================

	ginkgo.It("fails when data format conversion is enabled without a serializer arm", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled:      true,
			Deserializer: &AwsKinesisFirehoseDataFormatConversion_OpenXJson{OpenXJson: &AwsKinesisFirehoseOpenXJsonDeserializer{}},
			Schema: &AwsKinesisFirehoseGlueSchemaConfig{
				DatabaseName: "db",
				TableName:    "tbl",
				RoleArn:      strRef("arn:aws:iam::123456789012:role/glue"),
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when data format conversion is enabled without a deserializer arm", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled:    true,
			Serializer: &AwsKinesisFirehoseDataFormatConversion_Parquet{Parquet: &AwsKinesisFirehoseParquetSerializer{}},
			Schema: &AwsKinesisFirehoseGlueSchemaConfig{
				DatabaseName: "db",
				TableName:    "tbl",
				RoleArn:      strRef("arn:aws:iam::123456789012:role/glue"),
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when data format conversion is enabled without schema", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled:      true,
			Deserializer: &AwsKinesisFirehoseDataFormatConversion_OpenXJson{OpenXJson: &AwsKinesisFirehoseOpenXJsonDeserializer{}},
			Serializer:   &AwsKinesisFirehoseDataFormatConversion_Parquet{Parquet: &AwsKinesisFirehoseParquetSerializer{}},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("carries one deserializer and one serializer by construction (a second arm replaces the first)", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled:      true,
			Deserializer: &AwsKinesisFirehoseDataFormatConversion_OpenXJson{OpenXJson: &AwsKinesisFirehoseOpenXJsonDeserializer{}},
			Serializer:   &AwsKinesisFirehoseDataFormatConversion_Parquet{Parquet: &AwsKinesisFirehoseParquetSerializer{}},
			Schema: &AwsKinesisFirehoseGlueSchemaConfig{
				DatabaseName: "db",
				TableName:    "tbl",
				RoleArn:      strRef("arn:aws:iam::123456789012:role/glue"),
			},
		}
		s3.DataFormatConversion.Deserializer = &AwsKinesisFirehoseDataFormatConversion_HiveJson{HiveJson: &AwsKinesisFirehoseHiveJsonDeserializer{}}
		gomega.Expect(s3.DataFormatConversion.GetOpenXJson()).To(gomega.BeNil())
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("fails when parquet compression is not a legal codec", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled:      true,
			Deserializer: &AwsKinesisFirehoseDataFormatConversion_OpenXJson{OpenXJson: &AwsKinesisFirehoseOpenXJsonDeserializer{}},
			Serializer:   &AwsKinesisFirehoseDataFormatConversion_Parquet{Parquet: &AwsKinesisFirehoseParquetSerializer{Compression: "ZLIB"}},
			Schema: &AwsKinesisFirehoseGlueSchemaConfig{
				DatabaseName: "db",
				TableName:    "tbl",
				RoleArn:      strRef("arn:aws:iam::123456789012:role/glue"),
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when parquet page size is below the 64 KiB floor", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled:      true,
			Deserializer: &AwsKinesisFirehoseDataFormatConversion_OpenXJson{OpenXJson: &AwsKinesisFirehoseOpenXJsonDeserializer{}},
			Serializer:   &AwsKinesisFirehoseDataFormatConversion_Parquet{Parquet: &AwsKinesisFirehoseParquetSerializer{PageSizeBytes: 4096}},
			Schema: &AwsKinesisFirehoseGlueSchemaConfig{
				DatabaseName: "db",
				TableName:    "tbl",
				RoleArn:      strRef("arn:aws:iam::123456789012:role/glue"),
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when ORC stripe size is below the 8 MiB floor", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled:      true,
			Deserializer: &AwsKinesisFirehoseDataFormatConversion_OpenXJson{OpenXJson: &AwsKinesisFirehoseOpenXJsonDeserializer{}},
			Serializer:   &AwsKinesisFirehoseDataFormatConversion_Orc{Orc: &AwsKinesisFirehoseOrcSerializer{StripeSizeBytes: 1048576}},
			Schema: &AwsKinesisFirehoseGlueSchemaConfig{
				DatabaseName: "db",
				TableName:    "tbl",
				RoleArn:      strRef("arn:aws:iam::123456789012:role/glue"),
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when ORC bloom filter probability is above 1", func() {
		s3 := minimalExtendedS3()
		fpp := 1.5
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled:      true,
			Deserializer: &AwsKinesisFirehoseDataFormatConversion_OpenXJson{OpenXJson: &AwsKinesisFirehoseOpenXJsonDeserializer{}},
			Serializer:   &AwsKinesisFirehoseDataFormatConversion_Orc{Orc: &AwsKinesisFirehoseOrcSerializer{BloomFilterFalsePositiveProbability: &fpp}},
			Schema: &AwsKinesisFirehoseGlueSchemaConfig{
				DatabaseName: "db",
				TableName:    "tbl",
				RoleArn:      strRef("arn:aws:iam::123456789012:role/glue"),
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when ORC row index stride is below 1000", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled:      true,
			Deserializer: &AwsKinesisFirehoseDataFormatConversion_OpenXJson{OpenXJson: &AwsKinesisFirehoseOpenXJsonDeserializer{}},
			Serializer:   &AwsKinesisFirehoseDataFormatConversion_Orc{Orc: &AwsKinesisFirehoseOrcSerializer{RowIndexStride: 500}},
			Schema: &AwsKinesisFirehoseGlueSchemaConfig{
				DatabaseName: "db",
				TableName:    "tbl",
				RoleArn:      strRef("arn:aws:iam::123456789012:role/glue"),
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when ORC format version is not V0_11 or V0_12", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled:      true,
			Deserializer: &AwsKinesisFirehoseDataFormatConversion_OpenXJson{OpenXJson: &AwsKinesisFirehoseOpenXJsonDeserializer{}},
			Serializer:   &AwsKinesisFirehoseDataFormatConversion_Orc{Orc: &AwsKinesisFirehoseOrcSerializer{FormatVersion: "V1"}},
			Schema: &AwsKinesisFirehoseGlueSchemaConfig{
				DatabaseName: "db",
				TableName:    "tbl",
				RoleArn:      strRef("arn:aws:iam::123456789012:role/glue"),
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure: processing pipeline validations
	// =========================================================================

	ginkgo.It("fails when processing is enabled without processors", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when processors are configured without enabled", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Processors: []*AwsKinesisFirehoseProcessor{
				{AppendDelimiter: &AwsKinesisFirehoseAppendDelimiterProcessor{Delimiter: "\\n"}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a processor entry sets no typed arm", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled:    true,
			Processors: []*AwsKinesisFirehoseProcessor{{}},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a processor entry sets two typed arms", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{
					Lambda: &AwsKinesisFirehoseLambdaProcessor{
						LambdaArn: strRef("arn:aws:lambda:us-east-1:123456789012:function:fn"),
					},
					AppendDelimiter: &AwsKinesisFirehoseAppendDelimiterProcessor{Delimiter: "\\n"},
				},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a Lambda processor is missing lambda_arn", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{Lambda: &AwsKinesisFirehoseLambdaProcessor{}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a Lambda processor buffer_size_in_mbs exceeds 3", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{Lambda: &AwsKinesisFirehoseLambdaProcessor{
					LambdaArn:       strRef("arn:aws:lambda:us-east-1:123456789012:function:fn"),
					BufferSizeInMbs: 5,
				}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a Lambda processor with a fractional sub-1-MiB buffer size", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{Lambda: &AwsKinesisFirehoseLambdaProcessor{
					LambdaArn:       strRef("arn:aws:lambda:us-east-1:123456789012:function:fn"),
					BufferSizeInMbs: 0.5,
				}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when a Lambda processor buffer_size_in_mbs is below 0.2", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{Lambda: &AwsKinesisFirehoseLambdaProcessor{
					LambdaArn:       strRef("arn:aws:lambda:us-east-1:123456789012:function:fn"),
					BufferSizeInMbs: 0.1,
				}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a Lambda processor with a dedicated invocation role", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{Lambda: &AwsKinesisFirehoseLambdaProcessor{
					LambdaArn: strRef("arn:aws:lambda:us-east-1:123456789012:function:fn"),
					RoleArn:   strRef("arn:aws:iam::123456789012:role/firehose-lambda-invoke"),
				}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when a metadata extraction processor has an invalid parsing engine", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{MetadataExtraction: &AwsKinesisFirehoseMetadataExtractionProcessor{
					Query:             "{customer_id: .customer_id}",
					JsonParsingEngine: "JQ-2.0",
				}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a decompression processor has an invalid format", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{Decompression: &AwsKinesisFirehoseDecompressionProcessor{
					CompressionFormat: "ZSTD",
				}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when record deaggregation is DELIMITED without a delimiter", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{RecordDeaggregation: &AwsKinesisFirehoseRecordDeaggregationProcessor{
					SubRecordType: "DELIMITED",
				}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when record deaggregation has an invalid sub_record_type", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{RecordDeaggregation: &AwsKinesisFirehoseRecordDeaggregationProcessor{
					SubRecordType: "PROTOBUF",
				}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a non-S3 destination carries a record deaggregation processor", func() {
		sp := minimalSplunk()
		sp.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{RecordDeaggregation: &AwsKinesisFirehoseRecordDeaggregationProcessor{
					SubRecordType: "JSON",
				}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Splunk{Splunk: sp}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a non-S3 destination carries an append delimiter processor", func() {
		os := minimalOpenSearch()
		os.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{AppendDelimiter: &AwsKinesisFirehoseAppendDelimiterProcessor{Delimiter: "\\n"}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Opensearch{Opensearch: os}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a CloudWatch-Logs pipeline on a non-S3 destination", func() {
		sp := minimalSplunk()
		sp.Processing = &AwsKinesisFirehoseProcessing{
			Enabled: true,
			Processors: []*AwsKinesisFirehoseProcessor{
				{Decompression: &AwsKinesisFirehoseDecompressionProcessor{CompressionFormat: "GZIP"}},
				{CloudwatchLogProcessing: &AwsKinesisFirehoseCloudwatchLogProcessingProcessor{DataMessageExtraction: true}},
			},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Splunk{Splunk: sp}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Failure: OpenSearch validations
	// =========================================================================

	ginkgo.It("fails when OpenSearch has both domain_arn and cluster_endpoint", func() {
		os := minimalOpenSearch()
		os.ClusterEndpoint = "https://search-domain.us-east-1.es.amazonaws.com"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Opensearch{Opensearch: os}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when OpenSearch has neither domain_arn nor cluster_endpoint", func() {
		os := minimalOpenSearch()
		os.DomainArn = nil
		os.ClusterEndpoint = ""
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Opensearch{Opensearch: os}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when OpenSearch has invalid index_rotation_period", func() {
		os := minimalOpenSearch()
		os.IndexRotationPeriod = "TwoHours"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Opensearch{Opensearch: os}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when OpenSearch has invalid default_document_id_format", func() {
		os := minimalOpenSearch()
		os.DefaultDocumentIdFormat = "CUSTOM_ID"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Opensearch{Opensearch: os}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when OpenSearch buffering size exceeds the 100 MiB cap", func() {
		os := minimalOpenSearch()
		os.Buffering = &AwsKinesisFirehoseBufferingHints{SizeInMbs: 128}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Opensearch{Opensearch: os}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when OpenSearch has invalid s3_backup_mode", func() {
		os := minimalOpenSearch()
		os.S3BackupMode = "InvalidMode"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Opensearch{Opensearch: os}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when OpenSearch is missing s3_config", func() {
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Opensearch{
			Opensearch: &AwsKinesisFirehoseOpenSearchDestination{
				DomainArn: strRef("arn:aws:es:us-east-1:123456789012:domain/logs"),
				IndexName: "logs",
				RoleArn:   strRef("arn:aws:iam::123456789012:role/firehose-os"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure: OpenSearch Serverless validations
	// =========================================================================

	ginkgo.It("fails when OpenSearch Serverless collection_endpoint is not HTTPS", func() {
		aoss := minimalOpenSearchServerless()
		aoss.CollectionEndpoint = "http://abc123.us-east-1.aoss.amazonaws.com"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_OpensearchServerless{OpensearchServerless: aoss}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when OpenSearch Serverless is missing index_name", func() {
		aoss := minimalOpenSearchServerless()
		aoss.IndexName = ""
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_OpensearchServerless{OpensearchServerless: aoss}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when OpenSearch Serverless is missing s3_config", func() {
		aoss := minimalOpenSearchServerless()
		aoss.S3Config = nil
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_OpensearchServerless{OpensearchServerless: aoss}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when OpenSearch Serverless buffering size exceeds the 100 MiB cap", func() {
		aoss := minimalOpenSearchServerless()
		aoss.Buffering = &AwsKinesisFirehoseBufferingHints{SizeInMbs: 128}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_OpensearchServerless{OpensearchServerless: aoss}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure: HTTP endpoint validations
	// =========================================================================

	ginkgo.It("fails when HTTP endpoint URL is not HTTPS", func() {
		http := minimalHttpEndpoint()
		http.Url = "http://api.example.com/firehose"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_HttpEndpoint{HttpEndpoint: http}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when HTTP endpoint sets both access_key and secrets_manager", func() {
		http := minimalHttpEndpoint()
		http.AccessKey = "my-secret-key"
		http.SecretsManager = minimalSecretsManager()
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_HttpEndpoint{HttpEndpoint: http}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when HTTP endpoint buffering size exceeds the 100 MiB cap", func() {
		http := minimalHttpEndpoint()
		http.Buffering = &AwsKinesisFirehoseBufferingHints{SizeInMbs: 128}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_HttpEndpoint{HttpEndpoint: http}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when HTTP endpoint has invalid s3_backup_mode", func() {
		http := minimalHttpEndpoint()
		http.S3BackupMode = "InvalidMode"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_HttpEndpoint{HttpEndpoint: http}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when HTTP endpoint has invalid content_encoding", func() {
		http := minimalHttpEndpoint()
		http.RequestConfig = &AwsKinesisFirehoseRequestConfig{
			ContentEncoding: "BROTLI",
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_HttpEndpoint{HttpEndpoint: http}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure: Redshift validations
	// =========================================================================

	ginkgo.It("fails when Redshift is missing cluster_jdbcurl", func() {
		rs := minimalRedshift()
		rs.ClusterJdbcurl = ""
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Redshift{Redshift: rs}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Redshift has no credentials at all", func() {
		rs := minimalRedshift()
		rs.Username = ""
		rs.Password = ""
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Redshift{Redshift: rs}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Redshift sets both plaintext credentials and secrets_manager", func() {
		rs := minimalRedshift()
		rs.SecretsManager = minimalSecretsManager()
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Redshift{Redshift: rs}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Redshift sets username without password", func() {
		rs := minimalRedshift()
		rs.Password = ""
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Redshift{Redshift: rs}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Redshift s3_backup is set without s3_backup_mode Enabled", func() {
		rs := minimalRedshift()
		rs.S3Backup = minimalS3Config()
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Redshift{Redshift: rs}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Redshift retry_duration exceeds 7200", func() {
		rs := minimalRedshift()
		rs.RetryDurationInSeconds = 8000
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Redshift{Redshift: rs}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure: Splunk validations
	// =========================================================================

	ginkgo.It("fails when Splunk hec_endpoint is not HTTPS", func() {
		sp := minimalSplunk()
		sp.HecEndpoint = "http://splunk.example.com:8088"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Splunk{Splunk: sp}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Splunk has invalid hec_endpoint_type", func() {
		sp := minimalSplunk()
		sp.HecEndpointType = "Stream"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Splunk{Splunk: sp}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Splunk has no credentials at all", func() {
		sp := minimalSplunk()
		sp.HecToken = ""
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Splunk{Splunk: sp}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Splunk sets both hec_token and secrets_manager", func() {
		sp := minimalSplunk()
		sp.SecretsManager = minimalSecretsManager()
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Splunk{Splunk: sp}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Splunk ack timeout is below 180", func() {
		sp := minimalSplunk()
		sp.HecAcknowledgmentTimeoutInSeconds = 60
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Splunk{Splunk: sp}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Splunk buffering exceeds the 60s / 5 MiB caps", func() {
		sp := minimalSplunk()
		sp.Buffering = &AwsKinesisFirehoseBufferingHints{
			IntervalInSeconds: 300,
			SizeInMbs:         5,
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Splunk{Splunk: sp}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Splunk has invalid s3_backup_mode", func() {
		sp := minimalSplunk()
		sp.S3BackupMode = "FailedDataOnly"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Splunk{Splunk: sp}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure: Snowflake validations
	// =========================================================================

	ginkgo.It("fails when Snowflake account_url is not HTTPS", func() {
		sf := minimalSnowflake()
		sf.AccountUrl = "myaccount.snowflakecomputing.com"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Snowflake{Snowflake: sf}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Snowflake has no credentials at all", func() {
		sf := minimalSnowflake()
		sf.User = ""
		sf.PrivateKey = ""
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Snowflake{Snowflake: sf}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Snowflake sets both key-pair credentials and secrets_manager", func() {
		sf := minimalSnowflake()
		sf.SecretsManager = minimalSecretsManager()
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Snowflake{Snowflake: sf}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Snowflake sets user without private_key", func() {
		sf := minimalSnowflake()
		sf.PrivateKey = ""
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Snowflake{Snowflake: sf}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Snowflake sets key_passphrase without private_key", func() {
		sf := minimalSnowflake()
		sf.User = ""
		sf.PrivateKey = ""
		sf.KeyPassphrase = "my-passphrase"
		sf.SecretsManager = minimalSecretsManager()
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Snowflake{Snowflake: sf}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Snowflake has invalid data_loading_option", func() {
		sf := minimalSnowflake()
		sf.DataLoadingOption = "CSV_MAPPING"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Snowflake{Snowflake: sf}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Snowflake VARIANT mapping lacks content_column_name", func() {
		sf := minimalSnowflake()
		sf.DataLoadingOption = "VARIANT_CONTENT_MAPPING"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Snowflake{Snowflake: sf}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Snowflake content+metadata mapping lacks metadata_column_name", func() {
		sf := minimalSnowflake()
		sf.DataLoadingOption = "VARIANT_CONTENT_AND_METADATA_MAPPING"
		sf.ContentColumnName = "RECORD_CONTENT"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Snowflake{Snowflake: sf}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Snowflake is missing the target table", func() {
		sf := minimalSnowflake()
		sf.Table = ""
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Snowflake{Snowflake: sf}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure: Iceberg validations
	// =========================================================================

	ginkgo.It("fails when Iceberg is missing catalog_arn", func() {
		ib := minimalIceberg()
		ib.CatalogArn = nil
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Iceberg{Iceberg: ib}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when an Iceberg destination table is missing table_name", func() {
		ib := minimalIceberg()
		ib.DestinationTables = []*AwsKinesisFirehoseIcebergDestinationTable{
			{DatabaseName: "lakehouse"},
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Iceberg{Iceberg: ib}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Iceberg has invalid s3_backup_mode", func() {
		ib := minimalIceberg()
		ib.S3BackupMode = "Enabled"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Iceberg{Iceberg: ib}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Iceberg is missing s3_config", func() {
		ib := minimalIceberg()
		ib.S3Config = nil
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Iceberg{Iceberg: ib}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure: Secrets Manager config validations
	// =========================================================================

	ginkgo.It("fails when secrets_manager is missing secret_arn", func() {
		http := minimalHttpEndpoint()
		http.SecretsManager = &AwsKinesisFirehoseSecretsManagerConfig{}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_HttpEndpoint{HttpEndpoint: http}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure: CloudWatch logging validations
	// =========================================================================

	ginkgo.It("fails when logging is enabled without log_group_name", func() {
		s3 := minimalExtendedS3()
		s3.Logging = &AwsKinesisFirehoseCloudwatchLogging{
			Enabled:       true,
			LogStreamName: "S3Delivery",
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when logging is enabled without log_stream_name", func() {
		s3 := minimalExtendedS3()
		s3.Logging = &AwsKinesisFirehoseCloudwatchLogging{
			Enabled:      true,
			LogGroupName: "/aws/kinesisfirehose/stream",
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// API envelope validations (from api.proto)
	// =========================================================================

	ginkgo.It("fails when apiVersion is wrong", func() {
		envelope := &AwsKinesisFirehose{
			ApiVersion: "wrong/v1",
			Kind:       "AwsKinesisFirehose",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kind is wrong", func() {
		envelope := &AwsKinesisFirehose{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "WrongKind",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		envelope := &AwsKinesisFirehose{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsKinesisFirehose",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		envelope := &AwsKinesisFirehose{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsKinesisFirehose",
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
