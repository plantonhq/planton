package hetznercloudfloatingipv1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHetznerCloudFloatingIpSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HetznerCloudFloatingIpSpec Validation Suite")
}

var _ = Describe("HetznerCloudFloatingIpSpec validations", func() {

	Context("with valid specs", func() {
		It("should accept a minimal IPv4 spec", func() {
			spec := &HetznerCloudFloatingIpSpec{
				Type:         HetznerCloudFloatingIpSpec_ipv4,
				HomeLocation: "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a minimal IPv6 spec", func() {
			spec := &HetznerCloudFloatingIpSpec{
				Type:         HetznerCloudFloatingIpSpec_ipv6,
				HomeLocation: "nbg1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with description", func() {
			spec := &HetznerCloudFloatingIpSpec{
				Type:         HetznerCloudFloatingIpSpec_ipv4,
				HomeLocation: "hel1",
				Description:  "Production web frontend failover IP",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with dns_ptr", func() {
			spec := &HetznerCloudFloatingIpSpec{
				Type:         HetznerCloudFloatingIpSpec_ipv4,
				HomeLocation: "fsn1",
				DnsPtr:       "mail.example.com",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with delete_protection enabled", func() {
			spec := &HetznerCloudFloatingIpSpec{
				Type:             HetznerCloudFloatingIpSpec_ipv4,
				HomeLocation:     "fsn1",
				DeleteProtection: true,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a fully populated spec without server_id", func() {
			spec := &HetznerCloudFloatingIpSpec{
				Type:             HetznerCloudFloatingIpSpec_ipv6,
				HomeLocation:     "ash",
				Description:      "HA cluster failover IP",
				DnsPtr:           "server.example.com",
				DeleteProtection: true,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("with invalid specs", func() {
		It("should reject ip_type_unspecified", func() {
			spec := &HetznerCloudFloatingIpSpec{
				Type:         HetznerCloudFloatingIpSpec_ip_type_unspecified,
				HomeLocation: "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a missing type (zero value)", func() {
			spec := &HetznerCloudFloatingIpSpec{
				HomeLocation: "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject an empty home_location", func() {
			spec := &HetznerCloudFloatingIpSpec{
				Type:         HetznerCloudFloatingIpSpec_ipv4,
				HomeLocation: "",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a missing home_location (zero value)", func() {
			spec := &HetznerCloudFloatingIpSpec{
				Type: HetznerCloudFloatingIpSpec_ipv4,
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})
	})
})
