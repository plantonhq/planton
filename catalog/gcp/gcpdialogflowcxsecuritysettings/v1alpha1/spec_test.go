package gcpdialogflowcxsecuritysettingsv1alpha1

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
	ginkgo.RunSpecs(t, "GcpDialogflowCxSecuritySettingsSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpDialogflowCxSecuritySettingsSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpDialogflowCxSecuritySettings {
		return &GcpDialogflowCxSecuritySettings{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpDialogflowCxSecuritySettings",
			Metadata:   &shared.CloudResourceMetadata{Name: "pii-redaction"},
			Spec: &GcpDialogflowCxSecuritySettingsSpec{
				Location: "global",
			},
		}
	}

	ginkgo.It("should accept settings with only a location", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("support-project")
		msg.Spec.Location = "us-central1"
		msg.Spec.DisplayName = "PII redaction"
		msg.Spec.RedactionStrategy = "REDACT_WITH_SERVICE"
		msg.Spec.RedactionScope = "REDACT_DISK_STORAGE"
		msg.Spec.InspectTemplate = "projects/support-project/locations/us-central1/inspectTemplates/pii"
		msg.Spec.DeidentifyTemplate = "organizations/123/locations/us-central1/deidentifyTemplates/mask"
		msg.Spec.PurgeDataTypes = []string{"DIALOGFLOW_HISTORY"}
		msg.Spec.RetentionWindowDays = 30
		msg.Spec.AudioExportSettings = &GcpDialogflowCxSecuritySettingsAudioExport{
			GcsBucket:            litRef("support-call-audio"),
			AudioExportPattern:   "calls/{conversation}",
			AudioFormat:          "MP3",
			EnableAudioRedaction: true,
		}
		msg.Spec.EnableInsightsExport = true
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept removal after the conversation instead of a window", func() {
		msg := minimal()
		msg.Spec.RetentionStrategy = "REMOVE_AFTER_CONVERSATION"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject both retention rules at once", func() {
		msg := minimal()
		msg.Spec.RetentionStrategy = "REMOVE_AFTER_CONVERSATION"
		msg.Spec.RetentionWindowDays = 7
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a location and reject a malformed one", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Location = "US Central"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject values outside Google's lists", func() {
		msg := minimal()
		msg.Spec.RedactionStrategy = "REDACT_EVERYTHING"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.RedactionScope = "REDACT_MEMORY"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.PurgeDataTypes = []string{"ALL"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.AudioExportSettings = &GcpDialogflowCxSecuritySettingsAudioExport{AudioFormat: "WAV"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a negative retention window and malformed templates", func() {
		msg := minimal()
		msg.Spec.RetentionWindowDays = -1
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.InspectTemplate = "pii"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeidentifyTemplate = "projects/p/locations/global/inspectTemplates/wrong-kind"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
