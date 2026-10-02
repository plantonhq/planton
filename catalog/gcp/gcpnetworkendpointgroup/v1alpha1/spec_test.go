package gcpnetworkendpointgroupv1alpha1

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
	ginkgo.RunSpecs(t, "GcpNetworkEndpointGroupSpec Suite")
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

var _ = ginkgo.Describe("GcpNetworkEndpointGroupSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// A zonal VM group (the default GCE_VM_IP_PORT type) with one instance
	// endpoint.
	zonal := func() *GcpNetworkEndpointGroup {
		return &GcpNetworkEndpointGroup{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpNetworkEndpointGroup",
			Metadata:   &shared.CloudResourceMetadata{Name: "web-neg"},
			Spec: &GcpNetworkEndpointGroupSpec{
				Zone:        "us-central1-a",
				Network:     nameRef("main-vpc"),
				DefaultPort: proto.Int32(8080),
				Endpoints: []*GcpNetworkEndpoint{
					{Instance: nameRef("web-1"), IpAddress: "10.0.0.5", Port: proto.Int32(8080)},
				},
			},
		}
	}

	// A global internet FQDN group with one endpoint.
	global := func() *GcpNetworkEndpointGroup {
		return &GcpNetworkEndpointGroup{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpNetworkEndpointGroup",
			Metadata:   &shared.CloudResourceMetadata{Name: "origin-neg"},
			Spec: &GcpNetworkEndpointGroupSpec{
				NetworkEndpointType: "INTERNET_FQDN_PORT",
				DefaultPort:         proto.Int32(443),
				Endpoints:           []*GcpNetworkEndpoint{{Fqdn: "origin.example.com", Port: proto.Int32(443)}},
			},
		}
	}

	ginkgo.It("should accept the zonal VM group and the global FQDN group", func() {
		gomega.Expect(validator.Validate(zonal())).To(gomega.Succeed())
		gomega.Expect(validator.Validate(global())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a zonal group with no endpoints yet", func() {
		msg := zonal()
		msg.Spec.Endpoints = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a GCE_VM_IP group whose endpoints carry no port", func() {
		msg := zonal()
		msg.Spec.NetworkEndpointType = "GCE_VM_IP"
		msg.Spec.DefaultPort = nil
		msg.Spec.Endpoints[0].Port = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a port on a GCE_VM_IP group", func() {
		msg := zonal()
		msg.Spec.NetworkEndpointType = "GCE_VM_IP"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a hybrid group of addresses and reject an instance on it", func() {
		msg := zonal()
		msg.Spec.NetworkEndpointType = "NON_GCP_PRIVATE_IP_PORT"
		msg.Spec.Endpoints = []*GcpNetworkEndpoint{{IpAddress: "192.168.10.4", Port: proto.Int32(443)}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.Endpoints[0].Instance = nameRef("not-allowed")
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a VM endpoint without an instance", func() {
		msg := zonal()
		msg.Spec.Endpoints[0].Instance = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a hybrid or internet-IP endpoint without an address", func() {
		msg := zonal()
		msg.Spec.NetworkEndpointType = "INTERNET_IP_PORT"
		msg.Spec.Endpoints = []*GcpNetworkEndpoint{{Port: proto.Int32(443)}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should require a network zonally and reject one globally", func() {
		msg := zonal()
		msg.Spec.Network = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = global()
		msg.Spec.Network = litRef("https://www.googleapis.com/compute/v1/projects/p/global/networks/n")
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = global()
		msg.Spec.Subnetwork = nameRef("sub")
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a global group of a zonal-only type or with no type", func() {
		msg := global()
		msg.Spec.NetworkEndpointType = "GCE_VM_IP_PORT"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = global()
		msg.Spec.NetworkEndpointType = ""
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject SERVERLESS and PRIVATE_SERVICE_CONNECT (the regional kind's types)", func() {
		msg := zonal()
		msg.Spec.NetworkEndpointType = "SERVERLESS"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg.Spec.NetworkEndpointType = "PRIVATE_SERVICE_CONNECT"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should keep fqdn endpoints to FQDN groups and require them there", func() {
		msg := zonal()
		msg.Spec.Endpoints[0].Fqdn = "web.example.com"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = global()
		msg.Spec.Endpoints[0].Fqdn = ""
		msg.Spec.Endpoints[0].IpAddress = "203.0.113.10"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a global INTERNET_IP_PORT group and reject an endpoint with both address and fqdn", func() {
		msg := global()
		msg.Spec.NetworkEndpointType = "INTERNET_IP_PORT"
		msg.Spec.Endpoints = []*GcpNetworkEndpoint{{IpAddress: "203.0.113.10", Port: proto.Int32(443)}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.Endpoints[0].Fqdn = "also.example.com"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid zone, name, address, port, or deletion policy", func() {
		msg := zonal()
		msg.Spec.Zone = "us-central1"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = zonal()
		msg.Spec.NegName = "Web_NEG"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = zonal()
		msg.Spec.Endpoints[0].IpAddress = "10.0.0"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = zonal()
		msg.Spec.Endpoints[0].Port = proto.Int32(70000)
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = zonal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})
})
