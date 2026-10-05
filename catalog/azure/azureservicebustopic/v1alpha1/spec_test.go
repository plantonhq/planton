package azureservicebustopicv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureServiceBusTopicSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureServiceBusTopicSpec Validation Tests")
}

func minimalTopic() *AzureServiceBusTopic {
	return &AzureServiceBusTopic{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureServiceBusTopic",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-topic",
		},
		Spec: &AzureServiceBusTopicSpec{
			NamespaceId: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ServiceBus/namespaces/myapp-bus",
				},
			},
			TopicName: "events",
		},
	}
}

var _ = ginkgo.Describe("AzureServiceBusTopicSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_service_bus_topic", func() {

			ginkgo.It("should accept a minimal topic", func() {
				gomega.Expect(protovalidate.Validate(minimalTopic())).To(gomega.BeNil())
			})

			ginkgo.It("should accept a namespace reference by valueFrom", func() {
				input := minimalTopic()
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

			ginkgo.It("should accept hierarchical topic names with slashes", func() {
				input := minimalTopic()
				input.Spec.TopicName = "billing/invoices"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept every allowed max size", func() {
				for _, s := range []int32{1024, 2048, 3072, 4096, 5120, 10240, 20480, 40960, 81920} {
					size := s
					input := minimalTopic()
					input.Spec.MaxSizeInMegabytes = &size
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept a fully-dialed topic", func() {
				size := int32(5120)
				partitioning := true
				dedup := true
				window := "PT30S"
				ttl := "P30D"
				idle := "PT10M"
				batched := true
				ordering := true
				input := minimalTopic()
				input.Spec.MaxSizeInMegabytes = &size
				input.Spec.PartitioningEnabled = &partitioning
				input.Spec.RequiresDuplicateDetection = &dedup
				input.Spec.DuplicateDetectionHistoryTimeWindow = &window
				input.Spec.DefaultMessageTtl = &ttl
				input.Spec.AutoDeleteOnIdle = &idle
				input.Spec.BatchedOperationsEnabled = &batched
				input.Spec.SupportOrdering = &ordering
				input.Spec.Status = AzureServiceBusTopicStatusValue_ACTIVE
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept the DISABLED gate state", func() {
				input := minimalTopic()
				input.Spec.Status = AzureServiceBusTopicStatusValue_DISABLED
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept express on its own", func() {
				express := true
				input := minimalTopic()
				input.Spec.ExpressEnabled = &express
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept premium large-message bounds", func() {
				for _, k := range []int32{1024, 102400} {
					kb := k
					input := minimalTopic()
					input.Spec.MaxMessageSizeInKilobytes = &kb
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_service_bus_topic", func() {

			ginkgo.It("should reject a missing namespace reference", func() {
				input := minimalTopic()
				input.Spec.NamespaceId = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing topic name", func() {
				input := minimalTopic()
				input.Spec.TopicName = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a topic name ending with a period", func() {
				input := minimalTopic()
				input.Spec.TopicName = "events."
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a topic name with illegal characters", func() {
				input := minimalTopic()
				input.Spec.TopicName = "events queue"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a max size outside the sold set", func() {
				size := int32(6000)
				input := minimalTopic()
				input.Spec.MaxSizeInMegabytes = &size
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a large-message size outside the premium range", func() {
				for _, k := range []int32{512, 204800} {
					kb := k
					input := minimalTopic()
					input.Spec.MaxMessageSizeInKilobytes = &kb
					gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
				}
			})

			ginkgo.It("should reject express combined with duplicate detection", func() {
				express := true
				dedup := true
				input := minimalTopic()
				input.Spec.ExpressEnabled = &express
				input.Spec.RequiresDuplicateDetection = &dedup
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a dedup window without duplicate detection", func() {
				window := "PT30S"
				input := minimalTopic()
				input.Spec.DuplicateDetectionHistoryTimeWindow = &window
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an undeclared gate state", func() {
				input := minimalTopic()
				input.Spec.Status = AzureServiceBusTopicStatusValue(99)
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing metadata block", func() {
				input := minimalTopic()
				input.Metadata = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an incorrect kind", func() {
				input := minimalTopic()
				input.Kind = "WrongKind"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
