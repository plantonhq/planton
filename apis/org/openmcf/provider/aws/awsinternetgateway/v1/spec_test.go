package awsinternetgatewayv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsInternetGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsInternetGatewaySpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func newValueFromRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

func minimalValidInternetGateway() *AwsInternetGateway {
	return &AwsInternetGateway{
		ApiVersion: "aws.openmcf.org/v1",
		Kind:       "AwsInternetGateway",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-internet-gateway",
		},
		Spec: &AwsInternetGatewaySpec{
			Region: "us-west-2",
			VpcId:  newStringValueOrRef("vpc-0abc123"),
		},
	}
}

var _ = ginkgo.Describe("AwsInternetGatewaySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_internet_gateway", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(minimalValidInternetGateway())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with vpc_id via value_from ref", func() {
				input := minimalValidInternetGateway()
				input.Spec.VpcId = newValueFromRef("my-vpc")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full metadata set", func() {
				input := &AwsInternetGateway{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsInternetGateway",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-internet-gateway",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &AwsInternetGatewaySpec{
						Region: "us-west-2",
						VpcId:  newStringValueOrRef("vpc-0abc123"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_internet_gateway", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidInternetGateway()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidInternetGateway()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidInternetGateway()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AwsInternetGateway{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsInternetGateway",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-internet-gateway"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty", func() {
				input := minimalValidInternetGateway()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vpc_id is missing", func() {
				input := minimalValidInternetGateway()
				input.Spec.VpcId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
