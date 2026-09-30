package stripebillingportalconfigurationv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestStripeBillingPortalConfiguration(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripeBillingPortalConfiguration Suite")
}

func portal(spec *StripeBillingPortalConfigurationSpec) *StripeBillingPortalConfiguration {
	return &StripeBillingPortalConfiguration{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripeBillingPortalConfiguration",
		Metadata:   &shared.CloudResourceMetadata{Name: "customer-portal"},
		Spec:       spec,
	}
}

// selfServe is the minimal valid spec: invoices and cancellation at period end.
func selfServe() *StripeBillingPortalConfigurationSpec {
	return &StripeBillingPortalConfigurationSpec{
		Features: &StripeBillingPortalFeatures{
			InvoiceHistory: &StripeBillingPortalInvoiceHistory{Enabled: true},
			SubscriptionCancel: &StripeBillingPortalSubscriptionCancel{
				Enabled: true,
				Mode:    StripeBillingPortalSubscriptionCancel_at_period_end,
			},
		},
	}
}

func planSwitching() *StripeBillingPortalSubscriptionUpdate {
	return &StripeBillingPortalSubscriptionUpdate{
		Enabled: true,
		DefaultAllowedUpdates: []StripeBillingPortalSubscriptionUpdate_DefaultAllowedUpdate{
			StripeBillingPortalSubscriptionUpdate_price,
			StripeBillingPortalSubscriptionUpdate_quantity,
		},
		ProrationBehavior: StripeBillingPortalSubscriptionUpdate_create_prorations,
		Products: []*StripeBillingPortalProduct{{
			Product: literal("prod_team"),
			Prices:  []*foreignkeyv1.StringValueOrRef{literal("price_monthly"), literal("price_yearly")},
			AdjustableQuantity: &StripeBillingPortalAdjustableQuantity{
				Enabled: true, Minimum: int64Ptr(1), Maximum: int64Ptr(50),
			},
		}},
		ScheduleAtPeriodEnd: &StripeBillingPortalScheduleAtPeriodEnd{
			Conditions: []*StripeBillingPortalScheduleCondition{{Type: StripeBillingPortalScheduleCondition_decreasing_item_amount}},
		},
	}
}

func int64Ptr(v int64) *int64 { return &v }

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func ref(kind cloudresourcekind.CloudResourceKind, name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
		ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name},
	}}
}

