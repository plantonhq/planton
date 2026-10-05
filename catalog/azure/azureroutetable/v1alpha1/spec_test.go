package azureroutetablev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureRouteTableSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureRouteTableSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// ref builds a StringValueOrRef carrying a value_from reference.
func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

// validResource returns a minimal valid AzureRouteTable that individual
// cases then mutate into the shape under test.
func validResource() *AzureRouteTable {
	return &AzureRouteTable{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureRouteTable",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-route-table",
		},
		Spec: &AzureRouteTableSpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "test-rt",
		},
	}
}

var _ = ginkgo.Describe("AzureRouteTableSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_route_table", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the resource group as a reference", func() {
				input := validResource()
				input.Spec.ResourceGroup = ref("platform-rg")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an internet route", func() {
				input := validResource()
				input.Spec.Routes = []*AzureRouteTableRoute{
					{
						Name:          "dmz-direct",
						AddressPrefix: "0.0.0.0/0",
						NextHopType:   AzureRouteTableNextHopType_INTERNET,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a virtual appliance route with a literal next hop IP", func() {
				input := validResource()
				input.Spec.Routes = []*AzureRouteTableRoute{
					{
						Name:               "default-via-firewall",
						AddressPrefix:      "0.0.0.0/0",
						NextHopType:        AzureRouteTableNextHopType_VIRTUAL_APPLIANCE,
						NextHopInIpAddress: literal("10.0.1.4"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a virtual appliance route referencing a firewall's private IP output", func() {
				input := validResource()
				input.Spec.Routes = []*AzureRouteTableRoute{
					{
						Name:               "default-via-firewall",
						AddressPrefix:      "0.0.0.0/0",
						NextHopType:        AzureRouteTableNextHopType_VIRTUAL_APPLIANCE,
						NextHopInIpAddress: ref("hub-firewall"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a service tag address prefix", func() {
				input := validResource()
				input.Spec.Routes = []*AzureRouteTableRoute{
					{
						Name:          "backup-direct",
						AddressPrefix: "AzureBackup",
						NextHopType:   AzureRouteTableNextHopType_INTERNET,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a black-hole route", func() {
				input := validResource()
				input.Spec.Routes = []*AzureRouteTableRoute{
					{
						Name:          "blackhole-rfc1918",
						AddressPrefix: "192.168.0.0/16",
						NextHopType:   AzureRouteTableNextHopType_NONE,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept disabling BGP route propagation", func() {
				input := validResource()
				disabled := false
				input.Spec.BgpRoutePropagationEnabled = &disabled
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept user tags", func() {
				input := validResource()
				input.Spec.Tags = map[string]string{
					"cost-center": "platform",
					"owner":       "network-team",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_route_table", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name has invalid characters", func() {
				input := validResource()
				input.Spec.Name = "bad name!"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a route omits its name", func() {
				input := validResource()
				input.Spec.Routes = []*AzureRouteTableRoute{
					{
						AddressPrefix: "0.0.0.0/0",
						NextHopType:   AzureRouteTableNextHopType_INTERNET,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a route omits its address prefix", func() {
				input := validResource()
				input.Spec.Routes = []*AzureRouteTableRoute{
					{
						Name:        "incomplete",
						NextHopType: AzureRouteTableNextHopType_INTERNET,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a route omits its next hop type", func() {
				input := validResource()
				input.Spec.Routes = []*AzureRouteTableRoute{
					{
						Name:          "incomplete",
						AddressPrefix: "0.0.0.0/0",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a virtual appliance route omits the next hop IP", func() {
				input := validResource()
				input.Spec.Routes = []*AzureRouteTableRoute{
					{
						Name:          "default-via-firewall",
						AddressPrefix: "0.0.0.0/0",
						NextHopType:   AzureRouteTableNextHopType_VIRTUAL_APPLIANCE,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a non-appliance route carries a next hop IP", func() {
				input := validResource()
				input.Spec.Routes = []*AzureRouteTableRoute{
					{
						Name:               "invalid-hop-ip",
						AddressPrefix:      "0.0.0.0/0",
						NextHopType:        AzureRouteTableNextHopType_INTERNET,
						NextHopInIpAddress: literal("10.0.1.4"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := validResource()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := validResource()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := validResource()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := validResource()
				input.Spec = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
