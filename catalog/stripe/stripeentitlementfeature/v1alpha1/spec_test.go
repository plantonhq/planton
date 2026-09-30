package stripeentitlementfeaturev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestStripeEntitlementFeature(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripeEntitlementFeature Suite")
}

func feature(spec *StripeEntitlementFeatureSpec) *StripeEntitlementFeature {
	return &StripeEntitlementFeature{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripeEntitlementFeature",
		Metadata:   &shared.CloudResourceMetadata{Name: "api-access"},
		Spec:       spec,
	}
}

// apiAccess is the minimal valid spec: a lookup key and a name.
func apiAccess() *StripeEntitlementFeatureSpec {
	return &StripeEntitlementFeatureSpec{LookupKey: "api-access", Name: "API access"}
}

var _ = ginkgo.Describe("StripeEntitlementFeature Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts the minimal feature", func() {
			gomega.Expect(protovalidate.Validate(feature(apiAccess()))).To(gomega.Succeed())
		})

		ginkgo.It("accepts metadata and an 80-character lookup key", func() {
			spec := apiAccess()
			spec.LookupKey = strings.Repeat("k", 80)
			spec.Metadata = map[string]string{"team": "platform"}
			gomega.Expect(protovalidate.Validate(feature(spec))).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a missing lookup key", func() {
			spec := apiAccess()
			spec.LookupKey = ""
			gomega.Expect(protovalidate.Validate(feature(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a lookup key over 80 characters", func() {
			spec := apiAccess()
			spec.LookupKey = strings.Repeat("k", 81)
			gomega.Expect(protovalidate.Validate(feature(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a missing name", func() {
			spec := apiAccess()
			spec.Name = ""
			gomega.Expect(protovalidate.Validate(feature(spec))).NotTo(gomega.Succeed())
		})
	})
})
