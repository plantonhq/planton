package ocisecuritygroupv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestOciSecurityGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciSecurityGroupSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidNsg() *OciSecurityGroup {
	return &OciSecurityGroup{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciSecurityGroup",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-nsg",
		},
		Spec: &OciSecurityGroupSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			VcnId:         newStringValueOrRef("ocid1.vcn.oc1.iad.example"),
		},
	}
}

var _ = ginkgo.Describe("OciSecurityGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_network_security_group", func() {

			ginkgo.It("should not return a validation error for minimal valid fields (no rules)", func() {
				input := minimalValidNsg()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display_name set", func() {
				input := minimalValidNsg()
				input.Spec.DisplayName = "Web Tier NSG"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with ingress rules only", func() {
				input := minimalValidNsg()
				input.Spec.IngressRules = []*OciSecurityGroupSpec_IngressRule{
					{
						Source:   "0.0.0.0/0",
						Protocol: OciSecurityGroupSpec_tcp,
						TcpOptions: &OciSecurityGroupSpec_TcpOptions{
							DestinationPortRange: &OciSecurityGroupSpec_PortRange{
								Min: 443,
								Max: 443,
							},
						},
						Description: "Allow HTTPS from internet",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with egress rules only", func() {
				input := minimalValidNsg()
				input.Spec.EgressRules = []*OciSecurityGroupSpec_EgressRule{
					{
						Destination: "0.0.0.0/0",
						Protocol:    OciSecurityGroupSpec_all,
						Description: "Allow all outbound",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with both ingress and egress rules", func() {
				input := minimalValidNsg()
				input.Spec.IngressRules = []*OciSecurityGroupSpec_IngressRule{
					{
						Source:     "10.0.0.0/16",
						SourceType: OciSecurityGroupSpec_cidr_block,
						Protocol:   OciSecurityGroupSpec_tcp,
						TcpOptions: &OciSecurityGroupSpec_TcpOptions{
							DestinationPortRange: &OciSecurityGroupSpec_PortRange{
								Min: 80,
								Max: 80,
							},
						},
					},
				}
				input.Spec.EgressRules = []*OciSecurityGroupSpec_EgressRule{
					{
						Destination:     "0.0.0.0/0",
						DestinationType: OciSecurityGroupSpec_cidr_block,
						Protocol:        OciSecurityGroupSpec_all,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all protocol types", func() {
				input := minimalValidNsg()
				input.Spec.IngressRules = []*OciSecurityGroupSpec_IngressRule{
					{Source: "0.0.0.0/0", Protocol: OciSecurityGroupSpec_all},
					{
						Source: "10.0.0.0/8", Protocol: OciSecurityGroupSpec_tcp,
						TcpOptions: &OciSecurityGroupSpec_TcpOptions{
							DestinationPortRange: &OciSecurityGroupSpec_PortRange{Min: 22, Max: 22},
						},
					},
					{
						Source: "10.0.0.0/8", Protocol: OciSecurityGroupSpec_udp,
						UdpOptions: &OciSecurityGroupSpec_UdpOptions{
							DestinationPortRange: &OciSecurityGroupSpec_PortRange{Min: 53, Max: 53},
						},
					},
					{
						Source: "10.0.0.0/8", Protocol: OciSecurityGroupSpec_icmp,
						IcmpOptions: &OciSecurityGroupSpec_IcmpOptions{Type: 3, Code: proto.Int32(4)},
					},
					{
						Source: "::/0", Protocol: OciSecurityGroupSpec_icmpv6,
						IcmpOptions: &OciSecurityGroupSpec_IcmpOptions{Type: 1},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with service_cidr_block source type", func() {
				input := minimalValidNsg()
				input.Spec.IngressRules = []*OciSecurityGroupSpec_IngressRule{
					{
						Source:     "all-iad-services-in-oracle-services-network",
						SourceType: OciSecurityGroupSpec_service_cidr_block,
						Protocol:   OciSecurityGroupSpec_tcp,
						TcpOptions: &OciSecurityGroupSpec_TcpOptions{
							DestinationPortRange: &OciSecurityGroupSpec_PortRange{Min: 443, Max: 443},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with network_security_group source type", func() {
				input := minimalValidNsg()
				input.Spec.IngressRules = []*OciSecurityGroupSpec_IngressRule{
					{
						Source:     "ocid1.networksecuritygroup.oc1.iad.example",
						SourceType: OciSecurityGroupSpec_network_security_group,
						Protocol:   OciSecurityGroupSpec_all,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with stateless rules", func() {
				input := minimalValidNsg()
				input.Spec.IngressRules = []*OciSecurityGroupSpec_IngressRule{
					{
						Source:    "0.0.0.0/0",
						Protocol:  OciSecurityGroupSpec_tcp,
						Stateless: true,
						TcpOptions: &OciSecurityGroupSpec_TcpOptions{
							DestinationPortRange: &OciSecurityGroupSpec_PortRange{Min: 443, Max: 443},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with TCP source and destination port ranges", func() {
				input := minimalValidNsg()
				input.Spec.IngressRules = []*OciSecurityGroupSpec_IngressRule{
					{
						Source:   "10.0.0.0/8",
						Protocol: OciSecurityGroupSpec_tcp,
						TcpOptions: &OciSecurityGroupSpec_TcpOptions{
							DestinationPortRange: &OciSecurityGroupSpec_PortRange{Min: 8080, Max: 8443},
							SourcePortRange:      &OciSecurityGroupSpec_PortRange{Min: 1024, Max: 65535},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with ICMP type only (no code)", func() {
				input := minimalValidNsg()
				input.Spec.IngressRules = []*OciSecurityGroupSpec_IngressRule{
					{
						Source:      "10.0.0.0/16",
						Protocol:    OciSecurityGroupSpec_icmp,
						IcmpOptions: &OciSecurityGroupSpec_IcmpOptions{Type: 8},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with ICMP code 0 (valid code)", func() {
				input := minimalValidNsg()
				input.Spec.IngressRules = []*OciSecurityGroupSpec_IngressRule{
					{
						Source:      "10.0.0.0/16",
						Protocol:    OciSecurityGroupSpec_icmp,
						IcmpOptions: &OciSecurityGroupSpec_IcmpOptions{Type: 3, Code: proto.Int32(0)},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with exactly 120 total rules", func() {
				input := minimalValidNsg()
				for i := 0; i < 60; i++ {
					input.Spec.IngressRules = append(input.Spec.IngressRules,
						&OciSecurityGroupSpec_IngressRule{
							Source:   "10.0.0.0/8",
							Protocol: OciSecurityGroupSpec_all,
						})
				}
				for i := 0; i < 60; i++ {
					input.Spec.EgressRules = append(input.Spec.EgressRules,
						&OciSecurityGroupSpec_EgressRule{
							Destination: "0.0.0.0/0",
							Protocol:    OciSecurityGroupSpec_all,
						})
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via value_from ref", func() {
				input := minimalValidNsg()
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
				input := minimalValidNsg()
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

			ginkgo.It("should not return a validation error with all optional metadata fields", func() {
				input := minimalValidNsg()
				input.Metadata.Org = "acme-corp"
				input.Metadata.Env = "production"
				input.Metadata.Labels = map[string]string{"team": "platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_network_security_group", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidNsg()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidNsg()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidNsg()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciSecurityGroup{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciSecurityGroup",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-nsg"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidNsg()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vcn_id is missing", func() {
				input := minimalValidNsg()
				input.Spec.VcnId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ingress rule source is empty", func() {
				input := minimalValidNsg()
				input.Spec.IngressRules = []*OciSecurityGroupSpec_IngressRule{
					{Source: "", Protocol: OciSecurityGroupSpec_tcp},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when egress rule destination is empty", func() {
				input := minimalValidNsg()
				input.Spec.EgressRules = []*OciSecurityGroupSpec_EgressRule{
					{Destination: "", Protocol: OciSecurityGroupSpec_all},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when total rules exceed 120", func() {
				input := minimalValidNsg()
				for i := 0; i < 61; i++ {
					input.Spec.IngressRules = append(input.Spec.IngressRules,
						&OciSecurityGroupSpec_IngressRule{
							Source:   "10.0.0.0/8",
							Protocol: OciSecurityGroupSpec_all,
						})
				}
				for i := 0; i < 60; i++ {
					input.Spec.EgressRules = append(input.Spec.EgressRules,
						&OciSecurityGroupSpec_EgressRule{
							Destination: "0.0.0.0/0",
							Protocol:    OciSecurityGroupSpec_all,
						})
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when port range min > max", func() {
				input := minimalValidNsg()
				input.Spec.IngressRules = []*OciSecurityGroupSpec_IngressRule{
					{
						Source:   "10.0.0.0/8",
						Protocol: OciSecurityGroupSpec_tcp,
						TcpOptions: &OciSecurityGroupSpec_TcpOptions{
							DestinationPortRange: &OciSecurityGroupSpec_PortRange{
								Min: 8443,
								Max: 8080,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when port is out of range (0)", func() {
				input := minimalValidNsg()
				input.Spec.IngressRules = []*OciSecurityGroupSpec_IngressRule{
					{
						Source:   "10.0.0.0/8",
						Protocol: OciSecurityGroupSpec_tcp,
						TcpOptions: &OciSecurityGroupSpec_TcpOptions{
							DestinationPortRange: &OciSecurityGroupSpec_PortRange{
								Min: 0,
								Max: 80,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when port exceeds 65535", func() {
				input := minimalValidNsg()
				input.Spec.IngressRules = []*OciSecurityGroupSpec_IngressRule{
					{
						Source:   "10.0.0.0/8",
						Protocol: OciSecurityGroupSpec_tcp,
						TcpOptions: &OciSecurityGroupSpec_TcpOptions{
							DestinationPortRange: &OciSecurityGroupSpec_PortRange{
								Min: 80,
								Max: 70000,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ingress rule protocol is unspecified", func() {
				input := minimalValidNsg()
				input.Spec.IngressRules = []*OciSecurityGroupSpec_IngressRule{
					{Source: "10.0.0.0/8", Protocol: OciSecurityGroupSpec_protocol_unspecified},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
