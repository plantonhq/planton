package alicloudsecuritygroupv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAliCloudSecurityGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudSecurityGroupSpec Validation Tests")
}

var _ = ginkgo.Describe("AliCloudSecurityGroupSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields and no rules", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-sg",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-abc123"},
					},
					SecurityGroupName: "my-sg",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional SG-level fields populated", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-sg",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-shanghai",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-xyz789"},
					},
					SecurityGroupName: "prod-security-group",
					Description:       "Production security group for web tier",
					InnerAccessPolicy: proto.String("Drop"),
					ResourceGroupId:   "rg-abc123",
					Tags:              map[string]string{"team": "platform", "tier": "web"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with inner_access_policy Accept", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "accept-sg",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "us-west-1",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "accept-sg",
					InnerAccessPolicy: proto.String("Accept"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with a complete ingress rule", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "web-sg",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "web-sg",
					Rules: []*AliCloudSecurityGroupRule{
						{
							Type:        "ingress",
							IpProtocol:  "tcp",
							PortRange:   proto.String("443/443"),
							CidrIp:      "0.0.0.0/0",
							Priority:    proto.Int32(1),
							Policy:      proto.String("accept"),
							Description: "Allow HTTPS from anywhere",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with mixed ingress and egress rules", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "mixed-sg",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "mixed-sg",
					Rules: []*AliCloudSecurityGroupRule{
						{
							Type:       "ingress",
							IpProtocol: "tcp",
							PortRange:  proto.String("80/80"),
							CidrIp:     "10.0.0.0/8",
						},
						{
							Type:       "ingress",
							IpProtocol: "tcp",
							PortRange:  proto.String("443/443"),
							CidrIp:     "0.0.0.0/0",
						},
						{
							Type:       "egress",
							IpProtocol: "all",
							PortRange:  proto.String("-1/-1"),
							CidrIp:     "0.0.0.0/0",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with a SG-to-SG rule using source_security_group_id", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "sg-to-sg",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "sg-to-sg",
					Rules: []*AliCloudSecurityGroupRule{
						{
							Type:                  "ingress",
							IpProtocol:            "tcp",
							PortRange:             proto.String("3306/3306"),
							SourceSecurityGroupId: "sg-web-tier",
							Description:           "Allow MySQL from web tier SG",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with a drop-policy rule", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "deny-sg",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "deny-sg",
					Rules: []*AliCloudSecurityGroupRule{
						{
							Type:       "ingress",
							IpProtocol: "all",
							CidrIp:     "192.168.1.0/24",
							Policy:     proto.String("drop"),
							Priority:   proto.Int32(50),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with priority at boundary values", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "priority-sg",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "priority-sg",
					Rules: []*AliCloudSecurityGroupRule{
						{
							Type:       "ingress",
							IpProtocol: "tcp",
							PortRange:  proto.String("22/22"),
							CidrIp:     "10.0.0.0/8",
							Priority:   proto.Int32(1),
						},
						{
							Type:       "ingress",
							IpProtocol: "tcp",
							PortRange:  proto.String("22/22"),
							CidrIp:     "0.0.0.0/0",
							Priority:   proto.Int32(100),
							Policy:     proto.String("drop"),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with vpc_id as a reference", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "ref-sg",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_ValueFrom{
							ValueFrom: &fkv1.ValueFromRef{
								Name: "my-vpc",
							},
						},
					},
					SecurityGroupName: "ref-sg",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "my-sg",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_id is missing", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region:            "cn-hangzhou",
					SecurityGroupName: "my-sg",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when security_group_name is missing", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when security_group_name is too short", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "x",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when security_group_name exceeds max length", func() {
			longName := ""
			for i := 0; i < 129; i++ {
				longName += "a"
			}
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: longName,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "my-sg",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "my-sg",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "my-sg",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when inner_access_policy has invalid value", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "my-sg",
					InnerAccessPolicy: proto.String("Invalid"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when rule type is invalid", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "my-sg",
					Rules: []*AliCloudSecurityGroupRule{
						{
							Type:       "invalid",
							IpProtocol: "tcp",
							CidrIp:     "0.0.0.0/0",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when rule ip_protocol is invalid", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "my-sg",
					Rules: []*AliCloudSecurityGroupRule{
						{
							Type:       "ingress",
							IpProtocol: "ftp",
							CidrIp:     "0.0.0.0/0",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when rule ip_protocol is missing", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "my-sg",
					Rules: []*AliCloudSecurityGroupRule{
						{
							Type:   "ingress",
							CidrIp: "0.0.0.0/0",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when rule type is missing", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "my-sg",
					Rules: []*AliCloudSecurityGroupRule{
						{
							IpProtocol: "tcp",
							CidrIp:     "0.0.0.0/0",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when rule policy has invalid value", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "my-sg",
					Rules: []*AliCloudSecurityGroupRule{
						{
							Type:       "ingress",
							IpProtocol: "tcp",
							CidrIp:     "0.0.0.0/0",
							Policy:     proto.String("allow"),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when rule priority is below minimum", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "my-sg",
					Rules: []*AliCloudSecurityGroupRule{
						{
							Type:       "ingress",
							IpProtocol: "tcp",
							CidrIp:     "0.0.0.0/0",
							Priority:   proto.Int32(0),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when rule priority exceeds maximum", func() {
			input := &AliCloudSecurityGroup{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudSecurityGroup",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudSecurityGroupSpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					SecurityGroupName: "my-sg",
					Rules: []*AliCloudSecurityGroupRule{
						{
							Type:       "ingress",
							IpProtocol: "tcp",
							CidrIp:     "0.0.0.0/0",
							Priority:   proto.Int32(101),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
