package module

import (
	azurevirtualmachinescalesetv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualmachinescaleset/v1alpha1"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Map the spec enums to ARM's exact API values. Enum helpers return the
// ARM string for set values and "" for unspecified, so callers can skip
// unset optionals (an unspecified spec and Azure's default deploy
// identically on both engines).

func cachingToArm(c azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDiskCaching) string {
	switch c {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDiskCaching_NONE:
		return "None"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDiskCaching_READ_ONLY:
		return "ReadOnly"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDiskCaching_READ_WRITE:
		return "ReadWrite"
	default:
		return ""
	}
}

func osDiskStorageToArm(t azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetOsDiskStorageAccountType) string {
	switch t {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetOsDiskStorageAccountType_STANDARD_LRS:
		return "Standard_LRS"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetOsDiskStorageAccountType_STANDARD_SSD_LRS:
		return "StandardSSD_LRS"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetOsDiskStorageAccountType_PREMIUM_LRS:
		return "Premium_LRS"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetOsDiskStorageAccountType_STANDARD_SSD_ZRS:
		return "StandardSSD_ZRS"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetOsDiskStorageAccountType_PREMIUM_ZRS:
		return "Premium_ZRS"
	default:
		return ""
	}
}

func dataDiskStorageToArm(t azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDataDiskStorageAccountType) string {
	switch t {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDataDiskStorageAccountType_DATA_STANDARD_LRS:
		return "Standard_LRS"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDataDiskStorageAccountType_DATA_STANDARD_SSD_LRS:
		return "StandardSSD_LRS"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDataDiskStorageAccountType_DATA_PREMIUM_LRS:
		return "Premium_LRS"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDataDiskStorageAccountType_DATA_PREMIUM_ZRS:
		return "Premium_ZRS"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDataDiskStorageAccountType_ULTRA_SSD_LRS:
		return "UltraSSD_LRS"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDataDiskStorageAccountType_PREMIUM_V2_LRS:
		return "PremiumV2_LRS"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDataDiskStorageAccountType_DATA_STANDARD_SSD_ZRS:
		return "StandardSSD_ZRS"
	default:
		return ""
	}
}

func createOptionToArm(o azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDataDiskCreateOption) string {
	if o == azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDataDiskCreateOption_FROM_IMAGE {
		return "FromImage"
	}
	return "Empty"
}

func diffDiskPlacementToArm(p azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDiffDiskPlacement) string {
	switch p {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDiffDiskPlacement_CACHE_DISK:
		return "CacheDisk"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDiffDiskPlacement_RESOURCE_DISK:
		return "ResourceDisk"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetDiffDiskPlacement_NVME_DISK:
		return "NvmeDisk"
	default:
		return ""
	}
}

func securityEncryptionToArm(t azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetSecurityEncryptionType) string {
	switch t {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetSecurityEncryptionType_VM_GUEST_STATE_ONLY:
		return "VMGuestStateOnly"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetSecurityEncryptionType_DISK_WITH_VM_GUEST_STATE:
		return "DiskWithVMGuestState"
	default:
		return ""
	}
}

func ipVersionToArm(v azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetIpVersion) string {
	if v == azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetIpVersion_IPV6 {
		return "IPv6"
	}
	return "IPv4"
}

func upgradeModeToArm(m azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetUpgradeMode) string {
	switch m {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetUpgradeMode_AUTOMATIC:
		return "Automatic"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetUpgradeMode_ROLLING:
		return "Rolling"
	default:
		return "Manual"
	}
}

func evictionPolicyToArm(p azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetEvictionPolicy) string {
	if p == azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetEvictionPolicy_DEALLOCATE {
		return "Deallocate"
	}
	return "Delete"
}

func repairActionToArm(a azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetRepairAction) string {
	switch a {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetRepairAction_REPLACE:
		return "Replace"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetRepairAction_RESTART:
		return "Restart"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetRepairAction_REIMAGE:
		return "Reimage"
	default:
		return ""
	}
}

func scaleInRuleToArm(r azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetScaleInRule) string {
	switch r {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetScaleInRule_NEWEST_VM:
		return "NewestVM"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetScaleInRule_OLDEST_VM:
		return "OldestVM"
	default:
		return "Default"
	}
}

func identityTypeToArm(t azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetIdentityType) string {
	switch t {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetIdentityType_SYSTEM_ASSIGNED:
		return "SystemAssigned"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetIdentityType_USER_ASSIGNED:
		return "UserAssigned"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetIdentityType_SYSTEM_AND_USER_ASSIGNED:
		return "SystemAssigned, UserAssigned"
	default:
		return ""
	}
}

func linuxPatchModeToArm(m azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetLinuxPatchMode) string {
	switch m {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetLinuxPatchMode_LINUX_IMAGE_DEFAULT:
		return "ImageDefault"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetLinuxPatchMode_LINUX_AUTOMATIC_BY_PLATFORM:
		return "AutomaticByPlatform"
	default:
		return ""
	}
}

func windowsPatchModeToArm(m azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetWindowsPatchMode) string {
	switch m {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetWindowsPatchMode_WINDOWS_MANUAL:
		return "Manual"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetWindowsPatchMode_AUTOMATIC_BY_OS:
		return "AutomaticByOS"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetWindowsPatchMode_WINDOWS_AUTOMATIC_BY_PLATFORM:
		return "AutomaticByPlatform"
	default:
		return ""
	}
}

func assessmentModeToArm(m azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetPatchAssessmentMode) string {
	switch m {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetPatchAssessmentMode_ASSESSMENT_IMAGE_DEFAULT:
		return "ImageDefault"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetPatchAssessmentMode_ASSESSMENT_AUTOMATIC_BY_PLATFORM:
		return "AutomaticByPlatform"
	default:
		return ""
	}
}

func windowsLicenseToArm(t azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetWindowsLicenseType) string {
	switch t {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetWindowsLicenseType_WINDOWS_LICENSE_NONE:
		return "None"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetWindowsLicenseType_WINDOWS_CLIENT:
		return "Windows_Client"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetWindowsLicenseType_WINDOWS_SERVER:
		return "Windows_Server"
	default:
		return ""
	}
}

func winrmProtocolToArm(p azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetWinrmProtocol) string {
	if p == azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetWinrmProtocol_HTTPS {
		return "Https"
	}
	return "Http"
}

func unattendSettingToArm(s azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetUnattendSetting) string {
	if s == azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetUnattendSetting_FIRST_LOGON_COMMANDS {
		return "FirstLogonCommands"
	}
	return "AutoLogon"
}

func allocationStrategyToArm(s azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetAllocationStrategy) string {
	switch s {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetAllocationStrategy_LOWEST_PRICE:
		return "LowestPrice"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetAllocationStrategy_CAPACITY_OPTIMIZED:
		return "CapacityOptimized"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetAllocationStrategy_PRIORITIZED:
		return "Prioritized"
	default:
		return ""
	}
}

func auxiliaryModeToArm(m azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetAuxiliaryMode) string {
	switch m {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetAuxiliaryMode_ACCELERATED_CONNECTIONS:
		return "AcceleratedConnections"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetAuxiliaryMode_FLOATING:
		return "Floating"
	default:
		return ""
	}
}

func auxiliarySkuToArm(s azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetAuxiliarySku) string {
	switch s {
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetAuxiliarySku_A1:
		return "A1"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetAuxiliarySku_A2:
		return "A2"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetAuxiliarySku_A4:
		return "A4"
	case azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetAuxiliarySku_A8:
		return "A8"
	default:
		return ""
	}
}

// optionalBool resolves an optional bool field to its value or, when
// unset, the proto-declared default. IaC-input paths that bypass the
// manifest loader deliver unset optionals as nil, and a bare getter's
// false would silently diverge from the Terraform module's
// optional(bool, true) encodings on the true-default fields
// (provision_vm_agent, extension_operations_enabled, overprovision,
// disable_password_authentication, automatic_updates_enabled,
// auto_upgrade_minor_version_enabled).
func optionalBool(v *bool, def bool) bool {
	if v != nil {
		return *v
	}
	return def
}

// refValues resolves a list of StringValueOrRef to their literal values,
// skipping empties (references are resolved by the platform before the
// module runs).
func refValues(refs []*foreignkeyv1.StringValueOrRef) pulumi.StringArray {
	out := pulumi.StringArray{}
	for _, r := range refs {
		if r.GetValue() != "" {
			out = append(out, pulumi.String(r.GetValue()))
		}
	}
	return out
}
