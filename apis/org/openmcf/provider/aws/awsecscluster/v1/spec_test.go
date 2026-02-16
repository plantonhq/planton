package awsecsclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsEcsCluster(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsEcsCluster Suite")
}

var _ = ginkgo.Describe("AwsEcsCluster Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("minimal configuration", func() {
			var input *AwsEcsCluster

			ginkgo.BeforeEach(func() {
				input = &AwsEcsCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsEcsCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "valid-name",
					},
					Spec: &AwsEcsClusterSpec{
						EnableContainerInsights: true,
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with capacity providers", func() {
			var input *AwsEcsCluster

			ginkgo.BeforeEach(func() {
				input = &AwsEcsCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsEcsCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "valid-name",
					},
					Spec: &AwsEcsClusterSpec{
						EnableContainerInsights: true,
						CapacityProviders:       []string{"FARGATE", "FARGATE_SPOT"},
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with default capacity provider strategy", func() {
			var input *AwsEcsCluster

			ginkgo.BeforeEach(func() {
				input = &AwsEcsCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsEcsCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "production-cluster",
					},
					Spec: &AwsEcsClusterSpec{
						EnableContainerInsights: true,
						CapacityProviders:       []string{"FARGATE", "FARGATE_SPOT"},
						DefaultCapacityProviderStrategy: []*CapacityProviderStrategy{
							{
								CapacityProvider: "FARGATE",
								Base:             1,
								Weight:           1,
							},
							{
								CapacityProvider: "FARGATE_SPOT",
								Base:             0,
								Weight:           4,
							},
						},
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with exec configuration - DEFAULT", func() {
			var input *AwsEcsCluster

			ginkgo.BeforeEach(func() {
				input = &AwsEcsCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsEcsCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "debug-cluster",
					},
					Spec: &AwsEcsClusterSpec{
						EnableContainerInsights: true,
						ExecuteCommandConfiguration: &ExecConfiguration{
							Logging: ExecConfiguration_DEFAULT,
						},
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with exec configuration - OVERRIDE with CloudWatch", func() {
			var input *AwsEcsCluster

			ginkgo.BeforeEach(func() {
				input = &AwsEcsCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsEcsCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "audit-cluster",
					},
					Spec: &AwsEcsClusterSpec{
						EnableContainerInsights: true,
						ExecuteCommandConfiguration: &ExecConfiguration{
							Logging: ExecConfiguration_OVERRIDE,
							LogConfiguration: &ExecLogConfiguration{
								CloudWatchLogGroupName:      "/aws/ecs/exec-logs",
								CloudWatchEncryptionEnabled: true,
							},
							KmsKeyId: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
								},
							},
						},
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with exec configuration - OVERRIDE with S3", func() {
			var input *AwsEcsCluster

			ginkgo.BeforeEach(func() {
				input = &AwsEcsCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsEcsCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "compliance-cluster",
					},
					Spec: &AwsEcsClusterSpec{
						EnableContainerInsights: true,
						ExecuteCommandConfiguration: &ExecConfiguration{
							Logging: ExecConfiguration_OVERRIDE,
							LogConfiguration: &ExecLogConfiguration{
								S3BucketName:        "my-compliance-bucket",
								S3KeyPrefix:         "ecs-exec/",
								S3EncryptionEnabled: true,
							},
							KmsKeyId: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
								},
							},
						},
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("full production configuration", func() {
			var input *AwsEcsCluster

			ginkgo.BeforeEach(func() {
				input = &AwsEcsCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsEcsCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "production-full",
					},
					Spec: &AwsEcsClusterSpec{
						EnableContainerInsights: true,
						CapacityProviders:       []string{"FARGATE", "FARGATE_SPOT"},
						DefaultCapacityProviderStrategy: []*CapacityProviderStrategy{
							{
								CapacityProvider: "FARGATE",
								Base:             1,
								Weight:           1,
							},
							{
								CapacityProvider: "FARGATE_SPOT",
								Base:             0,
								Weight:           4,
							},
						},
						ExecuteCommandConfiguration: &ExecConfiguration{
							Logging: ExecConfiguration_OVERRIDE,
							LogConfiguration: &ExecLogConfiguration{
								CloudWatchLogGroupName:      "/aws/ecs/prod/exec",
								CloudWatchEncryptionEnabled: true,
								S3BucketName:                "prod-ecs-audit",
								S3KeyPrefix:                 "exec-logs/",
								S3EncryptionEnabled:         true,
							},
							KmsKeyId: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "arn:aws:kms:us-east-1:123456789012:key/prod-key",
								},
							},
						},
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("invalid capacity provider", func() {
			var input *AwsEcsCluster

			ginkgo.BeforeEach(func() {
				input = &AwsEcsCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsEcsCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "invalid-cluster",
					},
					Spec: &AwsEcsClusterSpec{
						CapacityProviders: []string{"INVALID_PROVIDER"},
					},
				}
			})

			ginkgo.It("should return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("duplicate capacity providers", func() {
			var input *AwsEcsCluster

			ginkgo.BeforeEach(func() {
				input = &AwsEcsCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsEcsCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "duplicate-cluster",
					},
					Spec: &AwsEcsClusterSpec{
						CapacityProviders: []string{"FARGATE", "FARGATE"},
					},
				}
			})

			ginkgo.It("should return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid capacity provider in strategy", func() {
			var input *AwsEcsCluster

			ginkgo.BeforeEach(func() {
				input = &AwsEcsCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsEcsCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "invalid-strategy",
					},
					Spec: &AwsEcsClusterSpec{
						DefaultCapacityProviderStrategy: []*CapacityProviderStrategy{
							{
								CapacityProvider: "INVALID",
								Base:             0,
								Weight:           1,
							},
						},
					},
				}
			})

			ginkgo.It("should return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("negative base in strategy", func() {
			var input *AwsEcsCluster

			ginkgo.BeforeEach(func() {
				input = &AwsEcsCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsEcsCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "negative-base",
					},
					Spec: &AwsEcsClusterSpec{
						DefaultCapacityProviderStrategy: []*CapacityProviderStrategy{
							{
								CapacityProvider: "FARGATE",
								Base:             -1,
								Weight:           1,
							},
						},
					},
				}
			})

			ginkgo.It("should return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("zero weight in strategy", func() {
			var input *AwsEcsCluster

			ginkgo.BeforeEach(func() {
				input = &AwsEcsCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsEcsCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "zero-weight",
					},
					Spec: &AwsEcsClusterSpec{
						DefaultCapacityProviderStrategy: []*CapacityProviderStrategy{
							{
								CapacityProvider: "FARGATE",
								Base:             0,
								Weight:           0,
							},
						},
					},
				}
			})

			ginkgo.It("should return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
