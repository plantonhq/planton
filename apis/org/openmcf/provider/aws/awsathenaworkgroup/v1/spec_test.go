package awsathenaworkgroupv1

import (
	"testing"

	"buf.build/go/protovalidate"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
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
		// Minimal valid spec: empty spec (all optional, all have sane defaults).
		spec = &AwsAthenaWorkgroupSpec{}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "WrongKind",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		envelope := &AwsAthenaWorkgroup{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsAthenaWorkgroup",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		envelope := &AwsAthenaWorkgroup{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsAthenaWorkgroup",
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
