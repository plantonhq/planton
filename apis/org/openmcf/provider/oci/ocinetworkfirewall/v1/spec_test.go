package ocinetworkfirewallv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciNetworkFirewallSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciNetworkFirewallSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidFirewall() *OciNetworkFirewall {
	return &OciNetworkFirewall{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciNetworkFirewall",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-firewall",
		},
		Spec: &OciNetworkFirewallSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			SubnetId:      newStringValueOrRef("ocid1.subnet.oc1..example"),
			Policy: &OciNetworkFirewallSpec_Policy{
				SecurityRules: []*OciNetworkFirewallSpec_SecurityRule{
					{
						Name:   "allow-all",
						Action: OciNetworkFirewallSpec_SecurityRule_allow,
						Condition: &OciNetworkFirewallSpec_SecurityRuleCondition{
							SourceAddresses: []string{},
						},
					},
				},
			},
		},
	}
}

func fullFirewall() *OciNetworkFirewall {
	maxPort := int32(443)
	return &OciNetworkFirewall{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciNetworkFirewall",
		Metadata: &shared.CloudResourceMetadata{
			Name: "prod-firewall",
			Id:   "ocifw-prod",
			Org:  "acme",
			Env:  "production",
		},
		Spec: &OciNetworkFirewallSpec{
			CompartmentId:      newStringValueOrRef("ocid1.compartment.oc1..example"),
			SubnetId:           newStringValueOrRef("ocid1.subnet.oc1..example"),
			DisplayName:        "production-firewall",
			AvailabilityDomain: "AD-1",
			Shape:              "NGFW1",
			NetworkSecurityGroupIds: []*foreignkeyv1.StringValueOrRef{
				newStringValueOrRef("ocid1.networksecuritygroup.oc1..example"),
			},
			NatConfiguration: &OciNetworkFirewallSpec_NatConfiguration{
				MustEnablePrivateNat: true,
			},
			Policy: &OciNetworkFirewallSpec_Policy{
				DisplayName: "prod-policy",
				Description: "Production firewall policy",
				AddressLists: []*OciNetworkFirewallSpec_AddressList{
					{
						Name:      "internal-networks",
						Type:      OciNetworkFirewallSpec_AddressList_ip,
						Addresses: []string{"10.0.0.0/8", "172.16.0.0/12"},
					},
					{
						Name:      "blocked-domains",
						Type:      OciNetworkFirewallSpec_AddressList_fqdn,
						Addresses: []string{"malware.example.com"},
					},
				},
				Services: []*OciNetworkFirewallSpec_Service{
					{
						Name: "https",
						Type: OciNetworkFirewallSpec_Service_tcp_service,
						PortRanges: []*OciNetworkFirewallSpec_PortRange{
							{MinimumPort: 443, MaximumPort: &maxPort},
						},
					},
					{
						Name: "http",
						Type: OciNetworkFirewallSpec_Service_tcp_service,
						PortRanges: []*OciNetworkFirewallSpec_PortRange{
							{MinimumPort: 80},
						},
					},
				},
				ServiceLists: []*OciNetworkFirewallSpec_ServiceList{
					{
						Name:     "web-services",
						Services: []string{"http", "https"},
					},
				},
				UrlLists: []*OciNetworkFirewallSpec_UrlList{
					{
						Name: "blocked-urls",
						Urls: []*OciNetworkFirewallSpec_UrlPattern{
							{Pattern: "*.malware.example.com"},
						},
					},
				},
				SecurityRules: []*OciNetworkFirewallSpec_SecurityRule{
					{
						Name:   "block-malware-urls",
						Action: OciNetworkFirewallSpec_SecurityRule_drop,
						Condition: &OciNetworkFirewallSpec_SecurityRuleCondition{
							Urls: []string{"blocked-urls"},
						},
						Description: "Drop traffic to known malware URLs",
					},
					{
						Name:   "allow-internal-web",
						Action: OciNetworkFirewallSpec_SecurityRule_allow,
						Condition: &OciNetworkFirewallSpec_SecurityRuleCondition{
							SourceAddresses: []string{"internal-networks"},
							Services:        []string{"web-services"},
						},
					},
					{
						Name:       "inspect-remaining",
						Action:     OciNetworkFirewallSpec_SecurityRule_inspect,
						Inspection: OciNetworkFirewallSpec_SecurityRule_intrusion_prevention,
						Condition: &OciNetworkFirewallSpec_SecurityRuleCondition{
							SourceAddresses: []string{"internal-networks"},
						},
					},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("OciNetworkFirewallSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_network_firewall", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidFirewall()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for full configuration", func() {
				input := fullFirewall()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display name", func() {
				input := minimalValidFirewall()
				input.Spec.DisplayName = "my-firewall"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with static IPv4 address", func() {
				input := minimalValidFirewall()
				input.Spec.Ipv4Address = "10.0.1.50"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with IPv6 address", func() {
				input := minimalValidFirewall()
				input.Spec.Ipv6Address = "2001:db8::1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with availability domain", func() {
				input := minimalValidFirewall()
				input.Spec.AvailabilityDomain = "AD-2"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with shape", func() {
				input := minimalValidFirewall()
				input.Spec.Shape = "NGFW1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with NSG IDs", func() {
				input := minimalValidFirewall()
				input.Spec.NetworkSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.networksecuritygroup.oc1..nsg1"),
					newStringValueOrRef("ocid1.networksecuritygroup.oc1..nsg2"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with NAT configuration", func() {
				input := minimalValidFirewall()
				input.Spec.NatConfiguration = &OciNetworkFirewallSpec_NatConfiguration{
					MustEnablePrivateNat: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidFirewall()
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

			ginkgo.It("should not return a validation error with subnet_id via valueFrom ref", func() {
				input := minimalValidFirewall()
				input.Spec.SubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-subnet",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with IP address list", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.AddressLists = []*OciNetworkFirewallSpec_AddressList{
					{
						Name:      "servers",
						Type:      OciNetworkFirewallSpec_AddressList_ip,
						Addresses: []string{"10.0.1.0/24"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with FQDN address list", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.AddressLists = []*OciNetworkFirewallSpec_AddressList{
					{
						Name:        "external-apis",
						Type:        OciNetworkFirewallSpec_AddressList_fqdn,
						Addresses:   []string{"api.example.com", "auth.example.com"},
						Description: "External API endpoints",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with TCP service", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.Services = []*OciNetworkFirewallSpec_Service{
					{
						Name: "ssh",
						Type: OciNetworkFirewallSpec_Service_tcp_service,
						PortRanges: []*OciNetworkFirewallSpec_PortRange{
							{MinimumPort: 22},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with UDP service", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.Services = []*OciNetworkFirewallSpec_Service{
					{
						Name: "dns",
						Type: OciNetworkFirewallSpec_Service_udp_service,
						PortRanges: []*OciNetworkFirewallSpec_PortRange{
							{MinimumPort: 53},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with port range", func() {
				maxPort := int32(8443)
				input := minimalValidFirewall()
				input.Spec.Policy.Services = []*OciNetworkFirewallSpec_Service{
					{
						Name: "high-ports",
						Type: OciNetworkFirewallSpec_Service_tcp_service,
						PortRanges: []*OciNetworkFirewallSpec_PortRange{
							{MinimumPort: 8080, MaximumPort: &maxPort},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with service list", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.Services = []*OciNetworkFirewallSpec_Service{
					{
						Name: "http",
						Type: OciNetworkFirewallSpec_Service_tcp_service,
						PortRanges: []*OciNetworkFirewallSpec_PortRange{
							{MinimumPort: 80},
						},
					},
				}
				input.Spec.Policy.ServiceLists = []*OciNetworkFirewallSpec_ServiceList{
					{
						Name:     "web",
						Services: []string{"http"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with URL list", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.UrlLists = []*OciNetworkFirewallSpec_UrlList{
					{
						Name: "blocked",
						Urls: []*OciNetworkFirewallSpec_UrlPattern{
							{Pattern: "*.bad-site.com"},
							{Pattern: "malware.example.com/payload"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all security rule actions", func() {
				for _, action := range []OciNetworkFirewallSpec_SecurityRule_Action{
					OciNetworkFirewallSpec_SecurityRule_allow,
					OciNetworkFirewallSpec_SecurityRule_drop,
					OciNetworkFirewallSpec_SecurityRule_reject,
				} {
					input := minimalValidFirewall()
					input.Spec.Policy.SecurityRules[0].Action = action
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error with inspect action and intrusion_detection", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.SecurityRules[0].Action = OciNetworkFirewallSpec_SecurityRule_inspect
				input.Spec.Policy.SecurityRules[0].Inspection = OciNetworkFirewallSpec_SecurityRule_intrusion_detection
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with inspect action and intrusion_prevention", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.SecurityRules[0].Action = OciNetworkFirewallSpec_SecurityRule_inspect
				input.Spec.Policy.SecurityRules[0].Inspection = OciNetworkFirewallSpec_SecurityRule_intrusion_prevention
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple security rules", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.SecurityRules = []*OciNetworkFirewallSpec_SecurityRule{
					{
						Name:   "rule-1",
						Action: OciNetworkFirewallSpec_SecurityRule_drop,
						Condition: &OciNetworkFirewallSpec_SecurityRuleCondition{
							DestinationAddresses: []string{"blocked-ips"},
						},
					},
					{
						Name:   "rule-2",
						Action: OciNetworkFirewallSpec_SecurityRule_allow,
						Condition: &OciNetworkFirewallSpec_SecurityRuleCondition{
							SourceAddresses: []string{"trusted-networks"},
							Services:        []string{"web-services"},
						},
					},
					{
						Name:       "rule-3",
						Action:     OciNetworkFirewallSpec_SecurityRule_inspect,
						Inspection: OciNetworkFirewallSpec_SecurityRule_intrusion_prevention,
						Condition: &OciNetworkFirewallSpec_SecurityRuleCondition{
							SourceAddresses: []string{"internal"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with policy display name and description", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.DisplayName = "custom-policy"
				input.Spec.Policy.Description = "A custom firewall policy"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with single-port range (min == max)", func() {
				maxPort := int32(443)
				input := minimalValidFirewall()
				input.Spec.Policy.Services = []*OciNetworkFirewallSpec_Service{
					{
						Name: "https-exact",
						Type: OciNetworkFirewallSpec_Service_tcp_service,
						PortRanges: []*OciNetworkFirewallSpec_PortRange{
							{MinimumPort: 443, MaximumPort: &maxPort},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_network_firewall", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidFirewall()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidFirewall()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidFirewall()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciNetworkFirewall{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciNetworkFirewall",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidFirewall()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet_id is missing", func() {
				input := minimalValidFirewall()
				input.Spec.SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when policy is missing", func() {
				input := minimalValidFirewall()
				input.Spec.Policy = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when address list name is empty", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.AddressLists = []*OciNetworkFirewallSpec_AddressList{
					{
						Name:      "",
						Type:      OciNetworkFirewallSpec_AddressList_ip,
						Addresses: []string{"10.0.0.0/8"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when address list type is unspecified", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.AddressLists = []*OciNetworkFirewallSpec_AddressList{
					{
						Name:      "test-list",
						Type:      OciNetworkFirewallSpec_AddressList_unspecified,
						Addresses: []string{"10.0.0.0/8"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when address list has no addresses", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.AddressLists = []*OciNetworkFirewallSpec_AddressList{
					{
						Name:      "empty-list",
						Type:      OciNetworkFirewallSpec_AddressList_ip,
						Addresses: []string{},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when service name is empty", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.Services = []*OciNetworkFirewallSpec_Service{
					{
						Name: "",
						Type: OciNetworkFirewallSpec_Service_tcp_service,
						PortRanges: []*OciNetworkFirewallSpec_PortRange{
							{MinimumPort: 80},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when service type is unspecified", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.Services = []*OciNetworkFirewallSpec_Service{
					{
						Name: "test-svc",
						Type: OciNetworkFirewallSpec_Service_service_type_unspecified,
						PortRanges: []*OciNetworkFirewallSpec_PortRange{
							{MinimumPort: 80},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when service has no port ranges", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.Services = []*OciNetworkFirewallSpec_Service{
					{
						Name:       "no-ports",
						Type:       OciNetworkFirewallSpec_Service_tcp_service,
						PortRanges: []*OciNetworkFirewallSpec_PortRange{},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when port range minimum is 0", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.Services = []*OciNetworkFirewallSpec_Service{
					{
						Name: "bad-port",
						Type: OciNetworkFirewallSpec_Service_tcp_service,
						PortRanges: []*OciNetworkFirewallSpec_PortRange{
							{MinimumPort: 0},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when port range minimum exceeds 65535", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.Services = []*OciNetworkFirewallSpec_Service{
					{
						Name: "overflow-port",
						Type: OciNetworkFirewallSpec_Service_tcp_service,
						PortRanges: []*OciNetworkFirewallSpec_PortRange{
							{MinimumPort: 65536},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when maximum_port < minimum_port", func() {
				maxPort := int32(79)
				input := minimalValidFirewall()
				input.Spec.Policy.Services = []*OciNetworkFirewallSpec_Service{
					{
						Name: "reversed-range",
						Type: OciNetworkFirewallSpec_Service_tcp_service,
						PortRanges: []*OciNetworkFirewallSpec_PortRange{
							{MinimumPort: 80, MaximumPort: &maxPort},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when URL list name is empty", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.UrlLists = []*OciNetworkFirewallSpec_UrlList{
					{
						Name: "",
						Urls: []*OciNetworkFirewallSpec_UrlPattern{
							{Pattern: "*.example.com"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when URL list has no URLs", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.UrlLists = []*OciNetworkFirewallSpec_UrlList{
					{
						Name: "empty-urls",
						Urls: []*OciNetworkFirewallSpec_UrlPattern{},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when URL pattern is empty", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.UrlLists = []*OciNetworkFirewallSpec_UrlList{
					{
						Name: "bad-url",
						Urls: []*OciNetworkFirewallSpec_UrlPattern{
							{Pattern: ""},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when security rule name is empty", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.SecurityRules[0].Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when security rule action is unspecified", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.SecurityRules[0].Action = OciNetworkFirewallSpec_SecurityRule_action_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when security rule condition is missing", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.SecurityRules[0].Condition = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when inspect action lacks inspection type", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.SecurityRules[0].Action = OciNetworkFirewallSpec_SecurityRule_inspect
				input.Spec.Policy.SecurityRules[0].Inspection = OciNetworkFirewallSpec_SecurityRule_inspection_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when service list name is empty", func() {
				input := minimalValidFirewall()
				input.Spec.Policy.ServiceLists = []*OciNetworkFirewallSpec_ServiceList{
					{
						Name:     "",
						Services: []string{"http"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is empty", func() {
				input := minimalValidFirewall()
				input.ApiVersion = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is empty", func() {
				input := minimalValidFirewall()
				input.Kind = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
