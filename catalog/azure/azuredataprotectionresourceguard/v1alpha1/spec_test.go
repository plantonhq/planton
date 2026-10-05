package azuredataprotectionresourceguardv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureDataProtectionResourceGuardSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDataProtectionResourceGuardSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// validResource returns a minimal valid guard that individual cases
// mutate into the shape under test.
func validResource() *AzureDataProtectionResourceGuard {
	return &AzureDataProtectionResourceGuard{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDataProtectionResourceGuard",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-resource-guard",
		},
		Spec: &AzureDataProtectionResourceGuardSpec{
			Region:        "eastus",
			ResourceGroup: literal("security-rg"),
			Name:          "backup-mua-guard",
		},
	}
}

var _ = ginkgo.Describe("AzureDataProtectionResourceGuardSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_data_protection_resource_guard", func() {

			ginkgo.It("should not return a validation error for a minimal guard", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an exclusion list of excludable operations", func() {
				input := validResource()
				input.Spec.VaultCriticalOperationExclusionList = []string{
					"Microsoft.RecoveryServices/vaults/backupResourceGuardProxies/write",
					"Microsoft.RecoveryServices/vaults/backupconfig/delete",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a 260-character name", func() {
				input := validResource()
				input.Spec.Name = strings.Repeat("a", 260)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept user tags", func() {
				input := validResource()
				input.Spec.Tags = map[string]string{"owner": "security-team"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_data_protection_resource_guard", func() {

			ginkgo.It("should reject a missing name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name over 260 characters", func() {
				input := validResource()
				input.Spec.Name = strings.Repeat("a", 261)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty exclusion-list entry", func() {
				input := validResource()
				input.Spec.VaultCriticalOperationExclusionList = []string{""}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject excluding an MUA-mandatory operation", func() {
				// ARM rejects these with
				// BMSUserErrorInvalidCriticalOperationExclusionList; the
				// spec mirrors that list so the failure lands at
				// admission.
				input := validResource()
				input.Spec.VaultCriticalOperationExclusionList = []string{
					"Microsoft.RecoveryServices/vaults/backupconfig/write",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a mandatory operation mixed into an otherwise-legal list", func() {
				input := validResource()
				input.Spec.VaultCriticalOperationExclusionList = []string{
					"Microsoft.RecoveryServices/vaults/backupResourceGuardProxies/write",
					"Microsoft.DataProtection/backupVaults/write#reduceSoftDeleteSecurity",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
