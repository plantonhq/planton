package stripecouponv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestStripeCoupon(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripeCoupon Suite")
}

func coupon(spec *StripeCouponSpec) *StripeCoupon {
	return &StripeCoupon{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripeCoupon",
		Metadata:   &shared.CloudResourceMetadata{Name: "launch-25"},
		Spec:       spec,
	}
}

func int64Ptr(v int64) *int64       { return &v }
func float64Ptr(v float64) *float64 { return &v }

func productRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
		ValueFrom: &foreignkeyv1.ValueFromRef{Kind: cloudresourcekind.CloudResourceKind_StripeProduct, Name: "pro-plan"},
	}}
}

// launch is 25% off for three months on a referenced product.
func launch() *StripeCouponSpec {
	return &StripeCouponSpec{
		Name:              "Launch 25% off",
		PercentOff:        float64Ptr(25),
		Duration:          StripeCouponSpec_repeating,
		DurationInMonths:  int64Ptr(3),
		AppliesToProducts: []*foreignkeyv1.StringValueOrRef{productRef()},
	}
}

// tenOff is 10 USD off once, with a euro amount.
func tenOff() *StripeCouponSpec {
	return &StripeCouponSpec{
		AmountOff:       int64Ptr(1000),
		Currency:        "usd",
		CurrencyOptions: map[string]int64{"eur": 900},
		Duration:        StripeCouponSpec_once,
	}
}

var _ = ginkgo.Describe("StripeCoupon Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a repeating percentage coupon on a referenced product", func() {
			gomega.Expect(protovalidate.Validate(coupon(launch()))).To(gomega.Succeed())
		})

		ginkgo.It("accepts an amount coupon with currency options, limits and a literal product", func() {
			spec := tenOff()
			spec.MaxRedemptions = int64Ptr(500)
			spec.RedeemBy = int64Ptr(1798761599)
			spec.Metadata = map[string]string{"campaign": "spring"}
			spec.AppliesToProducts = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "prod_123"}},
			}
			gomega.Expect(protovalidate.Validate(coupon(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts a percentage coupon with no duration (Stripe's once) and 100% off forever", func() {
			spec := &StripeCouponSpec{PercentOff: float64Ptr(12.5)}
			gomega.Expect(protovalidate.Validate(coupon(spec))).To(gomega.Succeed())

			spec = &StripeCouponSpec{PercentOff: float64Ptr(100), Duration: StripeCouponSpec_forever}
			gomega.Expect(protovalidate.Validate(coupon(spec))).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a coupon with no discount, or with both kinds", func() {
			err := protovalidate.Validate(coupon(&StripeCouponSpec{}))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one discount"))

			spec := tenOff()
			spec.PercentOff = float64Ptr(10)
			gomega.Expect(protovalidate.Validate(coupon(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a percentage out of range", func() {
			for _, pct := range []float64{0, -5, 100.5} {
				spec := launch()
				spec.PercentOff = float64Ptr(pct)
				gomega.Expect(protovalidate.Validate(coupon(spec))).NotTo(gomega.Succeed(), "percent_off %v", pct)
			}
		})

		ginkgo.It("refuses an amount without its currency, a currency on a percentage, or a malformed currency", func() {
			spec := tenOff()
			spec.Currency = ""
			spec.CurrencyOptions = nil
			err := protovalidate.Validate(coupon(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("currency belongs to amount_off"))

			spec = launch()
			spec.Currency = "usd"
			gomega.Expect(protovalidate.Validate(coupon(spec))).NotTo(gomega.Succeed())

			spec = tenOff()
			spec.Currency = "USD"
			gomega.Expect(protovalidate.Validate(coupon(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses currency options on a percentage, the main currency among them, or a zero amount", func() {
			spec := launch()
			spec.CurrencyOptions = map[string]int64{"eur": 500}
			err := protovalidate.Validate(coupon(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("currency_options are amount_off in other currencies"))

			spec = tenOff()
			spec.CurrencyOptions = map[string]int64{"usd": 1000}
			gomega.Expect(protovalidate.Validate(coupon(spec))).NotTo(gomega.Succeed())

			spec = tenOff()
			spec.CurrencyOptions = map[string]int64{"eur": 0}
			gomega.Expect(protovalidate.Validate(coupon(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses months without repeating, and repeating without months", func() {
			spec := launch()
			spec.DurationInMonths = nil
			err := protovalidate.Validate(coupon(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("duration_in_months goes with duration repeating"))

			spec = tenOff()
			spec.DurationInMonths = int64Ptr(3)
			gomega.Expect(protovalidate.Validate(coupon(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a literal that is not a product id, and a product named twice", func() {
			spec := launch()
			spec.AppliesToProducts = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "price_123"}},
			}
			err := protovalidate.Validate(coupon(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("Stripe product id"))

			literal := &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "prod_123"}}
			spec = launch()
			spec.AppliesToProducts = []*foreignkeyv1.StringValueOrRef{literal, literal}
			gomega.Expect(protovalidate.Validate(coupon(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a zero redemption limit or redeem-by", func() {
			spec := launch()
			spec.MaxRedemptions = int64Ptr(0)
			gomega.Expect(protovalidate.Validate(coupon(spec))).NotTo(gomega.Succeed())

			spec = launch()
			spec.RedeemBy = int64Ptr(0)
			gomega.Expect(protovalidate.Validate(coupon(spec))).NotTo(gomega.Succeed())
		})
	})
})
