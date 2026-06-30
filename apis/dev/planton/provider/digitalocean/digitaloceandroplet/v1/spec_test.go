package digitaloceandropletv1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestDigitalOceanDropletSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanDropletSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("DigitalOceanDropletSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("digitalocean_droplet", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &DigitalOceanDroplet{
					ApiVersion: "digital-ocean.planton.dev/v1",
					Kind:       "DigitalOceanDroplet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-droplet",
					},
					Spec: &DigitalOceanDropletSpec{
						DropletName: "test-droplet",
						Region:      digitalocean.DigitalOceanRegion_nyc3,
						Size:        "s-2vcpu-4gb",
						Image:       "ubuntu-22-04-x64",
						Vpc: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-vpc-id"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
