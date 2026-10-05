package azurekeyvaultkeyv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureKeyVaultKeySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureKeyVaultKeySpec Custom Validation Tests")
}

func validRsaSpec() *AzureKeyVaultKeySpec {
	return &AzureKeyVaultKeySpec{
		Name:       "cmk-storage",
		KeyVaultId: stringRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/my-vault"),
		KeyType:    AzureKeyVaultKeyType_RSA,
		KeySize:    int32Ptr(2048),
		KeyOpts: []AzureKeyVaultKeyOperation{
			AzureKeyVaultKeyOperation_WRAP_KEY,
			AzureKeyVaultKeyOperation_UNWRAP_KEY,
		},
	}
}

func key(spec *AzureKeyVaultKeySpec) *AzureKeyVaultKey {
	return &AzureKeyVaultKey{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureKeyVaultKey",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-key",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("AzureKeyVaultKeySpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_key_vault_key", func() {

			ginkgo.It("should accept a minimal RSA wrap/unwrap key", func() {
				err := protovalidate.Validate(key(validRsaSpec()))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every RSA key size", func() {
				for _, size := range []int32{2048, 3072, 4096} {
					spec := validRsaSpec()
					spec.KeySize = int32Ptr(size)
					gomega.Expect(protovalidate.Validate(key(spec))).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept an HSM-backed RSA key", func() {
				spec := validRsaSpec()
				spec.KeyType = AzureKeyVaultKeyType_RSA_HSM
				err := protovalidate.Validate(key(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an EC signing key with an explicit curve", func() {
				spec := validRsaSpec()
				spec.KeyType = AzureKeyVaultKeyType_EC
				spec.KeySize = nil
				spec.Curve = AzureKeyVaultKeyCurve_P_384
				spec.KeyOpts = []AzureKeyVaultKeyOperation{
					AzureKeyVaultKeyOperation_SIGN,
					AzureKeyVaultKeyOperation_VERIFY,
				}
				err := protovalidate.Validate(key(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an EC key without a curve (Azure defaults P-256)", func() {
				spec := validRsaSpec()
				spec.KeyType = AzureKeyVaultKeyType_EC_HSM
				spec.KeySize = nil
				spec.KeyOpts = []AzureKeyVaultKeyOperation{AzureKeyVaultKeyOperation_SIGN}
				err := protovalidate.Validate(key(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept activation and expiry timestamps", func() {
				spec := validRsaSpec()
				spec.NotBeforeDate = strPtr("2027-01-01T00:00:00Z")
				spec.ExpirationDate = strPtr("2028-01-01T00:00:00Z")
				err := protovalidate.Validate(key(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a rotation policy with expiry pair and automatic trigger", func() {
				spec := validRsaSpec()
				spec.RotationPolicy = &AzureKeyVaultKeyRotationPolicy{
					ExpireAfter:        strPtr("P90D"),
					NotifyBeforeExpiry: strPtr("P7D"),
					Automatic: &AzureKeyVaultKeyRotationPolicyAutomatic{
						TimeBeforeExpiry: strPtr("P30D"),
					},
				}
				err := protovalidate.Validate(key(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an automatic-only rotation policy", func() {
				spec := validRsaSpec()
				spec.RotationPolicy = &AzureKeyVaultKeyRotationPolicy{
					Automatic: &AzureKeyVaultKeyRotationPolicyAutomatic{
						TimeAfterCreation: strPtr("P83D"),
					},
				}
				err := protovalidate.Validate(key(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept user tags", func() {
				spec := validRsaSpec()
				spec.Tags = map[string]string{"purpose": "storage-cmk"}
				err := protovalidate.Validate(key(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_key_vault_key", func() {

			ginkgo.It("should return a validation error when name is missing", func() {
				spec := validRsaSpec()
				spec.Name = ""
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name carries invalid characters", func() {
				spec := validRsaSpec()
				spec.Name = "my_key"
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when key_vault_id is missing", func() {
				spec := validRsaSpec()
				spec.KeyVaultId = nil
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when key_type is unspecified", func() {
				spec := validRsaSpec()
				spec.KeyType = AzureKeyVaultKeyType_azure_key_vault_key_type_unspecified
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when an RSA key omits key_size", func() {
				spec := validRsaSpec()
				spec.KeySize = nil
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an unsupported RSA key size", func() {
				spec := validRsaSpec()
				spec.KeySize = int32Ptr(1024)
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when an EC key sets key_size", func() {
				spec := validRsaSpec()
				spec.KeyType = AzureKeyVaultKeyType_EC
				spec.KeySize = int32Ptr(2048)
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when an RSA key sets a curve", func() {
				spec := validRsaSpec()
				spec.Curve = AzureKeyVaultKeyCurve_P_256
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when key_opts is empty", func() {
				spec := validRsaSpec()
				spec.KeyOpts = nil
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when key_opts carries an unspecified operation", func() {
				spec := validRsaSpec()
				spec.KeyOpts = []AzureKeyVaultKeyOperation{AzureKeyVaultKeyOperation_azure_key_vault_key_operation_unspecified}
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed expiration_date", func() {
				spec := validRsaSpec()
				spec.ExpirationDate = strPtr("2028-01-01")
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an empty rotation policy", func() {
				spec := validRsaSpec()
				spec.RotationPolicy = &AzureKeyVaultKeyRotationPolicy{}
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when expire_after lacks notify_before_expiry", func() {
				spec := validRsaSpec()
				spec.RotationPolicy = &AzureKeyVaultKeyRotationPolicy{
					ExpireAfter: strPtr("P90D"),
				}
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed rotation duration", func() {
				spec := validRsaSpec()
				spec.RotationPolicy = &AzureKeyVaultKeyRotationPolicy{
					ExpireAfter:        strPtr("90days"),
					NotifyBeforeExpiry: strPtr("P7D"),
				}
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an automatic block without a trigger", func() {
				spec := validRsaSpec()
				spec.RotationPolicy = &AzureKeyVaultKeyRotationPolicy{
					Automatic: &AzureKeyVaultKeyRotationPolicyAutomatic{},
				}
				gomega.Expect(protovalidate.Validate(key(spec))).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := key(validRsaSpec())
				input.Spec = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})

// Helper functions for pointer types
func int32Ptr(i int32) *int32 {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}
