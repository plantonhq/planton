package awssnstopicv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAwsSnsTopicSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsSnsTopicSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

// helper to create a google.protobuf.Struct from a map.
func newStruct(m map[string]interface{}) *structpb.Struct {
	s, _ := structpb.NewStruct(m)
	return s
}

var _ = ginkgo.Describe("AwsSnsTopicSpec validations", func() {
	var spec *AwsSnsTopicSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: a standard topic with all AWS defaults.
		spec = &AwsSnsTopicSpec{
			Region: "us-west-2",
		}
	})

	// -------------------------------------------------------------------------
	// Happy path — Spec level
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal standard topic (all defaults)", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a FIFO topic with content-based deduplication", func() {
		spec.FifoTopic = true
		spec.ContentBasedDeduplication = true
		spec.FifoThroughputScope = "MessageGroup"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a FIFO topic with Topic-level throughput scope", func() {
		spec.FifoTopic = true
		spec.FifoThroughputScope = "Topic"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a topic with KMS encryption", func() {
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a topic with display name", func() {
		spec.DisplayName = "Order Notifications"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a topic with tracing_config Active", func() {
		spec.TracingConfig = "Active"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a topic with tracing_config PassThrough", func() {
		spec.TracingConfig = "PassThrough"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a topic with signature_version 1", func() {
		spec.SignatureVersion = 1
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a topic with signature_version 2", func() {
		spec.SignatureVersion = 2
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a topic with an IAM access policy", func() {
		spec.Policy = newStruct(map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []interface{}{
				map[string]interface{}{
					"Effect":    "Allow",
					"Principal": "*",
					"Action":    "SNS:Publish",
					"Resource":  "*",
				},
			},
		})
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: FIFO-only fields on Standard topics
	// -------------------------------------------------------------------------

	ginkgo.It("fails when content_based_deduplication is set on a standard topic", func() {
		spec.FifoTopic = false
		spec.ContentBasedDeduplication = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when fifo_throughput_scope is set on a standard topic", func() {
		spec.FifoTopic = false
		spec.FifoThroughputScope = "Topic"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when fifo_throughput_scope has an invalid value", func() {
		spec.FifoTopic = true
		spec.FifoThroughputScope = "invalid"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: signature_version validation
	// -------------------------------------------------------------------------

	ginkgo.It("fails when signature_version is not 1 or 2", func() {
		spec.SignatureVersion = 3
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when signature_version is negative", func() {
		spec.SignatureVersion = -1
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: tracing_config validation
	// -------------------------------------------------------------------------

	ginkgo.It("fails when tracing_config has an invalid value", func() {
		spec.TracingConfig = "Disabled"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Happy path — Subscription level
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a topic with a valid SQS subscription", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:     "order-queue",
				Protocol: "sqs",
				Endpoint: strRef("arn:aws:sqs:us-east-1:123456789012:order-queue"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a topic with a Lambda subscription", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:     "processor",
				Protocol: "lambda",
				Endpoint: strRef("arn:aws:lambda:us-east-1:123456789012:function:process-order"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a subscription with filter policy on MessageAttributes", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:              "filtered-queue",
				Protocol:          "sqs",
				Endpoint:          strRef("arn:aws:sqs:us-east-1:123456789012:filtered-queue"),
				FilterPolicy:      newStruct(map[string]interface{}{"store": []interface{}{"example_corp"}}),
				FilterPolicyScope: "MessageAttributes",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a subscription with filter policy on MessageBody", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:              "body-filtered",
				Protocol:          "sqs",
				Endpoint:          strRef("arn:aws:sqs:us-east-1:123456789012:body-filtered"),
				FilterPolicy:      newStruct(map[string]interface{}{"source": []interface{}{"order-service"}}),
				FilterPolicyScope: "MessageBody",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a subscription with raw message delivery", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:               "raw-queue",
				Protocol:           "sqs",
				Endpoint:           strRef("arn:aws:sqs:us-east-1:123456789012:raw-queue"),
				RawMessageDelivery: true,
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a subscription with redrive config (DLQ)", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:     "with-dlq",
				Protocol: "sqs",
				Endpoint: strRef("arn:aws:sqs:us-east-1:123456789012:my-queue"),
				RedriveConfig: &AwsSnsSubscriptionRedriveConfig{
					DeadLetterTargetArn: strRef("arn:aws:sqs:us-east-1:123456789012:my-dlq"),
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a Firehose subscription with subscription_role_arn", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:                "firehose-delivery",
				Protocol:            "firehose",
				Endpoint:            strRef("arn:aws:firehose:us-east-1:123456789012:deliverystream/my-stream"),
				SubscriptionRoleArn: strRef("arn:aws:iam::123456789012:role/sns-firehose-role"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an email subscription", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:     "alert-email",
				Protocol: "email",
				Endpoint: strRef("ops-team@example.com"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an HTTPS subscription", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:     "webhook",
				Protocol: "https",
				Endpoint: strRef("https://api.example.com/sns-webhook"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts multiple subscriptions", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:     "order-queue",
				Protocol: "sqs",
				Endpoint: strRef("arn:aws:sqs:us-east-1:123456789012:order-queue"),
			},
			{
				Name:     "audit-lambda",
				Protocol: "lambda",
				Endpoint: strRef("arn:aws:lambda:us-east-1:123456789012:function:audit"),
			},
			{
				Name:     "ops-email",
				Protocol: "email",
				Endpoint: strRef("ops@example.com"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Subscription-level validation failures
	// -------------------------------------------------------------------------

	ginkgo.It("fails when subscription name is missing", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Protocol: "sqs",
				Endpoint: strRef("arn:aws:sqs:us-east-1:123456789012:my-queue"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when subscription protocol is missing", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:     "my-sub",
				Endpoint: strRef("arn:aws:sqs:us-east-1:123456789012:my-queue"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when subscription endpoint is missing", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:     "my-sub",
				Protocol: "sqs",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when subscription protocol is invalid", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:     "my-sub",
				Protocol: "kafka",
				Endpoint: strRef("some-endpoint"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when filter_policy_scope is invalid", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:              "my-sub",
				Protocol:          "sqs",
				Endpoint:          strRef("arn:aws:sqs:us-east-1:123456789012:my-queue"),
				FilterPolicy:      newStruct(map[string]interface{}{"key": "value"}),
				FilterPolicyScope: "InvalidScope",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when filter_policy_scope is set without filter_policy", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:              "my-sub",
				Protocol:          "sqs",
				Endpoint:          strRef("arn:aws:sqs:us-east-1:123456789012:my-queue"),
				FilterPolicyScope: "MessageBody",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when firehose subscription is missing subscription_role_arn", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:     "my-firehose",
				Protocol: "firehose",
				Endpoint: strRef("arn:aws:firehose:us-east-1:123456789012:deliverystream/my-stream"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when redrive_config.dead_letter_target_arn is missing", func() {
		spec.Subscriptions = []*AwsSnsTopicSubscription{
			{
				Name:          "my-sub",
				Protocol:      "sqs",
				Endpoint:      strRef("arn:aws:sqs:us-east-1:123456789012:my-queue"),
				RedriveConfig: &AwsSnsSubscriptionRedriveConfig{},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
