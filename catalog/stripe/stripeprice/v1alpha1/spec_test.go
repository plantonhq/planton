package stripepricev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestStripePrice(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripePrice Suite")
}

func price(spec *StripePriceSpec) *StripePrice {
	return &StripePrice{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripePrice",
		Metadata:   &shared.CloudResourceMetadata{Name: "pro-monthly"},
		Spec:       spec,
	}
}

func int64Ptr(v int64) *int64 { return &v }

func productRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
		ValueFrom: &foreignkeyv1.ValueFromRef{Kind: cloudresourcekind.CloudResourceKind_StripeProduct, Name: "pro-plan"},
	}}
}

// proMonthly is the minimal recurring price: 49 USD a month on a referenced product.
func proMonthly() *StripePriceSpec {
	return &StripePriceSpec{
		Product:    productRef(),
		Currency:   "usd",
		UnitAmount: int64Ptr(4900),
		Recurring:  &StripePriceRecurring{Interval: StripePriceRecurring_month},
	}
}

// graduated is a valid tiered price.
func graduated() *StripePriceSpec {
	return &StripePriceSpec{
		Product:       productRef(),
		Currency:      "usd",
		BillingScheme: StripePriceSpec_tiered,
		TiersMode:     StripePriceSpec_graduated,
		Tiers: []*StripePriceTier{
			{UpTo: "1000", UnitAmount: int64Ptr(10)},
			{UpTo: "inf", UnitAmountDecimal: "0.5"},
		},
		Recurring: &StripePriceRecurring{Interval: StripePriceRecurring_month, UsageType: StripePriceRecurring_metered, Meter: "mtr_123"},
	}
}

