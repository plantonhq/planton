package gcpiamcustomrolev1alpha1

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
	ginkgo.RunSpecs(t, "GcpIamCustomRoleSpec Suite")
}

func ptr(s string) *string {
	return &s
}

var _ = ginkgo.Describe("GcpIamCustomRoleSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpIamCustomRole.
	minimal := func() *GcpIamCustomRole {
		return &GcpIamCustomRole{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpIamCustomRole",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-custom-role",
			},
			Spec: &GcpIamCustomRoleSpec{
				RoleId:      "logBucketWriter",
				Title:       "Log Bucket Writer",
				Permissions: []string{"storage.objects.create"},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		err := validator.Validate(minimal())
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fully populated spec", func() {
		msg := minimal()
		msg.Spec.ProjectId = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-gcp-project"},
		}
		msg.Spec.Description = "Grants exactly the permissions needed to write log objects"
		msg.Spec.Stage = ptr("BETA")
		msg.Spec.Permissions = []string{"storage.objects.create", "storage.objects.get"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a role_id with underscores and periods", func() {
		msg := minimal()
		msg.Spec.RoleId = "log_bucket.writer_v2"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a 3-character role_id (lower bound)", func() {
		msg := minimal()
		msg.Spec.RoleId = "abc"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a 64-character role_id (upper bound)", func() {
		msg := minimal()
		msg.Spec.RoleId = strings.Repeat("a", 64)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept every valid stage value", func() {
		for _, stage := range []string{"ALPHA", "BETA", "GA", "DEPRECATED", "DISABLED", "EAP"} {
			msg := minimal()
			msg.Spec.Stage = ptr(stage)
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "stage %s should be valid", stage)
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing role_id", func() {
		msg := minimal()
		msg.Spec.RoleId = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a role_id with hyphens", func() {
		msg := minimal()
		msg.Spec.RoleId = "log-bucket-writer"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a role_id shorter than 3 characters", func() {
		msg := minimal()
		msg.Spec.RoleId = "ab"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a role_id longer than 64 characters", func() {
		msg := minimal()
		msg.Spec.RoleId = strings.Repeat("a", 65)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing title", func() {
		msg := minimal()
		msg.Spec.Title = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a title longer than 100 characters", func() {
		msg := minimal()
		msg.Spec.Title = strings.Repeat("t", 101)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a description longer than 256 characters", func() {
		msg := minimal()
		msg.Spec.Description = strings.Repeat("d", 257)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an empty permissions list", func() {
		msg := minimal()
		msg.Spec.Permissions = []string{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an empty string permission", func() {
		msg := minimal()
		msg.Spec.Permissions = []string{""}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid stage value", func() {
		msg := minimal()
		msg.Spec.Stage = ptr("PRODUCTION")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = v
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
