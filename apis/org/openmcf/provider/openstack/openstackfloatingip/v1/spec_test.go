package openstackfloatingipv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOpenStackFloatingIpSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackFloatingIpSpec Validation Tests")
}

// newStringValueOrRef is a helper to create a literal StringValueOrRef.
func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

// newStringValueFromRef is a helper to create a value_from StringValueOrRef.
func newStringValueFromRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Name: name,
			},
		},
	}
}

// minimalValidFloatingIp returns a minimal valid OpenStackFloatingIp for test scaffolding.
// Only the required field (floating_network_id) is set.
func minimalValidFloatingIp() *OpenStackFloatingIp {
	return &OpenStackFloatingIp{
		ApiVersion: "openstack.openmcf.org/v1",
		Kind:       "OpenStackFloatingIp",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-fip",
		},
		Spec: &OpenStackFloatingIpSpec{
			FloatingNetworkId: newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
		},
	}
}

var _ = ginkgo.Describe("OpenStackFloatingIpSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_floating_ip", func() {

			ginkgo.It("should not return a validation error for minimal valid floating IP (allocation-only)", func() {
				input := minimalValidFloatingIp()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for floating_network_id via value_from ref", func() {
				input := minimalValidFloatingIp()
				input.Spec.FloatingNetworkId = newStringValueFromRef("my-external-network")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for floating IP with port_id (built-in association)", func() {
				input := minimalValidFloatingIp()
				input.Spec.PortId = newStringValueOrRef("port-uuid-1234")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for floating IP with port_id via value_from ref", func() {
				input := minimalValidFloatingIp()
				input.Spec.PortId = newStringValueFromRef("my-instance-port")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for floating IP with port_id and fixed_ip", func() {
				input := minimalValidFloatingIp()
				input.Spec.PortId = newStringValueOrRef("port-uuid-1234")
				input.Spec.FixedIp = "10.0.1.5"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for floating IP with subnet_id", func() {
				input := minimalValidFloatingIp()
				input.Spec.SubnetId = "ext-subnet-uuid-5678"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for floating IP with specific address", func() {
				input := minimalValidFloatingIp()
				input.Spec.Address = "203.0.113.42"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for floating IP with description", func() {
				input := minimalValidFloatingIp()
				input.Spec.Description = "Public IP for web server"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for floating IP with tags", func() {
				input := minimalValidFloatingIp()
				input.Spec.Tags = []string{"team:platform", "env:production"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for floating IP with region override", func() {
				input := minimalValidFloatingIp()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified floating IP", func() {
				input := &OpenStackFloatingIp{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackFloatingIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-fip",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackFloatingIpSpec{
						FloatingNetworkId: newStringValueOrRef("ext-net-uuid"),
						PortId:            newStringValueOrRef("port-uuid"),
						FixedIp:           "10.0.1.5",
						SubnetId:          "ext-subnet-uuid",
						Address:           "203.0.113.50",
						Description:       "Production web server public IP",
						Tags:              []string{"production", "managed"},
						Region:            "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_floating_ip", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidFloatingIp()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidFloatingIp()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidFloatingIp()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackFloatingIp{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackFloatingIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-fip",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when floating_network_id is missing", func() {
				input := &OpenStackFloatingIp{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackFloatingIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-fip",
					},
					Spec: &OpenStackFloatingIpSpec{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when fixed_ip is set without port_id", func() {
				input := minimalValidFloatingIp()
				input.Spec.FixedIp = "10.0.1.5"
				// port_id is NOT set -- CEL should reject this
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tags contain duplicates", func() {
				input := minimalValidFloatingIp()
				input.Spec.Tags = []string{"env:dev", "env:dev"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
