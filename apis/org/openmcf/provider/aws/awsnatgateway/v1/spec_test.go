package awsnatgatewayv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsNatGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsNatGatewaySpec Validation Tests")
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

// minimalValidPublicNatGateway is the common case: a public gateway in a subnet
// with an Elastic IP.
func minimalValidPublicNatGateway() *AwsNatGateway {
	return &AwsNatGateway{
		ApiVersion: "aws.openmcf.org/v1",
		Kind:       "AwsNatGateway",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-nat-gateway",
		},
		Spec: &AwsNatGatewaySpec{
			Region:           "us-west-2",
			ConnectivityType: "public",
			SubnetId:         newStringValueOrRef("subnet-0abc123"),
			AllocationId:     newStringValueOrRef("eipalloc-0abc123"),
		},
	}
}

func minimalValidPrivateNatGateway() *AwsNatGateway {
	return &AwsNatGateway{
		ApiVersion: "aws.openmcf.org/v1",
		Kind:       "AwsNatGateway",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-private-nat-gateway",
		},
		Spec: &AwsNatGatewaySpec{
			Region:           "us-west-2",
			ConnectivityType: "private",
			SubnetId:         newStringValueOrRef("subnet-0abc123"),
		},
	}
}

var _ = ginkgo.Describe("AwsNatGatewaySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_nat_gateway", func() {

			ginkgo.It("should not return a validation error for a minimal public gateway", func() {
				err := protovalidate.Validate(minimalValidPublicNatGateway())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a minimal private gateway", func() {
				err := protovalidate.Validate(minimalValidPrivateNatGateway())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with subnet_id and allocation_id via value_from refs", func() {
				input := minimalValidPublicNatGateway()
				input.Spec.SubnetId = newValueFromRef("my-public-subnet")
				input.Spec.AllocationId = newValueFromRef("my-eip")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a public gateway with secondary allocations", func() {
				input := minimalValidPublicNatGateway()
				input.Spec.SecondaryAllocationIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("eipalloc-0def456"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a private gateway with a private_ip", func() {
				input := minimalValidPrivateNatGateway()
				input.Spec.PrivateIp = "10.0.1.10"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a private gateway with secondary_private_ip_address_count", func() {
				input := minimalValidPrivateNatGateway()
				input.Spec.SecondaryPrivateIpAddressCount = 2
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full metadata set", func() {
				input := minimalValidPublicNatGateway()
				input.Metadata = &shared.CloudResourceMetadata{
					Name: "full-nat-gateway",
					Org:  "acme-corp",
					Env:  "production",
					Labels: map[string]string{
						"team": "platform",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_nat_gateway", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidPublicNatGateway()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidPublicNatGateway()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidPublicNatGateway()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AwsNatGateway{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsNatGateway",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-nat-gateway"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty", func() {
				input := minimalValidPublicNatGateway()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when connectivity_type is empty", func() {
				input := minimalValidPublicNatGateway()
				input.Spec.ConnectivityType = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when connectivity_type is unknown", func() {
				input := minimalValidPublicNatGateway()
				input.Spec.ConnectivityType = "hybrid"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet_id is missing", func() {
				input := minimalValidPublicNatGateway()
				input.Spec.SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a public gateway has no allocation_id", func() {
				input := minimalValidPublicNatGateway()
				input.Spec.AllocationId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a private gateway has an allocation_id", func() {
				input := minimalValidPrivateNatGateway()
				input.Spec.AllocationId = newStringValueOrRef("eipalloc-0abc123")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a private gateway has secondary_allocation_ids", func() {
				input := minimalValidPrivateNatGateway()
				input.Spec.SecondaryAllocationIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("eipalloc-0def456"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a public gateway has a private_ip", func() {
				input := minimalValidPublicNatGateway()
				input.Spec.PrivateIp = "10.0.1.10"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a public gateway has a secondary_private_ip_address_count", func() {
				input := minimalValidPublicNatGateway()
				input.Spec.SecondaryPrivateIpAddressCount = 2
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both secondary private IP forms are set", func() {
				input := minimalValidPrivateNatGateway()
				input.Spec.SecondaryPrivateIpAddresses = []string{"10.0.1.11"}
				input.Spec.SecondaryPrivateIpAddressCount = 2
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
