package azureservicebusnamespacev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAzureServiceBusNamespaceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureServiceBusNamespaceSpec Validation Tests")
}

// helper to create a minimal valid spec (Standard tier, one queue)
func minimalSpec() *AzureServiceBusNamespace {
	return &AzureServiceBusNamespace{
		ApiVersion: "azure.openmcf.org/v1",
		Kind:       "AzureServiceBusNamespace",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-sb",
		},
		Spec: &AzureServiceBusNamespaceSpec{
			Region: "eastus",
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-rg",
				},
			},
			Name: "myapp-servicebus",
		},
	}
}

var _ = ginkgo.Describe("AzureServiceBusNamespaceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_service_bus_namespace", func() {

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
				capacity := int32(2)
				input := minimalSpec()
				input.Spec.Sku = &sku
				input.Spec.Capacity = &capacity
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for Premium with partitions", func() {
				sku := "Premium"
				partitions := int32(4)
				input := minimalSpec()
				input.Spec.Sku = &sku
				input.Spec.PremiumMessagingPartitions = &partitions
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

			ginkgo.It("should not return a validation error for a namespace with queues", func() {
				input := minimalSpec()
				input.Spec.Queues = []*AzureServiceBusQueue{
					{Name: "orders"},
					{Name: "notifications"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a namespace with topics", func() {
				input := minimalSpec()
				input.Spec.Topics = []*AzureServiceBusTopic{
					{Name: "events"},
					{Name: "audit-logs"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a queue with all optional fields", func() {
				ttl := "P14D"
				lockDuration := "PT5M"
				maxDelivery := int32(5)
				maxSize := int32(5120)
				partitioning := false
				dupDetection := true
				session := true
				deadLettering := true
				forwardTo := "dead-letter-sink"
				forwardDlq := "dlq-monitor"
				input := minimalSpec()
				input.Spec.Queues = []*AzureServiceBusQueue{
					{
						Name:                               "full-featured-queue",
						MaxSizeInMegabytes:                 &maxSize,
						PartitioningEnabled:                &partitioning,
						DefaultMessageTtl:                  &ttl,
						LockDuration:                       &lockDuration,
						MaxDeliveryCount:                   &maxDelivery,
						RequiresDuplicateDetection:         &dupDetection,
						RequiresSession:                    &session,
						DeadLetteringOnMessageExpiration:   &deadLettering,
						ForwardTo:                          &forwardTo,
						ForwardDeadLetteredMessagesTo:      &forwardDlq,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a topic with all optional fields", func() {
				ttl := "P30D"
				maxSize := int32(4096)
				partitioning := true
				dupDetection := true
				ordering := true
				input := minimalSpec()
				input.Spec.Topics = []*AzureServiceBusTopic{
					{
						Name:                       "full-featured-topic",
						MaxSizeInMegabytes:         &maxSize,
						PartitioningEnabled:        &partitioning,
						DefaultMessageTtl:          &ttl,
						RequiresDuplicateDetection: &dupDetection,
						SupportOrdering:            &ordering,
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

			ginkgo.It("should not return a validation error with all optional fields set", func() {
				sku := "Premium"
				capacity := int32(4)
				partitions := int32(2)
				zoneRedundant := true
				tls := "1.2"
				publicAccess := false
				input := minimalSpec()
				input.Metadata.Org = "mycompany"
				input.Metadata.Env = "production"
				input.Spec.Sku = &sku
				input.Spec.Capacity = &capacity
				input.Spec.PremiumMessagingPartitions = &partitions
				input.Spec.ZoneRedundant = &zoneRedundant
				input.Spec.MinimumTlsVersion = &tls
				input.Spec.PublicNetworkAccessEnabled = &publicAccess
				input.Spec.Queues = []*AzureServiceBusQueue{
					{Name: "orders"},
				}
				input.Spec.Topics = []*AzureServiceBusTopic{
					{Name: "events"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a queue with max_delivery_count of 1", func() {
				maxDelivery := int32(1)
				input := minimalSpec()
				input.Spec.Queues = []*AzureServiceBusQueue{
					{Name: "fast-fail-queue", MaxDeliveryCount: &maxDelivery},
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
				// 50 chars: starts with letter, ends with number, contains hyphens
				input := minimalSpec()
				input.Spec.Name = "a-very-long-service-bus-namespace-name-for-test-01" // 50 chars
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_service_bus_namespace", func() {

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

			ginkgo.It("should return a validation error when capacity exceeds maximum (16)", func() {
				capacity := int32(17)
				input := minimalSpec()
				input.Spec.Capacity = &capacity
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when capacity is zero", func() {
				capacity := int32(0)
				input := minimalSpec()
				input.Spec.Capacity = &capacity
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when premium_messaging_partitions exceeds 4", func() {
				partitions := int32(5)
				input := minimalSpec()
				input.Spec.PremiumMessagingPartitions = &partitions
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when premium_messaging_partitions is zero", func() {
				partitions := int32(0)
				input := minimalSpec()
				input.Spec.PremiumMessagingPartitions = &partitions
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when queue name is empty", func() {
				input := minimalSpec()
				input.Spec.Queues = []*AzureServiceBusQueue{
					{Name: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when queue max_delivery_count is zero", func() {
				maxDelivery := int32(0)
				input := minimalSpec()
				input.Spec.Queues = []*AzureServiceBusQueue{
					{Name: "bad-queue", MaxDeliveryCount: &maxDelivery},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when topic name is empty", func() {
				input := minimalSpec()
				input.Spec.Topics = []*AzureServiceBusTopic{
					{Name: ""},
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
				input := &AzureServiceBusNamespace{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureServiceBusNamespace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-sb",
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
