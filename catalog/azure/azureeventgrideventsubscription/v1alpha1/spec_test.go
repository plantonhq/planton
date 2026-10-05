package azureeventgrideventsubscriptionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureEventgridEventSubscriptionSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventgridEventSubscriptionSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func strPtr(s string) *string { return &s }

func boolPtr(b bool) *bool { return &b }

func int32Ptr(i int32) *int32 { return &i }

const (
	testTopicId       = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.EventGrid/topics/orders-events"
	testSystemTopicId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.EventGrid/systemTopics/appdata-events"
	testStorageId     = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Storage/storageAccounts/appdata"
	testQueueDest     = "work-items"
	testIdentityId    = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-uai"
)

// storageQueueDestination returns the canonical test destination.
func storageQueueDestination() *AzureEventgridEventSubscriptionDestination {
	return &AzureEventgridEventSubscriptionDestination{
		StorageQueue: &AzureEventgridEventSubscriptionStorageQueueDestination{
			StorageAccountId: literal(testStorageId),
			QueueName:        testQueueDest,
		},
	}
}

// validResource returns a valid scope-addressed subscription that
// individual cases mutate into the shape under test.
func validResource() *AzureEventgridEventSubscription {
	return &AzureEventgridEventSubscription{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureEventgridEventSubscription",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-eges",
		},
		Spec: &AzureEventgridEventSubscriptionSpec{
			Scope:       literal(testTopicId),
			Name:        "orders-to-queue",
			Destination: storageQueueDestination(),
		},
	}
}

