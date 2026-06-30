package azurevirtualmachinev1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestAzureVirtualMachineSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVirtualMachineSpec Validation Suite")
}

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}

var _ = ginkgo.Describe("AzureVirtualMachineSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_virtual_machine with minimal configuration using SSH", func() {

			ginkgo.It("should not return a validation error for minimal valid fields with SSH key", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("my-rg"),
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						Image: &AzureVirtualMachineImage{
							Publisher: "Canonical",
							Offer:     "0001-com-ubuntu-server-jammy",
							Sku:       "22_04-lts-gen2",
						},
						SshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E test@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with password authentication", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("my-rg"),
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						Image: &AzureVirtualMachineImage{
							Publisher: "MicrosoftWindowsServer",
							Offer:     "WindowsServer",
							Sku:       "2022-datacenter-g2",
						},
						AdminPassword: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "SuperSecurePassword123!",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with custom image ID", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("my-rg"),
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						Image: &AzureVirtualMachineImage{
							CustomImageId: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Compute/images/myimage",
						},
						SshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E test@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with availability zone", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region:           "eastus",
						ResourceGroup:    stringRef("my-rg"),
						AvailabilityZone: "1",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						Image: &AzureVirtualMachineImage{
							Publisher: "Canonical",
							Offer:     "0001-com-ubuntu-server-jammy",
							Sku:       "22_04-lts-gen2",
						},
						SshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E test@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with spot instance configuration", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("my-rg"),
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						Image: &AzureVirtualMachineImage{
							Publisher: "Canonical",
							Offer:     "0001-com-ubuntu-server-jammy",
							Sku:       "22_04-lts-gen2",
						},
						SshPublicKey:   "ssh-rsa AAAAB3NzaC1yc2E test@host",
						IsSpotInstance: true,
						SpotMaxPrice:   0.5,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with data disks", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("my-rg"),
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						Image: &AzureVirtualMachineImage{
							Publisher: "Canonical",
							Offer:     "0001-com-ubuntu-server-jammy",
							Sku:       "22_04-lts-gen2",
						},
						SshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E test@host",
						DataDisks: []*AzureVirtualMachineDataDisk{
							{
								Name:   "data-disk-1",
								SizeGb: 256,
								Lun:    0,
							},
							{
								Name:   "data-disk-2",
								SizeGb: 512,
								Lun:    1,
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
		ginkgo.Context("azure_virtual_machine", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						ResourceGroup: stringRef("my-rg"),
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						Image: &AzureVirtualMachineImage{
							Publisher: "Canonical",
							Offer:     "0001-com-ubuntu-server-jammy",
							Sku:       "22_04-lts-gen2",
						},
						SshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E test@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region: "eastus",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						Image: &AzureVirtualMachineImage{
							Publisher: "Canonical",
							Offer:     "0001-com-ubuntu-server-jammy",
							Sku:       "22_04-lts-gen2",
						},
						SshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E test@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet_id is missing", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("my-rg"),
						Image: &AzureVirtualMachineImage{
							Publisher: "Canonical",
							Offer:     "0001-com-ubuntu-server-jammy",
							Sku:       "22_04-lts-gen2",
						},
						SshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E test@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when image is missing", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("my-rg"),
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						SshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E test@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when neither ssh_public_key nor admin_password is provided", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("my-rg"),
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						Image: &AzureVirtualMachineImage{
							Publisher: "Canonical",
							Offer:     "0001-com-ubuntu-server-jammy",
							Sku:       "22_04-lts-gen2",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when availability_zone is invalid", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region:           "eastus",
						ResourceGroup:    stringRef("my-rg"),
						AvailabilityZone: "4",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						Image: &AzureVirtualMachineImage{
							Publisher: "Canonical",
							Offer:     "0001-com-ubuntu-server-jammy",
							Sku:       "22_04-lts-gen2",
						},
						SshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E test@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spot_max_price is set but is_spot_instance is false", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("my-rg"),
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						Image: &AzureVirtualMachineImage{
							Publisher: "Canonical",
							Offer:     "0001-com-ubuntu-server-jammy",
							Sku:       "22_04-lts-gen2",
						},
						SshPublicKey:   "ssh-rsa AAAAB3NzaC1yc2E test@host",
						IsSpotInstance: false,
						SpotMaxPrice:   0.5,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when image has neither marketplace info nor custom_image_id", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("my-rg"),
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						Image:        &AzureVirtualMachineImage{},
						SshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E test@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when data disk LUN is out of range", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("my-rg"),
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						Image: &AzureVirtualMachineImage{
							Publisher: "Canonical",
							Offer:     "0001-com-ubuntu-server-jammy",
							Sku:       "22_04-lts-gen2",
						},
						SshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E test@host",
						DataDisks: []*AzureVirtualMachineDataDisk{
							{
								Name:   "data-disk-1",
								SizeGb: 256,
								Lun:    100,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when data disk name is empty", func() {
				input := &AzureVirtualMachine{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVirtualMachine",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vm",
					},
					Spec: &AzureVirtualMachineSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("my-rg"),
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						Image: &AzureVirtualMachineImage{
							Publisher: "Canonical",
							Offer:     "0001-com-ubuntu-server-jammy",
							Sku:       "22_04-lts-gen2",
						},
						SshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E test@host",
						DataDisks: []*AzureVirtualMachineDataDisk{
							{
								Name:   "",
								SizeGb: 256,
								Lun:    0,
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
