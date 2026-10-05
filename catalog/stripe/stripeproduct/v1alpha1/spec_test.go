package stripeproductv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestStripeProduct(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripeProduct Suite")
}

func product(spec *StripeProductSpec) *StripeProduct {
	return &StripeProduct{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripeProduct",
		Metadata:   &shared.CatalogObjectMetadata{Name: "pro-plan"},
		Spec:       spec,
	}
}

// pro is the minimal valid spec: a name.
func pro() *StripeProductSpec {
	return &StripeProductSpec{Name: "Pro"}
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
		ValueFrom: &foreignkeyv1.ValueFromRef{Kind: catalogkind.CatalogKind_StripeEntitlementFeature, Name: name},
	}}
}

func boolPtr(v bool) *bool { return &v }

var _ = ginkgo.Describe("StripeProduct Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts the minimal product", func() {
			gomega.Expect(protovalidate.Validate(product(pro()))).To(gomega.Succeed())
		})

		ginkgo.It("accepts every field on a service, with features by literal and by reference", func() {
			spec := pro()
			spec.Description = "For growing teams"
			spec.Active = boolPtr(true)
			spec.Type = StripeProductSpec_service
			spec.Images = []string{"https://example.com/pro.png"}
			spec.MarketingFeatures = []*StripeProductMarketingFeature{{Name: "Unlimited projects"}, {Name: "SSO"}}
			spec.StatementDescriptor = "ACME PRO"
			spec.TaxCode = "txcd_10103001"
			spec.UnitLabel = "seat"
			spec.Url = "https://example.com/pricing"
			spec.Metadata = map[string]string{"tier": "pro"}
			spec.Features = []*foreignkeyv1.StringValueOrRef{literal("feat_123"), ref("api-access"), ref("sso")}
			gomega.Expect(protovalidate.Validate(product(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts a shippable good with package dimensions", func() {
			spec := pro()
			spec.Type = StripeProductSpec_good
			spec.Shippable = boolPtr(true)
			spec.PackageDimensions = &StripeProductPackageDimensions{Height: 2, Length: 10, Weight: 16, Width: 8}
			gomega.Expect(protovalidate.Validate(product(spec))).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a missing name", func() {
			gomega.Expect(protovalidate.Validate(product(&StripeProductSpec{}))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses more than 8 images, or an image that is not a URL", func() {
			spec := pro()
			for i := 0; i < 9; i++ {
				spec.Images = append(spec.Images, "https://example.com/"+strings.Repeat("i", i+1)+".png")
			}
			gomega.Expect(protovalidate.Validate(product(spec))).NotTo(gomega.Succeed())

			spec = pro()
			spec.Images = []string{"pro.png"}
			gomega.Expect(protovalidate.Validate(product(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses more than 15 marketing features, or one over 80 characters", func() {
			spec := pro()
			for i := 0; i < 16; i++ {
				spec.MarketingFeatures = append(spec.MarketingFeatures, &StripeProductMarketingFeature{Name: "line"})
			}
			gomega.Expect(protovalidate.Validate(product(spec))).NotTo(gomega.Succeed())

			spec = pro()
			spec.MarketingFeatures = []*StripeProductMarketingFeature{{Name: strings.Repeat("x", 81)}}
			gomega.Expect(protovalidate.Validate(product(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a statement descriptor Stripe would reject", func() {
			for _, descriptor := range []string{strings.Repeat("A", 23), "12345", `ACME "PRO"`, "ACME<PRO>", `ACME\PRO`, "ACME'S"} {
				spec := pro()
				spec.StatementDescriptor = descriptor
				err := protovalidate.Validate(product(spec))
				gomega.Expect(err).To(gomega.HaveOccurred(), "descriptor %q", descriptor)
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("statement_descriptor is up to 22 characters"))
			}
		})

		ginkgo.It("refuses a statement descriptor or unit label on a good", func() {
			spec := pro()
			spec.Type = StripeProductSpec_good
			spec.UnitLabel = "seat"
			err := protovalidate.Validate(product(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("apply only to a service"))
		})

		ginkgo.It("refuses a tax code that is not a Stripe tax code id", func() {
			spec := pro()
			spec.TaxCode = "10103001"
			gomega.Expect(protovalidate.Validate(product(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a URL that is not http or https", func() {
			spec := pro()
			spec.Url = "example.com/pricing"
			gomega.Expect(protovalidate.Validate(product(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a package dimension of zero", func() {
			spec := pro()
			spec.PackageDimensions = &StripeProductPackageDimensions{Height: 2, Length: 10, Weight: 0, Width: 8}
			gomega.Expect(protovalidate.Validate(product(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a feature that is not an entitlement feature id", func() {
			spec := pro()
			spec.Features = []*foreignkeyv1.StringValueOrRef{literal("prod_123")}
			err := protovalidate.Validate(product(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("entitlement feature id"))
		})

		ginkgo.It("refuses a feature granted twice", func() {
			spec := pro()
			spec.Features = []*foreignkeyv1.StringValueOrRef{literal("feat_123"), literal("feat_123")}
			err := protovalidate.Validate(product(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("each feature is granted once"))
		})
	})
})
