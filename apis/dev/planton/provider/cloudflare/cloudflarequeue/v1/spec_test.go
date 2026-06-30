package cloudflarequeuev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func value(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func validQueue() *CloudflareQueue {
	return &CloudflareQueue{
		ApiVersion: "cloudflare.planton.dev/v1",
		Kind:       "CloudflareQueue",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-queue"},
		Spec: &CloudflareQueueSpec{
			AccountId: validAccountID,
			QueueName: "orders-queue",
		},
	}
}

func TestCloudflareQueueSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareQueueSpec Validation Suite")
}

var _ = ginkgo.Describe("CloudflareQueueSpec Validation", func() {
	ginkgo.Describe("Valid inputs", func() {
		ginkgo.It("accepts a minimal queue", func() {
			gomega.Expect(protovalidate.Validate(validQueue())).To(gomega.BeNil())
		})

		ginkgo.It("accepts queue settings within range", func() {
			q := validQueue()
			q.Spec.Settings = &CloudflareQueueSettings{
				DeliveryDelay:          60,
				DeliveryPaused:         true,
				MessageRetentionPeriod: 86400,
			}
			gomega.Expect(protovalidate.Validate(q)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a worker consumer with a script reference", func() {
			q := validQueue()
			q.Spec.Consumer = &CloudflareQueueConsumer{
				Type:       CloudflareQueueConsumer_worker,
				ScriptName: value("orders-consumer"),
				Settings: &CloudflareQueueConsumerSettings{
					BatchSize:      25,
					MaxConcurrency: 10,
					MaxRetries:     3,
					MaxWaitTimeMs:  1000,
					RetryDelay:     30,
				},
			}
			gomega.Expect(protovalidate.Validate(q)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an http_pull consumer with a dead-letter queue", func() {
			q := validQueue()
			q.Spec.Consumer = &CloudflareQueueConsumer{
				Type:            CloudflareQueueConsumer_http_pull,
				DeadLetterQueue: value("orders-dlq"),
				Settings: &CloudflareQueueConsumerSettings{
					BatchSize:           50,
					VisibilityTimeoutMs: 60000,
				},
			}
			gomega.Expect(protovalidate.Validate(q)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Invalid inputs", func() {
		ginkgo.It("rejects a non-hex account_id", func() {
			q := validQueue()
			q.Spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(q)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing queue_name", func() {
			q := validQueue()
			q.Spec.QueueName = ""
			gomega.Expect(protovalidate.Validate(q)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid queue_name with illegal characters", func() {
			q := validQueue()
			q.Spec.QueueName = "orders queue!"
			gomega.Expect(protovalidate.Validate(q)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a message_retention_period below the minimum", func() {
			q := validQueue()
			q.Spec.Settings = &CloudflareQueueSettings{MessageRetentionPeriod: 30}
			gomega.Expect(protovalidate.Validate(q)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a consumer with unspecified type", func() {
			q := validQueue()
			q.Spec.Consumer = &CloudflareQueueConsumer{Type: CloudflareQueueConsumer_consumer_type_unspecified}
			gomega.Expect(protovalidate.Validate(q)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a worker consumer without a script_name", func() {
			q := validQueue()
			q.Spec.Consumer = &CloudflareQueueConsumer{Type: CloudflareQueueConsumer_worker}
			gomega.Expect(protovalidate.Validate(q)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an http_pull consumer that sets a script_name", func() {
			q := validQueue()
			q.Spec.Consumer = &CloudflareQueueConsumer{
				Type:       CloudflareQueueConsumer_http_pull,
				ScriptName: value("orders-consumer"),
			}
			gomega.Expect(protovalidate.Validate(q)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a batch_size above the maximum", func() {
			q := validQueue()
			q.Spec.Consumer = &CloudflareQueueConsumer{
				Type:       CloudflareQueueConsumer_worker,
				ScriptName: value("orders-consumer"),
				Settings:   &CloudflareQueueConsumerSettings{BatchSize: 200},
			}
			gomega.Expect(protovalidate.Validate(q)).ToNot(gomega.BeNil())
		})
	})
})
