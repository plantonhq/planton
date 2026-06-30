package gcpkmskeyringv1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpKmsKeyRingSpec Suite")
}

var _ = ginkgo.Describe("GcpKmsKeyRingSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpKmsKeyRing.
	minimal := func() *GcpKmsKeyRing {
		return &GcpKmsKeyRing{
			ApiVersion: "gcp.planton.dev/v1",
			Kind:       "GcpKmsKeyRing",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-key-ring",
			},
			Spec: &GcpKmsKeyRingSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-gcp-project"},
				},
				KeyRingName: "prod-encryption",
				Location:    "us-central1",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept key_ring_name with uppercase letters", func() {
		msg := minimal()
		msg.Spec.KeyRingName = "Prod-Encryption-Keys"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept key_ring_name with underscores", func() {
		msg := minimal()
		msg.Spec.KeyRingName = "prod_encryption_keys"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept key_ring_name with mixed characters", func() {
		msg := minimal()
		msg.Spec.KeyRingName = "My_Key-Ring-01"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept key_ring_name starting with a digit", func() {
		msg := minimal()
		msg.Spec.KeyRingName = "1-key-ring"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept single character key_ring_name", func() {
		msg := minimal()
		msg.Spec.KeyRingName = "k"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept key_ring_name at maximum length (63 chars)", func() {
		msg := minimal()
		msg.Spec.KeyRingName = strings.Repeat("a", 63)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept location 'global'", func() {
		msg := minimal()
		msg.Spec.Location = "global"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept multi-region location", func() {
		msg := minimal()
		msg.Spec.Location = "us"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject when project_id is missing", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when key_ring_name is empty", func() {
		msg := minimal()
		msg.Spec.KeyRingName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject key_ring_name exceeding 63 characters", func() {
		msg := minimal()
		msg.Spec.KeyRingName = strings.Repeat("a", 64)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject key_ring_name with spaces", func() {
		msg := minimal()
		msg.Spec.KeyRingName = "invalid name"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject key_ring_name with dots", func() {
		msg := minimal()
		msg.Spec.KeyRingName = "invalid.name"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject key_ring_name with special characters", func() {
		msg := minimal()
		msg.Spec.KeyRingName = "invalid@name!"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when location is empty", func() {
		msg := minimal()
		msg.Spec.Location = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when metadata is missing", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when spec is missing", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
