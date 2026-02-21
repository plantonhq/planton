package alicloudnatgatewayv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAliCloudNatGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudNatGatewaySpec Validation Tests")
}

var _ = ginkgo.Describe("AliCloudNatGatewaySpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields and no SNAT entries", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-nat",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-abc123"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					NatGatewayName: "my-nat",
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-abc123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-nat",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-shanghai",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-xyz789"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-xyz789"},
					},
					NatGatewayName:     "prod-nat-gateway",
					Description:        "Production NAT Gateway for outbound internet access",
					NatType:            proto.String("Enhanced"),
					PaymentType:        proto.String("PayAsYouGo"),
					InternetChargeType: proto.String("PayByLcu"),
					DeletionProtection: proto.Bool(true),
					Tags:               map[string]string{"team": "platform", "env": "prod"},
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-xyz789"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with SNAT entries using source_vswitch_id", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "snat-nat",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-nat"},
					},
					NatGatewayName: "snat-nat",
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-123"},
					},
					SnatEntries: []*AliCloudSnatEntry{
						{
							SourceVswitchId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-app-a"},
							},
							SnatEntryName: "app-zone-a",
						},
						{
							SourceVswitchId: &fkv1.StringValueOrRef{
								LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-app-b"},
							},
							SnatEntryName: "app-zone-b",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with SNAT entries using source_cidr", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "cidr-nat",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "us-west-1",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-456"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-456"},
					},
					NatGatewayName: "cidr-nat",
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-456"},
					},
					SnatEntries: []*AliCloudSnatEntry{
						{
							SourceCidr:    "10.0.1.0/24",
							SnatEntryName: "subnet-a-only",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with nat_type Normal", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "normal-nat",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-normal"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-normal"},
					},
					NatGatewayName: "normal-nat",
					NatType:        proto.String("Normal"),
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-normal"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with PayBySpec and specification", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "spec-nat",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-spec"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-spec"},
					},
					NatGatewayName:     "spec-nat",
					InternetChargeType: proto.String("PayBySpec"),
					Specification:      proto.String("Large"),
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-spec"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with Subscription payment type", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "sub-nat",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-sub"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-sub"},
					},
					NatGatewayName: "sub-nat",
					PaymentType:    proto.String("Subscription"),
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-sub"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					NatGatewayName: "test",
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					NatGatewayName: "test",
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					NatGatewayName: "test",
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when region is empty", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					NatGatewayName: "test",
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_id is missing", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					NatGatewayName: "test",
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vswitch_id is missing", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					NatGatewayName: "test",
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when eip_id is missing", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					NatGatewayName: "test",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when nat_gateway_name is too short", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					NatGatewayName: "x",
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when nat_type is invalid", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					NatGatewayName: "test-nat",
					NatType:        proto.String("SuperFast"),
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when payment_type is invalid", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					NatGatewayName: "test-nat",
					PaymentType:    proto.String("Free"),
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when internet_charge_type is invalid", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					NatGatewayName:     "test-nat",
					InternetChargeType: proto.String("PayByMagic"),
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when specification is invalid", func() {
			input := &AliCloudNatGateway{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudNatGateway",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudNatGatewaySpec{
					Region: "cn-hangzhou",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-123"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					NatGatewayName:     "test-nat",
					InternetChargeType: proto.String("PayBySpec"),
					Specification:      proto.String("Jumbo"),
					EipId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "eip-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
