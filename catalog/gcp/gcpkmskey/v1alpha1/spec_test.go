package gcpkmskeyv1alpha1

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
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpKmsKey",
			Metadata: &shared.CatalogObjectMetadata{
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

	ginkgo.It("should accept a future purpose value the API adds (free-string contract)", func() {
		msg := minimal()
		msg.Spec.Purpose = "KEY_ENCAPSULATION"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept rotation_period at exactly the 24h minimum", func() {
		msg := minimal()
		msg.Spec.RotationPeriod = "86400s"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a 90-day rotation_period", func() {
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

	ginkgo.It("should accept rotation_period together with explicit ENCRYPT_DECRYPT purpose", func() {
		msg := minimal()
		msg.Spec.Purpose = "ENCRYPT_DECRYPT"
		msg.Spec.RotationPeriod = "2592000s"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept valid destroy_scheduled_duration", func() {
		msg := minimal()
		msg.Spec.DestroyScheduledDuration = "2592000s"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept version_template with algorithm and HSM protection", func() {
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

	ginkgo.It("should accept EXTERNAL protection level", func() {
		msg := minimal()
		msg.Spec.VersionTemplate = &GcpKmsKeyVersionTemplate{
			Algorithm:       "EXTERNAL_SYMMETRIC_ENCRYPTION",
			ProtectionLevel: "EXTERNAL",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept EXTERNAL_VPC protection with a crypto_key_backend", func() {
		msg := minimal()
		msg.Spec.VersionTemplate = &GcpKmsKeyVersionTemplate{
			Algorithm:       "EXTERNAL_SYMMETRIC_ENCRYPTION",
			ProtectionLevel: "EXTERNAL_VPC",
		}
		msg.Spec.CryptoKeyBackend = "projects/my-project/locations/us-central1/ekmConnections/my-ekm"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept import_only when skip_initial_version_creation is also set", func() {
		msg := minimal()
		msg.Spec.ImportOnly = true
		msg.Spec.SkipInitialVersionCreation = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept user labels", func() {
		msg := minimal()
		msg.Spec.Labels = map[string]string{"team": "payments", "cost-center": "cc-1234"}
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

	ginkgo.It("should reject rotation_period below the 24h minimum", func() {
		msg := minimal()
		msg.Spec.RotationPeriod = "86399s"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a one-hour rotation_period", func() {
		msg := minimal()
		msg.Spec.RotationPeriod = "3600s"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject rotation_period with more than 9 fractional digits", func() {
		msg := minimal()
		msg.Spec.RotationPeriod = "86400.1234567891s"
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

	ginkgo.It("should reject rotation_period on a non-ENCRYPT_DECRYPT key", func() {
		msg := minimal()
		msg.Spec.Purpose = "ASYMMETRIC_SIGN"
		msg.Spec.RotationPeriod = "7776000s"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid destroy_scheduled_duration format", func() {
		msg := minimal()
		msg.Spec.DestroyScheduledDuration = "30d"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid protection_level", func() {
		msg := minimal()
		msg.Spec.VersionTemplate = &GcpKmsKeyVersionTemplate{
			Algorithm:       "GOOGLE_SYMMETRIC_ENCRYPTION",
			ProtectionLevel: "FIPS_HSM",
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

	ginkgo.It("should reject import_only without skip_initial_version_creation", func() {
		msg := minimal()
		msg.Spec.ImportOnly = true
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject crypto_key_backend without EXTERNAL_VPC protection", func() {
		msg := minimal()
		msg.Spec.CryptoKeyBackend = "projects/my-project/locations/us-central1/ekmConnections/my-ekm"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject crypto_key_backend when version_template protection is HSM", func() {
		msg := minimal()
		msg.Spec.VersionTemplate = &GcpKmsKeyVersionTemplate{
			Algorithm:       "GOOGLE_SYMMETRIC_ENCRYPTION",
			ProtectionLevel: "HSM",
		}
		msg.Spec.CryptoKeyBackend = "projects/my-project/locations/us-central1/ekmConnections/my-ekm"
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
