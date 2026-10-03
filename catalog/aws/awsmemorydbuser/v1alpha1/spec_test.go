package awsmemorydbuserv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestAwsMemorydbUserSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsMemorydbUserSpec Validation Tests")
}

// passwordUser is the common production shape: one application identity
// authenticating with a password pair for zero-downtime rotation.
func passwordUser() *AwsMemorydbUser {
	return &AwsMemorydbUser{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsMemorydbUser",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "orders-service",
		},
		Spec: &AwsMemorydbUserSpec{
			Region:       "us-west-2",
			AccessString: "on ~orders:* +@read +@write",
			AuthenticationMode: &AwsMemorydbUserAuthenticationMode{
				Type:      "password",
				Passwords: []string{"a-very-strong-password"},
			},
		},
	}
}

// iamUser is the credential-free shape: the client signs short-lived tokens
// with its IAM identity.
func iamUser() *AwsMemorydbUser {
	return &AwsMemorydbUser{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsMemorydbUser",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "analytics-service",
		},
		Spec: &AwsMemorydbUserSpec{
			Region:       "us-west-2",
			AccessString: "on ~analytics:* +@read",
			AuthenticationMode: &AwsMemorydbUserAuthenticationMode{
				Type: "iam",
			},
		},
	}
}

var _ = ginkgo.Describe("AwsMemorydbUserSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_memorydb_user", func() {

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

			ginkgo.It("should accept an IAM-authenticated user", func() {
				err := protovalidate.Validate(iamUser())
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("required fields", func() {

			ginkgo.It("should reject a missing region", func() {
				input := passwordUser()
				input.Spec.Region = ""
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
		})

		ginkgo.Context("authentication mode couplings", func() {

			ginkgo.It("should reject an unknown authentication type", func() {
				input := passwordUser()
				input.Spec.AuthenticationMode.Type = "no-password-required"
				input.Spec.AuthenticationMode.Passwords = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject the password type with zero passwords", func() {
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
				input := iamUser()
				input.Spec.AuthenticationMode.Passwords = []string{"a-very-strong-password"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
