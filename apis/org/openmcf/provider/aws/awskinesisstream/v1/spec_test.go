package awskinesisstreamv1

import (
	"testing"

	"buf.build/go/protovalidate"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestAwsKinesisStreamSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsKinesisStreamSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

var _ = ginkgo.Describe("AwsKinesisStreamSpec validations", func() {
	var spec *AwsKinesisStreamSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: ON_DEMAND stream with all AWS defaults.
		spec = &AwsKinesisStreamSpec{
			StreamMode: "ON_DEMAND",
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal ON_DEMAND stream (all defaults)", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a minimal PROVISIONED stream with shard_count", func() {
		spec.StreamMode = "PROVISIONED"
		spec.ShardCount = 2
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a PROVISIONED stream with single shard", func() {
		spec.StreamMode = "PROVISIONED"
		spec.ShardCount = 1
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a stream with KMS encryption via direct key ARN", func() {
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a stream with KMS encryption via Kinesis-owned alias", func() {
		spec.KmsKeyId = strRef("alias/aws/kinesis")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a stream with KMS encryption via valueFrom reference", func() {
		spec.KmsKeyId = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{
					Kind:      219, // AwsKmsKey
					Name:      "my-encryption-key",
					FieldPath: "status.outputs.key_arn",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a stream with 48-hour retention", func() {
		spec.RetentionPeriodHours = 48
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a stream with maximum retention (365 days)", func() {
		spec.RetentionPeriodHours = 8760
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a stream with custom max record size", func() {
		spec.MaxRecordSizeInKib = 5120 // 5 MiB
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a stream with maximum record size (10 MiB)", func() {
		spec.MaxRecordSizeInKib = 10240
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a stream with enhanced shard-level metrics", func() {
		spec.ShardLevelMetrics = []string{
			"IncomingBytes",
			"IncomingRecords",
			"IteratorAgeMilliseconds",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a stream with all seven shard-level metrics", func() {
		spec.ShardLevelMetrics = []string{
			"IncomingBytes",
			"IncomingRecords",
			"OutgoingBytes",
			"OutgoingRecords",
			"WriteProvisionedThroughputExceeded",
			"ReadProvisionedThroughputExceeded",
			"IteratorAgeMilliseconds",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a stream with enforce_consumer_deletion enabled", func() {
		spec.EnforceConsumerDeletion = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a production-ready ON_DEMAND stream with all features", func() {
		spec.StreamMode = "ON_DEMAND"
		spec.RetentionPeriodHours = 168 // 7 days
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123")
		spec.MaxRecordSizeInKib = 2048 // 2 MiB
		spec.ShardLevelMetrics = []string{
			"IncomingBytes",
			"IncomingRecords",
			"OutgoingBytes",
			"OutgoingRecords",
			"WriteProvisionedThroughputExceeded",
			"ReadProvisionedThroughputExceeded",
			"IteratorAgeMilliseconds",
		}
		spec.EnforceConsumerDeletion = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a production-ready PROVISIONED stream with all features", func() {
		spec.StreamMode = "PROVISIONED"
		spec.ShardCount = 4
		spec.RetentionPeriodHours = 72 // 3 days
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123")
		spec.ShardLevelMetrics = []string{
			"WriteProvisionedThroughputExceeded",
			"ReadProvisionedThroughputExceeded",
			"IteratorAgeMilliseconds",
		}
		spec.EnforceConsumerDeletion = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: stream_mode_required
	// -------------------------------------------------------------------------

	ginkgo.It("fails when stream_mode is empty", func() {
		spec.StreamMode = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when stream_mode is invalid", func() {
		spec.StreamMode = "SERVERLESS"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when stream_mode has wrong casing", func() {
		spec.StreamMode = "on_demand"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: shard_count_required_for_provisioned
	// -------------------------------------------------------------------------

	ginkgo.It("fails when stream_mode is PROVISIONED but shard_count is 0", func() {
		spec.StreamMode = "PROVISIONED"
		spec.ShardCount = 0
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: shard_count_forbidden_for_on_demand
	// -------------------------------------------------------------------------

	ginkgo.It("fails when stream_mode is ON_DEMAND but shard_count is set", func() {
		spec.StreamMode = "ON_DEMAND"
		spec.ShardCount = 4
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: retention_period_range
	// -------------------------------------------------------------------------

	ginkgo.It("fails when retention_period_hours is below 24 (non-zero)", func() {
		spec.RetentionPeriodHours = 12
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when retention_period_hours exceeds 8760", func() {
		spec.RetentionPeriodHours = 8761
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when retention_period_hours is 1 (below minimum)", func() {
		spec.RetentionPeriodHours = 1
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: max_record_size_range
	// -------------------------------------------------------------------------

	ginkgo.It("fails when max_record_size_in_kib is below 1024 (non-zero)", func() {
		spec.MaxRecordSizeInKib = 512
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when max_record_size_in_kib exceeds 10240", func() {
		spec.MaxRecordSizeInKib = 10241
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: shard_level_metrics_valid
	// -------------------------------------------------------------------------

	ginkgo.It("fails when shard_level_metrics contains an invalid metric name", func() {
		spec.ShardLevelMetrics = []string{"IncomingBytes", "InvalidMetric"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when shard_level_metrics contains 'ALL' (must list individual metrics)", func() {
		spec.ShardLevelMetrics = []string{"ALL"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// API envelope validations (from api.proto)
	// -------------------------------------------------------------------------

	ginkgo.It("fails when apiVersion is wrong", func() {
		envelope := &AwsKinesisStream{
			ApiVersion: "wrong/v1",
			Kind:       "AwsKinesisStream",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kind is wrong", func() {
		envelope := &AwsKinesisStream{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "WrongKind",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		envelope := &AwsKinesisStream{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsKinesisStream",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		envelope := &AwsKinesisStream{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsKinesisStream",
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
