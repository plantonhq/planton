package awseksnodegroupv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsEksNodeGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsEksNodeGroupSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimalValidNodeGroup is the common case: a small inline-configured
// On-Demand pool in one private subnet pair.
func minimalValidNodeGroup() *AwsEksNodeGroup {
	return &AwsEksNodeGroup{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsEksNodeGroup",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "workers",
		},
		Spec: &AwsEksNodeGroupSpec{
			Region:      "us-west-2",
			ClusterName: literalRef("platform"),
			NodeRoleArn: literalRef("arn:aws:iam::123456789012:role/EksNodeRole"),
			SubnetIds: []*foreignkeyv1.StringValueOrRef{
				literalRef("subnet-0123456789abcdef0"),
				literalRef("subnet-0123456789abcdef1"),
			},
			InstanceTypes: []string{"m6i.large"},
			Scaling: &AwsEksNodeGroupScalingConfig{
				MinSize:     1,
				MaxSize:     3,
				DesiredSize: 1,
			},
		},
	}
}

var _ = ginkgo.Describe("AwsEksNodeGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_eks_node_group", func() {

			ginkgo.It("should not return a validation error for a minimal node group", func() {
				err := protovalidate.Validate(minimalValidNodeGroup())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a launch-template node group", func() {
				input := minimalValidNodeGroup()
				input.Spec.InstanceTypes = nil
				input.Spec.LaunchTemplate = &AwsEksNodeGroupLaunchTemplate{
					LaunchTemplateId: literalRef("lt-0123456789abcdef0"),
					Version:          "$Default",
				}
				input.Spec.AmiType = "AL2023_x86_64_STANDARD"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a dedicated tainted Spot pool with rollout and repair config", func() {
				input := minimalValidNodeGroup()
				input.Spec.InstanceTypes = []string{"m6i.large", "m5.large", "m6a.large"}
				input.Spec.CapacityType = AwsEksNodeGroupCapacityType_spot
				input.Spec.Labels = map[string]string{"pool": "batch"}
				input.Spec.Taints = []*AwsEksNodeGroupTaint{
					{Key: "dedicated", Value: "batch", Effect: "NO_SCHEDULE"},
				}
				input.Spec.UpdateConfig = &AwsEksNodeGroupUpdateConfig{
					MaxUnavailablePercentage: 25,
					UpdateStrategy:           "MINIMAL",
				}
				input.Spec.NodeRepairConfig = &AwsEksNodeGroupNodeRepairConfig{
					Enabled:                        true,
					MaxParallelNodesRepairedCount:  1,
					MaxUnhealthyNodeThresholdCount: 3,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a scale-to-zero pool", func() {
				input := minimalValidNodeGroup()
				input.Spec.Scaling = &AwsEksNodeGroupScalingConfig{
					MinSize:     0,
					MaxSize:     5,
					DesiredSize: 0,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a pinned version and release", func() {
				input := minimalValidNodeGroup()
				input.Spec.Version = "1.31"
				input.Spec.ReleaseVersion = "1.31.3-20241109"
				input.Spec.ForceUpdateVersion = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept scoped remote access", func() {
				input := minimalValidNodeGroup()
				input.Spec.RemoteAccess = &AwsEksNodeGroupRemoteAccess{
					Ec2SshKey: "ops-keypair",
					SourceSecurityGroupIds: []*foreignkeyv1.StringValueOrRef{
						literalRef("sg-0123456789abcdef0"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a warm pool with an explicit zero ceiling", func() {
				zero := int32(0)
				input := minimalValidNodeGroup()
				input.Spec.WarmPoolConfig = &AwsEksNodeGroupWarmPoolConfig{
					PoolState:                "STOPPED",
					MinSize:                  2,
					MaxGroupPreparedCapacity: &zero,
					ReuseOnScaleIn:           true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a warm pool with the ceiling left to AWS", func() {
				input := minimalValidNodeGroup()
				input.Spec.WarmPoolConfig = &AwsEksNodeGroupWarmPoolConfig{
					PoolState: "HIBERNATED",
					MinSize:   1,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_eks_node_group", func() {

			ginkgo.It("should reject a missing cluster reference", func() {
				input := minimalValidNodeGroup()
				input.Spec.ClusterName = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing scaling config", func() {
				input := minimalValidNodeGroup()
				input.Spec.Scaling = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject instance_types combined with a launch template", func() {
				input := minimalValidNodeGroup()
				input.Spec.LaunchTemplate = &AwsEksNodeGroupLaunchTemplate{
					LaunchTemplateId: literalRef("lt-0123456789abcdef0"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject disk_size_gb combined with a launch template", func() {
				input := minimalValidNodeGroup()
				input.Spec.InstanceTypes = nil
				input.Spec.DiskSizeGb = 100
				input.Spec.LaunchTemplate = &AwsEksNodeGroupLaunchTemplate{
					LaunchTemplateId: literalRef("lt-0123456789abcdef0"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject remote_access combined with a launch template", func() {
				input := minimalValidNodeGroup()
				input.Spec.InstanceTypes = nil
				input.Spec.RemoteAccess = &AwsEksNodeGroupRemoteAccess{Ec2SshKey: "ops-keypair"}
				input.Spec.LaunchTemplate = &AwsEksNodeGroupLaunchTemplate{
					LaunchTemplateId: literalRef("lt-0123456789abcdef0"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a launch template without a template id", func() {
				input := minimalValidNodeGroup()
				input.Spec.InstanceTypes = nil
				input.Spec.LaunchTemplate = &AwsEksNodeGroupLaunchTemplate{Version: "$Latest"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown ami type", func() {
				input := minimalValidNodeGroup()
				input.Spec.AmiType = "UBUNTU_x86_64"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a malformed Kubernetes version", func() {
				input := minimalValidNodeGroup()
				input.Spec.Version = "1.9"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject desired_size outside the min/max bounds", func() {
				input := minimalValidNodeGroup()
				input.Spec.Scaling = &AwsEksNodeGroupScalingConfig{
					MinSize:     1,
					MaxSize:     3,
					DesiredSize: 5,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject max_size below min_size", func() {
				input := minimalValidNodeGroup()
				input.Spec.Scaling = &AwsEksNodeGroupScalingConfig{
					MinSize:     3,
					MaxSize:     1,
					DesiredSize: 3,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an invalid taint effect", func() {
				input := minimalValidNodeGroup()
				input.Spec.Taints = []*AwsEksNodeGroupTaint{
					{Key: "dedicated", Effect: "NoSchedule"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an update config with both unavailability forms", func() {
				input := minimalValidNodeGroup()
				input.Spec.UpdateConfig = &AwsEksNodeGroupUpdateConfig{
					MaxUnavailable:           1,
					MaxUnavailablePercentage: 25,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an update config with neither unavailability form", func() {
				input := minimalValidNodeGroup()
				input.Spec.UpdateConfig = &AwsEksNodeGroupUpdateConfig{UpdateStrategy: "DEFAULT"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an invalid update strategy", func() {
				input := minimalValidNodeGroup()
				input.Spec.UpdateConfig = &AwsEksNodeGroupUpdateConfig{
					MaxUnavailable: 1,
					UpdateStrategy: "SURGE",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject repair config with both parallelism forms", func() {
				input := minimalValidNodeGroup()
				input.Spec.NodeRepairConfig = &AwsEksNodeGroupNodeRepairConfig{
					Enabled:                            true,
					MaxParallelNodesRepairedCount:      1,
					MaxParallelNodesRepairedPercentage: 10,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a repair override with an invalid action", func() {
				input := minimalValidNodeGroup()
				input.Spec.NodeRepairConfig = &AwsEksNodeGroupNodeRepairConfig{
					Enabled: true,
					Overrides: []*AwsEksNodeGroupNodeRepairOverride{
						{
							MinRepairWaitTimeMins:   10,
							NodeMonitoringCondition: "AcceleratedHardwareReady",
							NodeUnhealthyReason:     "NvidiaXID13Error",
							RepairAction:            "Terminate",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown warm pool state", func() {
				input := minimalValidNodeGroup()
				input.Spec.WarmPoolConfig = &AwsEksNodeGroupWarmPoolConfig{
					PoolState: "Stopped",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a negative warm pool ceiling", func() {
				negative := int32(-1)
				input := minimalValidNodeGroup()
				input.Spec.WarmPoolConfig = &AwsEksNodeGroupWarmPoolConfig{
					MaxGroupPreparedCapacity: &negative,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
