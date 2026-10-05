package azurecontainerappenvironmentmanagedcertificatev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureContainerAppEnvironmentManagedCertificateSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureContainerAppEnvironmentManagedCertificateSpec Validation Tests")
}

// ref builds a StringValueOrRef carrying a value_from reference.
func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

// validResource returns a minimal valid managed certificate that
// individual cases then mutate into the shape under test.
func validResource() *AzureContainerAppEnvironmentManagedCertificate {
	return &AzureContainerAppEnvironmentManagedCertificate{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureContainerAppEnvironmentManagedCertificate",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "app-managed-cert",
		},
		Spec: &AzureContainerAppEnvironmentManagedCertificateSpec{
			CertificateName:           "app-example-com",
			ContainerAppEnvironmentId: ref("my-environment"),
			SubjectName:               "app.example.com",
		},
	}
}

var _ = ginkgo.Describe("AzureContainerAppEnvironmentManagedCertificateSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_container_app_environment_managed_certificate", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept CNAME domain-control validation", func() {
				input := validResource()
				input.Spec.DomainControlValidation = AzureContainerAppManagedCertificateDomainControlValidation_CNAME
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept HTTP domain-control validation and tags", func() {
				input := validResource()
				input.Spec.DomainControlValidation = AzureContainerAppManagedCertificateDomainControlValidation_HTTP
				input.Spec.Tags = map[string]string{"team": "web"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a deeply nested subject name", func() {
				input := validResource()
				input.Spec.SubjectName = "api.v2.app.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_container_app_environment_managed_certificate", func() {

			ginkgo.It("should return a validation error when certificate_name is missing", func() {
				input := validResource()
				input.Spec.CertificateName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a certificate name with consecutive hyphens", func() {
				input := validResource()
				input.Spec.CertificateName = "app--cert"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when container_app_environment_id is missing", func() {
				input := validResource()
				input.Spec.ContainerAppEnvironmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subject_name is missing", func() {
				input := validResource()
				input.Spec.SubjectName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a wildcard subject name", func() {
				input := validResource()
				input.Spec.SubjectName = "*.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a single-label subject name", func() {
				input := validResource()
				input.Spec.SubjectName = "localhost"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an undefined validation enum value", func() {
				input := validResource()
				input.Spec.DomainControlValidation = AzureContainerAppManagedCertificateDomainControlValidation(99)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := validResource()
				input.Kind = "WrongKind"
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
