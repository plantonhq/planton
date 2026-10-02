package gcpdatastreamprivateconnectionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpDatastreamPrivateConnectionSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpDatastreamPrivateConnectionSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	peering := func() *GcpDatastreamPrivateConnectionVpcPeeringConfig {
		return &GcpDatastreamPrivateConnectionVpcPeeringConfig{
			Vpc:    litRef("projects/p/global/networks/data"),
			Subnet: "10.200.0.0/29",
		}
	}
	base := func() *GcpDatastreamPrivateConnection {
		return &GcpDatastreamPrivateConnection{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpDatastreamPrivateConnection",
			Metadata:   &shared.CloudResourceMetadata{Name: "data-vpc"},
			Spec: &GcpDatastreamPrivateConnectionSpec{
				Location:         "us-central1",
				VpcPeeringConfig: peering(),
			},
		}
	}

	ginkgo.It("should accept VPC peering with every optional field", func() {
		r := base()
		r.Spec.ProjectId = litRef("p")
		r.Spec.PrivateConnectionId = "data-vpc"
		r.Spec.DisplayName = "Data VPC"
		r.Spec.CreateWithoutValidation = true
		r.Spec.Labels = map[string]string{"team": "data"}
		r.Spec.DeletionPolicy = "FORCE"
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a PSC interface alone", func() {
		r := base()
		r.Spec.VpcPeeringConfig = nil
		r.Spec.PscInterfaceConfig = &GcpDatastreamPrivateConnectionPscInterfaceConfig{
			NetworkAttachment: "projects/p/regions/us-central1/networkAttachments/datastream",
		}
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
	})

	ginkgo.It("should reject neither connectivity option", func() {
		r := base()
		r.Spec.VpcPeeringConfig = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject both connectivity options", func() {
		r := base()
		r.Spec.PscInterfaceConfig = &GcpDatastreamPrivateConnectionPscInterfaceConfig{
			NetworkAttachment: "projects/p/regions/us-central1/networkAttachments/datastream",
		}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a peering range that is not a /29", func() {
		for _, subnet := range []string{"10.200.0.0/28", "10.200.0.0", "not-a-cidr", "10.200.0.1/29"} {
			r := base()
			r.Spec.VpcPeeringConfig.Subnet = subnet
			gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), subnet)
		}
	})

	ginkgo.It("should reject a peering without a VPC", func() {
		r := base()
		r.Spec.VpcPeeringConfig.Vpc = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed network attachment", func() {
		r := base()
		r.Spec.VpcPeeringConfig = nil
		r.Spec.PscInterfaceConfig = &GcpDatastreamPrivateConnectionPscInterfaceConfig{NetworkAttachment: "datastream"}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a missing or malformed location", func() {
		for _, location := range []string{"", "US", "us-central1-a"} {
			r := base()
			r.Spec.Location = location
			gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), location)
		}
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		r := base()
		r.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})
})
