package ocisubnetv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciSubnetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciSubnetSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidSubnet() *OciSubnet {
	return &OciSubnet{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciSubnet",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-subnet",
		},
		Spec: &OciSubnetSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			VcnId:         newStringValueOrRef("ocid1.vcn.oc1.iad.example"),
			CidrBlock:     "10.0.1.0/24",
		},
	}
}

var _ = ginkgo.Describe("OciSubnetSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_subnet", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidSubnet()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional fields set", func() {
				input := &OciSubnet{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-subnet",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OciSubnetSpec{
						CompartmentId:            newStringValueOrRef("ocid1.compartment.oc1..example"),
						VcnId:                    newStringValueOrRef("ocid1.vcn.oc1.iad.example"),
						CidrBlock:                "10.0.1.0/24",
						DisplayName:              "Production Private Subnet",
						DnsLabel:                 "priv1",
						AvailabilityDomain:       "Iocq:US-ASHBURN-AD-1",
						ProhibitPublicIpOnVnic:   true,
						ProhibitInternetIngress:  true,
						DhcpOptionsId:            newStringValueOrRef("ocid1.dhcpoptions.oc1.iad.example"),
						Ipv6CidrBlock:            "2001:0db8:0123:1111::/64",
						SecurityListIds: []*foreignkeyv1.StringValueOrRef{
							newStringValueOrRef("ocid1.securitylist.oc1.iad.example1"),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a public subnet", func() {
				input := minimalValidSubnet()
				input.Spec.ProhibitPublicIpOnVnic = false
				input.Spec.ProhibitInternetIngress = false
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with route_table_id only", func() {
				input := minimalValidSubnet()
				input.Spec.RouteTableId = newStringValueOrRef("ocid1.routetable.oc1.iad.example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with route_rules only", func() {
				input := minimalValidSubnet()
				input.Spec.RouteRules = []*OciSubnetSpec_RouteRule{
					{
						Destination:     "0.0.0.0/0",
						DestinationType: OciSubnetSpec_RouteRule_cidr_block,
						NetworkEntityId: newStringValueOrRef("ocid1.internetgateway.oc1.iad.example"),
						Description:     "Route all traffic to internet gateway",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple route rules", func() {
				input := minimalValidSubnet()
				input.Spec.RouteRules = []*OciSubnetSpec_RouteRule{
					{
						Destination:     "0.0.0.0/0",
						DestinationType: OciSubnetSpec_RouteRule_cidr_block,
						NetworkEntityId: newStringValueOrRef("ocid1.natgateway.oc1.iad.example"),
						Description:     "Route internet traffic via NAT gateway",
					},
					{
						Destination:     "all-iad-services-in-oracle-services-network",
						DestinationType: OciSubnetSpec_RouteRule_service_cidr_block,
						NetworkEntityId: newStringValueOrRef("ocid1.servicegateway.oc1.iad.example"),
						Description:     "Route OCI services via service gateway",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with neither route_table_id nor route_rules", func() {
				input := minimalValidSubnet()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via value_from ref", func() {
				input := minimalValidSubnet()
				input.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-compartment",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with vcn_id via value_from ref", func() {
				input := minimalValidSubnet()
				input.Spec.VcnId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-vcn",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with maximum 5 security lists", func() {
				input := minimalValidSubnet()
				input.Spec.SecurityListIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.securitylist.oc1.iad.a"),
					newStringValueOrRef("ocid1.securitylist.oc1.iad.b"),
					newStringValueOrRef("ocid1.securitylist.oc1.iad.c"),
					newStringValueOrRef("ocid1.securitylist.oc1.iad.d"),
					newStringValueOrRef("ocid1.securitylist.oc1.iad.e"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_subnet", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidSubnet()
				input.ApiVersion = "wrong.openmcf.org/v1"
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
				input := &OciSubnet{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidSubnet()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vcn_id is missing", func() {
				input := minimalValidSubnet()
				input.Spec.VcnId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cidr_block is empty", func() {
				input := minimalValidSubnet()
				input.Spec.CidrBlock = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when route_table_id and route_rules are both set", func() {
				input := minimalValidSubnet()
				input.Spec.RouteTableId = newStringValueOrRef("ocid1.routetable.oc1.iad.example")
				input.Spec.RouteRules = []*OciSubnetSpec_RouteRule{
					{
						Destination:     "0.0.0.0/0",
						DestinationType: OciSubnetSpec_RouteRule_cidr_block,
						NetworkEntityId: newStringValueOrRef("ocid1.internetgateway.oc1.iad.example"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when security_list_ids exceeds 5", func() {
				input := minimalValidSubnet()
				input.Spec.SecurityListIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.securitylist.oc1.iad.a"),
					newStringValueOrRef("ocid1.securitylist.oc1.iad.b"),
					newStringValueOrRef("ocid1.securitylist.oc1.iad.c"),
					newStringValueOrRef("ocid1.securitylist.oc1.iad.d"),
					newStringValueOrRef("ocid1.securitylist.oc1.iad.e"),
					newStringValueOrRef("ocid1.securitylist.oc1.iad.f"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when route_rule is missing destination", func() {
				input := minimalValidSubnet()
				input.Spec.RouteRules = []*OciSubnetSpec_RouteRule{
					{
						Destination:     "",
						DestinationType: OciSubnetSpec_RouteRule_cidr_block,
						NetworkEntityId: newStringValueOrRef("ocid1.internetgateway.oc1.iad.example"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when route_rule is missing network_entity_id", func() {
				input := minimalValidSubnet()
				input.Spec.RouteRules = []*OciSubnetSpec_RouteRule{
					{
						Destination:     "0.0.0.0/0",
						DestinationType: OciSubnetSpec_RouteRule_cidr_block,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
