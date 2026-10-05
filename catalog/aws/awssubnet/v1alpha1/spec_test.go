package awssubnetv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsSubnetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsSubnetSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func newValueFromRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

func ptr(s string) *string { return &s }

func iptr(i int32) *int32 { return &i }

func minimalValidSubnet() *AwsSubnet {
	return &AwsSubnet{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsSubnet",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-subnet",
		},
		Spec: &AwsSubnetSpec{
			Region:           "us-west-2",
			VpcId:            newStringValueOrRef("vpc-0abc123"),
			AvailabilityZone: "us-west-2a",
			CidrBlock:        "10.0.1.0/24",
		},
	}
}

var _ = ginkgo.Describe("AwsSubnetSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_subnet", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(minimalValidSubnet())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional fields set", func() {
				input := &AwsSubnet{
					ApiVersion: "aws.planton.dev/v1alpha1",
					Kind:       "AwsSubnet",
					Metadata: &shared.CatalogObjectMetadata{
						Name: "full-subnet",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &AwsSubnetSpec{
						Region:                                  "us-west-2",
						VpcId:                                   newStringValueOrRef("vpc-0abc123"),
						AvailabilityZone:                        "us-west-2a",
						CidrBlock:                               "10.0.1.0/24",
						MapPublicIpOnLaunch:                     true,
						AssignIpv6AddressOnCreation:             true,
						Ipv6CidrBlock:                           "2600:1f18:abcd:1200::/64",
						EnableDns64:                             true,
						EnableResourceNameDnsARecordOnLaunch:    true,
						EnableResourceNameDnsAaaaRecordOnLaunch: true,
						PrivateDnsHostnameTypeOnLaunch:          ptr("resource-name"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with vpc_id via value_from ref", func() {
				input := minimalValidSubnet()
				input.Spec.VpcId = newValueFromRef("my-vpc")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with route_table_id only", func() {
				input := minimalValidSubnet()
				input.Spec.RouteTableId = newStringValueOrRef("rtb-0abc123")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with an inline IPv4 default route", func() {
				input := minimalValidSubnet()
				input.Spec.Routes = []*AwsSubnetSpec_AwsSubnetRoute{
					{
						DestinationCidrBlock: "0.0.0.0/0",
						TargetType:           AwsSubnetSpec_AwsSubnetRoute_internet_gateway,
						TargetId:             newStringValueOrRef("igw-0abc123"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple inline routes", func() {
				input := minimalValidSubnet()
				input.Spec.Routes = []*AwsSubnetSpec_AwsSubnetRoute{
					{
						DestinationCidrBlock: "0.0.0.0/0",
						TargetType:           AwsSubnetSpec_AwsSubnetRoute_nat_gateway,
						TargetId:             newStringValueOrRef("nat-0abc123"),
					},
					{
						DestinationIpv6CidrBlock: "::/0",
						TargetType:               AwsSubnetSpec_AwsSubnetRoute_egress_only_internet_gateway,
						TargetId:                 newStringValueOrRef("eigw-0abc123"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a prefix-list route", func() {
				input := minimalValidSubnet()
				input.Spec.Routes = []*AwsSubnetSpec_AwsSubnetRoute{
					{
						DestinationPrefixListId: "pl-0123abcd",
						TargetType:              AwsSubnetSpec_AwsSubnetRoute_transit_gateway,
						TargetId:                newStringValueOrRef("tgw-0abc123"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with neither route_table_id nor routes", func() {
				err := protovalidate.Validate(minimalValidSubnet())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an empty private_dns_hostname_type_on_launch", func() {
				input := minimalValidSubnet()
				input.Spec.PrivateDnsHostnameTypeOnLaunch = ptr("")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with availability_zone_id instead of the zone name", func() {
				input := minimalValidSubnet()
				input.Spec.AvailabilityZone = ""
				input.Spec.AvailabilityZoneId = "usw2-az1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with an IPv4 IPAM allocation instead of cidr_block", func() {
				input := minimalValidSubnet()
				input.Spec.CidrBlock = ""
				input.Spec.Ipv4IpamPoolId = newStringValueOrRef("ipam-pool-0abc123")
				input.Spec.Ipv4NetmaskLength = iptr(24)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with an IPv6 IPAM allocation", func() {
				input := minimalValidSubnet()
				input.Spec.Ipv6IpamPoolId = newStringValueOrRef("ipam-pool-0def456")
				input.Spec.Ipv6NetmaskLength = iptr(64)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an IPv6-only subnet", func() {
				input := minimalValidSubnet()
				input.Spec.CidrBlock = ""
				input.Spec.Ipv6Native = true
				input.Spec.Ipv6CidrBlock = "2600:1f18:abcd:1200::/64"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for the newer route target kinds", func() {
				input := minimalValidSubnet()
				input.Spec.Routes = []*AwsSubnetSpec_AwsSubnetRoute{
					{
						DestinationCidrBlock: "10.100.0.0/16",
						TargetType:           AwsSubnetSpec_AwsSubnetRoute_core_network,
						TargetId:             newStringValueOrRef("arn:aws:networkmanager::123456789012:core-network/core-network-0abc123"),
					},
					{
						DestinationCidrBlock: "10.200.0.0/16",
						TargetType:           AwsSubnetSpec_AwsSubnetRoute_local_gateway,
						TargetId:             newStringValueOrRef("lgw-0abc123"),
					},
					{
						DestinationCidrBlock: "10.201.0.0/16",
						TargetType:           AwsSubnetSpec_AwsSubnetRoute_carrier_gateway,
						TargetId:             newStringValueOrRef("cagw-0abc123"),
					},
					{
						DestinationCidrBlock: "10.202.0.0/16",
						TargetType:           AwsSubnetSpec_AwsSubnetRoute_odb_network,
						TargetId:             newStringValueOrRef("arn:aws:odb:us-west-2:123456789012:odb-network/odb-0abc123"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for propagating_vgws with inline routes", func() {
				input := minimalValidSubnet()
				input.Spec.Routes = []*AwsSubnetSpec_AwsSubnetRoute{
					{
						DestinationCidrBlock: "0.0.0.0/0",
						TargetType:           AwsSubnetSpec_AwsSubnetRoute_internet_gateway,
						TargetId:             newStringValueOrRef("igw-0abc123"),
					},
				}
				input.Spec.PropagatingVgws = []string{"vgw-0abc123"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for propagating_vgws without inline routes", func() {
				input := minimalValidSubnet()
				input.Spec.PropagatingVgws = []string{"vgw-0abc123"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_subnet", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidSubnet()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidSubnet()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidSubnet()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AwsSubnet{
					ApiVersion: "aws.planton.dev/v1alpha1",
					Kind:       "AwsSubnet",
					Metadata:   &shared.CatalogObjectMetadata{Name: "test-subnet"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty", func() {
				input := minimalValidSubnet()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vpc_id is missing", func() {
				input := minimalValidSubnet()
				input.Spec.VpcId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when availability_zone is empty", func() {
				input := minimalValidSubnet()
				input.Spec.AvailabilityZone = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cidr_block is empty", func() {
				input := minimalValidSubnet()
				input.Spec.CidrBlock = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cidr_block is not a CIDR", func() {
				input := minimalValidSubnet()
				input.Spec.CidrBlock = "10.0.1.0"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when route_table_id and routes are both set", func() {
				input := minimalValidSubnet()
				input.Spec.RouteTableId = newStringValueOrRef("rtb-0abc123")
				input.Spec.Routes = []*AwsSubnetSpec_AwsSubnetRoute{
					{
						DestinationCidrBlock: "0.0.0.0/0",
						TargetType:           AwsSubnetSpec_AwsSubnetRoute_internet_gateway,
						TargetId:             newStringValueOrRef("igw-0abc123"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a route has no destination", func() {
				input := minimalValidSubnet()
				input.Spec.Routes = []*AwsSubnetSpec_AwsSubnetRoute{
					{
						TargetType: AwsSubnetSpec_AwsSubnetRoute_internet_gateway,
						TargetId:   newStringValueOrRef("igw-0abc123"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a route has more than one destination", func() {
				input := minimalValidSubnet()
				input.Spec.Routes = []*AwsSubnetSpec_AwsSubnetRoute{
					{
						DestinationCidrBlock:     "0.0.0.0/0",
						DestinationIpv6CidrBlock: "::/0",
						TargetType:               AwsSubnetSpec_AwsSubnetRoute_internet_gateway,
						TargetId:                 newStringValueOrRef("igw-0abc123"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a route target_type is unspecified", func() {
				input := minimalValidSubnet()
				input.Spec.Routes = []*AwsSubnetSpec_AwsSubnetRoute{
					{
						DestinationCidrBlock: "0.0.0.0/0",
						TargetId:             newStringValueOrRef("igw-0abc123"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a route is missing target_id", func() {
				input := minimalValidSubnet()
				input.Spec.Routes = []*AwsSubnetSpec_AwsSubnetRoute{
					{
						DestinationCidrBlock: "0.0.0.0/0",
						TargetType:           AwsSubnetSpec_AwsSubnetRoute_internet_gateway,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid private_dns_hostname_type_on_launch", func() {
				input := minimalValidSubnet()
				input.Spec.PrivateDnsHostnameTypeOnLaunch = ptr("invalid-type")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both availability_zone and availability_zone_id are set", func() {
				input := minimalValidSubnet()
				input.Spec.AvailabilityZoneId = "usw2-az1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cidr_block and ipv4_ipam_pool_id are both set", func() {
				input := minimalValidSubnet()
				input.Spec.Ipv4IpamPoolId = newStringValueOrRef("ipam-pool-0abc123")
				input.Spec.Ipv4NetmaskLength = iptr(24)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ipv4_netmask_length is set without its pool", func() {
				input := minimalValidSubnet()
				input.Spec.CidrBlock = ""
				input.Spec.Ipv4NetmaskLength = iptr(24)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range ipv4_netmask_length", func() {
				input := minimalValidSubnet()
				input.Spec.CidrBlock = ""
				input.Spec.Ipv4IpamPoolId = newStringValueOrRef("ipam-pool-0abc123")
				input.Spec.Ipv4NetmaskLength = iptr(30)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an ipv6_netmask_length AWS does not accept", func() {
				input := minimalValidSubnet()
				input.Spec.Ipv6IpamPoolId = newStringValueOrRef("ipam-pool-0def456")
				input.Spec.Ipv6NetmaskLength = iptr(63)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ipv6_cidr_block and ipv6_ipam_pool_id are both set", func() {
				input := minimalValidSubnet()
				input.Spec.Ipv6CidrBlock = "2600:1f18:abcd:1200::/64"
				input.Spec.Ipv6IpamPoolId = newStringValueOrRef("ipam-pool-0def456")
				input.Spec.Ipv6NetmaskLength = iptr(64)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for ipv6_native with IPv4 addressing", func() {
				input := minimalValidSubnet()
				input.Spec.Ipv6Native = true
				input.Spec.Ipv6CidrBlock = "2600:1f18:abcd:1200::/64"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for ipv6_native without any IPv6 addressing", func() {
				input := minimalValidSubnet()
				input.Spec.CidrBlock = ""
				input.Spec.Ipv6Native = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for propagating_vgws with an external route_table_id", func() {
				input := minimalValidSubnet()
				input.Spec.RouteTableId = newStringValueOrRef("rtb-0abc123")
				input.Spec.PropagatingVgws = []string{"vgw-0abc123"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
