package awscodebuildprojectv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsCodeBuildProjectSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsCodeBuildProjectSpec Validation Suite")
}

func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func svRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

func minimalGitHubSpec() *AwsCodeBuildProjectSpec {
	return &AwsCodeBuildProjectSpec{
		Source: &AwsCodeBuildSource{
			Type:     "GITHUB",
			Location: "https://github.com/example/repo.git",
		},
		Environment: &AwsCodeBuildEnvironment{
			Type:        "LINUX_CONTAINER",
			ComputeType: "BUILD_GENERAL1_SMALL",
			Image:       "aws/codebuild/amazonlinux2-x86_64-standard:5.0",
		},
		Artifacts: &AwsCodeBuildArtifacts{
			Type: "NO_ARTIFACTS",
		},
		ServiceRole: svRef("arn:aws:iam::123456789012:role/codebuild-role"),
	}
}

func minimalCodePipelineSpec() *AwsCodeBuildProjectSpec {
	return &AwsCodeBuildProjectSpec{
		Source: &AwsCodeBuildSource{
			Type: "CODEPIPELINE",
		},
		Environment: &AwsCodeBuildEnvironment{
			Type:        "LINUX_CONTAINER",
			ComputeType: "BUILD_GENERAL1_MEDIUM",
			Image:       "aws/codebuild/amazonlinux2-x86_64-standard:5.0",
		},
		Artifacts: &AwsCodeBuildArtifacts{
			Type: "CODEPIPELINE",
		},
		ServiceRole: svRef("arn:aws:iam::123456789012:role/codebuild-role"),
	}
}

