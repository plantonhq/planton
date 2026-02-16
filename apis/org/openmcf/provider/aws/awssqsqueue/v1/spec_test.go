package awssqsqueuev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsSqsQueueSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsSqsQueueSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

var _ = ginkgo.Describe("AwsSqsQueueSpec validations", func() {
	var spec *AwsSqsQueueSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: a standard queue with all AWS defaults.
		spec = &AwsSqsQueueSpec{}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal standard queue (all defaults)", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a valid FIFO queue with content-based deduplication", func() {
		spec.FifoQueue = true
		spec.ContentBasedDeduplication = true
		spec.DeduplicationScope = "messageGroup"
		spec.FifoThroughputLimit = "perMessageGroupId"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a standard queue with SQS-managed SSE", func() {
		spec.SqsManagedSseEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a standard queue with KMS encryption", func() {
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123")
		spec.KmsDataKeyReusePeriodSeconds = 600
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a queue with dead letter config", func() {
		spec.DeadLetterConfig = &AwsSqsQueueDeadLetterConfig{
			TargetArn:       strRef("arn:aws:sqs:us-east-1:123456789012:my-dlq"),
			MaxReceiveCount: 5,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a queue with all delivery settings", func() {
		spec.VisibilityTimeoutSeconds = 60
		spec.MessageRetentionSeconds = 86400
		spec.MaxMessageSizeBytes = 131072
		spec.DelaySeconds = 10
		spec.ReceiveWaitTimeSeconds = 20
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level range validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when visibility_timeout_seconds exceeds 43200", func() {
		spec.VisibilityTimeoutSeconds = 43201
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when delay_seconds exceeds 900", func() {
		spec.DelaySeconds = 901
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when receive_wait_time_seconds exceeds 20", func() {
		spec.ReceiveWaitTimeSeconds = 21
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: message_retention_range
	// -------------------------------------------------------------------------

	ginkgo.It("fails when message_retention_seconds is below 60 (non-zero)", func() {
		spec.MessageRetentionSeconds = 30
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when message_retention_seconds exceeds 1209600", func() {
		spec.MessageRetentionSeconds = 1209601
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: max_message_size_range
	// -------------------------------------------------------------------------

	ginkgo.It("fails when max_message_size_bytes is below 1024 (non-zero)", func() {
		spec.MaxMessageSizeBytes = 512
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when max_message_size_bytes exceeds 1048576", func() {
		spec.MaxMessageSizeBytes = 1048577
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: FIFO-only fields on Standard queues
	// -------------------------------------------------------------------------

	ginkgo.It("fails when content_based_deduplication is set on a standard queue", func() {
		spec.FifoQueue = false
		spec.ContentBasedDeduplication = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when deduplication_scope is set on a standard queue", func() {
		spec.FifoQueue = false
		spec.DeduplicationScope = "messageGroup"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when fifo_throughput_limit is set on a standard queue", func() {
		spec.FifoQueue = false
		spec.FifoThroughputLimit = "perQueue"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when deduplication_scope has an invalid value", func() {
		spec.FifoQueue = true
		spec.DeduplicationScope = "invalid"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when fifo_throughput_limit has an invalid value", func() {
		spec.FifoQueue = true
		spec.FifoThroughputLimit = "invalid"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: Encryption mutual exclusion
	// -------------------------------------------------------------------------

	ginkgo.It("fails when both kms_key_id and sqs_managed_sse_enabled are set", func() {
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123")
		spec.SqsManagedSseEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: KMS data key reuse requires KMS key
	// -------------------------------------------------------------------------

	ginkgo.It("fails when kms_data_key_reuse_period_seconds is set without kms_key_id", func() {
		spec.KmsDataKeyReusePeriodSeconds = 300
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kms_data_key_reuse_period_seconds is below 60 (non-zero)", func() {
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123")
		spec.KmsDataKeyReusePeriodSeconds = 30
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kms_data_key_reuse_period_seconds exceeds 86400", func() {
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123")
		spec.KmsDataKeyReusePeriodSeconds = 86401
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Dead letter config field-level validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when dead_letter_config.target_arn is missing", func() {
		spec.DeadLetterConfig = &AwsSqsQueueDeadLetterConfig{
			MaxReceiveCount: 5,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when dead_letter_config.max_receive_count is 0", func() {
		spec.DeadLetterConfig = &AwsSqsQueueDeadLetterConfig{
			TargetArn:       strRef("arn:aws:sqs:us-east-1:123456789012:my-dlq"),
			MaxReceiveCount: 0,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when dead_letter_config.max_receive_count exceeds 1000", func() {
		spec.DeadLetterConfig = &AwsSqsQueueDeadLetterConfig{
			TargetArn:       strRef("arn:aws:sqs:us-east-1:123456789012:my-dlq"),
			MaxReceiveCount: 1001,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
