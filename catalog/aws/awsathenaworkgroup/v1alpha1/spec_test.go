package awsathenaworkgroupv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAwsAthenaWorkgroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsAthenaWorkgroupSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

var _ = ginkgo.Describe("AwsAthenaWorkgroupSpec validations", func() {
	var spec *AwsAthenaWorkgroupSpec

	ginkgo.BeforeEach(func() {
		spec = &AwsAthenaWorkgroupSpec{
			Region: "us-west-2",
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("rejects a managed-results KMS alias -- this field is ARN-only", func() {
		spec.ManagedQueryResults = &AwsAthenaWorkgroupManagedQueryResults{
			KmsKey: strRef("alias/athena-results"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("aliases are not accepted here"))
	})

	ginkgo.It("accepts a managed-results KMS key ARN", func() {
		spec.ManagedQueryResults = &AwsAthenaWorkgroupManagedQueryResults{
			KmsKey: strRef("arn:aws:kms:us-west-2:111122223333:key/abc"),
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("rejects a CloudWatch log group name with illegal characters", func() {
		spec.Monitoring = &AwsAthenaWorkgroupMonitoringConfig{
			CloudWatchLogging: &AwsAthenaWorkgroupCloudWatchLoggingConfig{
				LogGroup: "my log group",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an s3 logging location with an uppercase bucket", func() {
		spec.Monitoring = &AwsAthenaWorkgroupMonitoringConfig{
			S3Logging: &AwsAthenaWorkgroupS3LoggingConfig{
				LogLocation: "s3://My-Bucket/logs/",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a literal execution role that is not an ARN", func() {
		spec.ExecutionRole = strRef("athena-spark-role")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("must be an IAM role ARN"))
	})

	ginkgo.It("accepts a minimal empty spec (all defaults)", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with S3 output location only", func() {
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation: "s3://my-athena-results/queries/",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with SSE_S3 encryption", func() {
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation:   "s3://my-athena-results/",
			EncryptionOption: "SSE_S3",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with SSE_KMS encryption and KMS key ARN", func() {
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation:   "s3://my-athena-results/",
			EncryptionOption: "SSE_KMS",
			KmsKeyArn:        strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with CSE_KMS encryption and KMS key ARN", func() {
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation:   "s3://my-athena-results/",
			EncryptionOption: "CSE_KMS",
			KmsKeyArn:        strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with KMS key via valueFrom reference", func() {
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation:   "s3://my-athena-results/",
			EncryptionOption: "SSE_KMS",
			KmsKeyArn: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      219, // AwsKmsKey
						Name:      "my-encryption-key",
						FieldPath: "status.outputs.key_arn",
					},
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with bytes_scanned_cutoff at minimum (10 MB)", func() {
		spec.BytesScannedCutoffPerQuery = 10485760
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with bytes_scanned_cutoff at 10 GB", func() {
		spec.BytesScannedCutoffPerQuery = 10737418240
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with enforce_workgroup_configuration set to false", func() {
		f := false
		spec.EnforceWorkgroupConfiguration = &f
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with publish_cloudwatch_metrics disabled", func() {
		f := false
		spec.PublishCloudwatchMetricsEnabled = &f
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with requester_pays enabled", func() {
		spec.RequesterPaysEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with minimum encryption enforcement enabled", func() {
		spec.EnableMinimumEncryptionConfiguration = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with specific engine version", func() {
		spec.SelectedEngineVersion = "Athena engine version 3"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with engine version set to AUTO", func() {
		spec.SelectedEngineVersion = "AUTO"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with force_destroy enabled", func() {
		spec.ForceDestroy = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with execution_role for Spark workgroup", func() {
		spec.ExecutionRole = strRef("arn:aws:iam::123456789012:role/AthenaSparkRole")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with expected_bucket_owner for cross-account", func() {
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation:      "s3://cross-account-bucket/results/",
			ExpectedBucketOwner: "987654321098",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with s3_acl_option for cross-account", func() {
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation:      "s3://cross-account-bucket/results/",
			ExpectedBucketOwner: "987654321098",
			S3AclOption:         "BUCKET_OWNER_FULL_CONTROL",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a production-ready SQL workgroup with all features", func() {
		t := true
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation:   "s3://prod-athena-results/queries/",
			EncryptionOption: "SSE_KMS",
			KmsKeyArn:        strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123"),
		}
		spec.BytesScannedCutoffPerQuery = 10737418240 // 10 GB
		spec.EnforceWorkgroupConfiguration = &t
		spec.PublishCloudwatchMetricsEnabled = &t
		spec.EnableMinimumEncryptionConfiguration = true
		spec.SelectedEngineVersion = "Athena engine version 3"
		spec.ForceDestroy = false
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a Spark workgroup with execution role", func() {
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation:   "s3://spark-results/notebooks/",
			EncryptionOption: "SSE_S3",
		}
		spec.ExecutionRole = strRef("arn:aws:iam::123456789012:role/AthenaSparkExecutionRole")
		spec.SelectedEngineVersion = "PySpark engine version 3"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// description + state
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a spec with a description", func() {
		spec.Description = "Interactive analytics workgroup for the data platform team"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when description exceeds 1024 characters", func() {
		spec.Description = strings.Repeat("x", 1025)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts state ENABLED", func() {
		s := "ENABLED"
		spec.State = &s
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts state DISABLED", func() {
		s := "DISABLED"
		spec.State = &s
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid state value", func() {
		s := "PAUSED"
		spec.State = &s
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Managed query results + CEL: managed_results_excludes_output_location
	// -------------------------------------------------------------------------

	ginkgo.It("accepts an empty managed_query_results block (AWS-owned key)", func() {
		spec.ManagedQueryResults = &AwsAthenaWorkgroupManagedQueryResults{}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts managed_query_results declared and switched off beside an S3 output_location", func() {
		spec.ManagedQueryResults = &AwsAthenaWorkgroupManagedQueryResults{Enabled: proto.Bool(false)}
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{OutputLocation: "s3://my-results/"}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("accepts every monitoring destination declared and switched off", func() {
		spec.Monitoring = &AwsAthenaWorkgroupMonitoringConfig{
			CloudWatchLogging: &AwsAthenaWorkgroupCloudWatchLoggingConfig{Enabled: proto.Bool(false), LogGroup: "/athena/quiet"},
			ManagedLogging:    &AwsAthenaWorkgroupManagedLoggingConfig{Enabled: proto.Bool(false)},
			S3Logging:         &AwsAthenaWorkgroupS3LoggingConfig{Enabled: proto.Bool(false), LogLocation: "s3://logs/athena/"},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("accepts managed_query_results with a customer KMS key", func() {
		spec.ManagedQueryResults = &AwsAthenaWorkgroupManagedQueryResults{
			KmsKey: strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when managed_query_results is combined with an S3 output_location", func() {
		spec.ManagedQueryResults = &AwsAthenaWorkgroupManagedQueryResults{}
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation: "s3://my-results/",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts managed_query_results alongside a result_configuration without output_location", func() {
		spec.ManagedQueryResults = &AwsAthenaWorkgroupManagedQueryResults{}
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			EncryptionOption: "SSE_S3",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Spark content encryption
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a customer content encryption KMS key", func() {
		spec.CustomerContentEncryptionKmsKey = strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Identity Center + S3 Access Grants
	// -------------------------------------------------------------------------

	ginkgo.It("accepts an identity_center block with a valid instance ARN", func() {
		spec.IdentityCenter = &AwsAthenaWorkgroupIdentityCenterConfig{
			EnableIdentityCenter:      true,
			IdentityCenterInstanceArn: "arn:aws:sso:::instance/ssoins-1234567890abcdef",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when identity_center_instance_arn is not an ARN", func() {
		spec.IdentityCenter = &AwsAthenaWorkgroupIdentityCenterConfig{
			EnableIdentityCenter:      true,
			IdentityCenterInstanceArn: "ssoins-1234567890abcdef",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts an s3_access_grants block with DIRECTORY_IDENTITY", func() {
		spec.S3AccessGrants = &AwsAthenaWorkgroupS3AccessGrantsConfig{
			EnableS3AccessGrants:  true,
			AuthenticationType:    "DIRECTORY_IDENTITY",
			CreateUserLevelPrefix: true,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when s3_access_grants omits authentication_type", func() {
		spec.S3AccessGrants = &AwsAthenaWorkgroupS3AccessGrantsConfig{
			EnableS3AccessGrants: true,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when s3_access_grants authentication_type is invalid", func() {
		spec.S3AccessGrants = &AwsAthenaWorkgroupS3AccessGrantsConfig{
			EnableS3AccessGrants: true,
			AuthenticationType:   "IAM_IDENTITY",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Monitoring
	// -------------------------------------------------------------------------

	ginkgo.It("accepts CloudWatch logging with log group, prefix, and Spark log types", func() {
		spec.Monitoring = &AwsAthenaWorkgroupMonitoringConfig{
			CloudWatchLogging: &AwsAthenaWorkgroupCloudWatchLoggingConfig{
				LogGroup:            "/athena/analytics-workgroup",
				LogStreamNamePrefix: "analytics",
				LogTypes: []*AwsAthenaWorkgroupLogTypeEntry{
					{Key: "SPARK_DRIVER", Values: []string{"STDOUT", "STDERR"}},
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when a log_types entry has an empty key", func() {
		spec.Monitoring = &AwsAthenaWorkgroupMonitoringConfig{
			CloudWatchLogging: &AwsAthenaWorkgroupCloudWatchLoggingConfig{
				LogTypes: []*AwsAthenaWorkgroupLogTypeEntry{
					{Key: "", Values: []string{"DRIVER"}},
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a log_types entry has no values", func() {
		spec.Monitoring = &AwsAthenaWorkgroupMonitoringConfig{
			CloudWatchLogging: &AwsAthenaWorkgroupCloudWatchLoggingConfig{
				LogTypes: []*AwsAthenaWorkgroupLogTypeEntry{
					{Key: "SPARK_DRIVER", Values: []string{}},
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts managed logging with a KMS key", func() {
		spec.Monitoring = &AwsAthenaWorkgroupMonitoringConfig{
			ManagedLogging: &AwsAthenaWorkgroupManagedLoggingConfig{
				KmsKey: strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts S3 logging with an s3:// location", func() {
		spec.Monitoring = &AwsAthenaWorkgroupMonitoringConfig{
			S3Logging: &AwsAthenaWorkgroupS3LoggingConfig{
				LogLocation: "s3://my-log-archive/athena/",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when s3_logging log_location is not an s3 URI", func() {
		spec.Monitoring = &AwsAthenaWorkgroupMonitoringConfig{
			S3Logging: &AwsAthenaWorkgroupS3LoggingConfig{
				LogLocation: "https://my-log-archive/athena/",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts all three monitoring arms together", func() {
		spec.Monitoring = &AwsAthenaWorkgroupMonitoringConfig{
			CloudWatchLogging: &AwsAthenaWorkgroupCloudWatchLoggingConfig{
				LogGroup: "/athena/wg",
			},
			ManagedLogging: &AwsAthenaWorkgroupManagedLoggingConfig{},
			S3Logging: &AwsAthenaWorkgroupS3LoggingConfig{
				LogLocation: "s3://archive/athena/",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: bytes_scanned_cutoff_range
	// -------------------------------------------------------------------------

	ginkgo.It("fails when bytes_scanned_cutoff_per_query is below 10 MB", func() {
		spec.BytesScannedCutoffPerQuery = 1048576 // 1 MB — below minimum
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when bytes_scanned_cutoff_per_query is 1 byte", func() {
		spec.BytesScannedCutoffPerQuery = 1
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when bytes_scanned_cutoff_per_query is just below minimum", func() {
		spec.BytesScannedCutoffPerQuery = 10485759 // 10 MB - 1
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: result_encryption_option_valid
	// -------------------------------------------------------------------------

	ginkgo.It("fails when encryption_option is invalid", func() {
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation:   "s3://my-results/",
			EncryptionOption: "AES256",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when encryption_option has wrong casing", func() {
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation:   "s3://my-results/",
			EncryptionOption: "sse_s3",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when encryption_option is a random string", func() {
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation:   "s3://my-results/",
			EncryptionOption: "ENCRYPT_ALL_THE_THINGS",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: s3_acl_option_valid
	// -------------------------------------------------------------------------

	ginkgo.It("fails when s3_acl_option is invalid", func() {
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation: "s3://my-results/",
			S3AclOption:    "PUBLIC_READ",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when s3_acl_option has wrong casing", func() {
		spec.ResultConfiguration = &AwsAthenaWorkgroupResultConfig{
			OutputLocation: "s3://my-results/",
			S3AclOption:    "bucket_owner_full_control",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// API envelope validations (from api.proto)
	// -------------------------------------------------------------------------

	ginkgo.It("fails when apiVersion is wrong", func() {
		envelope := &AwsAthenaWorkgroup{
			ApiVersion: "wrong/v1",
			Kind:       "AwsAthenaWorkgroup",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kind is wrong", func() {
		envelope := &AwsAthenaWorkgroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "WrongKind",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		envelope := &AwsAthenaWorkgroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsAthenaWorkgroup",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		envelope := &AwsAthenaWorkgroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsAthenaWorkgroup",
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
