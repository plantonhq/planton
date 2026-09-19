package gcptagkeyv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpTagKeySpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpTagKeySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpTagKey {
		return &GcpTagKey{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpTagKey",
			Metadata: &shared.CloudResourceMetadata{
				Name: "environment",
			},
			Spec: &GcpTagKeySpec{
				Parent: &GcpTagKeyParent{OrganizationId: "123456789012"},
			},
		}
	}

	expectError := func(target *GcpTagKey, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept an organization-owned key named from metadata", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project-owned key by literal or by reference", func() {
		target := minimal()
		target.Spec.Parent = &GcpTagKeyParent{ProjectId: litRef("my-gcp-project-123")}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.Parent = &GcpTagKeyParent{ProjectId: &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: "landing-zone-project"}},
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept short names with spaces, dashes, dots, and unicode", func() {
		for _, name := range []string{"environment", "cost-center", "team payments", "data.classification", "umgebung-ü"} {
			target := minimal()
			target.Spec.ShortName = name
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), name)
		}
	})

	ginkgo.It("should accept a firewall-purpose key with its network, a governance key, and a dynamic key", func() {
		target := minimal()
		target.Spec.Purpose = "GCE_FIREWALL"
		target.Spec.PurposeData = map[string]string{"network": "my-gcp-project-123/shared-vpc"}
		target.Spec.Description = "Secure tags for firewall policy rules"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target = minimal()
		target.Spec.Purpose = "DATA_GOVERNANCE"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target = minimal()
		target.Spec.AllowedValuesRegex = "^[A-Z]{2,4}-[0-9]{1,6}$"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every deletion_policy value and the empty default", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), policy)
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a key without a parent, with both arms, or with neither", func() {
		target := minimal()
		target.Spec.Parent = nil
		expectError(target, "parent")

		target.Spec.Parent = &GcpTagKeyParent{OrganizationId: "123456789012", ProjectId: litRef("my-gcp-project-123")}
		expectError(target, "exactly one of organization_id or project_id")

		target.Spec.Parent = &GcpTagKeyParent{}
		expectError(target, "exactly one of organization_id or project_id")
	})

	ginkgo.It("should reject a non-numeric organization ID", func() {
		target := minimal()
		target.Spec.Parent.OrganizationId = "organizations/123456789012"
		expectError(target, "numeric organization ID")
	})

	ginkgo.It("should reject short names with the characters Google forbids", func() {
		for _, name := range []string{"env/prod", `env\prod`, "env'prod", `env"prod`} {
			target := minimal()
			target.Spec.ShortName = name
			expectError(target, "must not contain a slash, backslash, or quote")
		}
	})

	ginkgo.It("should reject an over-long description and an unknown purpose", func() {
		target := minimal()
		target.Spec.Description = string(make([]byte, 257))
		expectError(target, "description")

		target = minimal()
		target.Spec.Purpose = "BILLING"
		expectError(target, "purpose must be empty, GCE_FIREWALL, or DATA_GOVERNANCE")
	})

	ginkgo.It("should reject purpose_data without a purpose", func() {
		target := minimal()
		target.Spec.PurposeData = map[string]string{"network": "my-gcp-project-123/shared-vpc"}
		expectError(target, "purpose_data is only meaningful with a purpose")
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "RETAIN"
		expectError(target, "deletion_policy must be one of")
	})
})
