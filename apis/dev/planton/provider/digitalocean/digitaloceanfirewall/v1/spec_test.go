package digitaloceanfirewallv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestDigitalOceanFirewallSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanFirewallSpec Validation Suite")
}

var _ = ginkgo.Describe("DigitalOceanFirewallSpec validations", func() {

	// Helper function to create a minimal valid firewall spec with tag-based targeting
	makeValidTagBasedSpec := func() *DigitalOceanFirewallSpec {
		return &DigitalOceanFirewallSpec{
			Name: "prod-web-firewall",
			Tags: []string{"web-tier"},
			InboundRules: []*DigitalOceanFirewallInboundRule{
				{
					Protocol:        "tcp",
					PortRange:       "443",
					SourceAddresses: []string{"0.0.0.0/0", "::/0"},
				},
			},
			OutboundRules: []*DigitalOceanFirewallOutboundRule{
				{
					Protocol:             "tcp",
					PortRange:            "443",
					DestinationAddresses: []string{"0.0.0.0/0", "::/0"},
				},
			},
		}
	}

	// Helper function to create a firewall spec with droplet IDs
	makeValidDropletIdSpec := func() *DigitalOceanFirewallSpec {
		return &DigitalOceanFirewallSpec{
			Name:       "dev-test-firewall",
			DropletIds: []int64{386734086},
			InboundRules: []*DigitalOceanFirewallInboundRule{
				{
					Protocol:        "tcp",
					PortRange:       "22",
					SourceAddresses: []string{"203.0.113.0/24"},
				},
			},
			OutboundRules: []*DigitalOceanFirewallOutboundRule{
				{
					Protocol:             "tcp",
					PortRange:            "1-65535",
					DestinationAddresses: []string{"0.0.0.0/0"},
				},
			},
		}
	}

	ginkgo.Context("Required fields", func() {
		ginkgo.It("accepts a minimal valid tag-based firewall spec", func() {
			spec := makeValidTagBasedSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a minimal valid droplet ID-based firewall spec", func() {
			spec := makeValidDropletIdSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing name", func() {
			spec := makeValidTagBasedSpec()
			spec.Name = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("name validation", func() {
		ginkgo.It("accepts name with valid characters (alphanumeric, hyphens)", func() {
			spec := makeValidTagBasedSpec()
			spec.Name = "prod-web-firewall-2025-v1"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts name with 255 characters (max)", func() {
			spec := makeValidTagBasedSpec()
			// Create a 255-character string
			spec.Name = string(make([]byte, 255))
			for i := range spec.Name {
				spec.Name = spec.Name[:i] + "a" + spec.Name[i+1:]
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects name exceeding 255 characters", func() {
			spec := makeValidTagBasedSpec()
			// Create a 256-character string
			spec.Name = string(make([]byte, 256))
			for i := range spec.Name {
				spec.Name = spec.Name[:i] + "a" + spec.Name[i+1:]
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects empty name", func() {
			spec := makeValidTagBasedSpec()
			spec.Name = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Inbound rules validation", func() {
		ginkgo.It("accepts firewall with no inbound rules (deny all inbound)", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts inbound rule with TCP protocol and port range", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:        "tcp",
					PortRange:       "8000-9000",
					SourceAddresses: []string{"192.168.1.0/24"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts inbound rule with UDP protocol", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:        "udp",
					PortRange:       "53",
					SourceAddresses: []string{"0.0.0.0/0"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts inbound rule with ICMP protocol (no port range)", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:        "icmp",
					SourceAddresses: []string{"0.0.0.0/0"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts inbound rule with source tags", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:   "tcp",
					PortRange:  "5432",
					SourceTags: []string{"web-tier", "api-tier"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts inbound rule with source droplet IDs", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:         "tcp",
					PortRange:        "3306",
					SourceDropletIds: []int64{386734086, 386734087},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts inbound rule with source Kubernetes cluster IDs", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:            "tcp",
					PortRange:           "443",
					SourceKubernetesIds: []string{"k8s-1-28-abc123"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts inbound rule with source Load Balancer UIDs", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:               "tcp",
					PortRange:              "8080",
					SourceLoadBalancerUids: []string{"lb-abc123"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts inbound rule with IPv6 CIDR", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:        "tcp",
					PortRange:       "443",
					SourceAddresses: []string{"::/0"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts inbound rule with multiple source types", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:        "tcp",
					PortRange:       "22",
					SourceAddresses: []string{"203.0.113.0/24"},
					SourceTags:      []string{"bastion"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects inbound rule with missing protocol", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:        "",
					PortRange:       "443",
					SourceAddresses: []string{"0.0.0.0/0"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Outbound rules validation", func() {
		ginkgo.It("accepts firewall with no outbound rules (deny all outbound)", func() {
			spec := makeValidTagBasedSpec()
			spec.OutboundRules = []*DigitalOceanFirewallOutboundRule{}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts outbound rule with TCP protocol and port range", func() {
			spec := makeValidTagBasedSpec()
			spec.OutboundRules = []*DigitalOceanFirewallOutboundRule{
				{
					Protocol:             "tcp",
					PortRange:            "5432",
					DestinationAddresses: []string{"10.0.0.0/8"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts outbound rule with destination tags", func() {
			spec := makeValidTagBasedSpec()
			spec.OutboundRules = []*DigitalOceanFirewallOutboundRule{
				{
					Protocol:        "tcp",
					PortRange:       "5432",
					DestinationTags: []string{"db-tier"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts outbound rule with destination droplet IDs", func() {
			spec := makeValidTagBasedSpec()
			spec.OutboundRules = []*DigitalOceanFirewallOutboundRule{
				{
					Protocol:              "tcp",
					PortRange:             "3306",
					DestinationDropletIds: []int64{386734088},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts outbound rule with destination Kubernetes cluster IDs", func() {
			spec := makeValidTagBasedSpec()
			spec.OutboundRules = []*DigitalOceanFirewallOutboundRule{
				{
					Protocol:                 "tcp",
					PortRange:                "6443",
					DestinationKubernetesIds: []string{"k8s-1-28-xyz789"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts outbound rule with destination Load Balancer UIDs", func() {
			spec := makeValidTagBasedSpec()
			spec.OutboundRules = []*DigitalOceanFirewallOutboundRule{
				{
					Protocol:                    "tcp",
					PortRange:                   "443",
					DestinationLoadBalancerUids: []string{"lb-xyz789"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts outbound rule allowing all ports (1-65535)", func() {
			spec := makeValidTagBasedSpec()
			spec.OutboundRules = []*DigitalOceanFirewallOutboundRule{
				{
					Protocol:             "tcp",
					PortRange:            "1-65535",
					DestinationAddresses: []string{"0.0.0.0/0"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts outbound rule with ICMP protocol (no port range)", func() {
			spec := makeValidTagBasedSpec()
			spec.OutboundRules = []*DigitalOceanFirewallOutboundRule{
				{
					Protocol:             "icmp",
					DestinationAddresses: []string{"0.0.0.0/0"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects outbound rule with missing protocol", func() {
			spec := makeValidTagBasedSpec()
			spec.OutboundRules = []*DigitalOceanFirewallOutboundRule{
				{
					Protocol:             "",
					PortRange:            "443",
					DestinationAddresses: []string{"0.0.0.0/0"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Target assignment validation", func() {
		ginkgo.It("accepts firewall with only tags (production pattern)", func() {
			spec := makeValidTagBasedSpec()
			spec.DropletIds = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts firewall with only droplet IDs (dev pattern)", func() {
			spec := makeValidDropletIdSpec()
			spec.Tags = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts firewall with both tags and droplet IDs", func() {
			spec := makeValidTagBasedSpec()
			spec.DropletIds = []int64{386734086}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts firewall with multiple tags", func() {
			spec := makeValidTagBasedSpec()
			spec.Tags = []string{"web-tier", "prod", "us-east"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts firewall with up to 10 droplet IDs", func() {
			spec := makeValidDropletIdSpec()
			spec.DropletIds = []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts firewall with neither tags nor droplet IDs (unassigned firewall)", func() {
			spec := &DigitalOceanFirewallSpec{
				Name: "unassigned-firewall",
				InboundRules: []*DigitalOceanFirewallInboundRule{
					{
						Protocol:        "tcp",
						PortRange:       "443",
						SourceAddresses: []string{"0.0.0.0/0"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Multi-tier architecture patterns", func() {
		ginkgo.It("accepts web tier firewall with Load Balancer source", func() {
			spec := &DigitalOceanFirewallSpec{
				Name: "web-tier-firewall",
				Tags: []string{"web-tier"},
				InboundRules: []*DigitalOceanFirewallInboundRule{
					{
						Protocol:               "tcp",
						PortRange:              "443",
						SourceLoadBalancerUids: []string{"lb-abc123"},
					},
					{
						Protocol:        "tcp",
						PortRange:       "22",
						SourceAddresses: []string{"203.0.113.10/32"},
					},
				},
				OutboundRules: []*DigitalOceanFirewallOutboundRule{
					{
						Protocol:        "tcp",
						PortRange:       "5432",
						DestinationTags: []string{"db-tier"},
					},
					{
						Protocol:             "tcp",
						PortRange:            "443",
						DestinationAddresses: []string{"0.0.0.0/0"},
					},
					{
						Protocol:             "udp",
						PortRange:            "53",
						DestinationAddresses: []string{"0.0.0.0/0"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts database tier firewall with restricted access", func() {
			spec := &DigitalOceanFirewallSpec{
				Name: "db-tier-firewall",
				Tags: []string{"db-tier"},
				InboundRules: []*DigitalOceanFirewallInboundRule{
					{
						Protocol:   "tcp",
						PortRange:  "5432",
						SourceTags: []string{"web-tier"},
					},
					{
						Protocol:        "tcp",
						PortRange:       "22",
						SourceAddresses: []string{"203.0.113.10/32"},
					},
				},
				OutboundRules: []*DigitalOceanFirewallOutboundRule{
					{
						Protocol:             "tcp",
						PortRange:            "443",
						DestinationAddresses: []string{"91.189.88.0/21"},
					},
					{
						Protocol:             "udp",
						PortRange:            "53",
						DestinationAddresses: []string{"1.1.1.1/32"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Edge cases", func() {
		ginkgo.It("accepts firewall with single port", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:        "tcp",
					PortRange:       "80",
					SourceAddresses: []string{"0.0.0.0/0"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts firewall with port range", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:        "tcp",
					PortRange:       "8000-9000",
					SourceAddresses: []string{"0.0.0.0/0"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts firewall with specific /32 CIDR (single IP)", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:        "tcp",
					PortRange:       "22",
					SourceAddresses: []string{"203.0.113.10/32"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts firewall with both IPv4 and IPv6 sources", func() {
			spec := makeValidTagBasedSpec()
			spec.InboundRules = []*DigitalOceanFirewallInboundRule{
				{
					Protocol:        "tcp",
					PortRange:       "443",
					SourceAddresses: []string{"0.0.0.0/0", "::/0"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts firewall name with underscores", func() {
			spec := makeValidTagBasedSpec()
			spec.Name = "prod_web_firewall_v2"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})
