package gcpserviceaccountiammemberv1alpha1

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
	ginkgo.RunSpecs(t, "GcpServiceAccountIamMemberSpec Suite")
}

func valueOf(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
	}
}

var _ = ginkgo.Describe("GcpServiceAccountIamMemberSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpServiceAccountIamMember.
	minimal := func() *GcpServiceAccountIamMember {
		return &GcpServiceAccountIamMember{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpServiceAccountIamMember",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-sa-iam-member",
			},
			Spec: &GcpServiceAccountIamMemberSpec{
				ServiceAccountId: valueOf("projects/my-project/serviceAccounts/deployer@my-project.iam.gserviceaccount.com"),
				Role:             valueOf("roles/iam.workloadIdentityUser"),
				Member:           valueOf("principalSet://iam.googleapis.com/projects/123456/locations/global/workloadIdentityPools/github/attribute.repository/my-org/my-repo"),
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal impersonation grant to a federated principal set", func() {
		err := validator.Validate(minimal())
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a token-creator grant to another service account", func() {
		msg := minimal()
		msg.Spec.Role = valueOf("roles/iam.serviceAccountTokenCreator")
		msg.Spec.Member = valueOf("serviceAccount:caller@my-project.iam.gserviceaccount.com")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an actAs grant to a user with a custom role name", func() {
		msg := minimal()
		msg.Spec.Role = valueOf("projects/my-project/roles/limitedActAs")
		msg.Spec.Member = valueOf("user:dev@example.com")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a grant with a full condition", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpServiceAccountIamMemberCondition{
			Title:       "expires-2027",
			Expression:  `request.time < timestamp("2027-01-01T00:00:00Z")`,
			Description: "Temporary impersonation for the migration",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a condition without a description", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpServiceAccountIamMemberCondition{
			Title:      "expires-2027",
			Expression: `request.time < timestamp("2027-01-01T00:00:00Z")`,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing service account id", func() {
		msg := minimal()
		msg.Spec.ServiceAccountId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

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
		msg.Spec.Condition = &GcpServiceAccountIamMemberCondition{
			Expression: `request.time < timestamp("2027-01-01T00:00:00Z")`,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a condition without an expression", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpServiceAccountIamMemberCondition{
			Title: "expires-2027",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a condition title longer than 100 characters", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpServiceAccountIamMemberCondition{
			Title:      strings.Repeat("t", 101),
			Expression: "true",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a condition description longer than 256 characters", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpServiceAccountIamMemberCondition{
			Title:       "expires-2027",
			Expression:  "true",
			Description: strings.Repeat("d", 257),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind literal", func() {
		msg := minimal()
		msg.Kind = "GcpProjectIamMember"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing metadata", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
