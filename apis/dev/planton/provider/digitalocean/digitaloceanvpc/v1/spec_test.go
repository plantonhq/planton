package digitaloceanvpcv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestDigitalOceanVpcSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanVpcSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("DigitalOceanVpcSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("digitalocean_vpc", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &DigitalOceanVpc{
					ApiVersion: "digital-ocean.planton.dev/v1",
					Kind:       "DigitalOceanVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vpc",
					},
					Spec: &DigitalOceanVpcSpec{
						Region:      digitalocean.DigitalOceanRegion_nyc3,
						IpRangeCidr: "10.10.0.0/16",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
