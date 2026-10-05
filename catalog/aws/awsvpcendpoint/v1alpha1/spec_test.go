package awsvpcendpointv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAwsVpcEndpointSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsVpcEndpointSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimalGatewayEndpoint is the most common shape: a free S3 gateway
// endpoint injected into one route table.
func minimalGatewayEndpoint() *AwsVpcEndpoint {
	return &AwsVpcEndpoint{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsVpcEndpoint",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "platform-s3-gateway",
		},
		Spec: &AwsVpcEndpointSpec{
			Region:      "us-west-2",
			VpcId:       literalRef("vpc-0a1b2c3d4e5f67890"),
			ServiceName: "com.amazonaws.us-west-2.s3",
			RouteTableIds: []*foreignkeyv1.StringValueOrRef{
				literalRef("rtb-0a1b2c3d4e5f67890"),
			},
		},
	}
}

// minimalInterfaceEndpoint is the PrivateLink shape: an ENI-based
// endpoint on two subnets with private DNS.
func minimalInterfaceEndpoint() *AwsVpcEndpoint {
	return &AwsVpcEndpoint{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsVpcEndpoint",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "platform-sts-endpoint",
		},
		Spec: &AwsVpcEndpointSpec{
			Region:       "us-west-2",
			VpcId:        literalRef("vpc-0a1b2c3d4e5f67890"),
			EndpointType: "Interface",
			ServiceName:  "com.amazonaws.us-west-2.sts",
			SubnetIds: []*foreignkeyv1.StringValueOrRef{
				literalRef("subnet-0a1b2c3d4e5f67890"),
				literalRef("subnet-0f9e8d7c6b5a43210"),
			},
			PrivateDnsEnabled: proto.Bool(true),
		},
	}
}

