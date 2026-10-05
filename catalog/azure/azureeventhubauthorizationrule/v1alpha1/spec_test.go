package azureeventhubauthorizationrulev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureEventHubAuthorizationRuleSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventHubAuthorizationRuleSpec Validation Tests")
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func boolPtr(v bool) *bool { return &v }

// helper for a valid namespace-scoped listen rule
func namespaceScopedRule() *AzureEventHubAuthorizationRule {
	return &AzureEventHubAuthorizationRule{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureEventHubAuthorizationRule",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-rule",
		},
		Spec: &AzureEventHubAuthorizationRuleSpec{
			RuleName:    "app-listen",
			NamespaceId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.EventHub/namespaces/my-ehns"),
			Listen:      boolPtr(true),
		},
	}
}

// helper for a valid hub-scoped send rule
func hubScopedRule() *AzureEventHubAuthorizationRule {
	input := namespaceScopedRule()
	input.Spec.NamespaceId = nil
	input.Spec.EventHubId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.EventHub/namespaces/my-ehns/eventhubs/telemetry")
	input.Spec.Listen = nil
	input.Spec.Send = boolPtr(true)
	return input
}

var _ = ginkgo.Describe("AzureEventHubAuthorizationRuleSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_event_hub_authorization_rule", func() {

			ginkgo.It("should accept a namespace-scoped listen rule", func() {
				gomega.Expect(protovalidate.Validate(namespaceScopedRule())).To(gomega.BeNil())
			})

			ginkgo.It("should accept a hub-scoped send rule", func() {
				gomega.Expect(protovalidate.Validate(hubScopedRule())).To(gomega.BeNil())
			})

			ginkgo.It("should accept a manage rule with listen and send", func() {
				input := namespaceScopedRule()
				input.Spec.Listen = boolPtr(true)
				input.Spec.Send = boolPtr(true)
				input.Spec.Manage = boolPtr(true)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a send-and-listen rule", func() {
				input := hubScopedRule()
				input.Spec.Listen = boolPtr(true)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a single-character rule name", func() {
				input := namespaceScopedRule()
				input.Spec.RuleName = "a"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("rule_name", func() {

			ginkgo.It("should reject a missing rule_name", func() {
				input := namespaceScopedRule()
				input.Spec.RuleName = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject the reserved root rule name", func() {
				input := namespaceScopedRule()
				input.Spec.RuleName = "RootManageSharedAccessKey"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a rule name over 60 characters", func() {
				input := namespaceScopedRule()
				input.Spec.RuleName = "a123456789b123456789c123456789d123456789e123456789f1234567890"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject invalid characters", func() {
				input := namespaceScopedRule()
				input.Spec.RuleName = "bad rule!"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("scope", func() {

			ginkgo.It("should reject NO scope set", func() {
				input := namespaceScopedRule()
				input.Spec.NamespaceId = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject BOTH scopes set", func() {
				input := namespaceScopedRule()
				input.Spec.EventHubId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.EventHub/namespaces/my-ehns/eventhubs/telemetry")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("rights", func() {

			ginkgo.It("should reject a rule with no rights", func() {
				input := namespaceScopedRule()
				input.Spec.Listen = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject manage without listen and send", func() {
				input := namespaceScopedRule()
				input.Spec.Listen = nil
				input.Spec.Manage = boolPtr(true)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject manage with listen but not send", func() {
				input := namespaceScopedRule()
				input.Spec.Manage = boolPtr(true)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
