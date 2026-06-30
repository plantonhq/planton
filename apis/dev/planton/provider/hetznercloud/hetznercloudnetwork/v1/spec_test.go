package hetznercloudnetworkv1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHetznerCloudNetworkSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HetznerCloudNetworkSpec Validation Suite")
}

var _ = Describe("HetznerCloudNetworkSpec validations", func() {

	Context("with valid specs", func() {
		It("should accept a minimal network with one cloud subnet", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/16",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{
					{
						Type:        HetznerCloudNetworkSpec_Subnet_cloud,
						NetworkZone: "eu-central",
						IpRange:     "10.0.1.0/24",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept multiple subnets in different zones", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/8",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{
					{
						Type:        HetznerCloudNetworkSpec_Subnet_cloud,
						NetworkZone: "eu-central",
						IpRange:     "10.0.1.0/24",
					},
					{
						Type:        HetznerCloudNetworkSpec_Subnet_cloud,
						NetworkZone: "us-east",
						IpRange:     "10.0.2.0/24",
					},
					{
						Type:        HetznerCloudNetworkSpec_Subnet_server,
						NetworkZone: "eu-central",
						IpRange:     "10.0.3.0/24",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept subnets with routes", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/16",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{
					{
						Type:        HetznerCloudNetworkSpec_Subnet_cloud,
						NetworkZone: "eu-central",
						IpRange:     "10.0.1.0/24",
					},
				},
				Routes: []*HetznerCloudNetworkSpec_Route{
					{
						Destination: "172.16.0.0/12",
						Gateway:     "10.0.0.1",
					},
					{
						Destination: "192.168.0.0/16",
						Gateway:     "10.0.0.2",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a vswitch subnet with vswitch_id", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/16",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{
					{
						Type:        HetznerCloudNetworkSpec_Subnet_vswitch,
						NetworkZone: "eu-central",
						IpRange:     "10.0.10.0/24",
						VswitchId:   12345,
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a network with delete_protection enabled", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/16",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{
					{
						Type:        HetznerCloudNetworkSpec_Subnet_cloud,
						NetworkZone: "eu-central",
						IpRange:     "10.0.1.0/24",
					},
				},
				DeleteProtection: true,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a network with expose_routes_to_vswitch enabled", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/16",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{
					{
						Type:        HetznerCloudNetworkSpec_Subnet_vswitch,
						NetworkZone: "eu-central",
						IpRange:     "10.0.10.0/24",
						VswitchId:   12345,
					},
				},
				ExposeRoutesToVswitch: true,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a cloud subnet without vswitch_id", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/16",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{
					{
						Type:        HetznerCloudNetworkSpec_Subnet_cloud,
						NetworkZone: "eu-central",
						IpRange:     "10.0.1.0/24",
						VswitchId:   0,
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("with invalid specs", func() {
		It("should reject a spec with empty ip_range", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{
					{
						Type:        HetznerCloudNetworkSpec_Subnet_cloud,
						NetworkZone: "eu-central",
						IpRange:     "10.0.1.0/24",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a spec with no subnets", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/16",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a spec with nil subnets", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/16",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a subnet with subnet_type_unspecified", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/16",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{
					{
						Type:        HetznerCloudNetworkSpec_Subnet_subnet_type_unspecified,
						NetworkZone: "eu-central",
						IpRange:     "10.0.1.0/24",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a subnet with empty network_zone", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/16",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{
					{
						Type:        HetznerCloudNetworkSpec_Subnet_cloud,
						NetworkZone: "",
						IpRange:     "10.0.1.0/24",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a subnet with empty ip_range", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/16",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{
					{
						Type:        HetznerCloudNetworkSpec_Subnet_cloud,
						NetworkZone: "eu-central",
						IpRange:     "",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a vswitch subnet without vswitch_id", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/16",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{
					{
						Type:        HetznerCloudNetworkSpec_Subnet_vswitch,
						NetworkZone: "eu-central",
						IpRange:     "10.0.10.0/24",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a route with empty destination", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/16",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{
					{
						Type:        HetznerCloudNetworkSpec_Subnet_cloud,
						NetworkZone: "eu-central",
						IpRange:     "10.0.1.0/24",
					},
				},
				Routes: []*HetznerCloudNetworkSpec_Route{
					{
						Destination: "",
						Gateway:     "10.0.0.1",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a route with empty gateway", func() {
			spec := &HetznerCloudNetworkSpec{
				IpRange: "10.0.0.0/16",
				Subnets: []*HetznerCloudNetworkSpec_Subnet{
					{
						Type:        HetznerCloudNetworkSpec_Subnet_cloud,
						NetworkZone: "eu-central",
						IpRange:     "10.0.1.0/24",
					},
				},
				Routes: []*HetznerCloudNetworkSpec_Route{
					{
						Destination: "172.16.0.0/12",
						Gateway:     "",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})
	})
})
