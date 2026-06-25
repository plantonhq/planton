package cloudflareturnstilewidgetv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func validWidget() *CloudflareTurnstileWidget {
	return &CloudflareTurnstileWidget{
		ApiVersion: "cloudflare.openmcf.org/v1",
		Kind:       "CloudflareTurnstileWidget",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-widget"},
		Spec: &CloudflareTurnstileWidgetSpec{
			AccountId: validAccountID,
			Name:      "login-form",
			Domains:   []string{"example.com"},
			Mode:      "managed",
		},
	}
}

func TestCloudflareTurnstileWidgetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareTurnstileWidgetSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareTurnstileWidgetSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal valid widget", func() {
			gomega.Expect(protovalidate.Validate(validWidget())).To(gomega.BeNil())
		})

		ginkgo.It("accepts each mode value", func() {
			for _, m := range []string{"non-interactive", "invisible", "managed"} {
				in := validWidget()
				in.Spec.Mode = m
				gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
			}
		})

		ginkgo.It("accepts optional clearance, region, and ENT flags", func() {
			in := validWidget()
			in.Spec.ClearanceLevel = "interactive"
			in.Spec.Region = "china"
			in.Spec.BotFightMode = true
			in.Spec.EphemeralId = true
			in.Spec.Offlabel = true
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a non-hex account_id", func() {
			in := validWidget()
			in.Spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing name", func() {
			in := validWidget()
			in.Spec.Name = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an empty domains list", func() {
			in := validWidget()
			in.Spec.Domains = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing mode", func() {
			in := validWidget()
			in.Spec.Mode = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid mode", func() {
			in := validWidget()
			in.Spec.Mode = "automatic"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid clearance_level", func() {
			in := validWidget()
			in.Spec.ClearanceLevel = "always"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid region", func() {
			in := validWidget()
			in.Spec.Region = "moon"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
