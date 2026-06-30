package openstacksubnetv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOpenStackSubnetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackSubnetSpec Validation Tests")
}

// newStringValueOrRef is a helper to create a literal StringValueOrRef.
func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

// minimalValidSubnet returns a minimal valid OpenStackSubnet for test scaffolding.
func minimalValidSubnet() *OpenStackSubnet {
	return &OpenStackSubnet{
		ApiVersion: "openstack.planton.dev/v1",
		Kind:       "OpenStackSubnet",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-subnet",
		},
		Spec: &OpenStackSubnetSpec{
			NetworkId: newStringValueOrRef("e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d"),
			Cidr:      "192.168.1.0/24",
		},
	}
}

var _ = ginkgo.Describe("OpenStackSubnetSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_subnet", func() {

			ginkgo.It("should not return a validation error for minimal valid subnet", func() {
				input := minimalValidSubnet()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for subnet with description", func() {
				input := minimalValidSubnet()
				input.Spec.Description = "Development subnet for team alpha"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for subnet with ip_version 4", func() {
				ipVersion := int32(4)
				input := minimalValidSubnet()
				input.Spec.IpVersion = &ipVersion
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for subnet with ip_version 6", func() {
				ipVersion := int32(6)
				input := minimalValidSubnet()
				input.Spec.IpVersion = &ipVersion
				input.Spec.Cidr = "2001:db8::/64"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for subnet with gateway_ip", func() {
				input := minimalValidSubnet()
				input.Spec.GatewayIp = "192.168.1.1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for subnet with no_gateway", func() {
				input := minimalValidSubnet()
				input.Spec.NoGateway = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for subnet with enable_dhcp disabled", func() {
				enableDhcp := false
				input := minimalValidSubnet()
				input.Spec.EnableDhcp = &enableDhcp
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for subnet with dns_nameservers", func() {
				input := minimalValidSubnet()
				input.Spec.DnsNameservers = []string{"8.8.8.8", "8.8.4.4"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for subnet with single allocation pool", func() {
				input := minimalValidSubnet()
				input.Spec.AllocationPools = []*AllocationPool{
					{Start: "192.168.1.100", End: "192.168.1.200"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for subnet with multiple allocation pools", func() {
				input := minimalValidSubnet()
				input.Spec.AllocationPools = []*AllocationPool{
					{Start: "192.168.1.10", End: "192.168.1.50"},
					{Start: "192.168.1.100", End: "192.168.1.200"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for subnet with tags", func() {
				input := minimalValidSubnet()
				input.Spec.Tags = []string{"team:platform", "env:dev"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for subnet with region override", func() {
				input := minimalValidSubnet()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for subnet with network_id via value_from ref", func() {
				input := minimalValidSubnet()
				input.Spec.NetworkId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-network",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified subnet", func() {
				ipVersion := int32(4)
				enableDhcp := true
				input := &OpenStackSubnet{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-subnet",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackSubnetSpec{
						NetworkId:      newStringValueOrRef("e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d"),
						Cidr:           "10.0.0.0/16",
						IpVersion:      &ipVersion,
						GatewayIp:      "10.0.0.1",
						EnableDhcp:     &enableDhcp,
						DnsNameservers: []string{"8.8.8.8", "8.8.4.4"},
						AllocationPools: []*AllocationPool{
							{Start: "10.0.1.0", End: "10.0.1.255"},
							{Start: "10.0.2.0", End: "10.0.2.255"},
						},
						Description: "Production subnet for ACME Corp",
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
		ginkgo.Context("openstack_subnet", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidSubnet()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidSubnet()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidSubnet()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackSubnet{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network_id is missing", func() {
				input := minimalValidSubnet()
				input.Spec.NetworkId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cidr is empty", func() {
				input := minimalValidSubnet()
				input.Spec.Cidr = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cidr is not valid CIDR format", func() {
				input := minimalValidSubnet()
				input.Spec.Cidr = "not-a-cidr"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cidr has no prefix length", func() {
				input := minimalValidSubnet()
				input.Spec.Cidr = "192.168.1.0"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ip_version is invalid", func() {
				ipVersion := int32(3)
				input := minimalValidSubnet()
				input.Spec.IpVersion = &ipVersion
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both gateway_ip and no_gateway are set", func() {
				input := minimalValidSubnet()
				input.Spec.GatewayIp = "192.168.1.1"
				input.Spec.NoGateway = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tags contain duplicates", func() {
				input := minimalValidSubnet()
				input.Spec.Tags = []string{"env:dev", "env:dev"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when allocation_pool start is empty", func() {
				input := minimalValidSubnet()
				input.Spec.AllocationPools = []*AllocationPool{
					{Start: "", End: "192.168.1.200"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when allocation_pool end is empty", func() {
				input := minimalValidSubnet()
				input.Spec.AllocationPools = []*AllocationPool{
					{Start: "192.168.1.100", End: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
