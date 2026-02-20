package alicloudkubernetesnodepoolv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAlicloudKubernetesNodePoolSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudKubernetesNodePoolSpec Validation Tests")
}

func strRef(s string) *fkv1.StringValueOrRef {
	return &fkv1.StringValueOrRef{
		LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: s},
	}
}

func minimalValidSpec() *AlicloudKubernetesNodePoolSpec {
	return &AlicloudKubernetesNodePoolSpec{
		Region:        "cn-hangzhou",
		ClusterId:     strRef("c-abc123"),
		Name:          "default-pool",
		VswitchIds:    []*fkv1.StringValueOrRef{strRef("vsw-aaa111"), strRef("vsw-bbb222")},
		InstanceTypes: []string{"ecs.g7.xlarge"},
	}
}

func minimalValidInput() *AlicloudKubernetesNodePool {
	return &AlicloudKubernetesNodePool{
		ApiVersion: "alicloud.openmcf.org/v1",
		Kind:       "AlicloudKubernetesNodePool",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-pool"},
		Spec:       minimalValidSpec(),
	}
}

var _ = ginkgo.Describe("AlicloudKubernetesNodePoolSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			err := protovalidate.Validate(minimalValidInput())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with a single vswitch", func() {
			input := minimalValidInput()
			input.Spec.VswitchIds = []*fkv1.StringValueOrRef{strRef("vsw-aaa111")}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with desired_size set", func() {
			input := minimalValidInput()
			input.Spec.DesiredSize = proto.Int32(3)
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with multiple instance types", func() {
			input := minimalValidInput()
			input.Spec.InstanceTypes = []string{"ecs.g7.xlarge", "ecs.g7.2xlarge", "ecs.c7.xlarge"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with system disk configuration", func() {
			input := minimalValidInput()
			input.Spec.SystemDisk = &AlicloudKubernetesNodePoolSystemDisk{
				Category:         proto.String("cloud_essd"),
				Size:             proto.Int32(200),
				PerformanceLevel: "PL1",
				Encrypted:        proto.Bool(true),
				KmsKeyId:         "kms-key-123",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with data disks", func() {
			input := minimalValidInput()
			input.Spec.DataDisks = []*AlicloudKubernetesNodePoolDataDisk{
				{
					Category:         proto.String("cloud_essd"),
					Size:             200,
					Name:             "data-disk-1",
					PerformanceLevel: "PL1",
				},
				{
					Size: 100,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with auto-scaling configuration", func() {
			input := minimalValidInput()
			input.Spec.DesiredSize = proto.Int32(2)
			input.Spec.ScalingConfig = &AlicloudKubernetesNodePoolScalingConfig{
				Enable:  proto.Bool(true),
				MinSize: 1,
				MaxSize: 10,
				Type:    proto.String("cpu"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with management configuration", func() {
			input := minimalValidInput()
			input.Spec.Management = &AlicloudKubernetesNodePoolManagement{
				Enable:         proto.Bool(true),
				AutoRepair:     proto.Bool(true),
				AutoUpgrade:    proto.Bool(true),
				MaxUnavailable: proto.Int32(1),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with spot strategy", func() {
			input := minimalValidInput()
			input.Spec.InstanceTypes = []string{"ecs.g7.xlarge", "ecs.c7.xlarge"}
			input.Spec.SpotStrategy = proto.String("SpotAsPriceGo")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with spot price limits", func() {
			input := minimalValidInput()
			input.Spec.SpotStrategy = proto.String("SpotWithPriceLimit")
			input.Spec.SpotPriceLimits = []*AlicloudKubernetesNodePoolSpotPriceLimit{
				{InstanceType: "ecs.g7.xlarge", PriceLimit: "0.98"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with labels and taints", func() {
			input := minimalValidInput()
			input.Spec.Labels = map[string]string{
				"workload-type": "compute",
				"team":          "platform",
			}
			input.Spec.Taints = []*AlicloudKubernetesNodePoolTaint{
				{Key: "dedicated", Value: "gpu", Effect: "NoSchedule"},
				{Key: "special", Effect: "PreferNoSchedule"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with PrePaid billing", func() {
			input := minimalValidInput()
			input.Spec.InstanceChargeType = proto.String("PrePaid")
			input.Spec.Period = proto.Int32(12)
			input.Spec.AutoRenew = proto.Bool(true)
			input.Spec.AutoRenewPeriod = proto.Int32(6)
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with authentication via key_name", func() {
			input := minimalValidInput()
			input.Spec.KeyName = "my-ssh-keypair"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with kubernetes configuration", func() {
			input := minimalValidInput()
			input.Spec.CpuPolicy = proto.String("static")
			input.Spec.RuntimeName = "containerd"
			input.Spec.RuntimeVersion = "1.6.28"
			input.Spec.Unschedulable = proto.Bool(false)
			input.Spec.InstallCloudMonitor = proto.Bool(true)
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with multi-AZ policy", func() {
			input := minimalValidInput()
			input.Spec.VswitchIds = []*fkv1.StringValueOrRef{
				strRef("vsw-az-a"), strRef("vsw-az-b"), strRef("vsw-az-c"),
			}
			input.Spec.MultiAzPolicy = proto.String("BALANCE")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with full production configuration", func() {
			input := minimalValidInput()
			input.Spec.Name = "production-pool"
			input.Spec.InstanceTypes = []string{"ecs.g7.2xlarge", "ecs.g7.xlarge"}
			input.Spec.DesiredSize = proto.Int32(6)
			input.Spec.ImageType = proto.String("AliyunLinux3")
			input.Spec.SystemDisk = &AlicloudKubernetesNodePoolSystemDisk{
				Category:         proto.String("cloud_essd"),
				Size:             proto.Int32(200),
				PerformanceLevel: "PL1",
				Encrypted:        proto.Bool(true),
			}
			input.Spec.DataDisks = []*AlicloudKubernetesNodePoolDataDisk{
				{Category: proto.String("cloud_essd"), Size: 500, PerformanceLevel: "PL1"},
			}
			input.Spec.KeyName = "prod-keypair"
			input.Spec.Labels = map[string]string{"env": "production", "tier": "compute"}
			input.Spec.Taints = []*AlicloudKubernetesNodePoolTaint{
				{Key: "dedicated", Value: "production", Effect: "NoSchedule"},
			}
			input.Spec.CpuPolicy = proto.String("none")
			input.Spec.RuntimeName = "containerd"
			input.Spec.InstallCloudMonitor = proto.Bool(true)
			input.Spec.ScalingConfig = &AlicloudKubernetesNodePoolScalingConfig{
				Enable:  proto.Bool(true),
				MinSize: 3,
				MaxSize: 20,
			}
			input.Spec.MultiAzPolicy = proto.String("BALANCE")
			input.Spec.Management = &AlicloudKubernetesNodePoolManagement{
				Enable:         proto.Bool(true),
				AutoRepair:     proto.Bool(true),
				AutoUpgrade:    proto.Bool(true),
				MaxUnavailable: proto.Int32(2),
			}
			input.Spec.Tags = map[string]string{"team": "platform", "cost-center": "infra-001"}
			input.Spec.ResourceGroupId = "rg-prod"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when api_version is wrong", func() {
			input := minimalValidInput()
			input.ApiVersion = "wrong/v1"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := minimalValidInput()
			input.Kind = "WrongKind"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := minimalValidInput()
			input.Metadata = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AlicloudKubernetesNodePool{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudKubernetesNodePool",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when region is empty", func() {
			input := minimalValidInput()
			input.Spec.Region = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cluster_id is missing", func() {
			input := minimalValidInput()
			input.Spec.ClusterId = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when name is empty", func() {
			input := minimalValidInput()
			input.Spec.Name = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when name exceeds 63 characters", func() {
			input := minimalValidInput()
			input.Spec.Name = "this-node-pool-name-is-way-too-long-and-exceeds-the-sixty-three-character-limit"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vswitch_ids is empty", func() {
			input := minimalValidInput()
			input.Spec.VswitchIds = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vswitch_ids exceeds 5", func() {
			input := minimalValidInput()
			input.Spec.VswitchIds = []*fkv1.StringValueOrRef{
				strRef("vsw-1"), strRef("vsw-2"), strRef("vsw-3"),
				strRef("vsw-4"), strRef("vsw-5"), strRef("vsw-6"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when instance_types is empty", func() {
			input := minimalValidInput()
			input.Spec.InstanceTypes = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when image_type is invalid", func() {
			input := minimalValidInput()
			input.Spec.ImageType = proto.String("InvalidOS")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spot_strategy is invalid", func() {
			input := minimalValidInput()
			input.Spec.SpotStrategy = proto.String("InvalidStrategy")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when multi_az_policy is invalid", func() {
			input := minimalValidInput()
			input.Spec.MultiAzPolicy = proto.String("ROUND_ROBIN")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cpu_policy is invalid", func() {
			input := minimalValidInput()
			input.Spec.CpuPolicy = proto.String("dedicated")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when instance_charge_type is invalid", func() {
			input := minimalValidInput()
			input.Spec.InstanceChargeType = proto.String("Hourly")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when period is invalid", func() {
			input := minimalValidInput()
			input.Spec.Period = proto.Int32(5)
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when desired_size exceeds 1000", func() {
			input := minimalValidInput()
			input.Spec.DesiredSize = proto.Int32(1001)
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when data_disk size is below 40", func() {
			input := minimalValidInput()
			input.Spec.DataDisks = []*AlicloudKubernetesNodePoolDataDisk{
				{Size: 10},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when taint key is empty", func() {
			input := minimalValidInput()
			input.Spec.Taints = []*AlicloudKubernetesNodePoolTaint{
				{Key: "", Effect: "NoSchedule"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when taint effect is invalid", func() {
			input := minimalValidInput()
			input.Spec.Taints = []*AlicloudKubernetesNodePoolTaint{
				{Key: "dedicated", Effect: "Drain"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when internet_charge_type is invalid", func() {
			input := minimalValidInput()
			input.Spec.InternetChargeType = proto.String("PayByMonth")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when system_disk category is invalid", func() {
			input := minimalValidInput()
			input.Spec.SystemDisk = &AlicloudKubernetesNodePoolSystemDisk{
				Category: proto.String("premium_ssd"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when scaling_config type is invalid", func() {
			input := minimalValidInput()
			input.Spec.ScalingConfig = &AlicloudKubernetesNodePoolScalingConfig{
				Enable:  proto.Bool(true),
				MinSize: 1,
				MaxSize: 10,
				Type:    proto.String("memory"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
