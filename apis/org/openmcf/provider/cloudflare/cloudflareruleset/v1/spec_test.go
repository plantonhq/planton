package cloudflarerulesetv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func boolPtr(b bool) *bool { return &b }
func rulesetKindPtr(k CloudflareRulesetSpec_RulesetKind) *CloudflareRulesetSpec_RulesetKind {
	return &k
}

func TestCloudflareRulesetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareRulesetSpec Validation Suite")
}

func validResource() *CloudflareRuleset {
	return &CloudflareRuleset{
		ApiVersion: "cloudflare.openmcf.org/v1",
		Kind:       "CloudflareRuleset",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-origin-rule"},
		Spec: &CloudflareRulesetSpec{
			ZoneId:      &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "zone-abc123"}},
			RulesetKind: rulesetKindPtr(CloudflareRulesetSpec_zone),
			Phase:       CloudflareRulesetSpec_http_request_origin,
			Name:        "Route app traffic to K8s",
			Rules: []*CloudflareRulesetRule{
				{
					Expression: "not http.request.uri.path starts_with \"/docs\"",
					Action:     CloudflareRulesetRule_route,
					Enabled:    boolPtr(true),
					ActionParameters: &CloudflareRulesetActionParameters{
						HostHeader: "planton.ai",
						Origin:     &CloudflareRulesetOrigin{Host: "k8s-lb.example.com", Port: 443},
					},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("CloudflareRulesetSpec Validation", func() {

	// ---- Positive cases ----

	ginkgo.Describe("Valid inputs", func() {

		ginkgo.It("accepts a minimal origin rule with zone_id", func() {
			gomega.Expect(protovalidate.Validate(validResource())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a ruleset with account_id instead of zone_id", func() {
			r := validResource()
			r.Spec.ZoneId = nil
			r.Spec.AccountId = "acct-xyz789"
			r.Spec.RulesetKind = rulesetKindPtr(CloudflareRulesetSpec_root)
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a block rule", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_http_request_firewall_custom
			r.Spec.Rules = []*CloudflareRulesetRule{
				{
					Expression: "ip.src eq 192.0.2.1",
					Action:     CloudflareRulesetRule_block,
					Enabled:    boolPtr(true),
					ActionParameters: &CloudflareRulesetActionParameters{
						Response: &CloudflareRulesetResponse{
							StatusCode:  403,
							Content:     "Forbidden",
							ContentType: "text/plain",
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts multiple rules", func() {
			r := validResource()
			r.Spec.Rules = append(r.Spec.Rules, &CloudflareRulesetRule{
				Expression: "http.request.uri.path starts_with \"/api\"",
				Action:     CloudflareRulesetRule_route,
				Enabled:    boolPtr(true),
			})
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a skip rule with phases and products", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_http_request_firewall_custom
			r.Spec.Rules = []*CloudflareRulesetRule{
				{
					Expression: "http.host eq \"example.com\"",
					Action:     CloudflareRulesetRule_skip,
					Enabled:    boolPtr(true),
					ActionParameters: &CloudflareRulesetActionParameters{
						Phases:   []string{"http_request_firewall_managed"},
						Products: []string{"waf"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a rule without action_parameters", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_http_request_firewall_custom
			r.Spec.Rules = []*CloudflareRulesetRule{
				{
					Expression: "true",
					Action:     CloudflareRulesetRule_challenge,
					Enabled:    boolPtr(true),
				},
			}
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})
	})

	// ---- Negative cases ----

	ginkgo.Describe("Invalid inputs", func() {

		ginkgo.It("rejects when neither zone_id nor account_id is set", func() {
			r := validResource()
			r.Spec.ZoneId = nil
			r.Spec.AccountId = ""
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("zone_id or account_id"))
		})

		ginkgo.It("rejects when both zone_id and account_id are set", func() {
			r := validResource()
			r.Spec.AccountId = "acct-xyz789"
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("zone_id or account_id"))
		})

		ginkgo.It("rejects unspecified phase", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_phase_unspecified
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("phase"))
		})

		ginkgo.It("rejects empty name", func() {
			r := validResource()
			r.Spec.Name = ""
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects empty rules list", func() {
			r := validResource()
			r.Spec.Rules = nil
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a rule with empty expression", func() {
			r := validResource()
			r.Spec.Rules[0].Expression = ""
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a rule with unspecified action", func() {
			r := validResource()
			r.Spec.Rules[0].Action = CloudflareRulesetRule_action_unspecified
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("action"))
		})
	})
})
