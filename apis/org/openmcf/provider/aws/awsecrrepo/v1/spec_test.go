package awsecrrepov1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"google.golang.org/protobuf/proto"
)

// stringOfLength is a helper for test scenarios needing an overly long string.
func stringOfLength(n int) string {
	s := make([]byte, n)
	for i := range s {
		s[i] = 'a'
	}
	return string(s)
}

func TestAwsEcrRepoSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsEcrRepoSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AwsEcrRepoSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_ecr_repo", func() {
			var input *AwsEcrRepo

			ginkgo.BeforeEach(func() {
				input = &AwsEcrRepo{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsEcrRepo",
					Metadata: &shared.CloudResourceMetadata{
						Name: "valid-repo-metadata",
					},
					Spec: &AwsEcrRepoSpec{
						Region:         "us-west-2",
						RepositoryName: "my-valid-repo",
						ImageImmutable: true,
						EncryptionType: proto.String("AES256"),
						ForceDelete:    false,
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.Context("repository_name length constraints", func() {
				ginkgo.It("should fail if repository_name is too short", func() {
					input.Spec.RepositoryName = "a"
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})

				ginkgo.It("should fail if repository_name is too long", func() {
					input.Spec.RepositoryName = stringOfLength(257)
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})

				ginkgo.It("should succeed if repository_name length is within the valid range", func() {
					input.Spec.RepositoryName = "valid-repo-123"
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				})
			})

			ginkgo.Context("encryption_type constraints", func() {
				ginkgo.It("should fail if encryption_type is invalid", func() {
					input.Spec.EncryptionType = proto.String("INVALID")
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})

				ginkgo.It("should succeed if encryption_type is AES256", func() {
					input.Spec.EncryptionType = proto.String("AES256")
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("should succeed if encryption_type is KMS", func() {
					input.Spec.EncryptionType = proto.String("KMS")
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				})
			})

			ginkgo.Context("lifecycle_policy constraints", func() {
				ginkgo.It("should succeed with valid lifecycle policy", func() {
					input.Spec.LifecyclePolicy = &AwsEcrRepoLifecyclePolicy{
						ExpireUntaggedAfterDays: proto.Int32(14),
						MaxImageCount:           proto.Int32(30),
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				})

				ginkgo.It("should fail if expire_untagged_after_days is less than 1", func() {
					input.Spec.LifecyclePolicy = &AwsEcrRepoLifecyclePolicy{
						ExpireUntaggedAfterDays: proto.Int32(0),
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})

				ginkgo.It("should fail if expire_untagged_after_days is greater than 365", func() {
					input.Spec.LifecyclePolicy = &AwsEcrRepoLifecyclePolicy{
						ExpireUntaggedAfterDays: proto.Int32(366),
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})

				ginkgo.It("should fail if max_image_count is less than 1", func() {
					input.Spec.LifecyclePolicy = &AwsEcrRepoLifecyclePolicy{
						MaxImageCount: proto.Int32(0),
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})

				ginkgo.It("should fail if max_image_count is greater than 1000", func() {
					input.Spec.LifecyclePolicy = &AwsEcrRepoLifecyclePolicy{
						MaxImageCount: proto.Int32(1001),
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil())
				})

				ginkgo.It("should succeed with no lifecycle policy", func() {
					input.Spec.LifecyclePolicy = nil
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				})
			})
		})
	})
})
