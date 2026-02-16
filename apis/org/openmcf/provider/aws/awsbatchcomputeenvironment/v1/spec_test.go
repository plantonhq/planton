package awsbatchcomputeenvironmentv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsBatchComputeEnvironmentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsBatchComputeEnvironmentSpec Validation Suite")
}

func int32Ptr(i int32) *int32 {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func svRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

func minimalFargateSpec() *AwsBatchComputeEnvironmentSpec {
	return &AwsBatchComputeEnvironmentSpec{
		ComputeResources: &AwsBatchComputeResources{
			Type:             "FARGATE",
			MaxVcpus:         256,
			SubnetIds:        []*foreignkeyv1.StringValueOrRef{svRef("subnet-aaa"), svRef("subnet-bbb")},
			SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{svRef("sg-111")},
		},
		JobQueues: []*AwsBatchJobQueue{
			{Name: "default", Priority: 1},
		},
	}
}

func minimalEc2Spec() *AwsBatchComputeEnvironmentSpec {
	return &AwsBatchComputeEnvironmentSpec{
		ComputeResources: &AwsBatchComputeResources{
			Type:          "EC2",
			MaxVcpus:      256,
			SubnetIds:     []*foreignkeyv1.StringValueOrRef{svRef("subnet-aaa")},
			InstanceTypes: []string{"optimal"},
			InstanceRole:  svRef("arn:aws:iam::123456789012:instance-profile/ecsInstanceRole"),
		},
		JobQueues: []*AwsBatchJobQueue{
			{Name: "default", Priority: 1},
		},
	}
}

var _ = ginkgo.Describe("AwsBatchComputeEnvironmentSpec validations", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.Context("with minimal Fargate configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(minimalFargateSpec())
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with minimal EC2 configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(minimalEc2Spec())
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with SPOT configuration including required fields", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsBatchComputeEnvironmentSpec{
					ComputeResources: &AwsBatchComputeResources{
						Type:               "SPOT",
						MaxVcpus:           512,
						SubnetIds:          []*foreignkeyv1.StringValueOrRef{svRef("subnet-aaa")},
						InstanceTypes:      []string{"m5.xlarge", "c5.xlarge"},
						InstanceRole:       svRef("arn:aws:iam::123456789012:instance-profile/ecsInstanceRole"),
						SpotIamFleetRole:   svRef("arn:aws:iam::123456789012:role/aws-ec2-spot-fleet-role"),
						BidPercentage:      int32Ptr(60),
						AllocationStrategy: "SPOT_CAPACITY_OPTIMIZED",
					},
					JobQueues: []*AwsBatchJobQueue{
						{Name: "high-priority", Priority: 10},
						{Name: "low-priority", Priority: 1},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with FARGATE_SPOT configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsBatchComputeEnvironmentSpec{
					ComputeResources: &AwsBatchComputeResources{
						Type:             "FARGATE_SPOT",
						MaxVcpus:         128,
						SubnetIds:        []*foreignkeyv1.StringValueOrRef{svRef("subnet-aaa"), svRef("subnet-bbb")},
						SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{svRef("sg-111")},
					},
					JobQueues: []*AwsBatchJobQueue{
						{Name: "spot-queue", Priority: 1},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with scheduling policy", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalFargateSpec()
				spec.SchedulingPolicy = &AwsBatchSchedulingPolicy{
					ComputeReservation: int32Ptr(10),
					ShareDecaySeconds:  int32Ptr(3600),
					ShareDistributions: []*AwsBatchShareDistribution{
						{ShareIdentifier: "team-a", WeightFactor: 0.5},
						{ShareIdentifier: "team-b", WeightFactor: 1.0},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with update policy", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalEc2Spec()
				spec.UpdatePolicy = &AwsBatchUpdatePolicy{
					TerminateJobsOnUpdate:       true,
					JobExecutionTimeoutMinutes: int32Ptr(30),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with launch template by ID", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalEc2Spec()
				spec.ComputeResources.LaunchTemplate = &AwsBatchLaunchTemplate{
					LaunchTemplateId: "lt-0123456789abcdef0",
					Version:          "$Latest",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with launch template by name", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalEc2Spec()
				spec.ComputeResources.LaunchTemplate = &AwsBatchLaunchTemplate{
					LaunchTemplateName: "my-batch-template",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with ec2_configurations", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalEc2Spec()
				spec.ComputeResources.Ec2Configurations = []*AwsBatchEc2Configuration{
					{ImageType: "ECS_AL2023"},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with job state time limit actions", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalFargateSpec()
				spec.JobQueues[0].JobStateTimeLimitActions = []*AwsBatchJobStateTimeLimitAction{
					{
						Action:         "CANCEL",
						MaxTimeSeconds: 3600,
						Reason:         "Job stuck in RUNNABLE for over 1 hour",
						State:          "RUNNABLE",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with state set to DISABLED", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalFargateSpec()
				spec.State = stringPtr("DISABLED")
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("with no job queues", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsBatchComputeEnvironmentSpec{
					ComputeResources: &AwsBatchComputeResources{
						Type:             "FARGATE",
						MaxVcpus:         256,
						SubnetIds:        []*foreignkeyv1.StringValueOrRef{svRef("subnet-aaa")},
						SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{svRef("sg-111")},
					},
					JobQueues: []*AwsBatchJobQueue{},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with no compute_resources", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsBatchComputeEnvironmentSpec{
					JobQueues: []*AwsBatchJobQueue{
						{Name: "default", Priority: 1},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid compute resource type", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsBatchComputeEnvironmentSpec{
					ComputeResources: &AwsBatchComputeResources{
						Type:      "INVALID_TYPE",
						MaxVcpus:  256,
						SubnetIds: []*foreignkeyv1.StringValueOrRef{svRef("subnet-aaa")},
					},
					JobQueues: []*AwsBatchJobQueue{
						{Name: "default", Priority: 1},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with max_vcpus less than 1", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalFargateSpec()
				spec.ComputeResources.MaxVcpus = 0
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with no subnets", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsBatchComputeEnvironmentSpec{
					ComputeResources: &AwsBatchComputeResources{
						Type:             "FARGATE",
						MaxVcpus:         256,
						SubnetIds:        []*foreignkeyv1.StringValueOrRef{},
						SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{svRef("sg-111")},
					},
					JobQueues: []*AwsBatchJobQueue{
						{Name: "default", Priority: 1},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with EC2 type missing instance_role", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsBatchComputeEnvironmentSpec{
					ComputeResources: &AwsBatchComputeResources{
						Type:      "EC2",
						MaxVcpus:  256,
						SubnetIds: []*foreignkeyv1.StringValueOrRef{svRef("subnet-aaa")},
					},
					JobQueues: []*AwsBatchJobQueue{
						{Name: "default", Priority: 1},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with SPOT type missing spot_iam_fleet_role", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsBatchComputeEnvironmentSpec{
					ComputeResources: &AwsBatchComputeResources{
						Type:         "SPOT",
						MaxVcpus:     256,
						SubnetIds:    []*foreignkeyv1.StringValueOrRef{svRef("subnet-aaa")},
						InstanceRole: svRef("arn:aws:iam::123456789012:instance-profile/ecsInstanceRole"),
					},
					JobQueues: []*AwsBatchJobQueue{
						{Name: "default", Priority: 1},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid allocation_strategy", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalEc2Spec()
				spec.ComputeResources.AllocationStrategy = "INVALID_STRATEGY"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with bid_percentage out of range", func() {
			ginkgo.It("should return a validation error when > 100", func() {
				spec := minimalEc2Spec()
				spec.ComputeResources.Type = "SPOT"
				spec.ComputeResources.SpotIamFleetRole = svRef("arn:aws:iam::123456789012:role/spot-fleet-role")
				spec.ComputeResources.BidPercentage = int32Ptr(150)
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid job queue name pattern", func() {
			ginkgo.It("should return a validation error for names starting with hyphen", func() {
				spec := minimalFargateSpec()
				spec.JobQueues[0].Name = "-invalid-name"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with launch template having both id and name", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalEc2Spec()
				spec.ComputeResources.LaunchTemplate = &AwsBatchLaunchTemplate{
					LaunchTemplateId:   "lt-0123456789abcdef0",
					LaunchTemplateName: "my-template",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with launch template having neither id nor name", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalEc2Spec()
				spec.ComputeResources.LaunchTemplate = &AwsBatchLaunchTemplate{}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with ec2_configurations exceeding max 2", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalEc2Spec()
				spec.ComputeResources.Ec2Configurations = []*AwsBatchEc2Configuration{
					{ImageType: "ECS_AL2"},
					{ImageType: "ECS_AL2023"},
					{ImageType: "ECS_AL2"},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with job_execution_timeout_minutes out of range", func() {
			ginkgo.It("should return a validation error when > 360", func() {
				spec := minimalEc2Spec()
				spec.UpdatePolicy = &AwsBatchUpdatePolicy{
					TerminateJobsOnUpdate:       true,
					JobExecutionTimeoutMinutes: int32Ptr(500),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with max_time_seconds out of range in job state time limit action", func() {
			ginkgo.It("should return a validation error when < 600", func() {
				spec := minimalFargateSpec()
				spec.JobQueues[0].JobStateTimeLimitActions = []*AwsBatchJobStateTimeLimitAction{
					{
						Action:         "CANCEL",
						MaxTimeSeconds: 100,
						Reason:         "Too short",
						State:          "RUNNABLE",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with compute_reservation out of range", func() {
			ginkgo.It("should return a validation error when > 99", func() {
				spec := minimalFargateSpec()
				spec.SchedulingPolicy = &AwsBatchSchedulingPolicy{
					ComputeReservation: int32Ptr(100),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
