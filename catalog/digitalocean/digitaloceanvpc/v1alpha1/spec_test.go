package digitaloceanvpcv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/catalog/digitalocean"
	"github.com/plantonhq/planton/shared"
)

func TestDigitalOceanVpcSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanVpcSpec Custom Validation Tests")
}

// vpc returns a minimal valid VPC the tests mutate per case.
func vpc() *DigitalOceanVpc {
	return &DigitalOceanVpc{
		ApiVersion: "digital-ocean.planton.dev/v1alpha1",
		Kind:       "DigitalOceanVpc",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-vpc",
		},
		Spec: &DigitalOceanVpcSpec{
			Region: digitalocean.DigitalOceanRegion_nyc3,
		},
	}
}

var _ = ginkgo.Describe("DigitalOceanVpcSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal VPC with only a region (DigitalOcean assigns the range)", func() {
			gomega.Expect(protovalidate.Validate(vpc())).To(gomega.BeNil())
		})

		ginkgo.It("accepts every prefix length DigitalOcean supports (/16 through /24)", func() {
			for _, cidr := range []string{
				"10.10.0.0/16", "10.10.0.0/17", "10.10.0.0/18", "10.10.0.0/19",
				"10.10.0.0/20", "10.10.0.0/21", "10.10.0.0/22", "10.10.0.0/23",
				"10.10.0.0/24",
			} {
				input := vpc()
				input.Spec.IpRangeCidr = cidr
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), cidr)
			}
		})

		ginkgo.It("accepts a description up to 255 characters", func() {
			input := vpc()
			desc := make([]byte, 255)
			for i := range desc {
				desc[i] = 'd'
			}
			input.Spec.Description = string(desc)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("region validation", func() {

		ginkgo.It("rejects a missing region", func() {
			input := vpc()
			input.Spec.Region = digitalocean.DigitalOceanRegion_digital_ocean_region_unspecified
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("ip_range_cidr validation", func() {

		ginkgo.It("rejects a /15 prefix (larger than DigitalOcean allows)", func() {
			input := vpc()
			input.Spec.IpRangeCidr = "10.10.0.0/15"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a /25 prefix (smaller than DigitalOcean allows)", func() {
			input := vpc()
			input.Spec.IpRangeCidr = "10.10.0.0/25"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an address without a prefix length", func() {
			input := vpc()
			input.Spec.IpRangeCidr = "10.10.0.0"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a non-CIDR string", func() {
			input := vpc()
			input.Spec.IpRangeCidr = "not-a-cidr"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("description validation", func() {

		ginkgo.It("rejects a description longer than 255 characters", func() {
			input := vpc()
			desc := make([]byte, 256)
			for i := range desc {
				desc[i] = 'd'
			}
			input.Spec.Description = string(desc)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
