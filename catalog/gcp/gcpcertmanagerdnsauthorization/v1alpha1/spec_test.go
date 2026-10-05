package gcpcertmanagerdnsauthorizationv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestGcpCertManagerDnsAuthorizationSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCertManagerDnsAuthorizationSpec Validation Tests")
}

// literal wraps a string in a StringValueOrRef literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// baseAuthorization returns a valid minimal authorization that cases mutate.
func baseAuthorization() *GcpCertManagerDnsAuthorization {
	return &GcpCertManagerDnsAuthorization{
		ApiVersion: "gcp.planton.dev/v1alpha1",
		Kind:       "GcpCertManagerDnsAuthorization",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-dns-authorization",
		},
		Spec: &GcpCertManagerDnsAuthorizationSpec{
			ProjectId: literal("test-project-123"),
			Domain:    "example.com",
		},
	}
}

var _ = ginkgo.Describe("GcpCertManagerDnsAuthorizationSpec Validation Tests", func() {

	ginkgo.Describe("Valid configurations", func() {

		ginkgo.It("should accept a minimal authorization", func() {
			gomega.Expect(protovalidate.Validate(baseAuthorization())).To(gomega.BeNil())
		})

		ginkgo.It("should accept an omitted project_id (ambient project)", func() {
			input := baseAuthorization()
			input.Spec.ProjectId = nil
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an explicit authorization name", func() {
			input := baseAuthorization()
			input.Spec.AuthorizationName = "prod_example-com-auth"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a subdomain", func() {
			input := baseAuthorization()
			input.Spec.Domain = "api.internal.example.co.uk"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept PER_PROJECT_RECORD type with labels and location", func() {
			input := baseAuthorization()
			input.Spec.Type = "PER_PROJECT_RECORD"
			input.Spec.Location = "us-central1"
			input.Spec.Labels = map[string]string{"team": "platform"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept each deletion_policy value", func() {
			for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
				input := baseAuthorization()
				input.Spec.DeletionPolicy = v
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})
	})

	ginkgo.Describe("Invalid configurations", func() {

		ginkgo.It("should reject a missing domain", func() {
			input := baseAuthorization()
			input.Spec.Domain = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a wildcard domain (the authorization covers the wildcard implicitly)", func() {
			input := baseAuthorization()
			input.Spec.Domain = "*.example.com"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a domain with a trailing dot", func() {
			input := baseAuthorization()
			input.Spec.Domain = "example.com."
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an authorization name starting with a digit", func() {
			input := baseAuthorization()
			input.Spec.AuthorizationName = "1-auth"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid type", func() {
			input := baseAuthorization()
			input.Spec.Type = "SHARED_RECORD"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid deletion_policy", func() {
			input := baseAuthorization()
			input.Spec.DeletionPolicy = "KEEP"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
