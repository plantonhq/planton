package awskinesisfirehosev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
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
		S3Config: &AwsKinesisFirehoseS3Config{
			BucketArn: strRef("arn:aws:s3:::backup-bucket"),
			RoleArn:   strRef("arn:aws:iam::123456789012:role/firehose-s3"),
		},
	}
}

// helper: minimal valid HTTP endpoint destination.
func minimalHttpEndpoint() *AwsKinesisFirehoseHttpEndpointDestination {
	return &AwsKinesisFirehoseHttpEndpointDestination{
		Url: "https://api.example.com/firehose",
		S3Config: &AwsKinesisFirehoseS3Config{
			BucketArn: strRef("arn:aws:s3:::backup-bucket"),
			RoleArn:   strRef("arn:aws:iam::123456789012:role/firehose-s3"),
		},
	}
}

// helper: minimal valid Redshift destination.
func minimalRedshift() *AwsKinesisFirehoseRedshiftDestination {
	return &AwsKinesisFirehoseRedshiftDestination{
		ClusterJdbcurl: "jdbc:redshift://cluster.abc.us-east-1.redshift.amazonaws.com:5439/mydb",
		RoleArn:        strRef("arn:aws:iam::123456789012:role/firehose-rs"),
		DataTableName:  "events",
		S3Config: &AwsKinesisFirehoseS3Config{
			BucketArn: strRef("arn:aws:s3:::staging-bucket"),
			RoleArn:   strRef("arn:aws:iam::123456789012:role/firehose-s3"),
		},
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

	ginkgo.It("accepts Extended S3 with Lambda processing", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseLambdaProcessing{
			Enabled:   true,
			LambdaArn: strRef("arn:aws:lambda:us-east-1:123456789012:function:transform"),
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Extended S3 with data format conversion to Parquet", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled:      true,
			InputFormat:  "OPENX_JSON",
			OutputFormat: "PARQUET",
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

	ginkgo.It("accepts Extended S3 with dynamic partitioning", func() {
		s3 := minimalExtendedS3()
		s3.DynamicPartitioning = &AwsKinesisFirehoseDynamicPartitioning{
			Enabled: true,
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Extended S3 with S3 backup enabled", func() {
		s3 := minimalExtendedS3()
		s3.S3BackupMode = "Enabled"
		s3.S3Backup = &AwsKinesisFirehoseS3Config{
			BucketArn: strRef("arn:aws:s3:::backup-bucket"),
			RoleArn:   strRef("arn:aws:iam::123456789012:role/firehose-backup"),
		}
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

	ginkgo.It("accepts OpenSearch with index rotation and backup mode", func() {
		os := minimalOpenSearch()
		os.IndexRotationPeriod = "OneHour"
		os.S3BackupMode = "AllDocuments"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Opensearch{Opensearch: os}
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

	// =========================================================================
	// Happy path: Redshift destination
	// =========================================================================

	ginkgo.It("accepts a minimal Redshift destination", func() {
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Redshift{
			Redshift: minimalRedshift(),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Redshift with credentials and COPY options", func() {
		rs := minimalRedshift()
		rs.Username = "admin"
		rs.Password = strRef("SecretPass123!")
		rs.CopyOptions = "JSON 'auto'"
		rs.DataTableColumns = "id,event_type,created_at"
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Redshift{Redshift: rs}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
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
			Prefix:            "data/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/",
			ErrorOutputPrefix: "errors/year=!{timestamp:yyyy}/",
			KmsKeyArn:         strRef("arn:aws:kms:us-east-1:123456789012:key/prod-key"),
			FileExtension:     ".parquet",
			Buffering: &AwsKinesisFirehoseBufferingHints{
				IntervalInSeconds: 120,
				SizeInMbs:         64,
			},
			Processing: &AwsKinesisFirehoseLambdaProcessing{
				Enabled:                 true,
				LambdaArn:               strRef("arn:aws:lambda:us-east-1:123456789012:function:enrich"),
				BufferSizeInMbs:         3,
				BufferIntervalInSeconds: 60,
				NumberOfRetries:         3,
			},
			DynamicPartitioning: &AwsKinesisFirehoseDynamicPartitioning{
				Enabled:                true,
				RetryDurationInSeconds: 600,
			},
			DataFormatConversion: &AwsKinesisFirehoseDataFormatConversion{
				Enabled:            true,
				InputFormat:        "OPENX_JSON",
				OutputFormat:       "PARQUET",
				ParquetCompression: "SNAPPY",
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
	// Failure: SSE conflicts with Kinesis source
	// =========================================================================

	ginkgo.It("fails when SSE is enabled with Kinesis stream source", func() {
		spec.SseEnabled = true
		spec.KinesisStreamSource = &AwsKinesisFirehoseKinesisStreamSource{
			StreamArn: strRef("arn:aws:kinesis:us-east-1:123456789012:stream/my-stream"),
			RoleArn:   strRef("arn:aws:iam::123456789012:role/firehose-kinesis"),
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
		s3.S3Backup = &AwsKinesisFirehoseS3Config{
			BucketArn: strRef("arn:aws:s3:::backup"),
			RoleArn:   strRef("arn:aws:iam::123456789012:role/backup"),
		}
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

	ginkgo.It("fails when data format conversion is enabled without output_format", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled: true,
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
			OutputFormat: "PARQUET",
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when data format conversion has invalid output_format", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled:      true,
			OutputFormat: "AVRO",
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

	ginkgo.It("fails when parquet_compression is set with ORC output_format", func() {
		s3 := minimalExtendedS3()
		s3.DataFormatConversion = &AwsKinesisFirehoseDataFormatConversion{
			Enabled:            true,
			OutputFormat:       "ORC",
			ParquetCompression: "SNAPPY",
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
	// Failure: Lambda processing validations
	// =========================================================================

	ginkgo.It("fails when processing is enabled without lambda_arn", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseLambdaProcessing{
			Enabled: true,
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when processing buffer_size_in_mbs exceeds 3", func() {
		s3 := minimalExtendedS3()
		s3.Processing = &AwsKinesisFirehoseLambdaProcessing{
			Enabled:         true,
			LambdaArn:       strRef("arn:aws:lambda:us-east-1:123456789012:function:fn"),
			BufferSizeInMbs: 5,
		}
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_ExtendedS3{ExtendedS3: s3}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
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
	// Failure: HTTP endpoint validations
	// =========================================================================

	ginkgo.It("fails when HTTP endpoint URL is not HTTPS", func() {
		http := minimalHttpEndpoint()
		http.Url = "http://api.example.com/firehose"
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
		spec.DestinationConfig = &AwsKinesisFirehoseSpec_Redshift{
			Redshift: &AwsKinesisFirehoseRedshiftDestination{
				RoleArn:       strRef("arn:aws:iam::123456789012:role/firehose-rs"),
				DataTableName: "events",
				S3Config: &AwsKinesisFirehoseS3Config{
					BucketArn: strRef("arn:aws:s3:::staging"),
					RoleArn:   strRef("arn:aws:iam::123456789012:role/s3"),
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when Redshift s3_backup is set without s3_backup_mode Enabled", func() {
		rs := minimalRedshift()
		rs.S3Backup = &AwsKinesisFirehoseS3Config{
			BucketArn: strRef("arn:aws:s3:::backup"),
			RoleArn:   strRef("arn:aws:iam::123456789012:role/backup"),
		}
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
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "WrongKind",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		envelope := &AwsKinesisFirehose{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsKinesisFirehose",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		envelope := &AwsKinesisFirehose{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsKinesisFirehose",
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
