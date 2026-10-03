package gcpgkeworkloadidentitybindingv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestGcpGkeWorkloadIdentityBindingSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpGkeWorkloadIdentityBindingSpec Validation Tests")
}

// literal wraps a string in a StringValueOrRef literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// baseBinding returns a valid minimal binding that individual cases mutate.
func baseBinding() *GcpGkeWorkloadIdentityBinding {
	return &GcpGkeWorkloadIdentityBinding{
		ApiVersion: "gcp.planton.dev/v1alpha1",
		Kind:       "GcpGkeWorkloadIdentityBinding",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-workload-identity-binding",
		},
		Spec: &GcpGkeWorkloadIdentityBindingSpec{
			ProjectId:           literal("test-project-123"),
			ServiceAccountEmail: literal("cert-manager@test-project-123.iam.gserviceaccount.com"),
			KsaNamespace:        "cert-manager",
			KsaName:             "cert-manager",
		},
	}
}

var _ = ginkgo.Describe("GcpGkeWorkloadIdentityBindingSpec Validation Tests", func() {

	ginkgo.Describe("Valid configurations", func() {

		ginkgo.It("should accept a minimal binding", func() {
			gomega.Expect(protovalidate.Validate(baseBinding())).To(gomega.BeNil())
		})

		ginkgo.It("should accept a KSA name with dots (DNS subdomain rules)", func() {
			input := baseBinding()
			input.Spec.KsaName = "workload.v2"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a binding with an IAM condition", func() {
			input := baseBinding()
			input.Spec.Condition = &GcpGkeWorkloadIdentityBindingCondition{
				Title:       "expires-2026-12-31",
				Expression:  `request.time < timestamp("2027-01-01T00:00:00Z")`,
				Description: "temporary grant for the migration window",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept single-character namespace and name", func() {
			input := baseBinding()
			input.Spec.KsaNamespace = "a"
			input.Spec.KsaName = "b"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Invalid configurations", func() {

		ginkgo.It("should accept an omitted project_id (ambient pool project)", func() {
			input := baseBinding()
			input.Spec.ProjectId = nil
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should reject a missing service_account_email", func() {
			input := baseBinding()
			input.Spec.ServiceAccountEmail = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a missing ksa_namespace", func() {
			input := baseBinding()
			input.Spec.KsaNamespace = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a missing ksa_name", func() {
			input := baseBinding()
			input.Spec.KsaName = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an uppercase namespace (RFC 1123 label)", func() {
			input := baseBinding()
			input.Spec.KsaNamespace = "CertManager"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a namespace with dots (labels take no dots)", func() {
			input := baseBinding()
			input.Spec.KsaNamespace = "kube.system"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a namespace ending with a hyphen", func() {
			input := baseBinding()
			input.Spec.KsaNamespace = "cert-manager-"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a namespace longer than 63 characters", func() {
			input := baseBinding()
			input.Spec.KsaNamespace = "a123456789a123456789a123456789a123456789a123456789a123456789abcd"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a KSA name with underscores", func() {
			input := baseBinding()
			input.Spec.KsaName = "cert_manager"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a condition without a title", func() {
			input := baseBinding()
			input.Spec.Condition = &GcpGkeWorkloadIdentityBindingCondition{
				Expression: `request.time < timestamp("2027-01-01T00:00:00Z")`,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a condition without an expression", func() {
			input := baseBinding()
			input.Spec.Condition = &GcpGkeWorkloadIdentityBindingCondition{
				Title: "expires-2026-12-31",
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
