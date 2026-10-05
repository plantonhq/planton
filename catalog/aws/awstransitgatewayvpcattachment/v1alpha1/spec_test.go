package awstransitgatewayvpcattachmentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAwsTransitGatewayVpcAttachmentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsTransitGatewayVpcAttachment Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

func validEnvelope(spec *AwsTransitGatewayVpcAttachmentSpec) *AwsTransitGatewayVpcAttachment {
	return &AwsTransitGatewayVpcAttachment{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsTransitGatewayVpcAttachment",
		Metadata:   &shared.CatalogObjectMetadata{Name: "test-attachment"},
		Spec:       spec,
	}
}

var _ = ginkgo.Describe("AwsTransitGatewayVpcAttachmentSpec validations", func() {
	var spec *AwsTransitGatewayVpcAttachmentSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: region + gateway + VPC + one subnet.
		spec = &AwsTransitGatewayVpcAttachmentSpec{
			Region:           "us-west-2",
			TransitGatewayId: strRef("tgw-0abc123"),
			VpcId:            strRef("vpc-0abc123"),
			SubnetIds: []*foreignkeyv1.StringValueOrRef{
				strRef("subnet-0abc123"),
			},
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal spec (region + gateway + vpc + one subnet)", func() {
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a fully-dialed multi-AZ attachment", func() {
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-0abc123"),
			strRef("subnet-0def456"),
			strRef("subnet-0ghi789"),
		}
		spec.DnsSupport = proto.Bool(false)
		spec.Ipv6Support = true
		spec.ApplianceModeSupport = true
		spec.SecurityGroupReferencingSupport = proto.Bool(true)
		spec.DefaultRouteTableAssociation = proto.Bool(false)
		spec.DefaultRouteTablePropagation = proto.Bool(false)
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Required field validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when region is empty", func() {
		spec.Region = ""
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when transit_gateway_id is missing", func() {
		spec.TransitGatewayId = nil
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when vpc_id is missing", func() {
		spec.VpcId = nil
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when subnet_ids is empty", func() {
		spec.SubnetIds = nil
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
