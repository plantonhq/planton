package azurevirtualmachinev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}

const (
	testNicId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Network/networkInterfaces/app-nic"
	testDiskId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Compute/disks/data"
	testDesId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/sec-rg/providers/Microsoft.Compute/diskEncryptionSets/cmk"
	testSshKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKp1QgHKux0e/js6p7UBR4jYOtb5aeedkl+0cNr5RB6Q planton-oss-e2e"
)

// validLinuxSpec returns a minimal valid SSH-key Linux VM the failure
// cases mutate one field at a time.
func validLinuxSpec() *AzureVirtualMachineSpec {
	return &AzureVirtualMachineSpec{
		Region:              "eastus",
		ResourceGroup:       stringRef("app-rg"),
		Name:                "app-vm",
		Size:                "Standard_D2s_v3",
		NetworkInterfaceIds: []*foreignkeyv1.StringValueOrRef{stringRef(testNicId)},
		OsProfile: &AzureVirtualMachineOsProfile{
			Linux: &AzureVirtualMachineLinuxProfile{
				AdminUsername: "azureuser",
				SshPublicKeys: []*AzureVirtualMachineSshPublicKey{{PublicKey: testSshKey}},
			},
		},
		OsDisk: &AzureVirtualMachineOsDisk{
			Caching:            AzureVirtualMachineDiskCaching_READ_WRITE,
			StorageAccountType: AzureVirtualMachineOsDiskStorageAccountType_PREMIUM_LRS,
		},
		SourceImageReference: &AzureVirtualMachineSourceImageReference{
			Publisher: "Canonical",
			Offer:     "ubuntu-24_04-lts",
			Sku:       "server",
			Version:   "latest",
		},
	}
}

// validWindowsSpec returns a minimal valid Windows VM.
func validWindowsSpec() *AzureVirtualMachineSpec {
	spec := validLinuxSpec()
	spec.OsProfile = &AzureVirtualMachineOsProfile{
		Windows: &AzureVirtualMachineWindowsProfile{
			AdminUsername: "azureadmin",
			AdminPassword: stringRef("C0mplex!Passw0rd"),
		},
	}
	spec.SourceImageReference = &AzureVirtualMachineSourceImageReference{
		Publisher: "MicrosoftWindowsServer",
		Offer:     "WindowsServer",
		Sku:       "2022-datacenter-g2",
		Version:   "latest",
	}
	return spec
}

func validInput(spec *AzureVirtualMachineSpec) *AzureVirtualMachine {
	return &AzureVirtualMachine{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureVirtualMachine",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-virtual-machine",
		},
		Spec: spec,
	}
}

func TestAzureVirtualMachineSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVirtualMachineSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AzureVirtualMachineSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal SSH-key Linux VM", func() {
			err := protovalidate.Validate(validInput(validLinuxSpec()))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept boot diagnostics declared and switched off", func() {
			spec := validLinuxSpec()
			spec.BootDiagnostics = &AzureVirtualMachineBootDiagnostics{Enabled: proto.Bool(false)}
			gomega.Expect(protovalidate.Validate(validInput(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a Linux VM with password auth explicitly enabled", func() {
			spec := validLinuxSpec()
			spec.OsProfile.Linux.SshPublicKeys = nil
			spec.OsProfile.Linux.DisablePasswordAuthentication = proto.Bool(false)
			spec.OsProfile.Linux.AdminPassword = stringRef("C0mplex!Passw0rd")
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a minimal Windows VM", func() {
			err := protovalidate.Validate(validInput(validWindowsSpec()))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a custom image by id", func() {
			spec := validLinuxSpec()
			spec.SourceImageReference = nil
			spec.SourceImageId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/galleries/g/images/i/versions/1.0.0"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept booting from an existing OS disk with no auth fields", func() {
			spec := validLinuxSpec()
			spec.SourceImageReference = nil
			spec.OsManagedDiskId = stringRef(testDiskId)
			spec.OsProfile.Linux = &AzureVirtualMachineLinuxProfile{}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept data-disk attachments, identity, and availability", func() {
			spec := validLinuxSpec()
			spec.DataDiskAttachments = []*AzureVirtualMachineDataDiskAttachment{
				{ManagedDiskId: stringRef(testDiskId), Lun: proto.Int32(0), Caching: AzureVirtualMachineDiskCaching_READ_ONLY},
			}
			spec.Identity = &AzureVirtualMachineIdentity{Type: AzureVirtualMachineIdentityType_SYSTEM_ASSIGNED}
			spec.Availability = &AzureVirtualMachineAvailability{Zone: "1"}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a spot VM with an eviction policy", func() {
			spec := validLinuxSpec()
			spec.Spot = &AzureVirtualMachineSpot{EvictionPolicy: AzureVirtualMachineEvictionPolicy_DEALLOCATE}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept trusted launch with platform patching and reboot control", func() {
			spec := validLinuxSpec()
			spec.Security = &AzureVirtualMachineSecurity{SecureBootEnabled: true, VtpmEnabled: true}
			spec.OsProfile.Linux.PatchMode = AzureVirtualMachineLinuxPatchMode_LINUX_AUTOMATIC_BY_PLATFORM
			spec.Patching = &AzureVirtualMachinePatching{RebootSetting: AzureVirtualMachineRebootSetting_IF_REQUIRED}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept confidential-VM guest-state encryption with vTPM and secure boot", func() {
			spec := validLinuxSpec()
			spec.Security = &AzureVirtualMachineSecurity{SecureBootEnabled: true, VtpmEnabled: true}
			spec.OsDisk.SecurityEncryptionType = AzureVirtualMachineSecurityEncryptionType_DISK_WITH_VM_GUEST_STATE
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an ephemeral OS disk", func() {
			spec := validLinuxSpec()
			spec.OsDisk.DiffDiskSettings = &AzureVirtualMachineDiffDiskSettings{
				Placement: AzureVirtualMachineDiffDiskPlacement_CACHE_DISK,
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept Windows hotpatching under platform patching", func() {
			spec := validWindowsSpec()
			spec.OsProfile.Windows.PatchMode = AzureVirtualMachineWindowsPatchMode_WINDOWS_AUTOMATIC_BY_PLATFORM
			spec.OsProfile.Windows.HotpatchingEnabled = true
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an HTTPS WinRM listener with a certificate", func() {
			spec := validWindowsSpec()
			spec.OsProfile.Windows.WinrmListeners = []*AzureVirtualMachineWinrmListener{
				{Protocol: AzureVirtualMachineWinrmProtocol_HTTPS, CertificateUrl: "https://vault.vault.azure.net/secrets/winrm/1"},
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing size", func() {
			spec := validLinuxSpec()
			spec.Size = ""
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a VM without network interfaces", func() {
			spec := validLinuxSpec()
			spec.NetworkInterfaceIds = nil
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a VM without an OS profile", func() {
			spec := validLinuxSpec()
			spec.OsProfile = nil
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject both OS profiles at once", func() {
			spec := validLinuxSpec()
			spec.OsProfile.Windows = &AzureVirtualMachineWindowsProfile{
				AdminUsername: "azureadmin",
				AdminPassword: stringRef("C0mplex!Passw0rd"),
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a VM without any image source", func() {
			spec := validLinuxSpec()
			spec.SourceImageReference = nil
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject two image sources at once", func() {
			spec := validLinuxSpec()
			spec.SourceImageId = "/subscriptions/s/.../versions/1.0.0"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a Linux VM with no credentials", func() {
			spec := validLinuxSpec()
			spec.OsProfile.Linux.SshPublicKeys = nil
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a Linux VM with password auth enabled but no password", func() {
			spec := validLinuxSpec()
			spec.OsProfile.Linux.SshPublicKeys = nil
			spec.OsProfile.Linux.DisablePasswordAuthentication = proto.Bool(false)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a Windows VM without a password", func() {
			spec := validWindowsSpec()
			spec.OsProfile.Windows.AdminPassword = nil
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject auth fields when booting from an existing OS disk", func() {
			spec := validLinuxSpec()
			spec.SourceImageReference = nil
			spec.OsManagedDiskId = stringRef(testDiskId)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a VM without an os_disk", func() {
			spec := validLinuxSpec()
			spec.OsDisk = nil
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an os_disk without explicit caching", func() {
			spec := validLinuxSpec()
			spec.OsDisk.Caching = AzureVirtualMachineDiskCaching_azure_virtual_machine_disk_caching_unspecified
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an incomplete source_image_reference", func() {
			spec := validLinuxSpec()
			spec.SourceImageReference.Version = ""
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a data-disk attachment with an out-of-range LUN", func() {
			spec := validLinuxSpec()
			spec.DataDiskAttachments = []*AzureVirtualMachineDataDiskAttachment{
				{ManagedDiskId: stringRef(testDiskId), Lun: proto.Int32(64), Caching: AzureVirtualMachineDiskCaching_NONE},
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a spot VM without an eviction policy", func() {
			spec := validLinuxSpec()
			spec.Spot = &AzureVirtualMachineSpot{}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a zone combined with an availability set", func() {
			spec := validLinuxSpec()
			spec.Availability = &AzureVirtualMachineAvailability{
				Zone:              "1",
				AvailabilitySetId: stringRef("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/availabilitySets/legacy"),
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a platform fault domain without a scale set", func() {
			spec := validLinuxSpec()
			spec.Availability = &AzureVirtualMachineAvailability{PlatformFaultDomain: proto.Int32(0)}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a USER_ASSIGNED identity without identity_ids", func() {
			spec := validLinuxSpec()
			spec.Identity = &AzureVirtualMachineIdentity{Type: AzureVirtualMachineIdentityType_USER_ASSIGNED}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a reboot setting without platform patching", func() {
			spec := validLinuxSpec()
			spec.Patching = &AzureVirtualMachinePatching{RebootSetting: AzureVirtualMachineRebootSetting_ALWAYS}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject guest-state encryption without vTPM", func() {
			spec := validLinuxSpec()
			spec.OsDisk.SecurityEncryptionType = AzureVirtualMachineSecurityEncryptionType_VM_GUEST_STATE_ONLY
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject full disk-with-guest-state encryption without secure boot", func() {
			spec := validLinuxSpec()
			spec.Security = &AzureVirtualMachineSecurity{VtpmEnabled: true}
			spec.OsDisk.SecurityEncryptionType = AzureVirtualMachineSecurityEncryptionType_DISK_WITH_VM_GUEST_STATE
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject both OS-disk encryption sets together", func() {
			spec := validLinuxSpec()
			spec.Security = &AzureVirtualMachineSecurity{VtpmEnabled: true}
			spec.OsDisk.SecurityEncryptionType = AzureVirtualMachineSecurityEncryptionType_VM_GUEST_STATE_ONLY
			spec.OsDisk.DiskEncryptionSetId = stringRef(testDesId)
			spec.OsDisk.SecureVmDiskEncryptionSetId = stringRef(testDesId)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject Windows hotpatching without platform patching", func() {
			spec := validWindowsSpec()
			spec.OsProfile.Windows.HotpatchingEnabled = true
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an HTTPS WinRM listener without a certificate", func() {
			spec := validWindowsSpec()
			spec.OsProfile.Windows.WinrmListeners = []*AzureVirtualMachineWinrmListener{
				{Protocol: AzureVirtualMachineWinrmProtocol_HTTPS},
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a Key Vault secret without certificates", func() {
			spec := validLinuxSpec()
			spec.Secrets = []*AzureVirtualMachineSecret{
				{KeyVaultId: stringRef("/subscriptions/s/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/v")},
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
