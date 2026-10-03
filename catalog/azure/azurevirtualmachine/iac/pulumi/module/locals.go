package module

import (
	"strings"

	azurevirtualmachinev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualmachine/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureVirtualMachine *azurevirtualmachinev1alpha1.AzureVirtualMachine

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// NetworkInterfaceIds are the resolved ARM IDs of the attached NICs
	// (repeated StringValueOrRef in the spec; the first is primary).
	NetworkInterfaceIds []string

	// IsLinux is the explicit OS discriminator -- which of ARM's two VM
	// surfaces this spec deploys through.
	IsLinux bool

	// Enum-to-ARM string mappings, empty when unspecified so both engines
	// send nothing and Azure's defaults apply.
	OsDiskCaching          string
	OsDiskStorageType      string
	DiffDiskPlacement      string
	SecurityEncryptionType string
	IdentityType           string
	IdentityIds            []string
	EvictionPolicy         string
	LinuxPatchMode         string
	WindowsPatchMode       string
	PatchAssessmentMode    string
	RebootSetting          string
	LinuxLicenseType       string
	WindowsLicenseType     string
	DiskControllerType     string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

// cachingToArm maps the shared disk-caching enum to ARM's strings.
func cachingToArm(caching azurevirtualmachinev1alpha1.AzureVirtualMachineDiskCaching) string {
	switch caching {
	case azurevirtualmachinev1alpha1.AzureVirtualMachineDiskCaching_NONE:
		return "None"
	case azurevirtualmachinev1alpha1.AzureVirtualMachineDiskCaching_READ_ONLY:
		return "ReadOnly"
	case azurevirtualmachinev1alpha1.AzureVirtualMachineDiskCaching_READ_WRITE:
		return "ReadWrite"
	}
	return ""
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurevirtualmachinev1alpha1.AzureVirtualMachineIacInput) *Locals {
	locals := &Locals{}

	locals.AzureVirtualMachine = iacInput.Target
	target := iacInput.Target
	spec := target.Spec

	locals.ResourceGroupName = spec.ResourceGroup.GetValue()

	for _, networkInterfaceId := range spec.NetworkInterfaceIds {
		locals.NetworkInterfaceIds = append(locals.NetworkInterfaceIds, networkInterfaceId.GetValue())
	}

	locals.IsLinux = spec.OsProfile.GetLinux() != nil

	locals.OsDiskCaching = cachingToArm(spec.OsDisk.Caching)

	switch spec.OsDisk.StorageAccountType {
	case azurevirtualmachinev1alpha1.AzureVirtualMachineOsDiskStorageAccountType_STANDARD_LRS:
		locals.OsDiskStorageType = "Standard_LRS"
	case azurevirtualmachinev1alpha1.AzureVirtualMachineOsDiskStorageAccountType_STANDARD_SSD_LRS:
		locals.OsDiskStorageType = "StandardSSD_LRS"
	case azurevirtualmachinev1alpha1.AzureVirtualMachineOsDiskStorageAccountType_PREMIUM_LRS:
		locals.OsDiskStorageType = "Premium_LRS"
	case azurevirtualmachinev1alpha1.AzureVirtualMachineOsDiskStorageAccountType_STANDARD_SSD_ZRS:
		locals.OsDiskStorageType = "StandardSSD_ZRS"
	case azurevirtualmachinev1alpha1.AzureVirtualMachineOsDiskStorageAccountType_PREMIUM_ZRS:
		locals.OsDiskStorageType = "Premium_ZRS"
	}

	if spec.OsDisk.DiffDiskSettings != nil {
		switch spec.OsDisk.DiffDiskSettings.Placement {
		case azurevirtualmachinev1alpha1.AzureVirtualMachineDiffDiskPlacement_CACHE_DISK:
			locals.DiffDiskPlacement = "CacheDisk"
		case azurevirtualmachinev1alpha1.AzureVirtualMachineDiffDiskPlacement_RESOURCE_DISK:
			locals.DiffDiskPlacement = "ResourceDisk"
		case azurevirtualmachinev1alpha1.AzureVirtualMachineDiffDiskPlacement_NVME_DISK:
			locals.DiffDiskPlacement = "NvmeDisk"
		}
	}

	switch spec.OsDisk.SecurityEncryptionType {
	case azurevirtualmachinev1alpha1.AzureVirtualMachineSecurityEncryptionType_VM_GUEST_STATE_ONLY:
		locals.SecurityEncryptionType = "VMGuestStateOnly"
	case azurevirtualmachinev1alpha1.AzureVirtualMachineSecurityEncryptionType_DISK_WITH_VM_GUEST_STATE:
		locals.SecurityEncryptionType = "DiskWithVMGuestState"
	}

	if spec.Identity != nil {
		switch spec.Identity.Type {
		case azurevirtualmachinev1alpha1.AzureVirtualMachineIdentityType_SYSTEM_ASSIGNED:
			locals.IdentityType = "SystemAssigned"
		case azurevirtualmachinev1alpha1.AzureVirtualMachineIdentityType_USER_ASSIGNED:
			locals.IdentityType = "UserAssigned"
		case azurevirtualmachinev1alpha1.AzureVirtualMachineIdentityType_SYSTEM_AND_USER_ASSIGNED:
			locals.IdentityType = "SystemAssigned, UserAssigned"
		}
		for _, identityId := range spec.Identity.IdentityIds {
			locals.IdentityIds = append(locals.IdentityIds, identityId.GetValue())
		}
	}

	if spec.Spot != nil {
		switch spec.Spot.EvictionPolicy {
		case azurevirtualmachinev1alpha1.AzureVirtualMachineEvictionPolicy_DEALLOCATE:
			locals.EvictionPolicy = "Deallocate"
		case azurevirtualmachinev1alpha1.AzureVirtualMachineEvictionPolicy_DELETE:
			locals.EvictionPolicy = "Delete"
		}
	}

	if linux := spec.OsProfile.GetLinux(); linux != nil {
		switch linux.PatchMode {
		case azurevirtualmachinev1alpha1.AzureVirtualMachineLinuxPatchMode_LINUX_IMAGE_DEFAULT:
			locals.LinuxPatchMode = "ImageDefault"
		case azurevirtualmachinev1alpha1.AzureVirtualMachineLinuxPatchMode_LINUX_AUTOMATIC_BY_PLATFORM:
			locals.LinuxPatchMode = "AutomaticByPlatform"
		}
		// The Linux license enum's names ARE ARM's literal values.
		if linux.LicenseType != azurevirtualmachinev1alpha1.AzureVirtualMachineLinuxLicenseType_azure_virtual_machine_linux_license_type_unspecified {
			locals.LinuxLicenseType = linux.LicenseType.String()
		}
	}

	if windows := spec.OsProfile.GetWindows(); windows != nil {
		switch windows.PatchMode {
		case azurevirtualmachinev1alpha1.AzureVirtualMachineWindowsPatchMode_MANUAL:
			locals.WindowsPatchMode = "Manual"
		case azurevirtualmachinev1alpha1.AzureVirtualMachineWindowsPatchMode_AUTOMATIC_BY_OS:
			locals.WindowsPatchMode = "AutomaticByOS"
		case azurevirtualmachinev1alpha1.AzureVirtualMachineWindowsPatchMode_WINDOWS_AUTOMATIC_BY_PLATFORM:
			locals.WindowsPatchMode = "AutomaticByPlatform"
		}
		switch windows.LicenseType {
		case azurevirtualmachinev1alpha1.AzureVirtualMachineWindowsLicenseType_WINDOWS_LICENSE_NONE:
			locals.WindowsLicenseType = "None"
		case azurevirtualmachinev1alpha1.AzureVirtualMachineWindowsLicenseType_WINDOWS_CLIENT:
			locals.WindowsLicenseType = "Windows_Client"
		case azurevirtualmachinev1alpha1.AzureVirtualMachineWindowsLicenseType_WINDOWS_SERVER:
			locals.WindowsLicenseType = "Windows_Server"
		}
	}

	if spec.Patching != nil {
		switch spec.Patching.AssessmentMode {
		case azurevirtualmachinev1alpha1.AzureVirtualMachinePatchAssessmentMode_ASSESSMENT_IMAGE_DEFAULT:
			locals.PatchAssessmentMode = "ImageDefault"
		case azurevirtualmachinev1alpha1.AzureVirtualMachinePatchAssessmentMode_ASSESSMENT_AUTOMATIC_BY_PLATFORM:
			locals.PatchAssessmentMode = "AutomaticByPlatform"
		}
		switch spec.Patching.RebootSetting {
		case azurevirtualmachinev1alpha1.AzureVirtualMachineRebootSetting_ALWAYS:
			locals.RebootSetting = "Always"
		case azurevirtualmachinev1alpha1.AzureVirtualMachineRebootSetting_IF_REQUIRED:
			locals.RebootSetting = "IfRequired"
		case azurevirtualmachinev1alpha1.AzureVirtualMachineRebootSetting_NEVER:
			locals.RebootSetting = "Never"
		}
	}

	switch spec.DiskControllerType {
	case azurevirtualmachinev1alpha1.AzureVirtualMachineDiskControllerType_SCSI:
		locals.DiskControllerType = "SCSI"
	case azurevirtualmachinev1alpha1.AzureVirtualMachineDiskControllerType_NVME:
		locals.DiskControllerType = "NVMe"
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureVirtualMachine.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	for k, v := range spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
