package awssagemakerdomainv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAwsSagemakerDomainSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsSagemakerDomainSpec Validation Tests")
}

func validMinimalSpec() *AwsSagemakerDomain {
	return &AwsSagemakerDomain{
		ApiVersion: "aws.planton.dev/v1",
		Kind:       "AwsSagemakerDomain",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-sagemaker-domain",
		},
		Spec: &AwsSagemakerDomainSpec{
			Region:   "us-west-2",
			AuthMode: "IAM",
			VpcId: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vpc-0abc123def456"},
			},
			SubnetIds: []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-aaa"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-bbb"}},
			},
			DefaultUserSettings: &AwsSagemakerDomainDefaultUserSettings{
				ExecutionRoleArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/SageMakerExecRole"},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AwsSagemakerDomainSpec Validation Tests", func() {

	// ===== HAPPY PATH TESTS =====

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal valid domain with IAM auth", func() {
			input := validMinimalSpec()
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with SSO auth", func() {
			input := validMinimalSpec()
			input.Spec.AuthMode = "SSO"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with VpcOnly network", func() {
			input := validMinimalSpec()
			input.Spec.AppNetworkAccessType = proto.String("VpcOnly")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with explicit PublicInternetOnly", func() {
			input := validMinimalSpec()
			input.Spec.AppNetworkAccessType = proto.String("PublicInternetOnly")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with KMS encryption", func() {
			input := validMinimalSpec()
			input.Spec.KmsKeyId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:kms:us-east-1:123456789012:key/mrk-abc123"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with user security groups", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-user1"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-user2"}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with domain security groups", func() {
			input := validMinimalSpec()
			input.Spec.DomainSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-domain1"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-domain2"}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with JupyterLab settings", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{
					InstanceType: "ml.t3.medium",
				},
				LifecycleConfigArns: []string{
					"arn:aws:sagemaker:us-east-1:123456789012:studio-lifecycle-config/install-packages",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with JupyterLab idle settings", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				IdleSettings: &AwsSagemakerDomainIdleSettings{
					LifecycleManagement:  "ENABLED",
					IdleTimeoutInMinutes: 120,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with JupyterLab code repositories", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				CodeRepositories: []*AwsSagemakerDomainCodeRepository{
					{RepositoryUrl: "https://github.com/org/ml-notebooks.git"},
					{RepositoryUrl: "https://github.com/org/shared-utils.git"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with custom images", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				CustomImages: []*AwsSagemakerDomainCustomImage{
					{
						AppImageConfigName: "pytorch-config",
						ImageName:          "pytorch-custom",
						ImageVersionNumber: proto.Int32(1),
					},
					{
						AppImageConfigName: "tensorflow-config",
						ImageName:          "tensorflow-custom",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with KernelGateway settings", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.KernelGatewayAppSettings = &AwsSagemakerDomainKernelGatewayAppSettings{
				DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{
					InstanceType:      "ml.g4dn.xlarge",
					SagemakerImageArn: "arn:aws:sagemaker:us-east-1:123456789012:image/gpu-kernel",
				},
				CustomImages: []*AwsSagemakerDomainCustomImage{
					{AppImageConfigName: "gpu-config", ImageName: "gpu-image"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with sharing enabled", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.SharingSettings = &AwsSagemakerDomainSharingSettings{
				NotebookOutputOption: proto.String("Allowed"),
				S3OutputPath:         "s3://my-bucket/notebook-outputs/",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with sharing disabled explicitly", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.SharingSettings = &AwsSagemakerDomainSharingSettings{
				NotebookOutputOption: proto.String("Disabled"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with sharing and KMS encryption", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.SharingSettings = &AwsSagemakerDomainSharingSettings{
				NotebookOutputOption: proto.String("Allowed"),
				S3OutputPath:         "s3://my-bucket/outputs/",
				S3KmsKeyId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:kms:us-east-1:123456789012:key/mrk-s3"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with EBS storage settings", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.SpaceStorageSettings = &AwsSagemakerDomainSpaceStorageSettings{
				DefaultEbsVolumeSizeInGb: 10,
				MaximumEbsVolumeSizeInGb: 100,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with Docker enabled", func() {
			input := validMinimalSpec()
			input.Spec.DockerSettings = &AwsSagemakerDomainDockerSettings{
				EnableDockerAccess: "ENABLED",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with Docker and trusted accounts", func() {
			input := validMinimalSpec()
			input.Spec.AppNetworkAccessType = proto.String("VpcOnly")
			input.Spec.DockerSettings = &AwsSagemakerDomainDockerSettings{
				EnableDockerAccess:     "ENABLED",
				VpcOnlyTrustedAccounts: []string{"111122223333", "444455556666"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with default landing URI", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.DefaultLandingUri = "studio::relative/JupyterLab"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with studio web portal disabled", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.StudioWebPortal = proto.String("DISABLED")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a domain with valueFrom references", func() {
			input := validMinimalSpec()
			input.Spec.VpcId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Name: "my-vpc",
					},
				},
			}
			input.Spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-vpc"},
				}},
			}
			input.Spec.DefaultUserSettings.ExecutionRoleArn = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "sagemaker-role"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a production-ready domain with all settings", func() {
			input := validMinimalSpec()
			input.Spec.AuthMode = "SSO"
			input.Spec.AppNetworkAccessType = proto.String("VpcOnly")
			input.Spec.KmsKeyId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:kms:us-east-1:123456789012:key/mrk-prod"},
			}
			input.Spec.DomainSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-domain1"}},
			}
			input.Spec.DockerSettings = &AwsSagemakerDomainDockerSettings{
				EnableDockerAccess:     "ENABLED",
				VpcOnlyTrustedAccounts: []string{"123456789012"},
			}
			input.Spec.DefaultUserSettings.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-user1"}},
			}
			input.Spec.DefaultUserSettings.StudioWebPortal = proto.String("ENABLED")
			input.Spec.DefaultUserSettings.DefaultLandingUri = "studio::relative/JupyterLab"
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{
					InstanceType: "ml.t3.medium",
				},
				IdleSettings: &AwsSagemakerDomainIdleSettings{
					LifecycleManagement:     "ENABLED",
					IdleTimeoutInMinutes:    120,
					MinIdleTimeoutInMinutes: 60,
					MaxIdleTimeoutInMinutes: 480,
				},
				CodeRepositories: []*AwsSagemakerDomainCodeRepository{
					{RepositoryUrl: "https://github.com/team/ml-platform.git"},
				},
			}
			input.Spec.DefaultUserSettings.KernelGatewayAppSettings = &AwsSagemakerDomainKernelGatewayAppSettings{
				DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{
					InstanceType: "ml.g4dn.xlarge",
				},
			}
			input.Spec.DefaultUserSettings.SharingSettings = &AwsSagemakerDomainSharingSettings{
				NotebookOutputOption: proto.String("Allowed"),
				S3OutputPath:         "s3://ml-team-bucket/notebook-outputs/",
				S3KmsKeyId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:kms:us-east-1:123456789012:key/mrk-s3"},
				},
			}
			input.Spec.DefaultUserSettings.SpaceStorageSettings = &AwsSagemakerDomainSpaceStorageSettings{
				DefaultEbsVolumeSizeInGb: 20,
				MaximumEbsVolumeSizeInGb: 200,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// ===== FAILURE TESTS =====

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should fail when auth_mode is missing", func() {
			input := validMinimalSpec()
			input.Spec.AuthMode = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when auth_mode is invalid", func() {
			input := validMinimalSpec()
			input.Spec.AuthMode = "LDAP"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_id is missing", func() {
			input := validMinimalSpec()
			input.Spec.VpcId = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when subnet_ids is empty", func() {
			input := validMinimalSpec()
			input.Spec.SubnetIds = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when default_user_settings is missing", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when execution_role_arn is missing", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.ExecutionRoleArn = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when app_network_access_type is invalid", func() {
			input := validMinimalSpec()
			input.Spec.AppNetworkAccessType = proto.String("DirectConnect")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when studio_web_portal is invalid", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.StudioWebPortal = proto.String("MAYBE")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when notebook_output_option is invalid", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.SharingSettings = &AwsSagemakerDomainSharingSettings{
				NotebookOutputOption: proto.String("AlwaysShare"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when s3_output_path is missing with Allowed sharing", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.SharingSettings = &AwsSagemakerDomainSharingSettings{
				NotebookOutputOption: proto.String("Allowed"),
				S3OutputPath:         "",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when max_ebs is less than default_ebs", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.SpaceStorageSettings = &AwsSagemakerDomainSpaceStorageSettings{
				DefaultEbsVolumeSizeInGb: 100,
				MaximumEbsVolumeSizeInGb: 50,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when enable_docker_access is invalid", func() {
			input := validMinimalSpec()
			input.Spec.DockerSettings = &AwsSagemakerDomainDockerSettings{
				EnableDockerAccess: "YES",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when lifecycle_management is invalid", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				IdleSettings: &AwsSagemakerDomainIdleSettings{
					LifecycleManagement: "AUTO",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when idle_timeout is set without lifecycle_management enabled", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				IdleSettings: &AwsSagemakerDomainIdleSettings{
					IdleTimeoutInMinutes: 120,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when idle_timeout_in_minutes is below minimum", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				IdleSettings: &AwsSagemakerDomainIdleSettings{
					LifecycleManagement:  "ENABLED",
					IdleTimeoutInMinutes: 30,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when idle_timeout_in_minutes exceeds maximum", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				IdleSettings: &AwsSagemakerDomainIdleSettings{
					LifecycleManagement:  "ENABLED",
					IdleTimeoutInMinutes: 600000,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when custom_image is missing app_image_config_name", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				CustomImages: []*AwsSagemakerDomainCustomImage{
					{ImageName: "my-image"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when custom_image is missing image_name", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				CustomImages: []*AwsSagemakerDomainCustomImage{
					{AppImageConfigName: "my-config"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when code_repository is missing repository_url", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				CodeRepositories: []*AwsSagemakerDomainCodeRepository{
					{RepositoryUrl: ""},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when domain_security_group_ids exceeds max of 3", func() {
			input := validMinimalSpec()
			input.Spec.DomainSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-1"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-2"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-3"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-4"}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when user security_group_ids exceeds max of 5", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-1"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-2"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-3"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-4"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-5"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-6"}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	// ===== API ENVELOPE TESTS =====

	ginkgo.Describe("API envelope validation", func() {

		ginkgo.It("should fail with wrong apiVersion", func() {
			input := validMinimalSpec()
			input.ApiVersion = "gcp.planton.dev/v1"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail with wrong kind", func() {
			input := validMinimalSpec()
			input.Kind = "AwsEksCluster"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail with missing metadata", func() {
			input := validMinimalSpec()
			input.Metadata = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail with missing spec", func() {
			input := validMinimalSpec()
			input.Spec = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept valid complete envelope", func() {
			input := validMinimalSpec()
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})
