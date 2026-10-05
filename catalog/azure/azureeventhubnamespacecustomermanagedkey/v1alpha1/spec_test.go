package azureeventhubnamespacecustomermanagedkeyv1alpha1

import (
	"fmt"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureEventHubNamespaceCustomerManagedKeySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventHubNamespaceCustomerManagedKeySpec Validation Tests")
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

// helper to create a minimal valid CMK configuration
func minimalCustomerManagedKey() *AzureEventHubNamespaceCustomerManagedKey {
	return &AzureEventHubNamespaceCustomerManagedKey{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureEventHubNamespaceCustomerManagedKey",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-eh-cmk",
		},
		Spec: &AzureEventHubNamespaceCustomerManagedKeySpec{
			EventhubNamespaceId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.EventHub/namespaces/ehns"),
			KeyVaultKeyIds: []*foreignkeyv1.StringValueOrRef{
				literal("https://vault.vault.azure.net/keys/cmk"),
			},
		},
	}
}

var _ = ginkgo.Describe("AzureEventHubNamespaceCustomerManagedKeySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_event_hub_namespace_customer_managed_key", func() {

			ginkgo.It("should accept a minimal configuration", func() {
				err := protovalidate.Validate(minimalCustomerManagedKey())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept two keys with infrastructure encryption and an identity", func() {
				infra := true
				input := minimalCustomerManagedKey()
				input.Spec.KeyVaultKeyIds = []*foreignkeyv1.StringValueOrRef{
					literal("https://vault.vault.azure.net/keys/cmk-a"),
					literal("https://vault.vault.azure.net/keys/cmk-b"),
				}
				input.Spec.InfrastructureEncryptionEnabled = &infra
				input.Spec.UserAssignedIdentityId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_event_hub_namespace_customer_managed_key", func() {

			ginkgo.It("should reject a missing eventhub_namespace_id", func() {
				input := minimalCustomerManagedKey()
				input.Spec.EventhubNamespaceId = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an empty key_vault_key_ids list", func() {
				input := minimalCustomerManagedKey()
				input.Spec.KeyVaultKeyIds = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject more than ten key ids", func() {
				input := minimalCustomerManagedKey()
				keyIds := make([]*foreignkeyv1.StringValueOrRef, 0, 11)
				for i := 0; i < 11; i++ {
					keyIds = append(keyIds, literal(fmt.Sprintf("https://vault.vault.azure.net/keys/cmk-%d", i)))
				}
				input.Spec.KeyVaultKeyIds = keyIds
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
