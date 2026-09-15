package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createUniformLinux realizes the spec as an
// azurerm_linux_virtual_machine_scale_set (UNIFORM orchestration,
// Linux OS profile).
func createUniformLinux(ctx *pulumi.Context, locals *Locals, azureProvider pulumi.ProviderResource) (*scaleSetOutputs, error) {
	spec := locals.AzureVirtualMachineScaleSet.Spec
	linux := spec.OsProfile.GetLinux()

	args := &compute.LinuxVirtualMachineScaleSetArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Sku:               pulumi.String(spec.SkuName),
		AdminUsername:     pulumi.String(linux.AdminUsername),
		Tags:              pulumi.ToStringMap(locals.AzureTags),

		// Azure-default-true gates: unset explicitly falls back to the
		// proto default so stack-input paths that bypass the manifest
		// loader deploy identically on both engines.
		DisablePasswordAuthentication: pulumi.Bool(optionalBool(linux.DisablePasswordAuthentication, true)),
		ProvisionVmAgent:              pulumi.Bool(optionalBool(spec.ProvisionVmAgent, true)),
		ExtensionOperationsEnabled:    pulumi.Bool(optionalBool(spec.ExtensionOperationsEnabled, true)),
		Overprovision:                 pulumi.Bool(optionalBool(spec.Overprovision, true)),
		ZoneBalance:                   pulumi.Bool(spec.GetZoneBalance()),
	}

	if spec.Instances != nil {
		args.Instances = pulumi.IntPtr(int(spec.GetInstances()))
	}
	if linux.AdminPassword.GetValue() != "" {
		args.AdminPassword = pulumi.StringPtr(linux.AdminPassword.GetValue())
	}
	if locals.ComputerNamePrefix != "" {
		args.ComputerNamePrefix = pulumi.StringPtr(locals.ComputerNamePrefix)
	}

	sshKeys := compute.LinuxVirtualMachineScaleSetAdminSshKeyArray{}
	for _, key := range linux.SshPublicKeys {
		// Unset key usernames default to the admin account -- the common
		// case.
		username := key.Username
		if username == "" {
			username = linux.AdminUsername
		}
		sshKeys = append(sshKeys, compute.LinuxVirtualMachineScaleSetAdminSshKeyArgs{
			PublicKey: pulumi.String(key.PublicKey),
			Username:  pulumi.String(username),
		})
	}
	if len(sshKeys) > 0 {
		args.AdminSshKeys = sshKeys
	}

	if spec.CustomData != "" {
		args.CustomData = pulumi.StringPtr(spec.CustomData)
	}
	if spec.UserData != "" {
		args.UserData = pulumi.StringPtr(spec.UserData)
	}

	if spec.SourceImageId != "" {
		args.SourceImageId = pulumi.StringPtr(spec.SourceImageId)
	}
	if spec.SourceImageReference != nil {
		args.SourceImageReference = compute.LinuxVirtualMachineScaleSetSourceImageReferenceArgs{
			Publisher: pulumi.String(spec.SourceImageReference.Publisher),
			Offer:     pulumi.String(spec.SourceImageReference.Offer),
			Sku:       pulumi.String(spec.SourceImageReference.Sku),
			Version:   pulumi.String(spec.SourceImageReference.Version),
		}
	}

	osDisk := compute.LinuxVirtualMachineScaleSetOsDiskArgs{
		Caching:                 pulumi.String(cachingToArm(spec.OsDisk.Caching)),
		StorageAccountType:      pulumi.String(osDiskStorageToArm(spec.OsDisk.StorageAccountType)),
		WriteAcceleratorEnabled: pulumi.Bool(spec.OsDisk.WriteAcceleratorEnabled),
	}
	if spec.OsDisk.DiskSizeGb != nil {
		osDisk.DiskSizeGb = pulumi.IntPtr(int(spec.OsDisk.GetDiskSizeGb()))
	}
	if spec.OsDisk.DiskEncryptionSetId.GetValue() != "" {
		osDisk.DiskEncryptionSetId = pulumi.StringPtr(spec.OsDisk.DiskEncryptionSetId.GetValue())
	}
	if spec.OsDisk.SecureVmDiskEncryptionSetId.GetValue() != "" {
		osDisk.SecureVmDiskEncryptionSetId = pulumi.StringPtr(spec.OsDisk.SecureVmDiskEncryptionSetId.GetValue())
	}
	if enc := securityEncryptionToArm(spec.OsDisk.SecurityEncryptionType); enc != "" {
		osDisk.SecurityEncryptionType = pulumi.StringPtr(enc)
	}
	if spec.OsDisk.DiffDiskSettings != nil {
		diff := compute.LinuxVirtualMachineScaleSetOsDiskDiffDiskSettingsArgs{
			Option: pulumi.String("Local"),
		}
		if placement := diffDiskPlacementToArm(spec.OsDisk.DiffDiskSettings.Placement); placement != "" {
			diff.Placement = pulumi.StringPtr(placement)
		}
		osDisk.DiffDiskSettings = diff
	}
	args.OsDisk = osDisk

	dataDisks := compute.LinuxVirtualMachineScaleSetDataDiskArray{}
	for _, disk := range spec.DataDisks {
		diskArgs := compute.LinuxVirtualMachineScaleSetDataDiskArgs{
			Lun:                     pulumi.Int(int(disk.GetLun())),
			Caching:                 pulumi.String(cachingToArm(disk.Caching)),
			DiskSizeGb:              pulumi.Int(int(disk.DiskSizeGb)),
			StorageAccountType:      pulumi.String(dataDiskStorageToArm(disk.StorageAccountType)),
			CreateOption:            pulumi.StringPtr(createOptionToArm(disk.CreateOption)),
			WriteAcceleratorEnabled: pulumi.Bool(disk.WriteAcceleratorEnabled),
		}
		if disk.Name != "" {
			diskArgs.Name = pulumi.StringPtr(disk.Name)
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

	nics := compute.LinuxVirtualMachineScaleSetNetworkInterfaceArray{}
	for _, nic := range spec.NetworkInterfaces {
		nicArgs := compute.LinuxVirtualMachineScaleSetNetworkInterfaceArgs{
			Name:                        pulumi.String(nic.Name),
			Primary:                     pulumi.Bool(nic.Primary),
			EnableAcceleratedNetworking: pulumi.Bool(nic.AcceleratedNetworkingEnabled),
			EnableIpForwarding:          pulumi.Bool(nic.IpForwardingEnabled),
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

		ipConfigs := compute.LinuxVirtualMachineScaleSetNetworkInterfaceIpConfigurationArray{}
		for _, config := range nic.IpConfigurations {
			configArgs := compute.LinuxVirtualMachineScaleSetNetworkInterfaceIpConfigurationArgs{
				Name:    pulumi.String(config.Name),
				Primary: pulumi.Bool(config.Primary),
				Version: pulumi.StringPtr(ipVersionToArm(config.Version)),
			}
			if config.SubnetId.GetValue() != "" {
				configArgs.SubnetId = pulumi.StringPtr(config.SubnetId.GetValue())
			}
			if pools := refValues(config.LoadBalancerBackendAddressPoolIds); len(pools) > 0 {
				configArgs.LoadBalancerBackendAddressPoolIds = pools
			}
			if natRules := refValues(config.LoadBalancerInboundNatRuleIds); len(natRules) > 0 {
				configArgs.LoadBalancerInboundNatRulesIds = natRules
			}
			if pools := refValues(config.ApplicationGatewayBackendAddressPoolIds); len(pools) > 0 {
				configArgs.ApplicationGatewayBackendAddressPoolIds = pools
			}
			if asgs := refValues(config.ApplicationSecurityGroupIds); len(asgs) > 0 {
				configArgs.ApplicationSecurityGroupIds = asgs
			}
			if pip := config.PublicIpAddress; pip != nil {
				pipArgs := compute.LinuxVirtualMachineScaleSetNetworkInterfaceIpConfigurationPublicIpAddressArgs{
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
				ipTags := compute.LinuxVirtualMachineScaleSetNetworkInterfaceIpConfigurationPublicIpAddressIpTagArray{}
				for _, tag := range pip.IpTags {
					ipTags = append(ipTags, compute.LinuxVirtualMachineScaleSetNetworkInterfaceIpConfigurationPublicIpAddressIpTagArgs{
						Type: pulumi.String(tag.Type),
						Tag:  pulumi.String(tag.Tag),
					})
				}
				if len(ipTags) > 0 {
					pipArgs.IpTags = ipTags
				}
				configArgs.PublicIpAddresses = compute.LinuxVirtualMachineScaleSetNetworkInterfaceIpConfigurationPublicIpAddressArray{pipArgs}
			}
			ipConfigs = append(ipConfigs, configArgs)
		}
		nicArgs.IpConfigurations = ipConfigs
		nics = append(nics, nicArgs)
	}
	args.NetworkInterfaces = nics

	args.UpgradeMode = pulumi.StringPtr(upgradeModeToArm(spec.GetUpgradePolicy().GetMode()))
	if policy := spec.UpgradePolicy; policy != nil {
		if policy.HealthProbeId.GetValue() != "" {
			args.HealthProbeId = pulumi.StringPtr(policy.HealthProbeId.GetValue())
		}
		if rolling := policy.Rolling; rolling != nil {
			args.RollingUpgradePolicy = compute.LinuxVirtualMachineScaleSetRollingUpgradePolicyArgs{
				MaxBatchInstancePercent:             pulumi.Int(int(rolling.MaxBatchInstancePercent)),
				MaxUnhealthyInstancePercent:         pulumi.Int(int(rolling.MaxUnhealthyInstancePercent)),
				MaxUnhealthyUpgradedInstancePercent: pulumi.Int(int(rolling.MaxUnhealthyUpgradedInstancePercent)),
				PauseTimeBetweenBatches:             pulumi.String(rolling.PauseTimeBetweenBatches),
				CrossZoneUpgradesEnabled:            pulumi.BoolPtr(rolling.CrossZoneUpgradesEnabled),
				PrioritizeUnhealthyInstancesEnabled: pulumi.BoolPtr(rolling.PrioritizeUnhealthyInstancesEnabled),
				MaximumSurgeInstancesEnabled:        pulumi.BoolPtr(rolling.MaximumSurgeInstancesEnabled),
			}
		}
		if osUpgrade := policy.AutomaticOsUpgrade; osUpgrade != nil {
			args.AutomaticOsUpgradePolicy = compute.LinuxVirtualMachineScaleSetAutomaticOsUpgradePolicyArgs{
				EnableAutomaticOsUpgrade: pulumi.Bool(osUpgrade.Enabled),
				DisableAutomaticRollback: pulumi.Bool(osUpgrade.DisableAutomaticRollback),
			}
		}
	}

	// Spot presence is the priority switch; the eviction policy is the
	// explicit fleet-level choice.
	if spot := spec.Spot; spot != nil {
		args.Priority = pulumi.StringPtr("Spot")
		args.EvictionPolicy = pulumi.StringPtr(evictionPolicyToArm(spot.EvictionPolicy))
		if spot.MaxBidPrice != nil {
			args.MaxBidPrice = pulumi.Float64Ptr(spot.GetMaxBidPrice())
		}
		if restore := spot.Restore; restore != nil {
			restoreArgs := compute.LinuxVirtualMachineScaleSetSpotRestoreArgs{
				Enabled: pulumi.BoolPtr(true),
			}
			if restore.Timeout != "" {
				restoreArgs.Timeout = pulumi.StringPtr(restore.Timeout)
			}
			args.SpotRestore = restoreArgs
		}
	}

	if repair := spec.AutomaticInstanceRepair; repair != nil {
		repairArgs := compute.LinuxVirtualMachineScaleSetAutomaticInstanceRepairArgs{
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
		notifArgs := compute.LinuxVirtualMachineScaleSetTerminationNotificationArgs{
			Enabled: pulumi.Bool(true),
		}
		if notification.Timeout != "" {
			notifArgs.Timeout = pulumi.StringPtr(notification.Timeout)
		}
		args.TerminationNotification = notifArgs
	}

	if scaleIn := spec.ScaleIn; scaleIn != nil {
		args.ScaleIn = compute.LinuxVirtualMachineScaleSetScaleInArgs{
			Rule:                 pulumi.StringPtr(scaleInRuleToArm(scaleIn.Rule)),
			ForceDeletionEnabled: pulumi.BoolPtr(scaleIn.ForceDeletionEnabled),
		}
	}

	if identity := spec.Identity; identity != nil {
		identityArgs := compute.LinuxVirtualMachineScaleSetIdentityArgs{
			Type: pulumi.String(identityTypeToArm(identity.Type)),
		}
		if ids := refValues(identity.IdentityIds); len(ids) > 0 {
			identityArgs.IdentityIds = ids
		}
		args.Identity = identityArgs
	}

	if diagnostics := spec.BootDiagnostics; diagnostics != nil && diagnostics.GetEnabled() {
		diagArgs := compute.LinuxVirtualMachineScaleSetBootDiagnosticsArgs{}
		if diagnostics.StorageAccountUri != "" {
			diagArgs.StorageAccountUri = pulumi.StringPtr(diagnostics.StorageAccountUri)
		}
		args.BootDiagnostics = diagArgs
	}

	secrets := compute.LinuxVirtualMachineScaleSetSecretArray{}
	for _, secret := range spec.Secrets {
		certs := compute.LinuxVirtualMachineScaleSetSecretCertificateArray{}
		for _, cert := range secret.Certificates {
			certs = append(certs, compute.LinuxVirtualMachineScaleSetSecretCertificateArgs{
				Url: pulumi.String(cert.Url),
			})
		}
		secrets = append(secrets, compute.LinuxVirtualMachineScaleSetSecretArgs{
			KeyVaultId:   pulumi.String(secret.KeyVaultId.GetValue()),
			Certificates: certs,
		})
	}
	if len(secrets) > 0 {
		args.Secrets = secrets
	}

	extensions := compute.LinuxVirtualMachineScaleSetExtensionArray{}
	for _, ext := range spec.Extensions {
		extArgs := compute.LinuxVirtualMachineScaleSetExtensionArgs{
			Name:                    pulumi.String(ext.Name),
			Publisher:               pulumi.String(ext.Publisher),
			Type:                    pulumi.String(ext.Type),
			TypeHandlerVersion:      pulumi.String(ext.TypeHandlerVersion),
			AutoUpgradeMinorVersion: pulumi.BoolPtr(optionalBool(ext.AutoUpgradeMinorVersionEnabled, true)),
			AutomaticUpgradeEnabled: pulumi.BoolPtr(ext.AutomaticUpgradeEnabled),
		}
		if ext.Settings != "" {
			extArgs.Settings = pulumi.StringPtr(ext.Settings)
		}
		if ext.ProtectedSettings != "" {
			extArgs.ProtectedSettings = pulumi.StringPtr(ext.ProtectedSettings)
		}
		if fromVault := ext.ProtectedSettingsFromKeyVault; fromVault != nil {
			extArgs.ProtectedSettingsFromKeyVault = compute.LinuxVirtualMachineScaleSetExtensionProtectedSettingsFromKeyVaultArgs{
				SecretUrl:     pulumi.String(fromVault.SecretUrl),
				SourceVaultId: pulumi.String(fromVault.SourceVaultId.GetValue()),
			}
		}
		if len(ext.ProvisionAfterExtensions) > 0 {
			extArgs.ProvisionAfterExtensions = pulumi.ToStringArray(ext.ProvisionAfterExtensions)
		}
		if ext.ForceUpdateTag != "" {
			extArgs.ForceUpdateTag = pulumi.StringPtr(ext.ForceUpdateTag)
		}
		extensions = append(extensions, extArgs)
	}
	if len(extensions) > 0 {
		args.Extensions = extensions
	}
	if spec.ExtensionsTimeBudget != "" {
		args.ExtensionsTimeBudget = pulumi.StringPtr(spec.ExtensionsTimeBudget)
	}
	if spec.DoNotRunExtensionsOnOverprovisionedMachines != nil {
		args.DoNotRunExtensionsOnOverprovisionedMachines = pulumi.BoolPtr(spec.GetDoNotRunExtensionsOnOverprovisionedMachines())
	}

	if len(spec.Zones) > 0 {
		args.Zones = pulumi.ToStringArray(spec.Zones)
	}
	if spec.PlatformFaultDomainCount != nil {
		args.PlatformFaultDomainCount = pulumi.IntPtr(int(spec.GetPlatformFaultDomainCount()))
	}

	if placement := spec.Placement; placement != nil {
		if placement.ProximityPlacementGroupId != "" {
			args.ProximityPlacementGroupId = pulumi.StringPtr(placement.ProximityPlacementGroupId)
		}
		if placement.CapacityReservationGroupId != "" {
			args.CapacityReservationGroupId = pulumi.StringPtr(placement.CapacityReservationGroupId)
		}
		if placement.HostGroupId != "" {
			args.HostGroupId = pulumi.StringPtr(placement.HostGroupId)
		}
		if placement.SinglePlacementGroup != nil {
			args.SinglePlacementGroup = pulumi.BoolPtr(placement.GetSinglePlacementGroup())
		}
	}

	if security := spec.Security; security != nil {
		args.EncryptionAtHostEnabled = pulumi.BoolPtr(security.EncryptionAtHostEnabled)
		args.SecureBootEnabled = pulumi.BoolPtr(security.SecureBootEnabled)
		args.VtpmEnabled = pulumi.BoolPtr(security.VtpmEnabled)
	}

	if plan := spec.Plan; plan != nil {
		args.Plan = compute.LinuxVirtualMachineScaleSetPlanArgs{
			Name:      pulumi.String(plan.Name),
			Product:   pulumi.String(plan.Product),
			Publisher: pulumi.String(plan.Publisher),
		}
	}

	galleryApps := compute.LinuxVirtualMachineScaleSetGalleryApplicationArray{}
	for _, app := range spec.GalleryApplications {
		appArgs := compute.LinuxVirtualMachineScaleSetGalleryApplicationArgs{
			VersionId: pulumi.String(app.VersionId),
		}
		if app.Order != nil {
			appArgs.Order = pulumi.IntPtr(int(app.GetOrder()))
		}
		if app.Tag != "" {
			appArgs.Tag = pulumi.StringPtr(app.Tag)
		}
		if app.ConfigurationBlobUri != "" {
			appArgs.ConfigurationBlobUri = pulumi.StringPtr(app.ConfigurationBlobUri)
		}
		galleryApps = append(galleryApps, appArgs)
	}
	if len(galleryApps) > 0 {
		args.GalleryApplications = galleryApps
	}

	if capabilities := spec.AdditionalCapabilities; capabilities != nil {
		args.AdditionalCapabilities = compute.LinuxVirtualMachineScaleSetAdditionalCapabilitiesArgs{
			UltraSsdEnabled: pulumi.BoolPtr(capabilities.UltraSsdEnabled),
		}
	}

	if spec.EdgeZone != "" {
		args.EdgeZone = pulumi.StringPtr(spec.EdgeZone)
	}

	scaleSet, err := compute.NewLinuxVirtualMachineScaleSet(ctx,
		spec.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create linux virtual machine scale set %s", spec.Name)
	}

	return &scaleSetOutputs{
		id:          scaleSet.ID().ToStringOutput(),
		name:        scaleSet.Name,
		uniqueId:    scaleSet.UniqueId,
		principalId: principalIdFromLinuxIdentity(scaleSet),
	}, nil
}

// principalIdFromLinuxIdentity surfaces the system-assigned principal,
// or an empty string when the set carries no system-assigned identity.
func principalIdFromLinuxIdentity(scaleSet *compute.LinuxVirtualMachineScaleSet) pulumi.StringOutput {
	return scaleSet.Identity.ApplyT(func(identity *compute.LinuxVirtualMachineScaleSetIdentity) string {
		if identity == nil || identity.PrincipalId == nil {
			return ""
		}
		return *identity.PrincipalId
	}).(pulumi.StringOutput)
}
