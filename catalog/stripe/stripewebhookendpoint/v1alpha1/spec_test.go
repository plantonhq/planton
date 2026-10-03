package stripewebhookendpointv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestStripeWebhookEndpoint(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripeWebhookEndpoint Suite")
}

func endpoint(spec *StripeWebhookEndpointSpec) *StripeWebhookEndpoint {
	return &StripeWebhookEndpoint{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripeWebhookEndpoint",
		Metadata:   &shared.CatalogObjectMetadata{Name: "billing-events"},
		Spec:       spec,
	}
}

// billingEvents is the minimal valid spec: a URL and the events it receives.
func billingEvents() *StripeWebhookEndpointSpec {
	return &StripeWebhookEndpointSpec{
		Url:           "https://app.example.com/stripe",
		EnabledEvents: []string{"checkout.session.completed", "invoice.paid", "charge.refunded"},
	}
}

var _ = ginkgo.Describe("StripeWebhookEndpoint Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts the minimal endpoint", func() {
			gomega.Expect(protovalidate.Validate(endpoint(billingEvents()))).To(gomega.Succeed())
		})

		ginkgo.It("accepts every optional field", func() {
			spec := billingEvents()
			spec.Description = "Billing events"
			spec.Metadata = map[string]string{"team": "billing"}
			spec.ApiVersion = "2025-09-30.clover"
			spec.Connect = true
			gomega.Expect(protovalidate.Validate(endpoint(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts the wildcard, dated api versions, and events with digits and underscores", func() {
			spec := billingEvents()
			spec.EnabledEvents = []string{"*", "treasury.outbound_payment.returned", "v1.billing.meter.error_report_triggered"}
			spec.ApiVersion = "2024-06-20"
			gomega.Expect(protovalidate.Validate(endpoint(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts an http URL for a sandbox tunnel", func() {
			spec := billingEvents()
			spec.Url = "http://localhost.example.test:4242/stripe"
			gomega.Expect(protovalidate.Validate(endpoint(spec))).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a missing URL", func() {
			spec := billingEvents()
			spec.Url = ""
			gomega.Expect(protovalidate.Validate(endpoint(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a URL that is not http or https", func() {
			spec := billingEvents()
			spec.Url = "ftp://example.com/stripe"
			err := protovalidate.Validate(endpoint(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("url is an https://"))
		})

		ginkgo.It("refuses an endpoint with no events", func() {
			spec := billingEvents()
			spec.EnabledEvents = nil
			gomega.Expect(protovalidate.Validate(endpoint(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a repeated event", func() {
			spec := billingEvents()
			spec.EnabledEvents = []string{"invoice.paid", "invoice.paid"}
			gomega.Expect(protovalidate.Validate(endpoint(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses an event that is not Stripe's dotted spelling", func() {
			for _, event := range []string{"invoice", "Invoice.Paid", "invoice paid", "invoice.*"} {
				spec := billingEvents()
				spec.EnabledEvents = []string{event}
				gomega.Expect(protovalidate.Validate(endpoint(spec))).NotTo(gomega.Succeed(), "event %q", event)
			}
		})

		ginkgo.It("refuses an api version that is not a Stripe version", func() {
			spec := billingEvents()
			spec.ApiVersion = "latest"
			err := protovalidate.Validate(endpoint(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("api_version is a Stripe API version"))
		})
	})
})
