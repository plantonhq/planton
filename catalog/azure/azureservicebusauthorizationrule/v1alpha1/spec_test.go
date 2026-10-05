package azureservicebusauthorizationrulev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureServiceBusAuthorizationRuleSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureServiceBusAuthorizationRuleSpec Validation Tests")
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

const nsID = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ServiceBus/namespaces/myapp-bus"

// helper: a listen-only rule scoped to the namespace
func minimalRule() *AzureServiceBusAuthorizationRule {
	listen := true
	return &AzureServiceBusAuthorizationRule{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureServiceBusAuthorizationRule",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-auth-rule",
		},
		Spec: &AzureServiceBusAuthorizationRuleSpec{
			RuleName:    "app-listener",
			NamespaceId: literal(nsID),
			Listen:      &listen,
		},
	}
}

var _ = ginkgo.Describe("AzureServiceBusAuthorizationRuleSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_service_bus_authorization_rule", func() {

			ginkgo.It("should accept a namespace-scoped listen rule", func() {
				gomega.Expect(protovalidate.Validate(minimalRule())).To(gomega.BeNil())
			})

			ginkgo.It("should accept a queue-scoped send rule", func() {
				send := true
				input := minimalRule()
				input.Spec.NamespaceId = nil
				input.Spec.QueueId = literal(nsID + "/queues/orders")
				input.Spec.Listen = nil
				input.Spec.Send = &send
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a topic-scoped rule by valueFrom", func() {
				send := true
				input := minimalRule()
				input.Spec.NamespaceId = nil
				input.Spec.Listen = nil
				input.Spec.Send = &send
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

			ginkgo.It("should accept a full manage rule with listen and send", func() {
				listen := true
				send := true
				manage := true
				input := minimalRule()
				input.Spec.Listen = &listen
				input.Spec.Send = &send
				input.Spec.Manage = &manage
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept listen and send without manage", func() {
				listen := true
				send := true
				input := minimalRule()
				input.Spec.Listen = &listen
				input.Spec.Send = &send
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a single-character rule name", func() {
				input := minimalRule()
				input.Spec.RuleName = "r"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_service_bus_authorization_rule", func() {

			ginkgo.It("should reject a missing rule name", func() {
				input := minimalRule()
				input.Spec.RuleName = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject the reserved root rule name", func() {
				input := minimalRule()
				input.Spec.RuleName = "RootManageSharedAccessKey"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a rule name with illegal characters", func() {
				input := minimalRule()
				input.Spec.RuleName = "app listener"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a rule name over 50 characters", func() {
				input := minimalRule()
				input.Spec.RuleName = "a-very-long-authorization-rule-name-exceeding-caps"
				input.Spec.RuleName += "x"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject zero scopes", func() {
				input := minimalRule()
				input.Spec.NamespaceId = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject two scopes at once", func() {
				input := minimalRule()
				input.Spec.QueueId = literal(nsID + "/queues/orders")
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject all three scopes at once", func() {
				input := minimalRule()
				input.Spec.QueueId = literal(nsID + "/queues/orders")
				input.Spec.TopicId = literal(nsID + "/topics/events")
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a rule with no rights", func() {
				input := minimalRule()
				input.Spec.Listen = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject manage without listen and send", func() {
				manage := true
				input := minimalRule()
				input.Spec.Listen = nil
				input.Spec.Manage = &manage
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject manage with listen but no send", func() {
				listen := true
				manage := true
				input := minimalRule()
				input.Spec.Listen = &listen
				input.Spec.Manage = &manage
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing metadata block", func() {
				input := minimalRule()
				input.Metadata = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an incorrect kind", func() {
				input := minimalRule()
				input.Kind = "WrongKind"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
