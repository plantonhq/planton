package azurediskencryptionsetv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureDiskEncryptionSetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDiskEncryptionSetSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

func validResource() *AzureDiskEncryptionSet {
	return &AzureDiskEncryptionSet{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDiskEncryptionSet",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-des",
		},
		Spec: &AzureDiskEncryptionSetSpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "test-des",
			KeyVaultKeyId: ref("cmk-key"),
			Identity: &AzureDiskEncryptionSetIdentity{
				Type: AzureDiskEncryptionSetIdentityType_SYSTEM_ASSIGNED,
			},
		},
	}
}

var _ = ginkgo.Describe("AzureDiskEncryptionSetSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_disk_encryption_set", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity with identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureDiskEncryptionSetIdentity{
					Type:        AzureDiskEncryptionSetIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{ref("des-uai")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept auto key rotation enabled", func() {
				input := validResource()
				rotate := true
				input.Spec.AutoKeyRotationEnabled = &rotate
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept double-encryption type", func() {
				input := validResource()
				input.Spec.EncryptionType = AzureDiskEncryptionSetEncryptionType_ENCRYPTION_AT_REST_WITH_PLATFORM_AND_CUSTOMER_KEYS
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a valid federated client id", func() {
				input := validResource()
				input.Spec.FederatedClientId = "11111111-2222-3333-4444-555555555555"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_disk_encryption_set", func() {

			ginkgo.It("should return a validation error when the key is missing", func() {
				input := validResource()
				input.Spec.KeyVaultKeyId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the identity is missing", func() {
				input := validResource()
				input.Spec.Identity = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when USER_ASSIGNED has no identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureDiskEncryptionSetIdentity{
					Type: AzureDiskEncryptionSetIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when SYSTEM_ASSIGNED carries identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureDiskEncryptionSetIdentity{
					Type:        AzureDiskEncryptionSetIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{ref("stray-uai")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the identity type is unspecified", func() {
				input := validResource()
				input.Spec.Identity = &AzureDiskEncryptionSetIdentity{
					Type: AzureDiskEncryptionSetIdentityType_azure_disk_encryption_set_identity_type_unspecified,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a non-UUID federated client id", func() {
				input := validResource()
				input.Spec.FederatedClientId = "not-a-uuid"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is missing", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := validResource()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := validResource()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
