package azurecontainerappenvironmentcertificatev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureContainerAppEnvironmentCertificateSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureContainerAppEnvironmentCertificateSpec Validation Tests")
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

// validResource returns a minimal valid inline-PFX certificate that
// individual cases then mutate into the shape under test.
func validResource() *AzureContainerAppEnvironmentCertificate {
	return &AzureContainerAppEnvironmentCertificate{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureContainerAppEnvironmentCertificate",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "app-tls-cert",
		},
		Spec: &AzureContainerAppEnvironmentCertificateSpec{
			CertificateName:           "app.example.com",
			ContainerAppEnvironmentId: ref("my-environment"),
			CertificateBlobBase64:     "TUlJRXZnSUJBREFOQmdrcWhraUc5dzBCQVFFRkFBU0NCS2c=",
			CertificatePassword:       "pfx-password",
		},
	}
}

var _ = ginkgo.Describe("AzureContainerAppEnvironmentCertificateSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_container_app_environment_certificate", func() {

			ginkgo.It("should not return a validation error for a minimal inline-PFX certificate", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a Key Vault-sourced certificate", func() {
				input := validResource()
				input.Spec.CertificateBlobBase64 = ""
				input.Spec.CertificatePassword = ""
				input.Spec.CertificateKeyVault = &AzureContainerAppEnvironmentCertificateKeyVault{
					KeyVaultSecretId: ref("my-kv-certificate"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a Key Vault certificate with a user-assigned identity", func() {
				input := validResource()
				input.Spec.CertificateBlobBase64 = ""
				input.Spec.CertificatePassword = ""
				input.Spec.CertificateKeyVault = &AzureContainerAppEnvironmentCertificateKeyVault{
					KeyVaultSecretId: literal("https://my-vault.vault.azure.net/secrets/app-cert"),
					Identity:         ref("env-identity"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a hyphenated certificate name and tags", func() {
				input := validResource()
				input.Spec.CertificateName = "wildcard-example-com"
				input.Spec.Tags = map[string]string{"team": "platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a passwordless PFX (blob with an empty password)", func() {
				input := validResource()
				input.Spec.CertificatePassword = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_container_app_environment_certificate", func() {

			ginkgo.It("should return a validation error when certificate_name is missing", func() {
				input := validResource()
				input.Spec.CertificateName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a certificate name with uppercase letters", func() {
				input := validResource()
				input.Spec.CertificateName = "App.Example.Com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a certificate name with consecutive hyphens", func() {
				input := validResource()
				input.Spec.CertificateName = "app--cert"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a certificate name ending in a hyphen", func() {
				input := validResource()
				input.Spec.CertificateName = "app-cert-"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when container_app_environment_id is missing", func() {
				input := validResource()
				input.Spec.ContainerAppEnvironmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when neither blob nor key vault is set", func() {
				input := validResource()
				input.Spec.CertificateBlobBase64 = ""
				input.Spec.CertificatePassword = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both blob and key vault are set", func() {
				input := validResource()
				input.Spec.CertificateKeyVault = &AzureContainerAppEnvironmentCertificateKeyVault{
					KeyVaultSecretId: ref("my-kv-certificate"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a password without a blob", func() {
				input := validResource()
				input.Spec.CertificateBlobBase64 = ""
				input.Spec.CertificateKeyVault = &AzureContainerAppEnvironmentCertificateKeyVault{
					KeyVaultSecretId: ref("my-kv-certificate"),
				}
				// certificate_password kept from the valid resource -- it
				// cannot accompany the Key Vault source.
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the key vault block lacks its secret id", func() {
				input := validResource()
				input.Spec.CertificateBlobBase64 = ""
				input.Spec.CertificatePassword = ""
				input.Spec.CertificateKeyVault = &AzureContainerAppEnvironmentCertificateKeyVault{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := validResource()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := validResource()
				input.Spec = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
