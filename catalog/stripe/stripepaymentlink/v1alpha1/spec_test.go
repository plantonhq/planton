package stripepaymentlinkv1alpha1

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

func TestStripePaymentLink(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripePaymentLink Suite")
}

func link(spec *StripePaymentLinkSpec) *StripePaymentLink {
	return &StripePaymentLink{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripePaymentLink",
		Metadata:   &shared.CloudResourceMetadata{Name: "pro-monthly-link"},
		Spec:       spec,
	}
}

func int64Ptr(v int64) *int64       { return &v }
func float64Ptr(v float64) *float64 { return &v }
func boolPtr(v bool) *bool          { return &v }

func priceRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
		ValueFrom: &foreignkeyv1.ValueFromRef{Kind: cloudresourcekind.CloudResourceKind_StripePrice, Name: name},
	}}
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

// proMonthly sells one referenced price.
func proMonthly() *StripePaymentLinkSpec {
	return &StripePaymentLinkSpec{
		LineItems: []*StripePaymentLinkLineItem{{Price: priceRef("pro-monthly"), Quantity: 1}},
	}
}

func items(n int) []*StripePaymentLinkLineItem {
	out := make([]*StripePaymentLinkLineItem, n)
	for i := range out {
		out[i] = &StripePaymentLinkLineItem{Price: literal("price_123"), Quantity: 1}
	}
	return out
}

