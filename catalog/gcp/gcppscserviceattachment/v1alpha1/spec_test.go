package gcppscserviceattachmentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpPscServiceAttachmentSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: v}},
	}
}

var _ = ginkgo.Describe("GcpPscServiceAttachmentSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// A producer publishing an internal passthrough load balancer to every
	// consumer, the smallest valid attachment.
	minimal := func() *GcpPscServiceAttachment {
		return &GcpPscServiceAttachment{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpPscServiceAttachment",
			Metadata:   &shared.CloudResourceMetadata{Name: "orders-db-psc"},
			Spec: &GcpPscServiceAttachmentSpec{
				Region:               "us-central1",
				TargetService:        nameRef("orders-ilb"),
				NatSubnets:           []*foreignkeyv1.StringValueOrRef{nameRef("psc-nat-a")},
				ConnectionPreference: "ACCEPT_AUTOMATIC",
			},
		}
	}

	manual := func() *GcpPscServiceAttachment {
		msg := minimal()
		msg.Spec.ConnectionPreference = "ACCEPT_MANUAL"
		msg.Spec.ConsumerAcceptLists = []*GcpPscServiceAttachmentConsumer{
			{ProjectId: litRef("consumer-project"), ConnectionLimit: 10},
			{Network: nameRef("partner-vpc"), ConnectionLimit: 2},
		}
		return msg
	}

	ginkgo.It("should accept the minimal automatic-acceptance attachment", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a manual-acceptance attachment with project and network consumers", func() {
		gomega.Expect(validator.Validate(manual())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every optional lever together", func() {
		msg := manual()
		msg.Spec.ProjectId = litRef("producer-project")
		msg.Spec.AttachmentName = "orders-db-psc"
		msg.Spec.Description = "orders database over PSC"
		msg.Spec.NatSubnets = append(msg.Spec.NatSubnets, litRef("https://www.googleapis.com/compute/v1/projects/p/regions/us-central1/subnetworks/psc-nat-b"))
		msg.Spec.ConsumerRejectLists = []*foreignkeyv1.StringValueOrRef{litRef("blocked-project")}
		msg.Spec.ReconcileConnections = proto.Bool(true)
		msg.Spec.EnableProxyProtocol = true
		msg.Spec.DomainNames = []string{"p.mycompany.com."}
		msg.Spec.PropagatedConnectionLimit = proto.Int32(0)
		msg.Spec.ShowNatIps = true
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require region, target_service, nat_subnets, and connection_preference", func() {
		msg := minimal()
		msg.Spec.Region = ""
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.TargetService = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.NatSubnets = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.ConnectionPreference = ""
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid region, name, or connection preference", func() {
		msg := minimal()
		msg.Spec.Region = "US_Central1"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.AttachmentName = "Orders-DB"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.ConnectionPreference = "ACCEPT_ALL"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an accept list under ACCEPT_AUTOMATIC", func() {
		msg := manual()
		msg.Spec.ConnectionPreference = "ACCEPT_AUTOMATIC"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a consumer naming none or two of project, network, endpoint", func() {
		msg := manual()
		msg.Spec.ConsumerAcceptLists[0] = &GcpPscServiceAttachmentConsumer{ConnectionLimit: 1}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = manual()
		msg.Spec.ConsumerAcceptLists[0].Network = nameRef("also-a-network")
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an endpoint-url consumer and reject a zero connection limit", func() {
		msg := manual()
		msg.Spec.ConsumerAcceptLists = []*GcpPscServiceAttachmentConsumer{{EndpointUrl: "https://www.googleapis.com/compute/v1/projects/c/regions/us-central1/forwardingRules/ep", ConnectionLimit: 1}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.ConsumerAcceptLists[0].ConnectionLimit = 0
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject more than one domain name or one without a trailing dot", func() {
		msg := minimal()
		msg.Spec.DomainNames = []string{"a.example.com.", "b.example.com."}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.DomainNames = []string{"p.mycompany.com"}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a negative propagated connection limit and an invalid deletion policy", func() {
		msg := minimal()
		msg.Spec.PropagatedConnectionLimit = proto.Int32(-1)
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})
})
