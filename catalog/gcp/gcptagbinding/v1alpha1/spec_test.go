package gcptagbindingv1alpha1

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
	ginkgo.RunSpecs(t, "GcpTagBindingSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: name}},
	}
}

var _ = ginkgo.Describe("GcpTagBindingSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// The smallest binding: a value bound to the provider's default project.
	minimal := func() *GcpTagBinding {
		return &GcpTagBinding{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpTagBinding",
			Metadata: &shared.CloudResourceMetadata{
				Name: "project-environment-prod",
			},
			Spec: &GcpTagBindingSpec{
				TagValue: litRef("tagValues/281476102962987"),
			},
		}
	}

	expectError := func(target *GcpTagBinding, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a value bound to the default project with no parent at all", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept the tag value by reference or in namespaced form", func() {
		target := minimal()
		target.Spec.TagValue = nameRef("environment-prod")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.TagValue = litRef("123456789012/environment/prod")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each parent arm alone: project by number, ID, or reference; folder; organization; resource name", func() {
		parents := []*GcpTagBindingParent{
			{ProjectId: litRef("123456789012")},
			{ProjectId: litRef("my-gcp-project-123")},
			{ProjectId: nameRef("landing-zone-project")},
			{FolderId: litRef("987654321098")},
			{FolderId: nameRef("environments")},
			{OrganizationId: "123456789012"},
			{ResourceName: "//compute.googleapis.com/projects/my-gcp-project-123/global/networks/shared-vpc"},
			{},
		}
		for _, parent := range parents {
			target := minimal()
			target.Spec.Parent = parent
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept a zonal and a regional resource with a location", func() {
		target := minimal()
		target.Spec.Parent = &GcpTagBindingParent{ResourceName: "//compute.googleapis.com/projects/my-gcp-project-123/zones/us-central1-a/instances/web-1"}
		target.Spec.Location = "us-central1-a"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.Parent = &GcpTagBindingParent{ResourceName: "//sqladmin.googleapis.com/projects/my-gcp-project-123/instances/orders-db"}
		target.Spec.Location = "us-central1"
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

	ginkgo.It("should reject a binding without a tag value or with a malformed literal", func() {
		target := minimal()
		target.Spec.TagValue = nil
		expectError(target, "tag_value")

		for _, v := range []string{"prod", "tagValues/prod", "environment/prod"} {
			target = minimal()
			target.Spec.TagValue = litRef(v)
			expectError(target, "tag_value must be tagValues/{numeric_id}")
		}
	})

	ginkgo.It("should reject two parent arms", func() {
		target := minimal()
		target.Spec.Parent = &GcpTagBindingParent{ProjectId: litRef("123456789012"), OrganizationId: "123456789012"}
		expectError(target, "at most one of project_id, folder_id, organization_id, or resource_name")
	})

	ginkgo.It("should reject a non-numeric organization ID and a resource name that is not a full resource name", func() {
		target := minimal()
		target.Spec.Parent = &GcpTagBindingParent{OrganizationId: "organizations/123456789012"}
		expectError(target, "numeric organization ID")

		target.Spec.Parent = &GcpTagBindingParent{ResourceName: "projects/my-gcp-project-123/zones/us-central1-a/instances/web-1"}
		expectError(target, "full resource name beginning with //")
	})

	ginkgo.It("should reject a location on a global parent and a malformed location", func() {
		target := minimal()
		target.Spec.Location = "us-central1"
		expectError(target, "location applies only to a regional or zonal resource")

		target.Spec.Parent = &GcpTagBindingParent{FolderId: litRef("987654321098")}
		expectError(target, "location applies only to a regional or zonal resource")

		target.Spec.Parent = &GcpTagBindingParent{ResourceName: "//compute.googleapis.com/projects/p/zones/us-central1-a/instances/web-1"}
		target.Spec.Location = "US Central"
		expectError(target, "location must be a region")
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "RETAIN"
		expectError(target, "deletion_policy must be one of")
	})
})
