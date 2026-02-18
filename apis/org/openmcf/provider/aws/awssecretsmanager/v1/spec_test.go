package awssecretsmanagerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAwsSecretsManagerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsSecretsManagerSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AwsSecretsManagerSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_secrets_manager", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AwsSecretsManager{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsSecretsManager",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-secrets-manager",
					},
					Spec: &AwsSecretsManagerSpec{
						Region:      "us-west-2",
						SecretNames: []string{"test-secret"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
