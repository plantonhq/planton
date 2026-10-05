package awstransitgatewayv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/proto"
)

func TestAwsTransitGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsTransitGateway Validation Suite")
}

func validEnvelope(spec *AwsTransitGatewaySpec) *AwsTransitGateway {
	return &AwsTransitGateway{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsTransitGateway",
		Metadata:   &shared.CatalogObjectMetadata{Name: "test-tgw"},
		Spec:       spec,
	}
}

var _ = ginkgo.Describe("AwsTransitGatewaySpec validations", func() {
	var spec *AwsTransitGatewaySpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: the gateway is a pure hub, so region alone is
		// a deployable manifest (every dial has an AWS-side default).
		spec = &AwsTransitGatewaySpec{
			Region: "us-west-2",
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal spec (region only)", func() {
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a fully-dialed spec", func() {
		spec.Description = "hub for the shared-services topology"
		spec.AmazonSideAsn = 64513
		spec.DefaultRouteTableAssociation = proto.Bool(false)
		spec.DefaultRouteTablePropagation = proto.Bool(false)
		spec.DnsSupport = proto.Bool(true)
		spec.VpnEcmpSupport = proto.Bool(false)
		spec.AutoAcceptSharedAttachments = true
		spec.SecurityGroupReferencingSupport = true
		spec.MulticastSupport = true
		spec.EncryptionSupport = proto.Bool(true)
		spec.TransitGatewayCidrBlocks = []string{"10.255.0.0/24", "fd00:1::/64"}
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

	// -------------------------------------------------------------------------
	// amazon_side_asn range
	// -------------------------------------------------------------------------

	ginkgo.It("accepts the 16-bit private ASN bounds", func() {
		spec.AmazonSideAsn = 64512
		gomega.Expect(protovalidate.Validate(validEnvelope(spec))).To(gomega.BeNil())
		spec.AmazonSideAsn = 65534
		gomega.Expect(protovalidate.Validate(validEnvelope(spec))).To(gomega.BeNil())
	})

	ginkgo.It("accepts the 32-bit private ASN bounds", func() {
		spec.AmazonSideAsn = 4200000000
		gomega.Expect(protovalidate.Validate(validEnvelope(spec))).To(gomega.BeNil())
		spec.AmazonSideAsn = 4294967294
		gomega.Expect(protovalidate.Validate(validEnvelope(spec))).To(gomega.BeNil())
	})

	ginkgo.It("fails on an ASN below the 16-bit private range", func() {
		spec.AmazonSideAsn = 64511
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an ASN between the private ranges", func() {
		spec.AmazonSideAsn = 65535
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an ASN above the 32-bit private range", func() {
		spec.AmazonSideAsn = 4294967295
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// transit_gateway_cidr_blocks
	// -------------------------------------------------------------------------

	ginkgo.It("fails on more than 5 CIDR blocks", func() {
		spec.TransitGatewayCidrBlocks = []string{
			"10.0.0.0/24", "10.1.0.0/24", "10.2.0.0/24",
			"10.3.0.0/24", "10.4.0.0/24", "10.5.0.0/24",
		}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an IPv4 block smaller than /24", func() {
		spec.TransitGatewayCidrBlocks = []string{"10.0.0.0/25"}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on an IPv6 block smaller than /64", func() {
		spec.TransitGatewayCidrBlocks = []string{"fd00:1::/80"}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on a link-local CIDR block", func() {
		spec.TransitGatewayCidrBlocks = []string{"169.254.10.0/24"}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on a CIDR block without a prefix length", func() {
		spec.TransitGatewayCidrBlocks = []string{"10.0.0.0"}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
