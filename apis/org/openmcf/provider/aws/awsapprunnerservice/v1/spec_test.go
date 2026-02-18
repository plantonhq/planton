package awsapprunnerservicev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsAppRunnerServiceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsAppRunnerService Validation Tests")
}

func strPtr(s string) *string { return &s }
func int32Ptr(i int32) *int32 { return &i }
func boolPtr(b bool) *bool    { return &b }

func stringValueOrRefLiteral(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

func stringValueOrRefFrom(kind cloudresourcekind.CloudResourceKind, env, name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind: kind,
				Env:  env,
				Name: name,
			},
		},
	}
}

func validEnvelope(spec *AwsAppRunnerServiceSpec) *AwsAppRunnerService {
	return &AwsAppRunnerService{
		ApiVersion: "aws.openmcf.org/v1",
		Kind:       "AwsAppRunnerService",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-app-runner-svc"},
		Spec:       spec,
	}
}

func minimalImageSpec() *AwsAppRunnerServiceSpec {
	return &AwsAppRunnerServiceSpec{
		Region: "us-west-2",
		ImageSource: &AwsAppRunnerServiceImageSource{
			ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
			ImageRepositoryType: "ECR_PUBLIC",
		},
	}
}

var _ = ginkgo.Describe("AwsAppRunnerService Validation Tests", func() {

	// =====================================================================
	// HAPPY PATH TESTS
	// =====================================================================
	ginkgo.Describe("Happy Path", func() {

		ginkgo.It("should accept minimal valid spec with ECR_PUBLIC image source", func() {
			input := validEnvelope(minimalImageSpec())
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept image source with private ECR and access_role_arn", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app:v1.0",
					ImageRepositoryType: "ECR",
					AccessRoleArn:       stringValueOrRefLiteral("arn:aws:iam::123456789012:role/apprunner-ecr-access"),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept image source with custom port, start_command, and env vars", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				Port:         strPtr("3000"),
				StartCommand: "node server.js",
				EnvironmentVariables: map[string]string{
					"NODE_ENV":  "production",
					"LOG_LEVEL": "info",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept image source with environment secrets", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				EnvironmentSecrets: map[string]string{
					"DB_PASSWORD": "arn:aws:secretsmanager:us-east-1:123456789012:secret:db-pass-abc123",
					"API_KEY":     "arn:aws:ssm:us-east-1:123456789012:parameter/api-key",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept code source with API configuration", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				CodeSource: &AwsAppRunnerServiceCodeSource{
					RepositoryUrl:       "https://github.com/my-org/my-app",
					Branch:              "main",
					ConnectionArn:       stringValueOrRefLiteral("arn:aws:apprunner:us-east-1:123456789012:connection/my-conn/abc123"),
					ConfigurationSource: "API",
					Runtime:             "NODEJS_18",
					BuildCommand:        "npm ci && npm run build",
				},
				StartCommand: "npm start",
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept code source with REPOSITORY configuration", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				CodeSource: &AwsAppRunnerServiceCodeSource{
					RepositoryUrl:       "https://github.com/my-org/my-app",
					Branch:              "production",
					SourceDirectory:     "services/api",
					ConnectionArn:       stringValueOrRefLiteral("arn:aws:apprunner:us-east-1:123456789012:connection/my-conn/abc123"),
					ConfigurationSource: "REPOSITORY",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept full production spec with image source + VPC + KMS + auto scaling + health check", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "123456789012.dkr.ecr.us-east-1.amazonaws.com/prod-app:v2.1",
					ImageRepositoryType: "ECR",
					AccessRoleArn:       stringValueOrRefLiteral("arn:aws:iam::123456789012:role/apprunner-ecr-access"),
				},
				Port:            strPtr("8080"),
				StartCommand:    "./start.sh",
				Cpu:             strPtr("2048"),
				Memory:          strPtr("4096"),
				InstanceRoleArn: stringValueOrRefLiteral("arn:aws:iam::123456789012:role/apprunner-instance"),
				HealthCheck: &AwsAppRunnerServiceHealthCheck{
					Protocol:           strPtr("HTTP"),
					Path:               strPtr("/health"),
					IntervalSeconds:    int32Ptr(10),
					TimeoutSeconds:     int32Ptr(5),
					HealthyThreshold:   int32Ptr(3),
					UnhealthyThreshold: int32Ptr(5),
				},
				AutoScaling: &AwsAppRunnerServiceAutoScaling{
					MinSize:        int32Ptr(2),
					MaxSize:        int32Ptr(10),
					MaxConcurrency: int32Ptr(50),
				},
				SubnetIds: []*foreignkeyv1.StringValueOrRef{
					stringValueOrRefLiteral("subnet-0abc1234"),
					stringValueOrRefLiteral("subnet-0def5678"),
				},
				SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{
					stringValueOrRefLiteral("sg-0abc1234"),
				},
				KmsKeyArn:                     stringValueOrRefLiteral("arn:aws:kms:us-east-1:123456789012:key/my-key"),
				ObservabilityEnabled:          true,
				ObservabilityConfigurationArn: stringValueOrRefLiteral("arn:aws:apprunner:us-east-1:123456789012:observabilityconfiguration/my-config/1"),
				IsPubliclyAccessible:          boolPtr(true),
				IpAddressType:                 strPtr("IPV4"),
				AutoDeploymentsEnabled:        boolPtr(true),
				EnvironmentVariables: map[string]string{
					"APP_ENV": "production",
				},
				EnvironmentSecrets: map[string]string{
					"DB_URL": "arn:aws:secretsmanager:us-east-1:123456789012:secret:db-url",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept VPC egress with inline subnet_ids and security_group_ids", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				SubnetIds: []*foreignkeyv1.StringValueOrRef{
					stringValueOrRefLiteral("subnet-0abc1234"),
					stringValueOrRefLiteral("subnet-0def5678"),
				},
				SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{
					stringValueOrRefLiteral("sg-0abc1234"),
					stringValueOrRefLiteral("sg-0def5678"),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept VPC egress with vpc_connector_arn only", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				VpcConnectorArn: stringValueOrRefLiteral("arn:aws:apprunner:us-east-1:123456789012:vpcconnector/my-vpc-connector/1/abc123"),
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept auto scaling with custom values", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				AutoScaling: &AwsAppRunnerServiceAutoScaling{
					MinSize:        int32Ptr(3),
					MaxSize:        int32Ptr(20),
					MaxConcurrency: int32Ptr(150),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept health check with HTTP protocol", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				HealthCheck: &AwsAppRunnerServiceHealthCheck{
					Protocol:        strPtr("HTTP"),
					Path:            strPtr("/readyz"),
					IntervalSeconds: int32Ptr(10),
					TimeoutSeconds:  int32Ptr(3),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept health check with TCP protocol (default)", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				HealthCheck: &AwsAppRunnerServiceHealthCheck{
					Protocol:        strPtr("TCP"),
					IntervalSeconds: int32Ptr(5),
					TimeoutSeconds:  int32Ptr(2),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept observability enabled with configuration ARN", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				ObservabilityEnabled:          true,
				ObservabilityConfigurationArn: stringValueOrRefLiteral("arn:aws:apprunner:us-east-1:123456789012:observabilityconfiguration/my-config/1"),
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept DUAL_STACK ip address type", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				IpAddressType: strPtr("DUAL_STACK"),
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept auto_deployments_enabled set to false", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				AutoDeploymentsEnabled: boolPtr(false),
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept human-readable CPU format '1 vCPU'", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				Cpu: strPtr("1 vCPU"),
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept human-readable memory format '2 GB'", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				Memory: strPtr("2 GB"),
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept StringValueOrRef with valueFrom references", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app:latest",
					ImageRepositoryType: "ECR",
					AccessRoleArn: stringValueOrRefFrom(
						cloudresourcekind.CloudResourceKind_AwsIamRole,
						"production",
						"ecr-access-role",
					),
				},
				InstanceRoleArn: stringValueOrRefFrom(
					cloudresourcekind.CloudResourceKind_AwsIamRole,
					"production",
					"instance-role",
				),
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept all valid numeric CPU values", func() {
			validCPUs := []string{"256", "512", "1024", "2048", "4096"}
			for _, cpu := range validCPUs {
				spec := minimalImageSpec()
				spec.Cpu = strPtr(cpu)
				input := validEnvelope(spec)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil(), "expected CPU=%s to be valid", cpu)
			}
		})

		ginkgo.It("should accept all valid human-readable CPU values", func() {
			validCPUs := []string{"0.25 vCPU", "0.5 vCPU", "1 vCPU", "2 vCPU", "4 vCPU"}
			for _, cpu := range validCPUs {
				spec := minimalImageSpec()
				spec.Cpu = strPtr(cpu)
				input := validEnvelope(spec)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil(), "expected CPU=%s to be valid", cpu)
			}
		})

		ginkgo.It("should accept all valid numeric memory values", func() {
			validMems := []string{"512", "1024", "2048", "3072", "4096", "6144", "8192", "10240", "12288"}
			for _, mem := range validMems {
				spec := minimalImageSpec()
				spec.Memory = strPtr(mem)
				input := validEnvelope(spec)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil(), "expected Memory=%s to be valid", mem)
			}
		})

		ginkgo.It("should accept is_publicly_accessible set to false", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				IsPubliclyAccessible: boolPtr(false),
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// =====================================================================
	// FAILURE TESTS
	// =====================================================================
	ginkgo.Describe("Failure Cases", func() {

		ginkgo.It("should fail when neither image_source nor code_source is set", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when both image_source and code_source are set", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				CodeSource: &AwsAppRunnerServiceCodeSource{
					RepositoryUrl:       "https://github.com/my-org/my-app",
					Branch:              "main",
					ConnectionArn:       stringValueOrRefLiteral("arn:aws:apprunner:us-east-1:123456789012:connection/conn/1"),
					ConfigurationSource: "REPOSITORY",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when image_repository_type is invalid", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "DOCKER_HUB",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ECR is used without access_role_arn", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app:latest",
					ImageRepositoryType: "ECR",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when code source is missing connection_arn", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				CodeSource: &AwsAppRunnerServiceCodeSource{
					RepositoryUrl:       "https://github.com/my-org/my-app",
					Branch:              "main",
					ConfigurationSource: "REPOSITORY",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when configuration_source is invalid", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				CodeSource: &AwsAppRunnerServiceCodeSource{
					RepositoryUrl:       "https://github.com/my-org/my-app",
					Branch:              "main",
					ConnectionArn:       stringValueOrRefLiteral("arn:aws:apprunner:us-east-1:123456789012:connection/conn/1"),
					ConfigurationSource: "INVALID",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when API configuration is used without runtime", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				CodeSource: &AwsAppRunnerServiceCodeSource{
					RepositoryUrl:       "https://github.com/my-org/my-app",
					Branch:              "main",
					ConnectionArn:       stringValueOrRefLiteral("arn:aws:apprunner:us-east-1:123456789012:connection/conn/1"),
					ConfigurationSource: "API",
					Runtime:             "",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when CPU value is invalid", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				Cpu: strPtr("999"),
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when memory value is invalid", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				Memory: strPtr("777"),
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when health check protocol is invalid", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				HealthCheck: &AwsAppRunnerServiceHealthCheck{
					Protocol: strPtr("GRPC"),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when health check interval_seconds is 0 (below range)", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				HealthCheck: &AwsAppRunnerServiceHealthCheck{
					IntervalSeconds: int32Ptr(0),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when health check interval_seconds is 21 (above range)", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				HealthCheck: &AwsAppRunnerServiceHealthCheck{
					IntervalSeconds: int32Ptr(21),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when health check timeout_seconds is out of range", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				HealthCheck: &AwsAppRunnerServiceHealthCheck{
					TimeoutSeconds: int32Ptr(25),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when auto scaling min_size > max_size", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				AutoScaling: &AwsAppRunnerServiceAutoScaling{
					MinSize: int32Ptr(10),
					MaxSize: int32Ptr(5),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when auto scaling max_concurrency is 0 (below range)", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				AutoScaling: &AwsAppRunnerServiceAutoScaling{
					MaxConcurrency: int32Ptr(0),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when auto scaling max_concurrency is 201 (above range)", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				AutoScaling: &AwsAppRunnerServiceAutoScaling{
					MaxConcurrency: int32Ptr(201),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ip_address_type is invalid", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				IpAddressType: strPtr("IPV6"),
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when both vpc_connector_arn and subnet_ids are provided", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				VpcConnectorArn: stringValueOrRefLiteral("arn:aws:apprunner:us-east-1:123456789012:vpcconnector/my-vpc/1/abc"),
				SubnetIds: []*foreignkeyv1.StringValueOrRef{
					stringValueOrRefLiteral("subnet-0abc1234"),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when observability_enabled is true without configuration ARN", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				ObservabilityEnabled: true,
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when image_identifier is empty", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "",
					ImageRepositoryType: "ECR_PUBLIC",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when image_repository_type is empty", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when auto scaling min_size is 0 (below range)", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				AutoScaling: &AwsAppRunnerServiceAutoScaling{
					MinSize: int32Ptr(0),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when auto scaling max_size is 26 (above range)", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				AutoScaling: &AwsAppRunnerServiceAutoScaling{
					MaxSize: int32Ptr(26),
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when code source repository_url is empty", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				CodeSource: &AwsAppRunnerServiceCodeSource{
					RepositoryUrl:       "",
					Branch:              "main",
					ConnectionArn:       stringValueOrRefLiteral("arn:aws:apprunner:us-east-1:123456789012:connection/conn/1"),
					ConfigurationSource: "REPOSITORY",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when code source branch is empty", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				CodeSource: &AwsAppRunnerServiceCodeSource{
					RepositoryUrl:       "https://github.com/my-org/my-app",
					Branch:              "",
					ConnectionArn:       stringValueOrRefLiteral("arn:aws:apprunner:us-east-1:123456789012:connection/conn/1"),
					ConfigurationSource: "REPOSITORY",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	// =====================================================================
	// API ENVELOPE TESTS
	// =====================================================================
	ginkgo.Describe("API Envelope", func() {

		ginkgo.It("should fail when apiVersion is wrong", func() {
			input := &AwsAppRunnerService{
				ApiVersion: "gcp.openmcf.org/v1",
				Kind:       "AwsAppRunnerService",
				Metadata:   &shared.CloudResourceMetadata{Name: "test-svc"},
				Spec:       minimalImageSpec(),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AwsAppRunnerService{
				ApiVersion: "aws.openmcf.org/v1",
				Kind:       "AwsLambda",
				Metadata:   &shared.CloudResourceMetadata{Name: "test-svc"},
				Spec:       minimalImageSpec(),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AwsAppRunnerService{
				ApiVersion: "aws.openmcf.org/v1",
				Kind:       "AwsAppRunnerService",
				Spec:       minimalImageSpec(),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AwsAppRunnerService{
				ApiVersion: "aws.openmcf.org/v1",
				Kind:       "AwsAppRunnerService",
				Metadata:   &shared.CloudResourceMetadata{Name: "test-svc"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept a valid complete envelope", func() {
			input := validEnvelope(minimalImageSpec())
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})
