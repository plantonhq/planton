package stripepaymentmethodconfigurationv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestStripePaymentMethodConfiguration(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "StripePaymentMethodConfiguration Suite")
}

func configuration(spec *StripePaymentMethodConfigurationSpec) *StripePaymentMethodConfiguration {
	return &StripePaymentMethodConfiguration{
		ApiVersion: "stripe.planton.dev/v1alpha1",
		Kind:       "StripePaymentMethodConfiguration",
		Metadata:   &shared.CatalogObjectMetadata{Name: "checkout-methods"},
		Spec:       spec,
	}
}

func pref(p StripePaymentMethodPreference_Preference) *StripePaymentMethodPreference {
	return &StripePaymentMethodPreference{Preference: p}
}

var _ = ginkgo.Describe("StripePaymentMethodConfiguration Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts an empty configuration, leaving every method to Stripe", func() {
			gomega.Expect(protovalidate.Validate(configuration(&StripePaymentMethodConfigurationSpec{}))).To(gomega.Succeed())
		})

		ginkgo.It("accepts on, off and none", func() {
			spec := &StripePaymentMethodConfigurationSpec{
				Name:     "Checkout methods",
				Card:     pref(StripePaymentMethodPreference_on),
				Link:     pref(StripePaymentMethodPreference_off),
				ApplePay: pref(StripePaymentMethodPreference_none),
				Parent:   "pmc_platform",
			}
			gomega.Expect(protovalidate.Validate(configuration(spec))).To(gomega.Succeed())
		})

		ginkgo.It("accepts a preference on every one of the 59 methods", func() {
			spec := &StripePaymentMethodConfigurationSpec{}
			methods := 0
			fields := spec.ProtoReflect().Descriptor().Fields()
			for i := 0; i < fields.Len(); i++ {
				field := fields.Get(i)
				if field.Kind() != protoreflect.MessageKind {
					continue
				}
				spec.ProtoReflect().Set(field, protoreflect.ValueOfMessage(pref(StripePaymentMethodPreference_on).ProtoReflect()))
				methods++
			}
			gomega.Expect(methods).To(gomega.Equal(59))
			gomega.Expect(protovalidate.Validate(configuration(spec))).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a method present with no preference", func() {
			spec := &StripePaymentMethodConfigurationSpec{Card: &StripePaymentMethodPreference{}}
			gomega.Expect(protovalidate.Validate(configuration(spec))).NotTo(gomega.Succeed())
		})

		ginkgo.It("refuses a parent that is not a payment method configuration", func() {
			spec := &StripePaymentMethodConfigurationSpec{Parent: "acct_123"}
			err := protovalidate.Validate(configuration(spec))
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("parent is a payment method configuration id"))
		})
	})
})
