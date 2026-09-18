package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// computeInstance enables the Compute Engine API and creates the instance.
//
// Sharp edges, all taught by the API rather than invented here:
//
//   - zone, boot source, NIC count/networks, scratch disks, hostname,
//     confidential mode, and reservation affinity are immutable — changing
//     them replaces the VM. machine_type, service_account, and several
//     others update via stop/start, which the provider performs only when
//     allow_stopping_for_update is true.
//
//   - Spot: provisioning_model = "SPOT" requires the API's legacy
//     preemptible flag and no automatic restart — both derived here,
//     identically to the Terraform module, so the spec's single switch
//     stays honest.
//
//   - desired_status starts/suspends/stops in place ("RUNNING",
//     "SUSPENDED", "TERMINATED"); unset follows GCP (running).
//
//   - deletion_protection guards the VM object only; data protection
//     lives on the disks (boot auto_delete, GcpComputeDisk lifecycles).
func computeInstance(
	ctx *pulumi.Context,
	locals *Locals,
	gcpProvider *gcp.Provider,
) (*compute.Instance, error) {

	spec := locals.GcpComputeInstance.Spec

	// Enable the Compute Engine API — the control plane that owns
	// instances. DisableOnDestroy stays false: tearing down one VM must
	// never disable the API for everything else in the project.
	computeApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("compute.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	// Honor the spec contract: an empty project_id falls back to the
	// provider's default project.
	if spec.ProjectId.GetValue() != "" {
		computeApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdComputeApi, err := projects.NewService(ctx,
		"gcpvm-compute.googleapis.com", computeApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to enable compute.googleapis.com api")
	}

	// Boot disk: exactly one source, enforced pre-deploy by the spec's
	// CEL. An existing bootable disk attaches via Source; image/snapshot
	// create a fresh disk through InitializeParams. Unset size/type
	// default server-side (image size, machine-family default type).
	bootDiskArgs := &compute.InstanceBootDiskArgs{}
	if spec.BootDisk.AutoDelete != nil {
		bootDiskArgs.AutoDelete = pulumi.BoolPtr(spec.BootDisk.GetAutoDelete())
	} else {
		bootDiskArgs.AutoDelete = pulumi.BoolPtr(true)
	}
	if spec.BootDisk.DeviceName != "" {
		bootDiskArgs.DeviceName = pulumi.StringPtr(spec.BootDisk.DeviceName)
	}
	if spec.BootDisk.KmsKey.GetValue() != "" {
		bootDiskArgs.KmsKeySelfLink = pulumi.StringPtr(spec.BootDisk.KmsKey.GetValue())
	}
	if spec.BootDisk.KmsKeyServiceAccount != "" {
		bootDiskArgs.DiskEncryptionServiceAccount = pulumi.StringPtr(spec.BootDisk.KmsKeyServiceAccount)
	}
	if spec.BootDisk.Mode != "" {
		bootDiskArgs.Mode = pulumi.StringPtr(spec.BootDisk.Mode)
	}
	// Google-advice-only lever — sent only when explicitly set, so the
	// API's auto-selected interface never produces a diff.
	if spec.BootDisk.Interface != "" {
		bootDiskArgs.Interface = pulumi.StringPtr(spec.BootDisk.Interface)
	}
	// Regional-disk takeover; forcing a zonal disk is an API error.
	if spec.BootDisk.ForceAttach {
		bootDiskArgs.ForceAttach = pulumi.BoolPtr(true)
	}
	if len(spec.BootDisk.GuestOsFeatures) > 0 {
		bootDiskArgs.GuestOsFeatures = pulumi.ToStringArray(spec.BootDisk.GuestOsFeatures)
	}
	if spec.BootDisk.SourceDisk.GetValue() != "" {
		bootDiskArgs.Source = pulumi.StringPtr(spec.BootDisk.SourceDisk.GetValue())
	} else {
		initializeParams := &compute.InstanceBootDiskInitializeParamsArgs{}
		if spec.BootDisk.Image != "" {
			initializeParams.Image = pulumi.StringPtr(spec.BootDisk.Image)
		}
		if spec.BootDisk.SourceSnapshot != "" {
			initializeParams.Snapshot = pulumi.StringPtr(spec.BootDisk.SourceSnapshot)
		}
		if spec.BootDisk.SizeGb > 0 {
			initializeParams.Size = pulumi.IntPtr(int(spec.BootDisk.SizeGb))
		}
		if spec.BootDisk.Type != "" {
			initializeParams.Type = pulumi.StringPtr(spec.BootDisk.Type)
		}
		if len(spec.BootDisk.DiskLabels) > 0 {
			initializeParams.Labels = pulumi.ToStringMap(spec.BootDisk.DiskLabels)
		}
		if spec.BootDisk.ProvisionedIops != nil {
			initializeParams.ProvisionedIops = pulumi.IntPtr(int(spec.BootDisk.GetProvisionedIops()))
		}
		if spec.BootDisk.ProvisionedThroughput != nil {
			initializeParams.ProvisionedThroughput = pulumi.IntPtr(int(spec.BootDisk.GetProvisionedThroughput()))
		}
		if spec.BootDisk.Architecture != "" {
			initializeParams.Architecture = pulumi.StringPtr(spec.BootDisk.Architecture)
		}
		if spec.BootDisk.EnableConfidentialCompute {
			initializeParams.EnableConfidentialCompute = pulumi.BoolPtr(true)
		}
		// The bridged provider flattens this max-1 list to a single
		// string; the spec's repeated shape (capped at 1 by validation)
		// mirrors the provider API.
		if len(spec.BootDisk.ResourcePolicies) > 0 {
			initializeParams.ResourcePolicies = pulumi.StringPtr(spec.BootDisk.ResourcePolicies[0])
		}
		if spec.BootDisk.StoragePool != "" {
			initializeParams.StoragePool = pulumi.StringPtr(spec.BootDisk.StoragePool)
		}
		// Exactly two zones (one the instance's own) converts the boot
		// disk to a regional disk — enforced pre-deploy by the spec.
		if len(spec.BootDisk.ReplicaZones) > 0 {
			initializeParams.ReplicaZones = pulumi.ToStringArray(spec.BootDisk.ReplicaZones)
		}
		// Create-time only; not returned by the API.
		if len(spec.BootDisk.ResourceManagerTags) > 0 {
			initializeParams.ResourceManagerTags = pulumi.ToStringMap(spec.BootDisk.ResourceManagerTags)
		}
		// CMEK decryption of an encrypted source image/snapshot. CSEK
		// raw-key arms are deliberately not modeled (secure-by-default).
		if spec.BootDisk.SourceImageEncryption != nil {
			keyArgs := &compute.InstanceBootDiskInitializeParamsSourceImageEncryptionKeyArgs{
				KmsKeySelfLink: pulumi.StringPtr(spec.BootDisk.SourceImageEncryption.KmsKey.GetValue()),
			}
			if spec.BootDisk.SourceImageEncryption.KmsKeyServiceAccount != "" {
				keyArgs.KmsKeyServiceAccount = pulumi.StringPtr(spec.BootDisk.SourceImageEncryption.KmsKeyServiceAccount)
			}
			initializeParams.SourceImageEncryptionKey = keyArgs
		}
		if spec.BootDisk.SourceSnapshotEncryption != nil {
			keyArgs := &compute.InstanceBootDiskInitializeParamsSourceSnapshotEncryptionKeyArgs{
				KmsKeySelfLink: pulumi.StringPtr(spec.BootDisk.SourceSnapshotEncryption.KmsKey.GetValue()),
			}
			if spec.BootDisk.SourceSnapshotEncryption.KmsKeyServiceAccount != "" {
				keyArgs.KmsKeyServiceAccount = pulumi.StringPtr(spec.BootDisk.SourceSnapshotEncryption.KmsKeyServiceAccount)
			}
			initializeParams.SourceSnapshotEncryptionKey = keyArgs
		}
		bootDiskArgs.InitializeParams = initializeParams
	}

	// Data disks are pre-existing GcpComputeDisk resources attached by
	// reference — the disk's own lifecycle protects the data.
	attachedDisks := compute.InstanceAttachedDiskArray{}
	for _, disk := range spec.AttachedDisks {
		diskArgs := &compute.InstanceAttachedDiskArgs{
			Source: pulumi.String(disk.Source.GetValue()),
		}
		if disk.DeviceName != "" {
			diskArgs.DeviceName = pulumi.StringPtr(disk.DeviceName)
		}
		if disk.Mode != "" {
			diskArgs.Mode = pulumi.StringPtr(disk.Mode)
		}
		if disk.KmsKey.GetValue() != "" {
			diskArgs.KmsKeySelfLink = pulumi.StringPtr(disk.KmsKey.GetValue())
		}
		if disk.KmsKeyServiceAccount != "" {
			diskArgs.DiskEncryptionServiceAccount = pulumi.StringPtr(disk.KmsKeyServiceAccount)
		}
		// Regional-disk takeover; forcing a zonal disk is an API error.
		if disk.ForceAttach {
			diskArgs.ForceAttach = pulumi.BoolPtr(true)
		}
		attachedDisks = append(attachedDisks, diskArgs)
	}

	// Ephemeral local SSDs — contents vanish on stop/preemption.
	scratchDisks := compute.InstanceScratchDiskArray{}
	for _, disk := range spec.ScratchDisks {
		diskArgs := &compute.InstanceScratchDiskArgs{
			Interface: pulumi.String(disk.Interface),
		}
		if disk.SizeGb > 0 {
			diskArgs.Size = pulumi.IntPtr(int(disk.SizeGb))
		}
		if disk.DeviceName != "" {
			diskArgs.DeviceName = pulumi.StringPtr(disk.DeviceName)
		}
		scratchDisks = append(scratchDisks, diskArgs)
	}

	networkInterfaces := compute.InstanceNetworkInterfaceArray{}
	for _, ni := range spec.NetworkInterfaces {
		niArgs := &compute.InstanceNetworkInterfaceArgs{}
		if ni.Network.GetValue() != "" {
			niArgs.Network = pulumi.StringPtr(ni.Network.GetValue())
		}
		if ni.Subnetwork.GetValue() != "" {
			niArgs.Subnetwork = pulumi.StringPtr(ni.Subnetwork.GetValue())
		}
		if ni.SubnetworkProject != "" {
			niArgs.SubnetworkProject = pulumi.StringPtr(ni.SubnetworkProject)
		}
		// PSC consumer interface — legal on its own with no
		// network/subnetwork (the spec's CEL mirrors that rule).
		if ni.NetworkAttachment != "" {
			niArgs.NetworkAttachment = pulumi.StringPtr(ni.NetworkAttachment)
		}
		if ni.NetworkIp.GetValue() != "" {
			niArgs.NetworkIp = pulumi.StringPtr(ni.NetworkIp.GetValue())
		}
		if ni.StackType != "" {
			niArgs.StackType = pulumi.StringPtr(ni.StackType)
		}
		if ni.NicType != "" {
			niArgs.NicType = pulumi.StringPtr(ni.NicType)
		}
		if ni.QueueCount != nil {
			niArgs.QueueCount = pulumi.IntPtr(int(ni.GetQueueCount()))
		}
		// VLAN tag marks a dynamic sub-interface (2-255).
		if ni.Vlan != nil {
			niArgs.Vlan = pulumi.IntPtr(int(ni.GetVlan()))
		}
		if ni.IgmpQuery != "" {
			niArgs.IgmpQuery = pulumi.StringPtr(ni.IgmpQuery)
		}
		// Static internal IPv6 — requires an IPv6-enabled stack_type and
		// subnetwork; unset lets GCP assign from the subnetwork range.
		if ni.Ipv6Address != "" {
			niArgs.Ipv6Address = pulumi.StringPtr(ni.Ipv6Address)
		}
		if ni.InternalIpv6PrefixLength != nil {
			niArgs.InternalIpv6PrefixLength = pulumi.IntPtr(int(ni.GetInternalIpv6PrefixLength()))
		}

		// An access config grants a static (nat_ip) or ephemeral external
		// IPv4 -- the manifest's `ephemeral: true` is the request for the
		// latter and renders as a config with no nat_ip; absence keeps the
		// VM private (pair with Cloud NAT for egress).
		if len(ni.AccessConfigs) > 0 {
			accessConfigs := compute.InstanceNetworkInterfaceAccessConfigArray{}
			for _, ac := range ni.AccessConfigs {
				acArgs := &compute.InstanceNetworkInterfaceAccessConfigArgs{}
				if ac.NatIp.GetValue() != "" {
					acArgs.NatIp = pulumi.StringPtr(ac.NatIp.GetValue())
				}
				if ac.NetworkTier != "" {
					acArgs.NetworkTier = pulumi.StringPtr(ac.NetworkTier)
				}
				if ac.PublicPtrDomainName != "" {
					acArgs.PublicPtrDomainName = pulumi.StringPtr(ac.PublicPtrDomainName)
				}
				accessConfigs = append(accessConfigs, acArgs)
			}
			niArgs.AccessConfigs = accessConfigs
		}

		if len(ni.Ipv6AccessConfigs) > 0 {
			ipv6Configs := compute.InstanceNetworkInterfaceIpv6AccessConfigArray{}
			for _, ac := range ni.Ipv6AccessConfigs {
				acArgs := &compute.InstanceNetworkInterfaceIpv6AccessConfigArgs{
					NetworkTier: pulumi.String(ac.NetworkTier),
				}
				if ac.PublicPtrDomainName != "" {
					acArgs.PublicPtrDomainName = pulumi.StringPtr(ac.PublicPtrDomainName)
				}
				// Unset lets GCP assign the external range; these three
				// are immutable — pinning or renaming replaces the VM.
				if ac.ExternalIpv6 != "" {
					acArgs.ExternalIpv6 = pulumi.StringPtr(ac.ExternalIpv6)
				}
				if ac.ExternalIpv6PrefixLength != "" {
					acArgs.ExternalIpv6PrefixLength = pulumi.StringPtr(ac.ExternalIpv6PrefixLength)
				}
				if ac.Name != "" {
					acArgs.Name = pulumi.StringPtr(ac.Name)
				}
				ipv6Configs = append(ipv6Configs, acArgs)
			}
			niArgs.Ipv6AccessConfigs = ipv6Configs
		}

		if len(ni.AliasIpRanges) > 0 {
			aliasRanges := compute.InstanceNetworkInterfaceAliasIpRangeArray{}
			for _, ar := range ni.AliasIpRanges {
				arArgs := &compute.InstanceNetworkInterfaceAliasIpRangeArgs{
					IpCidrRange: pulumi.String(ar.IpCidrRange),
				}
				if ar.SubnetworkRangeName != "" {
					arArgs.SubnetworkRangeName = pulumi.StringPtr(ar.SubnetworkRangeName)
				}
				aliasRanges = append(aliasRanges, arArgs)
			}
			niArgs.AliasIpRanges = aliasRanges
		}

		networkInterfaces = append(networkInterfaces, niArgs)
	}

	// SSH keys fold into the metadata "ssh-keys" key (newline-joined) —
	// byte-identical to the Terraform module's fold. The startup script
	// rides the dedicated MetadataStartupScript attribute, never plain
	// metadata, so it re-runs on every boot exactly as GCP documents.
	finalMetadata := map[string]string{}
	for key, value := range spec.Metadata {
		finalMetadata[key] = value
	}
	if len(spec.SshKeys) > 0 {
		sshKeys := spec.SshKeys[0]
		for _, key := range spec.SshKeys[1:] {
			sshKeys = sshKeys + "\n" + key
		}
		finalMetadata["ssh-keys"] = sshKeys
	}

	args := &compute.InstanceArgs{
		Name:        pulumi.String(locals.InstanceName),
		Zone:        pulumi.String(spec.Zone),
		MachineType: pulumi.String(spec.MachineType),
		Labels:      pulumi.ToStringMap(locals.GcpLabels),
		BootDisk:    bootDiskArgs,
	}

	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.Hostname != "" {
		args.Hostname = pulumi.StringPtr(spec.Hostname)
	}
	if len(spec.Tags) > 0 {
		args.Tags = pulumi.ToStringArray(spec.Tags)
	}
	if spec.MinCpuPlatform != "" {
		args.MinCpuPlatform = pulumi.StringPtr(spec.MinCpuPlatform)
	}
	if spec.CanIpForward {
		args.CanIpForward = pulumi.BoolPtr(true)
	}
	if spec.EnableDisplay {
		args.EnableDisplay = pulumi.BoolPtr(true)
	}
	args.DeletionProtection = pulumi.BoolPtr(spec.DeletionProtection)
	if spec.DesiredStatus != "" {
		args.DesiredStatus = pulumi.StringPtr(spec.DesiredStatus)
	}
	if spec.AllowStoppingForUpdate != nil {
		args.AllowStoppingForUpdate = pulumi.BoolPtr(spec.GetAllowStoppingForUpdate())
	}
	if spec.KeyRevocationActionType != "" {
		args.KeyRevocationActionType = pulumi.StringPtr(spec.KeyRevocationActionType)
	}
	// The bridged provider flattens this max-1 list to a single string;
	// the spec's repeated shape (capped at 1 by validation) mirrors the
	// provider API.
	if len(spec.ResourcePolicies) > 0 {
		args.ResourcePolicies = pulumi.StringPtr(spec.ResourcePolicies[0])
	}
	if len(finalMetadata) > 0 {
		args.Metadata = pulumi.ToStringMap(finalMetadata)
	}
	if spec.StartupScript != "" {
		args.MetadataStartupScript = pulumi.StringPtr(spec.StartupScript)
	}
	if len(attachedDisks) > 0 {
		args.AttachedDisks = attachedDisks
	}
	if len(scratchDisks) > 0 {
		args.ScratchDisks = scratchDisks
	}
	args.NetworkInterfaces = networkInterfaces

	// Omitted block = Compute Engine default service account with its
	// default scopes; an explicit block with no email pins scopes on the
	// default account.
	if spec.ServiceAccount != nil {
		saArgs := &compute.InstanceServiceAccountArgs{
			Scopes: pulumi.ToStringArray(spec.ServiceAccount.Scopes),
		}
		if spec.ServiceAccount.Email.GetValue() != "" {
			saArgs.Email = pulumi.StringPtr(spec.ServiceAccount.Email.GetValue())
		}
		args.ServiceAccount = saArgs
	}

	// Spot semantics: SPOT requires the API's legacy preemptible flag and
	// forbids automatic restart. Deriving both here (identically in the
	// Terraform module) keeps the spec's single provisioning_model switch
	// honest.
	isSpot := spec.Scheduling != nil && spec.Scheduling.ProvisioningModel == "SPOT"
	if spec.Scheduling != nil {
		schedulingArgs := &compute.InstanceSchedulingArgs{
			Preemptible: pulumi.BoolPtr(isSpot),
		}
		if spec.Scheduling.ProvisioningModel != "" {
			schedulingArgs.ProvisioningModel = pulumi.StringPtr(spec.Scheduling.ProvisioningModel)
		}
		automaticRestart := true
		if isSpot {
			automaticRestart = false
		} else if spec.Scheduling.AutomaticRestart != nil {
			automaticRestart = spec.Scheduling.GetAutomaticRestart()
		}
		schedulingArgs.AutomaticRestart = pulumi.BoolPtr(automaticRestart)
		if spec.Scheduling.OnHostMaintenance != "" {
			schedulingArgs.OnHostMaintenance = pulumi.StringPtr(spec.Scheduling.OnHostMaintenance)
		}
		if spec.Scheduling.InstanceTerminationAction != "" {
			schedulingArgs.InstanceTerminationAction = pulumi.StringPtr(spec.Scheduling.InstanceTerminationAction)
		}
		if spec.Scheduling.TerminationTime != "" {
			schedulingArgs.TerminationTime = pulumi.StringPtr(spec.Scheduling.TerminationTime)
		}
		if spec.Scheduling.MaxRunDurationSeconds != nil {
			schedulingArgs.MaxRunDuration = &compute.InstanceSchedulingMaxRunDurationArgs{
				Seconds: pulumi.Int(int(spec.Scheduling.GetMaxRunDurationSeconds())),
			}
		}
		if spec.Scheduling.DiscardLocalSsdsOnStop != nil {
			schedulingArgs.OnInstanceStopAction = &compute.InstanceSchedulingOnInstanceStopActionArgs{
				DiscardLocalSsd: pulumi.BoolPtr(spec.Scheduling.GetDiscardLocalSsdsOnStop()),
			}
		}
		if spec.Scheduling.AvailabilityDomain != nil {
			schedulingArgs.AvailabilityDomain = pulumi.IntPtr(int(spec.Scheduling.GetAvailabilityDomain()))
		}
		if spec.Scheduling.MinNodeCpus != nil {
			schedulingArgs.MinNodeCpus = pulumi.IntPtr(int(spec.Scheduling.GetMinNodeCpus()))
		}
		if len(spec.Scheduling.NodeAffinities) > 0 {
			affinities := compute.InstanceSchedulingNodeAffinityArray{}
			for _, na := range spec.Scheduling.NodeAffinities {
				affinities = append(affinities, &compute.InstanceSchedulingNodeAffinityArgs{
					Key:      pulumi.String(na.Key),
					Operator: pulumi.String(na.Operator),
					Values:   pulumi.ToStringArray(na.Values),
				})
			}
			schedulingArgs.NodeAffinities = affinities
		}
		if spec.Scheduling.LocalSsdRecoveryTimeoutSeconds != nil {
			schedulingArgs.LocalSsdRecoveryTimeout = &compute.InstanceSchedulingLocalSsdRecoveryTimeoutArgs{
				Seconds: pulumi.Int(int(spec.Scheduling.GetLocalSsdRecoveryTimeoutSeconds())),
			}
		}
		// Host-error detection timeout (90..330 s, steps of 30). Sent only
		// when set so an unset spec keeps Compute Engine's default recovery.
		if spec.Scheduling.HostErrorTimeoutSeconds != nil {
			schedulingArgs.HostErrorTimeoutSeconds = pulumi.IntPtr(int(spec.Scheduling.GetHostErrorTimeoutSeconds()))
		}
		args.Scheduling = schedulingArgs
	}

	// Managed workload identity: a SPIFFE ID issued to the VM, optionally
	// with X.509 identity certificates for mutual TLS. Immutable (replaces
	// the VM on change).
	if spec.WorkloadIdentityConfig != nil {
		args.WorkloadIdentityConfig = &compute.InstanceWorkloadIdentityConfigArgs{
			Identity:                   pulumi.StringPtr(spec.WorkloadIdentityConfig.Identity),
			IdentityCertificateEnabled: pulumi.BoolPtr(spec.WorkloadIdentityConfig.IdentityCertificateEnabled),
		}
	}

	// Shielded VM: unset booleans follow GCP defaults (secure boot off,
	// vTPM on, integrity monitoring on).
	if spec.ShieldedInstanceConfig != nil {
		shieldedArgs := &compute.InstanceShieldedInstanceConfigArgs{
			EnableSecureBoot:          pulumi.BoolPtr(false),
			EnableVtpm:                pulumi.BoolPtr(true),
			EnableIntegrityMonitoring: pulumi.BoolPtr(true),
		}
		if spec.ShieldedInstanceConfig.EnableSecureBoot != nil {
			shieldedArgs.EnableSecureBoot = pulumi.BoolPtr(spec.ShieldedInstanceConfig.GetEnableSecureBoot())
		}
		if spec.ShieldedInstanceConfig.EnableVtpm != nil {
			shieldedArgs.EnableVtpm = pulumi.BoolPtr(spec.ShieldedInstanceConfig.GetEnableVtpm())
		}
		if spec.ShieldedInstanceConfig.EnableIntegrityMonitoring != nil {
			shieldedArgs.EnableIntegrityMonitoring = pulumi.BoolPtr(spec.ShieldedInstanceConfig.GetEnableIntegrityMonitoring())
		}
		args.ShieldedInstanceConfig = shieldedArgs
	}

	// Confidential VM: the typed field is the modern surface; the legacy
	// enable flag stays unset (it only supports SEV and will be
	// deprecated).
	if spec.ConfidentialInstanceConfig != nil {
		args.ConfidentialInstanceConfig = &compute.InstanceConfidentialInstanceConfigArgs{
			ConfidentialInstanceType: pulumi.StringPtr(spec.ConfidentialInstanceConfig.ConfidentialInstanceType),
		}
	}

	if spec.AdvancedMachineFeatures != nil {
		amf := spec.AdvancedMachineFeatures
		amfArgs := &compute.InstanceAdvancedMachineFeaturesArgs{}
		if amf.EnableNestedVirtualization != nil {
			amfArgs.EnableNestedVirtualization = pulumi.BoolPtr(amf.GetEnableNestedVirtualization())
		}
		if amf.ThreadsPerCore != nil {
			amfArgs.ThreadsPerCore = pulumi.IntPtr(int(amf.GetThreadsPerCore()))
		}
		if amf.VisibleCoreCount != nil {
			amfArgs.VisibleCoreCount = pulumi.IntPtr(int(amf.GetVisibleCoreCount()))
		}
		if amf.EnableUefiNetworking != nil {
			amfArgs.EnableUefiNetworking = pulumi.BoolPtr(amf.GetEnableUefiNetworking())
		}
		if amf.PerformanceMonitoringUnit != "" {
			amfArgs.PerformanceMonitoringUnit = pulumi.StringPtr(amf.PerformanceMonitoringUnit)
		}
		if amf.TurboMode != "" {
			amfArgs.TurboMode = pulumi.StringPtr(amf.TurboMode)
		}
		args.AdvancedMachineFeatures = amfArgs
	}

	if len(spec.GuestAccelerators) > 0 {
		accelerators := compute.InstanceGuestAcceleratorArray{}
		for _, ga := range spec.GuestAccelerators {
			accelerators = append(accelerators, &compute.InstanceGuestAcceleratorArgs{
				Type:  pulumi.String(ga.Type),
				Count: pulumi.Int(int(ga.Count)),
			})
		}
		args.GuestAccelerators = accelerators
	}

	if spec.ReservationAffinity != nil {
		raArgs := &compute.InstanceReservationAffinityArgs{
			Type: pulumi.String(spec.ReservationAffinity.Type),
		}
		if spec.ReservationAffinity.SpecificReservation != nil {
			raArgs.SpecificReservation = &compute.InstanceReservationAffinitySpecificReservationArgs{
				Key:    pulumi.String(spec.ReservationAffinity.SpecificReservation.Key),
				Values: pulumi.ToStringArray(spec.ReservationAffinity.SpecificReservation.Values),
			}
		}
		args.ReservationAffinity = raArgs
	}

	if spec.TotalEgressBandwidthTier != "" {
		args.NetworkPerformanceConfig = &compute.InstanceNetworkPerformanceConfigArgs{
			TotalEgressBandwidthTier: pulumi.String(spec.TotalEgressBandwidthTier),
		}
	}

	// Resource Manager tags bind at create time only.
	if len(spec.ResourceManagerTags) > 0 {
		args.Params = &compute.InstanceParamsArgs{
			ResourceManagerTags: pulumi.ToStringMap(spec.ResourceManagerTags),
		}
	}

	// Instance-level CMEK (memory and other instance state) — distinct
	// from the per-disk keys. CSEK raw keys are deliberately not modeled.
	if spec.InstanceEncryptionKey != nil {
		keyArgs := &compute.InstanceInstanceEncryptionKeyArgs{
			KmsKeySelfLink: pulumi.StringPtr(spec.InstanceEncryptionKey.KmsKey.GetValue()),
		}
		if spec.InstanceEncryptionKey.KmsKeyServiceAccount != "" {
			keyArgs.KmsKeyServiceAccount = pulumi.StringPtr(spec.InstanceEncryptionKey.KmsKeyServiceAccount)
		}
		args.InstanceEncryptionKey = keyArgs
	}

	// Destroy behavior: unset follows the provider default (DELETE);
	// PREVENT fails the destroy; ABANDON forgets the VM but leaves it
	// running in GCP.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdInstance, err := compute.NewInstance(ctx,
		locals.InstanceName,
		args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdComputeApi}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create compute instance")
	}

	return createdInstance, nil
}