var _ = ginkgo.Describe("AwsVpcEndpointSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_vpc_endpoint", func() {

			ginkgo.It("should not return a validation error for a minimal gateway endpoint", func() {
				err := protovalidate.Validate(minimalGatewayEndpoint())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit Gateway type with a policy document", func() {
				input := minimalGatewayEndpoint()
				input.Spec.EndpointType = "Gateway"
				policy, err := structpb.NewStruct(map[string]any{
					"Version": "2012-10-17",
					"Statement": []any{
						map[string]any{
							"Effect":    "Allow",
							"Principal": "*",
							"Action":    "s3:GetObject",
							"Resource":  "*",
						},
					},
				})
				gomega.Expect(err).To(gomega.BeNil())
				input.Spec.Policy = policy
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a gateway endpoint with no route tables (attach later)", func() {
				input := minimalGatewayEndpoint()
				input.Spec.RouteTableIds = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a minimal interface endpoint", func() {
				err := protovalidate.Validate(minimalInterfaceEndpoint())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an interface endpoint with security groups, IP type, and DNS options", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					literalRef("sg-0a1b2c3d4e5f67890"),
				}
				input.Spec.IpAddressType = "ipv4"
				input.Spec.DnsOptions = &AwsVpcEndpointDnsOptions{
					DnsRecordIpType: "ipv4",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the S3 dual-stack inbound-resolver pattern", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.ServiceName = "com.amazonaws.us-west-2.s3"
				input.Spec.DnsOptions = &AwsVpcEndpointDnsOptions{
					PrivateDnsOnlyForInboundResolverEndpoint: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a cross-region interface endpoint", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.PrivateDnsEnabled = nil
				input.Spec.ServiceRegion = "us-east-1"
				input.Spec.ServiceName = "com.amazonaws.us-east-1.sts"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit-false private DNS on an interface endpoint (the disable path)", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.PrivateDnsEnabled = proto.Bool(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept pinned ENI addresses via subnet configurations", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.SubnetConfigurations = []*AwsVpcEndpointSubnetConfiguration{
					{
						SubnetId: literalRef("subnet-0a1b2c3d4e5f67890"),
						Ipv4:     "10.0.101.10",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a VPC Lattice Resource endpoint", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.EndpointType = "Resource"
				input.Spec.ServiceName = ""
				input.Spec.ResourceConfigurationArn = "arn:aws:vpc-lattice:us-west-2:123456789012:resourceconfiguration/rcfg-0a1b2c3d4e5f67890"
				input.Spec.PrivateDnsEnabled = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a VPC Lattice ServiceNetwork endpoint with a DNS preference", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.EndpointType = "ServiceNetwork"
				input.Spec.ServiceName = ""
				input.Spec.ServiceNetworkArn = "arn:aws:vpc-lattice:us-west-2:123456789012:servicenetwork/sn-0a1b2c3d4e5f67890"
				input.Spec.PrivateDnsEnabled = nil
				input.Spec.DnsOptions = &AwsVpcEndpointDnsOptions{
					PrivateDnsPreference:       "SPECIFIED_DOMAINS_ONLY",
					PrivateDnsSpecifiedDomains: []string{"internal.example.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept auto_accept for same-account PrivateLink services", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.ServiceName = "com.amazonaws.vpce.us-west-2.vpce-svc-0a1b2c3d4e5f67890"
				input.Spec.PrivateDnsEnabled = nil
				input.Spec.AutoAccept = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_vpc_endpoint", func() {

			ginkgo.It("should return an error when region is empty", func() {
				input := minimalGatewayEndpoint()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when vpc_id is missing", func() {
				input := minimalGatewayEndpoint()
				input.Spec.VpcId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when no service target is set", func() {
				input := minimalGatewayEndpoint()
				input.Spec.ServiceName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when two service targets are set", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.EndpointType = "Resource"
				input.Spec.ResourceConfigurationArn = "arn:aws:vpc-lattice:us-west-2:123456789012:resourceconfiguration/rcfg-0a1b2c3d4e5f67890"
				// service_name is still set from the fixture -- two targets.
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an unknown endpoint type", func() {
				input := minimalGatewayEndpoint()
				input.Spec.EndpointType = "PrivateLink"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a resource configuration ARN on a non-Resource type", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.ServiceName = ""
				input.Spec.ResourceConfigurationArn = "arn:aws:vpc-lattice:us-west-2:123456789012:resourceconfiguration/rcfg-0a1b2c3d4e5f67890"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a service network ARN on a non-ServiceNetwork type", func() {
				input := minimalGatewayEndpoint()
				input.Spec.ServiceName = ""
				input.Spec.ServiceNetworkArn = "arn:aws:vpc-lattice:us-west-2:123456789012:servicenetwork/sn-0a1b2c3d4e5f67890"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject route tables on an interface endpoint", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.RouteTableIds = []*foreignkeyv1.StringValueOrRef{
					literalRef("rtb-0a1b2c3d4e5f67890"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject subnets on a gateway endpoint", func() {
				input := minimalGatewayEndpoint()
				input.Spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
					literalRef("subnet-0a1b2c3d4e5f67890"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject private DNS on a gateway endpoint", func() {
				input := minimalGatewayEndpoint()
				input.Spec.PrivateDnsEnabled = proto.Bool(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an explicit-false private DNS on a gateway endpoint (presence matters)", func() {
				input := minimalGatewayEndpoint()
				input.Spec.PrivateDnsEnabled = proto.Bool(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject inbound-resolver-only DNS without private DNS enabled", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.PrivateDnsEnabled = nil
				input.Spec.DnsOptions = &AwsVpcEndpointDnsOptions{
					PrivateDnsOnlyForInboundResolverEndpoint: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject security groups on a GatewayLoadBalancer endpoint", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.EndpointType = "GatewayLoadBalancer"
				input.Spec.PrivateDnsEnabled = nil
				input.Spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					literalRef("sg-0a1b2c3d4e5f67890"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a cross-region target on a non-Interface endpoint", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.EndpointType = "GatewayLoadBalancer"
				input.Spec.PrivateDnsEnabled = nil
				input.Spec.ServiceRegion = "us-east-1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an unknown ip_address_type", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.IpAddressType = "ipv5"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an unknown dns_record_ip_type", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.DnsOptions = &AwsVpcEndpointDnsOptions{
					DnsRecordIpType: "any",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an unknown private_dns_preference", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.DnsOptions = &AwsVpcEndpointDnsOptions{
					PrivateDnsPreference: "SOME_DOMAINS",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject specified domains without a specified-domains preference", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.DnsOptions = &AwsVpcEndpointDnsOptions{
					PrivateDnsSpecifiedDomains: []string{"internal.example.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a specified-domains preference without domains", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.DnsOptions = &AwsVpcEndpointDnsOptions{
					PrivateDnsPreference: "SPECIFIED_DOMAINS_ONLY",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject more than 10 specified domains", func() {
				domains := make([]string, 11)
				for i := range domains {
					domains[i] = "internal.example.com"
				}
				input := minimalInterfaceEndpoint()
				input.Spec.DnsOptions = &AwsVpcEndpointDnsOptions{
					PrivateDnsPreference:       "SPECIFIED_DOMAINS_ONLY",
					PrivateDnsSpecifiedDomains: domains,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a subnet configuration without a subnet", func() {
				input := minimalInterfaceEndpoint()
				input.Spec.SubnetConfigurations = []*AwsVpcEndpointSubnetConfiguration{
					{Ipv4: "10.0.101.10"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
