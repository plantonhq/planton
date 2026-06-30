package azureeventhubnamespacev1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAzureEventHubNamespaceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventHubNamespaceSpec Validation Tests")
}

// helper to create a minimal valid spec (Standard tier, no event hubs)
func minimalSpec() *AzureEventHubNamespace {
	return &AzureEventHubNamespace{
		ApiVersion: "azure.planton.dev/v1",
		Kind:       "AzureEventHubNamespace",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-eh",
		},
		Spec: &AzureEventHubNamespaceSpec{
			Region: "eastus",
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-rg",
				},
			},
			Name: "myapp-eventhubs",
		},
	}
}

var _ = ginkgo.Describe("AzureEventHubNamespaceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_event_hub_namespace", func() {

			ginkgo.It("should not return a validation error for a minimal Standard namespace", func() {
				input := minimalSpec()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for each valid SKU", func() {
				skus := []string{"Basic", "Standard", "Premium"}
				for _, s := range skus {
					sku := s
					input := minimalSpec()
					input.Spec.Sku = &sku
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error for Premium with capacity", func() {
				sku := "Premium"
				capacity := int32(4)
				input := minimalSpec()
				input.Spec.Sku = &sku
				input.Spec.Capacity = &capacity
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for Premium with zone redundancy", func() {
				sku := "Premium"
				capacity := int32(1)
				zoneRedundant := true
				input := minimalSpec()
				input.Spec.Sku = &sku
				input.Spec.Capacity = &capacity
				input.Spec.ZoneRedundant = &zoneRedundant
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for each valid TLS version", func() {
				versions := []string{"1.0", "1.1", "1.2"}
				for _, v := range versions {
					ver := v
					input := minimalSpec()
					input.Spec.MinimumTlsVersion = &ver
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error for a namespace with event hubs", func() {
				retention := int32(3)
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{Name: "telemetry", PartitionCount: 4, MessageRetention: &retention},
					{Name: "audit-logs", PartitionCount: 8},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an event hub with consumer groups", func() {
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{
						Name:           "events",
						PartitionCount: 4,
						ConsumerGroups: []*AzureEventHubConsumerGroup{
							{Name: "analytics-consumer"},
							{Name: "archiver-consumer"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an event hub with all optional fields", func() {
				retention := int32(7)
				userMeta := "analytics pipeline consumer"
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{
						Name:             "full-featured-hub",
						PartitionCount:   16,
						MessageRetention: &retention,
						ConsumerGroups: []*AzureEventHubConsumerGroup{
							{Name: "pipeline-consumer", UserMetadata: &userMeta},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a consumer group with user_metadata", func() {
				userMeta := "Owned by data-engineering team for real-time analytics"
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{
						Name:           "metrics",
						PartitionCount: 2,
						ConsumerGroups: []*AzureEventHubConsumerGroup{
							{Name: "data-eng", UserMetadata: &userMeta},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom reference for resource_group", func() {
				input := minimalSpec()
				input.Spec.ResourceGroup = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureResourceGroup,
							Name:      "shared-rg",
							FieldPath: "status.outputs.resource_group_name",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with public_network_access disabled", func() {
				publicAccess := false
				input := minimalSpec()
				input.Spec.PublicNetworkAccessEnabled = &publicAccess
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with auto-inflate enabled and maximum_throughput_units", func() {
				autoInflate := true
				maxTU := int32(20)
				input := minimalSpec()
				input.Spec.AutoInflateEnabled = &autoInflate
				input.Spec.MaximumThroughputUnits = &maxTU
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional fields set", func() {
				sku := "Premium"
				capacity := int32(4)
				zoneRedundant := true
				tls := "1.2"
				publicAccess := false
				autoInflate := false
				retention := int32(7)
				userMeta := "team-data-platform"
				input := minimalSpec()
				input.Metadata.Org = "mycompany"
				input.Metadata.Env = "production"
				input.Spec.Sku = &sku
				input.Spec.Capacity = &capacity
				input.Spec.ZoneRedundant = &zoneRedundant
				input.Spec.MinimumTlsVersion = &tls
				input.Spec.PublicNetworkAccessEnabled = &publicAccess
				input.Spec.AutoInflateEnabled = &autoInflate
				input.Spec.EventHubs = []*AzureEventHub{
					{
						Name:             "orders",
						PartitionCount:   32,
						MessageRetention: &retention,
						ConsumerGroups: []*AzureEventHubConsumerGroup{
							{Name: "order-processor", UserMetadata: &userMeta},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for namespace name at min length (6)", func() {
				input := minimalSpec()
				input.Spec.Name = "abcde1" // 6 chars: starts letter, ends number
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for namespace name at max length (50)", func() {
				input := minimalSpec()
				// 50 chars: starts with letter, ends with number, contains hyphens
				input.Spec.Name = "a-very-long-eventhub-namespace-name-for-testing-01" // 50 chars
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for event hub name at max length (256)", func() {
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{Name: strings.Repeat("a", 256), PartitionCount: 1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for consumer group name at max length (50)", func() {
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{
						Name:           "hub",
						PartitionCount: 1,
						ConsumerGroups: []*AzureEventHubConsumerGroup{
							{Name: strings.Repeat("c", 50)},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for partition count at min boundary (1)", func() {
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{Name: "min-partitions", PartitionCount: 1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for partition count at max boundary (32)", func() {
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{Name: "max-partitions", PartitionCount: 32},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for message retention at min boundary (1)", func() {
				retention := int32(1)
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{Name: "min-retention", PartitionCount: 2, MessageRetention: &retention},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for message retention at max boundary (7)", func() {
				retention := int32(7)
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{Name: "max-retention", PartitionCount: 2, MessageRetention: &retention},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for maximum throughput units at boundary (40)", func() {
				autoInflate := true
				maxTU := int32(40)
				input := minimalSpec()
				input.Spec.AutoInflateEnabled = &autoInflate
				input.Spec.MaximumThroughputUnits = &maxTU
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_event_hub_namespace", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := minimalSpec()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := minimalSpec()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := minimalSpec()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is too short (5 chars)", func() {
				input := minimalSpec()
				input.Spec.Name = "abcde" // 5 chars, needs 6
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds 50 characters", func() {
				input := minimalSpec()
				input.Spec.Name = "a" + string(make([]byte, 50)) // > 50 chars
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name starts with a number", func() {
				input := minimalSpec()
				input.Spec.Name = "1-invalid-name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name ends with a hyphen", func() {
				input := minimalSpec()
				input.Spec.Name = "invalid-name-"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name starts with a hyphen", func() {
				input := minimalSpec()
				input.Spec.Name = "-invalid-name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when sku is invalid", func() {
				invalidSku := "Enterprise"
				input := minimalSpec()
				input.Spec.Sku = &invalidSku
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when minimum_tls_version is invalid", func() {
				invalidTls := "1.3"
				input := minimalSpec()
				input.Spec.MinimumTlsVersion = &invalidTls
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when partition count is zero", func() {
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{Name: "bad-hub", PartitionCount: 0},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when partition count exceeds maximum (32)", func() {
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{Name: "bad-hub", PartitionCount: 33},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when message retention is zero", func() {
				retention := int32(0)
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{Name: "bad-hub", PartitionCount: 2, MessageRetention: &retention},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when message retention exceeds maximum (7)", func() {
				retention := int32(8)
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{Name: "bad-hub", PartitionCount: 2, MessageRetention: &retention},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when maximum throughput units exceeds maximum (40)", func() {
				maxTU := int32(41)
				input := minimalSpec()
				input.Spec.MaximumThroughputUnits = &maxTU
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when maximum throughput units is negative", func() {
				maxTU := int32(-1)
				input := minimalSpec()
				input.Spec.MaximumThroughputUnits = &maxTU
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when event hub name is empty", func() {
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{Name: "", PartitionCount: 2},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when consumer group name is empty", func() {
				input := minimalSpec()
				input.Spec.EventHubs = []*AzureEventHub{
					{
						Name:           "hub",
						PartitionCount: 2,
						ConsumerGroups: []*AzureEventHubConsumerGroup{
							{Name: ""},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalSpec()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureEventHubNamespace{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureEventHubNamespace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-eh",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := minimalSpec()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := minimalSpec()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
