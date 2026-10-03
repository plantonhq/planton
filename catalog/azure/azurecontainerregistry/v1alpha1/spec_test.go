package azurecontainerregistryv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}

// validSpec returns a minimal valid Standard registry the failure cases
// mutate one field at a time.
func validSpec() *AzureContainerRegistrySpec {
	return &AzureContainerRegistrySpec{
		Region:        "eastus",
		ResourceGroup: stringRef("test-rg"),
		RegistryName:  "testregistry123",
	}
}

func validInput(spec *AzureContainerRegistrySpec) *AzureContainerRegistry {
	return &AzureContainerRegistry{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureContainerRegistry",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-container-registry",
		},
		Spec: spec,
	}
}

func TestAzureContainerRegistrySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureContainerRegistrySpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AzureContainerRegistrySpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal registry (unspecified SKU = Standard baseline)", func() {
			err := protovalidate.Validate(validInput(validSpec()))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a BASIC registry", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_BASIC
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a STANDARD registry with the admin user enabled", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_STANDARD
			spec.AdminUserEnabled = true
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept anonymous pull on the default (Standard) SKU", func() {
			spec := validSpec()
			spec.AnonymousPullEnabled = true
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a PREMIUM registry with geo-replications", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_PREMIUM
			spec.Georeplications = []*AzureContainerRegistryGeoreplication{
				{Location: "westeurope", ZoneRedundancyEnabled: true},
				{Location: "southeastasia", GlobalEndpointRoutingEnabled: true},
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a PREMIUM registry with zone redundancy and data endpoints", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_PREMIUM
			spec.ZoneRedundancyEnabled = true
			spec.DataEndpointEnabled = true
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept PREMIUM policies (quarantine, retention)", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_PREMIUM
			spec.QuarantinePolicyEnabled = true
			spec.RetentionPolicyInDays = proto.Int32(7)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a PREMIUM network rule set allowlist", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_PREMIUM
			spec.NetworkRuleSet = &AzureContainerRegistryNetworkRuleSet{
				DefaultAction: AzureContainerRegistryNetworkRuleDefaultAction_DENY,
				IpRules: []*AzureContainerRegistryIpRule{
					{IpRange: "203.0.113.0/24"},
				},
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept disabling export on a fully private PREMIUM registry", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_PREMIUM
			spec.PublicNetworkAccessEnabled = proto.Bool(false)
			spec.ExportPolicyEnabled = proto.Bool(false)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a system-assigned identity on any SKU", func() {
			spec := validSpec()
			spec.Identity = &AzureContainerRegistryIdentity{
				Type: AzureContainerRegistryIdentityType_SYSTEM_ASSIGNED,
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept PREMIUM CMK encryption with a user-assigned identity", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_PREMIUM
			spec.Identity = &AzureContainerRegistryIdentity{
				Type:        AzureContainerRegistryIdentityType_USER_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{stringRef("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/cmk")},
			}
			spec.Encryption = &AzureContainerRegistryEncryption{
				IdentityClientId: stringRef("00000000-0000-0000-0000-000000000000"),
				KeyVaultKeyId:    stringRef("https://vault.vault.azure.net/keys/acr-cmk"),
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept user tags", func() {
			spec := validSpec()
			spec.Tags = map[string]string{"cost-center": "platform"}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing region", func() {
			spec := validSpec()
			spec.Region = ""
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a missing registry_name", func() {
			spec := validSpec()
			spec.RegistryName = ""
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a registry_name shorter than 5 characters", func() {
			spec := validSpec()
			spec.RegistryName = "acr"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a registry_name with uppercase or special characters", func() {
			spec := validSpec()
			spec.RegistryName = "My-ACR-123"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject geo-replications on a non-PREMIUM SKU", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_STANDARD
			spec.Georeplications = []*AzureContainerRegistryGeoreplication{{Location: "westeurope"}}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a geo-replication into the registry's own region", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_PREMIUM
			spec.Georeplications = []*AzureContainerRegistryGeoreplication{{Location: "eastus"}}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject zone redundancy on a non-PREMIUM SKU", func() {
			spec := validSpec()
			spec.ZoneRedundancyEnabled = true
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject anonymous pull on the BASIC SKU", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_BASIC
			spec.AnonymousPullEnabled = true
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject data endpoints on a non-PREMIUM SKU", func() {
			spec := validSpec()
			spec.DataEndpointEnabled = true
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject the quarantine policy on a non-PREMIUM SKU", func() {
			spec := validSpec()
			spec.QuarantinePolicyEnabled = true
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a retention policy on a non-PREMIUM SKU", func() {
			spec := validSpec()
			spec.RetentionPolicyInDays = proto.Int32(7)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a retention window beyond 365 days", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_PREMIUM
			spec.RetentionPolicyInDays = proto.Int32(366)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject disabling export while public network access stays on", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_PREMIUM
			spec.ExportPolicyEnabled = proto.Bool(false)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a network rule set on a non-PREMIUM SKU", func() {
			spec := validSpec()
			spec.NetworkRuleSet = &AzureContainerRegistryNetworkRuleSet{
				DefaultAction: AzureContainerRegistryNetworkRuleDefaultAction_DENY,
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an ip_rule without an ip_range", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_PREMIUM
			spec.NetworkRuleSet = &AzureContainerRegistryNetworkRuleSet{
				IpRules: []*AzureContainerRegistryIpRule{{}},
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject CMK encryption on a non-PREMIUM SKU", func() {
			spec := validSpec()
			spec.Identity = &AzureContainerRegistryIdentity{
				Type:        AzureContainerRegistryIdentityType_USER_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{stringRef("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/cmk")},
			}
			spec.Encryption = &AzureContainerRegistryEncryption{
				IdentityClientId: stringRef("00000000-0000-0000-0000-000000000000"),
				KeyVaultKeyId:    stringRef("https://vault.vault.azure.net/keys/acr-cmk"),
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject CMK encryption without a user-assigned identity", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_PREMIUM
			spec.Identity = &AzureContainerRegistryIdentity{
				Type: AzureContainerRegistryIdentityType_SYSTEM_ASSIGNED,
			}
			spec.Encryption = &AzureContainerRegistryEncryption{
				IdentityClientId: stringRef("00000000-0000-0000-0000-000000000000"),
				KeyVaultKeyId:    stringRef("https://vault.vault.azure.net/keys/acr-cmk"),
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject encryption missing its key id", func() {
			spec := validSpec()
			spec.Sku = AzureContainerRegistrySku_PREMIUM
			spec.Identity = &AzureContainerRegistryIdentity{
				Type:        AzureContainerRegistryIdentityType_USER_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{stringRef("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/cmk")},
			}
			spec.Encryption = &AzureContainerRegistryEncryption{
				IdentityClientId: stringRef("00000000-0000-0000-0000-000000000000"),
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a USER_ASSIGNED identity without identity_ids", func() {
			spec := validSpec()
			spec.Identity = &AzureContainerRegistryIdentity{
				Type: AzureContainerRegistryIdentityType_USER_ASSIGNED,
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a SYSTEM_ASSIGNED identity carrying identity_ids", func() {
			spec := validSpec()
			spec.Identity = &AzureContainerRegistryIdentity{
				Type:        AzureContainerRegistryIdentityType_SYSTEM_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{stringRef("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/x")},
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
