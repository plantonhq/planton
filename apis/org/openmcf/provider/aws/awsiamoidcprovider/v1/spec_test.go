package awsiamoidcproviderv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAwsIamOidcProvider(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsIamOidcProvider Suite")
}

func newValidProvider() *AwsIamOidcProvider {
	return &AwsIamOidcProvider{
		ApiVersion: "aws.openmcf.org/v1",
		Kind:       "AwsIamOidcProvider",
		Metadata: &shared.CloudResourceMetadata{
			Name: "valid-name",
		},
		Spec: &AwsIamOidcProviderSpec{
			Region: "us-west-2",
			Url: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "https://token.actions.githubusercontent.com",
				},
			},
			ClientIdList: []string{"sts.amazonaws.com"},
		},
	}
}

var _ = ginkgo.Describe("AwsIamOidcProvider Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_iam_oidc_provider", func() {
			ginkgo.It("should not return a validation error for the minimal valid spec", func() {
				err := protovalidate.Validate(newValidProvider())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a well-formed 40-char thumbprint", func() {
				input := newValidProvider()
				input.Spec.ThumbprintList = []string{"990f4193972f2becf12ddeda5237f9c952f20d9e"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_iam_oidc_provider", func() {
			ginkgo.It("should reject a missing url", func() {
				input := newValidProvider()
				input.Spec.Url = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an empty client_id_list", func() {
				input := newValidProvider()
				input.Spec.ClientIdList = []string{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate client IDs", func() {
				input := newValidProvider()
				input.Spec.ClientIdList = []string{"sts.amazonaws.com", "sts.amazonaws.com"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a malformed thumbprint", func() {
				input := newValidProvider()
				input.Spec.ThumbprintList = []string{"not-a-valid-thumbprint"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an empty region", func() {
				input := newValidProvider()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
