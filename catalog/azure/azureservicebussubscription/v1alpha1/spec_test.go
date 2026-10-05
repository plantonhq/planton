package azureservicebussubscriptionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureServiceBusSubscriptionSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureServiceBusSubscriptionSpec Validation Tests")
}

func minimalSubscription() *AzureServiceBusSubscription {
	return &AzureServiceBusSubscription{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureServiceBusSubscription",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-subscription",
		},
		Spec: &AzureServiceBusSubscriptionSpec{
			TopicId: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ServiceBus/namespaces/myapp-bus/topics/events",
				},
			},
			SubscriptionName: "audit-consumer",
			MaxDeliveryCount: 10,
		},
	}
}

func sqlRule(name, expr string) *AzureServiceBusSubscriptionRule {
	return &AzureServiceBusSubscriptionRule{
		RuleName:   name,
		FilterType: AzureServiceBusFilterType_SQL_FILTER,
		SqlFilter:  &expr,
	}
}

var _ = ginkgo.Describe("AzureServiceBusSubscriptionSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_service_bus_subscription", func() {

			ginkgo.It("should accept a minimal subscription", func() {
				gomega.Expect(protovalidate.Validate(minimalSubscription())).To(gomega.BeNil())
			})

			ginkgo.It("should accept a topic reference by valueFrom", func() {
				input := minimalSubscription()
				input.Spec.TopicId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureServiceBusTopic,
							Name:      "events-topic",
							FieldPath: "status.outputs.topic_id",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a fully-dialed subscription", func() {
				lock := "PT2M"
				ttl := "P7D"
				idle := "PT10M"
				dlExpire := true
				dlFilter := false
				session := true
				batched := true
				forwardDlq := &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "poison-sink"},
				}
				input := minimalSubscription()
				input.Spec.LockDuration = &lock
				input.Spec.DefaultMessageTtl = &ttl
				input.Spec.AutoDeleteOnIdle = &idle
				input.Spec.DeadLetteringOnMessageExpiration = &dlExpire
				input.Spec.DeadLetteringOnFilterEvaluationError = &dlFilter
				input.Spec.RequiresSession = &session
				input.Spec.BatchedOperationsEnabled = &batched
				input.Spec.ForwardDeadLetteredMessagesTo = forwardDlq
				input.Spec.Status = AzureServiceBusSubscriptionStatusValue_ACTIVE
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a subscription forwarding to a queue", func() {
				input := minimalSubscription()
				input.Spec.ForwardTo = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureServiceBusQueue,
							Name:      "work-queue",
							FieldPath: "status.outputs.queue_name",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept every declared gate state", func() {
				for _, s := range []AzureServiceBusSubscriptionStatusValue{
					AzureServiceBusSubscriptionStatusValue_ACTIVE,
					AzureServiceBusSubscriptionStatusValue_DISABLED,
					AzureServiceBusSubscriptionStatusValue_RECEIVE_DISABLED,
				} {
					input := minimalSubscription()
					input.Spec.Status = s
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept a client-scoped subscription", func() {
				shareable := false
				input := minimalSubscription()
				input.Spec.ClientScopedSubscription = &AzureServiceBusClientScopedSubscription{
					ClientId:  "jms-client-1",
					Shareable: &shareable,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a SQL filter rule", func() {
				input := minimalSubscription()
				input.Spec.Rules = []*AzureServiceBusSubscriptionRule{
					sqlRule("important-only", "sys.Label = 'important' AND quantity > 10"),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept multiple rules of both families", func() {
				label := "order-created"
				input := minimalSubscription()
				input.Spec.Rules = []*AzureServiceBusSubscriptionRule{
					sqlRule("emea-orders", "region = 'emea'"),
					{
						RuleName:   "order-created",
						FilterType: AzureServiceBusFilterType_CORRELATION_FILTER,
						CorrelationFilter: &AzureServiceBusCorrelationFilter{
							Label: &label,
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a SQL rule with an action", func() {
				action := "SET sys.Label = 'routed'"
				rule := sqlRule("route-and-tag", "priority > 3")
				rule.Action = &action
				input := minimalSubscription()
				input.Spec.Rules = []*AzureServiceBusSubscriptionRule{rule}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a correlation filter rule", func() {
				label := "order-created"
				input := minimalSubscription()
				input.Spec.Rules = []*AzureServiceBusSubscriptionRule{
					{
						RuleName:   "order-events",
						FilterType: AzureServiceBusFilterType_CORRELATION_FILTER,
						CorrelationFilter: &AzureServiceBusCorrelationFilter{
							Label: &label,
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a correlation filter matching only user properties", func() {
				input := minimalSubscription()
				input.Spec.Rules = []*AzureServiceBusSubscriptionRule{
					{
						RuleName:   "tenant-router",
						FilterType: AzureServiceBusFilterType_CORRELATION_FILTER,
						CorrelationFilter: &AzureServiceBusCorrelationFilter{
							Properties: map[string]string{"tenant": "contoso"},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_service_bus_subscription", func() {

			ginkgo.It("should reject a missing topic reference", func() {
				input := minimalSubscription()
				input.Spec.TopicId = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing subscription name", func() {
				input := minimalSubscription()
				input.Spec.SubscriptionName = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a subscription name over 50 characters", func() {
				input := minimalSubscription()
				input.Spec.SubscriptionName = "a-very-long-subscription-name-that-exceeds-the-cap-x"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a subscription name with slashes", func() {
				input := minimalSubscription()
				input.Spec.SubscriptionName = "audit/consumer"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing max_delivery_count", func() {
				input := minimalSubscription()
				input.Spec.MaxDeliveryCount = 0
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an undeclared gate state", func() {
				input := minimalSubscription()
				input.Spec.Status = AzureServiceBusSubscriptionStatusValue(99)
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a rule without a name", func() {
				input := minimalSubscription()
				input.Spec.Rules = []*AzureServiceBusSubscriptionRule{
					sqlRule("", "1=1"),
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject the reserved $Default rule name", func() {
				input := minimalSubscription()
				input.Spec.Rules = []*AzureServiceBusSubscriptionRule{
					sqlRule("$Default", "region = 'emea'"),
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a rule without a filter type", func() {
				expr := "1=1"
				input := minimalSubscription()
				input.Spec.Rules = []*AzureServiceBusSubscriptionRule{
					{RuleName: "untyped", SqlFilter: &expr},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a SQL rule without its expression", func() {
				input := minimalSubscription()
				input.Spec.Rules = []*AzureServiceBusSubscriptionRule{
					{RuleName: "no-expr", FilterType: AzureServiceBusFilterType_SQL_FILTER},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a SQL rule carrying a correlation filter", func() {
				label := "x"
				rule := sqlRule("mixed", "1=1")
				rule.CorrelationFilter = &AzureServiceBusCorrelationFilter{Label: &label}
				input := minimalSubscription()
				input.Spec.Rules = []*AzureServiceBusSubscriptionRule{rule}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a correlation rule without its filter block", func() {
				input := minimalSubscription()
				input.Spec.Rules = []*AzureServiceBusSubscriptionRule{
					{RuleName: "no-block", FilterType: AzureServiceBusFilterType_CORRELATION_FILTER},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a correlation rule carrying a sql_filter", func() {
				expr := "1=1"
				label := "x"
				input := minimalSubscription()
				input.Spec.Rules = []*AzureServiceBusSubscriptionRule{
					{
						RuleName:          "mixed",
						FilterType:        AzureServiceBusFilterType_CORRELATION_FILTER,
						SqlFilter:         &expr,
						CorrelationFilter: &AzureServiceBusCorrelationFilter{Label: &label},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an empty correlation filter", func() {
				input := minimalSubscription()
				input.Spec.Rules = []*AzureServiceBusSubscriptionRule{
					{
						RuleName:          "empty-matcher",
						FilterType:        AzureServiceBusFilterType_CORRELATION_FILTER,
						CorrelationFilter: &AzureServiceBusCorrelationFilter{},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a SQL filter over 1024 characters", func() {
				long := make([]byte, 1025)
				for i := range long {
					long[i] = 'x'
				}
				input := minimalSubscription()
				input.Spec.Rules = []*AzureServiceBusSubscriptionRule{
					sqlRule("too-long", string(long)),
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing metadata block", func() {
				input := minimalSubscription()
				input.Metadata = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an incorrect kind", func() {
				input := minimalSubscription()
				input.Kind = "WrongKind"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
