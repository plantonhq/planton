package gcpfirestorebackupschedulev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpFirestoreBackupScheduleSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpFirestoreBackupScheduleSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimalDaily := func() *GcpFirestoreBackupSchedule {
		return &GcpFirestoreBackupSchedule{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpFirestoreBackupSchedule",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "daily-backups",
			},
			Spec: &GcpFirestoreBackupScheduleSpec{
				Database:  litRef("my-firestore-db"),
				Retention: "604800s",
				Daily:     true,
			},
		}
	}

	minimalWeekly := func() *GcpFirestoreBackupSchedule {
		msg := minimalDaily()
		msg.Spec.Daily = false
		msg.Spec.WeeklyRecurrence = &GcpFirestoreBackupScheduleWeeklyRecurrence{
			Day: "SUNDAY",
		}
		return msg
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal daily schedule", func() {
		gomega.Expect(validator.Validate(minimalDaily())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a minimal weekly schedule", func() {
		gomega.Expect(validator.Validate(minimalWeekly())).To(gomega.Succeed())
	})

	ginkgo.It("should accept an omitted project_id (ambient project)", func() {
		msg := minimalDaily()
		msg.Spec.ProjectId = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id literal", func() {
		msg := minimalDaily()
		msg.Spec.ProjectId = litRef("my-gcp-project")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each valid weekly day", func() {
		for _, day := range []string{
			"MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY",
			"FRIDAY", "SATURDAY", "SUNDAY",
		} {
			msg := minimalWeekly()
			msg.Spec.WeeklyRecurrence.Day = day
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept the 14-week maximum retention", func() {
		msg := minimalDaily()
		msg.Spec.Retention = "8467200s"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a reference-shaped database", func() {
		msg := minimalDaily()
		msg.Spec.Database = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{Name: "prod-firestore"},
			},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing database", func() {
		msg := minimalDaily()
		msg.Spec.Database = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing retention", func() {
		msg := minimalDaily()
		msg.Spec.Retention = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a retention without the trailing s", func() {
		msg := minimalDaily()
		msg.Spec.Retention = "604800"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a day-formatted retention", func() {
		msg := minimalDaily()
		msg.Spec.Retention = "7d"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject neither daily nor weekly recurrence", func() {
		msg := minimalDaily()
		msg.Spec.Daily = false
		msg.Spec.WeeklyRecurrence = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one recurrence"))
	})

	ginkgo.It("should reject both daily and weekly recurrence", func() {
		msg := minimalDaily()
		msg.Spec.WeeklyRecurrence = &GcpFirestoreBackupScheduleWeeklyRecurrence{
			Day: "MONDAY",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one recurrence"))
	})

	ginkgo.It("should reject an invalid weekly day", func() {
		msg := minimalWeekly()
		msg.Spec.WeeklyRecurrence.Day = "FUNDAY"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing weekly day", func() {
		msg := minimalWeekly()
		msg.Spec.WeeklyRecurrence.Day = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing spec", func() {
		msg := minimalDaily()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind literal", func() {
		msg := minimalDaily()
		msg.Kind = "GcpFirestoreBackupSchedules"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept every deletion_policy value", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			msg := minimalDaily()
			msg.Spec.DeletionPolicy = policy
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimalDaily()
		msg.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
