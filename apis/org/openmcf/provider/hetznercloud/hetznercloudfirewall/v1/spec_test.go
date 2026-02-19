package hetznercloudfirewallv1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHetznerCloudFirewallSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HetznerCloudFirewallSpec Validation Suite")
}

var _ = Describe("HetznerCloudFirewallSpec validations", func() {

	Context("with valid specs", func() {
		It("should accept a single inbound TCP rule", func() {
			dirIn := HetznerCloudFirewallSpec_Rule_in
			protoTcp := HetznerCloudFirewallSpec_Rule_tcp
			spec := &HetznerCloudFirewallSpec{
				Rules: []*HetznerCloudFirewallSpec_Rule{
					{
						Direction: dirIn,
						Protocol:  protoTcp,
						Port:      "22",
						SourceIps: []string{"0.0.0.0/0", "::/0"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept an outbound UDP rule", func() {
			dirOut := HetznerCloudFirewallSpec_Rule_out
			protoUdp := HetznerCloudFirewallSpec_Rule_udp
			spec := &HetznerCloudFirewallSpec{
				Rules: []*HetznerCloudFirewallSpec_Rule{
					{
						Direction:      dirOut,
						Protocol:       protoUdp,
						Port:           "53",
						DestinationIps: []string{"0.0.0.0/0", "::/0"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept an ICMP rule without port", func() {
			dirIn := HetznerCloudFirewallSpec_Rule_in
			protoIcmp := HetznerCloudFirewallSpec_Rule_icmp
			spec := &HetznerCloudFirewallSpec{
				Rules: []*HetznerCloudFirewallSpec_Rule{
					{
						Direction: dirIn,
						Protocol:  protoIcmp,
						SourceIps: []string{"0.0.0.0/0", "::/0"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a TCP rule with a port range", func() {
			dirIn := HetznerCloudFirewallSpec_Rule_in
			protoTcp := HetznerCloudFirewallSpec_Rule_tcp
			spec := &HetznerCloudFirewallSpec{
				Rules: []*HetznerCloudFirewallSpec_Rule{
					{
						Direction:   dirIn,
						Protocol:    protoTcp,
						Port:        "80-443",
						SourceIps:   []string{"10.0.0.0/8"},
						Description: "allow HTTP and HTTPS from private network",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept multiple mixed rules", func() {
			dirIn := HetznerCloudFirewallSpec_Rule_in
			dirOut := HetznerCloudFirewallSpec_Rule_out
			protoTcp := HetznerCloudFirewallSpec_Rule_tcp
			protoIcmp := HetznerCloudFirewallSpec_Rule_icmp
			spec := &HetznerCloudFirewallSpec{
				Rules: []*HetznerCloudFirewallSpec_Rule{
					{
						Direction:   dirIn,
						Protocol:    protoTcp,
						Port:        "22",
						SourceIps:   []string{"0.0.0.0/0", "::/0"},
						Description: "allow SSH",
					},
					{
						Direction:   dirIn,
						Protocol:    protoTcp,
						Port:        "80",
						SourceIps:   []string{"0.0.0.0/0", "::/0"},
						Description: "allow HTTP",
					},
					{
						Direction: dirIn,
						Protocol:  protoIcmp,
						SourceIps: []string{"0.0.0.0/0", "::/0"},
					},
					{
						Direction:      dirOut,
						Protocol:       protoTcp,
						Port:           "443",
						DestinationIps: []string{"0.0.0.0/0", "::/0"},
						Description:    "allow outbound HTTPS",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept an empty rules list", func() {
			spec := &HetznerCloudFirewallSpec{}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept an ESP rule without port", func() {
			dirIn := HetznerCloudFirewallSpec_Rule_in
			protoEsp := HetznerCloudFirewallSpec_Rule_esp
			spec := &HetznerCloudFirewallSpec{
				Rules: []*HetznerCloudFirewallSpec_Rule{
					{
						Direction:   dirIn,
						Protocol:    protoEsp,
						SourceIps:   []string{"10.0.0.0/8"},
						Description: "allow IPSec ESP",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a GRE rule without port", func() {
			dirIn := HetznerCloudFirewallSpec_Rule_in
			protoGre := HetznerCloudFirewallSpec_Rule_gre
			spec := &HetznerCloudFirewallSpec{
				Rules: []*HetznerCloudFirewallSpec_Rule{
					{
						Direction: dirIn,
						Protocol:  protoGre,
						SourceIps: []string{"10.0.0.0/8"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("with invalid specs", func() {
		It("should reject a rule with direction_unspecified", func() {
			protoTcp := HetznerCloudFirewallSpec_Rule_tcp
			spec := &HetznerCloudFirewallSpec{
				Rules: []*HetznerCloudFirewallSpec_Rule{
					{
						Protocol:  protoTcp,
						Port:      "80",
						SourceIps: []string{"0.0.0.0/0"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a rule with protocol_unspecified", func() {
			dirIn := HetznerCloudFirewallSpec_Rule_in
			spec := &HetznerCloudFirewallSpec{
				Rules: []*HetznerCloudFirewallSpec_Rule{
					{
						Direction: dirIn,
						Port:      "80",
						SourceIps: []string{"0.0.0.0/0"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a TCP rule without port", func() {
			dirIn := HetznerCloudFirewallSpec_Rule_in
			protoTcp := HetznerCloudFirewallSpec_Rule_tcp
			spec := &HetznerCloudFirewallSpec{
				Rules: []*HetznerCloudFirewallSpec_Rule{
					{
						Direction: dirIn,
						Protocol:  protoTcp,
						SourceIps: []string{"0.0.0.0/0"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a UDP rule without port", func() {
			dirIn := HetznerCloudFirewallSpec_Rule_in
			protoUdp := HetznerCloudFirewallSpec_Rule_udp
			spec := &HetznerCloudFirewallSpec{
				Rules: []*HetznerCloudFirewallSpec_Rule{
					{
						Direction: dirIn,
						Protocol:  protoUdp,
						SourceIps: []string{"0.0.0.0/0"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject an inbound rule without source_ips", func() {
			dirIn := HetznerCloudFirewallSpec_Rule_in
			protoTcp := HetznerCloudFirewallSpec_Rule_tcp
			spec := &HetznerCloudFirewallSpec{
				Rules: []*HetznerCloudFirewallSpec_Rule{
					{
						Direction: dirIn,
						Protocol:  protoTcp,
						Port:      "80",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject an outbound rule without destination_ips", func() {
			dirOut := HetznerCloudFirewallSpec_Rule_out
			protoTcp := HetznerCloudFirewallSpec_Rule_tcp
			spec := &HetznerCloudFirewallSpec{
				Rules: []*HetznerCloudFirewallSpec_Rule{
					{
						Direction: dirOut,
						Protocol:  protoTcp,
						Port:      "443",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})
	})
})
