package gcpmonitoringnotificationchannelv1alpha1

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
	ginkgo.RunSpecs(t, "GcpMonitoringNotificationChannelSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpMonitoringNotificationChannelSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpMonitoringNotificationChannel {
		return &GcpMonitoringNotificationChannel{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpMonitoringNotificationChannel",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-channel",
			},
			Spec: &GcpMonitoringNotificationChannelSpec{
				Type: "email",
				ChannelLabels: map[string]string{
					"email_address": "oncall@example.com",
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal email channel", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id literal", func() {
		target := minimal()
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a slack channel with its auth token in sensitive_labels", func() {
		target := minimal()
		target.Spec.Type = "slack"
		target.Spec.ChannelLabels = map[string]string{"channel_name": "#alerts"}
		target.Spec.SensitiveLabels = &GcpMonitoringNotificationChannelSensitiveLabels{
			AuthToken: "xoxb-test-token",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept display_name, description, and user labels", func() {
		target := minimal()
		target.Spec.DisplayName = "On-call email"
		target.Spec.Description = "Primary paging channel for the platform team"
		target.Spec.Labels = map[string]string{"team": "platform"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept enabled=false and force_delete", func() {
		target := minimal()
		enabled := false
		target.Spec.Enabled = &enabled
		target.Spec.ForceDelete = true
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing type", func() {
		target := minimal()
		target.Spec.Type = ""
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject credentials placed in channel_labels", func() {
		for _, key := range []string{"auth_token", "password", "service_key"} {
			target := minimal()
			target.Spec.ChannelLabels = map[string]string{key: "plaintext-secret"}
			err := validator.Validate(target)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(strings.Contains(err.Error(), "sensitive_labels")).To(gomega.BeTrue())
		}
	})

	ginkgo.It("should reject an over-long display_name", func() {
		target := minimal()
		target.Spec.DisplayName = strings.Repeat("x", 513)
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "ABANDON")).To(gomega.BeTrue())
	})
})
