package gcpprojectv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestGcpProjectSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpProjectSpec Validation Tests")
}

// baseProject returns a valid minimal project that individual cases mutate.
func baseProject() *GcpProject {
	return &GcpProject{
		ApiVersion: "gcp.planton.dev/v1alpha1",
		Kind:       "GcpProject",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-gcp-project",
		},
		Spec: &GcpProjectSpec{
			ProjectId:  "my-prod-project-123",
			ParentType: GcpProjectParentType_organization,
			ParentId:   "123456789012",
		},
	}
}

var _ = ginkgo.Describe("GcpProjectSpec Validation Tests", func() {

	ginkgo.Describe("Valid configurations", func() {

		ginkgo.It("should accept a minimal project", func() {
			gomega.Expect(protovalidate.Validate(baseProject())).To(gomega.BeNil())
		})

		ginkgo.It("should accept a project under a folder", func() {
			input := baseProject()
			input.Spec.ParentType = GcpProjectParentType_folder
			input.Spec.ParentId = "987654321"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a folder by reference with parent_type omitted or set to folder", func() {
			input := baseProject()
			input.Spec.ParentType = GcpProjectParentType_gcp_project_parent_type_unspecified
			input.Spec.ParentId = ""
			input.Spec.FolderId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: "environments"}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

			input.Spec.ParentType = GcpProjectParentType_folder
			input.Spec.FolderId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "987654321"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a billing account", func() {
			input := baseProject()
			input.Spec.BillingAccountId = "0123AB-4567CD-89EFGH"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a display name, labels, and tags", func() {
			input := baseProject()
			input.Spec.DisplayName = "Production Workloads"
			input.Spec.Labels = map[string]string{"team": "platform", "cost-center": "eng"}
			input.Spec.Tags = map[string]string{"tagKeys/123456789": "tagValues/987654321"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept enabled APIs", func() {
			input := baseProject()
			input.Spec.EnabledApis = []string{"compute.googleapis.com", "storage.googleapis.com"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an explicit auto_create_network", func() {
			autoCreate := true
			input := baseProject()
			input.Spec.AutoCreateNetwork = &autoCreate
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept every deletion policy", func() {
			for _, policy := range []string{"DELETE", "PREVENT", "ABANDON"} {
				input := baseProject()
				input.Spec.DeletionPolicy = policy
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "deletion_policy %s should be accepted", policy)
			}
		})

		ginkgo.It("should accept a six-character project ID (minimum length)", func() {
			input := baseProject()
			input.Spec.ProjectId = "abc-12"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Invalid configurations", func() {

		ginkgo.It("should reject a missing project_id", func() {
			input := baseProject()
			input.Spec.ProjectId = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a project_id shorter than 6 characters", func() {
			input := baseProject()
			input.Spec.ProjectId = "abc12"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a project_id longer than 30 characters", func() {
			input := baseProject()
			input.Spec.ProjectId = "a123456789b123456789c123456789d"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a project_id starting with a digit", func() {
			input := baseProject()
			input.Spec.ProjectId = "1my-project"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a project_id ending with a hyphen", func() {
			input := baseProject()
			input.Spec.ProjectId = "my-project-"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a project_id with uppercase letters", func() {
			input := baseProject()
			input.Spec.ProjectId = "My-Project-123"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a non-numeric parent_id", func() {
			input := baseProject()
			input.Spec.ParentId = "my-folder"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject folder_id beside parent_id, or beside parent_type organization", func() {
			input := baseProject()
			input.Spec.FolderId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "987654321"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("when folder_id is set it IS the parent"))

			input.Spec.ParentId = ""
			input.Spec.ParentType = GcpProjectParentType_organization
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed billing account ID", func() {
			input := baseProject()
			input.Spec.BillingAccountId = "not-a-billing-account"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an API entry without the googleapis.com suffix", func() {
			input := baseProject()
			input.Spec.EnabledApis = []string{"compute"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid deletion policy", func() {
			input := baseProject()
			input.Spec.DeletionPolicy = "PROTECT"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
