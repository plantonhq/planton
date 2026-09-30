package stripepromotioncodev1alpha1

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

func TestStripePromotionCode(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripePromotionCode Suite")
}

func promotionCode(spec *StripePromotionCodeSpec) *StripePromotionCode {
	return &StripePromotionCode{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripePromotionCode",
		Metadata:   &shared.CloudResourceMetadata{Name: "launch25"},
		Spec:       spec,
	}
}

func int64Ptr(v int64) *int64 { return &v }
func boolPtr(v bool) *bool    { return &v }

func couponRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
		ValueFrom: &foreignkeyv1.ValueFromRef{Kind: cloudresourcekind.CloudResourceKind_StripeCoupon, Name: "launch-25"},
	}}
}

func launch25() *StripePromotionCodeSpec {
	return &StripePromotionCodeSpec{Coupon: couponRef(), Code: "LAUNCH25"}
}

var _ = ginkgo.Describe("StripePromotionCode Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a code on a referenced coupon", func() {
			gomega.Expect(protovalidate.Validate(promotionCode(launch25()))).To(gomega.Succeed())
		})

		ginkgo.It("accepts a generated code (no code) on a literal coupon id", func() {
			spec := &StripePromotionCodeSpec{Coupon: &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "Z4OV52SU"}}}
			gomega.Expect(protovalidate.Validate(promotionCode(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts every limit, restrictions with currency options, and dashes in the code", func() {
			spec := launch25()
			spec.Code = "SPRING-2026-launch"
			spec.Customer = "cus_123"
			spec.ExpiresAt = int64Ptr(1798761599)
			spec.MaxRedemptions = int64Ptr(500)
			spec.Active = boolPtr(false)
			spec.Metadata = map[string]string{"campaign": "spring"}
			spec.Restrictions = &StripePromotionCodeRestrictions{
				FirstTimeTransaction:  boolPtr(true),
				MinimumAmount:         int64Ptr(5000),
				MinimumAmountCurrency: "usd",
				CurrencyOptions:       map[string]int64{"eur": 4500},
			}
			gomega.Expect(protovalidate.Validate(promotionCode(spec))).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a missing coupon", func() {
			spec := launch25()
			spec.Coupon = nil
			gomega.Expect(protovalidate.Validate(promotionCode(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a code with a space or another character, and one over 500 characters", func() {
			for _, code := range []string{"LAUNCH 25", "LAUNCH_25", "LAUNCH25!", strings.Repeat("A", 501)} {
				spec := launch25()
				spec.Code = code
				err := protovalidate.Validate(promotionCode(spec))
				gomega.Expect(err).To(gomega.HaveOccurred(), "code %q", code)
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("letters (a-z, A-Z), digits (0-9) and dashes"))
			}
		})

		ginkgo.It("refuses a customer that is not a customer id, and two customers", func() {
			spec := launch25()
			spec.Customer = "acct_123"
			gomega.Expect(protovalidate.Validate(promotionCode(spec))).NotTo(gomega.Succeed())

			spec = launch25()
			spec.Customer = "cus_123"
			spec.CustomerAccount = "acct_123"
			err := protovalidate.Validate(promotionCode(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("set customer or customer_account, not both"))
		})

		ginkgo.It("refuses a minimum amount without its currency, and currency options without a minimum", func() {
			spec := launch25()
			spec.Restrictions = &StripePromotionCodeRestrictions{MinimumAmount: int64Ptr(5000)}
			err := protovalidate.Validate(promotionCode(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("minimum_amount and minimum_amount_currency go together"))

			spec = launch25()
			spec.Restrictions = &StripePromotionCodeRestrictions{CurrencyOptions: map[string]int64{"eur": 4500}}
			gomega.Expect(protovalidate.Validate(promotionCode(spec))).NotTo(gomega.Succeed())

			spec = launch25()
			spec.Restrictions = &StripePromotionCodeRestrictions{MinimumAmount: int64Ptr(5000), MinimumAmountCurrency: "usd", CurrencyOptions: map[string]int64{"usd": 5000}}
			gomega.Expect(protovalidate.Validate(promotionCode(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a zero redemption limit or expiry", func() {
			spec := launch25()
			spec.MaxRedemptions = int64Ptr(0)
			gomega.Expect(protovalidate.Validate(promotionCode(spec))).NotTo(gomega.Succeed())

			spec = launch25()
			spec.ExpiresAt = int64Ptr(0)
			gomega.Expect(protovalidate.Validate(promotionCode(spec))).NotTo(gomega.Succeed())
		})
	})
})
