package alicloudrampolicyv1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAliCloudRamPolicySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudRamPolicySpec Validation Tests")
}

func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }

const validPolicyDocument = `{"Version":"1","Statement":[{"Effect":"Allow","Action":["oss:GetObject","oss:ListObjects"],"Resource":["acs:oss:*:*:my-bucket/*"]}]}`

var _ = ginkgo.Describe("AliCloudRamPolicySpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AliCloudRamPolicy{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudRamPolicy",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-policy",
				},
				Spec: &AliCloudRamPolicySpec{
					Region:         "cn-hangzhou",
					PolicyName:     "my-oss-reader",
					PolicyDocument: validPolicyDocument,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AliCloudRamPolicy{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudRamPolicy",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-config",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AliCloudRamPolicySpec{
					Region:         "us-west-1",
					PolicyName:     "full-config-policy",
					Description:    "Grants read-only access to audit logs",
					PolicyDocument: validPolicyDocument,
					RotateStrategy: stringPtr("DeleteOldestNonDefaultVersionWhenLimitExceeded"),
					Tags:           map[string]string{"team": "security", "cost-center": "compliance"},
					Force:          boolPtr(true),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with rotate_strategy set to None", func() {
			input := &AliCloudRamPolicy{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudRamPolicy",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamPolicySpec{
					Region:         "cn-hangzhou",
					PolicyName:     "none-rotation",
					PolicyDocument: validPolicyDocument,
					RotateStrategy: stringPtr("None"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with policy_name at max length (128 chars)", func() {
			input := &AliCloudRamPolicy{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudRamPolicy",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamPolicySpec{
					Region:         "cn-hangzhou",
					PolicyName:     strings.Repeat("a", 128),
					PolicyDocument: validPolicyDocument,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AliCloudRamPolicy{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudRamPolicy",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamPolicySpec{
					PolicyName:     "my-policy",
					PolicyDocument: validPolicyDocument,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when policy_name is missing", func() {
			input := &AliCloudRamPolicy{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudRamPolicy",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamPolicySpec{
					Region:         "cn-hangzhou",
					PolicyDocument: validPolicyDocument,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when policy_document is missing", func() {
			input := &AliCloudRamPolicy{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudRamPolicy",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamPolicySpec{
					Region:     "cn-hangzhou",
					PolicyName: "my-policy",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when policy_name exceeds 128 characters", func() {
			input := &AliCloudRamPolicy{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudRamPolicy",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamPolicySpec{
					Region:         "cn-hangzhou",
					PolicyName:     strings.Repeat("a", 129),
					PolicyDocument: validPolicyDocument,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when description exceeds 1024 characters", func() {
			input := &AliCloudRamPolicy{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudRamPolicy",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamPolicySpec{
					Region:         "cn-hangzhou",
					PolicyName:     "my-policy",
					Description:    strings.Repeat("x", 1025),
					PolicyDocument: validPolicyDocument,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when rotate_strategy is invalid", func() {
			input := &AliCloudRamPolicy{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudRamPolicy",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamPolicySpec{
					Region:         "cn-hangzhou",
					PolicyName:     "my-policy",
					PolicyDocument: validPolicyDocument,
					RotateStrategy: stringPtr("InvalidStrategy"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudRamPolicy{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudRamPolicy",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamPolicySpec{
					Region:         "cn-hangzhou",
					PolicyName:     "my-policy",
					PolicyDocument: validPolicyDocument,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudRamPolicy{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudRamPolicySpec{
					Region:         "cn-hangzhou",
					PolicyName:     "my-policy",
					PolicyDocument: validPolicyDocument,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AliCloudRamPolicy{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudRamPolicy",
				Spec: &AliCloudRamPolicySpec{
					Region:         "cn-hangzhou",
					PolicyName:     "my-policy",
					PolicyDocument: validPolicyDocument,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
