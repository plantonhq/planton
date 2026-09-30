package striperadarvaluelistv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestStripeRadarValueList(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripeRadarValueList Suite")
}

func list(spec *StripeRadarValueListSpec) *StripeRadarValueList {
	return &StripeRadarValueList{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripeRadarValueList",
		Metadata:   &shared.CloudResourceMetadata{Name: "blocked-countries"},
		Spec:       spec,
	}
}

func typed(itemType StripeRadarValueListSpec_ItemType, items ...string) *StripeRadarValueListSpec {
	return &StripeRadarValueListSpec{Alias: "the_list", Name: "The list", ItemType: itemType, Items: items}
}

var _ = ginkgo.Describe("StripeRadarValueList Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts an empty list of the default type", func() {
			gomega.Expect(protovalidate.Validate(list(&StripeRadarValueListSpec{Alias: "watch", Name: "Watch list"}))).To(gomega.Succeed())
		})

		ginkgo.It("accepts well-formed items of every checked type", func() {
			for _, spec := range []*StripeRadarValueListSpec{
				typed(StripeRadarValueListSpec_country, "KP", "IR"),
				typed(StripeRadarValueListSpec_email, "fraud@example.com"),
				typed(StripeRadarValueListSpec_ip_address, "203.0.113.7", "2001:db8::1"),
				typed(StripeRadarValueListSpec_card_bin, "424242", "42424242"),
				typed(StripeRadarValueListSpec_customer_id, "cus_123"),
				typed(StripeRadarValueListSpec_account, "acct_123"),
				typed(StripeRadarValueListSpec_case_sensitive_string, "Any Text"),
			} {
				gomega.Expect(protovalidate.Validate(list(spec))).To(gomega.Succeed(), "item type %s", spec.ItemType)
			}
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a missing alias or name", func() {
			gomega.Expect(protovalidate.Validate(list(&StripeRadarValueListSpec{Name: "Watch list"}))).NotTo(gomega.Succeed())
			gomega.Expect(protovalidate.Validate(list(&StripeRadarValueListSpec{Alias: "watch"}))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a repeated or empty item", func() {
			gomega.Expect(protovalidate.Validate(list(typed(StripeRadarValueListSpec_string, "a", "a")))).NotTo(gomega.Succeed())
			gomega.Expect(protovalidate.Validate(list(typed(StripeRadarValueListSpec_string, "")))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses items that do not match the list's type, naming the rule", func() {
			cases := map[string]*StripeRadarValueListSpec{
				"two-letter ISO country codes": typed(StripeRadarValueListSpec_country, "USA"),
				"email addresses":              typed(StripeRadarValueListSpec_email, "not-an-email"),
				"IPv4 or IPv6 addresses":       typed(StripeRadarValueListSpec_ip_address, "203.0.113"),
				"first 6 to 8 digits":          typed(StripeRadarValueListSpec_card_bin, "4242"),
				"Stripe customer ids":          typed(StripeRadarValueListSpec_customer_id, "acct_123"),
				"Stripe account ids":           typed(StripeRadarValueListSpec_account, "cus_123"),
			}
			for want, spec := range cases {
				err := protovalidate.Validate(list(spec))
				gomega.Expect(err).To(gomega.HaveOccurred(), "item type %s", spec.ItemType)
				gomega.Expect(err.Error()).To(gomega.ContainSubstring(want))
			}
		})
	})
})
