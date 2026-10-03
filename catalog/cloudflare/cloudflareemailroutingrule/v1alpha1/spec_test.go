package cloudflareemailroutingrulev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func value(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func literalMatcher(addr string) *CloudflareEmailRoutingRuleMatcher {
	return &CloudflareEmailRoutingRuleMatcher{
		Type:  CloudflareEmailRoutingRuleMatcherType_literal,
		Field: "to",
		Value: addr,
	}
}

func forwardAction(addrs ...string) *CloudflareEmailRoutingRuleAction {
	action := &CloudflareEmailRoutingRuleAction{Type: CloudflareEmailRoutingRuleActionType_forward}
	for _, a := range addrs {
		action.ForwardTo = append(action.ForwardTo, value(a))
	}
	return action
}

func validRule() *CloudflareEmailRoutingRule {
	return &CloudflareEmailRoutingRule{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareEmailRoutingRule",
		Metadata:   &shared.CatalogObjectMetadata{Name: "test-rule"},
		Spec: &CloudflareEmailRoutingRuleSpec{
			ZoneId:   value("023e105f4ecef8ad9ca31a8372d0c353"),
			Matchers: []*CloudflareEmailRoutingRuleMatcher{literalMatcher("support@example.com")},
			Actions:  []*CloudflareEmailRoutingRuleAction{forwardAction("ops@example.com")},
		},
	}
}

func TestCloudflareEmailRoutingRuleSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareEmailRoutingRuleSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareEmailRoutingRuleSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a forward rule with a literal matcher", func() {
			gomega.Expect(protovalidate.Validate(validRule())).To(gomega.BeNil())
		})

		ginkgo.It("accepts an all matcher with a worker action", func() {
			in := validRule()
			in.Spec.Matchers = []*CloudflareEmailRoutingRuleMatcher{{Type: CloudflareEmailRoutingRuleMatcherType_all}}
			in.Spec.Actions = []*CloudflareEmailRoutingRuleAction{{
				Type:   CloudflareEmailRoutingRuleActionType_worker,
				Worker: value("email-router"),
			}}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a drop action", func() {
			in := validRule()
			in.Spec.Actions = []*CloudflareEmailRoutingRuleAction{{Type: CloudflareEmailRoutingRuleActionType_drop}}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts multiple actions on one rule (forward AND worker)", func() {
			in := validRule()
			in.Spec.Actions = []*CloudflareEmailRoutingRuleAction{
				forwardAction("ops@example.com", "audit@example.com"),
				{
					Type:   CloudflareEmailRoutingRuleActionType_worker,
					Worker: value("email-router"),
				},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing zone_id", func() {
			in := validRule()
			in.Spec.ZoneId = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an empty matchers list", func() {
			in := validRule()
			in.Spec.Matchers = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an empty actions list", func() {
			in := validRule()
			in.Spec.Actions = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a forward action without forward_to", func() {
			in := validRule()
			in.Spec.Actions = []*CloudflareEmailRoutingRuleAction{{Type: CloudflareEmailRoutingRuleActionType_forward}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a literal matcher without a value", func() {
			in := validRule()
			in.Spec.Matchers = []*CloudflareEmailRoutingRuleMatcher{{Type: CloudflareEmailRoutingRuleMatcherType_literal, Field: "to"}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an all matcher that sets a value", func() {
			in := validRule()
			in.Spec.Matchers = []*CloudflareEmailRoutingRuleMatcher{{Type: CloudflareEmailRoutingRuleMatcherType_all, Value: "x@example.com"}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
