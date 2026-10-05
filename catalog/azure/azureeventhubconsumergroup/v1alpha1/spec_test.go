package azureeventhubconsumergroupv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureEventHubConsumerGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventHubConsumerGroupSpec Validation Tests")
}

func minimalConsumerGroup() *AzureEventHubConsumerGroup {
	return &AzureEventHubConsumerGroup{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureEventHubConsumerGroup",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-consumer-group",
		},
		Spec: &AzureEventHubConsumerGroupSpec{
			EventHubId: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "/subscriptions/s/resourceGroups/rg/providers/Microsoft.EventHub/namespaces/myapp-eh/eventhubs/orders-stream",
				},
			},
			ConsumerGroupName: "analytics-loader",
		},
	}
}

var _ = ginkgo.Describe("AzureEventHubConsumerGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_event_hub_consumer_group", func() {

			ginkgo.It("should accept a minimal consumer group", func() {
				gomega.Expect(protovalidate.Validate(minimalConsumerGroup())).To(gomega.BeNil())
			})

			ginkgo.It("should accept an event hub reference by valueFrom", func() {
				input := minimalConsumerGroup()
				input.Spec.EventHubId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureEventHub,
							Name:      "orders-stream",
							FieldPath: "status.outputs.event_hub_id",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept user metadata", func() {
				meta := "owner=analytics-team purpose=hourly-batch-load"
				input := minimalConsumerGroup()
				input.Spec.UserMetadata = &meta
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a single-character group name", func() {
				input := minimalConsumerGroup()
				input.Spec.ConsumerGroupName = "g"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a 50-character group name", func() {
				input := minimalConsumerGroup()
				input.Spec.ConsumerGroupName = "a" + strings.Repeat("b", 48) + "c"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_event_hub_consumer_group", func() {

			ginkgo.It("should reject a missing event hub reference", func() {
				input := minimalConsumerGroup()
				input.Spec.EventHubId = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing group name", func() {
				input := minimalConsumerGroup()
				input.Spec.ConsumerGroupName = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a group name over 50 characters", func() {
				input := minimalConsumerGroup()
				input.Spec.ConsumerGroupName = strings.Repeat("a", 51)
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a group name with illegal characters", func() {
				input := minimalConsumerGroup()
				input.Spec.ConsumerGroupName = "bad name!"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject the reserved $Default group name", func() {
				input := minimalConsumerGroup()
				input.Spec.ConsumerGroupName = "$Default"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject user metadata over 1024 characters", func() {
				meta := strings.Repeat("a", 1025)
				input := minimalConsumerGroup()
				input.Spec.UserMetadata = &meta
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing metadata block", func() {
				input := minimalConsumerGroup()
				input.Metadata = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an incorrect kind", func() {
				input := minimalConsumerGroup()
				input.Kind = "WrongKind"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