var _ = ginkgo.Describe("AwsCodeBuildProjectSpec validations", func() {

	// =========================================================================
	// Valid configurations
	// =========================================================================

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.Context("with minimal GitHub configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(minimalGitHubSpec())
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with minimal CodePipeline configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(minimalCodePipelineSpec())
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with NO_SOURCE and inline buildspec", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Source = &AwsCodeBuildSource{
					Type:      "NO_SOURCE",
					Buildspec: "version: 0.2\nphases:\n  build:\n    commands:\n      - echo hello",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with S3 source", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Source = &AwsCodeBuildSource{
					Type:     "S3",
					Location: "my-bucket/source.zip",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with S3 artifacts", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Artifacts = &AwsCodeBuildArtifacts{
					Type:     "S3",
					Location: svRef("my-artifacts-bucket"),
					Name:     "build-output",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with all optional fields populated", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsCodeBuildProjectSpec{
					Source: &AwsCodeBuildSource{
						Type:              "GITHUB",
						Location:          "https://github.com/example/repo.git",
						Buildspec:         "buildspec.yml",
						GitCloneDepth:     1,
						ReportBuildStatus: true,
						FetchSubmodules:   true,
					},
					Environment: &AwsCodeBuildEnvironment{
						Type:                     "LINUX_CONTAINER",
						ComputeType:              "BUILD_GENERAL1_LARGE",
						Image:                    "aws/codebuild/amazonlinux2-x86_64-standard:5.0",
						PrivilegedMode:           true,
						ImagePullCredentialsType: stringPtr("CODEBUILD"),
						EnvironmentVariables: []*AwsCodeBuildEnvironmentVariable{
							{Name: "ENV", Value: "production", Type: stringPtr("PLAINTEXT")},
							{Name: "DB_PASSWORD", Value: "my-secret", Type: stringPtr("SECRETS_MANAGER")},
						},
					},
					Artifacts: &AwsCodeBuildArtifacts{
						Type:          "S3",
						Location:      svRef("my-bucket"),
						Name:          "output",
						Path:          "builds",
						Packaging:     "ZIP",
						NamespaceType: "BUILD_ID",
					},
					ServiceRole:          svRef("arn:aws:iam::123456789012:role/codebuild-role"),
					Description:          "Production build project",
					EncryptionKey:        svRef("arn:aws:kms:us-east-1:123456789012:key/example"),
					BuildTimeout:         int32Ptr(120),
					QueuedTimeout:        int32Ptr(240),
					ConcurrentBuildLimit: 5,
					SourceVersion:        "main",
					Cache: &AwsCodeBuildCache{
						Type:     stringPtr("S3"),
						Location: svRef("my-cache-bucket/prefix"),
					},
					LogsConfig: &AwsCodeBuildLogsConfig{
						CloudwatchLogs: &AwsCodeBuildCloudWatchLogs{
							Status:     stringPtr("ENABLED"),
							GroupName:  svRef("/aws/codebuild/my-project"),
							StreamName: "build",
						},
					},
					VpcConfig: &AwsCodeBuildVpcConfig{
						VpcId:            svRef("vpc-abc123"),
						SubnetIds:        []*foreignkeyv1.StringValueOrRef{svRef("subnet-aaa"), svRef("subnet-bbb")},
						SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{svRef("sg-111")},
					},
					Webhook: &AwsCodeBuildWebhook{
						BuildType: "BUILD",
						FilterGroups: []*AwsCodeBuildWebhookFilterGroup{
							{
								Filters: []*AwsCodeBuildWebhookFilter{
									{Type: "EVENT", Pattern: "PUSH, PULL_REQUEST_CREATED"},
									{Type: "HEAD_REF", Pattern: "^refs/heads/main$"},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with Lambda environment type", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Environment = &AwsCodeBuildEnvironment{
					Type:        "LINUX_LAMBDA_CONTAINER",
					ComputeType: "BUILD_LAMBDA_4GB",
					Image:       "aws/codebuild/amazonlinux-aarch64-lambda-standard:go1.21",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with ARM container", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Environment = &AwsCodeBuildEnvironment{
					Type:        "ARM_CONTAINER",
					ComputeType: "BUILD_GENERAL1_LARGE",
					Image:       "aws/codebuild/amazonlinux2-aarch64-standard:3.0",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with LOCAL cache using Docker layer mode", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Cache = &AwsCodeBuildCache{
					Type:  stringPtr("LOCAL"),
					Modes: []string{"LOCAL_DOCKER_LAYER_CACHE", "LOCAL_SOURCE_CACHE"},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with webhook excluding a branch", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Webhook = &AwsCodeBuildWebhook{
					FilterGroups: []*AwsCodeBuildWebhookFilterGroup{
						{
							Filters: []*AwsCodeBuildWebhookFilter{
								{Type: "EVENT", Pattern: "PUSH"},
								{Type: "HEAD_REF", Pattern: "^refs/heads/release/.*$", ExcludeMatchedPattern: true},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	// =========================================================================
	// Required field validations
	// =========================================================================

	ginkgo.Describe("When required fields are missing", func() {

		ginkgo.Context("with missing source", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Source = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing environment", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Environment = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing artifacts", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Artifacts = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing service_role", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.ServiceRole = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing source.type", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Source.Type = ""
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing environment.image", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Environment.Image = ""
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing webhook filter in filter group", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Webhook = &AwsCodeBuildWebhook{
					FilterGroups: []*AwsCodeBuildWebhookFilterGroup{
						{Filters: []*AwsCodeBuildWebhookFilter{}},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	// =========================================================================
	// Enum / string-in validations
	// =========================================================================

	ginkgo.Describe("When invalid enum values are passed", func() {

		ginkgo.Context("with invalid source.type", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Source.Type = "INVALID_SOURCE"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid environment.type", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Environment.Type = "INVALID_ENV"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid environment.compute_type", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Environment.ComputeType = "BUILD_MEGA_XLARGE"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid artifacts.type", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Artifacts.Type = "INVALID_ARTIFACTS"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid artifacts.packaging", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Artifacts.Packaging = "TAR"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid cache.type", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Cache = &AwsCodeBuildCache{
					Type: stringPtr("REDIS"),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid webhook filter type", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Webhook = &AwsCodeBuildWebhook{
					FilterGroups: []*AwsCodeBuildWebhookFilterGroup{
						{
							Filters: []*AwsCodeBuildWebhookFilter{
								{Type: "INVALID_FILTER", Pattern: "test"},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	// =========================================================================
	// Range validations
	// =========================================================================

	ginkgo.Describe("When values are out of range", func() {

		ginkgo.Context("with build_timeout below minimum", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.BuildTimeout = int32Ptr(3)
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with build_timeout above maximum", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.BuildTimeout = int32Ptr(3000)
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with queued_timeout below minimum", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.QueuedTimeout = int32Ptr(2)
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with queued_timeout above maximum", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.QueuedTimeout = int32Ptr(600)
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with description exceeding max length", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				longDesc := ""
				for i := 0; i < 260; i++ {
					longDesc += "x"
				}
				spec.Description = longDesc
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	// =========================================================================
	// Cross-field (CEL) validations
	// =========================================================================

	ginkgo.Describe("When cross-field validations fail", func() {

		ginkgo.Context("with CODEPIPELINE source but non-CODEPIPELINE artifacts", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalCodePipelineSpec()
				spec.Artifacts = &AwsCodeBuildArtifacts{Type: "NO_ARTIFACTS"}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with non-CODEPIPELINE source but CODEPIPELINE artifacts", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Artifacts = &AwsCodeBuildArtifacts{Type: "CODEPIPELINE"}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with GITHUB source but missing location", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Source.Location = ""
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with BITBUCKET source but missing location", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Source.Type = "BITBUCKET"
				spec.Source.Location = ""
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with NO_SOURCE but missing buildspec", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Source = &AwsCodeBuildSource{
					Type: "NO_SOURCE",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with S3 artifacts but missing location", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Artifacts = &AwsCodeBuildArtifacts{Type: "S3"}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with webhook on CODEPIPELINE source", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalCodePipelineSpec()
				spec.Webhook = &AwsCodeBuildWebhook{
					FilterGroups: []*AwsCodeBuildWebhookFilterGroup{
						{
							Filters: []*AwsCodeBuildWebhookFilter{
								{Type: "EVENT", Pattern: "PUSH"},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with webhook on NO_SOURCE", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Source = &AwsCodeBuildSource{
					Type:      "NO_SOURCE",
					Buildspec: "version: 0.2",
				}
				spec.Webhook = &AwsCodeBuildWebhook{
					FilterGroups: []*AwsCodeBuildWebhookFilterGroup{
						{
							Filters: []*AwsCodeBuildWebhookFilter{
								{Type: "EVENT", Pattern: "PUSH"},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with webhook on S3 source", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.Source = &AwsCodeBuildSource{
					Type:     "S3",
					Location: "my-bucket/source.zip",
				}
				spec.Webhook = &AwsCodeBuildWebhook{
					FilterGroups: []*AwsCodeBuildWebhookFilterGroup{
						{
							Filters: []*AwsCodeBuildWebhookFilter{
								{Type: "EVENT", Pattern: "PUSH"},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	// =========================================================================
	// VPC config validations
	// =========================================================================

	ginkgo.Describe("When VPC config is incomplete", func() {

		ginkgo.Context("with vpc_config missing subnet_ids", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.VpcConfig = &AwsCodeBuildVpcConfig{
					VpcId:            svRef("vpc-abc123"),
					SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{svRef("sg-111")},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with vpc_config missing security_group_ids", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				spec.VpcConfig = &AwsCodeBuildVpcConfig{
					VpcId:     svRef("vpc-abc123"),
					SubnetIds: []*foreignkeyv1.StringValueOrRef{svRef("subnet-aaa")},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with vpc_config exceeding max subnets", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				var subnets []*foreignkeyv1.StringValueOrRef
				for i := 0; i < 17; i++ {
					subnets = append(subnets, svRef("subnet-xxx"))
				}
				spec.VpcConfig = &AwsCodeBuildVpcConfig{
					VpcId:            svRef("vpc-abc123"),
					SubnetIds:        subnets,
					SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{svRef("sg-111")},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with vpc_config exceeding max security groups", func() {
			ginkgo.It("should return a validation error", func() {
				spec := minimalGitHubSpec()
				var sgs []*foreignkeyv1.StringValueOrRef
				for i := 0; i < 6; i++ {
					sgs = append(sgs, svRef("sg-xxx"))
				}
				spec.VpcConfig = &AwsCodeBuildVpcConfig{
					VpcId:            svRef("vpc-abc123"),
					SubnetIds:        []*foreignkeyv1.StringValueOrRef{svRef("subnet-aaa")},
					SecurityGroupIds: sgs,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
