package stripetaxratev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestStripeTaxRate(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripeTaxRate Suite")
}

func taxRate(spec *StripeTaxRateSpec) *StripeTaxRate {
	return &StripeTaxRate{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripeTaxRate",
		Metadata:   &shared.CatalogObjectMetadata{Name: "de-vat"},
		Spec:       spec,
	}
}

func boolPtr(v bool) *bool { return &v }

// deVat is German VAT, 19%, exclusive.
func deVat() *StripeTaxRateSpec {
	return &StripeTaxRateSpec{DisplayName: "VAT", Percentage: 19, Country: "DE", TaxType: StripeTaxRateSpec_vat}
}

var _ = ginkgo.Describe("StripeTaxRate Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts an exclusive VAT rate", func() {
			gomega.Expect(protovalidate.Validate(taxRate(deVat()))).To(gomega.Succeed())
		})

		ginkgo.It("accepts an inclusive state rate with every in-place field, and a zero rate", func() {
			spec := &StripeTaxRateSpec{
				DisplayName:  "Sales tax",
				Percentage:   8.875,
				Inclusive:    true,
				Country:      "US",
				State:        "NY",
				Jurisdiction: "New York",
				Description:  "NYC combined sales tax",
				TaxType:      StripeTaxRateSpec_sales_tax,
				Active:       boolPtr(true),
				Metadata:     map[string]string{"region": "ny"},
			}
			gomega.Expect(protovalidate.Validate(taxRate(spec))).To(gomega.Succeed())

			spec = deVat()
			spec.Percentage = 0
			gomega.Expect(protovalidate.Validate(taxRate(spec))).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a rate with no display name", func() {
			spec := deVat()
			spec.DisplayName = ""
			gomega.Expect(protovalidate.Validate(taxRate(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a percentage out of range", func() {
			for _, pct := range []float64{-1, 100.01} {
				spec := deVat()
				spec.Percentage = pct
				gomega.Expect(protovalidate.Validate(taxRate(spec))).NotTo(gomega.Succeed(), "percentage %v", pct)
			}
		})

		ginkgo.It("refuses a country that is not a two-letter capital code", func() {
			for _, country := range []string{"de", "DEU", "Germany"} {
				spec := deVat()
				spec.Country = country
				err := protovalidate.Validate(taxRate(spec))
				gomega.Expect(err).To(gomega.HaveOccurred(), "country %q", country)
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("two-letter ISO country code"))
			}
		})

		ginkgo.It("refuses a tax type outside the provider's set", func() {
			spec := deVat()
			spec.TaxType = StripeTaxRateSpec_TaxType(99)
			gomega.Expect(protovalidate.Validate(taxRate(spec))).NotTo(gomega.Succeed())
		})
	})
})
