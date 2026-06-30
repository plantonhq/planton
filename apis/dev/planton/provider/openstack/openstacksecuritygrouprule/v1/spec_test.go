package openstacksecuritygrouprulev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOpenStackSecurityGroupRuleSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackSecurityGroupRuleSpec Validation Tests")
}

// int32Ptr is a helper to create a *int32 for optional int32 fields.
func int32Ptr(v int32) *int32 {
	return &v
}

// newStringValueOrRef is a helper to create a literal StringValueOrRef.
func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

// newValueFromRef is a helper to create a value_from StringValueOrRef.
func newValueFromRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Name: name,
			},
		},
	}
}

// minimalValidSecurityGroupRule returns a minimal valid OpenStackSecurityGroupRule.
// A minimal rule needs: security_group_id (required FK), direction, and ethertype.
func minimalValidSecurityGroupRule() *OpenStackSecurityGroupRule {
	return &OpenStackSecurityGroupRule{
		ApiVersion: "openstack.planton.dev/v1",
		Kind:       "OpenStackSecurityGroupRule",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-rule",
		},
		Spec: &OpenStackSecurityGroupRuleSpec{
			SecurityGroupId: newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
			Direction:       "ingress",
			Ethertype:       "IPv4",
		},
	}
}

var _ = ginkgo.Describe("OpenStackSecurityGroupRuleSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_security_group_rule", func() {

			ginkgo.It("should not return a validation error for minimal valid rule (ingress IPv4)", func() {
				input := minimalValidSecurityGroupRule()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for minimal egress IPv6 rule", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.Direction = "egress"
				input.Spec.Ethertype = "IPv6"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error when security_group_id uses value_from ref", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.SecurityGroupId = newValueFromRef("my-security-group")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for rule with protocol and port range", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.Protocol = "tcp"
				input.Spec.PortRangeMin = int32Ptr(22)
				input.Spec.PortRangeMax = int32Ptr(22)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for rule with protocol and wide port range", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.Protocol = "tcp"
				input.Spec.PortRangeMin = int32Ptr(80)
				input.Spec.PortRangeMax = int32Ptr(443)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for ICMP rule with type and code", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.Protocol = "icmp"
				input.Spec.PortRangeMin = int32Ptr(8) // Echo Request
				input.Spec.PortRangeMax = int32Ptr(0) // Code 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for rule with protocol only (all ports)", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.Protocol = "tcp"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for rule with remote_ip_prefix", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.RemoteIpPrefix = "0.0.0.0/0"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for rule with remote_group_id as literal", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.RemoteGroupId = newStringValueOrRef("b2c3d4e5-f6a7-8901-bcde-f12345678901")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for rule with remote_group_id as value_from ref", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.RemoteGroupId = newValueFromRef("bastion-sg")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for cross-SG rule with both FKs as value_from", func() {
				input := &OpenStackSecurityGroupRule{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackSecurityGroupRule",
					Metadata: &shared.CloudResourceMetadata{
						Name: "allow-ssh-from-bastion",
					},
					Spec: &OpenStackSecurityGroupRuleSpec{
						SecurityGroupId: newValueFromRef("app-sg"),
						Direction:       "ingress",
						Ethertype:       "IPv4",
						Protocol:        "tcp",
						PortRangeMin:    int32Ptr(22),
						PortRangeMax:    int32Ptr(22),
						RemoteGroupId:   newValueFromRef("bastion-sg"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for rule with description", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.Description = "Allow SSH from anywhere"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for rule with region override", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified rule", func() {
				input := &OpenStackSecurityGroupRule{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackSecurityGroupRule",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-rule",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackSecurityGroupRuleSpec{
						SecurityGroupId: newStringValueOrRef("sg-uuid-here"),
						Direction:       "ingress",
						Ethertype:       "IPv4",
						Protocol:        "tcp",
						PortRangeMin:    int32Ptr(443),
						PortRangeMax:    int32Ptr(443),
						RemoteIpPrefix:  "10.0.0.0/8",
						Description:     "Allow HTTPS from private network",
						Region:          "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for port range 0-65535 (full range)", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.Protocol = "tcp"
				input.Spec.PortRangeMin = int32Ptr(0)
				input.Spec.PortRangeMax = int32Ptr(65535)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_security_group_rule", func() {

			ginkgo.It("should return a validation error when security_group_id is missing", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.SecurityGroupId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when direction is invalid", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.Direction = "bidirectional"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when direction is empty", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.Direction = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ethertype is invalid", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.Ethertype = "IPv5"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ethertype is empty", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.Ethertype = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when only port_range_min is set", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.Protocol = "tcp"
				input.Spec.PortRangeMin = int32Ptr(22)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when only port_range_max is set", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.Protocol = "tcp"
				input.Spec.PortRangeMax = int32Ptr(22)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when port range is set without protocol", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.PortRangeMin = int32Ptr(22)
				input.Spec.PortRangeMax = int32Ptr(22)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when remote_group_id and remote_ip_prefix are both set", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.RemoteGroupId = newStringValueOrRef("other-sg-uuid")
				input.Spec.RemoteIpPrefix = "0.0.0.0/0"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when remote_group_id (value_from) and remote_ip_prefix are both set", func() {
				input := minimalValidSecurityGroupRule()
				input.Spec.RemoteGroupId = newValueFromRef("bastion-sg")
				input.Spec.RemoteIpPrefix = "10.0.0.0/8"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidSecurityGroupRule()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidSecurityGroupRule()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidSecurityGroupRule()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackSecurityGroupRule{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackSecurityGroupRule",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-rule",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
