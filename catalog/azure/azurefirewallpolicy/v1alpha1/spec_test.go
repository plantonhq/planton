package azurefirewallpolicyv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureFirewallPolicySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFirewallPolicySpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// ref builds a StringValueOrRef carrying a value_from reference.
func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

// validResource returns a minimal valid AzureFirewallPolicy that
// individual cases then mutate into the shape under test.
func validResource() *AzureFirewallPolicy {
	return &AzureFirewallPolicy{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFirewallPolicy",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-firewall-policy",
		},
		Spec: &AzureFirewallPolicySpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "egress-baseline",
		},
	}
}

// premiumResource returns a valid PREMIUM-tier policy for gating cases.
func premiumResource() *AzureFirewallPolicy {
	input := validResource()
	input.Spec.Sku = AzureFirewallPolicySku_PREMIUM
	return input
}

var _ = ginkgo.Describe("AzureFirewallPolicySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_firewall_policy", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit STANDARD sku", func() {
				input := validResource()
				input.Spec.Sku = AzureFirewallPolicySku_STANDARD
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a base policy reference", func() {
				input := validResource()
				input.Spec.BasePolicyId = ref("global-baseline")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a threat intelligence allowlist with only fqdns", func() {
				input := validResource()
				input.Spec.ThreatIntelligenceAllowlist = &AzureFirewallPolicyThreatIntelligenceAllowlist{
					Fqdns: []string{"partner.example.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept dns with servers and proxy", func() {
				input := validResource()
				input.Spec.Dns = &AzureFirewallPolicyDns{
					Servers:      []string{"10.0.0.4"},
					ProxyEnabled: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept intrusion detection on a PREMIUM policy", func() {
				input := premiumResource()
				input.Spec.IntrusionDetection = &AzureFirewallPolicyIntrusionDetection{
					Mode: AzureFirewallPolicyIntrusionDetectionState_IDPS_ALERT,
					SignatureOverrides: []*AzureFirewallPolicyIdpsSignatureOverride{
						{Id: "2024897", State: AzureFirewallPolicyIntrusionDetectionState_IDPS_OFF},
					},
					TrafficBypass: []*AzureFirewallPolicyIdpsTrafficBypass{
						{
							Name:             "trusted-backup-flow",
							Protocol:         AzureFirewallPolicyIdpsBypassProtocol_TCP,
							SourceAddresses:  []string{"10.0.1.0/24"},
							DestinationPorts: []string{"443"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a tls certificate with identity on a PREMIUM policy", func() {
				input := premiumResource()
				input.Spec.Identity = &AzureFirewallPolicyIdentity{
					Type:                    AzureFirewallPolicyIdentityType_USER_ASSIGNED,
					UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{ref("fw-tls-identity")},
				}
				input.Spec.TlsCertificate = &AzureFirewallPolicyTlsCertificate{
					KeyVaultSecretId: ref("egress-ca-cert"),
					Name:             "egress-ca",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept insights with an explicit false enabled", func() {
				input := validResource()
				input.Spec.Insights = &AzureFirewallPolicyInsights{
					Enabled:                        proto.Bool(false),
					DefaultLogAnalyticsWorkspaceId: ref("central-law"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit proxy configuration", func() {
				input := validResource()
				httpPort := int32(8087)
				httpsPort := int32(8088)
				input.Spec.ExplicitProxy = &AzureFirewallPolicyExplicitProxy{
					Enabled:   true,
					HttpPort:  &httpPort,
					HttpsPort: &httpsPort,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept private ip ranges and auto-learn", func() {
				input := validResource()
				input.Spec.PrivateIpRanges = []string{"10.0.0.0/8", "100.64.0.0/10"}
				input.Spec.AutoLearnPrivateRangesEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_firewall_policy", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a single-character name", func() {
				input := validResource()
				input.Spec.Name = "a"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the name ends with a period", func() {
				input := validResource()
				input.Spec.Name = "bad."
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when intrusion detection is set on a STANDARD policy", func() {
				input := validResource()
				input.Spec.Sku = AzureFirewallPolicySku_STANDARD
				input.Spec.IntrusionDetection = &AzureFirewallPolicyIntrusionDetection{
					Mode: AzureFirewallPolicyIntrusionDetectionState_IDPS_ALERT,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when intrusion detection is set with an unspecified sku", func() {
				input := validResource()
				input.Spec.IntrusionDetection = &AzureFirewallPolicyIntrusionDetection{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tls_certificate is set on a BASIC policy", func() {
				input := validResource()
				input.Spec.Sku = AzureFirewallPolicySku_BASIC
				input.Spec.TlsCertificate = &AzureFirewallPolicyTlsCertificate{
					KeyVaultSecretId: literal("https://kv.vault.azure.net/secrets/ca"),
					Name:             "ca",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an empty threat intelligence allowlist", func() {
				input := validResource()
				input.Spec.ThreatIntelligenceAllowlist = &AzureFirewallPolicyThreatIntelligenceAllowlist{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a USER_ASSIGNED identity without ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureFirewallPolicyIdentity{
					Type: AzureFirewallPolicyIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for identity ids with a SYSTEM_ASSIGNED type", func() {
				input := validResource()
				input.Spec.Identity = &AzureFirewallPolicyIdentity{
					Type:                    AzureFirewallPolicyIdentityType_SYSTEM_ASSIGNED,
					UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{ref("identity")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an identity block without a type", func() {
				input := validResource()
				input.Spec.Identity = &AzureFirewallPolicyIdentity{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a tls certificate omits the secret id", func() {
				input := premiumResource()
				input.Spec.TlsCertificate = &AzureFirewallPolicyTlsCertificate{
					Name: "ca",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when insights omits enabled", func() {
				input := validResource()
				input.Spec.Insights = &AzureFirewallPolicyInsights{
					DefaultLogAnalyticsWorkspaceId: ref("central-law"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when insights omits the default workspace", func() {
				input := validResource()
				input.Spec.Insights = &AzureFirewallPolicyInsights{
					Enabled: proto.Bool(true),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a proxy port above the provider bound", func() {
				input := validResource()
				port := int32(40000)
				input.Spec.ExplicitProxy = &AzureFirewallPolicyExplicitProxy{
					Enabled:  true,
					HttpPort: &port,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a traffic bypass without a protocol", func() {
				input := premiumResource()
				input.Spec.IntrusionDetection = &AzureFirewallPolicyIntrusionDetection{
					TrafficBypass: []*AzureFirewallPolicyIdpsTrafficBypass{
						{Name: "no-protocol"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a traffic bypass without a name", func() {
				input := premiumResource()
				input.Spec.IntrusionDetection = &AzureFirewallPolicyIntrusionDetection{
					TrafficBypass: []*AzureFirewallPolicyIdpsTrafficBypass{
						{Protocol: AzureFirewallPolicyIdpsBypassProtocol_ANY},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
