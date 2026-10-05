package gcpspannerbackupschedulev1alpha1

import (
	"strings"
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
	ginkgo.RunSpecs(t, "GcpSpannerBackupScheduleSpec Suite")
}

var _ = ginkgo.Describe("GcpSpannerBackupScheduleSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpSpannerBackupSchedule.
	minimal := func() *GcpSpannerBackupSchedule {
		return &GcpSpannerBackupSchedule{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpSpannerBackupSchedule",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-backup-schedule",
			},
			Spec: &GcpSpannerBackupScheduleSpec{
				Instance: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-spanner-instance",
					},
				},
				Database: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-database",
					},
				},
				Cron:              "0 2 * * *",
				RetentionDuration: "86400s",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an omitted project_id (ambient project)", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a project_id literal", func() {
		msg := minimal()
		msg.Spec.ProjectId = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-gcp-project"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an explicit schedule_name", func() {
		msg := minimal()
		msg.Spec.ScheduleName = "daily-backups"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an empty schedule_name (defaults to metadata.name)", func() {
		msg := minimal()
		msg.Spec.ScheduleName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fractional-seconds retention_duration", func() {
		msg := minimal()
		msg.Spec.RetentionDuration = "86400.5s"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept FULL backup_type", func() {
		msg := minimal()
		msg.Spec.BackupType = proto.String("FULL")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept INCREMENTAL backup_type", func() {
		msg := minimal()
		msg.Spec.BackupType = proto.String("INCREMENTAL")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an unset backup_type (defaults to FULL)", func() {
		msg := minimal()
		msg.Spec.BackupType = nil
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept USE_DATABASE_ENCRYPTION without keys", func() {
		msg := minimal()
		msg.Spec.EncryptionConfig = &GcpSpannerBackupScheduleEncryptionConfig{
			EncryptionType: "USE_DATABASE_ENCRYPTION",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept GOOGLE_DEFAULT_ENCRYPTION without keys", func() {
		msg := minimal()
		msg.Spec.EncryptionConfig = &GcpSpannerBackupScheduleEncryptionConfig{
			EncryptionType: "GOOGLE_DEFAULT_ENCRYPTION",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept CUSTOMER_MANAGED_ENCRYPTION with a single kms_key_name", func() {
		msg := minimal()
		msg.Spec.EncryptionConfig = &GcpSpannerBackupScheduleEncryptionConfig{
			EncryptionType: "CUSTOMER_MANAGED_ENCRYPTION",
			KmsKeyName: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "projects/p/locations/us-central1/keyRings/r/cryptoKeys/k",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept CUSTOMER_MANAGED_ENCRYPTION with multi-region kms_key_names", func() {
		msg := minimal()
		msg.Spec.EncryptionConfig = &GcpSpannerBackupScheduleEncryptionConfig{
			EncryptionType: "CUSTOMER_MANAGED_ENCRYPTION",
			KmsKeyNames: []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "projects/p/locations/us-east1/keyRings/r/cryptoKeys/k1",
				}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "projects/p/locations/us-west1/keyRings/r/cryptoKeys/k2",
				}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept the 366-day maximum retention", func() {
		msg := minimal()
		msg.Spec.RetentionDuration = "31622400s"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing instance", func() {
		msg := minimal()
		msg.Spec.Instance = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing database", func() {
		msg := minimal()
		msg.Spec.Database = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing cron", func() {
		msg := minimal()
		msg.Spec.Cron = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing retention_duration", func() {
		msg := minimal()
		msg.Spec.RetentionDuration = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a retention_duration without the trailing 's'", func() {
		msg := minimal()
		msg.Spec.RetentionDuration = "86400"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a day-formatted retention_duration", func() {
		msg := minimal()
		msg.Spec.RetentionDuration = "31d"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a schedule_name starting with a digit", func() {
		msg := minimal()
		msg.Spec.ScheduleName = "1daily"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a schedule_name with uppercase characters", func() {
		msg := minimal()
		msg.Spec.ScheduleName = "Daily-Backups"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a schedule_name ending with a hyphen", func() {
		msg := minimal()
		msg.Spec.ScheduleName = "daily-"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a schedule_name longer than 63 characters", func() {
		msg := minimal()
		msg.Spec.ScheduleName = "a" + strings.Repeat("b", 63)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid backup_type", func() {
		msg := minimal()
		msg.Spec.BackupType = proto.String("DIFFERENTIAL")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an empty backup_type when explicitly set", func() {
		msg := minimal()
		msg.Spec.BackupType = proto.String("")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an encryption_config without encryption_type", func() {
		msg := minimal()
		msg.Spec.EncryptionConfig = &GcpSpannerBackupScheduleEncryptionConfig{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid encryption_type", func() {
		msg := minimal()
		msg.Spec.EncryptionConfig = &GcpSpannerBackupScheduleEncryptionConfig{
			EncryptionType: "SOME_OTHER_ENCRYPTION",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject CUSTOMER_MANAGED_ENCRYPTION with no keys", func() {
		msg := minimal()
		msg.Spec.EncryptionConfig = &GcpSpannerBackupScheduleEncryptionConfig{
			EncryptionType: "CUSTOMER_MANAGED_ENCRYPTION",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject CUSTOMER_MANAGED_ENCRYPTION with BOTH key shapes", func() {
		msg := minimal()
		msg.Spec.EncryptionConfig = &GcpSpannerBackupScheduleEncryptionConfig{
			EncryptionType: "CUSTOMER_MANAGED_ENCRYPTION",
			KmsKeyName: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "projects/p/locations/us-central1/keyRings/r/cryptoKeys/k",
				},
			},
			KmsKeyNames: []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "projects/p/locations/us-east1/keyRings/r/cryptoKeys/k1",
				}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject USE_DATABASE_ENCRYPTION with a kms_key_name", func() {
		msg := minimal()
		msg.Spec.EncryptionConfig = &GcpSpannerBackupScheduleEncryptionConfig{
			EncryptionType: "USE_DATABASE_ENCRYPTION",
			KmsKeyName: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "projects/p/locations/us-central1/keyRings/r/cryptoKeys/k",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject GOOGLE_DEFAULT_ENCRYPTION with kms_key_names", func() {
		msg := minimal()
		msg.Spec.EncryptionConfig = &GcpSpannerBackupScheduleEncryptionConfig{
			EncryptionType: "GOOGLE_DEFAULT_ENCRYPTION",
			KmsKeyNames: []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "projects/p/locations/us-east1/keyRings/r/cryptoKeys/k1",
				}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing spec", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind constant", func() {
		msg := minimal()
		msg.Kind = "GcpSpannerBackupSchedules"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept deletion_policy DELETE, PREVENT, ABANDON, and empty", func() {
		for _, policy := range []string{"DELETE", "PREVENT", "ABANDON", ""} {
			msg := minimal()
			msg.Spec.DeletionPolicy = policy
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "policy %q", policy)
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "RETAIN"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
