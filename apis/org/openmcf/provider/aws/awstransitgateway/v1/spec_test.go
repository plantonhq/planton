package awstransitgatewayv1

import (
	"testing"

	"buf.build/go/protovalidate"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestAwsTransitGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsTransitGatewaySpec Validation Suite")
}

func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

func minimalAttachment(name, vpcId, subnetId string) *AwsTransitGatewayVpcAttachment {
	return &AwsTransitGatewayVpcAttachment{
		Name:      name,
		VpcId:     strRef(vpcId),
		SubnetIds: []*foreignkeyv1.StringValueOrRef{strRef(subnetId)},
	}
}

var _ = ginkgo.Describe("AwsTransitGatewaySpec validations", func() {
	var spec *AwsTransitGatewaySpec

	ginkgo.BeforeEach(func() {
		spec = &AwsTransitGatewaySpec{
			VpcAttachments: []*AwsTransitGatewayVpcAttachment{
				minimalAttachment("primary-vpc", "vpc-abc123", "subnet-abc123"),
			},
		}
	})

	// =========================================================================
	// Happy path — Spec level
	// =========================================================================

	ginkgo.It("accepts a minimal valid spec (one VPC attachment)", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with description", func() {
		spec.Description = "Production multi-VPC hub"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts custom 16-bit ASN", func() {
		spec.AmazonSideAsn = 65000
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts custom 32-bit ASN", func() {
		spec.AmazonSideAsn = 4200000001
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts default ASN (64512)", func() {
		spec.AmazonSideAsn = 64512
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts multiple VPC attachments", func() {
		spec.VpcAttachments = []*AwsTransitGatewayVpcAttachment{
			minimalAttachment("vpc-a", "vpc-111", "subnet-111"),
			minimalAttachment("vpc-b", "vpc-222", "subnet-222"),
			minimalAttachment("vpc-c", "vpc-333", "subnet-333"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts attachment with multiple subnets (multi-AZ)", func() {
		spec.VpcAttachments = []*AwsTransitGatewayVpcAttachment{
			{
				Name:  "multi-az-vpc",
				VpcId: strRef("vpc-abc123"),
				SubnetIds: []*foreignkeyv1.StringValueOrRef{
					strRef("subnet-az1"),
					strRef("subnet-az2"),
					strRef("subnet-az3"),
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts transit gateway CIDR blocks", func() {
		spec.TransitGatewayCidrBlocks = []string{
			"10.99.0.0/24",
			"10.99.1.0/24",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts maximum 5 CIDR blocks", func() {
		spec.TransitGatewayCidrBlocks = []string{
			"10.99.0.0/24", "10.99.1.0/24", "10.99.2.0/24",
			"10.99.3.0/24", "10.99.4.0/24",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts all feature toggles enabled", func() {
		spec.DefaultRouteTableAssociation = true
		spec.DefaultRouteTablePropagation = true
		spec.DnsSupport = true
		spec.VpnEcmpSupport = true
		spec.AutoAcceptSharedAttachments = true
		spec.SecurityGroupReferencingSupport = true
		spec.MulticastSupport = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts attachment with appliance mode for firewall VPC", func() {
		spec.VpcAttachments = []*AwsTransitGatewayVpcAttachment{
			{
				Name:                 "inspection-vpc",
				VpcId:                strRef("vpc-fw123"),
				SubnetIds:            []*foreignkeyv1.StringValueOrRef{strRef("subnet-fw1")},
				ApplianceModeSupport: true,
				DnsSupport:           true,
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts attachment with IPv6 support", func() {
		spec.VpcAttachments[0].Ipv6Support = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts attachment with explicit route table overrides", func() {
		spec.VpcAttachments[0].DefaultRouteTableAssociation = false
		spec.VpcAttachments[0].DefaultRouteTablePropagation = false
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path — Production-ready configuration
	// =========================================================================

	ginkgo.It("accepts a production-ready multi-VPC configuration", func() {
		spec = &AwsTransitGatewaySpec{
			Description:                 "Production hub connecting application and shared-services VPCs",
			AmazonSideAsn:               64512,
			DefaultRouteTableAssociation: true,
			DefaultRouteTablePropagation: true,
			DnsSupport:                  true,
			VpnEcmpSupport:              true,
			VpcAttachments: []*AwsTransitGatewayVpcAttachment{
				{
					Name:  "app-vpc",
					VpcId: strRef("vpc-app001"),
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						strRef("subnet-app-az1"),
						strRef("subnet-app-az2"),
					},
					DnsSupport:                  true,
					DefaultRouteTableAssociation: true,
					DefaultRouteTablePropagation: true,
				},
				{
					Name:  "shared-services-vpc",
					VpcId: strRef("vpc-shared001"),
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						strRef("subnet-shared-az1"),
						strRef("subnet-shared-az2"),
					},
					DnsSupport:                  true,
					DefaultRouteTableAssociation: true,
					DefaultRouteTablePropagation: true,
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Failure — Spec level
	// =========================================================================

	ginkgo.It("rejects empty vpc_attachments", func() {
		spec.VpcAttachments = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("vpc_attachments"))
	})

	ginkgo.It("rejects more than 5 CIDR blocks", func() {
		spec.TransitGatewayCidrBlocks = []string{
			"10.99.0.0/24", "10.99.1.0/24", "10.99.2.0/24",
			"10.99.3.0/24", "10.99.4.0/24", "10.99.5.0/24",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("transit_gateway_cidr_blocks"))
	})

	ginkgo.It("rejects ASN below valid 16-bit range", func() {
		spec.AmazonSideAsn = 100
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("amazon_side_asn"))
	})

	ginkgo.It("rejects ASN between 16-bit and 32-bit ranges", func() {
		spec.AmazonSideAsn = 100000
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("amazon_side_asn"))
	})

	// =========================================================================
	// Failure — VPC attachment validations
	// =========================================================================

	ginkgo.It("rejects attachment without name", func() {
		spec.VpcAttachments = []*AwsTransitGatewayVpcAttachment{
			{
				VpcId:     strRef("vpc-abc123"),
				SubnetIds: []*foreignkeyv1.StringValueOrRef{strRef("subnet-abc123")},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("name"))
	})

	ginkgo.It("rejects attachment with invalid name format (uppercase)", func() {
		spec.VpcAttachments[0].Name = "Primary-VPC"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("name"))
	})

	ginkgo.It("rejects attachment with invalid name format (starts with number)", func() {
		spec.VpcAttachments[0].Name = "1st-vpc"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("name"))
	})

	ginkgo.It("rejects attachment without vpc_id", func() {
		spec.VpcAttachments = []*AwsTransitGatewayVpcAttachment{
			{
				Name:      "test-vpc",
				SubnetIds: []*foreignkeyv1.StringValueOrRef{strRef("subnet-abc123")},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("vpc_id"))
	})

	ginkgo.It("rejects attachment without subnet_ids", func() {
		spec.VpcAttachments = []*AwsTransitGatewayVpcAttachment{
			{
				Name:  "test-vpc",
				VpcId: strRef("vpc-abc123"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("subnet_ids"))
	})

	ginkgo.It("rejects attachment with empty subnet_ids", func() {
		spec.VpcAttachments = []*AwsTransitGatewayVpcAttachment{
			{
				Name:      "test-vpc",
				VpcId:     strRef("vpc-abc123"),
				SubnetIds: []*foreignkeyv1.StringValueOrRef{},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("subnet_ids"))
	})
})
