package azurecontainerappcustomdomainv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureContainerAppCustomDomainSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureContainerAppCustomDomainSpec Validation Tests")
}

// ref builds a StringValueOrRef carrying a value_from reference.
func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

// validResource returns a minimal valid managed-certificate-flow binding
// that individual cases then mutate into the shape under test.
func validResource() *AzureContainerAppCustomDomain {
	return &AzureContainerAppCustomDomain{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureContainerAppCustomDomain",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "app-custom-domain",
		},
		Spec: &AzureContainerAppCustomDomainSpec{
			DomainName:     "app.example.com",
			ContainerAppId: ref("my-app"),
		},
	}
}

var _ = ginkgo.Describe("AzureContainerAppCustomDomainSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_container_app_custom_domain", func() {

			ginkgo.It("should not return a validation error for the managed-certificate flow (no certificate fields)", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a bring-your-own certificate with SNI binding", func() {
				input := validResource()
				input.Spec.ContainerAppEnvironmentCertificateId = ref("my-byo-cert")
				input.Spec.CertificateBindingType = AzureContainerAppCustomDomainCertificateBindingType_SNI_ENABLED
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a certificate with the DISABLED binding type", func() {
				input := validResource()
				input.Spec.ContainerAppEnvironmentCertificateId = ref("my-byo-cert")
				input.Spec.CertificateBindingType = AzureContainerAppCustomDomainCertificateBindingType_DISABLED
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a deeply nested domain name", func() {
				input := validResource()
				input.Spec.DomainName = "api.v2.app.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_container_app_custom_domain", func() {

			ginkgo.It("should return a validation error when domain_name is missing", func() {
				input := validResource()
				input.Spec.DomainName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a wildcard domain name", func() {
				input := validResource()
				input.Spec.DomainName = "*.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a single-label domain name", func() {
				input := validResource()
				input.Spec.DomainName = "app"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when container_app_id is missing", func() {
				input := validResource()
				input.Spec.ContainerAppId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a certificate without its binding type", func() {
				input := validResource()
				input.Spec.ContainerAppEnvironmentCertificateId = ref("my-byo-cert")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a binding type without a certificate", func() {
				input := validResource()
				input.Spec.CertificateBindingType = AzureContainerAppCustomDomainCertificateBindingType_SNI_ENABLED
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an undefined binding-type enum value", func() {
				input := validResource()
				input.Spec.ContainerAppEnvironmentCertificateId = ref("my-byo-cert")
				input.Spec.CertificateBindingType = AzureContainerAppCustomDomainCertificateBindingType(99)
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