var _ = ginkgo.Describe("StripePrice Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts the minimal recurring price", func() {
			gomega.Expect(protovalidate.Validate(price(proMonthly()))).To(gomega.Succeed())
		})

		ginkgo.It("accepts every in-place field, currency options, and a literal product id", func() {
			spec := proMonthly()
			spec.Product = &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "prod_123"}}
			spec.LookupKey = "pro-monthly"
			spec.TransferLookupKey = true
			spec.Nickname = "Pro, monthly"
			spec.TaxBehavior = StripePriceSpec_exclusive
			spec.Metadata = map[string]string{"plan": "pro"}
			spec.CurrencyOptions = map[string]*StripePriceCurrencyOption{
				"eur": {UnitAmount: int64Ptr(4500), TaxBehavior: StripePriceCurrencyOption_inclusive},
			}
			spec.Recurring.IntervalCount = int64Ptr(3)
			spec.Recurring.TrialPeriodDays = int64Ptr(14)
			gomega.Expect(protovalidate.Validate(price(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts a graduated metered price with tiered currency options", func() {
			spec := graduated()
			spec.CurrencyOptions = map[string]*StripePriceCurrencyOption{
				"eur": {Tiers: []*StripePriceTier{{UpTo: "inf", UnitAmount: int64Ptr(9)}}},
			}
			gomega.Expect(protovalidate.Validate(price(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts a one-time price the customer chooses, and a transformed per-unit price", func() {
			spec := &StripePriceSpec{
				Product:          productRef(),
				Currency:         "usd",
				CustomUnitAmount: &StripePriceCustomUnitAmount{Minimum: int64Ptr(500), Preset: int64Ptr(1000), Maximum: int64Ptr(10000)},
			}
			gomega.Expect(protovalidate.Validate(price(spec))).To(gomega.Succeed())

			spec = proMonthly()
			spec.UnitAmount = nil
			spec.UnitAmountDecimal = "0.25"
			spec.TransformQuantity = &StripePriceTransformQuantity{DivideBy: 1000, Round: StripePriceTransformQuantity_up}
			gomega.Expect(protovalidate.Validate(price(spec))).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a missing product, or a literal that is not a product id", func() {
			spec := proMonthly()
			spec.Product = nil
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())

			spec = proMonthly()
			spec.Product = &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "price_123"}}
			err := protovalidate.Validate(price(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("Stripe product id"))
		})

		ginkgo.It("refuses a currency that is not a lowercase ISO code", func() {
			for _, currency := range []string{"", "USD", "us", "dollars"} {
				spec := proMonthly()
				spec.Currency = currency
				gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed(), "currency %q", currency)
			}
		})

		ginkgo.It("refuses a per-unit price with no amount, or with two", func() {
			spec := proMonthly()
			spec.UnitAmount = nil
			err := protovalidate.Validate(price(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one of unit_amount"))

			spec = proMonthly()
			spec.UnitAmountDecimal = "4900"
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses tiers on a per-unit price", func() {
			spec := proMonthly()
			spec.Tiers = []*StripePriceTier{{UpTo: "inf", UnitAmount: int64Ptr(1)}}
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a tiered price without a mode, without an inf last tier, or with a unit amount", func() {
			spec := graduated()
			spec.TiersMode = StripePriceSpec_tiers_mode_unspecified
			err := protovalidate.Validate(price(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("a tiered price sets tiers_mode and tiers"))

			spec = graduated()
			spec.Tiers[1].UpTo = "5000"
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())

			spec = graduated()
			spec.UnitAmount = int64Ptr(10)
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a tier with no amount, two forms of one amount, or a malformed bound", func() {
			spec := graduated()
			spec.Tiers[0] = &StripePriceTier{UpTo: "1000"}
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())

			spec = graduated()
			spec.Tiers[0] = &StripePriceTier{UpTo: "1000", UnitAmount: int64Ptr(1), UnitAmountDecimal: "1"}
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())

			spec = graduated()
			spec.Tiers[0].UpTo = "0"
			err := protovalidate.Validate(price(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("up_to is a whole number of units"))
		})

		ginkgo.It("refuses transform_quantity on a tiered price", func() {
			spec := graduated()
			spec.TransformQuantity = &StripePriceTransformQuantity{DivideBy: 10, Round: StripePriceTransformQuantity_down}
			err := protovalidate.Validate(price(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("transform_quantity cannot be combined with tiers"))
		})

		ginkgo.It("refuses a transform with no divisor or no rounding", func() {
			spec := proMonthly()
			spec.TransformQuantity = &StripePriceTransformQuantity{Round: StripePriceTransformQuantity_up}
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())

			spec = proMonthly()
			spec.TransformQuantity = &StripePriceTransformQuantity{DivideBy: 10}
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a malformed decimal amount", func() {
			spec := proMonthly()
			spec.UnitAmount = nil
			spec.UnitAmountDecimal = "0." + strings.Repeat("1", 13)
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a custom amount range out of order", func() {
			spec := &StripePriceSpec{
				Product:          productRef(),
				Currency:         "usd",
				CustomUnitAmount: &StripePriceCustomUnitAmount{Minimum: int64Ptr(1000), Preset: int64Ptr(500)},
			}
			err := protovalidate.Validate(price(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("minimum <= preset <= maximum"))
		})

		ginkgo.It("refuses a recurring price without an interval, or beyond three years", func() {
			spec := proMonthly()
			spec.Recurring.Interval = StripePriceRecurring_interval_unspecified
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())

			spec = proMonthly()
			spec.Recurring.IntervalCount = int64Ptr(37)
			err := protovalidate.Validate(price(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("at most three years"))
		})

		ginkgo.It("refuses a meter on a licensed price, or a meter that is not a meter id", func() {
			spec := proMonthly()
			spec.Recurring.Meter = "mtr_123"
			err := protovalidate.Validate(price(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("needs usage_type metered"))

			spec = graduated()
			spec.Recurring.Meter = "meter_123"
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses the main currency among the options, and options of the wrong scheme", func() {
			spec := proMonthly()
			spec.CurrencyOptions = map[string]*StripePriceCurrencyOption{"usd": {UnitAmount: int64Ptr(4900)}}
			err := protovalidate.Validate(price(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("the main currency is not one of them"))

			spec = graduated()
			spec.CurrencyOptions = map[string]*StripePriceCurrencyOption{"eur": {UnitAmount: int64Ptr(9)}}
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())

			spec = proMonthly()
			spec.CurrencyOptions = map[string]*StripePriceCurrencyOption{"EUR": {UnitAmount: int64Ptr(4500)}}
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a lookup key over 200 characters", func() {
			spec := proMonthly()
			spec.LookupKey = strings.Repeat("k", 201)
			gomega.Expect(protovalidate.Validate(price(spec))).NotTo(gomega.Succeed())
		})
	})
})
