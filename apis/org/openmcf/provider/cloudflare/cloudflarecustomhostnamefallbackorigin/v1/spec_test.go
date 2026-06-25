package cloudflarecustomhostnamefallbackoriginv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func ref(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func validFallbackOrigin() *CloudflareCustomHostnameFallbackOrigin {
	return &CloudflareCustomHostnameFallbackOrigin{
		ApiVersion: "cloudflare.openmcf.org/v1",
		Kind:       "CloudflareCustomHostnameFallbackOrigin",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-fallback-origin"},
		Spec: &CloudflareCustomHostnameFallbackOriginSpec{
			ZoneId: ref("0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"),
			Origin: ref("origin.helpdesk.io"),
		},
	}
}

func TestCloudflareCustomHostnameFallbackOriginSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareCustomHostnameFallbackOriginSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareCustomHostnameFallbackOriginSpec validations", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a valid fallback origin", func() {
			gomega.Expect(protovalidate.Validate(validFallbackOrigin())).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing zone_id", func() {
			in := validFallbackOrigin()
			in.Spec.ZoneId = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing origin", func() {
			in := validFallbackOrigin()
			in.Spec.Origin = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an empty origin value", func() {
			in := validFallbackOrigin()
			in.Spec.Origin = &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: ""}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
