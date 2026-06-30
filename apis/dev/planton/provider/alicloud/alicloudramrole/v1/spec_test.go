package alicloudramrolev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestAliCloudRamRoleSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudRamRoleSpec Validation Tests")
}

func int32Ptr(i int32) *int32    { return &i }
func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }

const validTrustPolicy = `{"Statement":[{"Action":"sts:AssumeRole","Effect":"Allow","Principal":{"Service":["ecs.aliyuncs.com"]}}],"Version":"1"}`

var _ = ginkgo.Describe("AliCloudRamRoleSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AliCloudRamRole{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRamRole",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-role",
				},
				Spec: &AliCloudRamRoleSpec{
					Region:                   "cn-hangzhou",
					RoleName:                 "my-ecs-role",
					AssumeRolePolicyDocument: validTrustPolicy,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with policy attachments", func() {
			input := &AliCloudRamRole{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRamRole",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-role",
				},
				Spec: &AliCloudRamRoleSpec{
					Region:                   "cn-shanghai",
					RoleName:                 "ack-worker-role",
					Description:              "Role for ACK worker nodes",
					AssumeRolePolicyDocument: validTrustPolicy,
					PolicyAttachments: []*AliCloudRamRolePolicyAttachment{
						{
							PolicyName: "AliyunECSFullAccess",
							PolicyType: stringPtr("System"),
						},
						{
							PolicyName: "AliyunOSSReadOnlyAccess",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AliCloudRamRole{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRamRole",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-config",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AliCloudRamRoleSpec{
					Region:                   "us-west-1",
					RoleName:                 "full-config-role",
					Description:              "Fully configured role",
					AssumeRolePolicyDocument: validTrustPolicy,
					MaxSessionDuration:       int32Ptr(7200),
					Tags:                     map[string]string{"team": "platform", "cost-center": "eng"},
					Force:                    boolPtr(true),
					PolicyAttachments: []*AliCloudRamRolePolicyAttachment{
						{
							PolicyName: "AliyunECSFullAccess",
							PolicyType: stringPtr("System"),
						},
						{
							PolicyName: "my-custom-policy",
							PolicyType: stringPtr("Custom"),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with max_session_duration at boundary values", func() {
			input := &AliCloudRamRole{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRamRole",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamRoleSpec{
					Region:                   "cn-hangzhou",
					RoleName:                 "boundary-test",
					AssumeRolePolicyDocument: validTrustPolicy,
					MaxSessionDuration:       int32Ptr(43200),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AliCloudRamRole{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRamRole",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamRoleSpec{
					RoleName:                 "my-role",
					AssumeRolePolicyDocument: validTrustPolicy,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when role_name is missing", func() {
			input := &AliCloudRamRole{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRamRole",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamRoleSpec{
					Region:                   "cn-hangzhou",
					AssumeRolePolicyDocument: validTrustPolicy,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when assume_role_policy_document is missing", func() {
			input := &AliCloudRamRole{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRamRole",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamRoleSpec{
					Region:   "cn-hangzhou",
					RoleName: "my-role",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when role_name exceeds 64 characters", func() {
			input := &AliCloudRamRole{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRamRole",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamRoleSpec{
					Region:                   "cn-hangzhou",
					RoleName:                 "this-role-name-is-intentionally-longer-than-sixty-four-characters-which-exceeds-the-maximum",
					AssumeRolePolicyDocument: validTrustPolicy,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when max_session_duration is below minimum", func() {
			input := &AliCloudRamRole{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRamRole",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamRoleSpec{
					Region:                   "cn-hangzhou",
					RoleName:                 "my-role",
					AssumeRolePolicyDocument: validTrustPolicy,
					MaxSessionDuration:       int32Ptr(1800),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when max_session_duration exceeds maximum", func() {
			input := &AliCloudRamRole{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRamRole",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamRoleSpec{
					Region:                   "cn-hangzhou",
					RoleName:                 "my-role",
					AssumeRolePolicyDocument: validTrustPolicy,
					MaxSessionDuration:       int32Ptr(50000),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when policy_type is invalid", func() {
			input := &AliCloudRamRole{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRamRole",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamRoleSpec{
					Region:                   "cn-hangzhou",
					RoleName:                 "my-role",
					AssumeRolePolicyDocument: validTrustPolicy,
					PolicyAttachments: []*AliCloudRamRolePolicyAttachment{
						{
							PolicyName: "some-policy",
							PolicyType: stringPtr("Invalid"),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when policy_name is missing in attachment", func() {
			input := &AliCloudRamRole{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRamRole",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamRoleSpec{
					Region:                   "cn-hangzhou",
					RoleName:                 "my-role",
					AssumeRolePolicyDocument: validTrustPolicy,
					PolicyAttachments: []*AliCloudRamRolePolicyAttachment{
						{
							PolicyType: stringPtr("System"),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudRamRole{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudRamRole",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamRoleSpec{
					Region:                   "cn-hangzhou",
					RoleName:                 "my-role",
					AssumeRolePolicyDocument: validTrustPolicy,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudRamRole{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamRoleSpec{
					Region:                   "cn-hangzhou",
					RoleName:                 "my-role",
					AssumeRolePolicyDocument: validTrustPolicy,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
