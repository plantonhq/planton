package azurebastionhostv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureBastionHostSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureBastionHostSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

func boolPtr(v bool) *bool { return &v }

const (
	testSubnetId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/virtualNetworks/platform-vnet/subnets/AzureBastionSubnet"
	testPipId    = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/publicIPAddresses/bastion-pip"
	testVnetId   = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/virtualNetworks/platform-vnet"
)

// dedicatedIpConfiguration returns the Basic/Standard-shaped binding:
// the AzureBastionSubnet plus a public IP.
func dedicatedIpConfiguration() *AzureBastionHostIpConfiguration {
	return &AzureBastionHostIpConfiguration{
		Name:              "bastion-ip-config",
		SubnetId:          literal(testSubnetId),
		PublicIpAddressId: literal(testPipId),
	}
}

// validResource returns a minimal valid host (the default BASIC shape)
// that individual cases mutate into the shape under test.
func validResource() *AzureBastionHost {
	return &AzureBastionHost{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureBastionHost",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-bastion",
		},
		Spec: &AzureBastionHostSpec{
			Region:          "eastus",
			ResourceGroup:   literal("platform-rg"),
			Name:            "platform-bastion",
			IpConfiguration: dedicatedIpConfiguration(),
		},
	}
}

var _ = ginkgo.Describe("AzureBastionHostSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_bastion_host", func() {

			ginkgo.It("should not return a validation error for the minimal (default BASIC) host", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit BASIC host with its dedicated binding", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_BASIC
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit scale_units of 2 on BASIC (the fixed capacity)", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_BASIC
				input.Spec.ScaleUnits = int32Ptr(2)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a STANDARD host with every feature knob and 50 scale units", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_STANDARD
				input.Spec.ScaleUnits = int32Ptr(50)
				input.Spec.FileCopyEnabled = true
				input.Spec.IpConnectEnabled = true
				input.Spec.KerberosEnabled = true
				input.Spec.ShareableLinkEnabled = true
				input.Spec.TunnelingEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a PREMIUM host with session recording", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_PREMIUM
				input.Spec.SessionRecordingEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a private-only PREMIUM host (no public IP)", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_PREMIUM
				input.Spec.IpConfiguration.PublicIpAddressId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a DEVELOPER host attached to a virtual network", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_DEVELOPER
				input.Spec.IpConfiguration = nil
				input.Spec.VirtualNetworkId = literal(testVnetId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a zone-redundant host", func() {
				input := validResource()
				input.Spec.Zones = []string{"1", "2", "3"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept copy_paste explicitly disabled", func() {
				input := validResource()
				input.Spec.CopyPasteEnabled = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an 80-character name", func() {
				input := validResource()
				input.Spec.Name = "b" + strings.Repeat("a", 78) + "9"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_bastion_host", func() {

			ginkgo.It("should reject a missing name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name starting with a period", func() {
				input := validResource()
				input.Spec.Name = ".bastion"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name ending with a hyphen", func() {
				input := validResource()
				input.Spec.Name = "bastion-"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name over 80 characters", func() {
				input := validResource()
				input.Spec.Name = "b" + strings.Repeat("a", 79) + "9"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject the default (BASIC) shape without ip_configuration", func() {
				input := validResource()
				input.Spec.IpConfiguration = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an explicit BASIC host without a public IP", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_BASIC
				input.Spec.IpConfiguration.PublicIpAddressId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a STANDARD host without a public IP", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_STANDARD
				input.Spec.IpConfiguration.PublicIpAddressId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a DEVELOPER host without a virtual network", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_DEVELOPER
				input.Spec.IpConfiguration = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject virtual_network_id on a BASIC host", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_BASIC
				input.Spec.VirtualNetworkId = literal(testVnetId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject non-default scale_units on BASIC", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_BASIC
				input.Spec.ScaleUnits = int32Ptr(5)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject scale_units below 2", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_STANDARD
				input.Spec.ScaleUnits = int32Ptr(1)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject scale_units above 50", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_STANDARD
				input.Spec.ScaleUnits = int32Ptr(51)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject file copy on BASIC", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_BASIC
				input.Spec.FileCopyEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject IP connect on the default (BASIC) shape", func() {
				input := validResource()
				input.Spec.IpConnectEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject Kerberos on a DEVELOPER host", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_DEVELOPER
				input.Spec.IpConfiguration = nil
				input.Spec.VirtualNetworkId = literal(testVnetId)
				input.Spec.KerberosEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject shareable links on BASIC", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_BASIC
				input.Spec.ShareableLinkEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject tunneling on BASIC", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_BASIC
				input.Spec.TunnelingEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject session recording on STANDARD", func() {
				input := validResource()
				input.Spec.Sku = AzureBastionHostSku_STANDARD
				input.Spec.SessionRecordingEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an ip_configuration without a subnet", func() {
				input := validResource()
				input.Spec.IpConfiguration.SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an ip_configuration without a name", func() {
				input := validResource()
				input.Spec.IpConfiguration.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
