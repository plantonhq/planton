package awssubnetv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
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

func minimalValidSubnet() *AwsSubnet {
	return &AwsSubnet{
		ApiVersion: "aws.planton.dev/v1",
		Kind:       "AwsSubnet",
		Metadata: &shared.CloudResourceMetadata{
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
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsSubnet",
					Metadata: &shared.CloudResourceMetadata{
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
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsSubnet",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-subnet"},
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
		})
	})
})
