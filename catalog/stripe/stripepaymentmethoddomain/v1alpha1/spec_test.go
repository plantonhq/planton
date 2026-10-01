package stripepaymentmethoddomainv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestStripePaymentMethodDomain(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripePaymentMethodDomain Suite")
}

func domain(spec *StripePaymentMethodDomainSpec) *StripePaymentMethodDomain {
	return &StripePaymentMethodDomain{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripePaymentMethodDomain",
		Metadata:   &shared.CloudResourceMetadata{Name: "checkout-domain"},
		Spec:       spec,
	}
}

func boolPtr(v bool) *bool { return &v }

var _ = ginkgo.Describe("StripePaymentMethodDomain Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a hostname", func() {
			gomega.Expect(protovalidate.Validate(domain(&StripePaymentMethodDomainSpec{DomainName: "pay.example.com"}))).To(gomega.Succeed())
		})

		ginkgo.It("accepts a disabled registration", func() {
			spec := &StripePaymentMethodDomainSpec{DomainName: "example.com", Enabled: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(domain(spec))).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a missing domain", func() {
			gomega.Expect(protovalidate.Validate(domain(&StripePaymentMethodDomainSpec{}))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a URL or a path where a hostname belongs", func() {
			for _, name := range []string{"https://pay.example.com", "pay.example.com/checkout", "pay example.com"} {
				gomega.Expect(protovalidate.Validate(domain(&StripePaymentMethodDomainSpec{DomainName: name}))).NotTo(gomega.Succeed(), "domain %q", name)
			}
		})
	})
})
