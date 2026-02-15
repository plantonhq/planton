package gcpkmskeyv1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpKmsKeySpec Suite")
}

var _ = ginkgo.Describe("GcpKmsKeySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpKmsKey.
	minimal := func() *GcpKmsKey {
		return &GcpKmsKey{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpKmsKey",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-kms-key",
			},
			Spec: &GcpKmsKeySpec{
				KeyRingId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "projects/my-project/locations/us-central1/keyRings/my-key-ring",
					},
				},
				KeyName: "cmek-encrypt-key",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec (key_ring_id + key_name only)", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept purpose ENCRYPT_DECRYPT", func() {
		msg := minimal()
		msg.Spec.Purpose = "ENCRYPT_DECRYPT"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept purpose ASYMMETRIC_SIGN", func() {
		msg := minimal()
		msg.Spec.Purpose = "ASYMMETRIC_SIGN"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept purpose ASYMMETRIC_DECRYPT", func() {
		msg := minimal()
		msg.Spec.Purpose = "ASYMMETRIC_DECRYPT"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept purpose MAC", func() {
		msg := minimal()
		msg.Spec.Purpose = "MAC"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept purpose RAW_ENCRYPT_DECRYPT", func() {
		msg := minimal()
		msg.Spec.Purpose = "RAW_ENCRYPT_DECRYPT"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept valid rotation_period", func() {
		msg := minimal()
		msg.Spec.RotationPeriod = "7776000s"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept rotation_period with fractional seconds", func() {
		msg := minimal()
		msg.Spec.RotationPeriod = "86400.5s"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept valid destroy_scheduled_duration", func() {
		msg := minimal()
		msg.Spec.DestroyScheduledDuration = "2592000s"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept version_template with algorithm and protection_level", func() {
		msg := minimal()
		msg.Spec.VersionTemplate = &GcpKmsKeyVersionTemplate{
			Algorithm:       "GOOGLE_SYMMETRIC_ENCRYPTION",
			ProtectionLevel: "HSM",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept version_template with algorithm only (SOFTWARE default)", func() {
		msg := minimal()
		msg.Spec.VersionTemplate = &GcpKmsKeyVersionTemplate{
			Algorithm: "EC_SIGN_P256_SHA256",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept key_name with uppercase letters", func() {
		msg := minimal()
		msg.Spec.KeyName = "Prod-CMEK-Key"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept key_name with underscores", func() {
		msg := minimal()
		msg.Spec.KeyName = "prod_cmek_key"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept key_name at maximum length (63 chars)", func() {
		msg := minimal()
		msg.Spec.KeyName = strings.Repeat("a", 63)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept skip_initial_version_creation set to true", func() {
		msg := minimal()
		msg.Spec.SkipInitialVersionCreation = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject when key_ring_id is missing", func() {
		msg := minimal()
		msg.Spec.KeyRingId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when key_name is empty", func() {
		msg := minimal()
		msg.Spec.KeyName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject key_name exceeding 63 characters", func() {
		msg := minimal()
		msg.Spec.KeyName = strings.Repeat("a", 64)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject key_name with spaces", func() {
		msg := minimal()
		msg.Spec.KeyName = "invalid name"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject key_name with dots", func() {
		msg := minimal()
		msg.Spec.KeyName = "invalid.name"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject key_name with special characters", func() {
		msg := minimal()
		msg.Spec.KeyName = "invalid@name!"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid purpose value", func() {
		msg := minimal()
		msg.Spec.Purpose = "INVALID_PURPOSE"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid rotation_period format (missing 's' suffix)", func() {
		msg := minimal()
		msg.Spec.RotationPeriod = "7776000"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid rotation_period format (non-numeric)", func() {
		msg := minimal()
		msg.Spec.RotationPeriod = "90days"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid destroy_scheduled_duration format", func() {
		msg := minimal()
		msg.Spec.DestroyScheduledDuration = "30d"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid protection_level", func() {
		msg := minimal()
		msg.Spec.VersionTemplate = &GcpKmsKeyVersionTemplate{
			Algorithm:       "GOOGLE_SYMMETRIC_ENCRYPTION",
			ProtectionLevel: "EXTERNAL",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject version_template without algorithm", func() {
		msg := minimal()
		msg.Spec.VersionTemplate = &GcpKmsKeyVersionTemplate{
			ProtectionLevel: "SOFTWARE",
		}
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
