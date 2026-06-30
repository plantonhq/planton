package civovpcv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	civo "github.com/plantonhq/planton/apis/dev/planton/provider/civo"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestCivoVpcSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CivoVpcSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CivoVpcSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("civo_vpc with minimal valid fields", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &CivoVpc{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-network",
					},
					Spec: &CivoVpcSpec{
						NetworkName: "test-network",
						Region:      civo.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with auto-allocated CIDR (empty ip_range_cidr)", func() {
				input := &CivoVpc{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "dev-network",
					},
					Spec: &CivoVpcSpec{
						NetworkName: "dev-network",
						Region:      civo.CivoRegion_nyc1,
						IpRangeCidr: "", // Empty = auto-allocate
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with explicit CIDR", func() {
				input := &CivoVpc{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-network",
					},
					Spec: &CivoVpcSpec{
						NetworkName: "prod-network",
						Region:      civo.CivoRegion_fra1,
						IpRangeCidr: "10.10.1.0/24",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error when set as default network", func() {
				input := &CivoVpc{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "default-network",
					},
					Spec: &CivoVpcSpec{
						NetworkName:        "default-network",
						Region:             civo.CivoRegion_lon1,
						IpRangeCidr:        "10.0.0.0/24",
						IsDefaultForRegion: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with description", func() {
				input := &CivoVpc{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "staging-network",
					},
					Spec: &CivoVpcSpec{
						NetworkName: "staging-network",
						Region:      civo.CivoRegion_nyc1,
						IpRangeCidr: "10.20.2.0/24",
						Description: "Staging environment network",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with maximum length description", func() {
				input := &CivoVpc{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-network",
					},
					Spec: &CivoVpcSpec{
						NetworkName: "prod-network",
						Region:      civo.CivoRegion_lon1,
						IpRangeCidr: "10.10.1.0/24",
						Description: "This is exactly one hundred characters long for testing maximum length constraint validation rule", // exactly 100 chars
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all fields", func() {
				input := &CivoVpc{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "complete-network",
					},
					Spec: &CivoVpcSpec{
						NetworkName:        "complete-network",
						Region:             civo.CivoRegion_fra1,
						IpRangeCidr:        "10.30.1.0/24",
						IsDefaultForRegion: false,
						Description:        "Production network for Frankfurt region",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with different regions", func() {
				regions := []civo.CivoRegion{
					civo.CivoRegion_lon1,
					civo.CivoRegion_nyc1,
					civo.CivoRegion_fra1,
					civo.CivoRegion_phx1,
					civo.CivoRegion_mum1,
				}
				for _, region := range regions {
					input := &CivoVpc{
						ApiVersion: "civo.planton.dev/v1",
						Kind:       "CivoVpc",
						Metadata: &shared.CloudResourceMetadata{
							Name: "test-network",
						},
						Spec: &CivoVpcSpec{
							NetworkName: "test-network",
							Region:      region,
							IpRangeCidr: "10.0.1.0/24",
						},
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error with various valid CIDR blocks", func() {
				validCidrs := []string{
					"10.0.0.0/24",
					"172.16.1.0/24",
					"192.168.100.0/24",
					"10.10.10.0/24",
				}
				for _, cidr := range validCidrs {
					input := &CivoVpc{
						ApiVersion: "civo.planton.dev/v1",
						Kind:       "CivoVpc",
						Metadata: &shared.CloudResourceMetadata{
							Name: "test-network",
						},
						Spec: &CivoVpcSpec{
							NetworkName: "test-network",
							Region:      civo.CivoRegion_lon1,
							IpRangeCidr: cidr,
						},
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("network_name validation", func() {

			ginkgo.It("should return a validation error when network_name is empty", func() {
				input := &CivoVpc{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-network",
					},
					Spec: &CivoVpcSpec{
						NetworkName: "", // Empty
						Region:      civo.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("region validation", func() {

			ginkgo.It("should return a validation error when region is unspecified", func() {
				input := &CivoVpc{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-network",
					},
					Spec: &CivoVpcSpec{
						NetworkName: "test-network",
						Region:      civo.CivoRegion_civo_region_unspecified,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("description validation", func() {

			ginkgo.It("should return a validation error when description exceeds max length", func() {
				input := &CivoVpc{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-network",
					},
					Spec: &CivoVpcSpec{
						NetworkName: "test-network",
						Region:      civo.CivoRegion_lon1,
						Description: "This description is definitely way too long because it exceeds the maximum allowed length of one hundred characters which should cause a validation error to be returned",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
