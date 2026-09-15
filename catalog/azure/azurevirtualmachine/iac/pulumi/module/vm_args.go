package module

import (
	azurevirtualmachinev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualmachine/v1alpha1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildLinuxArgs assembles the Linux VM's full argument surface from the
// spec. Only explicit spec choices are sent; absent optionals stay nil so
// Azure's defaults apply identically on both engines.
func buildLinuxArgs(locals *Locals) *compute.LinuxVirtualMachineArgs {
	spec := locals.AzureVirtualMachine.Spec
	linux := spec.OsProfile.GetLinux()

	args := &compute.LinuxVirtualMachineArgs{
		Name:                pulumi.String(spec.Name),
		Location:            pulumi.String(spec.Region),
		ResourceGroupName:   pulumi.String(locals.ResourceGroupName),
		Size:                pulumi.String(spec.Size),
		NetworkInterfaceIds: pulumi.ToStringArray(locals.NetworkInterfaceIds),
		Tags:                pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.OsProfile.ComputerName != "" {
		args.ComputerName = pulumi.String(spec.OsProfile.ComputerName)
	}

	// Authentication: SSH-first. Absent entirely when booting from an
	// existing OS disk (the disk already contains its users; spec-level
	// validation enforces the pairing).
	if linux.AdminUsername != "" {
		args.AdminUsername = pulumi.String(linux.AdminUsername)
	}
	if linux.AdminPassword.GetValue() != "" {
		args.AdminPassword = pulumi.String(linux.AdminPassword.GetValue())
	}
	// Presence-guarded true-default optional bool: an absent spec value
	// explicitly falls back to Azure's default (keys only).
	if linux.DisablePasswordAuthentication != nil {
		args.DisablePasswordAuthentication = pulumi.Bool(linux.GetDisablePasswordAuthentication())
	} else {
		args.DisablePasswordAuthentication = pulumi.Bool(true)
	}
	if len(linux.SshPublicKeys) > 0 {
		sshKeys := compute.LinuxVirtualMachineAdminSshKeyArray{}
		for _, sshKey := range linux.SshPublicKeys {
			// An unset key username defaults to the admin account -- the
			// common case.
			username := sshKey.Username
			if username == "" {
				username = linux.AdminUsername
			}
			sshKeys = append(sshKeys, compute.LinuxVirtualMachineAdminSshKeyArgs{
				PublicKey: pulumi.String(sshKey.PublicKey),
				Username:  pulumi.String(username),
			})
		}
		args.AdminSshKeys = sshKeys
	}

	args.OsDisk = buildLinuxOsDisk(locals)

	// Exactly one image source (spec-level validation): marketplace
	// coordinates, a custom/gallery image id, or an existing OS disk.
	if spec.SourceImageReference != nil {
		args.SourceImageReference = compute.LinuxVirtualMachineSourceImageReferenceArgs{
			Publisher: pulumi.String(spec.SourceImageReference.Publisher),
			Offer:     pulumi.String(spec.SourceImageReference.Offer),
			Sku:       pulumi.String(spec.SourceImageReference.Sku),
			Version:   pulumi.String(spec.SourceImageReference.Version),
		}
	}
	if spec.SourceImageId != "" {
		args.SourceImageId = pulumi.String(spec.SourceImageId)
	}
	if spec.OsManagedDiskId.GetValue() != "" {
		args.OsManagedDiskId = pulumi.String(spec.OsManagedDiskId.GetValue())
	}

	if locals.IdentityType != "" {
		args.Identity = compute.LinuxVirtualMachineIdentityArgs{
			Type:        pulumi.String(locals.IdentityType),
			IdentityIds: pulumi.ToStringArray(locals.IdentityIds),
		}
	}

	// Spot: presence of the spec's spot message makes the VM evictable,
	// deeply discounted capacity.
	if spec.Spot != nil {
		args.Priority = pulumi.String("Spot")
		args.EvictionPolicy = pulumi.String(locals.EvictionPolicy)
		if spec.Spot.MaxBidPrice != nil {
			args.MaxBidPrice = pulumi.Float64(spec.Spot.GetMaxBidPrice())
		}
	}

	if availability := spec.Availability; availability != nil {
		if availability.Zone != "" {
			args.Zone = pulumi.String(availability.Zone)
		}
		if availability.AvailabilitySetId.GetValue() != "" {
			args.AvailabilitySetId = pulumi.String(availability.AvailabilitySetId.GetValue())
		}
		if availability.ProximityPlacementGroupId != "" {
			args.ProximityPlacementGroupId = pulumi.String(availability.ProximityPlacementGroupId)
		}
		if availability.CapacityReservationGroupId != "" {
			args.CapacityReservationGroupId = pulumi.String(availability.CapacityReservationGroupId)
		}
		if availability.DedicatedHostId != "" {
			args.DedicatedHostId = pulumi.String(availability.DedicatedHostId)
		}
		if availability.DedicatedHostGroupId != "" {
			args.DedicatedHostGroupId = pulumi.String(availability.DedicatedHostGroupId)
		}
		if availability.VirtualMachineScaleSetId.GetValue() != "" {
			args.VirtualMachineScaleSetId = pulumi.String(availability.VirtualMachineScaleSetId.GetValue())
		}
		if availability.PlatformFaultDomain != nil {
			args.PlatformFaultDomain = pulumi.Int(int(availability.GetPlatformFaultDomain()))
		}
	}

	if security := spec.Security; security != nil {
		args.SecureBootEnabled = pulumi.Bool(security.SecureBootEnabled)
		args.VtpmEnabled = pulumi.Bool(security.VtpmEnabled)
		args.EncryptionAtHostEnabled = pulumi.Bool(security.EncryptionAtHostEnabled)
	}

	// Patch orchestration: the MODE vocabulary is Linux-specific; the
	// shared dials come from spec.patching.
	if locals.LinuxPatchMode != "" {
		args.PatchMode = pulumi.String(locals.LinuxPatchMode)
	}
	if locals.PatchAssessmentMode != "" {
		args.PatchAssessmentMode = pulumi.String(locals.PatchAssessmentMode)
	}
	if locals.RebootSetting != "" {
		args.RebootSetting = pulumi.String(locals.RebootSetting)
	}
	if spec.Patching != nil && spec.Patching.BypassPlatformSafetyChecksOnUserScheduleEnabled {
		args.BypassPlatformSafetyChecksOnUserScheduleEnabled = pulumi.Bool(true)
	}

	if locals.LinuxLicenseType != "" {
		args.LicenseType = pulumi.String(locals.LinuxLicenseType)
	}

	// Presence enables boot diagnostics; an empty URI uses Azure's
	// managed storage (the right default).
	if spec.BootDiagnostics != nil && spec.BootDiagnostics.GetEnabled() {
		bootDiagnostics := compute.LinuxVirtualMachineBootDiagnosticsArgs{}
		if spec.BootDiagnostics.StorageAccountUri != "" {
			bootDiagnostics.StorageAccountUri = pulumi.String(spec.BootDiagnostics.StorageAccountUri)
		}
		args.BootDiagnostics = bootDiagnostics
	}

	if len(spec.GalleryApplications) > 0 {
		galleryApplications := compute.LinuxVirtualMachineGalleryApplicationArray{}
		for _, application := range spec.GalleryApplications {
			applicationArgs := compute.LinuxVirtualMachineGalleryApplicationArgs{
				VersionId:                              pulumi.String(application.VersionId),
				AutomaticUpgradeEnabled:                pulumi.Bool(application.AutomaticUpgradeEnabled),
				TreatFailureAsDeploymentFailureEnabled: pulumi.Bool(application.TreatFailureAsDeploymentFailureEnabled),
			}
			if application.Order != nil {
				applicationArgs.Order = pulumi.Int(int(application.GetOrder()))
			}
			if application.Tag != "" {
				applicationArgs.Tag = pulumi.String(application.Tag)
			}
			if application.ConfigurationBlobUri != "" {
				applicationArgs.ConfigurationBlobUri = pulumi.String(application.ConfigurationBlobUri)
			}
			galleryApplications = append(galleryApplications, applicationArgs)
		}
		args.GalleryApplications = galleryApplications
	}

	if spec.TerminationNotification != nil {
		terminationNotification := compute.LinuxVirtualMachineTerminationNotificationArgs{
			Enabled: pulumi.Bool(true),
		}
		if spec.TerminationNotification.Timeout != "" {
			terminationNotification.Timeout = pulumi.String(spec.TerminationNotification.Timeout)
		}
		args.TerminationNotification = terminationNotification
	}
	if spec.OsImageNotification != nil {
		osImageNotification := compute.LinuxVirtualMachineOsImageNotificationArgs{}
		if spec.OsImageNotification.Timeout != "" {
			osImageNotification.Timeout = pulumi.String(spec.OsImageNotification.Timeout)
		}
		args.OsImageNotification = osImageNotification
	}

	if spec.Plan != nil {
		args.Plan = compute.LinuxVirtualMachinePlanArgs{
			Name:      pulumi.String(spec.Plan.Name),
			Product:   pulumi.String(spec.Plan.Product),
			Publisher: pulumi.String(spec.Plan.Publisher),
		}
	}

	// custom_data is delivered once at first boot (may embed bootstrap
	// secrets); user_data is IMDS-readable and updatable -- never secret.
	if spec.CustomData != "" {
		args.CustomData = pulumi.String(spec.CustomData)
	}
	if spec.UserData != "" {
		args.UserData = pulumi.String(spec.UserData)
	}

	if spec.ExtensionsTimeBudget != "" {
		args.ExtensionsTimeBudget = pulumi.String(spec.ExtensionsTimeBudget)
	}
	// Presence-guarded true-default optional bools.
	if spec.ProvisionVmAgent != nil {
		args.ProvisionVmAgent = pulumi.Bool(spec.GetProvisionVmAgent())
	} else {
		args.ProvisionVmAgent = pulumi.Bool(true)
	}
	if spec.AllowExtensionOperations != nil {
		args.AllowExtensionOperations = pulumi.Bool(spec.GetAllowExtensionOperations())
	} else {
		args.AllowExtensionOperations = pulumi.Bool(true)
	}

	if locals.DiskControllerType != "" {
		args.DiskControllerType = pulumi.String(locals.DiskControllerType)
	}

	if spec.AdditionalCapabilities != nil {
		args.AdditionalCapabilities = compute.LinuxVirtualMachineAdditionalCapabilitiesArgs{
			UltraSsdEnabled:    pulumi.Bool(spec.AdditionalCapabilities.UltraSsdEnabled),
			HibernationEnabled: pulumi.Bool(spec.AdditionalCapabilities.HibernationEnabled),
		}
	}

	if len(spec.Secrets) > 0 {
		secrets := compute.LinuxVirtualMachineSecretArray{}
		for _, secret := range spec.Secrets {
			certificates := compute.LinuxVirtualMachineSecretCertificateArray{}
			for _, certificate := range secret.Certificates {
				certificates = append(certificates, compute.LinuxVirtualMachineSecretCertificateArgs{
					Url: pulumi.String(certificate.Url),
				})
			}
			secrets = append(secrets, compute.LinuxVirtualMachineSecretArgs{
				KeyVaultId:   pulumi.String(secret.KeyVaultId.GetValue()),
				Certificates: certificates,
			})
		}
		args.Secrets = secrets
	}

	if spec.EdgeZone != "" {
		args.EdgeZone = pulumi.String(spec.EdgeZone)
	}

	return args
}

func buildLinuxOsDisk(locals *Locals) compute.LinuxVirtualMachineOsDiskArgs {
	osDisk := locals.AzureVirtualMachine.Spec.OsDisk

	osDiskArgs := compute.LinuxVirtualMachineOsDiskArgs{
		Caching:            pulumi.String(locals.OsDiskCaching),
		StorageAccountType: pulumi.String(locals.OsDiskStorageType),
	}
	if osDisk.DiskSizeGb != nil {
		osDiskArgs.DiskSizeGb = pulumi.Int(int(osDisk.GetDiskSizeGb()))
	}
	if osDisk.Name != "" {
		osDiskArgs.Name = pulumi.String(osDisk.Name)
	}
	// Ephemeral OS disk: lives on local VM storage, wiped on every
	// stop/deallocate -- stateless fleets only.
	if osDisk.DiffDiskSettings != nil {
		diffDiskArgs := compute.LinuxVirtualMachineOsDiskDiffDiskSettingsArgs{
			Option: pulumi.String("Local"),
		}
		if locals.DiffDiskPlacement != "" {
			diffDiskArgs.Placement = pulumi.String(locals.DiffDiskPlacement)
		}
		osDiskArgs.DiffDiskSettings = diffDiskArgs
	}
	if osDisk.DiskEncryptionSetId.GetValue() != "" {
		osDiskArgs.DiskEncryptionSetId = pulumi.String(osDisk.DiskEncryptionSetId.GetValue())
	}
	if osDisk.SecureVmDiskEncryptionSetId.GetValue() != "" {
		osDiskArgs.SecureVmDiskEncryptionSetId = pulumi.String(osDisk.SecureVmDiskEncryptionSetId.GetValue())
	}
	if locals.SecurityEncryptionType != "" {
		osDiskArgs.SecurityEncryptionType = pulumi.String(locals.SecurityEncryptionType)
	}
	if osDisk.WriteAcceleratorEnabled {
		osDiskArgs.WriteAcceleratorEnabled = pulumi.Bool(true)
	}
	return osDiskArgs
}

// buildWindowsArgs assembles the Windows VM's full argument surface from
// the spec -- the same shape as Linux plus the Windows-only management
// surface (automatic updates, hotpatching, timezone, WinRM, unattend
// content) and minus the SSH-key concept.
func buildWindowsArgs(locals *Locals) *compute.WindowsVirtualMachineArgs {
	spec := locals.AzureVirtualMachine.Spec
	windows := spec.OsProfile.GetWindows()

	args := &compute.WindowsVirtualMachineArgs{
		Name:                pulumi.String(spec.Name),
		Location:            pulumi.String(spec.Region),
		ResourceGroupName:   pulumi.String(locals.ResourceGroupName),
		Size:                pulumi.String(spec.Size),
		NetworkInterfaceIds: pulumi.ToStringArray(locals.NetworkInterfaceIds),
		Tags:                pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.OsProfile.ComputerName != "" {
		args.ComputerName = pulumi.String(spec.OsProfile.ComputerName)
	}

	// Authentication: username + password (Windows has no SSH-key
	// concept). Absent entirely when booting from an existing OS disk.
	if windows.AdminUsername != "" {
		args.AdminUsername = pulumi.String(windows.AdminUsername)
	}
	if windows.AdminPassword.GetValue() != "" {
		args.AdminPassword = pulumi.String(windows.AdminPassword.GetValue())
	}

	args.OsDisk = buildWindowsOsDisk(locals)

	if spec.SourceImageReference != nil {
		args.SourceImageReference = compute.WindowsVirtualMachineSourceImageReferenceArgs{
			Publisher: pulumi.String(spec.SourceImageReference.Publisher),
			Offer:     pulumi.String(spec.SourceImageReference.Offer),
			Sku:       pulumi.String(spec.SourceImageReference.Sku),
			Version:   pulumi.String(spec.SourceImageReference.Version),
		}
	}
	if spec.SourceImageId != "" {
		args.SourceImageId = pulumi.String(spec.SourceImageId)
	}
	if spec.OsManagedDiskId.GetValue() != "" {
		args.OsManagedDiskId = pulumi.String(spec.OsManagedDiskId.GetValue())
	}

	if locals.IdentityType != "" {
		args.Identity = compute.WindowsVirtualMachineIdentityArgs{
			Type:        pulumi.String(locals.IdentityType),
			IdentityIds: pulumi.ToStringArray(locals.IdentityIds),
		}
	}

	if spec.Spot != nil {
		args.Priority = pulumi.String("Spot")
		args.EvictionPolicy = pulumi.String(locals.EvictionPolicy)
		if spec.Spot.MaxBidPrice != nil {
			args.MaxBidPrice = pulumi.Float64(spec.Spot.GetMaxBidPrice())
		}
	}

	if availability := spec.Availability; availability != nil {
		if availability.Zone != "" {
			args.Zone = pulumi.String(availability.Zone)
		}
		if availability.AvailabilitySetId.GetValue() != "" {
			args.AvailabilitySetId = pulumi.String(availability.AvailabilitySetId.GetValue())
		}
		if availability.ProximityPlacementGroupId != "" {
			args.ProximityPlacementGroupId = pulumi.String(availability.ProximityPlacementGroupId)
		}
		if availability.CapacityReservationGroupId != "" {
			args.CapacityReservationGroupId = pulumi.String(availability.CapacityReservationGroupId)
		}
		if availability.DedicatedHostId != "" {
			args.DedicatedHostId = pulumi.String(availability.DedicatedHostId)
		}
		if availability.DedicatedHostGroupId != "" {
			args.DedicatedHostGroupId = pulumi.String(availability.DedicatedHostGroupId)
		}
		if availability.VirtualMachineScaleSetId.GetValue() != "" {
			args.VirtualMachineScaleSetId = pulumi.String(availability.VirtualMachineScaleSetId.GetValue())
		}
		if availability.PlatformFaultDomain != nil {
			args.PlatformFaultDomain = pulumi.Int(int(availability.GetPlatformFaultDomain()))
		}
	}

	if security := spec.Security; security != nil {
		args.SecureBootEnabled = pulumi.Bool(security.SecureBootEnabled)
		args.VtpmEnabled = pulumi.Bool(security.VtpmEnabled)
		args.EncryptionAtHostEnabled = pulumi.Bool(security.EncryptionAtHostEnabled)
	}

	// Patch orchestration: the Windows MODE vocabulary, plus the
	// Windows-only knobs (automatic updates, hotpatching, timezone).
	if locals.WindowsPatchMode != "" {
		args.PatchMode = pulumi.String(locals.WindowsPatchMode)
	}
	if locals.PatchAssessmentMode != "" {
		args.PatchAssessmentMode = pulumi.String(locals.PatchAssessmentMode)
	}
	if locals.RebootSetting != "" {
		args.RebootSetting = pulumi.String(locals.RebootSetting)
	}
	if spec.Patching != nil && spec.Patching.BypassPlatformSafetyChecksOnUserScheduleEnabled {
		args.BypassPlatformSafetyChecksOnUserScheduleEnabled = pulumi.Bool(true)
	}

	// Presence-guarded true-default optional bool.
	if windows.AutomaticUpdatesEnabled != nil {
		args.AutomaticUpdatesEnabled = pulumi.Bool(windows.GetAutomaticUpdatesEnabled())
	} else {
		args.AutomaticUpdatesEnabled = pulumi.Bool(true)
	}
	if windows.HotpatchingEnabled {
		args.HotpatchingEnabled = pulumi.Bool(true)
	}
	if windows.Timezone != "" {
		args.Timezone = pulumi.String(windows.Timezone)
	}

	if len(windows.WinrmListeners) > 0 {
		winrmListeners := compute.WindowsVirtualMachineWinrmListenerArray{}
		for _, listener := range windows.WinrmListeners {
			protocol := "Http"
			if listener.Protocol == azurevirtualmachinev1alpha1.AzureVirtualMachineWinrmProtocol_HTTPS {
				protocol = "Https"
			}
			listenerArgs := compute.WindowsVirtualMachineWinrmListenerArgs{
				Protocol: pulumi.String(protocol),
			}
			if listener.CertificateUrl != "" {
				listenerArgs.CertificateUrl = pulumi.String(listener.CertificateUrl)
			}
			winrmListeners = append(winrmListeners, listenerArgs)
		}
		args.WinrmListeners = winrmListeners
	}

	if len(windows.AdditionalUnattendContents) > 0 {
		unattendContents := compute.WindowsVirtualMachineAdditionalUnattendContentArray{}
		for _, unattendContent := range windows.AdditionalUnattendContents {
			setting := "AutoLogon"
			if unattendContent.Setting == azurevirtualmachinev1alpha1.AzureVirtualMachineUnattendSetting_FIRST_LOGON_COMMANDS {
				setting = "FirstLogonCommands"
			}
			unattendContents = append(unattendContents, compute.WindowsVirtualMachineAdditionalUnattendContentArgs{
				Setting: pulumi.String(setting),
				Content: pulumi.String(unattendContent.Content),
			})
		}
		args.AdditionalUnattendContents = unattendContents
	}

	if locals.WindowsLicenseType != "" {
		args.LicenseType = pulumi.String(locals.WindowsLicenseType)
	}

	if spec.BootDiagnostics != nil && spec.BootDiagnostics.GetEnabled() {
		bootDiagnostics := compute.WindowsVirtualMachineBootDiagnosticsArgs{}
		if spec.BootDiagnostics.StorageAccountUri != "" {
			bootDiagnostics.StorageAccountUri = pulumi.String(spec.BootDiagnostics.StorageAccountUri)
		}
		args.BootDiagnostics = bootDiagnostics
	}

	if len(spec.GalleryApplications) > 0 {
		galleryApplications := compute.WindowsVirtualMachineGalleryApplicationArray{}
		for _, application := range spec.GalleryApplications {
			applicationArgs := compute.WindowsVirtualMachineGalleryApplicationArgs{
				VersionId:                              pulumi.String(application.VersionId),
				AutomaticUpgradeEnabled:                pulumi.Bool(application.AutomaticUpgradeEnabled),
				TreatFailureAsDeploymentFailureEnabled: pulumi.Bool(application.TreatFailureAsDeploymentFailureEnabled),
			}
			if application.Order != nil {
				applicationArgs.Order = pulumi.Int(int(application.GetOrder()))
			}
			if application.Tag != "" {
				applicationArgs.Tag = pulumi.String(application.Tag)
			}
			if application.ConfigurationBlobUri != "" {
				applicationArgs.ConfigurationBlobUri = pulumi.String(application.ConfigurationBlobUri)
			}
			galleryApplications = append(galleryApplications, applicationArgs)
		}
		args.GalleryApplications = galleryApplications
	}

	if spec.TerminationNotification != nil {
		terminationNotification := compute.WindowsVirtualMachineTerminationNotificationArgs{
			Enabled: pulumi.Bool(true),
		}
		if spec.TerminationNotification.Timeout != "" {
			terminationNotification.Timeout = pulumi.String(spec.TerminationNotification.Timeout)
		}
		args.TerminationNotification = terminationNotification
	}
	if spec.OsImageNotification != nil {
		osImageNotification := compute.WindowsVirtualMachineOsImageNotificationArgs{}
		if spec.OsImageNotification.Timeout != "" {
			osImageNotification.Timeout = pulumi.String(spec.OsImageNotification.Timeout)
		}
		args.OsImageNotification = osImageNotification
	}

	if spec.Plan != nil {
		args.Plan = compute.WindowsVirtualMachinePlanArgs{
			Name:      pulumi.String(spec.Plan.Name),
			Product:   pulumi.String(spec.Plan.Product),
			Publisher: pulumi.String(spec.Plan.Publisher),
		}
	}

	if spec.CustomData != "" {
		args.CustomData = pulumi.String(spec.CustomData)
	}
	if spec.UserData != "" {
		args.UserData = pulumi.String(spec.UserData)
	}

	if spec.ExtensionsTimeBudget != "" {
		args.ExtensionsTimeBudget = pulumi.String(spec.ExtensionsTimeBudget)
	}
	if spec.ProvisionVmAgent != nil {
		args.ProvisionVmAgent = pulumi.Bool(spec.GetProvisionVmAgent())
	} else {
		args.ProvisionVmAgent = pulumi.Bool(true)
	}
	if spec.AllowExtensionOperations != nil {
		args.AllowExtensionOperations = pulumi.Bool(spec.GetAllowExtensionOperations())
	} else {
		args.AllowExtensionOperations = pulumi.Bool(true)
	}

	if locals.DiskControllerType != "" {
		args.DiskControllerType = pulumi.String(locals.DiskControllerType)
	}

	if spec.AdditionalCapabilities != nil {
		args.AdditionalCapabilities = compute.WindowsVirtualMachineAdditionalCapabilitiesArgs{
			UltraSsdEnabled:    pulumi.Bool(spec.AdditionalCapabilities.UltraSsdEnabled),
			HibernationEnabled: pulumi.Bool(spec.AdditionalCapabilities.HibernationEnabled),
		}
	}

	if len(spec.Secrets) > 0 {
		secrets := compute.WindowsVirtualMachineSecretArray{}
		for _, secret := range spec.Secrets {
			certificates := compute.WindowsVirtualMachineSecretCertificateArray{}
			for _, certificate := range secret.Certificates {
				certificates = append(certificates, compute.WindowsVirtualMachineSecretCertificateArgs{
					Url: pulumi.String(certificate.Url),
					// Windows installs into a named certificate store.
					Store: pulumi.String(certificate.Store),
				})
			}
			secrets = append(secrets, compute.WindowsVirtualMachineSecretArgs{
				KeyVaultId:   pulumi.String(secret.KeyVaultId.GetValue()),
				Certificates: certificates,
			})
		}
		args.Secrets = secrets
	}

	if spec.EdgeZone != "" {
		args.EdgeZone = pulumi.String(spec.EdgeZone)
	}

	return args
}

func buildWindowsOsDisk(locals *Locals) compute.WindowsVirtualMachineOsDiskArgs {
	osDisk := locals.AzureVirtualMachine.Spec.OsDisk

	osDiskArgs := compute.WindowsVirtualMachineOsDiskArgs{
		Caching:            pulumi.String(locals.OsDiskCaching),
		StorageAccountType: pulumi.String(locals.OsDiskStorageType),
	}
	if osDisk.DiskSizeGb != nil {
		osDiskArgs.DiskSizeGb = pulumi.Int(int(osDisk.GetDiskSizeGb()))
	}
	if osDisk.Name != "" {
		osDiskArgs.Name = pulumi.String(osDisk.Name)
	}
	if osDisk.DiffDiskSettings != nil {
		diffDiskArgs := compute.WindowsVirtualMachineOsDiskDiffDiskSettingsArgs{
			Option: pulumi.String("Local"),
		}
		if locals.DiffDiskPlacement != "" {
			diffDiskArgs.Placement = pulumi.String(locals.DiffDiskPlacement)
		}
		osDiskArgs.DiffDiskSettings = diffDiskArgs
	}
	if osDisk.DiskEncryptionSetId.GetValue() != "" {
		osDiskArgs.DiskEncryptionSetId = pulumi.String(osDisk.DiskEncryptionSetId.GetValue())
	}
	if osDisk.SecureVmDiskEncryptionSetId.GetValue() != "" {
		osDiskArgs.SecureVmDiskEncryptionSetId = pulumi.String(osDisk.SecureVmDiskEncryptionSetId.GetValue())
	}
	if locals.SecurityEncryptionType != "" {
		osDiskArgs.SecurityEncryptionType = pulumi.String(locals.SecurityEncryptionType)
	}
	if osDisk.WriteAcceleratorEnabled {
		osDiskArgs.WriteAcceleratorEnabled = pulumi.Bool(true)
	}
	return osDiskArgs
}
