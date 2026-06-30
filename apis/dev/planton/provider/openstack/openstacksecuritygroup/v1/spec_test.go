package openstacksecuritygroupv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestOpenStackSecurityGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackSecurityGroupSpec Validation Tests")
}

// boolPtr is a helper to create a *bool for optional bool fields.
func boolPtr(b bool) *bool {
	return &b
}

// int32Ptr is a helper to create a *int32 for optional int32 fields.
func int32Ptr(v int32) *int32 {
	return &v
}

// minimalValidSecurityGroup returns a minimal valid OpenStackSecurityGroup for test scaffolding.
// A security group can be created with an empty spec (no rules, no description).
func minimalValidSecurityGroup() *OpenStackSecurityGroup {
	return &OpenStackSecurityGroup{
		ApiVersion: "openstack.planton.dev/v1",
		Kind:       "OpenStackSecurityGroup",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-sg",
		},
		Spec: &OpenStackSecurityGroupSpec{},
	}
}

// validIngressRule returns a minimal valid ingress rule for test scaffolding.
func validIngressRule(key string) *SecurityGroupRule {
	return &SecurityGroupRule{
		Key:            key,
		Direction:      "ingress",
		Ethertype:      "IPv4",
		Protocol:       "tcp",
		PortRangeMin:   int32Ptr(22),
		PortRangeMax:   int32Ptr(22),
		RemoteIpPrefix: "0.0.0.0/0",
	}
}

