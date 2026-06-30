package awsecsservicev1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"
)

func TestAwsEcsService(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsEcsService Suite")
}

var _ = ginkgo.Describe("AwsEcsService Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_ecs_service", func() {
			var validInput *AwsEcsService

			ginkgo.BeforeEach(func() {
				validInput = &AwsEcsService{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsEcsService",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-service",
					},
					Spec: &AwsEcsServiceSpec{
						Region: "us-east-1",
						ClusterArn: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster"},
						},
						Container: &AwsEcsServiceContainer{
							Image: &AwsEcsServiceContainerImage{
								Repo: "example-repo",
								Tag:  "latest",
							},
							Cpu:    512,
							Memory: 1024,
						},
						Network: &AwsEcsServiceNetwork{
							Subnets: []*foreignkeyv1.StringValueOrRef{
								{
									LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-abc123"},
								},
							},
						},
						Alb: &AwsEcsServiceAlb{
							Enabled: true,
							Arn: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/myAlb"},
							},
							RoutingType:      "path",
							ListenerPort:     80,
							ListenerPriority: proto.Int32(100),
						},
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(validInput)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Context("api_version validations", func() {
		var input *AwsEcsService

		ginkgo.BeforeEach(func() {
			input = &AwsEcsService{
				ApiVersion: "aws.planton.dev/v1",
				Kind:       "AwsEcsService",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-service",
				},
				Spec: &AwsEcsServiceSpec{
					Region: "us-east-1",
					ClusterArn: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
							Value: "arn:aws:ecs:us-east-1:123456789012:cluster/test-cluster"},
					},
					Container: &AwsEcsServiceContainer{
						Image: &AwsEcsServiceContainerImage{
							Repo: "example-repo",
							Tag:  "latest",
						},
						Cpu:    256,
						Memory: 512,
					},
					Network: &AwsEcsServiceNetwork{
						Subnets: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-abc123"},
							},
						},
					},
				},
			}
		})

		ginkgo.It("should fail if api_version does not match 'aws.planton.dev/v1'", func() {
			input.ApiVersion = "invalid-version"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("kind validations", func() {
		var input *AwsEcsService

		ginkgo.BeforeEach(func() {
			input = &AwsEcsService{
				ApiVersion: "aws.planton.dev/v1",
				Kind:       "AwsEcsService",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-service",
				},
				Spec: &AwsEcsServiceSpec{
					Region: "us-east-1",
					ClusterArn: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
							Value: "arn:aws:ecs:us-east-1:123456789012:cluster/test-cluster"},
					},
					Container: &AwsEcsServiceContainer{
						Image: &AwsEcsServiceContainerImage{
							Repo: "example-repo",
							Tag:  "latest",
						},
						Cpu:    256,
						Memory: 512,
					},
					Network: &AwsEcsServiceNetwork{
						Subnets: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-abc123"},
							},
						},
					},
				},
			}
		})

		ginkgo.It("should fail if kind does not match 'AwsEcsService'", func() {
			input.Kind = "SomeOtherKind"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("routing_type validations", func() {
		var input *AwsEcsService

		ginkgo.BeforeEach(func() {
			input = &AwsEcsService{
				ApiVersion: "aws.planton.dev/v1",
				Kind:       "AwsEcsService",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-service",
				},
				Spec: &AwsEcsServiceSpec{
					Region: "us-east-1",
					ClusterArn: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
							Value: "arn:aws:ecs:us-east-1:123456789012:cluster/test-cluster"},
					},
					Container: &AwsEcsServiceContainer{
						Image: &AwsEcsServiceContainerImage{
							Repo: "example-repo",
							Tag:  "latest",
						},
						Cpu:    256,
						Memory: 512,
					},
					Network: &AwsEcsServiceNetwork{
						Subnets: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-abc123"},
							},
						},
					},
					Alb: &AwsEcsServiceAlb{
						Enabled: true,
						Arn: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/myAlb"},
						},
						RoutingType:      "path",
						ListenerPort:     80,
						ListenerPriority: proto.Int32(100),
					},
				},
			}
		})

		ginkgo.It("should allow 'path'", func() {
			input.Spec.Alb.RoutingType = "path"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should allow 'hostname'", func() {
			input.Spec.Alb.RoutingType = "hostname"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail for any other value", func() {
			input.Spec.Alb.RoutingType = "invalid"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("autoscaling validations", func() {
		var input *AwsEcsService

		ginkgo.BeforeEach(func() {
			input = &AwsEcsService{
				ApiVersion: "aws.planton.dev/v1",
				Kind:       "AwsEcsService",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-service",
				},
				Spec: &AwsEcsServiceSpec{
					Region: "us-east-1",
					ClusterArn: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
							Value: "arn:aws:ecs:us-east-1:123456789012:cluster/test-cluster"},
					},
					Container: &AwsEcsServiceContainer{
						Image: &AwsEcsServiceContainerImage{
							Repo: "example-repo",
							Tag:  "latest",
						},
						Cpu:    256,
						Memory: 512,
					},
					Network: &AwsEcsServiceNetwork{
						Subnets: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-abc123"},
							},
						},
					},
					Autoscaling: &AwsEcsServiceAutoscaling{
						Enabled:             true,
						MinTasks:            1,
						MaxTasks:            10,
						TargetCpuPercent:    proto.Int32(75),
						TargetMemoryPercent: proto.Int32(80),
					},
				},
			}
		})

		ginkgo.It("should pass with valid autoscaling configuration", func() {
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail if min_tasks is less than 1", func() {
			input.Spec.Autoscaling.MinTasks = 0
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail if max_tasks is less than 1", func() {
			input.Spec.Autoscaling.MaxTasks = 0
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail if target_cpu_percent is less than 1", func() {
			input.Spec.Autoscaling.TargetCpuPercent = proto.Int32(0)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail if target_cpu_percent is greater than 100", func() {
			input.Spec.Autoscaling.TargetCpuPercent = proto.Int32(101)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail if target_memory_percent is less than 1", func() {
			input.Spec.Autoscaling.TargetMemoryPercent = proto.Int32(0)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail if target_memory_percent is greater than 100", func() {
			input.Spec.Autoscaling.TargetMemoryPercent = proto.Int32(101)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("health_check_grace_period validations", func() {
		var input *AwsEcsService

		ginkgo.BeforeEach(func() {
			input = &AwsEcsService{
				ApiVersion: "aws.planton.dev/v1",
				Kind:       "AwsEcsService",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-service",
				},
				Spec: &AwsEcsServiceSpec{
					Region: "us-east-1",
					ClusterArn: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
							Value: "arn:aws:ecs:us-east-1:123456789012:cluster/test-cluster"},
					},
					Container: &AwsEcsServiceContainer{
						Image: &AwsEcsServiceContainerImage{
							Repo: "example-repo",
							Tag:  "latest",
						},
						Cpu:    256,
						Memory: 512,
					},
					Network: &AwsEcsServiceNetwork{
						Subnets: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-abc123"},
							},
						},
					},
					Alb: &AwsEcsServiceAlb{
						Enabled: true,
						Arn: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/myAlb"},
						},
						RoutingType:      "path",
						ListenerPort:     80,
						ListenerPriority: proto.Int32(100),
					},
					HealthCheckGracePeriodSeconds: proto.Int32(60),
				},
			}
		})

		ginkgo.It("should pass with valid health_check_grace_period_seconds", func() {
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with health_check_grace_period_seconds set to 120", func() {
			input.Spec.HealthCheckGracePeriodSeconds = proto.Int32(120)
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})
