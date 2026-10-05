package stripetaxregistrationv1alpha1

import (
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestStripeTaxRegistration(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripeTaxRegistration Suite")
}

func registration(spec *StripeTaxRegistrationSpec) *StripeTaxRegistration {
	return &StripeTaxRegistration{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripeTaxRegistration",
		Metadata:   &shared.CatalogObjectMetadata{Name: "de-oss"},
		Spec:       spec,
	}
}

func int64Ptr(v int64) *int64 { return &v }

// startOf2100 is 2100-01-01T00:00:00Z.
const startOf2100 = 4102444800

// deOss is the EU One-Stop Shop, declared in Germany.
func deOss() *StripeTaxRegistrationSpec {
	return &StripeTaxRegistrationSpec{Country: "DE", Type: StripeTaxRegistrationSpec_oss_union, ActiveFrom: startOf2100}
}

// txSalesTax is Texas state sales tax.
func txSalesTax() *StripeTaxRegistrationSpec {
	return &StripeTaxRegistrationSpec{Country: "US", Type: StripeTaxRegistrationSpec_state_sales_tax, State: "TX", ActiveFrom: startOf2100}
}

func expectValid(spec *StripeTaxRegistrationSpec) {
	gomega.ExpectWithOffset(1, protovalidate.Validate(registration(spec))).To(gomega.Succeed())
}

// violatedRules lists the rule ids a validation error names, so a case pins the rule it exists
// for rather than any failure at all.
func violatedRules(err error) []string {
	var validationErr *protovalidate.ValidationError
	if !errors.As(err, &validationErr) {
		return nil
	}
	ids := make([]string, 0, len(validationErr.Violations))
	for _, violation := range validationErr.Violations {
		ids = append(ids, violation.Proto.GetRuleId())
	}
	return ids
}

func expectRefused(spec *StripeTaxRegistrationSpec, ruleID string) {
	err := protovalidate.Validate(registration(spec))
	gomega.ExpectWithOffset(1, err).To(gomega.HaveOccurred())
	gomega.ExpectWithOffset(1, violatedRules(err)).To(gomega.ContainElement(ruleID))
}

var _ = ginkgo.Describe("StripeTaxRegistration Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts every EU type, and a standard registration's schemes", func() {
			expectValid(deOss())
			for _, t := range []StripeTaxRegistrationSpec_Type{
				StripeTaxRegistrationSpec_standard, StripeTaxRegistrationSpec_oss_non_union, StripeTaxRegistrationSpec_ioss,
			} {
				spec := deOss()
				spec.Type = t
				expectValid(spec)
			}
			for _, s := range []StripeTaxRegistrationPlaceOfSupplyScheme{
				StripeTaxRegistrationPlaceOfSupplyScheme_standard,
				StripeTaxRegistrationPlaceOfSupplyScheme_small_seller,
				StripeTaxRegistrationPlaceOfSupplyScheme_inbound_goods,
			} {
				spec := deOss()
				spec.Type = StripeTaxRegistrationSpec_standard
				spec.PlaceOfSupplyScheme = s
				expectValid(spec)
			}
		})

		ginkgo.It("accepts a standard-only and a simplified-only country", func() {
			expectValid(&StripeTaxRegistrationSpec{
				Country: "GB", Type: StripeTaxRegistrationSpec_standard, ActiveFrom: startOf2100,
				PlaceOfSupplyScheme: StripeTaxRegistrationPlaceOfSupplyScheme_inbound_goods,
			})
			expectValid(&StripeTaxRegistrationSpec{Country: "AL", Type: StripeTaxRegistrationSpec_standard, ActiveFrom: startOf2100})
			expectValid(&StripeTaxRegistrationSpec{Country: "IN", Type: StripeTaxRegistrationSpec_simplified, ActiveFrom: startOf2100})
		})

		ginkgo.It("accepts Canada's three types, a province with province_standard", func() {
			expectValid(&StripeTaxRegistrationSpec{Country: "CA", Type: StripeTaxRegistrationSpec_standard, ActiveFrom: startOf2100})
			expectValid(&StripeTaxRegistrationSpec{Country: "CA", Type: StripeTaxRegistrationSpec_simplified, ActiveFrom: startOf2100})
			expectValid(&StripeTaxRegistrationSpec{
				Country: "CA", Type: StripeTaxRegistrationSpec_province_standard, Province: "QC", ActiveFrom: startOf2100,
			})
		})

		ginkgo.It("accepts US state registrations, local taxes with a jurisdiction, and elections", func() {
			expectValid(txSalesTax())
			for _, t := range []StripeTaxRegistrationSpec_Type{
				StripeTaxRegistrationSpec_state_communications_tax, StripeTaxRegistrationSpec_state_retail_delivery_fee,
			} {
				spec := txSalesTax()
				spec.Type = t
				expectValid(spec)
			}
			for _, t := range []StripeTaxRegistrationSpec_Type{
				StripeTaxRegistrationSpec_local_amusement_tax, StripeTaxRegistrationSpec_local_lease_tax,
			} {
				spec := txSalesTax()
				spec.State = "IL"
				spec.Type = t
				spec.Jurisdiction = "14000"
				expectValid(spec)
			}
			spec := txSalesTax()
			spec.StateSalesTaxElections = []*StripeTaxRegistrationStateSalesTaxElection{
				{Type: StripeTaxRegistrationStateSalesTaxElection_local_use_tax, Jurisdiction: "02154"},
				{Type: StripeTaxRegistrationStateSalesTaxElection_simplified_sellers_use_tax},
			}
			expectValid(spec)
		})

		ginkgo.It("accepts an expiry after the start", func() {
			spec := deOss()
			spec.ExpiresAt = int64Ptr(startOf2100 + 86400)
			expectValid(spec)
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a missing, lowercase or unsupported country", func() {
			spec := deOss()
			spec.Country = ""
			gomega.Expect(protovalidate.Validate(registration(spec))).NotTo(gomega.Succeed())
			spec.Country = "de"
			expectRefused(spec, "spec.country.format")
			spec.Country = "XX"
			expectRefused(spec, "spec.country.supported")
		})

		ginkgo.It("refuses a registration with no type, or no start", func() {
			spec := deOss()
			spec.Type = StripeTaxRegistrationSpec_type_unspecified
			gomega.Expect(protovalidate.Validate(registration(spec))).NotTo(gomega.Succeed())
			spec = deOss()
			spec.ActiveFrom = 0
			gomega.Expect(protovalidate.Validate(registration(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a type the country does not take", func() {
			spec := deOss()
			spec.Type = StripeTaxRegistrationSpec_simplified
			expectRefused(spec, "spec.type.eu")

			expectRefused(&StripeTaxRegistrationSpec{Country: "GB", Type: StripeTaxRegistrationSpec_oss_union, ActiveFrom: startOf2100}, "spec.type.standard_only")
			expectRefused(&StripeTaxRegistrationSpec{Country: "IN", Type: StripeTaxRegistrationSpec_standard, ActiveFrom: startOf2100}, "spec.type.simplified_only")
			expectRefused(&StripeTaxRegistrationSpec{Country: "CA", Type: StripeTaxRegistrationSpec_state_sales_tax, ActiveFrom: startOf2100}, "spec.type.ca")

			spec = txSalesTax()
			spec.Type = StripeTaxRegistrationSpec_standard
			expectRefused(spec, "spec.type.us")
		})

		ginkgo.It("refuses a place-of-supply scheme outside a standard registration that has one", func() {
			spec := deOss()
			spec.PlaceOfSupplyScheme = StripeTaxRegistrationPlaceOfSupplyScheme_standard
			expectRefused(spec, "spec.place_of_supply_scheme.where")

			expectRefused(&StripeTaxRegistrationSpec{
				Country: "IN", Type: StripeTaxRegistrationSpec_simplified, ActiveFrom: startOf2100,
				PlaceOfSupplyScheme: StripeTaxRegistrationPlaceOfSupplyScheme_standard,
			}, "spec.place_of_supply_scheme.where")

			expectRefused(&StripeTaxRegistrationSpec{
				Country: "GB", Type: StripeTaxRegistrationSpec_standard, ActiveFrom: startOf2100,
				PlaceOfSupplyScheme: StripeTaxRegistrationPlaceOfSupplyScheme_small_seller,
			}, "spec.place_of_supply_scheme.small_seller_eu")
		})

		ginkgo.It("refuses a province, state or jurisdiction where it does not belong, or missing where it does", func() {
			expectRefused(&StripeTaxRegistrationSpec{Country: "CA", Type: StripeTaxRegistrationSpec_province_standard, ActiveFrom: startOf2100}, "spec.province.province_standard")
			expectRefused(&StripeTaxRegistrationSpec{Country: "CA", Type: StripeTaxRegistrationSpec_standard, Province: "QC", ActiveFrom: startOf2100}, "spec.province.province_standard")

			spec := txSalesTax()
			spec.State = ""
			expectRefused(spec, "spec.state.us")
			spec = deOss()
			spec.State = "TX"
			expectRefused(spec, "spec.state.us")

			spec = txSalesTax()
			spec.Type = StripeTaxRegistrationSpec_local_amusement_tax
			expectRefused(spec, "spec.jurisdiction.local_tax")
			spec = txSalesTax()
			spec.Jurisdiction = "14000"
			expectRefused(spec, "spec.jurisdiction.local_tax")
		})

		ginkgo.It("refuses elections outside a state sales tax registration, and an election with no type", func() {
			spec := txSalesTax()
			spec.Type = StripeTaxRegistrationSpec_state_communications_tax
			spec.StateSalesTaxElections = []*StripeTaxRegistrationStateSalesTaxElection{
				{Type: StripeTaxRegistrationStateSalesTaxElection_local_use_tax},
			}
			expectRefused(spec, "spec.state_sales_tax_elections.state_sales_tax")

			spec = txSalesTax()
			spec.StateSalesTaxElections = []*StripeTaxRegistrationStateSalesTaxElection{{}}
			gomega.Expect(protovalidate.Validate(registration(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses an expiry at or before the start", func() {
			spec := deOss()
			spec.ExpiresAt = int64Ptr(startOf2100)
			expectRefused(spec, "spec.expires_at.after_active_from")
		})
	})
})
