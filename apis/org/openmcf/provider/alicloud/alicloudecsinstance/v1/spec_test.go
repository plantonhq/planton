package alicloudecsinstancev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAlicloudEcsInstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudEcsInstanceSpec Validation Tests")
}

var _ = ginkgo.Describe("AlicloudEcsInstanceSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-hangzhou",
					InstanceType: "ecs.g7.large",
					ImageId:      "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "prod-web-server",
					Org:  "acme-corp",
					Env:  "production",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-shanghai",
					InstanceType: "ecs.g7.2xlarge",
					ImageId:      "aliyun_3_x64_20G_alibase_20230727.vhd",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-xyz789"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-web123"}},
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-mgmt456"}},
					},
					InstanceName: "prod-web-01",
					HostName:     "web-01",
					Description:  "Production web server",
					SystemDisk: &AlicloudEcsSystemDisk{
						Category:         proto.String("cloud_essd"),
						Size:             proto.Int32(100),
						PerformanceLevel: proto.String("PL1"),
						Encrypted:        proto.Bool(true),
						KmsKeyId:         "kms-abc123",
					},
					DataDisks: []*AlicloudEcsDataDisk{
						{
							Size:               200,
							Category:           proto.String("cloud_essd"),
							Name:               "data-vol-01",
							PerformanceLevel:   proto.String("PL1"),
							Encrypted:          proto.Bool(true),
							KmsKeyId:           "kms-abc123",
							DeleteWithInstance: proto.Bool(true),
						},
						{
							Size:               500,
							Category:           proto.String("cloud_ssd"),
							Name:               "log-vol-01",
							DeleteWithInstance: proto.Bool(false),
							Description:        "Log storage volume",
						},
					},
					KeyName:                "my-keypair",
					InternetMaxBandwidthOut: proto.Int32(10),
					InternetChargeType:     proto.String("PayByTraffic"),
					InstanceChargeType:     proto.String("PostPaid"),
					UserData:               "IyEvYmluL2Jhc2gKZWNobyBoZWxsbw==",
					RoleName:               "EcsInstanceRole",
					DeletionProtection:     proto.Bool(true),
					SecurityEnhancementStrategy: proto.String("Active"),
					ResourceGroupId:        "rg-abc123",
					Tags:                   map[string]string{"team": "platform", "env": "prod"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with spot instance configuration", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "spot-worker",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-hangzhou",
					InstanceType: "ecs.c7.xlarge",
					ImageId:      "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
					SpotStrategy:   proto.String("SpotWithPriceLimit"),
					SpotPriceLimit: proto.Float64(0.5),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with PrePaid billing configuration", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "prepaid-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-hangzhou",
					InstanceType: "ecs.g7.large",
					ImageId:      "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
					InstanceChargeType: proto.String("PrePaid"),
					Period:             proto.Int32(12),
					PeriodUnit:         proto.String("Month"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with password authentication", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "password-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-hangzhou",
					InstanceType: "ecs.g7.large",
					ImageId:      "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
					Password: "SecureP@ss123!",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail with wrong api_version", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "wrong/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-hangzhou",
					InstanceType: "ecs.g7.large",
					ImageId:      "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail with wrong kind", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-hangzhou",
					InstanceType: "ecs.g7.large",
					ImageId:      "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail without region", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					InstanceType: "ecs.g7.large",
					ImageId:      "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail without vswitch_id", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-hangzhou",
					InstanceType: "ecs.g7.large",
					ImageId:      "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail without security_group_ids", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-hangzhou",
					InstanceType: "ecs.g7.large",
					ImageId:      "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail with invalid instance_type prefix", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-hangzhou",
					InstanceType: "g7.large",
					ImageId:      "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail without image_id", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-hangzhou",
					InstanceType: "ecs.g7.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail with invalid instance_charge_type", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:             "cn-hangzhou",
					InstanceType:       "ecs.g7.large",
					ImageId:            "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					InstanceChargeType: proto.String("InvalidType"),
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail with invalid spot_strategy", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-hangzhou",
					InstanceType: "ecs.g7.large",
					ImageId:      "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					SpotStrategy: proto.String("BadStrategy"),
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail with invalid system_disk category", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-hangzhou",
					InstanceType: "ecs.g7.large",
					ImageId:      "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
					SystemDisk: &AlicloudEcsSystemDisk{
						Category: proto.String("invalid_disk_type"),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail with password too short", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-hangzhou",
					InstanceType: "ecs.g7.large",
					ImageId:      "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					Password:     "short",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail with invalid period value", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:             "cn-hangzhou",
					InstanceType:       "ecs.g7.large",
					ImageId:            "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					InstanceChargeType: proto.String("PrePaid"),
					Period:             proto.Int32(13),
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail with data disk size below minimum", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:       "cn-hangzhou",
					InstanceType: "ecs.g7.large",
					ImageId:      "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
					DataDisks: []*AlicloudEcsDataDisk{
						{Size: 5},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail with internet_max_bandwidth_out exceeding 100", func() {
			input := &AlicloudEcsInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudEcsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-ecs",
				},
				Spec: &AlicloudEcsInstanceSpec{
					Region:                  "cn-hangzhou",
					InstanceType:            "ecs.g7.large",
					ImageId:                 "ubuntu_22_04_x64_20G_alibase_20230515.vhd",
					InternetMaxBandwidthOut: proto.Int32(200),
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					SecurityGroupIds: []*fkv1.StringValueOrRef{
						{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "sg-abc123"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
