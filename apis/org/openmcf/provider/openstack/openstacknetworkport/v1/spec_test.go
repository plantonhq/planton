package openstacknetworkportv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOpenStackNetworkPortSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackNetworkPortSpec Validation Tests")
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

// boolPtr is a helper to create a *bool for optional bool fields.
func boolPtr(v bool) *bool {
	return &v
}

// minimalValidNetworkPort returns a minimal valid OpenStackNetworkPort for test scaffolding.
// Only the required field (network_id) is set.
func minimalValidNetworkPort() *OpenStackNetworkPort {
	return &OpenStackNetworkPort{
		ApiVersion: "openstack.openmcf.org/v1",
		Kind:       "OpenStackNetworkPort",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-port",
		},
		Spec: &OpenStackNetworkPortSpec{
			NetworkId: newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
		},
	}
}

var _ = ginkgo.Describe("OpenStackNetworkPortSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_network_port", func() {

			ginkgo.It("should not return a validation error for minimal valid port (network_id only)", func() {
				input := minimalValidNetworkPort()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for network_id via value_from ref", func() {
				input := minimalValidNetworkPort()
				input.Spec.NetworkId = newValueFromRef("my-network")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for port with a single fixed IP", func() {
				input := minimalValidNetworkPort()
				input.Spec.FixedIps = []*FixedIp{
					{
						SubnetId:  newStringValueOrRef("subnet-uuid-1234"),
						IpAddress: "192.168.1.10",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fixed IP with subnet_id via value_from ref", func() {
				input := minimalValidNetworkPort()
				input.Spec.FixedIps = []*FixedIp{
					{
						SubnetId: newValueFromRef("my-subnet"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for multi-homed port with two fixed IPs", func() {
				input := minimalValidNetworkPort()
				input.Spec.FixedIps = []*FixedIp{
					{SubnetId: newStringValueOrRef("subnet-a")},
					{SubnetId: newStringValueOrRef("subnet-b"), IpAddress: "10.0.0.5"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for port with security_group_ids as literals", func() {
				input := minimalValidNetworkPort()
				input.Spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("sg-uuid-1"),
					newStringValueOrRef("sg-uuid-2"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for port with security_group_ids via value_from refs", func() {
				input := minimalValidNetworkPort()
				input.Spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					newValueFromRef("app-sg"),
					newValueFromRef("bastion-sg"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for port with no_security_groups = true", func() {
				input := minimalValidNetworkPort()
				input.Spec.NoSecurityGroups = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for port with admin_state_up = false", func() {
				input := minimalValidNetworkPort()
				input.Spec.AdminStateUp = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for port with mac_address", func() {
				input := minimalValidNetworkPort()
				input.Spec.MacAddress = "fa:16:3e:aa:bb:cc"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for port with port_security_enabled = false", func() {
				input := minimalValidNetworkPort()
				input.Spec.PortSecurityEnabled = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for port with description", func() {
				input := minimalValidNetworkPort()
				input.Spec.Description = "Web server primary network interface"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for port with tags", func() {
				input := minimalValidNetworkPort()
				input.Spec.Tags = []string{"team:platform", "env:production"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for port with region override", func() {
				input := minimalValidNetworkPort()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified port", func() {
				input := &OpenStackNetworkPort{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackNetworkPort",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-port",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackNetworkPortSpec{
						NetworkId: newValueFromRef("my-network"),
						FixedIps: []*FixedIp{
							{SubnetId: newValueFromRef("my-subnet"), IpAddress: "192.168.1.100"},
						},
						SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{
							newValueFromRef("app-sg"),
						},
						AdminStateUp:        boolPtr(true),
						MacAddress:          "fa:16:3e:11:22:33",
						PortSecurityEnabled: boolPtr(true),
						Description:         "Production web server port",
						Tags:                []string{"managed", "production"},
						Region:              "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for port with fixed IP having only ip_address (no subnet_id)", func() {
				input := minimalValidNetworkPort()
				input.Spec.FixedIps = []*FixedIp{
					{IpAddress: "192.168.1.50"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for empty security_group_ids (uses default SG)", func() {
				input := minimalValidNetworkPort()
				input.Spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_network_port", func() {

			ginkgo.It("should return a validation error when network_id is missing", func() {
				input := &OpenStackNetworkPort{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackNetworkPort",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-port",
					},
					Spec: &OpenStackNetworkPortSpec{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when no_security_groups and security_group_ids are both set", func() {
				input := minimalValidNetworkPort()
				input.Spec.NoSecurityGroups = true
				input.Spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("sg-uuid-1"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tags contain duplicates", func() {
				input := minimalValidNetworkPort()
				input.Spec.Tags = []string{"env:dev", "env:dev"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidNetworkPort()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidNetworkPort()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidNetworkPort()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackNetworkPort{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackNetworkPort",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-port",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
