package awsvpcv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAwsVpcSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsVpcSpec Validation Tests")
}

func boolPtr(b bool) *bool { return &b }

// minimalValidVpc is the common case: a region and a primary IPv4 CIDR.
func minimalValidVpc() *AwsVpc {
	return &AwsVpc{
		ApiVersion: "aws.openmcf.org/v1",
		Kind:       "AwsVpc",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-vpc",
		},
		Spec: &AwsVpcSpec{
			Region:    "us-west-2",
			CidrBlock: "10.0.0.0/16",
		},
	}
}

var _ = ginkgo.Describe("AwsVpcSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_vpc", func() {

			ginkgo.It("should not return a validation error for a minimal VPC", func() {
				err := protovalidate.Validate(minimalValidVpc())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a fully-specified IPv4+IPv6 VPC", func() {
				input := minimalValidVpc()
				input.Spec.SecondaryIpv4CidrBlocks = []string{"10.1.0.0/16", "100.64.0.0/16"}
				input.Spec.InstanceTenancy = "dedicated"
				input.Spec.EnableDnsSupport = boolPtr(true)
				input.Spec.EnableDnsHostnames = true
				input.Spec.EnableNetworkAddressUsageMetrics = true
				input.Spec.AssignGeneratedIpv6CidrBlock = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an IPv4 IPAM-allocated VPC", func() {
				input := minimalValidVpc()
				input.Spec.CidrBlock = ""
				input.Spec.Ipv4IpamPoolId = "ipam-pool-0abc123"
				input.Spec.Ipv4NetmaskLength = 16
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an IPv6 IPAM-allocated VPC", func() {
				input := minimalValidVpc()
				input.Spec.Ipv6IpamPoolId = "ipam-pool-ipv6-0abc123"
				input.Spec.Ipv6NetmaskLength = 56
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for amazon-provided IPv6 with a border group", func() {
				input := minimalValidVpc()
				input.Spec.AssignGeneratedIpv6CidrBlock = true
				input.Spec.Ipv6CidrBlockNetworkBorderGroup = "us-west-2-lax-1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error when DNS support is explicitly disabled", func() {
				input := minimalValidVpc()
				input.Spec.EnableDnsSupport = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full metadata set", func() {
				input := minimalValidVpc()
				input.Metadata = &shared.CloudResourceMetadata{
					Name:   "full-vpc",
					Org:    "acme-corp",
					Env:    "production",
					Labels: map[string]string{"team": "platform"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_vpc", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidVpc()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidVpc()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidVpc()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AwsVpc{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsVpc",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-vpc"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty", func() {
				input := minimalValidVpc()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when no primary IPv4 source is set", func() {
				input := minimalValidVpc()
				input.Spec.CidrBlock = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cidr_block and ipv4_netmask_length are both set", func() {
				input := minimalValidVpc()
				input.Spec.Ipv4IpamPoolId = "ipam-pool-0abc123"
				input.Spec.Ipv4NetmaskLength = 16
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cidr_block is malformed", func() {
				input := minimalValidVpc()
				input.Spec.CidrBlock = "not-a-cidr"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ipv4_netmask_length is out of range", func() {
				input := minimalValidVpc()
				input.Spec.CidrBlock = ""
				input.Spec.Ipv4IpamPoolId = "ipam-pool-0abc123"
				input.Spec.Ipv4NetmaskLength = 40
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ipv4_netmask_length has no IPAM pool", func() {
				input := minimalValidVpc()
				input.Spec.CidrBlock = ""
				input.Spec.Ipv4NetmaskLength = 16
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when instance_tenancy is unknown", func() {
				input := minimalValidVpc()
				input.Spec.InstanceTenancy = "weird"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when amazon-provided IPv6 is combined with IPAM", func() {
				input := minimalValidVpc()
				input.Spec.AssignGeneratedIpv6CidrBlock = true
				input.Spec.Ipv6IpamPoolId = "ipam-pool-ipv6-0abc123"
				input.Spec.Ipv6NetmaskLength = 56
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ipv6_cidr_block has no IPAM pool", func() {
				input := minimalValidVpc()
				input.Spec.Ipv6CidrBlock = "2600:1f18:abcd:1200::/56"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ipv6_cidr_block and ipv6_netmask_length are both set", func() {
				input := minimalValidVpc()
				input.Spec.Ipv6IpamPoolId = "ipam-pool-ipv6-0abc123"
				input.Spec.Ipv6CidrBlock = "2600:1f18:abcd:1200::/56"
				input.Spec.Ipv6NetmaskLength = 56
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ipv6_netmask_length is invalid", func() {
				input := minimalValidVpc()
				input.Spec.Ipv6IpamPoolId = "ipam-pool-ipv6-0abc123"
				input.Spec.Ipv6NetmaskLength = 50
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the IPv6 border group has no amazon-provided block", func() {
				input := minimalValidVpc()
				input.Spec.Ipv6CidrBlockNetworkBorderGroup = "us-west-2-lax-1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
