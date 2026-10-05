package awsapprunnerservicev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
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

func stringValueOrRefFrom(kind catalogkind.CatalogKind, env, name string) *foreignkeyv1.StringValueOrRef {
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
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsAppRunnerService",
		Metadata:   &shared.CatalogObjectMetadata{Name: "test-app-runner-svc"},
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

		ginkgo.It("should accept auto deployments for private ECR", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app:latest",
					ImageRepositoryType: "ECR",
					AccessRoleArn:       stringValueOrRefLiteral("arn:aws:iam::123456789012:role/apprunner-ecr-access"),
				},
				AutoDeploymentsEnabled: true,
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
					Runtime:             "NODEJS_22",
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

		ginkgo.It("should accept every managed runtime App Runner supports", func() {
			runtimes := []string{
				"PYTHON_3", "PYTHON_311",
				"NODEJS_12", "NODEJS_14", "NODEJS_16", "NODEJS_18", "NODEJS_22",
				"CORRETTO_8", "CORRETTO_11",
				"GO_1", "DOTNET_6", "PHP_81", "RUBY_31",
			}
			for _, rt := range runtimes {
				input := validEnvelope(&AwsAppRunnerServiceSpec{
					Region: "us-west-2",
					CodeSource: &AwsAppRunnerServiceCodeSource{
						RepositoryUrl:       "https://github.com/my-org/my-app",
						Branch:              "main",
						ConnectionArn:       stringValueOrRefLiteral("arn:aws:apprunner:us-east-1:123456789012:connection/conn/1"),
						ConfigurationSource: "API",
						Runtime:             rt,
					},
				})
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil(), "expected runtime=%s to be valid", rt)
			}
		})

		ginkgo.It("should accept full production spec with companion refs + KMS + health check + WAF", func() {
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
					Interval:           int32Ptr(10),
					Timeout:            int32Ptr(5),
					HealthyThreshold:   int32Ptr(3),
					UnhealthyThreshold: int32Ptr(5),
				},
				AutoScalingConfigurationArn:   stringValueOrRefLiteral("arn:aws:apprunner:us-west-2:123456789012:autoscalingconfiguration/prod-asc/2/abc"),
				VpcConnectorArn:               stringValueOrRefLiteral("arn:aws:apprunner:us-west-2:123456789012:vpcconnector/prod-vc/1/abc"),
				ObservabilityConfigurationArn: stringValueOrRefLiteral("arn:aws:apprunner:us-west-2:123456789012:observabilityconfiguration/prod-oc/1/abc"),
				KmsKeyArn:                     stringValueOrRefLiteral("arn:aws:kms:us-west-2:123456789012:key/my-key"),
				IsPubliclyAccessible:          boolPtr(true),
				IpAddressType:                 strPtr("IPV4"),
				AutoDeploymentsEnabled:        true,
				WebAclArn:                     stringValueOrRefLiteral("arn:aws:wafv2:us-west-2:123456789012:regional/webacl/prod-acl/abc"),
				CustomDomains: []*AwsAppRunnerServiceCustomDomain{
					{DomainName: "api.example.com"},
				},
				EnvironmentVariables: map[string]string{
					"APP_ENV": "production",
				},
				EnvironmentSecrets: map[string]string{
					"DB_URL": "arn:aws:secretsmanager:us-west-2:123456789012:secret:db-url",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept custom domains with www subdomain disabled", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				CustomDomains: []*AwsAppRunnerServiceCustomDomain{
					{DomainName: "app.example.com", EnableWwwSubdomain: boolPtr(false)},
					{DomainName: "example.com"},
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
					Protocol: strPtr("HTTP"),
					Path:     strPtr("/readyz"),
					Interval: int32Ptr(10),
					Timeout:  int32Ptr(3),
				},
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

		ginkgo.It("should accept a private service (is_publicly_accessible false)", func() {
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

		ginkgo.It("should accept StringValueOrRef with valueFrom references", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app:latest",
					ImageRepositoryType: "ECR",
					AccessRoleArn: stringValueOrRefFrom(
						catalogkind.CatalogKind_AwsIamRole,
						"production",
						"ecr-access-role",
					),
				},
				InstanceRoleArn: stringValueOrRefFrom(
					catalogkind.CatalogKind_AwsIamRole,
					"production",
					"instance-role",
				),
				AutoScalingConfigurationArn: stringValueOrRefFrom(
					catalogkind.CatalogKind_AwsAppRunnerAutoScalingConfiguration,
					"production",
					"prod-scaling",
				),
				VpcConnectorArn: stringValueOrRefFrom(
					catalogkind.CatalogKind_AwsAppRunnerVpcConnector,
					"production",
					"prod-connector",
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
	})

	// =====================================================================
	// FAILURE TESTS
	// =====================================================================
	ginkgo.Describe("Failure Cases", func() {

		ginkgo.It("should fail when neither image_source nor code_source is set", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
			})
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

		ginkgo.It("should fail when auto deployments are enabled for ECR_PUBLIC", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
				AutoDeploymentsEnabled: true,
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

		ginkgo.It("should fail when runtime is not a managed runtime", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				CodeSource: &AwsAppRunnerServiceCodeSource{
					RepositoryUrl:       "https://github.com/my-org/my-app",
					Branch:              "main",
					ConnectionArn:       stringValueOrRefLiteral("arn:aws:apprunner:us-east-1:123456789012:connection/conn/1"),
					ConfigurationSource: "API",
					Runtime:             "RUST_1",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when CPU value is invalid", func() {
			spec := minimalImageSpec()
			spec.Cpu = strPtr("999")
			err := protovalidate.Validate(validEnvelope(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when memory value is invalid", func() {
			spec := minimalImageSpec()
			spec.Memory = strPtr("777")
			err := protovalidate.Validate(validEnvelope(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when health check protocol is invalid", func() {
			spec := minimalImageSpec()
			spec.HealthCheck = &AwsAppRunnerServiceHealthCheck{
				Protocol: strPtr("GRPC"),
			}
			err := protovalidate.Validate(validEnvelope(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when health check interval is 0 (below range)", func() {
			spec := minimalImageSpec()
			spec.HealthCheck = &AwsAppRunnerServiceHealthCheck{
				Interval: int32Ptr(0),
			}
			err := protovalidate.Validate(validEnvelope(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when health check interval is 21 (above range)", func() {
			spec := minimalImageSpec()
			spec.HealthCheck = &AwsAppRunnerServiceHealthCheck{
				Interval: int32Ptr(21),
			}
			err := protovalidate.Validate(validEnvelope(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when health check timeout is out of range", func() {
			spec := minimalImageSpec()
			spec.HealthCheck = &AwsAppRunnerServiceHealthCheck{
				Timeout: int32Ptr(25),
			}
			err := protovalidate.Validate(validEnvelope(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ip_address_type is invalid", func() {
			spec := minimalImageSpec()
			spec.IpAddressType = strPtr("IPV6")
			err := protovalidate.Validate(validEnvelope(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when a custom domain name is empty", func() {
			spec := minimalImageSpec()
			spec.CustomDomains = []*AwsAppRunnerServiceCustomDomain{
				{DomainName: ""},
			}
			err := protovalidate.Validate(validEnvelope(spec))
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

		ginkgo.It("should fail when region is empty", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "public.ecr.aws/nginx/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
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
	// VPC INGRESS CONNECTIONS
	// =====================================================================
	ginkgo.Describe("VPC Ingress Connections", func() {

		ginkgo.It("should accept an ingress connection with literal VPC and endpoint ids", func() {
			spec := minimalImageSpec()
			spec.IsPubliclyAccessible = boolPtr(false)
			spec.VpcIngressConnections = []*AwsAppRunnerServiceVpcIngressConnection{
				{
					Name:          "my-svc-private",
					VpcId:         stringValueOrRefLiteral("vpc-0abc123def456"),
					VpcEndpointId: stringValueOrRefLiteral("vpce-0abc123def456"),
				},
			}
			err := protovalidate.Validate(validEnvelope(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an ingress connection wired by reference", func() {
			spec := minimalImageSpec()
			spec.VpcIngressConnections = []*AwsAppRunnerServiceVpcIngressConnection{
				{
					Name:          "my-svc-private",
					VpcId:         stringValueOrRefFrom(catalogkind.CatalogKind_AwsVpc, "dev", "main-vpc"),
					VpcEndpointId: stringValueOrRefFrom(catalogkind.CatalogKind_AwsVpcEndpoint, "dev", "apprunner-endpoint"),
				},
			}
			err := protovalidate.Validate(validEnvelope(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail when the connection name is shorter than 4 characters", func() {
			spec := minimalImageSpec()
			spec.VpcIngressConnections = []*AwsAppRunnerServiceVpcIngressConnection{
				{
					Name:          "abc",
					VpcId:         stringValueOrRefLiteral("vpc-0abc123def456"),
					VpcEndpointId: stringValueOrRefLiteral("vpce-0abc123def456"),
				},
			}
			err := protovalidate.Validate(validEnvelope(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when the connection name exceeds 40 characters", func() {
			spec := minimalImageSpec()
			spec.VpcIngressConnections = []*AwsAppRunnerServiceVpcIngressConnection{
				{
					Name:          "this-connection-name-is-way-too-long-for-aws",
					VpcId:         stringValueOrRefLiteral("vpc-0abc123def456"),
					VpcEndpointId: stringValueOrRefLiteral("vpce-0abc123def456"),
				},
			}
			err := protovalidate.Validate(validEnvelope(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_id is missing", func() {
			spec := minimalImageSpec()
			spec.VpcIngressConnections = []*AwsAppRunnerServiceVpcIngressConnection{
				{
					Name:          "my-svc-private",
					VpcEndpointId: stringValueOrRefLiteral("vpce-0abc123def456"),
				},
			}
			err := protovalidate.Validate(validEnvelope(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_endpoint_id is missing", func() {
			spec := minimalImageSpec()
			spec.VpcIngressConnections = []*AwsAppRunnerServiceVpcIngressConnection{
				{
					Name:  "my-svc-private",
					VpcId: stringValueOrRefLiteral("vpc-0abc123def456"),
				},
			}
			err := protovalidate.Validate(validEnvelope(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	// =====================================================================
	// IMAGE IDENTIFIER FORMAT
	// =====================================================================
	ginkgo.Describe("Image identifier format", func() {

		ginkgo.It("should accept a private ECR image path", func() {
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

		ginkgo.It("should accept an ECR Public Gallery image path", func() {
			input := validEnvelope(minimalImageSpec())
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail for a bare Docker Hub image name", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
				},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail for a non-ECR registry host", func() {
			input := validEnvelope(&AwsAppRunnerServiceSpec{
				Region: "us-west-2",
				ImageSource: &AwsAppRunnerServiceImageSource{
					ImageIdentifier:     "docker.io/library/nginx:latest",
					ImageRepositoryType: "ECR_PUBLIC",
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
				ApiVersion: "gcp.planton.dev/v1alpha1",
				Kind:       "AwsAppRunnerService",
				Metadata:   &shared.CatalogObjectMetadata{Name: "test-svc"},
				Spec:       minimalImageSpec(),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AwsAppRunnerService{
				ApiVersion: "aws.planton.dev/v1alpha1",
				Kind:       "AwsLambda",
				Metadata:   &shared.CatalogObjectMetadata{Name: "test-svc"},
				Spec:       minimalImageSpec(),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AwsAppRunnerService{
				ApiVersion: "aws.planton.dev/v1alpha1",
				Kind:       "AwsAppRunnerService",
				Spec:       minimalImageSpec(),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AwsAppRunnerService{
				ApiVersion: "aws.planton.dev/v1alpha1",
				Kind:       "AwsAppRunnerService",
				Metadata:   &shared.CatalogObjectMetadata{Name: "test-svc"},
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
