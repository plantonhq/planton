package gcpprojectiammemberv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpProjectIamMemberSpec Suite")
}

func valueOf(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
	}
}

var _ = ginkgo.Describe("GcpProjectIamMemberSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpProjectIamMember.
	minimal := func() *GcpProjectIamMember {
		return &GcpProjectIamMember{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpProjectIamMember",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-iam-member",
			},
			Spec: &GcpProjectIamMemberSpec{
				Role:   valueOf("roles/storage.objectViewer"),
				Member: valueOf("serviceAccount:my-sa@my-project.iam.gserviceaccount.com"),
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal grant with predefined role and service account member", func() {
		err := validator.Validate(minimal())
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a grant with an explicit project and custom role name", func() {
		msg := minimal()
		msg.Spec.ProjectId = valueOf("my-gcp-project")
		msg.Spec.Role = valueOf("projects/my-gcp-project/roles/logBucketWriter")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a grant with a full condition", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpProjectIamMemberCondition{
			Title:       "expires-2027",
			Expression:  `request.time < timestamp("2027-01-01T00:00:00Z")`,
			Description: "Temporary access for the migration",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a condition without a description", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpProjectIamMemberCondition{
			Title:      "prod-buckets-only",
			Expression: `resource.name.startsWith("projects/_/buckets/prod-")`,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing role", func() {
		msg := minimal()
		msg.Spec.Role = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing member", func() {
		msg := minimal()
		msg.Spec.Member = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a condition without a title", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpProjectIamMemberCondition{
			Expression: `request.time < timestamp("2027-01-01T00:00:00Z")`,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a condition without an expression", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpProjectIamMemberCondition{
			Title: "expires-2027",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a condition title longer than 100 characters", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpProjectIamMemberCondition{
			Title:      strings.Repeat("t", 101),
			Expression: "true",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a condition description longer than 256 characters", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpProjectIamMemberCondition{
			Title:       "expires-2027",
			Expression:  "true",
			Description: strings.Repeat("d", 257),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
