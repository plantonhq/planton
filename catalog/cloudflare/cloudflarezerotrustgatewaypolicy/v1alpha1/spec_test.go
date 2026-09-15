package cloudflarezerotrustgatewaypolicyv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareZeroTrustGatewayPolicySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustGatewayPolicySpec Custom Validation Tests")
}

const testAccountId = "023e105f4ecef8ad9ca31a8372d0c353"

func boolPtr(b bool) *bool { return &b }

func vnetRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func validPolicy(spec *CloudflareZeroTrustGatewayPolicySpec) *CloudflareZeroTrustGatewayPolicy {
	return &CloudflareZeroTrustGatewayPolicy{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZeroTrustGatewayPolicy",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-gateway-policy",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareZeroTrustGatewayPolicySpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal DNS block policy", func() {
			input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
				AccountId: testAccountId,
				Name:      "block-gambling",
				Action:    "block",
				Filter:    "dns",
				Enabled:   boolPtr(true),
				Traffic:   "any(dns.domains[*] == \"example.com\")",
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an HTTP allow policy with settings", func() {
			input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
				AccountId: testAccountId,
				Name:      "allow-with-session-check",
				Action:    "allow",
				Filter:    "http",
				Enabled:   boolPtr(true),
				Traffic:   "http.request.uri matches \".*example.com.*\"",
				RuleSettings: &CloudflareZeroTrustGatewayPolicyRuleSettings{
					CheckSession: &CloudflareZeroTrustGatewayPolicyCheckSession{
						Duration: "24h",
						Enforce:  true,
					},
					BlockPage: &CloudflareZeroTrustGatewayPolicyBlockPage{
						TargetUri: "https://intranet.example.com/blocked",
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a resolver policy with custom upstreams over a vnet", func() {
			input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
				AccountId: testAccountId,
				Name:      "resolve-internal",
				Action:    "resolve",
				Filter:    "dns_resolver",
				Enabled:   boolPtr(true),
				RuleSettings: &CloudflareZeroTrustGatewayPolicyRuleSettings{
					DnsResolvers: &CloudflareZeroTrustGatewayPolicyDnsResolvers{
						Ipv4: []*CloudflareZeroTrustGatewayPolicyDnsResolverV4{
							{
								Ip:                         "10.0.0.53",
								RouteThroughPrivateNetwork: true,
								VnetId:                     vnetRef("6a7e50b8-8e0c-4d0a-9d1c-000000000000"),
							},
						},
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an expiration and schedule on a DNS policy", func() {
			input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
				AccountId: testAccountId,
				Name:      "temporary-block",
				Action:    "block",
				Filter:    "dns",
				Enabled:   boolPtr(true),
				Expiration: &CloudflareZeroTrustGatewayPolicyExpiration{
					ExpiresAt: "2026-09-01T00:00:00Z",
				},
				Schedule: &CloudflareZeroTrustGatewayPolicySchedule{
					Mon:      "08:00-17:00",
					TimeZone: "America/New_York",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept logging blocks that switch each surface off", func() {
			input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
				AccountId: testAccountId,
				Name:      "quiet-allow",
				Action:    "allow",
				Filter:    "http",
				Enabled:   boolPtr(true),
				Traffic:   "http.request.uri matches \".*example.com.*\"",
				RuleSettings: &CloudflareZeroTrustGatewayPolicyRuleSettings{
					AuditSsh:     &CloudflareZeroTrustGatewayPolicyAuditSsh{CommandLogging: boolPtr(false)},
					ForensicCopy: &CloudflareZeroTrustGatewayPolicyForensicCopy{Enabled: boolPtr(false)},
					PayloadLog:   &CloudflareZeroTrustGatewayPolicyPayloadLog{Enabled: boolPtr(false)},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a logging block that does not say on or off", func() {
			for _, settings := range []*CloudflareZeroTrustGatewayPolicyRuleSettings{
				{AuditSsh: &CloudflareZeroTrustGatewayPolicyAuditSsh{}},
				{ForensicCopy: &CloudflareZeroTrustGatewayPolicyForensicCopy{}},
				{PayloadLog: &CloudflareZeroTrustGatewayPolicyPayloadLog{}},
			} {
				input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
					AccountId:    testAccountId,
					Name:         "half-declared",
					Action:       "allow",
					Filter:       "http",
					Enabled:      boolPtr(true),
					Traffic:      "http.request.uri matches \".*example.com.*\"",
					RuleSettings: settings,
				})
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("required"))
			}
		})

		ginkgo.It("should reject a missing account_id", func() {
			input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
				Name:   "block-gambling",
				Action: "block",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown action", func() {
			input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
				AccountId: testAccountId,
				Name:      "block-gambling",
				Action:    "reject",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown filter", func() {
			input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
				AccountId: testAccountId,
				Name:      "block-gambling",
				Action:    "block",
				Filter:    "smtp",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an expiration without a valid timestamp", func() {
			input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
				AccountId: testAccountId,
				Name:      "temporary-block",
				Action:    "block",
				Expiration: &CloudflareZeroTrustGatewayPolicyExpiration{
					ExpiresAt: "tomorrow",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a vnet resolver that does not route through the private network", func() {
			input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
				AccountId: testAccountId,
				Name:      "resolve-internal",
				Action:    "resolve",
				RuleSettings: &CloudflareZeroTrustGatewayPolicyRuleSettings{
					DnsResolvers: &CloudflareZeroTrustGatewayPolicyDnsResolvers{
						Ipv4: []*CloudflareZeroTrustGatewayPolicyDnsResolverV4{
							{
								Ip:     "10.0.0.53",
								VnetId: vnetRef("6a7e50b8-8e0c-4d0a-9d1c-000000000000"),
							},
						},
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject two resolver mechanisms on one policy", func() {
			input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
				AccountId: testAccountId,
				Name:      "resolve-internal",
				Action:    "resolve",
				RuleSettings: &CloudflareZeroTrustGatewayPolicyRuleSettings{
					ResolveDnsThroughCloudflare: boolPtr(true),
					ResolveDnsInternally: &CloudflareZeroTrustGatewayPolicyResolveDnsInternally{
						Fallback: "public_dns",
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid quarantine file type", func() {
			input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
				AccountId: testAccountId,
				Name:      "sandbox-downloads",
				Action:    "quarantine",
				Filter:    "http",
				RuleSettings: &CloudflareZeroTrustGatewayPolicyRuleSettings{
					Quarantine: &CloudflareZeroTrustGatewayPolicyQuarantine{
						FileTypes: []string{"exe", "iso"},
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty header-value list", func() {
			input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
				AccountId: testAccountId,
				Name:      "allow-with-headers",
				Action:    "allow",
				Filter:    "http",
				RuleSettings: &CloudflareZeroTrustGatewayPolicyRuleSettings{
					AddHeaders: map[string]*CloudflareZeroTrustGatewayPolicyStringList{
						"X-Custom": {Values: []string{}},
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid biso admin control value", func() {
			input := validPolicy(&CloudflareZeroTrustGatewayPolicySpec{
				AccountId: testAccountId,
				Name:      "isolate-risky",
				Action:    "isolate",
				Filter:    "http",
				RuleSettings: &CloudflareZeroTrustGatewayPolicyRuleSettings{
					BisoAdminControls: &CloudflareZeroTrustGatewayPolicyBisoAdminControls{
						Version: "v2",
						Copy:    "sometimes",
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