var _ = ginkgo.Describe("AzureEventgridEventSubscriptionSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_eventgrid_event_subscription", func() {

			ginkgo.It("should not return a validation error for the minimal scope-addressed shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a system-topic-addressed subscription", func() {
				input := validResource()
				input.Spec.Scope = nil
				input.Spec.SystemTopicId = literal(testSystemTopicId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept each id-arm destination", func() {
				for _, dest := range []*AzureEventgridEventSubscriptionDestination{
					{EventhubId: literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.EventHub/namespaces/ns/eventhubs/hub")},
					{HybridConnectionId: literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Relay/namespaces/ns/hybridConnections/hc")},
					{ServiceBusQueueId: literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ServiceBus/namespaces/ns/queues/q")},
					{ServiceBusTopicId: literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ServiceBus/namespaces/ns/topics/t")},
				} {
					input := validResource()
					input.Spec.Destination = dest
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept an azure function destination with batch tuning", func() {
				input := validResource()
				input.Spec.Destination = &AzureEventgridEventSubscriptionDestination{
					AzureFunction: &AzureEventgridEventSubscriptionAzureFunctionDestination{
						FunctionId:                    literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Web/sites/app/functions/handler"),
						MaxEventsPerBatch:             int32Ptr(10),
						PreferredBatchSizeInKilobytes: int32Ptr(64),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a webhook destination with Entra fields and batch bounds", func() {
				input := validResource()
				input.Spec.Destination = &AzureEventgridEventSubscriptionDestination{
					Webhook: &AzureEventgridEventSubscriptionWebhookDestination{
						Url:                           "https://handler.example.com/events",
						MaxEventsPerBatch:             int32Ptr(5000),
						PreferredBatchSizeInKilobytes: int32Ptr(1024),
						ActiveDirectoryTenantId:       "00000000-0000-0000-0000-000000000000",
						ActiveDirectoryAppIdOrUri:     "api://handler",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a storage queue destination with a message TTL", func() {
				input := validResource()
				input.Spec.Destination.StorageQueue.QueueMessageTimeToLiveInSeconds = int32Ptr(-1)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept each delivery schema token", func() {
				for _, schema := range []string{"EventGridSchema", "CloudEventSchemaV1_0", "CustomInputSchema"} {
					input := validResource()
					input.Spec.EventDeliverySchema = strPtr(schema)
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept a system-assigned delivery identity", func() {
				input := validResource()
				input.Spec.DeliveryIdentity = &AzureEventgridEventSubscriptionIdentity{
					Type: AzureEventgridEventSubscriptionIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned delivery identity carrying an identity", func() {
				input := validResource()
				input.Spec.DeliveryIdentity = &AzureEventgridEventSubscriptionIdentity{
					Type:                 AzureEventgridEventSubscriptionIdentityType_USER_ASSIGNED,
					UserAssignedIdentity: literal(testIdentityId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept static and dynamic delivery properties", func() {
				input := validResource()
				input.Spec.Destination = &AzureEventgridEventSubscriptionDestination{
					Webhook: &AzureEventgridEventSubscriptionWebhookDestination{Url: "https://handler.example.com/events"},
				}
				input.Spec.DeliveryProperties = []*AzureEventgridEventSubscriptionDeliveryProperty{
					{HeaderName: "x-api-key", Type: "Static", Value: literal("s3cret"), Secret: true},
					{HeaderName: "x-source-system", Type: "Dynamic", SourceField: "data.system"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a dead letter with a matching identity", func() {
				input := validResource()
				input.Spec.DeadLetter = &AzureEventgridEventSubscriptionDeadLetter{
					StorageAccountId:         literal(testStorageId),
					StorageBlobContainerName: "dead-letters",
				}
				input.Spec.DeadLetterIdentity = &AzureEventgridEventSubscriptionIdentity{
					Type: AzureEventgridEventSubscriptionIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept subject and advanced filters", func() {
				input := validResource()
				input.Spec.SubjectFilter = &AzureEventgridEventSubscriptionSubjectFilter{
					SubjectBeginsWith: "/blobServices/default/containers/invoices/",
					SubjectEndsWith:   ".pdf",
					CaseSensitive:     boolPtr(false),
				}
				input.Spec.AdvancedFilter = &AzureEventgridEventSubscriptionAdvancedFilter{
					StringIn: []*AzureEventgridEventSubscriptionStringListFilter{
						{Key: "data.api", Values: []string{"PutBlob", "PutBlockList"}},
					},
					NumberGreaterThan: []*AzureEventgridEventSubscriptionNumberFilter{
						{Key: "data.contentLength", Value: 0},
					},
					IsNotNull: []*AzureEventgridEventSubscriptionKeyFilter{
						{Key: "data.contentType"},
					},
					NumberInRange: []*AzureEventgridEventSubscriptionNumberRangeFilter{
						{Key: "data.contentLength", Ranges: []*AzureEventgridEventSubscriptionNumberRange{{From: 1, To: 1048576}}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept event types, labels, expiration, and a retry policy", func() {
				input := validResource()
				input.Spec.IncludedEventTypes = []string{"Microsoft.Storage.BlobCreated"}
				input.Spec.Labels = []string{"orders", "critical"}
				input.Spec.ExpirationTimeUtc = strPtr("2027-01-01T00:00:00Z")
				input.Spec.RetryPolicy = &AzureEventgridEventSubscriptionRetryPolicy{
					MaxDeliveryAttempts: 10,
					EventTimeToLive:     240,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_eventgrid_event_subscription", func() {

			ginkgo.It("should reject a subscription with neither scope nor system_topic_id", func() {
				input := validResource()
				input.Spec.Scope = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a subscription with both scope and system_topic_id", func() {
				input := validResource()
				input.Spec.SystemTopicId = literal(testSystemTopicId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a 2-character name", func() {
				input := validResource()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name with invalid characters", func() {
				input := validResource()
				input.Spec.Name = "orders.to.queue"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing destination", func() {
				input := validResource()
				input.Spec.Destination = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty destination block", func() {
				input := validResource()
				input.Spec.Destination = &AzureEventgridEventSubscriptionDestination{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a destination with two arms", func() {
				input := validResource()
				input.Spec.Destination.EventhubId = literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.EventHub/namespaces/ns/eventhubs/hub")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a storage queue destination without a queue name", func() {
				input := validResource()
				input.Spec.Destination.StorageQueue.QueueName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an http webhook url", func() {
				input := validResource()
				input.Spec.Destination = &AzureEventgridEventSubscriptionDestination{
					Webhook: &AzureEventgridEventSubscriptionWebhookDestination{Url: "http://handler.example.com/events"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject webhook batch values outside the provider bounds", func() {
				input := validResource()
				input.Spec.Destination = &AzureEventgridEventSubscriptionDestination{
					Webhook: &AzureEventgridEventSubscriptionWebhookDestination{
						Url:               "https://handler.example.com/events",
						MaxEventsPerBatch: int32Ptr(5001),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown delivery schema", func() {
				input := validResource()
				input.Spec.EventDeliverySchema = strPtr("CustomEventSchema")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without an identity", func() {
				input := validResource()
				input.Spec.DeliveryIdentity = &AzureEventgridEventSubscriptionIdentity{
					Type: AzureEventgridEventSubscriptionIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying an identity", func() {
				input := validResource()
				input.Spec.DeliveryIdentity = &AzureEventgridEventSubscriptionIdentity{
					Type:                 AzureEventgridEventSubscriptionIdentityType_SYSTEM_ASSIGNED,
					UserAssignedIdentity: literal(testIdentityId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a static delivery property without a value", func() {
				input := validResource()
				input.Spec.DeliveryProperties = []*AzureEventgridEventSubscriptionDeliveryProperty{
					{HeaderName: "x-api-key", Type: "Static"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a dynamic delivery property without a source field", func() {
				input := validResource()
				input.Spec.DeliveryProperties = []*AzureEventgridEventSubscriptionDeliveryProperty{
					{HeaderName: "x-source-system", Type: "Dynamic"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a dynamic delivery property marked secret", func() {
				input := validResource()
				input.Spec.DeliveryProperties = []*AzureEventgridEventSubscriptionDeliveryProperty{
					{HeaderName: "x-source-system", Type: "Dynamic", SourceField: "data.system", Secret: true},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a dead-letter identity without a dead letter", func() {
				input := validResource()
				input.Spec.DeadLetterIdentity = &AzureEventgridEventSubscriptionIdentity{
					Type: AzureEventgridEventSubscriptionIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty subject filter", func() {
				input := validResource()
				input.Spec.SubjectFilter = &AzureEventgridEventSubscriptionSubjectFilter{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty advanced filter", func() {
				input := validResource()
				input.Spec.AdvancedFilter = &AzureEventgridEventSubscriptionAdvancedFilter{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a list condition with empty values", func() {
				input := validResource()
				input.Spec.AdvancedFilter = &AzureEventgridEventSubscriptionAdvancedFilter{
					StringIn: []*AzureEventgridEventSubscriptionStringListFilter{
						{Key: "data.api"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a list condition with more than 25 values", func() {
				values := make([]string, 26)
				for i := range values {
					values[i] = "v"
				}
				input := validResource()
				input.Spec.AdvancedFilter = &AzureEventgridEventSubscriptionAdvancedFilter{
					StringIn: []*AzureEventgridEventSubscriptionStringListFilter{
						{Key: "data.api", Values: values},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a condition without a key", func() {
				input := validResource()
				input.Spec.AdvancedFilter = &AzureEventgridEventSubscriptionAdvancedFilter{
					IsNotNull: []*AzureEventgridEventSubscriptionKeyFilter{{}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a malformed expiration timestamp", func() {
				input := validResource()
				input.Spec.ExpirationTimeUtc = strPtr("2027-01-01 00:00:00")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject retry attempts outside 1-30", func() {
				input := validResource()
				input.Spec.RetryPolicy = &AzureEventgridEventSubscriptionRetryPolicy{
					MaxDeliveryAttempts: 31,
					EventTimeToLive:     240,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a retry time-to-live outside 1-1440", func() {
				input := validResource()
				input.Spec.RetryPolicy = &AzureEventgridEventSubscriptionRetryPolicy{
					MaxDeliveryAttempts: 10,
					EventTimeToLive:     1441,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
