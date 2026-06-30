package azurepublicipv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAzurePublicIpSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePublicIpSpec Validation Tests")
}

var _ = ginkgo.Describe("AzurePublicIpSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_public_ip", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "my-public-ip",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full metadata", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-pip",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzurePublicIpSpec{
						Region: "westeurope",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-network-rg",
							},
						},
						Name: "prod-gateway-pip",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with domain_name_label", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name:            "my-public-ip",
						DomainNameLabel: "myapp-gateway",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with zone-redundant zones", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name:  "my-public-ip",
						Zones: []string{"1", "2", "3"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with single zone", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name:  "my-public-ip",
						Zones: []string{"1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with idle_timeout_in_minutes at minimum", func() {
				minTimeout := int32(4)
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name:                 "my-public-ip",
						IdleTimeoutInMinutes: &minTimeout,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with idle_timeout_in_minutes at maximum", func() {
				maxTimeout := int32(30)
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name:                 "my-public-ip",
						IdleTimeoutInMinutes: &maxTimeout,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all fields populated", func() {
				timeout := int32(10)
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-pip",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-network-rg",
							},
						},
						Name:                 "prod-gateway-pip",
						DomainNameLabel:      "prod-gateway",
						Zones:                []string{"1", "2", "3"},
						IdleTimeoutInMinutes: &timeout,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with maximum length name", func() {
				// Azure allows up to 80 characters for Public IP names
				longName := ""
				for len(longName) < 80 {
					longName += "a"
				}
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "long-name-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: longName,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with domain_name_label at minimum length", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name:            "my-public-ip",
						DomainNameLabel: "ab1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_public_ip", func() {

			ginkgo.It("should return a validation error when name is missing", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
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

			ginkgo.It("should return a validation error when name is empty string", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is missing", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "my-public-ip",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty string", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "my-public-ip",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						Name:   "my-public-ip",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds maximum length", func() {
				// Azure limit is 80 characters
				tooLongName := ""
				for len(tooLongName) < 81 {
					tooLongName += "a"
				}
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
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

			ginkgo.It("should return a validation error when domain_name_label starts with a digit", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name:            "my-public-ip",
						DomainNameLabel: "1invalid-label",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when domain_name_label contains uppercase", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name:            "my-public-ip",
						DomainNameLabel: "MyInvalidLabel",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when domain_name_label ends with a hyphen", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name:            "my-public-ip",
						DomainNameLabel: "invalid-label-",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when idle_timeout_in_minutes is below minimum", func() {
				belowMin := int32(3)
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name:                 "my-public-ip",
						IdleTimeoutInMinutes: &belowMin,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when idle_timeout_in_minutes is above maximum", func() {
				aboveMax := int32(31)
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name:                 "my-public-ip",
						IdleTimeoutInMinutes: &aboveMax,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := &AzurePublicIp{
					ApiVersion: "wrong.version/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "my-public-ip",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "my-public-ip",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Spec: &AzurePublicIpSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "my-public-ip",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzurePublicIp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePublicIp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pip",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
