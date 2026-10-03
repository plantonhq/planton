package awsautoscalinggroupv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsAutoScalingGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsAutoScalingGroupSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimalValidGroup is the common case: a launch-template fleet across two
// subnets.
func minimalValidGroup() *AwsAutoScalingGroup {
	return &AwsAutoScalingGroup{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsAutoScalingGroup",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "web",
		},
		Spec: &AwsAutoScalingGroupSpec{
			Region: "us-west-2",
			Subnets: []*foreignkeyv1.StringValueOrRef{
				literalRef("subnet-0123456789abcdef0"),
				literalRef("subnet-0123456789abcdef1"),
			},
			LaunchTemplate: &AwsAutoScalingGroupLaunchTemplateRef{
				LaunchTemplateId: literalRef("lt-0123456789abcdef0"),
			},
			MinSize: 1,
			MaxSize: 4,
		},
	}
}

var _ = ginkgo.Describe("AwsAutoScalingGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_autoscaling_group", func() {

			ginkgo.It("should not return a validation error for a minimal group", func() {
				err := protovalidate.Validate(minimalValidGroup())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a web service behind a target group", func() {
				input := minimalValidGroup()
				input.Spec.DesiredCapacity = 2
				input.Spec.HealthCheckType = "ELB"
				input.Spec.HealthCheckGracePeriodSeconds = 120
				input.Spec.TargetGroups = []*foreignkeyv1.StringValueOrRef{
					literalRef("arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/api/50dc6c495c0c9188"),
				}
				input.Spec.TerminationPolicies = []string{"OldestLaunchTemplate", "Default"}
				minHealthy := int32(90)
				input.Spec.InstanceRefresh = &AwsAutoScalingGroupInstanceRefresh{
					Strategy: "Rolling",
					Preferences: &AwsAutoScalingGroupInstanceRefreshPreferences{
						MinHealthyPercentage: &minHealthy,
						MaxHealthyPercentage: 110,
						AutoRollback:         true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spot mixed-instances fleet", func() {
				input := minimalValidGroup()
				input.Spec.LaunchTemplate = nil
				onDemandPercentage := int32(20)
				input.Spec.MixedInstancesPolicy = &AwsAutoScalingGroupMixedInstancesPolicy{
					LaunchTemplate: &AwsAutoScalingGroupLaunchTemplateRef{
						LaunchTemplateId: literalRef("lt-0123456789abcdef0"),
					},
					Overrides: []*AwsAutoScalingGroupMixedInstancesOverride{
						{InstanceType: "m5.large"},
						{InstanceType: "m5a.large"},
						{InstanceType: "m6i.large", WeightedCapacity: 1},
					},
					InstancesDistribution: &AwsAutoScalingGroupInstancesDistribution{
						OnDemandBaseCapacity:                1,
						OnDemandPercentageAboveBaseCapacity: &onDemandPercentage,
						SpotAllocationStrategy:              "price-capacity-optimized",
					},
				}
				input.Spec.CapacityRebalance = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for attribute-based overrides", func() {
				input := minimalValidGroup()
				input.Spec.LaunchTemplate = nil
				input.Spec.MixedInstancesPolicy = &AwsAutoScalingGroupMixedInstancesPolicy{
					LaunchTemplate: &AwsAutoScalingGroupLaunchTemplateRef{
						LaunchTemplateId: literalRef("lt-0123456789abcdef0"),
					},
					Overrides: []*AwsAutoScalingGroupMixedInstancesOverride{
						{
							InstanceRequirements: &AwsAutoScalingGroupInstanceRequirements{
								MemoryMib:        &AwsAutoScalingGroupIntRange{Min: 4096, Max: 16384},
								VcpuCount:        &AwsAutoScalingGroupIntRange{Min: 2, Max: 8},
								CpuManufacturers: []string{"intel", "amd"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for scaling policies of every engine", func() {
				input := minimalValidGroup()
				returnData := false
				input.Spec.ScalingPolicies = []*AwsAutoScalingGroupScalingPolicy{
					{
						Name:       "cpu-target",
						PolicyType: "TargetTrackingScaling",
						TargetTracking: &AwsAutoScalingGroupTargetTrackingConfig{
							TargetValue:          60,
							PredefinedMetricType: "ASGAverageCPUUtilization",
						},
					},
					{
						Name:       "backlog-target",
						PolicyType: "TargetTrackingScaling",
						TargetTracking: &AwsAutoScalingGroupTargetTrackingConfig{
							TargetValue: 100,
							CustomizedMetric: &AwsAutoScalingGroupCustomizedMetric{
								Metrics: []*AwsAutoScalingGroupMetricDataQuery{
									{
										Id: "m1",
										MetricStat: &AwsAutoScalingGroupMetricStat{
											MetricName: "ApproximateNumberOfMessagesVisible",
											Namespace:  "AWS/SQS",
											Stat:       "Sum",
											Dimensions: []*AwsAutoScalingGroupMetricDimension{
												{Name: "QueueName", Value: "orders"},
											},
										},
										ReturnData: &returnData,
									},
									{Id: "e1", Expression: "m1 / 10"},
								},
							},
						},
					},
					{
						Name:       "step-up",
						PolicyType: "StepScaling",
						StepScaling: &AwsAutoScalingGroupStepScalingConfig{
							AdjustmentType: "ChangeInCapacity",
							StepAdjustments: []*AwsAutoScalingGroupStepAdjustment{
								{ScalingAdjustment: 1, MetricIntervalLowerBound: "0", MetricIntervalUpperBound: "10"},
								{ScalingAdjustment: 2, MetricIntervalLowerBound: "10"},
							},
						},
					},
					{
						Name:       "forecast",
						PolicyType: "PredictiveScaling",
						PredictiveScaling: &AwsAutoScalingGroupPredictiveScalingConfig{
							TargetValue:              60,
							PredefinedMetricPairType: "ASGCPUUtilization",
							Mode:                     "ForecastAndScale",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for schedules, hooks, warm pool, and notifications", func() {
				input := minimalValidGroup()
				minSize := int32(4)
				desired := int32(6)
				maxPrepared := int32(2)
				input.Spec.ScheduledActions = []*AwsAutoScalingGroupScheduledAction{
					{
						Name:            "business-hours",
						Recurrence:      "0 8 * * MON-FRI",
						TimeZone:        "America/New_York",
						MinSize:         &minSize,
						DesiredCapacity: &desired,
					},
				}
				input.Spec.LifecycleHooks = []*AwsAutoScalingGroupLifecycleHook{
					{
						Name:                    "drain",
						LifecycleTransition:     "autoscaling:EC2_INSTANCE_TERMINATING",
						DefaultResult:           "CONTINUE",
						HeartbeatTimeoutSeconds: 300,
					},
				}
				input.Spec.WarmPool = &AwsAutoScalingGroupWarmPool{
					PoolState:                "Stopped",
					MinSize:                  1,
					MaxGroupPreparedCapacity: &maxPrepared,
					ReuseOnScaleIn:           true,
				}
				input.Spec.Notifications = &AwsAutoScalingGroupNotifications{
					Topic: literalRef("arn:aws:sns:us-west-2:123456789012:fleet-events"),
					EventTypes: []string{
						"autoscaling:EC2_INSTANCE_LAUNCH_ERROR",
						"autoscaling:EC2_INSTANCE_TERMINATE_ERROR",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a reserved fleet with lifecycle policy and waits", func() {
				input := minimalValidGroup()
				input.Spec.CapacityReservation = &AwsAutoScalingGroupCapacityReservation{
					Preference:             "capacity-reservations-first",
					CapacityReservationIds: []string{"cr-0123456789abcdef0"},
				}
				input.Spec.InstanceLifecyclePolicy = &AwsAutoScalingGroupInstanceLifecyclePolicy{
					TerminateHookAbandon: "retain",
				}
				input.Spec.IgnoreFailedScalingActivities = true
				input.Spec.ForceDeleteWarmPool = true
				input.Spec.TargetGroups = []*foreignkeyv1.StringValueOrRef{
					literalRef("arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/api/50dc6c495c0c9188"),
				}
				input.Spec.MinElbCapacity = 1
				input.Spec.WaitForElbCapacity = 2
				input.Spec.LifecycleHooks = []*AwsAutoScalingGroupLifecycleHook{
					{
						Name:                "warm-cache",
						LifecycleTransition: "autoscaling:EC2_INSTANCE_LAUNCHING",
						ApplyAtLaunch:       true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for VPC Lattice traffic sources", func() {
				input := minimalValidGroup()
				input.Spec.TrafficSources = []*AwsAutoScalingGroupTrafficSource{
					{
						Identifier: literalRef("arn:aws:vpc-lattice:us-west-2:123456789012:targetgroup/tg-0123456789abcdef0"),
						Type:       "vpc-lattice",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a disabled policy with split predictive metrics", func() {
				input := minimalValidGroup()
				input.Spec.ScalingPolicies = []*AwsAutoScalingGroupScalingPolicy{
					{
						Name:       "forecast-split",
						PolicyType: "PredictiveScaling",
						Disabled:   true,
						PredictiveScaling: &AwsAutoScalingGroupPredictiveScalingConfig{
							TargetValue: 60,
							PredefinedLoadMetric: &AwsAutoScalingGroupPredictiveScalingPredefinedMetric{
								MetricType: "ASGTotalCPUUtilization",
							},
							PredefinedScalingMetric: &AwsAutoScalingGroupPredictiveScalingPredefinedMetric{
								MetricType: "ASGAverageCPUUtilization",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully customized predictive metrics", func() {
				input := minimalValidGroup()
				returnData := false
				input.Spec.ScalingPolicies = []*AwsAutoScalingGroupScalingPolicy{
					{
						Name:       "forecast-custom",
						PolicyType: "PredictiveScaling",
						PredictiveScaling: &AwsAutoScalingGroupPredictiveScalingConfig{
							TargetValue: 100,
							CustomizedLoadMetricQueries: []*AwsAutoScalingGroupMetricDataQuery{
								{
									Id: "load",
									MetricStat: &AwsAutoScalingGroupMetricStat{
										MetricName: "RequestsTotal",
										Namespace:  "MyApp",
										Stat:       "Sum",
									},
								},
							},
							CustomizedScalingMetricQueries: []*AwsAutoScalingGroupMetricDataQuery{
								{
									Id: "m1",
									MetricStat: &AwsAutoScalingGroupMetricStat{
										MetricName: "RequestsPerInstance",
										Namespace:  "MyApp",
										Stat:       "Average",
									},
									ReturnData: &returnData,
								},
								{Id: "scaling", Expression: "m1 * 1.1"},
							},
							CustomizedCapacityMetricQueries: []*AwsAutoScalingGroupMetricDataQuery{
								{
									Id: "capacity",
									MetricStat: &AwsAutoScalingGroupMetricStat{
										MetricName: "GroupInServiceInstances",
										Namespace:  "AWS/AutoScaling",
										Stat:       "Average",
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("spec-level rules", func() {

			ginkgo.It("should return an error when subnets are missing", func() {
				input := minimalValidGroup()
				input.Spec.Subnets = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when neither launch_template nor mixed_instances_policy is set", func() {
				input := minimalValidGroup()
				input.Spec.LaunchTemplate = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when both launch_template and mixed_instances_policy are set", func() {
				input := minimalValidGroup()
				input.Spec.MixedInstancesPolicy = &AwsAutoScalingGroupMixedInstancesPolicy{
					LaunchTemplate: &AwsAutoScalingGroupLaunchTemplateRef{
						LaunchTemplateId: literalRef("lt-0123456789abcdef0"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when max_size is below min_size", func() {
				input := minimalValidGroup()
				input.Spec.MinSize = 4
				input.Spec.MaxSize = 2
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when desired_capacity is outside the bounds", func() {
				input := minimalValidGroup()
				input.Spec.DesiredCapacity = 10
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an invalid health_check_type", func() {
				input := minimalValidGroup()
				input.Spec.HealthCheckType = "TCP"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a max_instance_lifetime below one day", func() {
				input := minimalValidGroup()
				input.Spec.MaxInstanceLifetimeSeconds = 3600
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an unknown suspended process", func() {
				input := minimalValidGroup()
				input.Spec.SuspendedProcesses = []string{"Reboot"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when traffic_sources and target_groups are both set", func() {
				input := minimalValidGroup()
				input.Spec.TargetGroups = []*foreignkeyv1.StringValueOrRef{
					literalRef("arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/api/50dc6c495c0c9188"),
				}
				input.Spec.TrafficSources = []*AwsAutoScalingGroupTrafficSource{
					{Identifier: literalRef("arn:aws:vpc-lattice:us-west-2:123456789012:targetgroup/tg-0123456789abcdef0")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an unknown traffic source type", func() {
				input := minimalValidGroup()
				input.Spec.TrafficSources = []*AwsAutoScalingGroupTrafficSource{
					{Identifier: literalRef("my-classic-elb"), Type: "clb"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when capacity reservation ids and resource groups are both set", func() {
				input := minimalValidGroup()
				input.Spec.CapacityReservation = &AwsAutoScalingGroupCapacityReservation{
					CapacityReservationIds:               []string{"cr-0123456789abcdef0"},
					CapacityReservationResourceGroupArns: []string{"arn:aws:resource-groups:us-west-2:123456789012:group/crs"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an unknown capacity reservation preference", func() {
				input := minimalValidGroup()
				input.Spec.CapacityReservation = &AwsAutoScalingGroupCapacityReservation{
					Preference: "always",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an unknown terminate_hook_abandon action", func() {
				input := minimalValidGroup()
				input.Spec.InstanceLifecyclePolicy = &AwsAutoScalingGroupInstanceLifecyclePolicy{
					TerminateHookAbandon: "keep",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("predictive scaling metric-form rules", func() {

			ginkgo.It("should return an error when the pair combines with customized queries", func() {
				input := minimalValidGroup()
				input.Spec.ScalingPolicies = []*AwsAutoScalingGroupScalingPolicy{
					{
						Name:       "forecast",
						PolicyType: "PredictiveScaling",
						PredictiveScaling: &AwsAutoScalingGroupPredictiveScalingConfig{
							TargetValue:              60,
							PredefinedMetricPairType: "ASGCPUUtilization",
							CustomizedScalingMetricQueries: []*AwsAutoScalingGroupMetricDataQuery{
								{Id: "m1", Expression: "10"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when no scaling metric is given", func() {
				input := minimalValidGroup()
				input.Spec.ScalingPolicies = []*AwsAutoScalingGroupScalingPolicy{
					{
						Name:       "forecast",
						PolicyType: "PredictiveScaling",
						PredictiveScaling: &AwsAutoScalingGroupPredictiveScalingConfig{
							TargetValue: 60,
							PredefinedLoadMetric: &AwsAutoScalingGroupPredictiveScalingPredefinedMetric{
								MetricType: "ASGTotalCPUUtilization",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when predefined and customized load metrics are both set", func() {
				input := minimalValidGroup()
				input.Spec.ScalingPolicies = []*AwsAutoScalingGroupScalingPolicy{
					{
						Name:       "forecast",
						PolicyType: "PredictiveScaling",
						PredictiveScaling: &AwsAutoScalingGroupPredictiveScalingConfig{
							TargetValue: 60,
							PredefinedLoadMetric: &AwsAutoScalingGroupPredictiveScalingPredefinedMetric{
								MetricType: "ASGTotalCPUUtilization",
							},
							CustomizedLoadMetricQueries: []*AwsAutoScalingGroupMetricDataQuery{
								{Id: "load", Expression: "10"},
							},
							PredefinedScalingMetric: &AwsAutoScalingGroupPredictiveScalingPredefinedMetric{
								MetricType: "ASGAverageCPUUtilization",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a scaling-side metric type on the load side", func() {
				input := minimalValidGroup()
				input.Spec.ScalingPolicies = []*AwsAutoScalingGroupScalingPolicy{
					{
						Name:       "forecast",
						PolicyType: "PredictiveScaling",
						PredictiveScaling: &AwsAutoScalingGroupPredictiveScalingConfig{
							TargetValue: 60,
							PredefinedLoadMetric: &AwsAutoScalingGroupPredictiveScalingPredefinedMetric{
								MetricType: "ASGAverageCPUUtilization",
							},
							PredefinedScalingMetric: &AwsAutoScalingGroupPredictiveScalingPredefinedMetric{
								MetricType: "ASGAverageCPUUtilization",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("mixed-instances rules", func() {

			ginkgo.It("should return an error when an override sets both type and requirements", func() {
				input := minimalValidGroup()
				input.Spec.LaunchTemplate = nil
				input.Spec.MixedInstancesPolicy = &AwsAutoScalingGroupMixedInstancesPolicy{
					LaunchTemplate: &AwsAutoScalingGroupLaunchTemplateRef{
						LaunchTemplateId: literalRef("lt-0123456789abcdef0"),
					},
					Overrides: []*AwsAutoScalingGroupMixedInstancesOverride{
						{
							InstanceType: "m5.large",
							InstanceRequirements: &AwsAutoScalingGroupInstanceRequirements{
								MemoryMib: &AwsAutoScalingGroupIntRange{Min: 4096},
								VcpuCount: &AwsAutoScalingGroupIntRange{Min: 2},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an invalid spot allocation strategy", func() {
				input := minimalValidGroup()
				input.Spec.LaunchTemplate = nil
				input.Spec.MixedInstancesPolicy = &AwsAutoScalingGroupMixedInstancesPolicy{
					LaunchTemplate: &AwsAutoScalingGroupLaunchTemplateRef{
						LaunchTemplateId: literalRef("lt-0123456789abcdef0"),
					},
					InstancesDistribution: &AwsAutoScalingGroupInstancesDistribution{
						SpotAllocationStrategy: "cheapest",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when spot_instance_pools is used without lowest-price", func() {
				input := minimalValidGroup()
				input.Spec.LaunchTemplate = nil
				input.Spec.MixedInstancesPolicy = &AwsAutoScalingGroupMixedInstancesPolicy{
					LaunchTemplate: &AwsAutoScalingGroupLaunchTemplateRef{
						LaunchTemplateId: literalRef("lt-0123456789abcdef0"),
					},
					InstancesDistribution: &AwsAutoScalingGroupInstancesDistribution{
						SpotAllocationStrategy: "capacity-optimized",
						SpotInstancePools:      4,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("instance refresh rules", func() {

			ginkgo.It("should return an error for an unknown strategy", func() {
				input := minimalValidGroup()
				input.Spec.InstanceRefresh = &AwsAutoScalingGroupInstanceRefresh{Strategy: "BlueGreen"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a max_healthy_percentage below 100", func() {
				input := minimalValidGroup()
				input.Spec.InstanceRefresh = &AwsAutoScalingGroupInstanceRefresh{
					Strategy: "Rolling",
					Preferences: &AwsAutoScalingGroupInstanceRefreshPreferences{
						MaxHealthyPercentage: 90,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("scaling policy rules", func() {

			ginkgo.It("should return an error when the config block does not match the policy type", func() {
				input := minimalValidGroup()
				input.Spec.ScalingPolicies = []*AwsAutoScalingGroupScalingPolicy{
					{
						Name:       "mismatched",
						PolicyType: "StepScaling",
						TargetTracking: &AwsAutoScalingGroupTargetTrackingConfig{
							TargetValue:          60,
							PredefinedMetricType: "ASGAverageCPUUtilization",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when target tracking sets both predefined and customized metrics", func() {
				input := minimalValidGroup()
				input.Spec.ScalingPolicies = []*AwsAutoScalingGroupScalingPolicy{
					{
						Name:       "double-metric",
						PolicyType: "TargetTrackingScaling",
						TargetTracking: &AwsAutoScalingGroupTargetTrackingConfig{
							TargetValue:          60,
							PredefinedMetricType: "ASGAverageCPUUtilization",
							CustomizedMetric: &AwsAutoScalingGroupCustomizedMetric{
								MetricName: "Depth",
								Namespace:  "MyApp",
								Statistic:  "Average",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a zero step adjustment on a delta policy", func() {
				input := minimalValidGroup()
				input.Spec.ScalingPolicies = []*AwsAutoScalingGroupScalingPolicy{
					{
						Name:       "noop-step",
						PolicyType: "StepScaling",
						StepScaling: &AwsAutoScalingGroupStepScalingConfig{
							AdjustmentType: "ChangeInCapacity",
							StepAdjustments: []*AwsAutoScalingGroupStepAdjustment{
								{ScalingAdjustment: 0, MetricIntervalLowerBound: "0"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a metric data query with neither expression nor metric_stat", func() {
				input := minimalValidGroup()
				input.Spec.ScalingPolicies = []*AwsAutoScalingGroupScalingPolicy{
					{
						Name:       "empty-query",
						PolicyType: "TargetTrackingScaling",
						TargetTracking: &AwsAutoScalingGroupTargetTrackingConfig{
							TargetValue: 100,
							CustomizedMetric: &AwsAutoScalingGroupCustomizedMetric{
								Metrics: []*AwsAutoScalingGroupMetricDataQuery{{Id: "m1"}},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("scheduled action rules", func() {

			ginkgo.It("should return an error when no capacity value changes", func() {
				input := minimalValidGroup()
				input.Spec.ScheduledActions = []*AwsAutoScalingGroupScheduledAction{
					{Name: "noop", Recurrence: "0 8 * * *"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when neither recurrence nor start_time is set", func() {
				input := minimalValidGroup()
				desired := int32(3)
				input.Spec.ScheduledActions = []*AwsAutoScalingGroupScheduledAction{
					{Name: "floating", DesiredCapacity: &desired},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("lifecycle hook rules", func() {

			ginkgo.It("should return an error for an invalid lifecycle transition", func() {
				input := minimalValidGroup()
				input.Spec.LifecycleHooks = []*AwsAutoScalingGroupLifecycleHook{
					{Name: "bad", LifecycleTransition: "autoscaling:EC2_INSTANCE_REBOOTING"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a heartbeat timeout out of range", func() {
				input := minimalValidGroup()
				input.Spec.LifecycleHooks = []*AwsAutoScalingGroupLifecycleHook{
					{
						Name:                    "short",
						LifecycleTransition:     "autoscaling:EC2_INSTANCE_LAUNCHING",
						HeartbeatTimeoutSeconds: 10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("warm pool and notification rules", func() {

			ginkgo.It("should return an error for an invalid pool state", func() {
				input := minimalValidGroup()
				input.Spec.WarmPool = &AwsAutoScalingGroupWarmPool{PoolState: "Paused"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an unknown notification event type", func() {
				input := minimalValidGroup()
				input.Spec.Notifications = &AwsAutoScalingGroupNotifications{
					Topic:      literalRef("arn:aws:sns:us-west-2:123456789012:fleet-events"),
					EventTypes: []string{"autoscaling:EC2_INSTANCE_REBOOT"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
