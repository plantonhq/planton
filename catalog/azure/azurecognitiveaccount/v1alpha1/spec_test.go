package azurecognitiveaccountv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureCognitiveAccountSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureCognitiveAccountSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// validResource returns a minimal valid OpenAI account that individual
// cases mutate into the shape under test.
func validResource() *AzureCognitiveAccount {
	return &AzureCognitiveAccount{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureCognitiveAccount",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-cognitive-account",
		},
		Spec: &AzureCognitiveAccountSpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "openai-prod",
			Kind:          "OpenAI",
			SkuName:       "S0",
		},
	}
}

// systemIdentity returns a system-assigned identity block.
func systemIdentity() *AzureCognitiveAccountIdentity {
	return &AzureCognitiveAccountIdentity{
		Type: AzureCognitiveAccountIdentityType_SYSTEM_ASSIGNED,
	}
}

var _ = ginkgo.Describe("AzureCognitiveAccountSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_cognitive_account", func() {

			ginkgo.It("should not return a validation error for a minimal OpenAI account", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an AIServices account with project management and an identity", func() {
				input := validResource()
				input.Spec.Kind = "AIServices"
				input.Spec.ProjectManagementEnabled = true
				input.Spec.Identity = systemIdentity()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept network ACLs with a custom subdomain and bypass on an OpenAI account", func() {
				input := validResource()
				input.Spec.CustomSubdomainName = "openai-prod"
				input.Spec.NetworkAcls = &AzureCognitiveAccountNetworkAcls{
					DefaultAction: "Deny",
					IpRules:       []string{"203.0.113.0/24", "198.51.100.7"},
					VirtualNetworkRules: []*AzureCognitiveAccountVirtualNetworkRule{
						{SubnetId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/app")},
					},
					Bypass: AzureCognitiveAccountNetworkAclsBypass_AZURE_SERVICES,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a customer-managed key with a user-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureCognitiveAccountIdentity{
					Type:        AzureCognitiveAccountIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/cmk")},
				}
				input.Spec.CustomerManagedKey = &AzureCognitiveAccountCustomerManagedKey{
					KeyVaultKeyId:    literal("https://kv.vault.azure.net/keys/cog-cmk"),
					IdentityClientId: "41b23e61-6c1e-4545-b367-cd054e0ed4b4",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept user-owned storage and an FQDN allowlist with outbound restriction", func() {
				input := validResource()
				input.Spec.Identity = systemIdentity()
				input.Spec.Storage = []*AzureCognitiveAccountStorage{
					{StorageAccountId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/sa")},
				}
				input.Spec.Fqdns = []string{"search.contoso.net"}
				input.Spec.OutboundNetworkAccessRestricted = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a MetricsAdvisor account with all four metrics-advisor fields", func() {
				input := validResource()
				input.Spec.Kind = "MetricsAdvisor"
				input.Spec.MetricsAdvisorAadClientId = "41b23e61-6c1e-4545-b367-cd054e0ed4b4"
				input.Spec.MetricsAdvisorAadTenantId = "72f988bf-86f1-41af-91ab-2d7cd011db47"
				input.Spec.MetricsAdvisorSuperUserName = "admin"
				input.Spec.MetricsAdvisorWebsiteName = "metrics-portal"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a TextAnalytics account with the custom-question-answering search fields", func() {
				input := validResource()
				input.Spec.Kind = "TextAnalytics"
				input.Spec.CustomQuestionAnsweringSearchServiceId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Search/searchServices/qna"
				input.Spec.CustomQuestionAnsweringSearchServiceKey = literal("search-api-key")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a QnAMaker account with its runtime endpoint", func() {
				input := validResource()
				input.Spec.Kind = "QnAMaker"
				input.Spec.QnaRuntimeEndpoint = "https://qna-runtime.azurewebsites.net"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept responsible-AI blocklists and policies", func() {
				input := validResource()
				input.Spec.RaiBlocklists = []*AzureCognitiveAccountRaiBlocklist{
					{Name: "competitor-names", Description: "blocked competitor mentions"},
				}
				input.Spec.RaiPolicies = []*AzureCognitiveAccountRaiPolicy{
					{
						Name:           "strict-chat",
						BasePolicyName: "Microsoft.Default",
						ContentFilters: []*AzureCognitiveAccountRaiPolicyContentFilter{
							{Name: "Hate", FilterEnabled: true, BlockEnabled: true, Source: "Prompt", SeverityThreshold: AzureCognitiveAccountRaiPolicyContentLevel_LOW},
							{Name: "Jailbreak", FilterEnabled: true, BlockEnabled: true, Source: "Prompt"},
						},
						Mode: AzureCognitiveAccountRaiPolicyMode_BLOCKING,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept agent network injection on an AIServices account", func() {
				input := validResource()
				input.Spec.Kind = "AIServices"
				input.Spec.NetworkInjection = &AzureCognitiveAccountNetworkInjection{
					Scenario: "agent",
					SubnetId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/agents"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept dynamic throttling on a Speech account", func() {
				input := validResource()
				input.Spec.Kind = "SpeechServices"
				input.Spec.DynamicThrottlingEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_cognitive_account", func() {

			ginkgo.It("should reject a missing kind", func() {
				input := validResource()
				input.Spec.Kind = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a kind outside the vocabulary", func() {
				input := validResource()
				input.Spec.Kind = "OpenAi"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a sku outside the vocabulary", func() {
				input := validResource()
				input.Spec.SkuName = "S99"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a single-character account name", func() {
				input := validResource()
				input.Spec.Name = "a"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an account name starting with a dash", func() {
				input := validResource()
				input.Spec.Name = "-openai"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject project management on a non-AIServices kind", func() {
				input := validResource()
				input.Spec.ProjectManagementEnabled = true
				input.Spec.Identity = systemIdentity()
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject project management without an identity", func() {
				input := validResource()
				input.Spec.Kind = "AIServices"
				input.Spec.ProjectManagementEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject dynamic throttling on an OpenAI account", func() {
				input := validResource()
				input.Spec.DynamicThrottlingEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject network ACLs without a custom subdomain", func() {
				input := validResource()
				input.Spec.NetworkAcls = &AzureCognitiveAccountNetworkAcls{DefaultAction: "Deny"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject bypass on a kind outside OpenAI/AIServices/TextAnalytics", func() {
				input := validResource()
				input.Spec.Kind = "SpeechServices"
				input.Spec.CustomSubdomainName = "speech-prod"
				input.Spec.NetworkAcls = &AzureCognitiveAccountNetworkAcls{
					DefaultAction: "Deny",
					Bypass:        AzureCognitiveAccountNetworkAclsBypass_AZURE_SERVICES,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an ip rule that is neither an IPv4 address nor a CIDR", func() {
				input := validResource()
				input.Spec.CustomSubdomainName = "openai-prod"
				input.Spec.NetworkAcls = &AzureCognitiveAccountNetworkAcls{
					DefaultAction: "Deny",
					IpRules:       []string{"not-an-ip"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject network injection on a non-AIServices kind", func() {
				input := validResource()
				input.Spec.NetworkInjection = &AzureCognitiveAccountNetworkInjection{
					Scenario: "agent",
					SubnetId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/agents"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a QnAMaker account without the runtime endpoint", func() {
				input := validResource()
				input.Spec.Kind = "QnAMaker"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a runtime endpoint on a non-QnAMaker kind", func() {
				input := validResource()
				input.Spec.QnaRuntimeEndpoint = "https://qna-runtime.azurewebsites.net"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject the QnA search service id on a non-TextAnalytics kind", func() {
				input := validResource()
				input.Spec.CustomQuestionAnsweringSearchServiceId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Search/searchServices/qna"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a metrics-advisor field on a non-MetricsAdvisor kind", func() {
				input := validResource()
				input.Spec.MetricsAdvisorSuperUserName = "admin"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate responsible-AI policy names", func() {
				input := validResource()
				policy := &AzureCognitiveAccountRaiPolicy{
					Name:           "strict-chat",
					BasePolicyName: "Microsoft.Default",
					ContentFilters: []*AzureCognitiveAccountRaiPolicyContentFilter{
						{Name: "Hate", FilterEnabled: true, BlockEnabled: true, Source: "Prompt"},
					},
				}
				input.Spec.RaiPolicies = []*AzureCognitiveAccountRaiPolicy{policy, policy}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate responsible-AI blocklist names", func() {
				input := validResource()
				blocklist := &AzureCognitiveAccountRaiBlocklist{Name: "competitor-names"}
				input.Spec.RaiBlocklists = []*AzureCognitiveAccountRaiBlocklist{blocklist, blocklist}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a blocklist name with illegal characters", func() {
				input := validResource()
				input.Spec.RaiBlocklists = []*AzureCognitiveAccountRaiBlocklist{{Name: "bad name!"}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a severity threshold on the binary Jailbreak filter", func() {
				input := validResource()
				input.Spec.RaiPolicies = []*AzureCognitiveAccountRaiPolicy{
					{
						Name:           "strict-chat",
						BasePolicyName: "Microsoft.Default",
						ContentFilters: []*AzureCognitiveAccountRaiPolicyContentFilter{
							{Name: "Jailbreak", FilterEnabled: true, BlockEnabled: true, Source: "Prompt", SeverityThreshold: AzureCognitiveAccountRaiPolicyContentLevel_LOW},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a policy without content filters", func() {
				input := validResource()
				input.Spec.RaiPolicies = []*AzureCognitiveAccountRaiPolicy{
					{Name: "strict-chat", BasePolicyName: "Microsoft.Default"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a content filter with a source outside the vocabulary", func() {
				input := validResource()
				input.Spec.RaiPolicies = []*AzureCognitiveAccountRaiPolicy{
					{
						Name:           "strict-chat",
						BasePolicyName: "Microsoft.Default",
						ContentFilters: []*AzureCognitiveAccountRaiPolicyContentFilter{
							{Name: "Hate", FilterEnabled: true, BlockEnabled: true, Source: "Input"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureCognitiveAccountIdentity{
					Type:        AzureCognitiveAccountIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/x")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureCognitiveAccountIdentity{
					Type: AzureCognitiveAccountIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a customer-managed key with a malformed identity client id", func() {
				input := validResource()
				input.Spec.CustomerManagedKey = &AzureCognitiveAccountCustomerManagedKey{
					KeyVaultKeyId:    literal("https://kv.vault.azure.net/keys/cog-cmk"),
					IdentityClientId: "not-a-uuid",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
