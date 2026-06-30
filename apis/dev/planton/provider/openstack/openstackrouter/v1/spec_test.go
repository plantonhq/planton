package openstackrouterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOpenStackRouterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackRouterSpec Validation Tests")
}

// newStringValueOrRef is a helper to create a literal StringValueOrRef.
func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

// boolPtr is a helper to create a *bool for optional bool fields.
func boolPtr(b bool) *bool {
	return &b
}

// minimalValidRouter returns a minimal valid OpenStackRouter for test scaffolding.
// No spec fields are required -- a router can be created with an empty spec
// (internal-only router with no external gateway).
func minimalValidRouter() *OpenStackRouter {
	return &OpenStackRouter{
		ApiVersion: "openstack.planton.dev/v1",
		Kind:       "OpenStackRouter",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-router",
		},
		Spec: &OpenStackRouterSpec{},
	}
}

var _ = ginkgo.Describe("OpenStackRouterSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_router", func() {

			ginkgo.It("should not return a validation error for minimal valid router (internal-only)", func() {
				input := minimalValidRouter()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router with external_network_id (literal)", func() {
				input := minimalValidRouter()
				input.Spec.ExternalNetworkId = newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router with external_network_id via value_from ref", func() {
				input := minimalValidRouter()
				input.Spec.ExternalNetworkId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-external-network",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router with admin_state_up disabled", func() {
				input := minimalValidRouter()
				input.Spec.AdminStateUp = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router with enable_snat and external_network_id", func() {
				input := minimalValidRouter()
				input.Spec.ExternalNetworkId = newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890")
				input.Spec.EnableSnat = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router with enable_snat=false and external_network_id", func() {
				input := minimalValidRouter()
				input.Spec.ExternalNetworkId = newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890")
				input.Spec.EnableSnat = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router with distributed enabled", func() {
				input := minimalValidRouter()
				input.Spec.Distributed = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router with distributed disabled", func() {
				input := minimalValidRouter()
				input.Spec.Distributed = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router with external_fixed_ips and external_network_id", func() {
				input := minimalValidRouter()
				input.Spec.ExternalNetworkId = newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890")
				input.Spec.ExternalFixedIps = []*ExternalFixedIp{
					{SubnetId: "subnet-uuid-1", IpAddress: "203.0.113.10"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router with multiple external_fixed_ips", func() {
				input := minimalValidRouter()
				input.Spec.ExternalNetworkId = newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890")
				input.Spec.ExternalFixedIps = []*ExternalFixedIp{
					{SubnetId: "subnet-uuid-1", IpAddress: "203.0.113.10"},
					{SubnetId: "subnet-uuid-2"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router with external_fixed_ip with only subnet_id", func() {
				input := minimalValidRouter()
				input.Spec.ExternalNetworkId = newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890")
				input.Spec.ExternalFixedIps = []*ExternalFixedIp{
					{SubnetId: "subnet-uuid-1"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router with external_fixed_ip with only ip_address", func() {
				input := minimalValidRouter()
				input.Spec.ExternalNetworkId = newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890")
				input.Spec.ExternalFixedIps = []*ExternalFixedIp{
					{IpAddress: "203.0.113.10"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router with description", func() {
				input := minimalValidRouter()
				input.Spec.Description = "Edge router for team alpha"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router with tags", func() {
				input := minimalValidRouter()
				input.Spec.Tags = []string{"team:platform", "env:dev"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for router with region override", func() {
				input := minimalValidRouter()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified router", func() {
				input := &OpenStackRouter{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackRouter",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-router",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackRouterSpec{
						ExternalNetworkId: newStringValueOrRef("ext-net-uuid"),
						AdminStateUp:      boolPtr(true),
						EnableSnat:        boolPtr(true),
						Distributed:       boolPtr(false),
						ExternalFixedIps: []*ExternalFixedIp{
							{SubnetId: "ext-subnet-uuid", IpAddress: "203.0.113.50"},
						},
						Description: "Production edge router for ACME Corp",
						Tags:        []string{"production", "managed"},
						Region:      "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_router", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidRouter()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidRouter()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidRouter()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackRouter{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackRouter",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when enable_snat is set without external_network_id", func() {
				input := minimalValidRouter()
				input.Spec.EnableSnat = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when enable_snat=false is set without external_network_id", func() {
				input := minimalValidRouter()
				input.Spec.EnableSnat = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when external_fixed_ips is set without external_network_id", func() {
				input := minimalValidRouter()
				input.Spec.ExternalFixedIps = []*ExternalFixedIp{
					{SubnetId: "subnet-uuid-1", IpAddress: "203.0.113.10"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tags contain duplicates", func() {
				input := minimalValidRouter()
				input.Spec.Tags = []string{"env:dev", "env:dev"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
