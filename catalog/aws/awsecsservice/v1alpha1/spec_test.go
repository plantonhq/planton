package awsecsservicev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsEcsServiceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsEcsServiceSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

// minimalValidService is the common case: a Fargate service running a
// referenced task-definition revision in a referenced cluster, on private
// subnets.
func minimalValidService() *AwsEcsService {
	return &AwsEcsService{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsEcsService",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "api",
		},
		Spec: &AwsEcsServiceSpec{
			Region:         "us-west-2",
			ClusterArn:     literalRef("arn:aws:ecs:us-west-2:123456789012:cluster/prod"),
			TaskDefinition: literalRef("arn:aws:ecs:us-west-2:123456789012:task-definition/api:3"),
			Network: &AwsEcsServiceNetwork{
				Subnets: []*foreignkeyv1.StringValueOrRef{
					literalRef("subnet-0123456789abcdef0"),
					literalRef("subnet-0fedcba9876543210"),
				},
			},
		},
	}
}

func validLoadBalancer() *AwsEcsServiceLoadBalancer {
	return &AwsEcsServiceLoadBalancer{
		TargetGroupArn: literalRef("arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/api/50dc6c495c0c9188"),
		ContainerName:  "api",
		ContainerPort:  8080,
	}
}

var _ = ginkgo.Describe("AwsEcsServiceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_ecs_service", func() {

			ginkgo.It("should not return a validation error for a minimal service", func() {
				err := protovalidate.Validate(minimalValidService())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept explicit zero desired_count", func() {
				input := minimalValidService()
				input.Spec.DesiredCount = int32Ptr(0)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a capacity-provider blend without a launch type", func() {
				input := minimalValidService()
				input.Spec.CapacityProviderStrategy = []*AwsEcsServiceCapacityProviderStrategy{
					{CapacityProvider: "FARGATE", Base: 1, Weight: 1},
					{CapacityProvider: "FARGATE_SPOT", Weight: 4},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept load-balancer wiring with a grace period", func() {
				input := minimalValidService()
				input.Spec.LoadBalancers = []*AwsEcsServiceLoadBalancer{validLoadBalancer()}
				input.Spec.HealthCheckGracePeriodSeconds = int32Ptr(90)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept deployment guards (circuit breaker + alarms)", func() {
				input := minimalValidService()
				input.Spec.DeploymentCircuitBreaker = &AwsEcsServiceDeploymentCircuitBreaker{
					Enable:   true,
					Rollback: true,
				}
				input.Spec.Alarms = &AwsEcsServiceDeploymentAlarms{
					AlarmNames: []*foreignkeyv1.StringValueOrRef{literalRef("api-5xx-rate")},
					Enable:     true,
					Rollback:   true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a blue/green deployment with canary shifting", func() {
				input := minimalValidService()
				lb := validLoadBalancer()
				lb.AdvancedConfiguration = &AwsEcsServiceLoadBalancerAdvancedConfiguration{
					AlternateTargetGroupArn: literalRef("arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/api-green/943f017f100becff"),
					ProductionListenerRule:  literalRef("arn:aws:elasticloadbalancing:us-west-2:123456789012:listener-rule/app/main/50dc6c495c0c9188/f2f7dc8efc522ab2/9683b2d02a6cabee"),
					RoleArn:                 literalRef("arn:aws:iam::123456789012:role/ecs-blue-green"),
				}
				input.Spec.LoadBalancers = []*AwsEcsServiceLoadBalancer{lb}
				input.Spec.DeploymentConfiguration = &AwsEcsServiceDeploymentConfiguration{
					Strategy:          "BLUE_GREEN",
					BakeTimeInMinutes: int32Ptr(10),
					CanaryConfiguration: &AwsEcsServiceDeploymentCanary{
						CanaryPercent:           5,
						CanaryBakeTimeInMinutes: int32Ptr(5),
					},
					LifecycleHooks: []*AwsEcsServiceDeploymentLifecycleHook{
						{
							HookTargetArn:   "arn:aws:lambda:us-west-2:123456789012:function:smoke-tests",
							RoleArn:         literalRef("arn:aws:iam::123456789012:role/ecs-hooks"),
							LifecycleStages: []string{"POST_TEST_TRAFFIC_SHIFT"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept Service Connect with an exposed service", func() {
				input := minimalValidService()
				input.Spec.ServiceConnect = &AwsEcsServiceServiceConnect{
					Enabled:   true,
					Namespace: "internal",
					Services: []*AwsEcsServiceServiceConnectService{
						{
							PortName: "http",
							ClientAlias: &AwsEcsServiceServiceConnectClientAlias{
								Port:    80,
								DnsName: "api",
							},
							Timeout: &AwsEcsServiceServiceConnectTimeout{
								PerRequestTimeoutSeconds: 30,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a managed EBS volume configuration", func() {
				input := minimalValidService()
				input.Spec.VolumeConfiguration = &AwsEcsServiceVolumeConfiguration{
					Name: "scratch",
					ManagedEbsVolume: &AwsEcsServiceManagedEbsVolume{
						RoleArn:    literalRef("arn:aws:iam::123456789012:role/ecs-volumes"),
						SizeInGb:   100,
						VolumeType: "gp3",
						Throughput: 250,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept EC2 placement strategies and constraints", func() {
				input := minimalValidService()
				input.Spec.LaunchType = "EC2"
				input.Spec.OrderedPlacementStrategy = []*AwsEcsServicePlacementStrategy{
					{Type: "spread", Field: "attribute:ecs.availability-zone"},
					{Type: "binpack", Field: "memory"},
				}
				input.Spec.PlacementConstraints = []*AwsEcsServicePlacementConstraint{
					{Type: "distinctInstance"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept autoscaling on CPU and requests per target", func() {
				input := minimalValidService()
				input.Spec.Autoscaling = &AwsEcsServiceAutoscaling{
					MinTasks: 2,
					MaxTasks: 20,
					Cpu: &AwsEcsServiceAutoscalingTarget{
						TargetPercent: 70,
					},
					RequestsPerTarget: &AwsEcsServiceAutoscalingRequestCountTarget{
						TargetRequestsPerTarget: 1000,
						LoadBalancerArnSuffix:   literalRef("app/main/50dc6c495c0c9188"),
						TargetGroupArnSuffix:    literalRef("targetgroup/api/943f017f100becff"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a DAEMON service on EC2 without autoscaling", func() {
				input := minimalValidService()
				input.Spec.LaunchType = "EC2"
				input.Spec.SchedulingStrategy = "DAEMON"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_ecs_service", func() {

			ginkgo.It("should return an error when region is empty", func() {
				input := minimalValidService()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when cluster_arn is missing", func() {
				input := minimalValidService()
				input.Spec.ClusterArn = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when task_definition is missing", func() {
				input := minimalValidService()
				input.Spec.TaskDefinition = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown launch type", func() {
				input := minimalValidService()
				input.Spec.LaunchType = "SERVERLESS"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a launch type combined with capacity providers", func() {
				input := minimalValidService()
				input.Spec.LaunchType = "FARGATE"
				input.Spec.CapacityProviderStrategy = []*AwsEcsServiceCapacityProviderStrategy{
					{CapacityProvider: "FARGATE_SPOT", Weight: 1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject DAEMON scheduling on Fargate", func() {
				input := minimalValidService()
				input.Spec.SchedulingStrategy = "DAEMON"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject autoscaling on a DAEMON service", func() {
				input := minimalValidService()
				input.Spec.LaunchType = "EC2"
				input.Spec.SchedulingStrategy = "DAEMON"
				input.Spec.Autoscaling = &AwsEcsServiceAutoscaling{
					MinTasks: 1,
					MaxTasks: 2,
					Cpu:      &AwsEcsServiceAutoscalingTarget{TargetPercent: 70},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a grace period without load balancers", func() {
				input := minimalValidService()
				input.Spec.HealthCheckGracePeriodSeconds = int32Ptr(60)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a platform version on the EC2 launch type", func() {
				input := minimalValidService()
				input.Spec.LaunchType = "EC2"
				input.Spec.PlatformVersion = "1.4.0"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject placement strategies on Fargate", func() {
				input := minimalValidService()
				input.Spec.OrderedPlacementStrategy = []*AwsEcsServicePlacementStrategy{
					{Type: "binpack", Field: "memory"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown AZ-rebalancing value", func() {
				input := minimalValidService()
				input.Spec.AvailabilityZoneRebalancing = "AUTO"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown propagate_tags value", func() {
				input := minimalValidService()
				input.Spec.PropagateTags = "CLUSTER"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown deployment controller", func() {
				input := minimalValidService()
				input.Spec.DeploymentController = "SPINNAKER"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when network subnets are missing", func() {
				input := minimalValidService()
				input.Spec.Network.Subnets = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a load balancer without a target group", func() {
				input := minimalValidService()
				lb := validLoadBalancer()
				lb.TargetGroupArn = nil
				input.Spec.LoadBalancers = []*AwsEcsServiceLoadBalancer{lb}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a load balancer without a container name", func() {
				input := minimalValidService()
				lb := validLoadBalancer()
				lb.ContainerName = ""
				input.Spec.LoadBalancers = []*AwsEcsServiceLoadBalancer{lb}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an out-of-range load-balancer container port", func() {
				input := minimalValidService()
				lb := validLoadBalancer()
				lb.ContainerPort = 70000
				input.Spec.LoadBalancers = []*AwsEcsServiceLoadBalancer{lb}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject circuit-breaker rollback without enable", func() {
				input := minimalValidService()
				input.Spec.DeploymentCircuitBreaker = &AwsEcsServiceDeploymentCircuitBreaker{
					Rollback: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject alarm rollback without enable", func() {
				input := minimalValidService()
				input.Spec.Alarms = &AwsEcsServiceDeploymentAlarms{
					AlarmNames: []*foreignkeyv1.StringValueOrRef{literalRef("api-5xx-rate")},
					Rollback:   true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown deployment strategy", func() {
				input := minimalValidService()
				input.Spec.DeploymentConfiguration = &AwsEcsServiceDeploymentConfiguration{
					Strategy: "CANARY",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject canary and linear shifting together", func() {
				input := minimalValidService()
				input.Spec.DeploymentConfiguration = &AwsEcsServiceDeploymentConfiguration{
					Strategy:            "BLUE_GREEN",
					CanaryConfiguration: &AwsEcsServiceDeploymentCanary{CanaryPercent: 5},
					LinearConfiguration: &AwsEcsServiceDeploymentLinear{StepPercent: 10},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject traffic shifting without the BLUE_GREEN strategy", func() {
				input := minimalValidService()
				input.Spec.DeploymentConfiguration = &AwsEcsServiceDeploymentConfiguration{
					CanaryConfiguration: &AwsEcsServiceDeploymentCanary{CanaryPercent: 5},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a lifecycle hook with an unknown stage", func() {
				input := minimalValidService()
				input.Spec.DeploymentConfiguration = &AwsEcsServiceDeploymentConfiguration{
					Strategy: "BLUE_GREEN",
					LifecycleHooks: []*AwsEcsServiceDeploymentLifecycleHook{
						{
							HookTargetArn:   "arn:aws:lambda:us-west-2:123456789012:function:smoke-tests",
							RoleArn:         literalRef("arn:aws:iam::123456789012:role/ecs-hooks"),
							LifecycleStages: []string{"BEFORE_DEPLOY"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject exposed Service Connect services when not enabled", func() {
				input := minimalValidService()
				input.Spec.ServiceConnect = &AwsEcsServiceServiceConnect{
					Services: []*AwsEcsServiceServiceConnectService{
						{PortName: "http"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown Service Connect log driver", func() {
				input := minimalValidService()
				input.Spec.ServiceConnect = &AwsEcsServiceServiceConnect{
					Enabled: true,
					LogConfiguration: &AwsEcsServiceLogConfiguration{
						LogDriver: "logstash",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a service-registries block without a registry ARN", func() {
				input := minimalValidService()
				input.Spec.ServiceRegistries = &AwsEcsServiceServiceRegistries{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a managed EBS volume with neither size nor snapshot", func() {
				input := minimalValidService()
				input.Spec.VolumeConfiguration = &AwsEcsServiceVolumeConfiguration{
					Name: "scratch",
					ManagedEbsVolume: &AwsEcsServiceManagedEbsVolume{
						RoleArn: literalRef("arn:aws:iam::123456789012:role/ecs-volumes"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject EBS throughput on a non-gp3 volume", func() {
				input := minimalValidService()
				input.Spec.VolumeConfiguration = &AwsEcsServiceVolumeConfiguration{
					Name: "scratch",
					ManagedEbsVolume: &AwsEcsServiceManagedEbsVolume{
						RoleArn:    literalRef("arn:aws:iam::123456789012:role/ecs-volumes"),
						SizeInGb:   100,
						VolumeType: "gp2",
						Throughput: 250,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject autoscaling with max below min", func() {
				input := minimalValidService()
				input.Spec.Autoscaling = &AwsEcsServiceAutoscaling{
					MinTasks: 5,
					MaxTasks: 2,
					Cpu:      &AwsEcsServiceAutoscalingTarget{TargetPercent: 70},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject autoscaling with no tracking policy", func() {
				input := minimalValidService()
				input.Spec.Autoscaling = &AwsEcsServiceAutoscaling{
					MinTasks: 1,
					MaxTasks: 5,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a request-count policy without the ARN suffixes", func() {
				input := minimalValidService()
				input.Spec.Autoscaling = &AwsEcsServiceAutoscaling{
					MinTasks: 1,
					MaxTasks: 5,
					RequestsPerTarget: &AwsEcsServiceAutoscalingRequestCountTarget{
						TargetRequestsPerTarget: 1000,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("vpc lattice configurations", func() {
		ginkgo.It("should accept a complete target-group attachment", func() {
			input := minimalValidService()
			input.Spec.VpcLatticeConfigurations = []*AwsEcsServiceVpcLatticeConfiguration{{
				RoleArn:        literalRef("arn:aws:iam::123456789012:role/ecs-lattice-infrastructure"),
				TargetGroupArn: "arn:aws:vpc-lattice:us-west-2:123456789012:targetgroup/tg-0123456789abcdef0",
				PortName:       "http",
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should reject an attachment without a role", func() {
			input := minimalValidService()
			input.Spec.VpcLatticeConfigurations = []*AwsEcsServiceVpcLatticeConfiguration{{
				TargetGroupArn: "arn:aws:vpc-lattice:us-west-2:123456789012:targetgroup/tg-0123456789abcdef0",
				PortName:       "http",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid port name", func() {
			input := minimalValidService()
			input.Spec.VpcLatticeConfigurations = []*AwsEcsServiceVpcLatticeConfiguration{{
				RoleArn:        literalRef("arn:aws:iam::123456789012:role/ecs-lattice-infrastructure"),
				TargetGroupArn: "arn:aws:vpc-lattice:us-west-2:123456789012:targetgroup/tg-0123456789abcdef0",
				PortName:       "HTTP Port",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("service connect access logs", func() {
		ginkgo.It("should accept JSON access logging on an enabled mesh", func() {
			input := minimalValidService()
			input.Spec.ServiceConnect = &AwsEcsServiceServiceConnect{
				Enabled: true,
				AccessLogConfiguration: &AwsEcsServiceServiceConnectAccessLog{
					Format:                 "JSON",
					IncludeQueryParameters: "DISABLED",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should reject access logs when Service Connect is not enabled", func() {
			input := minimalValidService()
			input.Spec.ServiceConnect = &AwsEcsServiceServiceConnect{
				AccessLogConfiguration: &AwsEcsServiceServiceConnectAccessLog{
					Format: "TEXT",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid access log format", func() {
			input := minimalValidService()
			input.Spec.ServiceConnect = &AwsEcsServiceServiceConnect{
				Enabled: true,
				AccessLogConfiguration: &AwsEcsServiceServiceConnectAccessLog{
					Format: "CSV",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
