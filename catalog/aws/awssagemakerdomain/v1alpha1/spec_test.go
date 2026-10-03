package awssagemakerdomainv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAwsSagemakerDomainSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsSagemakerDomainSpec Validation Tests")
}

func validMinimalSpec() *AwsSagemakerDomain {
	return &AwsSagemakerDomain{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsSagemakerDomain",
		Metadata: &shared.CatalogObjectMetadata{
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
			DefaultUserSettings: &AwsSagemakerDomainUserSettings{
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
					IdleTimeoutInMinutes:    120,
					MinIdleTimeoutInMinutes: 60,
					MaxIdleTimeoutInMinutes: 480,
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

		ginkgo.It("should fail when idle settings omit the user min/max bounds", func() {
			// The live API rejects a partial idle block (absent members
			// transmit as 0, below the 60-minute floor), so all three
			// timeouts are required whenever the block is set.
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
					IdleTimeoutInMinutes:    30,
					MinIdleTimeoutInMinutes: 60,
					MaxIdleTimeoutInMinutes: 480,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when idle_timeout_in_minutes exceeds maximum", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				IdleSettings: &AwsSagemakerDomainIdleSettings{
					IdleTimeoutInMinutes:    600000,
					MinIdleTimeoutInMinutes: 60,
					MaxIdleTimeoutInMinutes: 480,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when the max idle timeout is below the min", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				IdleSettings: &AwsSagemakerDomainIdleSettings{
					IdleTimeoutInMinutes:    120,
					MinIdleTimeoutInMinutes: 240,
					MaxIdleTimeoutInMinutes: 120,
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

	// ===== DOMAIN-SCOPED ADMINISTRATION =====

	ginkgo.Describe("Domain-scoped administration settings", func() {

		ginkgo.It("should accept tag_propagation ENABLED", func() {
			input := validMinimalSpec()
			input.Spec.TagPropagation = proto.String("ENABLED")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail when tag_propagation is invalid", func() {
			input := validMinimalSpec()
			input.Spec.TagPropagation = proto.String("ALWAYS")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept home_efs_retention_policy Delete", func() {
			input := validMinimalSpec()
			input.Spec.HomeEfsRetentionPolicy = proto.String("Delete")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail when home_efs_retention_policy is invalid", func() {
			input := validMinimalSpec()
			input.Spec.HomeEfsRetentionPolicy = proto.String("Keep")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept execution_role_identity_config USER_PROFILE_NAME", func() {
			input := validMinimalSpec()
			input.Spec.ExecutionRoleIdentityConfig = proto.String("USER_PROFILE_NAME")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail when execution_role_identity_config is invalid", func() {
			input := validMinimalSpec()
			input.Spec.ExecutionRoleIdentityConfig = proto.String("SESSION_NAME")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept trusted identity propagation ENABLED on an SSO domain", func() {
			input := validMinimalSpec()
			input.Spec.AuthMode = "SSO"
			input.Spec.TrustedIdentityPropagationStatus = proto.String("ENABLED")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept trusted identity propagation DISABLED on an SSO domain", func() {
			input := validMinimalSpec()
			input.Spec.AuthMode = "SSO"
			input.Spec.TrustedIdentityPropagationStatus = proto.String("DISABLED")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail when trusted identity propagation is set on an IAM domain", func() {
			// The live API rejects the setting outright on IAM-auth domains,
			// even "DISABLED" -- stricter than value-gating on ENABLED alone.
			for _, v := range []string{"ENABLED", "DISABLED"} {
				input := validMinimalSpec()
				input.Spec.TrustedIdentityPropagationStatus = proto.String(v)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			}
		})

		ginkgo.It("should fail when trusted_identity_propagation_status is invalid", func() {
			input := validMinimalSpec()
			input.Spec.TrustedIdentityPropagationStatus = proto.String("ON")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept app_security_group_management with RStudio configured", func() {
			input := validMinimalSpec()
			input.Spec.AppSecurityGroupManagement = proto.String("Customer")
			input.Spec.RStudioServerProDomainSettings = &AwsSagemakerDomainRStudioServerProDomainSettings{
				DomainExecutionRoleArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/RStudioDomainRole"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail when app_security_group_management is set without RStudio", func() {
			input := validMinimalSpec()
			input.Spec.AppSecurityGroupManagement = proto.String("Service")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when app_security_group_management is invalid", func() {
			input := validMinimalSpec()
			input.Spec.AppSecurityGroupManagement = proto.String("Managed")
			input.Spec.RStudioServerProDomainSettings = &AwsSagemakerDomainRStudioServerProDomainSettings{
				DomainExecutionRoleArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/RStudioDomainRole"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when RStudio domain settings are missing the domain execution role", func() {
			input := validMinimalSpec()
			input.Spec.RStudioServerProDomainSettings = &AwsSagemakerDomainRStudioServerProDomainSettings{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when subnet_ids exceeds max of 16", func() {
			input := validMinimalSpec()
			var subnets []*foreignkeyv1.StringValueOrRef
			for i := 0; i < 17; i++ {
				subnets = append(subnets, &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-x"},
				})
			}
			input.Spec.SubnetIds = subnets
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	// ===== USER-PLANE ADDITIONS =====

	ginkgo.Describe("Default user settings additions", func() {

		ginkgo.It("should accept auto_mount_home_efs values", func() {
			for _, v := range []string{"Enabled", "Disabled"} {
				input := validMinimalSpec()
				input.Spec.DefaultUserSettings.AutoMountHomeEfs = proto.String(v)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			}
		})

		ginkgo.It("should fail when auto_mount_home_efs is invalid", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.AutoMountHomeEfs = proto.String("ENABLED")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject the profile-only DefaultAsDomain value at the domain level", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.AutoMountHomeEfs = proto.String("DefaultAsDomain")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept Code Editor settings with idle shutdown", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.CodeEditorAppSettings = &AwsSagemakerDomainCodeEditorAppSettings{
				DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{InstanceType: "ml.t3.large"},
				IdleSettings: &AwsSagemakerDomainIdleSettings{
					IdleTimeoutInMinutes:    120,
					MinIdleTimeoutInMinutes: 60,
					MaxIdleTimeoutInMinutes: 480,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept classic Jupyter Server settings", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterServerAppSettings = &AwsSagemakerDomainJupyterServerAppSettings{
				DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{InstanceType: "system"},
				CodeRepositories: []*AwsSagemakerDomainCodeRepository{
					{RepositoryUrl: "https://github.com/org/classic-notebooks.git"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept TensorBoard settings", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.TensorBoardAppSettings = &AwsSagemakerDomainTensorBoardAppSettings{
				DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{InstanceType: "ml.m5.large"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept JupyterLab EMR settings with role refs", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				EmrSettings: &AwsSagemakerDomainEmrSettings{
					AssumableRoleArns: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/EmrConnect"}},
					},
					ExecutionRoleArns: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/EmrRuntime"}},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a built-in lifecycle config on JupyterLab", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				BuiltInLifecycleConfigArn: "arn:aws:sagemaker:us-west-2:123456789012:studio-lifecycle-config/built-in",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a resource spec pinning an image version by alias", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{
					SagemakerImageArn:          "arn:aws:sagemaker:us-west-2:123456789012:image/custom",
					SagemakerImageVersionAlias: "v2.0",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail when a resource spec sets both image version alias and ARN", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
				DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{
					SagemakerImageVersionAlias: "v2.0",
					SagemakerImageVersionArn:   "arn:aws:sagemaker:us-west-2:123456789012:image-version/custom/2",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept a custom EFS file system mount", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.CustomFileSystemConfigs = []*AwsSagemakerDomainCustomFileSystemConfig{
				{
					EfsFileSystemConfig: &AwsSagemakerDomainEfsFileSystemConfig{
						FileSystemId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "fs-0123456789abcdef0"},
						},
						FileSystemPath: "/shared/datasets",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail when a custom file system config has no EFS arm", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.CustomFileSystemConfigs = []*AwsSagemakerDomainCustomFileSystemConfig{{}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when an EFS mount is missing the file system path", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.CustomFileSystemConfigs = []*AwsSagemakerDomainCustomFileSystemConfig{
				{
					EfsFileSystemConfig: &AwsSagemakerDomainEfsFileSystemConfig{
						FileSystemId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "fs-0123456789abcdef0"},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept a valid POSIX user config", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.CustomPosixUserConfig = &AwsSagemakerDomainCustomPosixUserConfig{
				Uid: 10000,
				Gid: 1001,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail when POSIX uid is below 10000", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.CustomPosixUserConfig = &AwsSagemakerDomainCustomPosixUserConfig{
				Uid: 500,
				Gid: 1001,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when POSIX gid is below 1001", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.CustomPosixUserConfig = &AwsSagemakerDomainCustomPosixUserConfig{
				Uid: 10000,
				Gid: 1000,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept studio web portal hiding settings", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.StudioWebPortalSettings = &AwsSagemakerDomainStudioWebPortalSettings{
				HiddenAppTypes:      []string{"JupyterServer", "Canvas"},
				HiddenInstanceTypes: []string{"ml.p3.2xlarge"},
				HiddenMlTools:       []string{"DataWrangler"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept RStudio app access with a user group", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.RStudioServerProAppSettings = &AwsSagemakerDomainRStudioServerProAppSettings{
				AccessStatus: "ENABLED",
				UserGroup:    "R_STUDIO_ADMIN",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail when RStudio user_group is set without ENABLED access", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.RStudioServerProAppSettings = &AwsSagemakerDomainRStudioServerProAppSettings{
				AccessStatus: "DISABLED",
				UserGroup:    "R_STUDIO_USER",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when RStudio user_group is invalid", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.RStudioServerProAppSettings = &AwsSagemakerDomainRStudioServerProAppSettings{
				AccessStatus: "ENABLED",
				UserGroup:    "R_STUDIO_ROOT",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept RSession settings with custom images", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.RSessionAppSettings = &AwsSagemakerDomainRSessionAppSettings{
				DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{InstanceType: "ml.m5.large"},
				CustomImages: []*AwsSagemakerDomainCustomImage{
					{AppImageConfigName: "r-config", ImageName: "r-image"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// ===== CANVAS =====

	ginkgo.Describe("Canvas app settings", func() {

		ginkgo.It("should accept a full Canvas configuration", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.CanvasAppSettings = &AwsSagemakerDomainCanvasAppSettings{
				DirectDeployStatus: proto.String("ENABLED"),
				EmrServerlessSettings: &AwsSagemakerDomainCanvasEmrServerlessSettings{
					ExecutionRoleArn: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/CanvasEmr"},
					},
					Status: "ENABLED",
				},
				GenerativeAiBedrockRoleArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/CanvasBedrock"},
				},
				IdentityProviderOauthSettings: []*AwsSagemakerDomainCanvasIdentityProviderOauthSettings{
					{
						DataSourceName: "Snowflake",
						SecretArn:      "arn:aws:secretsmanager:us-west-2:123456789012:secret:snowflake-oauth",
						Status:         "ENABLED",
					},
				},
				KendraSettingsStatus: proto.String("ENABLED"),
				ModelRegisterSettings: &AwsSagemakerDomainCanvasModelRegisterSettings{
					Status: "ENABLED",
				},
				TimeSeriesForecastingSettings: &AwsSagemakerDomainCanvasTimeSeriesForecastingSettings{
					AmazonForecastRoleArn: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/CanvasForecast"},
					},
					Status: "ENABLED",
				},
				WorkspaceSettings: &AwsSagemakerDomainCanvasWorkspaceSettings{
					S3ArtifactPath: "s3://canvas-workspace/artifacts/",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail when direct_deploy_status is invalid", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.CanvasAppSettings = &AwsSagemakerDomainCanvasAppSettings{
				DirectDeployStatus: proto.String("ON"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when an OAuth connector is missing the secret ARN", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.CanvasAppSettings = &AwsSagemakerDomainCanvasAppSettings{
				IdentityProviderOauthSettings: []*AwsSagemakerDomainCanvasIdentityProviderOauthSettings{
					{DataSourceName: "Snowflake"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when an OAuth connector names an unknown data source", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.CanvasAppSettings = &AwsSagemakerDomainCanvasAppSettings{
				IdentityProviderOauthSettings: []*AwsSagemakerDomainCanvasIdentityProviderOauthSettings{
					{
						DataSourceName: "Databricks",
						SecretArn:      "arn:aws:secretsmanager:us-west-2:123456789012:secret:x",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when the Canvas workspace path is not an s3:// or https:// URI", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.CanvasAppSettings = &AwsSagemakerDomainCanvasAppSettings{
				WorkspaceSettings: &AwsSagemakerDomainCanvasWorkspaceSettings{
					S3ArtifactPath: "gs://wrong-cloud/artifacts/",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when EMR Serverless status is invalid", func() {
			input := validMinimalSpec()
			input.Spec.DefaultUserSettings.CanvasAppSettings = &AwsSagemakerDomainCanvasAppSettings{
				EmrServerlessSettings: &AwsSagemakerDomainCanvasEmrServerlessSettings{
					Status: "ACTIVE",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	// ===== SPACE PLANE =====

	ginkgo.Describe("Default space settings", func() {

		ginkgo.It("should accept space settings with their own execution role", func() {
			input := validMinimalSpec()
			input.Spec.DefaultSpaceSettings = &AwsSagemakerDomainDefaultSpaceSettings{
				ExecutionRoleArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/SpaceExecRole"},
				},
				JupyterLabAppSettings: &AwsSagemakerDomainJupyterLabAppSettings{
					DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{InstanceType: "ml.t3.medium"},
				},
				SpaceStorageSettings: &AwsSagemakerDomainSpaceStorageSettings{
					DefaultEbsVolumeSizeInGb: 10,
					MaximumEbsVolumeSizeInGb: 100,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail when space settings are missing the execution role", func() {
			input := validMinimalSpec()
			input.Spec.DefaultSpaceSettings = &AwsSagemakerDomainDefaultSpaceSettings{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when space security groups exceed max of 5", func() {
			input := validMinimalSpec()
			var sgs []*foreignkeyv1.StringValueOrRef
			for i := 0; i < 6; i++ {
				sgs = append(sgs, &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-x"},
				})
			}
			input.Spec.DefaultSpaceSettings = &AwsSagemakerDomainDefaultSpaceSettings{
				ExecutionRoleArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/SpaceExecRole"},
				},
				SecurityGroupIds: sgs,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	// ===== API ENVELOPE TESTS =====

	ginkgo.Describe("API envelope validation", func() {

		ginkgo.It("should fail with wrong apiVersion", func() {
			input := validMinimalSpec()
			input.ApiVersion = "gcp.planton.dev/v1alpha1"
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

// Coverage for the folded user-profile and space satellites, the idle-settings
// lifecycle_management switch, and the value-domain promotions.
var _ = ginkgo.Describe("AwsSagemakerDomain folded satellites and promotions", func() {

	ginkgo.It("accepts a domain with user profiles and a space", func() {
		input := validMinimalSpec()
		input.Spec.UserProfiles = []*AwsSagemakerDomainUserProfile{
			{UserProfileName: "alice"},
			{UserProfileName: "bob", UserSettings: &AwsSagemakerDomainUserSettings{
				ExecutionRoleArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/BobRole"},
				},
			}},
		}
		input.Spec.Spaces = []*AwsSagemakerDomainSpace{
			{
				SpaceName:            "team-analytics",
				DisplayName:          "Team Analytics",
				OwnershipSettings:    &AwsSagemakerDomainSpaceOwnership{OwnerUserProfileName: "alice"},
				SpaceSharingSettings: &AwsSagemakerDomainSpaceSharing{SharingType: "Shared"},
				SpaceSettings: &AwsSagemakerDomainSpaceSettings{
					AppType: proto.String("JupyterLab"),
					JupyterLabAppSettings: &AwsSagemakerDomainSpaceJupyterLabAppSettings{
						DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{InstanceType: "ml.t3.medium"},
						IdleSettings:        &AwsSagemakerDomainSpaceIdleSettings{IdleTimeoutInMinutes: proto.Int32(120)},
					},
					SpaceStorageSettings: &AwsSagemakerDomainSpaceStorage{EbsVolumeSizeInGb: 50},
				},
			},
		}
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("rejects duplicate user profile names", func() {
		input := validMinimalSpec()
		input.Spec.UserProfiles = []*AwsSagemakerDomainUserProfile{
			{UserProfileName: "alice"},
			{UserProfileName: "alice"},
		}
		gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
	})

	ginkgo.It("rejects duplicate space names", func() {
		input := validMinimalSpec()
		input.Spec.Spaces = []*AwsSagemakerDomainSpace{
			{SpaceName: "x"},
			{SpaceName: "x"},
		}
		gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid user profile name", func() {
		input := validMinimalSpec()
		input.Spec.UserProfiles = []*AwsSagemakerDomainUserProfile{{UserProfileName: "alice_smith"}}
		gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
	})

	ginkgo.It("rejects a lone SSO identifier without its value", func() {
		input := validMinimalSpec()
		input.Spec.UserProfiles = []*AwsSagemakerDomainUserProfile{{
			UserProfileName:            "alice",
			SingleSignOnUserIdentifier: "UserName",
		}}
		gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
	})

	ginkgo.It("rejects an SSO identifier other than UserName", func() {
		input := validMinimalSpec()
		input.Spec.UserProfiles = []*AwsSagemakerDomainUserProfile{{
			UserProfileName:            "alice",
			SingleSignOnUserIdentifier: "Email",
			SingleSignOnUserValue:      "alice@example.com",
		}}
		gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
	})

	ginkgo.It("rejects ownership without sharing settings", func() {
		input := validMinimalSpec()
		input.Spec.Spaces = []*AwsSagemakerDomainSpace{{
			SpaceName:         "solo",
			OwnershipSettings: &AwsSagemakerDomainSpaceOwnership{OwnerUserProfileName: "alice"},
		}}
		gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
	})

	ginkgo.It("rejects a space jupyter server block without a resource spec", func() {
		input := validMinimalSpec()
		input.Spec.Spaces = []*AwsSagemakerDomainSpace{{
			SpaceName: "legacy",
			SpaceSettings: &AwsSagemakerDomainSpaceSettings{
				JupyterServerAppSettings: &AwsSagemakerDomainJupyterServerAppSettings{},
			},
		}}
		gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
	})

	ginkgo.It("rejects a space EBS volume below 5 GB", func() {
		input := validMinimalSpec()
		input.Spec.Spaces = []*AwsSagemakerDomainSpace{{
			SpaceName: "small",
			SpaceSettings: &AwsSagemakerDomainSpaceSettings{
				SpaceStorageSettings: &AwsSagemakerDomainSpaceStorage{EbsVolumeSizeInGb: 4},
			},
		}}
		gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
	})

	// The CreateSpace server contract (live-caught 2026-08-13): a space idle
	// timeout is rejected wherever idle shutdown resolves DISABLED for the
	// space. The CELs resolve the in-manifest inheritance chain — the owner
	// profile's explicit setting wins, a silent owner inherits the domain's
	// default_user_settings plane.
	ginkgo.It("rejects a space idle timeout owned by a DISABLED-lifecycle profile (JupyterLab)", func() {
		input := validMinimalSpec()
		input.Spec.UserProfiles = []*AwsSagemakerDomainUserProfile{{
			UserProfileName: "quiet-owner",
			UserSettings: &AwsSagemakerDomainUserSettings{
				ExecutionRoleArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/Owner"},
				},
				JupyterLabAppSettings: &AwsSagemakerDomainJupyterLabAppSettings{
					IdleSettings: &AwsSagemakerDomainIdleSettings{
						LifecycleManagement:     proto.String("DISABLED"),
						IdleTimeoutInMinutes:    120,
						MinIdleTimeoutInMinutes: 60,
						MaxIdleTimeoutInMinutes: 240,
					},
				},
			},
		}}
		input.Spec.Spaces = []*AwsSagemakerDomainSpace{{
			SpaceName:            "contradiction",
			OwnershipSettings:    &AwsSagemakerDomainSpaceOwnership{OwnerUserProfileName: "quiet-owner"},
			SpaceSharingSettings: &AwsSagemakerDomainSpaceSharing{SharingType: "Private"},
			SpaceSettings: &AwsSagemakerDomainSpaceSettings{
				JupyterLabAppSettings: &AwsSagemakerDomainSpaceJupyterLabAppSettings{
					DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{InstanceType: "ml.t3.medium"},
					IdleSettings:        &AwsSagemakerDomainSpaceIdleSettings{IdleTimeoutInMinutes: proto.Int32(60)},
				},
			},
		}}
		gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
	})

	ginkgo.It("rejects a space idle timeout owned by a DISABLED-lifecycle profile (Code Editor)", func() {
		input := validMinimalSpec()
		input.Spec.UserProfiles = []*AwsSagemakerDomainUserProfile{{
			UserProfileName: "quiet-owner",
			UserSettings: &AwsSagemakerDomainUserSettings{
				ExecutionRoleArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/Owner"},
				},
				CodeEditorAppSettings: &AwsSagemakerDomainCodeEditorAppSettings{
					IdleSettings: &AwsSagemakerDomainIdleSettings{
						LifecycleManagement:     proto.String("DISABLED"),
						IdleTimeoutInMinutes:    120,
						MinIdleTimeoutInMinutes: 60,
						MaxIdleTimeoutInMinutes: 240,
					},
				},
			},
		}}
		input.Spec.Spaces = []*AwsSagemakerDomainSpace{{
			SpaceName:            "contradiction",
			OwnershipSettings:    &AwsSagemakerDomainSpaceOwnership{OwnerUserProfileName: "quiet-owner"},
			SpaceSharingSettings: &AwsSagemakerDomainSpaceSharing{SharingType: "Private"},
			SpaceSettings: &AwsSagemakerDomainSpaceSettings{
				CodeEditorAppSettings: &AwsSagemakerDomainSpaceCodeEditorAppSettings{
					DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{InstanceType: "ml.t3.medium"},
					IdleSettings:        &AwsSagemakerDomainSpaceIdleSettings{IdleTimeoutInMinutes: proto.Int32(60)},
				},
			},
		}}
		gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
	})

	ginkgo.It("accepts a space idle timeout under an ENABLED-lifecycle owner profile", func() {
		input := validMinimalSpec()
		input.Spec.UserProfiles = []*AwsSagemakerDomainUserProfile{{
			UserProfileName: "active-owner",
			UserSettings: &AwsSagemakerDomainUserSettings{
				ExecutionRoleArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/Owner"},
				},
				JupyterLabAppSettings: &AwsSagemakerDomainJupyterLabAppSettings{
					IdleSettings: &AwsSagemakerDomainIdleSettings{
						LifecycleManagement:     proto.String("ENABLED"),
						IdleTimeoutInMinutes:    120,
						MinIdleTimeoutInMinutes: 60,
						MaxIdleTimeoutInMinutes: 240,
					},
				},
			},
		}}
		input.Spec.Spaces = []*AwsSagemakerDomainSpace{{
			SpaceName:            "productive",
			OwnershipSettings:    &AwsSagemakerDomainSpaceOwnership{OwnerUserProfileName: "active-owner"},
			SpaceSharingSettings: &AwsSagemakerDomainSpaceSharing{SharingType: "Private"},
			SpaceSettings: &AwsSagemakerDomainSpaceSettings{
				JupyterLabAppSettings: &AwsSagemakerDomainSpaceJupyterLabAppSettings{
					DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{InstanceType: "ml.t3.medium"},
					IdleSettings:        &AwsSagemakerDomainSpaceIdleSettings{IdleTimeoutInMinutes: proto.Int32(60)},
				},
			},
		}}
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts a space idle timeout whose owner profile is outside the manifest", func() {
		// A profile provisioned outside the manifest resolves server-side —
		// the CEL must not reject what it cannot see, even when the domain
		// plane is DISABLED (the external profile may override to ENABLED).
		input := validMinimalSpec()
		input.Spec.Spaces = []*AwsSagemakerDomainSpace{{
			SpaceName:            "external-owner",
			OwnershipSettings:    &AwsSagemakerDomainSpaceOwnership{OwnerUserProfileName: "sso-provisioned"},
			SpaceSharingSettings: &AwsSagemakerDomainSpaceSharing{SharingType: "Private"},
			SpaceSettings: &AwsSagemakerDomainSpaceSettings{
				JupyterLabAppSettings: &AwsSagemakerDomainSpaceJupyterLabAppSettings{
					DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{InstanceType: "ml.t3.medium"},
					IdleSettings:        &AwsSagemakerDomainSpaceIdleSettings{IdleTimeoutInMinutes: proto.Int32(60)},
				},
			},
		}}
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("rejects a space idle timeout when its silent owner inherits a DISABLED domain plane (JupyterLab)", func() {
		// The inheritance blind spot, closed: the owner profile says nothing
		// about JupyterLab idle settings, so AWS resolves the domain's
		// default_user_settings plane — which disables idle shutdown. The
		// manifest must fail here, not at CreateSpace.
		input := validMinimalSpec()
		input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
			IdleSettings: &AwsSagemakerDomainIdleSettings{
				LifecycleManagement:     proto.String("DISABLED"),
				IdleTimeoutInMinutes:    120,
				MinIdleTimeoutInMinutes: 60,
				MaxIdleTimeoutInMinutes: 240,
			},
		}
		input.Spec.UserProfiles = []*AwsSagemakerDomainUserProfile{{
			UserProfileName: "silent-owner",
			UserSettings: &AwsSagemakerDomainUserSettings{
				ExecutionRoleArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/Owner"},
				},
			},
		}}
		input.Spec.Spaces = []*AwsSagemakerDomainSpace{{
			SpaceName:            "inherited-contradiction",
			OwnershipSettings:    &AwsSagemakerDomainSpaceOwnership{OwnerUserProfileName: "silent-owner"},
			SpaceSharingSettings: &AwsSagemakerDomainSpaceSharing{SharingType: "Private"},
			SpaceSettings: &AwsSagemakerDomainSpaceSettings{
				JupyterLabAppSettings: &AwsSagemakerDomainSpaceJupyterLabAppSettings{
					DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{InstanceType: "ml.t3.medium"},
					IdleSettings:        &AwsSagemakerDomainSpaceIdleSettings{IdleTimeoutInMinutes: proto.Int32(60)},
				},
			},
		}}
		gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
	})

	ginkgo.It("accepts a space idle timeout when the owner's explicit ENABLED overrides a DISABLED domain plane", func() {
		// Inheritance resolution order: the owner profile's explicit value
		// wins over the domain plane — a DISABLED default with an ENABLED
		// owner override is a legal, common shape.
		input := validMinimalSpec()
		input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
			IdleSettings: &AwsSagemakerDomainIdleSettings{
				LifecycleManagement:     proto.String("DISABLED"),
				IdleTimeoutInMinutes:    120,
				MinIdleTimeoutInMinutes: 60,
				MaxIdleTimeoutInMinutes: 240,
			},
		}
		input.Spec.UserProfiles = []*AwsSagemakerDomainUserProfile{{
			UserProfileName: "opted-in-owner",
			UserSettings: &AwsSagemakerDomainUserSettings{
				ExecutionRoleArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/Owner"},
				},
				JupyterLabAppSettings: &AwsSagemakerDomainJupyterLabAppSettings{
					IdleSettings: &AwsSagemakerDomainIdleSettings{
						LifecycleManagement:     proto.String("ENABLED"),
						IdleTimeoutInMinutes:    120,
						MinIdleTimeoutInMinutes: 60,
						MaxIdleTimeoutInMinutes: 240,
					},
				},
			},
		}}
		input.Spec.Spaces = []*AwsSagemakerDomainSpace{{
			SpaceName:            "opted-in",
			OwnershipSettings:    &AwsSagemakerDomainSpaceOwnership{OwnerUserProfileName: "opted-in-owner"},
			SpaceSharingSettings: &AwsSagemakerDomainSpaceSharing{SharingType: "Private"},
			SpaceSettings: &AwsSagemakerDomainSpaceSettings{
				JupyterLabAppSettings: &AwsSagemakerDomainSpaceJupyterLabAppSettings{
					DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{InstanceType: "ml.t3.medium"},
					IdleSettings:        &AwsSagemakerDomainSpaceIdleSettings{IdleTimeoutInMinutes: proto.Int32(60)},
				},
			},
		}}
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts the defined-but-disabled idle settings state", func() {
		input := validMinimalSpec()
		input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
			IdleSettings: &AwsSagemakerDomainIdleSettings{
				LifecycleManagement:     proto.String("DISABLED"),
				IdleTimeoutInMinutes:    120,
				MinIdleTimeoutInMinutes: 60,
				MaxIdleTimeoutInMinutes: 240,
			},
		}
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid lifecycle management value", func() {
		input := validMinimalSpec()
		input.Spec.DefaultUserSettings.JupyterLabAppSettings = &AwsSagemakerDomainJupyterLabAppSettings{
			IdleSettings: &AwsSagemakerDomainIdleSettings{
				LifecycleManagement:     proto.String("PAUSED"),
				IdleTimeoutInMinutes:    120,
				MinIdleTimeoutInMinutes: 60,
				MaxIdleTimeoutInMinutes: 240,
			},
		}
		gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
	})

	ginkgo.It("rejects a malformed trusted account id", func() {
		input := validMinimalSpec()
		input.Spec.DockerSettings = &AwsSagemakerDomainDockerSettings{
			EnableDockerAccess:     "ENABLED",
			VpcOnlyTrustedAccounts: []string{"12345"},
		}
		gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
	})

	ginkgo.It("rejects a malformed instance type and accepts system", func() {
		input := validMinimalSpec()
		input.Spec.DefaultUserSettings.JupyterServerAppSettings = &AwsSagemakerDomainJupyterServerAppSettings{
			DefaultResourceSpec: &AwsSagemakerDomainResourceSpec{InstanceType: "t3.medium"},
		}
		gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		input.Spec.DefaultUserSettings.JupyterServerAppSettings.DefaultResourceSpec.InstanceType = "system"
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})
})
