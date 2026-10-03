package digitaloceanfirewallv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestDigitalOceanFirewallSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanFirewallSpec Custom Validation Tests")
}

// strVal builds a StringValueOrRef carrying a literal value.
func strVal(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
	}
}

// validFirewall returns a minimal valid firewall the tests mutate per case.
func validFirewall() *DigitalOceanFirewall {
	return &DigitalOceanFirewall{
		ApiVersion: "digital-ocean.planton.dev/v1alpha1",
		Kind:       "DigitalOceanFirewall",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-firewall",
		},
		Spec: &DigitalOceanFirewallSpec{
			FirewallName: "web-firewall",
			InboundRules: []*DigitalOceanFirewallInboundRule{
				{
					Protocol:        "tcp",
					PortRange:       "80",
					SourceAddresses: []string{"0.0.0.0/0", "::/0"},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("DigitalOceanFirewallSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal firewall with one inbound rule", func() {
			gomega.Expect(protovalidate.Validate(validFirewall())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a firewall with only outbound rules", func() {
			input := validFirewall()
			input.Spec.InboundRules = nil
			input.Spec.OutboundRules = []*DigitalOceanFirewallOutboundRule{
				{
					Protocol:             "tcp",
					PortRange:            "all",
					DestinationAddresses: []string{"0.0.0.0/0"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an icmp rule with no port_range", func() {
			input := validFirewall()
			input.Spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:        "icmp",
					SourceAddresses: []string{"192.0.2.0/24"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts the literal 'all' and a port range", func() {
			input := validFirewall()
			input.Spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{Protocol: "udp", PortRange: "all", SourceAddresses: []string{"::/0"}},
				{Protocol: "tcp", PortRange: "8000-9000", SourceAddresses: []string{"10.0.0.0/8"}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts droplet targeting by reference and by literal id", func() {
			input := validFirewall()
			input.Spec.DropletIds = []*foreignkeyv1.StringValueOrRef{
				strVal("123456789"),
				{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-droplet"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts tag targeting with valid tag names", func() {
			input := validFirewall()
			input.Spec.Tags = []string{"web-tier", "env:production", "app_backend"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts rule-level droplet, kubernetes, and load balancer sources", func() {
			input := validFirewall()
			input.Spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:               "tcp",
					PortRange:              "443",
					SourceDropletIds:       []*foreignkeyv1.StringValueOrRef{strVal("123456")},
					SourceKubernetesIds:    []*foreignkeyv1.StringValueOrRef{strVal("2f9a1b2c-aaaa-bbbb-cccc-1234567890ab")},
					SourceLoadBalancerUids: []*foreignkeyv1.StringValueOrRef{strVal("4de7ac8b-495b-4884-9a69-1050c6793cd6")},
					SourceTags:             []string{"trusted"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts rule-level destinations on outbound rules", func() {
			input := validFirewall()
			input.Spec.OutboundRules = []*DigitalOceanFirewallOutboundRule{
				{
					Protocol:                    "udp",
					PortRange:                   "53",
					DestinationAddresses:        []string{"0.0.0.0/0"},
					DestinationDropletIds:       []*foreignkeyv1.StringValueOrRef{strVal("987654")},
					DestinationKubernetesIds:    []*foreignkeyv1.StringValueOrRef{strVal("3f9a1b2c-dddd-eeee-ffff-1234567890ab")},
					DestinationLoadBalancerUids: []*foreignkeyv1.StringValueOrRef{strVal("5de7ac8b-495b-4884-9a69-1050c6793cd7")},
					DestinationTags:             []string{"db-tier"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("rejects a missing firewall_name", func() {
			input := validFirewall()
			input.Spec.FirewallName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a firewall_name longer than 255 characters", func() {
			input := validFirewall()
			long := make([]byte, 256)
			for i := range long {
				long[i] = 'a'
			}
			input.Spec.FirewallName = string(long)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a firewall with no rules at all", func() {
			input := validFirewall()
			input.Spec.InboundRules = nil
			input.Spec.OutboundRules = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an unsupported protocol", func() {
			input := validFirewall()
			input.Spec.InboundRules[0].Protocol = "http"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing protocol", func() {
			input := validFirewall()
			input.Spec.InboundRules[0].Protocol = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a tcp inbound rule without port_range", func() {
			input := validFirewall()
			input.Spec.InboundRules[0].PortRange = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a udp outbound rule without port_range", func() {
			input := validFirewall()
			input.Spec.OutboundRules = []*DigitalOceanFirewallOutboundRule{
				{Protocol: "udp", DestinationAddresses: []string{"0.0.0.0/0"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an empty source address entry", func() {
			input := validFirewall()
			input.Spec.InboundRules[0].SourceAddresses = []string{""}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid firewall tag", func() {
			input := validFirewall()
			input.Spec.Tags = []string{"has space"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid rule-level source tag", func() {
			input := validFirewall()
			input.Spec.InboundRules[0].SourceTags = []string{"bad tag!"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
