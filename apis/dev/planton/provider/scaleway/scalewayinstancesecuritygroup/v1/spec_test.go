package scalewayinstancesecuritygroupv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestScalewayInstanceSecurityGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "ScalewayInstanceSecurityGroupSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("ScalewayInstanceSecurityGroupSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("minimal valid security group", func() {
			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-sg",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone: "fr-par-1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("security group with description and all optional fields", func() {
			ginkgo.It("should accept a fully specified security group", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-sg",
						Org:  "acme-corp",
						Env:  "production",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone:                  "fr-par-1",
						Description:           "Production web tier firewall",
						Stateful:              true,
						InboundDefaultPolicy:  "drop",
						OutboundDefaultPolicy: "accept",
						EnableDefaultSecurity: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("security group with inbound rules", func() {
			ginkgo.It("should accept a single TCP inbound rule", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "web-sg",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone:                 "fr-par-1",
						InboundDefaultPolicy: "drop",
						InboundRules: []*ScalewaySecurityGroupInboundRule{
							{
								Action:    "accept",
								Protocol:  "TCP",
								PortRange: "80",
								IpRange:   "0.0.0.0/0",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept multiple inbound rules with different protocols", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "multi-rule-sg",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone:                 "fr-par-1",
						InboundDefaultPolicy: "drop",
						InboundRules: []*ScalewaySecurityGroupInboundRule{
							{
								Action:    "accept",
								Protocol:  "TCP",
								PortRange: "22",
								IpRange:   "203.0.113.10/32",
							},
							{
								Action:    "accept",
								Protocol:  "TCP",
								PortRange: "443",
								IpRange:   "0.0.0.0/0",
							},
							{
								Action:   "accept",
								Protocol: "ICMP",
								IpRange:  "0.0.0.0/0",
							},
							{
								Action:    "accept",
								Protocol:  "UDP",
								PortRange: "53",
								IpRange:   "0.0.0.0/0",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept port ranges", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "range-sg",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone: "nl-ams-1",
						InboundRules: []*ScalewaySecurityGroupInboundRule{
							{
								Action:    "accept",
								Protocol:  "TCP",
								PortRange: "8000-9000",
								IpRange:   "10.0.0.0/8",
							},
							{
								Action:    "accept",
								Protocol:  "TCP",
								PortRange: "30000-32767",
								IpRange:   "0.0.0.0/0",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept ANY protocol", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "any-protocol-sg",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone: "fr-par-1",
						InboundRules: []*ScalewaySecurityGroupInboundRule{
							{
								Action:   "accept",
								Protocol: "ANY",
								IpRange:  "10.0.0.0/8",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept drop action", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "drop-rule-sg",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone: "fr-par-1",
						InboundRules: []*ScalewaySecurityGroupInboundRule{
							{
								Action:    "drop",
								Protocol:  "TCP",
								PortRange: "22",
								IpRange:   "0.0.0.0/0",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("security group with outbound rules", func() {
			ginkgo.It("should accept outbound rules", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "outbound-sg",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone:                  "fr-par-1",
						OutboundDefaultPolicy: "drop",
						OutboundRules: []*ScalewaySecurityGroupOutboundRule{
							{
								Action:    "accept",
								Protocol:  "TCP",
								PortRange: "443",
								IpRange:   "0.0.0.0/0",
							},
							{
								Action:    "accept",
								Protocol:  "UDP",
								PortRange: "53",
								IpRange:   "0.0.0.0/0",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept both inbound and outbound rules", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "bidirectional-sg",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone:                  "fr-par-1",
						InboundDefaultPolicy:  "drop",
						OutboundDefaultPolicy: "drop",
						InboundRules: []*ScalewaySecurityGroupInboundRule{
							{
								Action:    "accept",
								Protocol:  "TCP",
								PortRange: "443",
								IpRange:   "0.0.0.0/0",
							},
						},
						OutboundRules: []*ScalewaySecurityGroupOutboundRule{
							{
								Action:    "accept",
								Protocol:  "TCP",
								PortRange: "443",
								IpRange:   "0.0.0.0/0",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("real-world patterns", func() {
			ginkgo.It("should accept a Kubernetes worker node pattern", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "k8s-workers-sg",
						Org:  "acme-corp",
						Env:  "staging",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone:                 "nl-ams-1",
						Description:          "Kubernetes worker nodes",
						InboundDefaultPolicy: "drop",
						InboundRules: []*ScalewaySecurityGroupInboundRule{
							{Action: "accept", Protocol: "TCP", PortRange: "6443", IpRange: "0.0.0.0/0"},
							{Action: "accept", Protocol: "TCP", PortRange: "30000-32767", IpRange: "0.0.0.0/0"},
							{Action: "accept", Protocol: "TCP", PortRange: "80", IpRange: "0.0.0.0/0"},
							{Action: "accept", Protocol: "TCP", PortRange: "443", IpRange: "0.0.0.0/0"},
							{Action: "accept", Protocol: "ANY", IpRange: "10.0.0.0/8"},
							{Action: "accept", Protocol: "ICMP", IpRange: "0.0.0.0/0"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing required fields", func() {
			ginkgo.It("should return a validation error for missing zone", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-zone-sg",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone: "", // Missing required field
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing metadata", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata:   nil, // Missing required metadata
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone: "fr-par-1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing spec", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-spec-sg",
					},
					Spec: nil, // Missing required spec
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid rule actions", func() {
			ginkgo.It("should return a validation error for invalid inbound action", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "bad-action-sg",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone: "fr-par-1",
						InboundRules: []*ScalewaySecurityGroupInboundRule{
							{
								Action:    "allow", // Invalid -- must be "accept" or "drop"
								Protocol:  "TCP",
								PortRange: "80",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for invalid outbound action", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "bad-outbound-action-sg",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone: "fr-par-1",
						OutboundRules: []*ScalewaySecurityGroupOutboundRule{
							{
								Action:    "deny", // Invalid -- must be "accept" or "drop"
								Protocol:  "TCP",
								PortRange: "443",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for empty action", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "empty-action-sg",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone: "fr-par-1",
						InboundRules: []*ScalewaySecurityGroupInboundRule{
							{
								Action:    "", // Empty action -- required field
								Protocol:  "TCP",
								PortRange: "80",
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
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "wrong.api.version/v1",
					Kind:       "ScalewayInstanceSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "wrong-api-sg",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone: "fr-par-1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for wrong kind", func() {
				input := &ScalewayInstanceSecurityGroup{
					ApiVersion: "scaleway.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "wrong-kind-sg",
					},
					Spec: &ScalewayInstanceSecurityGroupSpec{
						Zone: "fr-par-1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
