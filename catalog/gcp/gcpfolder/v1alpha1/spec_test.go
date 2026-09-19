package gcpfolderv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpFolderSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpFolderSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpFolder {
		return &GcpFolder{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpFolder",
			Metadata: &shared.CloudResourceMetadata{
				Name: "production",
			},
			Spec: &GcpFolderSpec{
				Parent: &GcpFolderParent{OrganizationId: "123456789012"},
			},
		}
	}

	expectError := func(target *GcpFolder, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a top-level folder under the organization", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a nested folder whose parent is another folder by literal or by reference", func() {
		target := minimal()
		target.Spec.Parent = &GcpFolderParent{FolderId: litRef("987654321098")}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.Parent = &GcpFolderParent{FolderId: &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: "environments"}},
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every legal display name shape", func() {
		for _, name := range []string{"abc", "Production", "team payments", "team-payments_2", "a12345678901234567890123456789"} {
			target := minimal()
			target.Spec.DisplayName = name
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), name)
		}
	})

	ginkgo.It("should accept the destroy guard in both positions, create-time tags, and every deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionProtection = proto.Bool(false)
		target.Spec.Tags = map[string]string{"tagKeys/281475647562788": "tagValues/281476102962987"}
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), policy)
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a folder without a parent", func() {
		target := minimal()
		target.Spec.Parent = nil
		expectError(target, "parent")
	})

	ginkgo.It("should reject a parent with both arms or neither", func() {
		target := minimal()
		target.Spec.Parent = &GcpFolderParent{OrganizationId: "123456789012", FolderId: litRef("987654321098")}
		expectError(target, "exactly one of organization_id or folder_id")

		target.Spec.Parent = &GcpFolderParent{}
		expectError(target, "exactly one of organization_id or folder_id")
	})

	ginkgo.It("should reject a non-numeric organization ID", func() {
		target := minimal()
		target.Spec.Parent.OrganizationId = "organizations/123456789012"
		expectError(target, "numeric organization ID")
	})

	ginkgo.It("should reject display names Google rejects", func() {
		for _, name := range []string{"ab", "-production", "production-", "prod/uction", "a123456789012345678901234567890"} {
			target := minimal()
			target.Spec.DisplayName = name
			expectError(target, "display_name must be 3-30 characters")
		}
	})

	ginkgo.It("should reject tags that are not tagKeys/{id} -> tagValues/{id}", func() {
		target := minimal()
		target.Spec.Tags = map[string]string{"environment": "tagValues/281476102962987"}
		expectError(target, "tags")

		target.Spec.Tags = map[string]string{"tagKeys/281475647562788": "prod"}
		expectError(target, "tags")
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "RETAIN"
		expectError(target, "deletion_policy must be one of")
	})
})
