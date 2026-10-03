package azureservicebusqueuev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureServiceBusQueueSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureServiceBusQueueSpec Validation Tests")
}

func minimalQueue() *AzureServiceBusQueue {
	return &AzureServiceBusQueue{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureServiceBusQueue",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-queue",
		},
		Spec: &AzureServiceBusQueueSpec{
			NamespaceId: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ServiceBus/namespaces/myapp-bus",
				},
			},
			QueueName: "orders",
		},
	}
}

var _ = ginkgo.Describe("AzureServiceBusQueueSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_service_bus_queue", func() {

			ginkgo.It("should accept a minimal queue", func() {
				gomega.Expect(protovalidate.Validate(minimalQueue())).To(gomega.BeNil())
			})

			ginkgo.It("should accept a namespace reference by valueFrom", func() {
				input := minimalQueue()
				input.Spec.NamespaceId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureServiceBusNamespace,
							Name:      "shared-bus",
							FieldPath: "status.outputs.namespace_id",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept hierarchical queue names with slashes", func() {
				input := minimalQueue()
				input.Spec.QueueName = "orders/priority"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a single-character queue name", func() {
				input := minimalQueue()
				input.Spec.QueueName = "q"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept every allowed max size", func() {
				for _, s := range []int32{1024, 2048, 3072, 4096, 5120, 10240, 20480, 40960, 81920} {
					size := s
					input := minimalQueue()
					input.Spec.MaxSizeInMegabytes = &size
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept premium large-message bounds", func() {
				for _, k := range []int32{1024, 102400} {
					kb := k
					input := minimalQueue()
					input.Spec.MaxMessageSizeInKilobytes = &kb
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept a fully-dialed queue", func() {
				size := int32(5120)
				partitioning := true
				dedup := true
				window := "PT30S"
				ttl := "P14D"
				lock := "PT2M"
				maxDelivery := int32(5)
				dlOnExpire := true
				session := true
				idle := "PT10M"
				batched := true
				forwardDlq := &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "poison-sink"},
				}
				input := minimalQueue()
				input.Spec.MaxSizeInMegabytes = &size
				input.Spec.PartitioningEnabled = &partitioning
				input.Spec.RequiresDuplicateDetection = &dedup
				input.Spec.DuplicateDetectionHistoryTimeWindow = &window
				input.Spec.DefaultMessageTtl = &ttl
				input.Spec.LockDuration = &lock
				input.Spec.MaxDeliveryCount = &maxDelivery
				input.Spec.DeadLetteringOnMessageExpiration = &dlOnExpire
				input.Spec.RequiresSession = &session
				input.Spec.AutoDeleteOnIdle = &idle
				input.Spec.BatchedOperationsEnabled = &batched
				input.Spec.ForwardDeadLetteredMessagesTo = forwardDlq
				input.Spec.Status = AzureServiceBusEntityStatus_ACTIVE
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept forwarding to another entity", func() {
				input := minimalQueue()
				input.Spec.ForwardTo = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "downstream-topic"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept every declared gate state", func() {
				for _, s := range []AzureServiceBusEntityStatus{
					AzureServiceBusEntityStatus_ACTIVE,
					AzureServiceBusEntityStatus_DISABLED,
					AzureServiceBusEntityStatus_SEND_DISABLED,
					AzureServiceBusEntityStatus_RECEIVE_DISABLED,
				} {
					input := minimalQueue()
					input.Spec.Status = s
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept express on its own", func() {
				express := true
				input := minimalQueue()
				input.Spec.ExpressEnabled = &express
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_service_bus_queue", func() {

			ginkgo.It("should reject a missing namespace reference", func() {
				input := minimalQueue()
				input.Spec.NamespaceId = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing queue name", func() {
				input := minimalQueue()
				input.Spec.QueueName = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a queue name starting with a hyphen", func() {
				input := minimalQueue()
				input.Spec.QueueName = "-orders"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a queue name ending with a slash", func() {
				input := minimalQueue()
				input.Spec.QueueName = "orders/"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a queue name with illegal characters", func() {
				input := minimalQueue()
				input.Spec.QueueName = "orders queue"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a max size outside the sold set", func() {
				size := int32(1500)
				input := minimalQueue()
				input.Spec.MaxSizeInMegabytes = &size
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a large-message size below 1024 KB", func() {
				kb := int32(512)
				input := minimalQueue()
				input.Spec.MaxMessageSizeInKilobytes = &kb
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a large-message size above 102400 KB", func() {
				kb := int32(204800)
				input := minimalQueue()
				input.Spec.MaxMessageSizeInKilobytes = &kb
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a zero max_delivery_count", func() {
				maxDelivery := int32(0)
				input := minimalQueue()
				input.Spec.MaxDeliveryCount = &maxDelivery
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject express combined with duplicate detection", func() {
				express := true
				dedup := true
				input := minimalQueue()
				input.Spec.ExpressEnabled = &express
				input.Spec.RequiresDuplicateDetection = &dedup
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a dedup window without duplicate detection", func() {
				window := "PT30S"
				input := minimalQueue()
				input.Spec.DuplicateDetectionHistoryTimeWindow = &window
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an undeclared gate state", func() {
				input := minimalQueue()
				input.Spec.Status = AzureServiceBusEntityStatus(99)
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing metadata block", func() {
				input := minimalQueue()
				input.Metadata = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an incorrect kind", func() {
				input := minimalQueue()
				input.Kind = "WrongKind"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
