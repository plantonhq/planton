package alicloudkmskeyv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"google.golang.org/protobuf/proto"
)

func TestAliCloudKmsKeySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudKmsKeySpec Validation Tests")
}

var _ = ginkgo.Describe("AliCloudKmsKeySpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AliCloudKmsKey{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudKmsKey",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-key",
				},
				Spec: &AliCloudKmsKeySpec{
					Region: "cn-hangzhou",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AliCloudKmsKey{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudKmsKey",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-key",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AliCloudKmsKeySpec{
					Region:                        "cn-shanghai",
					Description:                   "Production encryption key for RDS and OSS",
					KeySpec:                       proto.String("Aliyun_AES_256"),
					KeyUsage:                      proto.String("ENCRYPT/DECRYPT"),
					ProtectionLevel:               proto.String("SOFTWARE"),
					AutomaticRotation:             proto.Bool(true),
					RotationInterval:              "365d",
					PendingWindowInDays:           proto.Int32(30),
					DeletionProtection:            proto.Bool(true),
					DeletionProtectionDescription: "Protects RDS TDE master key",
					Tags:                          map[string]string{"team": "platform", "env": "prod"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with HSM protection level", func() {
			input := &AliCloudKmsKey{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudKmsKey",
				Metadata: &shared.CloudResourceMetadata{
					Name: "hsm-key",
				},
				Spec: &AliCloudKmsKeySpec{
					Region:          "cn-hangzhou",
					ProtectionLevel: proto.String("HSM"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with SIGN/VERIFY key usage and asymmetric key specs", func() {
			for _, spec := range []string{"RSA_2048", "RSA_3072", "EC_P256", "EC_P256K", "EC_SM2"} {
				input := &AliCloudKmsKey{
					ApiVersion: "alicloud.planton.dev/v1",
					Kind:       "AliCloudKmsKey",
					Metadata: &shared.CloudResourceMetadata{
						Name: "sign-key",
					},
					Spec: &AliCloudKmsKeySpec{
						Region:   "cn-hangzhou",
						KeySpec:  proto.String(spec),
						KeyUsage: proto.String("SIGN/VERIFY"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			}
		})

		ginkgo.It("should pass with all symmetric key specs", func() {
			for _, spec := range []string{"Aliyun_AES_256", "Aliyun_AES_128", "Aliyun_AES_192", "Aliyun_SM4"} {
				input := &AliCloudKmsKey{
					ApiVersion: "alicloud.planton.dev/v1",
					Kind:       "AliCloudKmsKey",
					Metadata: &shared.CloudResourceMetadata{
						Name: "sym-key",
					},
					Spec: &AliCloudKmsKeySpec{
						Region:  "cn-hangzhou",
						KeySpec: proto.String(spec),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			}
		})

		ginkgo.It("should pass with pending_window_in_days at boundary values", func() {
			input := &AliCloudKmsKey{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudKmsKey",
				Metadata: &shared.CloudResourceMetadata{
					Name: "min-window-key",
				},
				Spec: &AliCloudKmsKeySpec{
					Region:              "cn-hangzhou",
					PendingWindowInDays: proto.Int32(7),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())

			input.Spec.PendingWindowInDays = proto.Int32(366)
			err = protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with rotation disabled and no rotation_interval", func() {
			input := &AliCloudKmsKey{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudKmsKey",
				Metadata: &shared.CloudResourceMetadata{
					Name: "no-rotation-key",
				},
				Spec: &AliCloudKmsKeySpec{
					Region:            "us-west-1",
					AutomaticRotation: proto.Bool(false),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AliCloudKmsKey{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudKmsKey",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudKmsKeySpec{},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudKmsKey{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudKmsKey",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudKmsKeySpec{
					Region: "cn-hangzhou",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudKmsKey{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudKmsKeySpec{
					Region: "cn-hangzhou",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AliCloudKmsKey{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudKmsKey",
				Spec: &AliCloudKmsKeySpec{
					Region: "cn-hangzhou",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when key_spec has invalid value", func() {
			input := &AliCloudKmsKey{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudKmsKey",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudKmsKeySpec{
					Region:  "cn-hangzhou",
					KeySpec: proto.String("AES_512"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when key_usage has invalid value", func() {
			input := &AliCloudKmsKey{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudKmsKey",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudKmsKeySpec{
					Region:   "cn-hangzhou",
					KeyUsage: proto.String("WRAP/UNWRAP"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when protection_level has invalid value", func() {
			input := &AliCloudKmsKey{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudKmsKey",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudKmsKeySpec{
					Region:          "cn-hangzhou",
					ProtectionLevel: proto.String("CLOUD_HSM"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when pending_window_in_days is below minimum", func() {
			input := &AliCloudKmsKey{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudKmsKey",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudKmsKeySpec{
					Region:              "cn-hangzhou",
					PendingWindowInDays: proto.Int32(6),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when pending_window_in_days exceeds maximum", func() {
			input := &AliCloudKmsKey{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudKmsKey",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudKmsKeySpec{
					Region:              "cn-hangzhou",
					PendingWindowInDays: proto.Int32(367),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
