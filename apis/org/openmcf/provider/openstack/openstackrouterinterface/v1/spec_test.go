package openstackrouterinterfacev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOpenStackRouterInterfaceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackRouterInterfaceSpec Validation Tests")
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

// minimalValidRouterInterface returns a minimal valid OpenStackRouterInterface
// with both required FKs as literal values.
func minimalValidRouterInterface() *OpenStackRouterInterface {
	return &OpenStackRouterInterface{
		ApiVersion: "openstack.openmcf.org/v1",
		Kind:       "OpenStackRouterInterface",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-router-interface",
		},
		Spec: &OpenStackRouterInterfaceSpec{
			RouterId: newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
			SubnetId: newStringValueOrRef("b2c3d4e5-f6a7-8901-bcde-f12345678901"),
		},
	}
}

var _ = ginkgo.Describe("OpenStackRouterInterfaceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_router_interface", func() {

			ginkgo.It("should not return a validation error for minimal valid router interface (both FKs as literals)", func() {
				input := minimalValidRouterInterface()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error when both FKs use value_from refs", func() {
				input := &OpenStackRouterInterface{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackRouterInterface",
					Metadata: &shared.CloudResourceMetadata{
						Name: "ref-router-interface",
					},
					Spec: &OpenStackRouterInterfaceSpec{
						RouterId: newValueFromRef("my-router"),
						SubnetId: newValueFromRef("my-subnet"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for mixed mode: router_id literal, subnet_id value_from", func() {
				input := minimalValidRouterInterface()
				input.Spec.SubnetId = newValueFromRef("my-subnet")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for mixed mode: router_id value_from, subnet_id literal", func() {
				input := minimalValidRouterInterface()
				input.Spec.RouterId = newValueFromRef("my-router")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router interface with region override", func() {
				input := minimalValidRouterInterface()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified router interface", func() {
				input := &OpenStackRouterInterface{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackRouterInterface",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-router-interface",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackRouterInterfaceSpec{
						RouterId: newStringValueOrRef("router-uuid-here"),
						SubnetId: newStringValueOrRef("subnet-uuid-here"),
						Region:   "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_router_interface", func() {

			ginkgo.It("should return a validation error when router_id is missing", func() {
				input := minimalValidRouterInterface()
				input.Spec.RouterId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet_id is missing", func() {
				input := minimalValidRouterInterface()
				input.Spec.SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidRouterInterface()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidRouterInterface()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidRouterInterface()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackRouterInterface{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackRouterInterface",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-interface",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
