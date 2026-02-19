package ocidynamicroutinggatewayv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciDynamicRoutingGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciDynamicRoutingGatewaySpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidDrg() *OciDynamicRoutingGateway {
	return &OciDynamicRoutingGateway{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciDynamicRoutingGateway",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-drg",
		},
		Spec: &OciDynamicRoutingGatewaySpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
		},
	}
}

func drgWithVcnAttachment() *OciDynamicRoutingGateway {
	drg := minimalValidDrg()
	drg.Spec.Attachments = []*OciDynamicRoutingGatewaySpec_DrgAttachment{
		{
			DisplayName: "spoke-vcn-1",
			NetworkDetails: &OciDynamicRoutingGatewaySpec_NetworkDetails{
				Type: OciDynamicRoutingGatewaySpec_NetworkDetails_vcn,
				Id:   newStringValueOrRef("ocid1.vcn.oc1.iad.example"),
			},
		},
	}
	return drg
}

var _ = ginkgo.Describe("OciDynamicRoutingGatewaySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_core_drg", func() {

			ginkgo.It("should not return a validation error for minimal valid fields (DRG only)", func() {
				input := minimalValidDrg()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display name set", func() {
				input := minimalValidDrg()
				input.Spec.DisplayName = "hub-drg-prod"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via value_from ref", func() {
				input := minimalValidDrg()
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

			ginkgo.It("should not return a validation error with a single VCN attachment", func() {
				input := drgWithVcnAttachment()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple VCN attachments", func() {
				input := minimalValidDrg()
				input.Spec.Attachments = []*OciDynamicRoutingGatewaySpec_DrgAttachment{
					{
						DisplayName: "spoke-vcn-1",
						NetworkDetails: &OciDynamicRoutingGatewaySpec_NetworkDetails{
							Type: OciDynamicRoutingGatewaySpec_NetworkDetails_vcn,
							Id:   newStringValueOrRef("ocid1.vcn.oc1.iad.aaa"),
						},
					},
					{
						DisplayName: "spoke-vcn-2",
						NetworkDetails: &OciDynamicRoutingGatewaySpec_NetworkDetails{
							Type: OciDynamicRoutingGatewaySpec_NetworkDetails_vcn,
							Id:   newStringValueOrRef("ocid1.vcn.oc1.iad.bbb"),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with IPSec tunnel attachment", func() {
				input := minimalValidDrg()
				input.Spec.Attachments = []*OciDynamicRoutingGatewaySpec_DrgAttachment{
					{
						DisplayName: "onprem-vpn",
						NetworkDetails: &OciDynamicRoutingGatewaySpec_NetworkDetails{
							Type: OciDynamicRoutingGatewaySpec_NetworkDetails_ipsec_tunnel,
							Id:   newStringValueOrRef("ocid1.ipsecconnection.oc1.iad.example"),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with virtual circuit attachment", func() {
				input := minimalValidDrg()
				input.Spec.Attachments = []*OciDynamicRoutingGatewaySpec_DrgAttachment{
					{
						DisplayName: "fastconnect",
						NetworkDetails: &OciDynamicRoutingGatewaySpec_NetworkDetails{
							Type: OciDynamicRoutingGatewaySpec_NetworkDetails_virtual_circuit,
							Id:   newStringValueOrRef("ocid1.virtualcircuit.oc1.iad.example"),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with remote peering connection attachment", func() {
				input := minimalValidDrg()
				input.Spec.Attachments = []*OciDynamicRoutingGatewaySpec_DrgAttachment{
					{
						DisplayName: "cross-region-peer",
						NetworkDetails: &OciDynamicRoutingGatewaySpec_NetworkDetails{
							Type: OciDynamicRoutingGatewaySpec_NetworkDetails_remote_peering_connection,
							Id:   newStringValueOrRef("ocid1.remotepeeringconnection.oc1.iad.example"),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with loopback attachment", func() {
				input := minimalValidDrg()
				input.Spec.Attachments = []*OciDynamicRoutingGatewaySpec_DrgAttachment{
					{
						DisplayName: "loopback-1",
						NetworkDetails: &OciDynamicRoutingGatewaySpec_NetworkDetails{
							Type: OciDynamicRoutingGatewaySpec_NetworkDetails_loopback,
							Id:   newStringValueOrRef("ocid1.drg.oc1.iad.example"),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with VCN attachment using transit routing", func() {
				input := drgWithVcnAttachment()
				input.Spec.Attachments[0].NetworkDetails.RouteTableId = "ocid1.routetable.oc1.iad.example"
				input.Spec.Attachments[0].NetworkDetails.VcnRouteType = OciDynamicRoutingGatewaySpec_NetworkDetails_subnet_cidrs
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with VCN attachment using vcn_cidrs route type", func() {
				input := drgWithVcnAttachment()
				input.Spec.Attachments[0].NetworkDetails.VcnRouteType = OciDynamicRoutingGatewaySpec_NetworkDetails_vcn_cidrs
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with attachment referencing a route table by name", func() {
				input := drgWithVcnAttachment()
				input.Spec.Attachments[0].DrgRouteTableName = "custom-rt"
				input.Spec.RouteTables = []*OciDynamicRoutingGatewaySpec_DrgRouteTable{
					{DisplayName: "custom-rt"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with attachment referencing an export distribution by name", func() {
				input := drgWithVcnAttachment()
				input.Spec.Attachments[0].ExportDrgRouteDistributionName = "custom-export"
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "custom-export",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_export_routes,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a custom route table", func() {
				input := minimalValidDrg()
				input.Spec.RouteTables = []*OciDynamicRoutingGatewaySpec_DrgRouteTable{
					{DisplayName: "spoke-rt"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with ECMP enabled on route table", func() {
				input := minimalValidDrg()
				input.Spec.RouteTables = []*OciDynamicRoutingGatewaySpec_DrgRouteTable{
					{
						DisplayName:   "ecmp-rt",
						IsEcmpEnabled: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with route table importing from a distribution", func() {
				input := minimalValidDrg()
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "vcn-import",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_import_routes,
					},
				}
				input.Spec.RouteTables = []*OciDynamicRoutingGatewaySpec_DrgRouteTable{
					{
						DisplayName:                    "spoke-rt",
						ImportDrgRouteDistributionName: "vcn-import",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with static route rules", func() {
				input := drgWithVcnAttachment()
				input.Spec.RouteTables = []*OciDynamicRoutingGatewaySpec_DrgRouteTable{
					{
						DisplayName: "hub-rt",
						StaticRouteRules: []*OciDynamicRoutingGatewaySpec_StaticRouteRule{
							{
								Destination:          "10.0.0.0/8",
								NextHopAttachmentName: "spoke-vcn-1",
							},
							{
								Destination:          "172.16.0.0/12",
								NextHopAttachmentName: "spoke-vcn-1",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with an import route distribution", func() {
				input := minimalValidDrg()
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "vcn-import",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_import_routes,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with an export route distribution", func() {
				input := minimalValidDrg()
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "custom-export",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_export_routes,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with distribution statements using match_all", func() {
				input := minimalValidDrg()
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "accept-all",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_import_routes,
						Statements: []*OciDynamicRoutingGatewaySpec_DistributionStatement{
							{
								Priority: 1,
								MatchCriteria: &OciDynamicRoutingGatewaySpec_MatchCriteria{
									MatchType: OciDynamicRoutingGatewaySpec_MatchCriteria_match_all,
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with distribution statement using drg_attachment_type", func() {
				input := minimalValidDrg()
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "vcn-only",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_import_routes,
						Statements: []*OciDynamicRoutingGatewaySpec_DistributionStatement{
							{
								Priority: 10,
								MatchCriteria: &OciDynamicRoutingGatewaySpec_MatchCriteria{
									MatchType:      OciDynamicRoutingGatewaySpec_MatchCriteria_drg_attachment_type,
									AttachmentType: "VCN",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with distribution statement using drg_attachment_id", func() {
				input := drgWithVcnAttachment()
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "specific-attachment",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_import_routes,
						Statements: []*OciDynamicRoutingGatewaySpec_DistributionStatement{
							{
								Priority: 5,
								MatchCriteria: &OciDynamicRoutingGatewaySpec_MatchCriteria{
									MatchType:         OciDynamicRoutingGatewaySpec_MatchCriteria_drg_attachment_id,
									DrgAttachmentName: "spoke-vcn-1",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple distribution statements", func() {
				input := minimalValidDrg()
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "multi-statement",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_import_routes,
						Statements: []*OciDynamicRoutingGatewaySpec_DistributionStatement{
							{
								Priority: 1,
								MatchCriteria: &OciDynamicRoutingGatewaySpec_MatchCriteria{
									MatchType:      OciDynamicRoutingGatewaySpec_MatchCriteria_drg_attachment_type,
									AttachmentType: "VCN",
								},
							},
							{
								Priority: 10,
								MatchCriteria: &OciDynamicRoutingGatewaySpec_MatchCriteria{
									MatchType:      OciDynamicRoutingGatewaySpec_MatchCriteria_drg_attachment_type,
									AttachmentType: "IPSEC_TUNNEL",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a full hub-and-spoke setup", func() {
				input := minimalValidDrg()
				input.Spec.DisplayName = "hub-drg"
				input.Spec.Attachments = []*OciDynamicRoutingGatewaySpec_DrgAttachment{
					{
						DisplayName: "spoke-a",
						NetworkDetails: &OciDynamicRoutingGatewaySpec_NetworkDetails{
							Type: OciDynamicRoutingGatewaySpec_NetworkDetails_vcn,
							Id:   newStringValueOrRef("ocid1.vcn.oc1.iad.aaa"),
						},
						DrgRouteTableName: "spoke-rt",
					},
					{
						DisplayName: "spoke-b",
						NetworkDetails: &OciDynamicRoutingGatewaySpec_NetworkDetails{
							Type: OciDynamicRoutingGatewaySpec_NetworkDetails_vcn,
							Id:   newStringValueOrRef("ocid1.vcn.oc1.iad.bbb"),
						},
						DrgRouteTableName: "spoke-rt",
					},
				}
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "vcn-import",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_import_routes,
						Statements: []*OciDynamicRoutingGatewaySpec_DistributionStatement{
							{
								Priority: 1,
								MatchCriteria: &OciDynamicRoutingGatewaySpec_MatchCriteria{
									MatchType:      OciDynamicRoutingGatewaySpec_MatchCriteria_drg_attachment_type,
									AttachmentType: "VCN",
								},
							},
						},
					},
				}
				input.Spec.RouteTables = []*OciDynamicRoutingGatewaySpec_DrgRouteTable{
					{
						DisplayName:                    "spoke-rt",
						ImportDrgRouteDistributionName: "vcn-import",
						StaticRouteRules: []*OciDynamicRoutingGatewaySpec_StaticRouteRule{
							{
								Destination:          "10.0.0.0/16",
								NextHopAttachmentName: "spoke-a",
							},
							{
								Destination:          "10.1.0.0/16",
								NextHopAttachmentName: "spoke-b",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with network details id via value_from ref", func() {
				input := minimalValidDrg()
				input.Spec.Attachments = []*OciDynamicRoutingGatewaySpec_DrgAttachment{
					{
						DisplayName: "spoke-vcn",
						NetworkDetails: &OciDynamicRoutingGatewaySpec_NetworkDetails{
							Type: OciDynamicRoutingGatewaySpec_NetworkDetails_vcn,
							Id: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
									ValueFrom: &foreignkeyv1.ValueFromRef{
										Name: "my-vcn",
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_core_drg", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidDrg()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidDrg()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidDrg()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciDynamicRoutingGateway{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciDynamicRoutingGateway",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-drg"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidDrg()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when attachment display_name is empty", func() {
				input := minimalValidDrg()
				input.Spec.Attachments = []*OciDynamicRoutingGatewaySpec_DrgAttachment{
					{
						DisplayName: "",
						NetworkDetails: &OciDynamicRoutingGatewaySpec_NetworkDetails{
							Type: OciDynamicRoutingGatewaySpec_NetworkDetails_vcn,
							Id:   newStringValueOrRef("ocid1.vcn.oc1.iad.example"),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when attachment network_details is missing", func() {
				input := minimalValidDrg()
				input.Spec.Attachments = []*OciDynamicRoutingGatewaySpec_DrgAttachment{
					{
						DisplayName:    "broken",
						NetworkDetails: nil,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network_details type is unspecified", func() {
				input := minimalValidDrg()
				input.Spec.Attachments = []*OciDynamicRoutingGatewaySpec_DrgAttachment{
					{
						DisplayName: "broken",
						NetworkDetails: &OciDynamicRoutingGatewaySpec_NetworkDetails{
							Type: OciDynamicRoutingGatewaySpec_NetworkDetails_network_type_unspecified,
							Id:   newStringValueOrRef("ocid1.vcn.oc1.iad.example"),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network_details id is missing", func() {
				input := minimalValidDrg()
				input.Spec.Attachments = []*OciDynamicRoutingGatewaySpec_DrgAttachment{
					{
						DisplayName: "broken",
						NetworkDetails: &OciDynamicRoutingGatewaySpec_NetworkDetails{
							Type: OciDynamicRoutingGatewaySpec_NetworkDetails_vcn,
							Id:   nil,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when route table display_name is empty", func() {
				input := minimalValidDrg()
				input.Spec.RouteTables = []*OciDynamicRoutingGatewaySpec_DrgRouteTable{
					{DisplayName: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when static route rule destination is empty", func() {
				input := drgWithVcnAttachment()
				input.Spec.RouteTables = []*OciDynamicRoutingGatewaySpec_DrgRouteTable{
					{
						DisplayName: "bad-rt",
						StaticRouteRules: []*OciDynamicRoutingGatewaySpec_StaticRouteRule{
							{
								Destination:          "",
								NextHopAttachmentName: "spoke-vcn-1",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when static route rule next_hop_attachment_name is empty", func() {
				input := drgWithVcnAttachment()
				input.Spec.RouteTables = []*OciDynamicRoutingGatewaySpec_DrgRouteTable{
					{
						DisplayName: "bad-rt",
						StaticRouteRules: []*OciDynamicRoutingGatewaySpec_StaticRouteRule{
							{
								Destination:          "10.0.0.0/8",
								NextHopAttachmentName: "",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when distribution display_name is empty", func() {
				input := minimalValidDrg()
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_import_routes,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when distribution type is unspecified", func() {
				input := minimalValidDrg()
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "bad-dist",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_distribution_type_unspecified,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when statement priority is zero", func() {
				input := minimalValidDrg()
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "bad-dist",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_import_routes,
						Statements: []*OciDynamicRoutingGatewaySpec_DistributionStatement{
							{
								Priority: 0,
								MatchCriteria: &OciDynamicRoutingGatewaySpec_MatchCriteria{
									MatchType: OciDynamicRoutingGatewaySpec_MatchCriteria_match_all,
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when statement priority exceeds 65535", func() {
				input := minimalValidDrg()
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "bad-dist",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_import_routes,
						Statements: []*OciDynamicRoutingGatewaySpec_DistributionStatement{
							{
								Priority: 70000,
								MatchCriteria: &OciDynamicRoutingGatewaySpec_MatchCriteria{
									MatchType: OciDynamicRoutingGatewaySpec_MatchCriteria_match_all,
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when statement match_criteria is missing", func() {
				input := minimalValidDrg()
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "bad-dist",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_import_routes,
						Statements: []*OciDynamicRoutingGatewaySpec_DistributionStatement{
							{
								Priority:      1,
								MatchCriteria: nil,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when match_criteria match_type is unspecified", func() {
				input := minimalValidDrg()
				input.Spec.RouteDistributions = []*OciDynamicRoutingGatewaySpec_DrgRouteDistribution{
					{
						DisplayName:      "bad-dist",
						DistributionType: OciDynamicRoutingGatewaySpec_DrgRouteDistribution_import_routes,
						Statements: []*OciDynamicRoutingGatewaySpec_DistributionStatement{
							{
								Priority: 1,
								MatchCriteria: &OciDynamicRoutingGatewaySpec_MatchCriteria{
									MatchType: OciDynamicRoutingGatewaySpec_MatchCriteria_match_type_unspecified,
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
