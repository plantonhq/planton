package azurenetworksecuritygroupv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAzureNetworkSecurityGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureNetworkSecurityGroupSpec Validation Tests")
}

var _ = ginkgo.Describe("AzureNetworkSecurityGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_network_security_group", func() {

			ginkgo.It("should not return a validation error for minimal valid fields (no rules)", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full metadata", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-web-nsg",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "westeurope",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-network-rg",
							},
						},
						Name: "prod-web-nsg",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a single inbound allow rule", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "allow-https",
								Priority:             100,
								Direction:            "Inbound",
								Access:               "Allow",
								Protocol:             "Tcp",
								DestinationPortRange: "443",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple rules", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "web-tier-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-rg",
							},
						},
						Name: "web-tier-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "allow-https",
								Priority:             100,
								Direction:            "Inbound",
								Access:               "Allow",
								Protocol:             "Tcp",
								DestinationPortRange: "443",
							},
							{
								Name:                 "allow-http",
								Priority:             200,
								Direction:            "Inbound",
								Access:               "Allow",
								Protocol:             "Tcp",
								DestinationPortRange: "80",
							},
							{
								Name:                 "deny-all-inbound",
								Priority:             4096,
								Direction:            "Inbound",
								Access:               "Deny",
								Protocol:             "*",
								DestinationPortRange: "*",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with rule description", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "allow-ssh",
								Description:          "Allow SSH from corporate VPN range",
								Priority:             100,
								Direction:            "Inbound",
								Access:               "Allow",
								Protocol:             "Tcp",
								DestinationPortRange: "22",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with source_address_prefixes (plural)", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                  "allow-vpn-ssh",
								Priority:              100,
								Direction:             "Inbound",
								Access:                "Allow",
								Protocol:              "Tcp",
								DestinationPortRange:  "22",
								SourceAddressPrefixes: []string{"10.0.0.0/8", "172.16.0.0/12"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with destination_address_prefixes (plural)", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                       "allow-app-to-db",
								Priority:                   100,
								Direction:                  "Outbound",
								Access:                     "Allow",
								Protocol:                   "Tcp",
								DestinationPortRange:       "5432",
								DestinationAddressPrefixes: []string{"10.0.2.0/24", "10.0.3.0/24"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with protocol wildcard", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "deny-all",
								Priority:             4096,
								Direction:            "Inbound",
								Access:               "Deny",
								Protocol:             "*",
								DestinationPortRange: "*",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with Udp protocol", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "allow-dns",
								Priority:             100,
								Direction:            "Inbound",
								Access:               "Allow",
								Protocol:             "Udp",
								DestinationPortRange: "53",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with Icmp protocol", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "allow-ping",
								Priority:             100,
								Direction:            "Inbound",
								Access:               "Allow",
								Protocol:             "Icmp",
								DestinationPortRange: "*",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with Outbound direction", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "deny-internet-outbound",
								Priority:             100,
								Direction:            "Outbound",
								Access:               "Deny",
								Protocol:             "*",
								DestinationPortRange: "*",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with priority at minimum boundary (100)", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "high-priority-rule",
								Priority:             100,
								Direction:            "Inbound",
								Access:               "Allow",
								Protocol:             "Tcp",
								DestinationPortRange: "443",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with priority at maximum boundary (4096)", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "low-priority-deny",
								Priority:             4096,
								Direction:            "Inbound",
								Access:               "Deny",
								Protocol:             "*",
								DestinationPortRange: "*",
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
		ginkgo.Context("azure_network_security_group", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						Name:   "test-nsg",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds maximum length", func() {
				tooLongName := ""
				for len(tooLongName) < 81 {
					tooLongName += "a"
				}
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: tooLongName,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "wrong.version/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule direction is invalid", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "bad-direction",
								Priority:             100,
								Direction:            "INBOUND",
								Access:               "Allow",
								Protocol:             "Tcp",
								DestinationPortRange: "443",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule access is invalid", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "bad-access",
								Priority:             100,
								Direction:            "Inbound",
								Access:               "ALLOW",
								Protocol:             "Tcp",
								DestinationPortRange: "443",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule protocol is invalid", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "bad-protocol",
								Priority:             100,
								Direction:            "Inbound",
								Access:               "Allow",
								Protocol:             "HTTP",
								DestinationPortRange: "443",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule priority is below 100", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "too-high-priority",
								Priority:             99,
								Direction:            "Inbound",
								Access:               "Allow",
								Protocol:             "Tcp",
								DestinationPortRange: "443",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule priority exceeds 4096", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "too-low-priority",
								Priority:             4097,
								Direction:            "Inbound",
								Access:               "Allow",
								Protocol:             "Tcp",
								DestinationPortRange: "443",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule name is missing", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Priority:             100,
								Direction:            "Inbound",
								Access:               "Allow",
								Protocol:             "Tcp",
								DestinationPortRange: "443",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule destination_port_range is missing", func() {
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:      "missing-dest-port",
								Priority:  100,
								Direction: "Inbound",
								Access:    "Allow",
								Protocol:  "Tcp",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule description exceeds 140 characters", func() {
				longDesc := ""
				for len(longDesc) < 141 {
					longDesc += "a"
				}
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 "long-desc",
								Description:          longDesc,
								Priority:             100,
								Direction:            "Inbound",
								Access:               "Allow",
								Protocol:             "Tcp",
								DestinationPortRange: "443",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule name exceeds 80 characters", func() {
				longRuleName := ""
				for len(longRuleName) < 81 {
					longRuleName += "a"
				}
				input := &AzureNetworkSecurityGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNetworkSecurityGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nsg",
					},
					Spec: &AzureNetworkSecurityGroupSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-nsg",
						SecurityRules: []*AzureSecurityRule{
							{
								Name:                 longRuleName,
								Priority:             100,
								Direction:            "Inbound",
								Access:               "Allow",
								Protocol:             "Tcp",
								DestinationPortRange: "443",
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
