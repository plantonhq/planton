package azurekeyvaultv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureKeyVaultSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureKeyVaultSpec Custom Validation Tests")
}

func validSpec() *AzureKeyVaultSpec {
	return &AzureKeyVaultSpec{
		Region:        "eastus",
		ResourceGroup: stringRef("test-resource-group"),
		VaultName:     "test-vault",
	}
}

func vault(spec *AzureKeyVaultSpec) *AzureKeyVault {
	return &AzureKeyVault{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureKeyVault",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-key-vault",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("AzureKeyVaultSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_key_vault", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(vault(validSpec()))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a production configuration with premium SKU and purge protection", func() {
				spec := validSpec()
				spec.Sku = AzureKeyVaultSku_PREMIUM
				spec.RbacAuthorizationEnabled = boolPtr(true)
				spec.PurgeProtectionEnabled = true
				spec.SoftDeleteRetentionDays = int32Ptr(90)
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a vault name at the length boundaries", func() {
				spec := validSpec()
				spec.VaultName = "abc"
				gomega.Expect(protovalidate.Validate(vault(spec))).To(gomega.BeNil())
				spec.VaultName = "a12345678901234567890123" // 24 chars
				gomega.Expect(protovalidate.Validate(vault(spec))).To(gomega.BeNil())
			})

			ginkgo.It("should accept network ACLs with DENY, bypass, IP rules and subnet references", func() {
				spec := validSpec()
				spec.NetworkAcls = &AzureKeyVaultNetworkAcls{
					DefaultAction: AzureKeyVaultNetworkAclsDefaultAction_DENY,
					Bypass:        AzureKeyVaultNetworkAclsBypass_AZURE_SERVICES,
					IpRules:       []string{"203.0.113.0/24", "198.51.100.42"},
					VirtualNetworkSubnetIds: []*foreignkeyv1.StringValueOrRef{
						stringRef("/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet1"),
					},
				}
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept network ACLs allowing all traffic with bypass NONE", func() {
				spec := validSpec()
				spec.NetworkAcls = &AzureKeyVaultNetworkAcls{
					DefaultAction: AzureKeyVaultNetworkAclsDefaultAction_ALLOW,
					Bypass:        AzureKeyVaultNetworkAclsBypass_NONE,
				}
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a legacy access-policy vault", func() {
				spec := validSpec()
				spec.RbacAuthorizationEnabled = boolPtr(false)
				spec.AccessPolicies = []*AzureKeyVaultAccessPolicy{
					{
						ObjectId: stringRef("00000000-0000-0000-0000-000000000001"),
						KeyPermissions: []AzureKeyVaultKeyPermission{
							AzureKeyVaultKeyPermission_KEY_GET,
							AzureKeyVaultKeyPermission_KEY_UNWRAP_KEY,
							AzureKeyVaultKeyPermission_KEY_WRAP_KEY,
						},
						SecretPermissions: []AzureKeyVaultSecretPermission{
							AzureKeyVaultSecretPermission_SECRET_GET,
						},
						CertificatePermissions: []AzureKeyVaultCertificatePermission{
							AzureKeyVaultCertificatePermission_CERTIFICATE_GET,
						},
					},
				}
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an access policy with explicit tenant and application ids", func() {
				spec := validSpec()
				spec.AccessPolicies = []*AzureKeyVaultAccessPolicy{
					{
						ObjectId:      stringRef("00000000-0000-0000-0000-000000000001"),
						TenantId:      strPtr("11111111-1111-1111-1111-111111111111"),
						ApplicationId: strPtr("22222222-2222-2222-2222-222222222222"),
						SecretPermissions: []AzureKeyVaultSecretPermission{
							AzureKeyVaultSecretPermission_SECRET_GET,
							AzureKeyVaultSecretPermission_SECRET_LIST,
						},
					},
				}
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the deployment-integration flags and user tags", func() {
				spec := validSpec()
				spec.EnabledForDeployment = true
				spec.EnabledForDiskEncryption = true
				spec.EnabledForTemplateDeployment = true
				spec.PublicNetworkAccessEnabled = boolPtr(false)
				spec.Tags = map[string]string{"team": "security", "cost-center": "1234"}
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept soft delete retention at the bounds", func() {
				spec := validSpec()
				spec.SoftDeleteRetentionDays = int32Ptr(7)
				gomega.Expect(protovalidate.Validate(vault(spec))).To(gomega.BeNil())
				spec.SoftDeleteRetentionDays = int32Ptr(90)
				gomega.Expect(protovalidate.Validate(vault(spec))).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_key_vault", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				spec := validSpec()
				spec.Region = ""
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				spec := validSpec()
				spec.ResourceGroup = nil
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vault_name is missing", func() {
				spec := validSpec()
				spec.VaultName = ""
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vault_name is too short", func() {
				spec := validSpec()
				spec.VaultName = "ab"
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vault_name is too long", func() {
				spec := validSpec()
				spec.VaultName = "a123456789012345678901234" // 25 chars
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vault_name starts with a digit", func() {
				spec := validSpec()
				spec.VaultName = "1vault"
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vault_name ends with a hyphen", func() {
				spec := validSpec()
				spec.VaultName = "vault-"
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vault_name contains consecutive hyphens", func() {
				spec := validSpec()
				spec.VaultName = "my--vault"
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vault_name contains invalid characters", func() {
				spec := validSpec()
				spec.VaultName = "my_vault"
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when soft_delete_retention_days is out of range", func() {
				spec := validSpec()
				spec.SoftDeleteRetentionDays = int32Ptr(6)
				gomega.Expect(protovalidate.Validate(vault(spec))).ToNot(gomega.BeNil())
				spec.SoftDeleteRetentionDays = int32Ptr(91)
				gomega.Expect(protovalidate.Validate(vault(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network ACLs omit default_action", func() {
				spec := validSpec()
				spec.NetworkAcls = &AzureKeyVaultNetworkAcls{
					IpRules: []string{"203.0.113.0/24"},
				}
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network ACLs has too many IP rules (> 200)", func() {
				ipRules := make([]string, 201)
				for i := 0; i < 201; i++ {
					ipRules[i] = "10.0.0.1/32"
				}
				spec := validSpec()
				spec.NetworkAcls = &AzureKeyVaultNetworkAcls{
					DefaultAction: AzureKeyVaultNetworkAclsDefaultAction_DENY,
					IpRules:       ipRules,
				}
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when an access policy omits object_id", func() {
				spec := validSpec()
				spec.AccessPolicies = []*AzureKeyVaultAccessPolicy{
					{
						SecretPermissions: []AzureKeyVaultSecretPermission{
							AzureKeyVaultSecretPermission_SECRET_GET,
						},
					},
				}
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when an access policy carries an unspecified permission", func() {
				spec := validSpec()
				spec.AccessPolicies = []*AzureKeyVaultAccessPolicy{
					{
						ObjectId: stringRef("00000000-0000-0000-0000-000000000001"),
						KeyPermissions: []AzureKeyVaultKeyPermission{
							AzureKeyVaultKeyPermission_azure_key_vault_key_permission_unspecified,
						},
					},
				}
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when an access policy tenant_id is not a UUID", func() {
				spec := validSpec()
				spec.AccessPolicies = []*AzureKeyVaultAccessPolicy{
					{
						ObjectId: stringRef("00000000-0000-0000-0000-000000000001"),
						TenantId: strPtr("not-a-uuid"),
					},
				}
				err := protovalidate.Validate(vault(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := vault(validSpec())
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := vault(validSpec())
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := vault(validSpec())
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := vault(validSpec())
				input.Spec = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

// Helper functions for pointer types
func boolPtr(b bool) *bool {
	return &b
}

func int32Ptr(i int32) *int32 {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}