var _ = ginkgo.Describe("StripePaymentLink Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts one referenced price", func() {
			gomega.Expect(protovalidate.Validate(link(proMonthly()))).To(gomega.Succeed())
		})

		ginkgo.It("accepts a subscription page with a trial, promotion codes, terms and a redirect", func() {
			spec := proMonthly()
			spec.AllowPromotionCodes = boolPtr(true)
			spec.ConsentCollection = &StripePaymentLinkConsentCollection{TermsOfService: StripePaymentLinkTermsOfService_required}
			spec.AfterCompletion = &StripePaymentLinkAfterCompletion{Behavior: &StripePaymentLinkAfterCompletion_Redirect{Redirect: &StripePaymentLinkRedirect{Url: "https://example.com/welcome?session={CHECKOUT_SESSION_ID}"}}}
			spec.SubscriptionData = &StripePaymentLinkSubscriptionData{
				TrialPeriodDays: int64Ptr(14),
				TrialSettings: &StripePaymentLinkTrialSettings{EndBehavior: &StripePaymentLinkTrialEndBehavior{
					MissingPaymentMethod: StripePaymentLinkTrialEndBehavior_cancel,
				}},
			}
			spec.PaymentMethodCollection = StripePaymentLinkPaymentMethodCollection_if_required
			spec.InactiveMessage = "This offer has ended."
			gomega.Expect(protovalidate.Validate(link(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts a one-time page with shipping, custom fields, optional items and Connect fields", func() {
			spec := &StripePaymentLinkSpec{
				LineItems: []*StripePaymentLinkLineItem{{
					Price:              literal("price_123"),
					Quantity:           2,
					AdjustableQuantity: &StripePaymentLinkAdjustableQuantity{Enabled: true, Minimum: int64Ptr(1), Maximum: int64Ptr(10)},
				}},
				OptionalItems:             []*StripePaymentLinkOptionalItem{{Price: priceRef("gift-wrap"), Quantity: 1}},
				ShippingAddressCollection: &StripePaymentLinkShippingAddressCollection{AllowedCountries: []string{"US", "CA"}},
				ShippingOptions: []*StripePaymentLinkShippingOption{{ShippingRate: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{Kind: cloudresourcekind.CloudResourceKind_StripeShippingRate, Name: "standard-shipping"},
					},
				}}},
				CustomFields: []*StripePaymentLinkCustomField{
					{Key: "engraving", Label: "Engraving", Type: &StripePaymentLinkCustomField_Text{Text: &StripePaymentLinkTextBounds{MaximumLength: int64Ptr(20)}}, Optional: boolPtr(true)},
					{Key: "seats", Label: "Seats", Type: &StripePaymentLinkCustomField_Numeric{Numeric: &StripePaymentLinkTextBounds{}}},
					{Key: "size", Label: "Size", Type: &StripePaymentLinkCustomField_Dropdown{Dropdown: &StripePaymentLinkDropdown{Options: []*StripePaymentLinkDropdownOption{{Label: "Small", Value: "s"}}}}},
				},
				AfterCompletion:          &StripePaymentLinkAfterCompletion{Behavior: &StripePaymentLinkAfterCompletion_HostedConfirmation{HostedConfirmation: &StripePaymentLinkHostedConfirmation{CustomMessage: "Thanks!"}}},
				BillingAddressCollection: StripePaymentLinkSpec_required,
				CustomerCreation:         StripePaymentLinkSpec_always,
				SubmitType:               StripePaymentLinkSubmitType_pay,
				AutomaticTax:             &StripePaymentLinkAutomaticTax{Enabled: true, Liability: &StripePaymentLinkAccountRef{Account: "acct_123"}},
				InvoiceCreation:          &StripePaymentLinkInvoiceCreation{Enabled: true, InvoiceData: &StripePaymentLinkInvoiceData{Footer: "Thank you", Issuer: &StripePaymentLinkAccountRef{}}},
				ApplicationFeeAmount:     int64Ptr(100),
				OnBehalfOf:               "acct_123",
				TransferData:             &StripePaymentLinkTransferData{Destination: "acct_123"},
				Restrictions:             &StripePaymentLinkRestrictions{CompletedSessions: &StripePaymentLinkCompletedSessions{Limit: 100}},
			}
			gomega.Expect(protovalidate.Validate(link(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts twenty items in all", func() {
			spec := &StripePaymentLinkSpec{LineItems: items(10)}
			for i := 0; i < 10; i++ {
				spec.OptionalItems = append(spec.OptionalItems, &StripePaymentLinkOptionalItem{Price: literal("price_456"), Quantity: 1})
			}
			gomega.Expect(protovalidate.Validate(link(spec))).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a link with no line items, or more than 20", func() {
			gomega.Expect(protovalidate.Validate(link(&StripePaymentLinkSpec{}))).NotTo(gomega.Succeed())
			gomega.Expect(protovalidate.Validate(link(&StripePaymentLinkSpec{LineItems: items(21)}))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses more than 10 optional items, or more than 20 items together", func() {
			spec := proMonthly()
			for i := 0; i < 11; i++ {
				spec.OptionalItems = append(spec.OptionalItems, &StripePaymentLinkOptionalItem{Price: literal("price_456"), Quantity: 1})
			}
			gomega.Expect(protovalidate.Validate(link(spec))).NotTo(gomega.Succeed())

			spec = &StripePaymentLinkSpec{LineItems: items(15)}
			for i := 0; i < 6; i++ {
				spec.OptionalItems = append(spec.OptionalItems, &StripePaymentLinkOptionalItem{Price: literal("price_456"), Quantity: 1})
			}
			err := protovalidate.Validate(link(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("at most 20 prices"))
		})

		ginkgo.It("refuses a line item with no price, a literal that is not a price id, or a zero quantity", func() {
			spec := proMonthly()
			spec.LineItems[0].Price = nil
			gomega.Expect(protovalidate.Validate(link(spec))).NotTo(gomega.Succeed())

			spec = proMonthly()
			spec.LineItems[0].Price = literal("prod_123")
			err := protovalidate.Validate(link(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("Stripe price id"))

			spec = proMonthly()
			spec.LineItems[0].Quantity = 0
			gomega.Expect(protovalidate.Validate(link(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a shipping option literal that is not a shipping rate id", func() {
			spec := proMonthly()
			spec.ShippingOptions = []*StripePaymentLinkShippingOption{{ShippingRate: literal("rate_123")}}
			err := protovalidate.Validate(link(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("Stripe shipping rate id"))
		})

		ginkgo.It("refuses more than 3 custom fields, a field of no type, and a malformed key", func() {
			field := func() *StripePaymentLinkCustomField {
				return &StripePaymentLinkCustomField{Key: "note", Label: "Note", Type: &StripePaymentLinkCustomField_Text{Text: &StripePaymentLinkTextBounds{}}}
			}
			spec := proMonthly()
			spec.CustomFields = []*StripePaymentLinkCustomField{field(), field(), field(), field()}
			gomega.Expect(protovalidate.Validate(link(spec))).NotTo(gomega.Succeed())

			spec = proMonthly()
			spec.CustomFields = []*StripePaymentLinkCustomField{{Key: "note", Label: "Note"}}
			err := protovalidate.Validate(link(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("custom_fields[0].type"))

			spec = proMonthly()
			f := field()
			f.Key = "gift note"
			spec.CustomFields = []*StripePaymentLinkCustomField{f}
			gomega.Expect(protovalidate.Validate(link(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("accepts Stripe's confirmation page with no message", func() {
			spec := proMonthly()
			spec.AfterCompletion = &StripePaymentLinkAfterCompletion{Behavior: &StripePaymentLinkAfterCompletion_HostedConfirmation{HostedConfirmation: &StripePaymentLinkHostedConfirmation{}}}
			gomega.Expect(protovalidate.Validate(link(spec))).To(gomega.Succeed())
		})

		ginkgo.It("refuses after_completion with no behavior, and a redirect that is not a URL", func() {
			spec := proMonthly()
			spec.AfterCompletion = &StripePaymentLinkAfterCompletion{}
			err := protovalidate.Validate(link(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("after_completion.behavior"))

			spec = proMonthly()
			spec.AfterCompletion = &StripePaymentLinkAfterCompletion{Behavior: &StripePaymentLinkAfterCompletion_Redirect{Redirect: &StripePaymentLinkRedirect{Url: "example.com"}}}
			gomega.Expect(protovalidate.Validate(link(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses both platform fees, a fee percent out of range, and malformed connected accounts", func() {
			spec := proMonthly()
			spec.ApplicationFeeAmount = int64Ptr(100)
			spec.ApplicationFeePercent = float64Ptr(10)
			err := protovalidate.Validate(link(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("one kind of platform fee"))

			spec = proMonthly()
			spec.ApplicationFeePercent = float64Ptr(101)
			gomega.Expect(protovalidate.Validate(link(spec))).NotTo(gomega.Succeed())

			spec = proMonthly()
			spec.OnBehalfOf = "cus_123"
			gomega.Expect(protovalidate.Validate(link(spec))).NotTo(gomega.Succeed())

			spec = proMonthly()
			spec.TransferData = &StripePaymentLinkTransferData{Destination: "cus_123"}
			gomega.Expect(protovalidate.Validate(link(spec))).NotTo(gomega.Succeed())

			spec = proMonthly()
			spec.AutomaticTax = &StripePaymentLinkAutomaticTax{Enabled: true, Liability: &StripePaymentLinkAccountRef{Account: "cus_123"}}
			gomega.Expect(protovalidate.Validate(link(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses an inactive message over 500 characters and a completed-session limit of 0", func() {
			spec := proMonthly()
			spec.InactiveMessage = strings.Repeat("x", 501)
			gomega.Expect(protovalidate.Validate(link(spec))).NotTo(gomega.Succeed())

			spec = proMonthly()
			spec.Restrictions = &StripePaymentLinkRestrictions{CompletedSessions: &StripePaymentLinkCompletedSessions{}}
			gomega.Expect(protovalidate.Validate(link(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a shipping country that is not a two-letter capital code", func() {
			spec := proMonthly()
			spec.ShippingAddressCollection = &StripePaymentLinkShippingAddressCollection{AllowedCountries: []string{"usa"}}
			gomega.Expect(protovalidate.Validate(link(spec))).NotTo(gomega.Succeed())
		})
	})
})
