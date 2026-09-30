package stripeeventdestinationv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestStripeEventDestination(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripeEventDestination Suite")
}

func destination(spec *StripeEventDestinationSpec) *StripeEventDestination {
	return &StripeEventDestination{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripeEventDestination",
		Metadata:   &shared.CloudResourceMetadata{Name: "billing-thin-events"},
		Spec:       spec,
	}
}

// thinWebhook is the minimal valid spec: thin events to a webhook URL.
func thinWebhook() *StripeEventDestinationSpec {
	return &StripeEventDestinationSpec{
		Name:          "Billing thin events",
		EventPayload:  StripeEventDestinationSpec_thin,
		EnabledEvents: []string{"v1.billing.meter.error_report_triggered"},
		Destination: &StripeEventDestinationSpec_WebhookEndpoint{
			WebhookEndpoint: &StripeEventDestinationWebhookEndpoint{Url: "https://app.example.com/stripe/thin"},
		},
	}
}

func eventBridge() *StripeEventDestinationSpec {
	return &StripeEventDestinationSpec{
		Name:          "Billing to EventBridge",
		EventPayload:  StripeEventDestinationSpec_snapshot,
		EnabledEvents: []string{"invoice.paid", "customer.subscription.updated"},
		Destination: &StripeEventDestinationSpec_AmazonEventbridge{
			AmazonEventbridge: &StripeEventDestinationAmazonEventBridge{AwsAccountId: "123456789012", AwsRegion: "us-east-1"},
		},
	}
}

var _ = ginkgo.Describe("StripeEventDestination Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a thin-event webhook destination", func() {
			gomega.Expect(protovalidate.Validate(destination(thinWebhook()))).To(gomega.Succeed())
		})

		ginkgo.It("accepts an EventBridge destination with a pinned snapshot version and routing", func() {
			spec := eventBridge()
			spec.Description = "Billing events for the data platform"
			spec.SnapshotApiVersion = "2025-09-30.clover"
			spec.EventsFrom = []string{"@self", "@organization_members/@accounts"}
			spec.Metadata = map[string]string{"team": "data"}
			gomega.Expect(protovalidate.Validate(destination(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts an Event Grid destination", func() {
			spec := eventBridge()
			spec.Destination = &StripeEventDestinationSpec_AzureEventGrid{AzureEventGrid: &StripeEventDestinationAzureEventGrid{
				AzureSubscriptionId:    "5f0c3b7e-8a52-4c1f-9f7e-3d2b1a0c9e8d",
				AzureResourceGroupName: "billing",
				AzureRegion:            "eastus",
			}}
			gomega.Expect(protovalidate.Validate(destination(spec))).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a destination with no block", func() {
			spec := thinWebhook()
			spec.Destination = nil
			gomega.Expect(protovalidate.Validate(destination(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a missing name, payload, or events", func() {
			spec := thinWebhook()
			spec.Name = ""
			gomega.Expect(protovalidate.Validate(destination(spec))).NotTo(gomega.Succeed())

			spec = thinWebhook()
			spec.EventPayload = StripeEventDestinationSpec_event_payload_unspecified
			gomega.Expect(protovalidate.Validate(destination(spec))).NotTo(gomega.Succeed())

			spec = thinWebhook()
			spec.EnabledEvents = nil
			gomega.Expect(protovalidate.Validate(destination(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses events that are not Stripe's dotted spelling", func() {
			spec := thinWebhook()
			spec.EnabledEvents = []string{"Invoice.Paid"}
			gomega.Expect(protovalidate.Validate(destination(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a snapshot version on thin events", func() {
			spec := thinWebhook()
			spec.SnapshotApiVersion = "2025-09-30.clover"
			err := protovalidate.Validate(destination(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("set event_payload snapshot"))
		})

		ginkgo.It("refuses a malformed snapshot version or routing entry", func() {
			spec := eventBridge()
			spec.SnapshotApiVersion = "latest"
			gomega.Expect(protovalidate.Validate(destination(spec))).NotTo(gomega.Succeed())

			spec = eventBridge()
			spec.EventsFrom = []string{"self"}
			gomega.Expect(protovalidate.Validate(destination(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a webhook URL that is not http or https", func() {
			spec := thinWebhook()
			spec.GetWebhookEndpoint().Url = "ftp://example.com/stripe"
			err := protovalidate.Validate(destination(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("url is an https://"))
		})

		ginkgo.It("refuses a malformed AWS account or region", func() {
			spec := eventBridge()
			spec.GetAmazonEventbridge().AwsAccountId = "12345"
			gomega.Expect(protovalidate.Validate(destination(spec))).NotTo(gomega.Succeed())

			spec = eventBridge()
			spec.GetAmazonEventbridge().AwsRegion = "US East"
			gomega.Expect(protovalidate.Validate(destination(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses an Azure destination without a UUID subscription or a resource group", func() {
			spec := eventBridge()
			spec.Destination = &StripeEventDestinationSpec_AzureEventGrid{AzureEventGrid: &StripeEventDestinationAzureEventGrid{
				AzureSubscriptionId: "not-a-uuid", AzureResourceGroupName: "billing", AzureRegion: "eastus",
			}}
			gomega.Expect(protovalidate.Validate(destination(spec))).NotTo(gomega.Succeed())

			spec.Destination = &StripeEventDestinationSpec_AzureEventGrid{AzureEventGrid: &StripeEventDestinationAzureEventGrid{
				AzureSubscriptionId: "5f0c3b7e-8a52-4c1f-9f7e-3d2b1a0c9e8d", AzureRegion: "eastus",
			}}
			gomega.Expect(protovalidate.Validate(destination(spec))).NotTo(gomega.Succeed())
		})
	})
})
