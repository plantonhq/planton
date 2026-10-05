package gcpkmskeyiammemberv1alpha1

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
	ginkgo.RunSpecs(t, "GcpKmsKeyIamMemberSpec Suite")
}

func valueOf(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
	}
}

var _ = ginkgo.Describe("GcpKmsKeyIamMemberSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpKmsKeyIamMember.
	minimal := func() *GcpKmsKeyIamMember {
		return &GcpKmsKeyIamMember{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpKmsKeyIamMember",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-kms-iam-member",
			},
			Spec: &GcpKmsKeyIamMemberSpec{
				CryptoKeyId: valueOf("projects/my-project/locations/us-central1/keyRings/app-ring/cryptoKeys/state-key"),
				Role:        valueOf("roles/cloudkms.cryptoKeyEncrypterDecrypter"),
				Member:      valueOf("serviceAccount:service-123456@gs-project-accounts.iam.gserviceaccount.com"),
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal CMEK grant to a service agent", func() {
		err := validator.Validate(minimal())
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a viewer grant to a group", func() {
		msg := minimal()
		msg.Spec.Role = valueOf("roles/cloudkms.viewer")
		msg.Spec.Member = valueOf("group:security-team@example.com")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a custom role grant", func() {
		msg := minimal()
		msg.Spec.Role = valueOf("projects/my-project/roles/kmsKeyAuditor")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a grant with a full condition", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpKmsKeyIamMemberCondition{
			Title:       "expires-2027",
			Expression:  `request.time < timestamp("2027-01-01T00:00:00Z")`,
			Description: "Temporary access for the migration",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a condition without a description", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpKmsKeyIamMemberCondition{
			Title:      "expires-2027",
			Expression: `request.time < timestamp("2027-01-01T00:00:00Z")`,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing crypto key id", func() {
		msg := minimal()
		msg.Spec.CryptoKeyId = nil
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
		msg.Spec.Condition = &GcpKmsKeyIamMemberCondition{
			Expression: `request.time < timestamp("2027-01-01T00:00:00Z")`,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a condition without an expression", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpKmsKeyIamMemberCondition{
			Title: "expires-2027",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a condition title longer than 100 characters", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpKmsKeyIamMemberCondition{
			Title:      strings.Repeat("t", 101),
			Expression: "true",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a condition description longer than 256 characters", func() {
		msg := minimal()
		msg.Spec.Condition = &GcpKmsKeyIamMemberCondition{
			Title:       "expires-2027",
			Expression:  "true",
			Description: strings.Repeat("d", 257),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind literal", func() {
		msg := minimal()
		msg.Kind = "GcpKmsKey"
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