var _ = ginkgo.Describe("StripeBillingPortalConfiguration Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts the minimal portal", func() {
			gomega.Expect(protovalidate.Validate(portal(selfServe()))).To(gomega.Succeed())
		})

		ginkgo.It("accepts every feature, the business profile, the login page, and a payment-method configuration reference", func() {
			spec := selfServe()
			spec.Name = "Customer portal"
			spec.DefaultReturnUrl = "https://app.example.com/billing"
			spec.BusinessProfile = &StripeBillingPortalBusinessProfile{
				Headline:          "Manage your subscription",
				PrivacyPolicyUrl:  "https://example.com/privacy",
				TermsOfServiceUrl: "https://example.com/terms",
			}
			spec.LoginPage = &StripeBillingPortalLoginPage{Enabled: true}
			spec.Features.CustomerUpdate = &StripeBillingPortalCustomerUpdate{
				Enabled: true,
				AllowedUpdates: []StripeBillingPortalCustomerUpdate_AllowedUpdate{
					StripeBillingPortalCustomerUpdate_email, StripeBillingPortalCustomerUpdate_tax_id,
				},
			}
			spec.Features.PaymentMethodUpdate = &StripeBillingPortalPaymentMethodUpdate{
				Enabled: true,
				PaymentMethodConfiguration: literal("pmc_123"),
			}
			spec.Features.SubscriptionCancel.CancellationReason = &StripeBillingPortalCancellationReason{
				Enabled: true,
				Options: []StripeBillingPortalCancellationReason_Option{
					StripeBillingPortalCancellationReason_too_expensive, StripeBillingPortalCancellationReason_other,
				},
			}
			spec.Features.SubscriptionUpdate = planSwitching()
			gomega.Expect(protovalidate.Validate(portal(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts prorations on an immediate cancellation", func() {
			spec := selfServe()
			spec.Features.SubscriptionCancel.Mode = StripeBillingPortalSubscriptionCancel_immediately
			spec.Features.SubscriptionCancel.ProrationBehavior = StripeBillingPortalSubscriptionCancel_create_prorations
			gomega.Expect(protovalidate.Validate(portal(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts products and prices named by reference", func() {
			spec := selfServe()
			update := planSwitching()
			update.Products[0].Product = ref(cloudresourcekind.CloudResourceKind_StripeProduct, "team-plan")
			update.Products[0].Prices = []*foreignkeyv1.StringValueOrRef{
				ref(cloudresourcekind.CloudResourceKind_StripePrice, "team-monthly"),
				ref(cloudresourcekind.CloudResourceKind_StripePrice, "team-yearly"),
			}
			spec.Features.SubscriptionUpdate = update
			gomega.Expect(protovalidate.Validate(portal(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts quantity changes without products", func() {
			spec := selfServe()
			spec.Features.SubscriptionUpdate = &StripeBillingPortalSubscriptionUpdate{
				Enabled: true,
				DefaultAllowedUpdates: []StripeBillingPortalSubscriptionUpdate_DefaultAllowedUpdate{
					StripeBillingPortalSubscriptionUpdate_quantity,
				},
			}
			gomega.Expect(protovalidate.Validate(portal(spec))).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a portal with no features", func() {
			gomega.Expect(protovalidate.Validate(portal(&StripeBillingPortalConfigurationSpec{}))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses prorations on a cancellation at period end", func() {
			spec := selfServe()
			spec.Features.SubscriptionCancel.ProrationBehavior = StripeBillingPortalSubscriptionCancel_create_prorations
			err := protovalidate.Validate(portal(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("only an immediate cancellation leaves"))
		})

		ginkgo.It("refuses price changes with no products to switch between", func() {
			spec := selfServe()
			update := planSwitching()
			update.Products = nil
			spec.Features.SubscriptionUpdate = update
			err := protovalidate.Validate(portal(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("switching prices needs the prices to switch between"))
		})

		ginkgo.It("refuses more than ten products", func() {
			spec := selfServe()
			update := planSwitching()
			for len(update.Products) < 11 {
				update.Products = append(update.Products, &StripeBillingPortalProduct{Product: literal("prod_x"), Prices: []*foreignkeyv1.StringValueOrRef{literal("price_x")}})
			}
			spec.Features.SubscriptionUpdate = update
			gomega.Expect(protovalidate.Validate(portal(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses product and price ids of the wrong kind", func() {
			spec := selfServe()
			update := planSwitching()
			update.Products[0].Product = literal("price_monthly")
			spec.Features.SubscriptionUpdate = update
			gomega.Expect(protovalidate.Validate(portal(spec))).NotTo(gomega.Succeed())

			update = planSwitching()
			update.Products[0].Prices = []*foreignkeyv1.StringValueOrRef{literal("prod_team")}
			spec.Features.SubscriptionUpdate = update
			gomega.Expect(protovalidate.Validate(portal(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a price listed twice", func() {
			spec := selfServe()
			update := planSwitching()
			update.Products[0].Prices = []*foreignkeyv1.StringValueOrRef{literal("price_monthly"), literal("price_monthly")}
			spec.Features.SubscriptionUpdate = update
			err := protovalidate.Validate(portal(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("each price is listed once"))
		})

		ginkgo.It("refuses a quantity range whose minimum exceeds its maximum", func() {
			spec := selfServe()
			update := planSwitching()
			update.Products[0].AdjustableQuantity.Minimum = int64Ptr(10)
			update.Products[0].AdjustableQuantity.Maximum = int64Ptr(5)
			spec.Features.SubscriptionUpdate = update
			err := protovalidate.Validate(portal(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("minimum must not exceed maximum"))
		})

		ginkgo.It("refuses a cancellation reason with no options, or an unset one", func() {
			spec := selfServe()
			spec.Features.SubscriptionCancel.CancellationReason = &StripeBillingPortalCancellationReason{Enabled: true}
			gomega.Expect(protovalidate.Validate(portal(spec))).NotTo(gomega.Succeed())

			spec.Features.SubscriptionCancel.CancellationReason.Options = []StripeBillingPortalCancellationReason_Option{
				StripeBillingPortalCancellationReason_option_unspecified,
			}
			gomega.Expect(protovalidate.Validate(portal(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a schedule condition with no type", func() {
			spec := selfServe()
			update := planSwitching()
			update.ScheduleAtPeriodEnd.Conditions = []*StripeBillingPortalScheduleCondition{{}}
			spec.Features.SubscriptionUpdate = update
			gomega.Expect(protovalidate.Validate(portal(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses URLs that are not http or https", func() {
			spec := selfServe()
			spec.DefaultReturnUrl = "app.example.com/billing"
			gomega.Expect(protovalidate.Validate(portal(spec))).NotTo(gomega.Succeed())

			spec = selfServe()
			spec.BusinessProfile = &StripeBillingPortalBusinessProfile{PrivacyPolicyUrl: "mailto:privacy@example.com"}
			gomega.Expect(protovalidate.Validate(portal(spec))).NotTo(gomega.Succeed())
		})
	})
})
