package cloudflarezerotrustaccessapplicationv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestCloudflareZeroTrustAccessApplicationSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustAccessApplicationSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareZeroTrustAccessApplicationSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("cloudflare_zero_trust_access_application", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &CloudflareZeroTrustAccessApplication{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareZeroTrustAccessApplication",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-access-app",
					},
					Spec: &CloudflareZeroTrustAccessApplicationSpec{
						ApplicationName: "Test Access Application",
						ZoneId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-zone-123",
							},
						},
						Hostname: "app.example.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
