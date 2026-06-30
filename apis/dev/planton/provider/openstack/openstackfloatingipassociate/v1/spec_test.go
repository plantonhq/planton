package openstackfloatingipassociatev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOpenStackFloatingIpAssociateSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackFloatingIpAssociateSpec Validation Tests")
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

// minimalValidFloatingIpAssociate returns a minimal valid OpenStackFloatingIpAssociate.
// Both required FKs (floating_ip and port_id) are set with literal values.
func minimalValidFloatingIpAssociate() *OpenStackFloatingIpAssociate {
	return &OpenStackFloatingIpAssociate{
		ApiVersion: "openstack.planton.dev/v1",
		Kind:       "OpenStackFloatingIpAssociate",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-fipa",
		},
		Spec: &OpenStackFloatingIpAssociateSpec{
			FloatingIp: newStringValueOrRef("203.0.113.42"),
			PortId:     newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
		},
	}
}

var _ = ginkgo.Describe("OpenStackFloatingIpAssociateSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_floating_ip_associate", func() {

			ginkgo.It("should not return a validation error for minimal valid association", func() {
				input := minimalValidFloatingIpAssociate()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for floating_ip via value_from ref", func() {
				input := minimalValidFloatingIpAssociate()
				input.Spec.FloatingIp = newValueFromRef("my-floating-ip")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for port_id via value_from ref", func() {
				input := minimalValidFloatingIpAssociate()
				input.Spec.PortId = newValueFromRef("my-port")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for both FKs via value_from refs", func() {
				input := &OpenStackFloatingIpAssociate{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackFloatingIpAssociate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "web-fipa",
					},
					Spec: &OpenStackFloatingIpAssociateSpec{
						FloatingIp: newValueFromRef("web-fip"),
						PortId:     newValueFromRef("web-port"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for association with fixed_ip", func() {
				input := minimalValidFloatingIpAssociate()
				input.Spec.FixedIp = "192.168.1.10"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for association with region override", func() {
				input := minimalValidFloatingIpAssociate()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified association", func() {
				input := &OpenStackFloatingIpAssociate{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackFloatingIpAssociate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-fipa",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackFloatingIpAssociateSpec{
						FloatingIp: newValueFromRef("prod-fip"),
						PortId:     newValueFromRef("prod-web-port"),
						FixedIp:    "10.0.1.100",
						Region:     "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for floating_ip using UUID format", func() {
				input := minimalValidFloatingIpAssociate()
				input.Spec.FloatingIp = newStringValueOrRef("c3d4e5f6-a7b8-9012-cdef-123456789012")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_floating_ip_associate", func() {

			ginkgo.It("should return a validation error when floating_ip is missing", func() {
				input := minimalValidFloatingIpAssociate()
				input.Spec.FloatingIp = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when port_id is missing", func() {
				input := minimalValidFloatingIpAssociate()
				input.Spec.PortId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both FKs are missing", func() {
				input := &OpenStackFloatingIpAssociate{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackFloatingIpAssociate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-fipa",
					},
					Spec: &OpenStackFloatingIpAssociateSpec{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidFloatingIpAssociate()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidFloatingIpAssociate()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidFloatingIpAssociate()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackFloatingIpAssociate{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackFloatingIpAssociate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-fipa",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
