package stripebillingmeterv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestStripeBillingMeter(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripeBillingMeter Suite")
}

func meter(spec *StripeBillingMeterSpec) *StripeBillingMeter {
	return &StripeBillingMeter{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripeBillingMeter",
		Metadata:   &shared.CatalogObjectMetadata{Name: "api-requests"},
		Spec:       spec,
	}
}

// apiRequests sums the value of api_requests events, with one alert.
func apiRequests() *StripeBillingMeterSpec {
	return &StripeBillingMeterSpec{
		DisplayName:        "API requests",
		EventName:          "api_requests",
		DefaultAggregation: &StripeBillingMeterDefaultAggregation{Formula: StripeBillingMeterDefaultAggregation_sum},
		Alerts:             []*StripeBillingMeterAlert{{Title: "10k requests", Gte: 10000}},
	}
}

var _ = ginkgo.Describe("StripeBillingMeter Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a summing meter with an alert", func() {
			gomega.Expect(protovalidate.Validate(meter(apiRequests()))).To(gomega.Succeed())
		})

		ginkgo.It("accepts every field: mapping, value key, time window, a customer alert with recurrence", func() {
			spec := apiRequests()
			spec.CustomerMapping = &StripeBillingMeterCustomerMapping{EventPayloadKey: "stripe_customer_id"}
			spec.ValueSettings = &StripeBillingMeterValueSettings{EventPayloadKey: "requests"}
			spec.EventTimeWindow = StripeBillingMeterSpec_hour
			spec.Alerts = append(spec.Alerts, &StripeBillingMeterAlert{
				Title: "big customer", Gte: 1000000, Customer: "cus_123", Recurrence: StripeBillingMeterAlert_one_time,
			})
			gomega.Expect(protovalidate.Validate(meter(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts a counting meter with no alerts", func() {
			spec := &StripeBillingMeterSpec{
				DisplayName:        "Active seats",
				EventName:          "seat_active",
				DefaultAggregation: &StripeBillingMeterDefaultAggregation{Formula: StripeBillingMeterDefaultAggregation_last},
			}
			gomega.Expect(protovalidate.Validate(meter(spec))).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a meter with no display name, event name or aggregation", func() {
			spec := apiRequests()
			spec.DisplayName = ""
			gomega.Expect(protovalidate.Validate(meter(spec))).NotTo(gomega.Succeed())

			spec = apiRequests()
			spec.EventName = ""
			gomega.Expect(protovalidate.Validate(meter(spec))).NotTo(gomega.Succeed())

			spec = apiRequests()
			spec.DefaultAggregation = nil
			gomega.Expect(protovalidate.Validate(meter(spec))).NotTo(gomega.Succeed())

			spec = apiRequests()
			spec.DefaultAggregation.Formula = StripeBillingMeterDefaultAggregation_formula_unspecified
			gomega.Expect(protovalidate.Validate(meter(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a mapping or value setting with no key", func() {
			spec := apiRequests()
			spec.CustomerMapping = &StripeBillingMeterCustomerMapping{}
			gomega.Expect(protovalidate.Validate(meter(spec))).NotTo(gomega.Succeed())

			spec = apiRequests()
			spec.ValueSettings = &StripeBillingMeterValueSettings{}
			gomega.Expect(protovalidate.Validate(meter(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses two alerts with one title, an alert with no title or threshold, and a malformed customer", func() {
			spec := apiRequests()
			spec.Alerts = append(spec.Alerts, &StripeBillingMeterAlert{Title: "10k requests", Gte: 20000})
			err := protovalidate.Validate(meter(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("each alert has its own title"))

			spec = apiRequests()
			spec.Alerts[0].Title = ""
			gomega.Expect(protovalidate.Validate(meter(spec))).NotTo(gomega.Succeed())

			spec = apiRequests()
			spec.Alerts[0].Gte = 0
			gomega.Expect(protovalidate.Validate(meter(spec))).NotTo(gomega.Succeed())

			spec = apiRequests()
			spec.Alerts[0].Customer = "acct_123"
			gomega.Expect(protovalidate.Validate(meter(spec))).NotTo(gomega.Succeed())
		})
	})
})