var _ = ginkgo.Describe("OpenStackSecurityGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_security_group", func() {

			ginkgo.It("should not return a validation error for minimal valid security group (no rules)", func() {
				input := minimalValidSecurityGroup()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for security group with description", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Description = "Web tier security group for developer environments"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for security group with delete_default_rules=true", func() {
				input := minimalValidSecurityGroup()
				input.Spec.DeleteDefaultRules = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for security group with delete_default_rules=false", func() {
				input := minimalValidSecurityGroup()
				input.Spec.DeleteDefaultRules = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for security group with stateful=true", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Stateful = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for security group with stateful=false (stateless)", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Stateful = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for security group with tags", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Tags = []string{"team:platform", "env:dev"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for security group with region override", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for security group with a single ingress TCP rule", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					validIngressRule("allow-ssh"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for security group with an egress rule", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					{
						Key:            "egress-all-ipv4",
						Direction:      "egress",
						Ethertype:      "IPv4",
						RemoteIpPrefix: "0.0.0.0/0",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a rule with protocol but no ports (e.g., all ICMP)", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					{
						Key:            "allow-all-icmp",
						Direction:      "ingress",
						Ethertype:      "IPv4",
						Protocol:       "icmp",
						RemoteIpPrefix: "0.0.0.0/0",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an ICMP rule where port_range_min > port_range_max", func() {
				// ICMP type 8 (Echo Request), code 0 -- min > max is valid for ICMP
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					{
						Key:            "allow-icmp-echo-request",
						Direction:      "ingress",
						Ethertype:      "IPv4",
						Protocol:       "icmp",
						PortRangeMin:   int32Ptr(8),
						PortRangeMax:   int32Ptr(0),
						RemoteIpPrefix: "0.0.0.0/0",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a rule with remote_group_id instead of remote_ip_prefix", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					{
						Key:           "allow-from-db-sg",
						Direction:     "ingress",
						Ethertype:     "IPv4",
						Protocol:      "tcp",
						PortRangeMin:  int32Ptr(5432),
						PortRangeMax:  int32Ptr(5432),
						RemoteGroupId: "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a rule with only direction and ethertype (allow all)", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					{
						Key:       "allow-all-ingress-ipv4",
						Direction: "ingress",
						Ethertype: "IPv4",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a rule with IPv6 ethertype", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					{
						Key:            "allow-ssh-ipv6",
						Direction:      "ingress",
						Ethertype:      "IPv6",
						Protocol:       "tcp",
						PortRangeMin:   int32Ptr(22),
						PortRangeMax:   int32Ptr(22),
						RemoteIpPrefix: "::/0",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for security group with multiple rules (different keys)", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					validIngressRule("allow-ssh"),
					{
						Key:            "allow-https",
						Direction:      "ingress",
						Ethertype:      "IPv4",
						Protocol:       "tcp",
						PortRangeMin:   int32Ptr(443),
						PortRangeMax:   int32Ptr(443),
						RemoteIpPrefix: "0.0.0.0/0",
					},
					{
						Key:       "egress-all",
						Direction: "egress",
						Ethertype: "IPv4",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a rule with description", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					{
						Key:            "allow-ssh",
						Direction:      "ingress",
						Ethertype:      "IPv4",
						Protocol:       "tcp",
						PortRangeMin:   int32Ptr(22),
						PortRangeMax:   int32Ptr(22),
						RemoteIpPrefix: "0.0.0.0/0",
						Description:    "Allow SSH access from anywhere",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified security group", func() {
				input := &OpenStackSecurityGroup{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "web-tier-sg",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackSecurityGroupSpec{
						Description:        "Web tier security group for production",
						DeleteDefaultRules: boolPtr(true),
						Stateful:           boolPtr(true),
						Rules: []*SecurityGroupRule{
							{
								Key:            "allow-http",
								Direction:      "ingress",
								Ethertype:      "IPv4",
								Protocol:       "tcp",
								PortRangeMin:   int32Ptr(80),
								PortRangeMax:   int32Ptr(80),
								RemoteIpPrefix: "0.0.0.0/0",
								Description:    "Allow HTTP from anywhere",
							},
							{
								Key:            "allow-https",
								Direction:      "ingress",
								Ethertype:      "IPv4",
								Protocol:       "tcp",
								PortRangeMin:   int32Ptr(443),
								PortRangeMax:   int32Ptr(443),
								RemoteIpPrefix: "0.0.0.0/0",
								Description:    "Allow HTTPS from anywhere",
							},
							{
								Key:       "egress-all-ipv4",
								Direction: "egress",
								Ethertype: "IPv4",
							},
						},
						Tags:   []string{"production", "web-tier"},
						Region: "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_security_group", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidSecurityGroup()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidSecurityGroup()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidSecurityGroup()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackSecurityGroup{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-sg",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tags contain duplicates", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Tags = []string{"env:dev", "env:dev"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule key is empty", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					{
						Key:       "",
						Direction: "ingress",
						Ethertype: "IPv4",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule direction is invalid", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					{
						Key:       "bad-direction",
						Direction: "inbound",
						Ethertype: "IPv4",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule ethertype is invalid", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					{
						Key:       "bad-ethertype",
						Direction: "ingress",
						Ethertype: "ipv4",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when port_range_min is set without port_range_max", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					{
						Key:          "partial-port-range",
						Direction:    "ingress",
						Ethertype:    "IPv4",
						Protocol:     "tcp",
						PortRangeMin: int32Ptr(22),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when port_range_max is set without port_range_min", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					{
						Key:          "partial-port-range",
						Direction:    "ingress",
						Ethertype:    "IPv4",
						Protocol:     "tcp",
						PortRangeMax: int32Ptr(22),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ports are set without protocol", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					{
						Key:          "ports-without-protocol",
						Direction:    "ingress",
						Ethertype:    "IPv4",
						PortRangeMin: int32Ptr(80),
						PortRangeMax: int32Ptr(80),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both remote_group_id and remote_ip_prefix are set", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					{
						Key:            "conflicting-remotes",
						Direction:      "ingress",
						Ethertype:      "IPv4",
						RemoteGroupId:  "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
						RemoteIpPrefix: "10.0.0.0/8",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when inline rule keys are duplicated", func() {
				input := minimalValidSecurityGroup()
				input.Spec.Rules = []*SecurityGroupRule{
					validIngressRule("allow-ssh"),
					{
						Key:            "allow-ssh",
						Direction:      "ingress",
						Ethertype:      "IPv4",
						Protocol:       "tcp",
						PortRangeMin:   int32Ptr(443),
						PortRangeMax:   int32Ptr(443),
						RemoteIpPrefix: "0.0.0.0/0",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
