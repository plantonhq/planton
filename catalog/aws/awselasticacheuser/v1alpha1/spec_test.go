package awselasticacheuserv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestAwsElasticacheUserSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsElasticacheUserSpec Validation Tests")
}

// passwordUser is the common production shape: one application identity
// authenticating with a password pair for zero-downtime rotation.
func passwordUser() *AwsElasticacheUser {
	return &AwsElasticacheUser{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsElasticacheUser",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "orders-service",
		},
		Spec: &AwsElasticacheUserSpec{
			Region:       "us-west-2",
			Engine:       "redis",
			UserName:     "orders-service",
			AccessString: "on ~orders:* +@read +@write",
			AuthenticationMode: &AwsElasticacheUserAuthenticationMode{
				Type:      "password",
				Passwords: []string{"a-very-strong-password"},
			},
		},
	}
}

// lockedDefaultUser is the mandatory group member: a "default" user that
// rejects unauthenticated access outright.
func lockedDefaultUser() *AwsElasticacheUser {
	return &AwsElasticacheUser{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsElasticacheUser",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "rbac-default-user",
		},
		Spec: &AwsElasticacheUserSpec{
			Region:       "us-west-2",
			Engine:       "redis",
			UserName:     "default",
			AccessString: "off ~* +@all",
			AuthenticationMode: &AwsElasticacheUserAuthenticationMode{
				Type: "no-password-required",
			},
		},
	}
}

var _ = ginkgo.Describe("AwsElasticacheUserSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_elasticache_user", func() {

			ginkgo.It("should accept a password-authenticated user", func() {
				err := protovalidate.Validate(passwordUser())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept two passwords for zero-downtime rotation", func() {
				input := passwordUser()
				input.Spec.AuthenticationMode.Passwords = []string{
					"the-current-password", "the-next-password-x",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a locked-down no-password default user", func() {
				err := protovalidate.Validate(lockedDefaultUser())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an IAM-authenticated user", func() {
				input := passwordUser()
				input.Spec.AuthenticationMode = &AwsElasticacheUserAuthenticationMode{
					Type: "iam",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the valkey engine", func() {
				input := passwordUser()
				input.Spec.Engine = "valkey"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_elasticache_user", func() {

			ginkgo.It("should reject a missing region", func() {
				input := passwordUser()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unsupported engine", func() {
				input := passwordUser()
				input.Spec.Engine = "memcached"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing user_name", func() {
				input := passwordUser()
				input.Spec.UserName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing access_string", func() {
				input := passwordUser()
				input.Spec.AccessString = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing authentication_mode", func() {
				input := passwordUser()
				input.Spec.AuthenticationMode = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unsupported authentication type", func() {
				input := passwordUser()
				input.Spec.AuthenticationMode.Type = "certificate"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a password type with no passwords", func() {
				input := passwordUser()
				input.Spec.AuthenticationMode.Passwords = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject more than two passwords", func() {
				input := passwordUser()
				input.Spec.AuthenticationMode.Passwords = []string{
					"the-first-password!", "the-second-password", "the-third-password!",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a password shorter than 16 characters", func() {
				input := passwordUser()
				input.Spec.AuthenticationMode.Passwords = []string{"too-short"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject passwords on the iam type", func() {
				input := passwordUser()
				input.Spec.AuthenticationMode = &AwsElasticacheUserAuthenticationMode{
					Type:      "iam",
					Passwords: []string{"a-very-strong-password"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject passwords on the no-password-required type", func() {
				input := lockedDefaultUser()
				input.Spec.AuthenticationMode.Passwords = []string{"a-very-strong-password"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
