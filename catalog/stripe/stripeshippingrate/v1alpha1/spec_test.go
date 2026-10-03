package stripeshippingratev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestStripeShippingRate(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripeShippingRate Suite")
}

func shippingRate(spec *StripeShippingRateSpec) *StripeShippingRate {
	return &StripeShippingRate{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripeShippingRate",
		Metadata:   &shared.CatalogObjectMetadata{Name: "standard-shipping"},
		Spec:       spec,
	}
}

// standard is 5 USD, 3 to 5 business days.
func standard() *StripeShippingRateSpec {
	return &StripeShippingRateSpec{
		DisplayName: "Standard",
		FixedAmount: &StripeShippingRateFixedAmount{Amount: 500, Currency: "usd"},
		DeliveryEstimate: &StripeShippingRateDeliveryEstimate{
			Minimum: &StripeShippingRateDeliveryBound{Unit: StripeShippingRateDeliveryBound_business_day, Value: 3},
			Maximum: &StripeShippingRateDeliveryBound{Unit: StripeShippingRateDeliveryBound_business_day, Value: 5},
		},
	}
}

var _ = ginkgo.Describe("StripeShippingRate Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a fixed amount with a delivery window", func() {
			gomega.Expect(protovalidate.Validate(shippingRate(standard()))).To(gomega.Succeed())
		})

		ginkgo.It("accepts free shipping with no estimate, and every in-place field", func() {
			spec := &StripeShippingRateSpec{
				DisplayName: "Free express",
				FixedAmount: &StripeShippingRateFixedAmount{
					Currency:        "usd",
					CurrencyOptions: map[string]*StripeShippingRateCurrencyOption{"eur": {TaxBehavior: StripeShippingRateCurrencyOption_inclusive}},
				},
				TaxBehavior: StripeShippingRateSpec_exclusive,
				TaxCode:     "txcd_92010001",
				Metadata:    map[string]string{"carrier": "dhl"},
			}
			gomega.Expect(protovalidate.Validate(shippingRate(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts an open-ended estimate", func() {
			spec := standard()
			spec.DeliveryEstimate.Maximum = nil
			gomega.Expect(protovalidate.Validate(shippingRate(spec))).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a rate with no name or no amount", func() {
			spec := standard()
			spec.DisplayName = ""
			gomega.Expect(protovalidate.Validate(shippingRate(spec))).NotTo(gomega.Succeed())

			spec = standard()
			spec.FixedAmount = nil
			gomega.Expect(protovalidate.Validate(shippingRate(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a negative amount, a missing or uppercase currency, and the main currency among the options", func() {
			spec := standard()
			spec.FixedAmount.Amount = -1
			gomega.Expect(protovalidate.Validate(shippingRate(spec))).NotTo(gomega.Succeed())

			for _, currency := range []string{"", "USD"} {
				spec = standard()
				spec.FixedAmount.Currency = currency
				gomega.Expect(protovalidate.Validate(shippingRate(spec))).NotTo(gomega.Succeed(), "currency %q", currency)
			}

			spec = standard()
			spec.FixedAmount.CurrencyOptions = map[string]*StripeShippingRateCurrencyOption{"usd": {Amount: 500}}
			err := protovalidate.Validate(shippingRate(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("the main currency is not one of them"))
		})

		ginkgo.It("refuses an empty estimate, a bound with no unit, and a zero value", func() {
			spec := standard()
			spec.DeliveryEstimate = &StripeShippingRateDeliveryEstimate{}
			err := protovalidate.Validate(shippingRate(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("sets minimum, maximum, or both"))

			spec = standard()
			spec.DeliveryEstimate.Minimum.Unit = StripeShippingRateDeliveryBound_unit_unspecified
			gomega.Expect(protovalidate.Validate(shippingRate(spec))).NotTo(gomega.Succeed())

			spec = standard()
			spec.DeliveryEstimate.Maximum.Value = 0
			gomega.Expect(protovalidate.Validate(shippingRate(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a malformed tax code", func() {
			spec := standard()
			spec.TaxCode = "92010001"
			gomega.Expect(protovalidate.Validate(shippingRate(spec))).NotTo(gomega.Succeed())
		})
	})
})
