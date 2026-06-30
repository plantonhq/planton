package civofirewallv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestCivoFirewallSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CivoFirewallSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CivoFirewallSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("minimal valid firewall", func() {
			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "test-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("firewall with inbound rules", func() {
			ginkgo.It("should accept single TCP inbound rule", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "web-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "web-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-456",
							},
						},
						InboundRules: []*CivoFirewallInboundRule{
							{
								Protocol:  "tcp",
								PortRange: "80",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "HTTP",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept multiple inbound rules with different protocols", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "multi-rule-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "multi-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-789",
							},
						},
						InboundRules: []*CivoFirewallInboundRule{
							{
								Protocol:  "tcp",
								PortRange: "22",
								Cidrs:     []string{"203.0.113.10/32"},
								Action:    "allow",
								Label:     "SSH from office",
							},
							{
								Protocol:  "tcp",
								PortRange: "80",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "HTTP",
							},
							{
								Protocol:  "tcp",
								PortRange: "443",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "HTTPS",
							},
							{
								Protocol:  "icmp",
								PortRange: "",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "Ping",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept port ranges", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "range-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "range-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-abc",
							},
						},
						InboundRules: []*CivoFirewallInboundRule{
							{
								Protocol:  "tcp",
								PortRange: "8000-9000",
								Cidrs:     []string{"10.0.0.0/8"},
								Action:    "allow",
								Label:     "Application port range",
							},
							{
								Protocol:  "udp",
								PortRange: "30000-32767",
								Cidrs:     []string{"192.168.0.0/16"},
								Action:    "allow",
								Label:     "Kubernetes NodePort range",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept UDP protocol", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "udp-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "udp-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-def",
							},
						},
						InboundRules: []*CivoFirewallInboundRule{
							{
								Protocol:  "udp",
								PortRange: "53",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "DNS",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept ICMP protocol", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "icmp-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "icmp-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-ghi",
							},
						},
						InboundRules: []*CivoFirewallInboundRule{
							{
								Protocol:  "icmp",
								PortRange: "",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "Allow ping",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept multiple CIDR blocks", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "multi-cidr-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "multi-cidr-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-jkl",
							},
						},
						InboundRules: []*CivoFirewallInboundRule{
							{
								Protocol:  "tcp",
								PortRange: "22",
								Cidrs:     []string{"203.0.113.10/32", "198.51.100.0/24", "192.0.2.0/24"},
								Action:    "allow",
								Label:     "SSH from multiple offices",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("firewall with outbound rules", func() {
			ginkgo.It("should accept single outbound rule", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "outbound-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "outbound-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-mno",
							},
						},
						OutboundRules: []*CivoFirewallOutboundRule{
							{
								Protocol:  "tcp",
								PortRange: "443",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "HTTPS outbound",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept both inbound and outbound rules", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "bidirectional-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "bidirectional-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-pqr",
							},
						},
						InboundRules: []*CivoFirewallInboundRule{
							{
								Protocol:  "tcp",
								PortRange: "80",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "HTTP inbound",
							},
						},
						OutboundRules: []*CivoFirewallOutboundRule{
							{
								Protocol:  "tcp",
								PortRange: "443",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "HTTPS outbound",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("firewall with tags", func() {
			ginkgo.It("should accept instance tags", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "tagged-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "tagged-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-stu",
							},
						},
						Tags: []string{"web-server", "production", "tier-frontend"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("real-world firewall patterns", func() {
			ginkgo.It("should accept web server firewall pattern", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "web-server-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "web-server-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "vpc-web-123",
							},
						},
						InboundRules: []*CivoFirewallInboundRule{
							{
								Protocol:  "tcp",
								PortRange: "22",
								Cidrs:     []string{"203.0.113.10/32"},
								Action:    "allow",
								Label:     "SSH from office",
							},
							{
								Protocol:  "tcp",
								PortRange: "80",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "HTTP",
							},
							{
								Protocol:  "tcp",
								PortRange: "443",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "HTTPS",
							},
						},
						Tags: []string{"web-server"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept database firewall pattern", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "database-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "database-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "vpc-db-456",
							},
						},
						InboundRules: []*CivoFirewallInboundRule{
							{
								Protocol:  "tcp",
								PortRange: "5432",
								Cidrs:     []string{"10.0.1.0/24"},
								Action:    "allow",
								Label:     "PostgreSQL from app tier",
							},
							{
								Protocol:  "tcp",
								PortRange: "22",
								Cidrs:     []string{"203.0.113.10/32"},
								Action:    "allow",
								Label:     "SSH from bastion",
							},
						},
						Tags: []string{"database", "postgresql"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept Kubernetes cluster firewall pattern", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "k8s-cluster-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "k8s-cluster-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "vpc-k8s-789",
							},
						},
						InboundRules: []*CivoFirewallInboundRule{
							{
								Protocol:  "tcp",
								PortRange: "6443",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "Kubernetes API server",
							},
							{
								Protocol:  "tcp",
								PortRange: "30000-32767",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "Kubernetes NodePort range",
							},
							{
								Protocol:  "tcp",
								PortRange: "80",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "HTTP ingress",
							},
							{
								Protocol:  "tcp",
								PortRange: "443",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "HTTPS ingress",
							},
						},
						Tags: []string{"kubernetes", "cluster"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing required fields", func() {
			ginkgo.It("should return a validation error for missing name", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "", // Missing required field
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing network_id", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name:      "test-fw",
						NetworkId: nil, // Missing required field
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing metadata", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata:   nil, // Missing required metadata
					Spec: &CivoFirewallSpec{
						Name: "test-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing spec", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-firewall",
					},
					Spec: nil, // Missing required spec
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid protocol values", func() {
			ginkgo.It("should return a validation error for invalid inbound protocol", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "invalid-protocol-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "invalid-protocol-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
						InboundRules: []*CivoFirewallInboundRule{
							{
								Protocol:  "invalid", // Invalid protocol
								PortRange: "80",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "Invalid",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for invalid outbound protocol", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "invalid-protocol-outbound",
					},
					Spec: &CivoFirewallSpec{
						Name: "invalid-outbound-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-456",
							},
						},
						OutboundRules: []*CivoFirewallOutboundRule{
							{
								Protocol:  "http", // Invalid protocol
								PortRange: "443",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "Invalid",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for empty protocol", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "empty-protocol-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "empty-protocol-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-789",
							},
						},
						InboundRules: []*CivoFirewallInboundRule{
							{
								Protocol:  "", // Empty protocol
								PortRange: "80",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "Empty protocol",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for uppercase protocol", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "uppercase-protocol-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "uppercase-protocol-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-abc",
							},
						},
						InboundRules: []*CivoFirewallInboundRule{
							{
								Protocol:  "TCP", // Uppercase not allowed
								PortRange: "80",
								Cidrs:     []string{"0.0.0.0/0"},
								Action:    "allow",
								Label:     "Uppercase",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid API version and kind", func() {
			ginkgo.It("should return a validation error for wrong api_version", func() {
				input := &CivoFirewall{
					ApiVersion: "wrong.api.version/v1", // Wrong value
					Kind:       "CivoFirewall",
					Metadata: &shared.CloudResourceMetadata{
						Name: "wrong-api-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "test-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for wrong kind", func() {
				input := &CivoFirewall{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "WrongKind", // Wrong kind
					Metadata: &shared.CloudResourceMetadata{
						Name: "wrong-kind-firewall",
					},
					Spec: &CivoFirewallSpec{
						Name: "test-fw",
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
