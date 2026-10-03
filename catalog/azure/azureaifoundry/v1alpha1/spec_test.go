package azureaifoundryv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureAiFoundrySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureAiFoundrySpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testVaultId    = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/hub-vault"
	testStorageId  = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/hubstorage"
	testInsightsId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Insights/components/hub-insights"
	testRegistryId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ContainerRegistry/registries/hubacr"
	testUaiId      = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/hub-uai"
	// The hub's key must be a VERSIONED Key Vault key URL (the
	// provider's hub contract -- versionless is rejected).
	testVersionedKeyId = "https://hub-vault.vault.azure.net/keys/hub-cmk/0123456789abcdef0123456789abcdef"
)

// validResource returns a minimal valid hub that individual cases
// mutate into the shape under test.
func validResource() *AzureAiFoundry {
	return &AzureAiFoundry{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureAiFoundry",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ai-foundry",
		},
		Spec: &AzureAiFoundrySpec{
			Region:           "eastus",
			ResourceGroup:    literal("ai-rg"),
			Name:             "team-hub",
			KeyVaultId:       literal(testVaultId),
			StorageAccountId: literal(testStorageId),
			Identity: &AzureAiFoundryIdentity{
				Type: AzureAiFoundryIdentityType_SYSTEM_ASSIGNED,
			},
		},
	}
}

var _ = ginkgo.Describe("AzureAiFoundrySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_ai_foundry", func() {

			ginkgo.It("should not return a validation error for a minimal hub", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a name with underscores (the provider's code regex allows them)", func() {
				input := validResource()
				input.Spec.Name = "team_hub_01"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity with a primary identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureAiFoundryIdentity{
					Type:        AzureAiFoundryIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testUaiId)},
				}
				input.Spec.PrimaryUserAssignedIdentity = literal(testUaiId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept CMK encryption with a versioned key URL", func() {
				input := validResource()
				input.Spec.Encryption = &AzureAiFoundryEncryption{
					KeyVaultId:             literal(testVaultId),
					KeyId:                  literal(testVersionedKeyId),
					UserAssignedIdentityId: literal(testUaiId),
				}
				input.Spec.HighBusinessImpactEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the full optional surface (insights, registry, network, access)", func() {
				input := validResource()
				input.Spec.ApplicationInsightsId = literal(testInsightsId)
				input.Spec.ContainerRegistryId = literal(testRegistryId)
				input.Spec.ManagedNetwork = &AzureAiFoundryManagedNetwork{
					IsolationMode: AzureAiFoundryIsolationMode_ALLOW_ONLY_APPROVED_OUTBOUND,
				}
				publicAccess := false
				input.Spec.PublicNetworkAccessEnabled = &publicAccess
				input.Spec.Description = "The shared AI foundation"
				input.Spec.FriendlyName = "Team Hub"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_ai_foundry", func() {

			ginkgo.It("should reject a name shorter than 3 characters", func() {
				input := validResource()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name longer than 33 characters", func() {
				input := validResource()
				input.Spec.Name = "a123456789012345678901234567890123"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name starting with a hyphen", func() {
				input := validResource()
				input.Spec.Name = "-team-hub"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a hub without a key vault", func() {
				input := validResource()
				input.Spec.KeyVaultId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a hub without a storage account", func() {
				input := validResource()
				input.Spec.StorageAccountId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a hub without an identity", func() {
				input := validResource()
				input.Spec.Identity = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying identity_ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureAiFoundryIdentity{
					Type:        AzureAiFoundryIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testUaiId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity_ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureAiFoundryIdentity{
					Type: AzureAiFoundryIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject encryption without a key", func() {
				input := validResource()
				input.Spec.Encryption = &AzureAiFoundryEncryption{
					KeyVaultId: literal(testVaultId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a hub without a region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
