package hetznercloudprimaryipv1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHetznerCloudPrimaryIpSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HetznerCloudPrimaryIpSpec Validation Suite")
}

var _ = Describe("HetznerCloudPrimaryIpSpec validations", func() {

	Context("with valid specs", func() {
		It("should accept a minimal IPv4 spec", func() {
			spec := &HetznerCloudPrimaryIpSpec{
				Type:     HetznerCloudPrimaryIpSpec_ipv4,
				Location: "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a minimal IPv6 spec", func() {
			spec := &HetznerCloudPrimaryIpSpec{
				Type:     HetznerCloudPrimaryIpSpec_ipv6,
				Location: "nbg1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept an IPv4 spec with dns_ptr", func() {
			spec := &HetznerCloudPrimaryIpSpec{
				Type:     HetznerCloudPrimaryIpSpec_ipv4,
				Location: "hel1",
				DnsPtr:   "mail.example.com",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept an IPv6 spec with dns_ptr", func() {
			spec := &HetznerCloudPrimaryIpSpec{
				Type:     HetznerCloudPrimaryIpSpec_ipv6,
				Location: "ash",
				DnsPtr:   "web.example.com",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with delete_protection enabled", func() {
			spec := &HetznerCloudPrimaryIpSpec{
				Type:             HetznerCloudPrimaryIpSpec_ipv4,
				Location:         "fsn1",
				DeleteProtection: true,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a fully populated spec", func() {
			spec := &HetznerCloudPrimaryIpSpec{
				Type:             HetznerCloudPrimaryIpSpec_ipv6,
				Location:         "hel1",
				DnsPtr:           "server.example.com",
				DeleteProtection: true,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("with invalid specs", func() {
		It("should reject ip_type_unspecified", func() {
			spec := &HetznerCloudPrimaryIpSpec{
				Type:     HetznerCloudPrimaryIpSpec_ip_type_unspecified,
				Location: "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a missing type (zero value)", func() {
			spec := &HetznerCloudPrimaryIpSpec{
				Location: "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject an empty location", func() {
			spec := &HetznerCloudPrimaryIpSpec{
				Type:     HetznerCloudPrimaryIpSpec_ipv4,
				Location: "",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a missing location (zero value)", func() {
			spec := &HetznerCloudPrimaryIpSpec{
				Type: HetznerCloudPrimaryIpSpec_ipv4,
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})
	})
})
