package openstacknetworkv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestOpenStackNetworkSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackNetworkSpec Validation Tests")
}

var _ = ginkgo.Describe("OpenStackNetworkSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_network", func() {

			ginkgo.It("should not return a validation error for minimal valid network", func() {
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "my-network",
					},
					Spec: &OpenStackNetworkSpec{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for network with description", func() {
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "described-network",
					},
					Spec: &OpenStackNetworkSpec{
						Description: "A development network for team alpha",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for network with admin_state_up set to false", func() {
				adminStateUp := false
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "admin-down-network",
					},
					Spec: &OpenStackNetworkSpec{
						AdminStateUp: &adminStateUp,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for shared external network", func() {
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "provider-net",
					},
					Spec: &OpenStackNetworkSpec{
						Shared:   true,
						External: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for network with valid MTU", func() {
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "jumbo-network",
					},
					Spec: &OpenStackNetworkSpec{
						Mtu: 9000,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for network with valid dns_domain", func() {
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "dns-network",
					},
					Spec: &OpenStackNetworkSpec{
						DnsDomain: "dev.example.com.",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for network with port_security_enabled", func() {
				portSecurity := true
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "secure-network",
					},
					Spec: &OpenStackNetworkSpec{
						PortSecurityEnabled: &portSecurity,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for network with tags", func() {
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "tagged-network",
					},
					Spec: &OpenStackNetworkSpec{
						Tags: []string{"team:platform", "env:dev"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for network with region override", func() {
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "regional-network",
					},
					Spec: &OpenStackNetworkSpec{
						Region: "RegionTwo",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified network", func() {
				adminStateUp := true
				portSecurity := true
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-network",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackNetworkSpec{
						Description:         "Production network for ACME Corp",
						AdminStateUp:        &adminStateUp,
						Shared:              false,
						External:            false,
						Mtu:                 1500,
						DnsDomain:           "prod.acme.com.",
						PortSecurityEnabled: &portSecurity,
						Tags:                []string{"production", "managed"},
						Region:              "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_network", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := &OpenStackNetwork{
					ApiVersion: "wrong.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-network",
					},
					Spec: &OpenStackNetworkSpec{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-network",
					},
					Spec: &OpenStackNetworkSpec{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Spec:       &OpenStackNetworkSpec{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-network",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when mtu is negative", func() {
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-network",
					},
					Spec: &OpenStackNetworkSpec{
						Mtu: -1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when dns_domain does not end with a dot", func() {
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-network",
					},
					Spec: &OpenStackNetworkSpec{
						DnsDomain: "example.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tags contain duplicates", func() {
				input := &OpenStackNetwork{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackNetwork",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-network",
					},
					Spec: &OpenStackNetworkSpec{
						Tags: []string{"env:dev", "env:dev"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
