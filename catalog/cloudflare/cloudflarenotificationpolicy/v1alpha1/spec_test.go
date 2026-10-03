package cloudflarenotificationpolicyv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareNotificationPolicySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareNotificationPolicySpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func webhookRef(id string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: id},
	}
}

func validPolicy(spec *CloudflareNotificationPolicySpec) *CloudflareNotificationPolicy {
	return &CloudflareNotificationPolicy{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareNotificationPolicy",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-notification-policy",
		},
		Spec: spec,
	}
}

func baseSpec() *CloudflareNotificationPolicySpec {
	return &CloudflareNotificationPolicySpec{
		AccountId: testAccountID,
		Name:      "tunnel-health",
		AlertType: "tunnel_health_event",
		Mechanisms: &CloudflareNotificationPolicyMechanisms{
			Emails: []string{"oncall@example.com"},
		},
	}
}

var _ = ginkgo.Describe("CloudflareNotificationPolicySpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept an email-delivered policy", func() {
			gomega.Expect(protovalidate.Validate(validPolicy(baseSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("should accept webhook and pagerduty destinations", func() {
			spec := baseSpec()
			spec.Mechanisms = &CloudflareNotificationPolicyMechanisms{
				PagerdutyIds: []string{"9430334d-cf60-4147-8b76-8d6cbea1b099"},
				WebhookIds:   []*foreignkeyv1.StringValueOrRef{webhookRef("0da2b59e-f118-42de-95bd-14fdd8ff4d7a")},
			}
			gomega.Expect(protovalidate.Validate(validPolicy(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept an incident alert with impact filters", func() {
			spec := baseSpec()
			spec.AlertType = "incident_alert"
			spec.Filters = &CloudflareNotificationPolicyFilters{
				IncidentImpact: []string{"INCIDENT_IMPACT_MAJOR", "INCIDENT_IMPACT_CRITICAL"},
			}
			gomega.Expect(protovalidate.Validate(validPolicy(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a traffic anomalies policy with the exclusion filter", func() {
			spec := baseSpec()
			spec.AlertType = "traffic_anomalies_alert"
			spec.Filters = &CloudflareNotificationPolicyFilters{
				TrafficExclusions: []string{"security_events"},
				Zones:             []string{"023e105f4ecef8ad9ca31a8372d0c353"},
			}
			gomega.Expect(protovalidate.Validate(validPolicy(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject an unknown alert type", func() {
			spec := baseSpec()
			spec.AlertType = "tunnel_health"
			gomega.Expect(protovalidate.Validate(validPolicy(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a policy without any destination", func() {
			spec := baseSpec()
			spec.Mechanisms = &CloudflareNotificationPolicyMechanisms{}
			gomega.Expect(protovalidate.Validate(validPolicy(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing mechanisms message", func() {
			spec := baseSpec()
			spec.Mechanisms = nil
			gomega.Expect(protovalidate.Validate(validPolicy(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid incident impact value", func() {
			spec := baseSpec()
			spec.AlertType = "incident_alert"
			spec.Filters = &CloudflareNotificationPolicyFilters{
				IncidentImpact: []string{"MAJOR"},
			}
			gomega.Expect(protovalidate.Validate(validPolicy(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid traffic exclusion value", func() {
			spec := baseSpec()
			spec.Filters = &CloudflareNotificationPolicyFilters{
				TrafficExclusions: []string{"all_events"},
			}
			gomega.Expect(protovalidate.Validate(validPolicy(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed account id", func() {
			spec := baseSpec()
			spec.AccountId = "not-hex"
			gomega.Expect(protovalidate.Validate(validPolicy(spec))).NotTo(gomega.BeNil())
		})
	})
})
