package ocikmsvaultv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciKmsVaultSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciKmsVaultSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidDefaultVault() *OciKmsVault {
	return &OciKmsVault{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciKmsVault",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-vault",
		},
		Spec: &OciKmsVaultSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			VaultType:     OciKmsVaultSpec_default_vault,
		},
	}
}

func minimalValidExternalVault() *OciKmsVault {
	return &OciKmsVault{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciKmsVault",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-external-vault",
		},
		Spec: &OciKmsVaultSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			VaultType:     OciKmsVaultSpec_external,
			ExternalKeyManagerMetadata: &OciKmsVaultSpec_ExternalKeyManagerMetadata{
				ExternalVaultEndpointUrl: "https://ekm.example.com/vault/keys",
				OauthMetadata: &OciKmsVaultSpec_ExternalKeyManagerMetadata_OAuthMetadata{
					ClientAppId:        "app-id-12345",
					ClientAppSecret:    "secret-value",
					IdcsAccountNameUrl: "https://idcs-xxx.identity.oraclecloud.com",
				},
				PrivateEndpointId: "ocid1.kmsekmsendpoint.oc1..example",
			},
		},
	}
}

var _ = ginkgo.Describe("OciKmsVaultSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_kms_vault", func() {

			ginkgo.It("should not return a validation error for minimal default_vault", func() {
				input := minimalValidDefaultVault()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for virtual_private vault", func() {
				input := minimalValidDefaultVault()
				input.Spec.VaultType = OciKmsVaultSpec_virtual_private
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for external vault with full metadata", func() {
				input := minimalValidExternalVault()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display_name set", func() {
				input := minimalValidDefaultVault()
				input.Spec.DisplayName = "my-encryption-vault"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidDefaultVault()
				input.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-compartment",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for full default_vault configuration", func() {
				input := minimalValidDefaultVault()
				input.Spec.DisplayName = "production-vault"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for full external vault configuration", func() {
				input := minimalValidExternalVault()
				input.Spec.DisplayName = "enterprise-ekms-vault"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_kms_vault", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidDefaultVault()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidDefaultVault()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidDefaultVault()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciKmsVault{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciKmsVault",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-vault"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidDefaultVault()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vault_type is unspecified", func() {
				input := minimalValidDefaultVault()
				input.Spec.VaultType = OciKmsVaultSpec_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for external vault without metadata", func() {
				input := minimalValidDefaultVault()
				input.Spec.VaultType = OciKmsVaultSpec_external
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for default_vault with metadata present", func() {
				input := minimalValidExternalVault()
				input.Spec.VaultType = OciKmsVaultSpec_default_vault
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for virtual_private with metadata present", func() {
				input := minimalValidExternalVault()
				input.Spec.VaultType = OciKmsVaultSpec_virtual_private
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for external vault with empty endpoint url", func() {
				input := minimalValidExternalVault()
				input.Spec.ExternalKeyManagerMetadata.ExternalVaultEndpointUrl = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for external vault with missing oauth_metadata", func() {
				input := minimalValidExternalVault()
				input.Spec.ExternalKeyManagerMetadata.OauthMetadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for external vault with empty private_endpoint_id", func() {
				input := minimalValidExternalVault()
				input.Spec.ExternalKeyManagerMetadata.PrivateEndpointId = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for external vault with empty client_app_id", func() {
				input := minimalValidExternalVault()
				input.Spec.ExternalKeyManagerMetadata.OauthMetadata.ClientAppId = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for external vault with empty client_app_secret", func() {
				input := minimalValidExternalVault()
				input.Spec.ExternalKeyManagerMetadata.OauthMetadata.ClientAppSecret = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for external vault with empty idcs_account_name_url", func() {
				input := minimalValidExternalVault()
				input.Spec.ExternalKeyManagerMetadata.OauthMetadata.IdcsAccountNameUrl = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
