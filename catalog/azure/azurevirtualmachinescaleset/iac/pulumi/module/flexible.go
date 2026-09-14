package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createFlexible realizes the spec as an
// azurerm_orchestrated_virtual_machine_scale_set (FLEXIBLE
// orchestration, either OS profile).
func createFlexible(ctx *pulumi.Context, locals *Locals, azureProvider pulumi.ProviderResource) (*scaleSetOutputs, error) {
	spec := locals.AzureVirtualMachineScaleSet.Spec

	args := &compute.OrchestratedVirtualMachineScaleSetArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		SkuName:           pulumi.StringPtr(spec.SkuName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),

		// Fault domains are the FLEXIBLE resilience contract: 1 with
		// zones (zones are the resilience unit) or the region's max for
		// regional spreading. Required by ARM; spec-level validation
		// guarantees presence.
		PlatformFaultDomainCount: pulumi.Int(int(spec.GetPlatformFaultDomainCount())),

		ExtensionOperationsEnabled: pulumi.Bool(optionalBool(spec.ExtensionOperationsEnabled, true)),
		ZoneBalance:                pulumi.Bool(spec.GetZoneBalance()),
	}

	if spec.Instances != nil {
		args.Instances = pulumi.IntPtr(int(spec.GetInstances()))
	}
	if spec.NetworkApiVersion != "" {
		args.NetworkApiVersion = pulumi.StringPtr(spec.NetworkApiVersion)
	}

	if profile := spec.SkuProfile; profile != nil {
		// PARITY-EXCEPTION: the pinned pulumi-azure v6 SDK bridges the
		// legacy sku_profile shape (plain vm_sizes, no per-size ranks)
		// while the Terraform module realizes ranked virtual_machine_size
		// blocks. Sizes deploy identically on both engines; a PRIORITIZED
		// rank cannot be honored here, so it fails loudly rather than
		// silently degrading. Output-neutral (sku_profile never feeds
		// stack outputs); revisit when the SDK catches up.
		vmSizes := pulumi.StringArray{}
		for _, size := range profile.VmSizes {
			if size.Rank != nil {
				return nil, errors.Errorf(
					"sku_profile vm_size %s carries a rank, which the pinned pulumi-azure SDK cannot express -- deploy ranked profiles with the Terraform module or drop the ranks",
					size.Name)
			}
			vmSizes = append(vmSizes, pulumi.String(size.Name))
		}
		args.SkuProfile = compute.OrchestratedVirtualMachineScaleSetSkuProfileArgs{
			AllocationStrategy: pulumi.String(allocationStrategyToArm(profile.AllocationStrategy)),
			VmSizes:            vmSizes,
		}
	}

	osProfile := compute.OrchestratedVirtualMachineScaleSetOsProfileArgs{}
	if spec.CustomData != "" {
		osProfile.CustomData = pulumi.StringPtr(spec.CustomData)
	}

	if linux := spec.OsProfile.GetLinux(); linux != nil {
		linuxConfig := compute.OrchestratedVirtualMachineScaleSetOsProfileLinuxConfigurationArgs{
			AdminUsername: pulumi.String(linux.AdminUsername),
			// Azure-default-true gates: unset explicitly falls back to
			// the proto default so both engines deploy identically.
			DisablePasswordAuthentication: pulumi.BoolPtr(optionalBool(linux.DisablePasswordAuthentication, true)),
			ProvisionVmAgent:              pulumi.BoolPtr(optionalBool(spec.ProvisionVmAgent, true)),
		}
		if linux.AdminPassword.GetValue() != "" {
			linuxConfig.AdminPassword = pulumi.StringPtr(linux.AdminPassword.GetValue())
		}
		if locals.ComputerNamePrefix != "" {
			linuxConfig.ComputerNamePrefix = pulumi.StringPtr(locals.ComputerNamePrefix)
		}
		if mode := linuxPatchModeToArm(linux.PatchMode); mode != "" {
			linuxConfig.PatchMode = pulumi.StringPtr(mode)
		}
		if mode := assessmentModeToArm(linux.PatchAssessmentMode); mode != "" {
			linuxConfig.PatchAssessmentMode = pulumi.StringPtr(mode)
		}

		sshKeys := compute.OrchestratedVirtualMachineScaleSetOsProfileLinuxConfigurationAdminSshKeyArray{}
		for _, key := range linux.SshPublicKeys {
			username := key.Username
			if username == "" {
				username = linux.AdminUsername
			}
			sshKeys = append(sshKeys, compute.OrchestratedVirtualMachineScaleSetOsProfileLinuxConfigurationAdminSshKeyArgs{
				PublicKey: pulumi.String(key.PublicKey),
				Username:  pulumi.String(username),
			})
		}
		if len(sshKeys) > 0 {
			linuxConfig.AdminSshKeys = sshKeys
		}

		linuxSecrets := compute.OrchestratedVirtualMachineScaleSetOsProfileLinuxConfigurationSecretArray{}
		for _, secret := range spec.Secrets {
			certs := compute.OrchestratedVirtualMachineScaleSetOsProfileLinuxConfigurationSecretCertificateArray{}
			for _, cert := range secret.Certificates {
				certs = append(certs, compute.OrchestratedVirtualMachineScaleSetOsProfileLinuxConfigurationSecretCertificateArgs{
					Url: pulumi.String(cert.Url),
				})
			}
			linuxSecrets = append(linuxSecrets, compute.OrchestratedVirtualMachineScaleSetOsProfileLinuxConfigurationSecretArgs{
				KeyVaultId:   pulumi.String(secret.KeyVaultId.GetValue()),
				Certificates: certs,
			})
		}
		if len(linuxSecrets) > 0 {
			linuxConfig.Secrets = linuxSecrets
		}

		osProfile.LinuxConfiguration = linuxConfig
	}

	if windows := spec.OsProfile.GetWindows(); windows != nil {
		windowsConfig := compute.OrchestratedVirtualMachineScaleSetOsProfileWindowsConfigurationArgs{
			AdminUsername:          pulumi.String(windows.AdminUsername),
			AdminPassword:          pulumi.String(windows.AdminPassword.GetValue()),
			EnableAutomaticUpdates: pulumi.BoolPtr(optionalBool(windows.AutomaticUpdatesEnabled, true)),
			HotpatchingEnabled:     pulumi.BoolPtr(windows.HotpatchingEnabled),
			ProvisionVmAgent:       pulumi.BoolPtr(optionalBool(spec.ProvisionVmAgent, true)),
		}
		if locals.ComputerNamePrefix != "" {
			windowsConfig.ComputerNamePrefix = pulumi.StringPtr(locals.ComputerNamePrefix)
		}
		if mode := windowsPatchModeToArm(windows.PatchMode); mode != "" {
			windowsConfig.PatchMode = pulumi.StringPtr(mode)
		}
		if mode := assessmentModeToArm(windows.PatchAssessmentMode); mode != "" {
			windowsConfig.PatchAssessmentMode = pulumi.StringPtr(mode)
		}
		if windows.Timezone != "" {
			windowsConfig.Timezone = pulumi.StringPtr(windows.Timezone)
		}

		winrmListeners := compute.OrchestratedVirtualMachineScaleSetOsProfileWindowsConfigurationWinrmListenerArray{}
		for _, listener := range windows.WinrmListeners {
			listenerArgs := compute.OrchestratedVirtualMachineScaleSetOsProfileWindowsConfigurationWinrmListenerArgs{
				Protocol: pulumi.String(winrmProtocolToArm(listener.Protocol)),
			}
			if listener.CertificateUrl != "" {
				listenerArgs.CertificateUrl = pulumi.StringPtr(listener.CertificateUrl)
			}
			winrmListeners = append(winrmListeners, listenerArgs)
		}
		if len(winrmListeners) > 0 {
			windowsConfig.WinrmListeners = winrmListeners
		}

		unattendContents := compute.OrchestratedVirtualMachineScaleSetOsProfileWindowsConfigurationAdditionalUnattendContentArray{}
		for _, content := range windows.AdditionalUnattendContents {
			unattendContents = append(unattendContents, compute.OrchestratedVirtualMachineScaleSetOsProfileWindowsConfigurationAdditionalUnattendContentArgs{
				Setting: pulumi.String(unattendSettingToArm(content.Setting)),
				Content: pulumi.String(content.Content),
			})
		}
		if len(unattendContents) > 0 {
			windowsConfig.AdditionalUnattendContents = unattendContents
		}

		windowsSecrets := compute.OrchestratedVirtualMachineScaleSetOsProfileWindowsConfigurationSecretArray{}
		for _, secret := range spec.Secrets {
			certs := compute.OrchestratedVirtualMachineScaleSetOsProfileWindowsConfigurationSecretCertificateArray{}
			for _, cert := range secret.Certificates {
				certs = append(certs, compute.OrchestratedVirtualMachineScaleSetOsProfileWindowsConfigurationSecretCertificateArgs{
					Url:   pulumi.String(cert.Url),
					Store: pulumi.String(cert.Store),
				})
			}
			windowsSecrets = append(windowsSecrets, compute.OrchestratedVirtualMachineScaleSetOsProfileWindowsConfigurationSecretArgs{
				KeyVaultId:   pulumi.String(secret.KeyVaultId.GetValue()),
				Certificates: certs,
			})
		}
		if len(windowsSecrets) > 0 {
			windowsConfig.Secrets = windowsSecrets
		}

		osProfile.WindowsConfiguration = windowsConfig

		// Azure Hybrid Benefit lives top-level on the FLEXIBLE resource
		// (Windows fleets only).
		if license := windowsLicenseToArm(windows.LicenseType); license != "" {
			args.LicenseType = pulumi.StringPtr(license)
		}
	}
	args.OsProfile = osProfile

	// The FLEXIBLE resource takes user data pre-encoded (the spec
	// carries base64 already).
	if spec.UserData != "" {
		args.UserDataBase64 = pulumi.StringPtr(spec.UserData)
	}

	if spec.SourceImageId != "" {
		args.SourceImageId = pulumi.StringPtr(spec.SourceImageId)
	}
	if spec.SourceImageReference != nil {
		args.SourceImageReference = compute.OrchestratedVirtualMachineScaleSetSourceImageReferenceArgs{
			Publisher: pulumi.String(spec.SourceImageReference.Publisher),
			Offer:     pulumi.String(spec.SourceImageReference.Offer),
			Sku:       pulumi.String(spec.SourceImageReference.Sku),
			Version:   pulumi.String(spec.SourceImageReference.Version),
		}
	}

	osDisk := compute.OrchestratedVirtualMachineScaleSetOsDiskArgs{
		Caching:                 pulumi.String(cachingToArm(spec.OsDisk.Caching)),
		StorageAccountType:      pulumi.String(osDiskStorageToArm(spec.OsDisk.StorageAccountType)),
		WriteAcceleratorEnabled: pulumi.BoolPtr(spec.OsDisk.WriteAcceleratorEnabled),
	}
	if spec.OsDisk.DiskSizeGb != nil {
		osDisk.DiskSizeGb = pulumi.IntPtr(int(spec.OsDisk.GetDiskSizeGb()))
	}
	if spec.OsDisk.DiskEncryptionSetId.GetValue() != "" {
		osDisk.DiskEncryptionSetId = pulumi.StringPtr(spec.OsDisk.DiskEncryptionSetId.GetValue())
	}
	if spec.OsDisk.DiffDiskSettings != nil {
		diff := compute.OrchestratedVirtualMachineScaleSetOsDiskDiffDiskSettingsArgs{
			Option: pulumi.String("Local"),
		}
		if placement := diffDiskPlacementToArm(spec.OsDisk.DiffDiskSettings.Placement); placement != "" {
			diff.Placement = pulumi.StringPtr(placement)
		}
		osDisk.DiffDiskSettings = diff
	}
	args.OsDisk = osDisk

	dataDisks := compute.OrchestratedVirtualMachineScaleSetDataDiskArray{}
	for _, disk := range spec.DataDisks {
		diskArgs := compute.OrchestratedVirtualMachineScaleSetDataDiskArgs{
			Lun:                     pulumi.IntPtr(int(disk.GetLun())),
			Caching:                 pulumi.String(cachingToArm(disk.Caching)),
			DiskSizeGb:              pulumi.Int(int(disk.DiskSizeGb)),
			StorageAccountType:      pulumi.String(dataDiskStorageToArm(disk.StorageAccountType)),
			CreateOption:            pulumi.StringPtr(createOptionToArm(disk.CreateOption)),
			WriteAcceleratorEnabled: pulumi.BoolPtr(disk.WriteAcceleratorEnabled),
		}
		if disk.DiskEncryptionSetId.GetValue() != "" {
			diskArgs.DiskEncryptionSetId = pulumi.StringPtr(disk.DiskEncryptionSetId.GetValue())
		}
		if disk.UltraSsdDiskIopsReadWrite != nil {
			diskArgs.UltraSsdDiskIopsReadWrite = pulumi.IntPtr(int(disk.GetUltraSsdDiskIopsReadWrite()))
		}
		if disk.UltraSsdDiskMbpsReadWrite != nil {
			diskArgs.UltraSsdDiskMbpsReadWrite = pulumi.IntPtr(int(disk.GetUltraSsdDiskMbpsReadWrite()))
		}
		dataDisks = append(dataDisks, diskArgs)
	}
	if len(dataDisks) > 0 {
		args.DataDisks = dataDisks
	}

	nics := compute.OrchestratedVirtualMachineScaleSetNetworkInterfaceArray{}
	for _, nic := range spec.NetworkInterfaces {
		nicArgs := compute.OrchestratedVirtualMachineScaleSetNetworkInterfaceArgs{
			Name:                        pulumi.String(nic.Name),
			Primary:                     pulumi.BoolPtr(nic.Primary),
			EnableAcceleratedNetworking: pulumi.BoolPtr(nic.AcceleratedNetworkingEnabled),
			EnableIpForwarding:          pulumi.BoolPtr(nic.IpForwardingEnabled),
		}
		if len(nic.DnsServers) > 0 {
			nicArgs.DnsServers = pulumi.ToStringArray(nic.DnsServers)
		}
		if nic.NetworkSecurityGroupId.GetValue() != "" {
			nicArgs.NetworkSecurityGroupId = pulumi.StringPtr(nic.NetworkSecurityGroupId.GetValue())
		}
		if mode := auxiliaryModeToArm(nic.AuxiliaryMode); mode != "" {
			nicArgs.AuxiliaryMode = pulumi.StringPtr(mode)
			nicArgs.AuxiliarySku = pulumi.StringPtr(auxiliarySkuToArm(nic.AuxiliarySku))
		}

		ipConfigs := compute.OrchestratedVirtualMachineScaleSetNetworkInterfaceIpConfigurationArray{}
		for _, config := range nic.IpConfigurations {
			configArgs := compute.OrchestratedVirtualMachineScaleSetNetworkInterfaceIpConfigurationArgs{
				Name:    pulumi.String(config.Name),
				Primary: pulumi.BoolPtr(config.Primary),
				Version: pulumi.StringPtr(ipVersionToArm(config.Version)),
			}
			if config.SubnetId.GetValue() != "" {
				configArgs.SubnetId = pulumi.StringPtr(config.SubnetId.GetValue())
			}
			if pools := refValues(config.LoadBalancerBackendAddressPoolIds); len(pools) > 0 {
				configArgs.LoadBalancerBackendAddressPoolIds = pools
			}
			if pools := refValues(config.ApplicationGatewayBackendAddressPoolIds); len(pools) > 0 {
				configArgs.ApplicationGatewayBackendAddressPoolIds = pools
			}
			if asgs := refValues(config.ApplicationSecurityGroupIds); len(asgs) > 0 {
				configArgs.ApplicationSecurityGroupIds = asgs
			}
			if pip := config.PublicIpAddress; pip != nil {
				pipArgs := compute.OrchestratedVirtualMachineScaleSetNetworkInterfaceIpConfigurationPublicIpAddressArgs{
					Name: pulumi.String(pip.Name),
				}
				if pip.DomainNameLabel != "" {
					pipArgs.DomainNameLabel = pulumi.StringPtr(pip.DomainNameLabel)
				}
				if pip.IdleTimeoutInMinutes != nil {
					pipArgs.IdleTimeoutInMinutes = pulumi.IntPtr(int(pip.GetIdleTimeoutInMinutes()))
				}
				if pip.Version != 0 {
					pipArgs.Version = pulumi.StringPtr(ipVersionToArm(pip.Version))
				}
				if pip.PublicIpPrefixId.GetValue() != "" {
					pipArgs.PublicIpPrefixId = pulumi.StringPtr(pip.PublicIpPrefixId.GetValue())
				}
				ipTags := compute.OrchestratedVirtualMachineScaleSetNetworkInterfaceIpConfigurationPublicIpAddressIpTagArray{}
				for _, tag := range pip.IpTags {
					ipTags = append(ipTags, compute.OrchestratedVirtualMachineScaleSetNetworkInterfaceIpConfigurationPublicIpAddressIpTagArgs{
						Type: pulumi.String(tag.Type),
						Tag:  pulumi.String(tag.Tag),
					})
				}
				if len(ipTags) > 0 {
					pipArgs.IpTags = ipTags
				}
				configArgs.PublicIpAddresses = compute.OrchestratedVirtualMachineScaleSetNetworkInterfaceIpConfigurationPublicIpAddressArray{pipArgs}
			}
			ipConfigs = append(ipConfigs, configArgs)
		}
		nicArgs.IpConfigurations = ipConfigs
		nics = append(nics, nicArgs)
	}
	args.NetworkInterfaces = nics

	args.UpgradeMode = pulumi.StringPtr(upgradeModeToArm(spec.GetUpgradePolicy().GetMode()))
	if policy := spec.UpgradePolicy; policy != nil {
		if rolling := policy.Rolling; rolling != nil {
			args.RollingUpgradePolicy = compute.OrchestratedVirtualMachineScaleSetRollingUpgradePolicyArgs{
				MaxBatchInstancePercent:             pulumi.Int(int(rolling.MaxBatchInstancePercent)),
				MaxUnhealthyInstancePercent:         pulumi.Int(int(rolling.MaxUnhealthyInstancePercent)),
				MaxUnhealthyUpgradedInstancePercent: pulumi.Int(int(rolling.MaxUnhealthyUpgradedInstancePercent)),
				PauseTimeBetweenBatches:             pulumi.String(rolling.PauseTimeBetweenBatches),
				CrossZoneUpgradesEnabled:            pulumi.BoolPtr(rolling.CrossZoneUpgradesEnabled),
				PrioritizeUnhealthyInstancesEnabled: pulumi.BoolPtr(rolling.PrioritizeUnhealthyInstancesEnabled),
				MaximumSurgeInstancesEnabled:        pulumi.BoolPtr(rolling.MaximumSurgeInstancesEnabled),
			}
		}
	}

	// Spot presence is the priority switch; priority_mix is FLEXIBLE's
	// spot/on-demand blending.
	if spot := spec.Spot; spot != nil {
		args.Priority = pulumi.StringPtr("Spot")
		args.EvictionPolicy = pulumi.StringPtr(evictionPolicyToArm(spot.EvictionPolicy))
		if spot.MaxBidPrice != nil {
			args.MaxBidPrice = pulumi.Float64Ptr(spot.GetMaxBidPrice())
		}
		if mix := spot.PriorityMix; mix != nil {
			mixArgs := compute.OrchestratedVirtualMachineScaleSetPriorityMixArgs{}
			if mix.BaseRegularCount != nil {
				mixArgs.BaseRegularCount = pulumi.IntPtr(int(mix.GetBaseRegularCount()))
			}
			if mix.RegularPercentageAboveBase != nil {
				mixArgs.RegularPercentageAboveBase = pulumi.IntPtr(int(mix.GetRegularPercentageAboveBase()))
			}
			args.PriorityMix = mixArgs
		}
	}

	if repair := spec.AutomaticInstanceRepair; repair != nil {
		repairArgs := compute.OrchestratedVirtualMachineScaleSetAutomaticInstanceRepairArgs{
			Enabled: pulumi.Bool(repair.Enabled),
		}
		if repair.GracePeriod != "" {
			repairArgs.GracePeriod = pulumi.StringPtr(repair.GracePeriod)
		}
		if action := repairActionToArm(repair.Action); action != "" {
			repairArgs.Action = pulumi.StringPtr(action)
		}
		args.AutomaticInstanceRepair = repairArgs
	}

	if notification := spec.TerminationNotification; notification != nil {
		notifArgs := compute.OrchestratedVirtualMachineScaleSetTerminationNotificationArgs{
			Enabled: pulumi.Bool(true),
		}
		if notification.Timeout != "" {
			notifArgs.Timeout = pulumi.StringPtr(notification.Timeout)
		}
		args.TerminationNotification = notifArgs
	}

	if identity := spec.Identity; identity != nil {
		args.Identity = compute.OrchestratedVirtualMachineScaleSetIdentityArgs{
			Type:        pulumi.String(identityTypeToArm(identity.Type)),
			IdentityIds: refValues(identity.IdentityIds),
		}
	}

	if diagnostics := spec.BootDiagnostics; diagnostics != nil && diagnostics.GetEnabled() {
		diagArgs := compute.OrchestratedVirtualMachineScaleSetBootDiagnosticsArgs{}
		if diagnostics.StorageAccountUri != "" {
			diagArgs.StorageAccountUri = pulumi.StringPtr(diagnostics.StorageAccountUri)
		}
		args.BootDiagnostics = diagArgs
	}

	extensions := compute.OrchestratedVirtualMachineScaleSetExtensionArray{}
	for _, ext := range spec.Extensions {
		extArgs := compute.OrchestratedVirtualMachineScaleSetExtensionArgs{
			Name:                           pulumi.String(ext.Name),
			Publisher:                      pulumi.String(ext.Publisher),
			Type:                           pulumi.String(ext.Type),
			TypeHandlerVersion:             pulumi.String(ext.TypeHandlerVersion),
			AutoUpgradeMinorVersionEnabled: pulumi.BoolPtr(optionalBool(ext.AutoUpgradeMinorVersionEnabled, true)),
			FailureSuppressionEnabled:      pulumi.BoolPtr(ext.FailureSuppressionEnabled),
		}
		if ext.Settings != "" {
			extArgs.Settings = pulumi.StringPtr(ext.Settings)
		}
		if ext.ProtectedSettings != "" {
			extArgs.ProtectedSettings = pulumi.StringPtr(ext.ProtectedSettings)
		}
		if fromVault := ext.ProtectedSettingsFromKeyVault; fromVault != nil {
			extArgs.ProtectedSettingsFromKeyVault = compute.OrchestratedVirtualMachineScaleSetExtensionProtectedSettingsFromKeyVaultArgs{
				SecretUrl:     pulumi.String(fromVault.SecretUrl),
				SourceVaultId: pulumi.String(fromVault.SourceVaultId.GetValue()),
			}
		}
		if len(ext.ProvisionAfterExtensions) > 0 {
			extArgs.ExtensionsToProvisionAfterVmCreations = pulumi.ToStringArray(ext.ProvisionAfterExtensions)
		}
		if ext.ForceUpdateTag != "" {
			extArgs.ForceExtensionExecutionOnChange = pulumi.StringPtr(ext.ForceUpdateTag)
		}
		extensions = append(extensions, extArgs)
	}
	if len(extensions) > 0 {
		args.Extensions = extensions
	}
	if spec.ExtensionsTimeBudget != "" {
		args.ExtensionsTimeBudget = pulumi.StringPtr(spec.ExtensionsTimeBudget)
	}

	if len(spec.Zones) > 0 {
		args.Zones = pulumi.ToStringArray(spec.Zones)
	}

	if placement := spec.Placement; placement != nil {
		if placement.ProximityPlacementGroupId != "" {
			args.ProximityPlacementGroupId = pulumi.StringPtr(placement.ProximityPlacementGroupId)
		}
		if placement.CapacityReservationGroupId != "" {
			args.CapacityReservationGroupId = pulumi.StringPtr(placement.CapacityReservationGroupId)
		}
		if placement.SinglePlacementGroup != nil {
			args.SinglePlacementGroup = pulumi.BoolPtr(placement.GetSinglePlacementGroup())
		}
	}

	if security := spec.Security; security != nil {
		args.EncryptionAtHostEnabled = pulumi.BoolPtr(security.EncryptionAtHostEnabled)
	}

	if plan := spec.Plan; plan != nil {
		args.Plan = compute.OrchestratedVirtualMachineScaleSetPlanArgs{
			Name:      pulumi.String(plan.Name),
			Product:   pulumi.String(plan.Product),
			Publisher: pulumi.String(plan.Publisher),
		}
	}

	if capabilities := spec.AdditionalCapabilities; capabilities != nil {
		args.AdditionalCapabilities = compute.OrchestratedVirtualMachineScaleSetAdditionalCapabilitiesArgs{
			UltraSsdEnabled: pulumi.BoolPtr(capabilities.UltraSsdEnabled),
		}
	}

	scaleSet, err := compute.NewOrchestratedVirtualMachineScaleSet(ctx,
		spec.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create orchestrated virtual machine scale set %s", spec.Name)
	}

	return &scaleSetOutputs{
		id:       scaleSet.ID().ToStringOutput(),
		name:     scaleSet.Name,
		uniqueId: scaleSet.UniqueId,
		// FLEXIBLE sets support user-assigned identities only; the
		// system-assigned principal seam is a UNIFORM capability.
		principalId: pulumi.String("").ToStringOutput(),
	}, nil
}
