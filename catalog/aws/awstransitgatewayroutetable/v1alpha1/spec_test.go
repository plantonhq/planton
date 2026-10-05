package awstransitgatewayroutetablev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsTransitGatewayRouteTableSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsTransitGatewayRouteTable Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

func validEnvelope(spec *AwsTransitGatewayRouteTableSpec) *AwsTransitGatewayRouteTable {
	return &AwsTransitGatewayRouteTable{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsTransitGatewayRouteTable",
		Metadata:   &shared.CatalogObjectMetadata{Name: "test-route-table"},
		Spec:       spec,
	}
}

var _ = ginkgo.Describe("AwsTransitGatewayRouteTableSpec validations", func() {
	var spec *AwsTransitGatewayRouteTableSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: region + gateway. A route table can exist empty
		// -- membership and routes are added as the domain grows.
		spec = &AwsTransitGatewayRouteTableSpec{
			Region:           "us-west-2",
			TransitGatewayId: strRef("tgw-0abc123"),
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal spec (region + gateway, empty table)", func() {
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a full routing domain (associations, propagations, routes, prefix lists)", func() {
		spec.Associations = []*AwsTransitGatewayRouteTableAssociation{
			{AttachmentId: strRef("tgw-attach-0aaa111")},
		}
		spec.Propagations = []*foreignkeyv1.StringValueOrRef{
			strRef("tgw-attach-0aaa111"),
			strRef("tgw-attach-0bbb222"),
		}
		spec.Routes = []*AwsTransitGatewayRouteTableRoute{
			{DestinationCidrBlock: "0.0.0.0/0", AttachmentId: strRef("tgw-attach-0bbb222")},
			{DestinationCidrBlock: "10.99.0.0/16", Blackhole: true},
		}
		spec.PrefixListReferences = []*AwsTransitGatewayRouteTablePrefixListReference{
			{PrefixListId: "pl-0abc123", AttachmentId: strRef("tgw-attach-0aaa111")},
		}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an association that takes over an existing association", func() {
		spec.Associations = []*AwsTransitGatewayRouteTableAssociation{
			{AttachmentId: strRef("tgw-attach-0aaa111"), ReplaceExistingAssociation: true},
		}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts both gateway default-table designations on one table", func() {
		spec.SetAsDefaultAssociationTable = true
		spec.SetAsDefaultPropagationTable = true
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

	ginkgo.It("fails an association without an attachment", func() {
		spec.Associations = []*AwsTransitGatewayRouteTableAssociation{
			{ReplaceExistingAssociation: true},
		}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Static route target: exactly one of attachment_id / blackhole
	// -------------------------------------------------------------------------

	ginkgo.It("fails a route with neither attachment nor blackhole", func() {
		spec.Routes = []*AwsTransitGatewayRouteTableRoute{
			{DestinationCidrBlock: "10.1.0.0/16"},
		}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails a route with both attachment and blackhole", func() {
		spec.Routes = []*AwsTransitGatewayRouteTableRoute{
			{DestinationCidrBlock: "10.1.0.0/16", AttachmentId: strRef("tgw-attach-0aaa111"), Blackhole: true},
		}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails a route with an empty destination", func() {
		spec.Routes = []*AwsTransitGatewayRouteTableRoute{
			{DestinationCidrBlock: "", Blackhole: true},
		}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on duplicate route destinations", func() {
		spec.Routes = []*AwsTransitGatewayRouteTableRoute{
			{DestinationCidrBlock: "10.1.0.0/16", Blackhole: true},
			{DestinationCidrBlock: "10.1.0.0/16", AttachmentId: strRef("tgw-attach-0aaa111")},
		}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Prefix list references
	// -------------------------------------------------------------------------

	ginkgo.It("fails a prefix list reference with neither attachment nor blackhole", func() {
		spec.PrefixListReferences = []*AwsTransitGatewayRouteTablePrefixListReference{
			{PrefixListId: "pl-0abc123"},
		}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails a prefix list reference with both attachment and blackhole", func() {
		spec.PrefixListReferences = []*AwsTransitGatewayRouteTablePrefixListReference{
			{PrefixListId: "pl-0abc123", AttachmentId: strRef("tgw-attach-0aaa111"), Blackhole: true},
		}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on duplicate prefix list IDs", func() {
		spec.PrefixListReferences = []*AwsTransitGatewayRouteTablePrefixListReference{
			{PrefixListId: "pl-0abc123", Blackhole: true},
			{PrefixListId: "pl-0abc123", AttachmentId: strRef("tgw-attach-0aaa111")},
		}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails a prefix list reference with an empty prefix list ID", func() {
		spec.PrefixListReferences = []*AwsTransitGatewayRouteTablePrefixListReference{
			{PrefixListId: "", Blackhole: true},
		}
		err := protovalidate.Validate(validEnvelope(spec))
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
