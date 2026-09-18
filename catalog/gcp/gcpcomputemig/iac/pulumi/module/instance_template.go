package module

import (
	"github.com/pkg/errors"
	gcpcomputemigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcomputemig/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// templateResult carries what downstream resources need from the created
// template.
type templateResult struct {
	// TemplateRef is what the group manager's version block references:
	// the zonal template's self_link_unique (carries the ?uniqueId= the
	// provider's diff suppression keys rotation on), or the regional
	// template's self_link (the regional resource has no unique variant —
	// rotation is carried by the name_prefix change producing a new
	// link).
	TemplateRef pulumi.StringOutput
}

// instanceTemplate creates the group's instance template — zonal
// (google_compute_instance_template, a global resource) or regional per
// the spec's location selector.
//
// Templates are IMMUTABLE in GCP (labels excepted): the module manages
// rotation natively via name_prefix — every template change creates a
// fresh "<mig-name>-<timestamp>" template first, repoints the group
// manager, then deletes the old template (Pulumi's default
// create-before-delete replacement), so the group is never left pointing
// at a deleted template.
//
// Spot semantics: provisioning_model = "SPOT" requires the API's legacy
// preemptible flag and no automatic restart — both derived here,
// identically to the Terraform module, so the spec's single switch stays
// honest.
func instanceTemplate(
	ctx *pulumi.Context,
	locals *Locals,
	gcpProvider *gcp.Provider,
	dependsOn pulumi.Resource,
) (*templateResult, error) {

	spec := locals.GcpComputeMig.Spec
	template := spec.Template

	if locals.IsRegional {
		args := &compute.RegionInstanceTemplateArgs{
			NamePrefix:  pulumi.StringPtr(locals.TemplateNamePrefix),
			Region:      pulumi.String(spec.Region),
			MachineType: pulumi.String(template.MachineType),
			Labels:      pulumi.ToStringMap(locals.GcpLabels),
			Disks:       regionalTemplateDisks(template),
		}
		if spec.ProjectId.GetValue() != "" {
			args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
		}
		if template.Description != "" {
			args.Description = pulumi.StringPtr(template.Description)
		}
		if template.InstanceDescription != "" {
			args.InstanceDescription = pulumi.StringPtr(template.InstanceDescription)
		}
		args.NetworkInterfaces = regionalTemplateNetworkInterfaces(template)
		if template.ServiceAccount != nil {
			saArgs := &compute.RegionInstanceTemplateServiceAccountArgs{
				Scopes: pulumi.ToStringArray(template.ServiceAccount.Scopes),
			}
			if template.ServiceAccount.Email.GetValue() != "" {
				saArgs.Email = pulumi.StringPtr(template.ServiceAccount.Email.GetValue())
			}
			args.ServiceAccount = saArgs
		}
		if template.Scheduling != nil {
			args.Scheduling = regionalTemplateScheduling(template.Scheduling)
		}
		if template.ShieldedInstanceConfig != nil {
			shielded := &compute.RegionInstanceTemplateShieldedInstanceConfigArgs{
				EnableSecureBoot:          pulumi.BoolPtr(false),
				EnableVtpm:                pulumi.BoolPtr(true),
				EnableIntegrityMonitoring: pulumi.BoolPtr(true),
			}
			if template.ShieldedInstanceConfig.EnableSecureBoot != nil {
				shielded.EnableSecureBoot = pulumi.BoolPtr(template.ShieldedInstanceConfig.GetEnableSecureBoot())
			}
			if template.ShieldedInstanceConfig.EnableVtpm != nil {
				shielded.EnableVtpm = pulumi.BoolPtr(template.ShieldedInstanceConfig.GetEnableVtpm())
			}
			if template.ShieldedInstanceConfig.EnableIntegrityMonitoring != nil {
				shielded.EnableIntegrityMonitoring = pulumi.BoolPtr(template.ShieldedInstanceConfig.GetEnableIntegrityMonitoring())
			}
			args.ShieldedInstanceConfig = shielded
		}
		// The typed field is the modern surface; the legacy enable flag
		// stays unset (SEV-only, headed for deprecation).
		if template.ConfidentialInstanceConfig != nil {
			args.ConfidentialInstanceConfig = &compute.RegionInstanceTemplateConfidentialInstanceConfigArgs{
				ConfidentialInstanceType: pulumi.StringPtr(template.ConfidentialInstanceConfig.ConfidentialInstanceType),
			}
		}
		if template.AdvancedMachineFeatures != nil {
			amf := template.AdvancedMachineFeatures
			amfArgs := &compute.RegionInstanceTemplateAdvancedMachineFeaturesArgs{}
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
		if len(template.GuestAccelerators) > 0 {
			accelerators := compute.RegionInstanceTemplateGuestAcceleratorArray{}
			for _, ga := range template.GuestAccelerators {
				accelerators = append(accelerators, &compute.RegionInstanceTemplateGuestAcceleratorArgs{
					Type:  pulumi.String(ga.Type),
					Count: pulumi.Int(int(ga.Count)),
				})
			}
			args.GuestAccelerators = accelerators
		}
		if template.ReservationAffinity != nil {
			ra := &compute.RegionInstanceTemplateReservationAffinityArgs{
				Type: pulumi.String(template.ReservationAffinity.Type),
			}
			if template.ReservationAffinity.SpecificReservation != nil {
				ra.SpecificReservation = &compute.RegionInstanceTemplateReservationAffinitySpecificReservationArgs{
					Key:    pulumi.String(template.ReservationAffinity.SpecificReservation.Key),
					Values: pulumi.ToStringArray(template.ReservationAffinity.SpecificReservation.Values),
				}
			}
			args.ReservationAffinity = ra
		}
		if template.TotalEgressBandwidthTier != "" {
			args.NetworkPerformanceConfig = &compute.RegionInstanceTemplateNetworkPerformanceConfigArgs{
				TotalEgressBandwidthTier: pulumi.String(template.TotalEgressBandwidthTier),
			}
		}
		if len(template.Metadata) > 0 {
			args.Metadata = pulumi.ToStringMap(template.Metadata)
		}
		if template.StartupScript != "" {
			args.MetadataStartupScript = pulumi.StringPtr(template.StartupScript)
		}
		if len(template.Tags) > 0 {
			args.Tags = pulumi.ToStringArray(template.Tags)
		}
		if len(template.ResourceManagerTags) > 0 {
			args.ResourceManagerTags = pulumi.ToStringMap(template.ResourceManagerTags)
		}
		if template.MinCpuPlatform != "" {
			args.MinCpuPlatform = pulumi.StringPtr(template.MinCpuPlatform)
		}
		if template.CanIpForward {
			args.CanIpForward = pulumi.BoolPtr(true)
		}
		if template.KeyRevocationActionType != "" {
			args.KeyRevocationActionType = pulumi.StringPtr(template.KeyRevocationActionType)
		}
		// The bridged provider flattens this max-1 list to a single
		// string; the spec's repeated shape (capped at 1 by validation)
		// mirrors the provider API.
		if len(template.ResourcePolicies) > 0 {
			args.ResourcePolicies = pulumi.StringPtr(template.ResourcePolicies[0])
		}
		// Managed workload identity: a SPIFFE ID issued to each VM,
		// optionally with X.509 identity certificates for mutual TLS. A
		// template field, so changing it rotates the template.
		if template.WorkloadIdentityConfig != nil {
			args.WorkloadIdentityConfig = &compute.RegionInstanceTemplateWorkloadIdentityConfigArgs{
				Identity:                   pulumi.StringPtr(template.WorkloadIdentityConfig.Identity),
				IdentityCertificateEnabled: pulumi.BoolPtr(template.WorkloadIdentityConfig.IdentityCertificateEnabled),
			}
		}
		// Destroy behavior: only the REGIONAL template carries a
		// deletion_policy in the provider — the zonal one has none (it is
		// always deleted on destroy). The spec's single field lands here
		// honestly.
		if spec.DeletionPolicy != "" {
			args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
		}

		createdTemplate, err := compute.NewRegionInstanceTemplate(ctx,
			locals.MigName,
			args,
			pulumi.Provider(gcpProvider),
			pulumi.DependsOn([]pulumi.Resource{dependsOn}),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create regional instance template")
		}
		return &templateResult{TemplateRef: createdTemplate.SelfLink}, nil
	}

	// Zonal groups use the global instance template resource, whose
	// self_link_unique carries the ?uniqueId= that keys the group
	// manager's rotation-aware diff suppression.
	args := &compute.InstanceTemplateArgs{
		NamePrefix:  pulumi.StringPtr(locals.TemplateNamePrefix),
		MachineType: pulumi.String(template.MachineType),
		Labels:      pulumi.ToStringMap(locals.GcpLabels),
		Disks:       zonalTemplateDisks(template),
	}
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}
	if template.Description != "" {
		args.Description = pulumi.StringPtr(template.Description)
	}
	if template.InstanceDescription != "" {
		args.InstanceDescription = pulumi.StringPtr(template.InstanceDescription)
	}
	args.NetworkInterfaces = zonalTemplateNetworkInterfaces(template)
	if template.ServiceAccount != nil {
		saArgs := &compute.InstanceTemplateServiceAccountArgs{
			Scopes: pulumi.ToStringArray(template.ServiceAccount.Scopes),
		}
		if template.ServiceAccount.Email.GetValue() != "" {
			saArgs.Email = pulumi.StringPtr(template.ServiceAccount.Email.GetValue())
		}
		args.ServiceAccount = saArgs
	}
	if template.Scheduling != nil {
		args.Scheduling = zonalTemplateScheduling(template.Scheduling)
	}
	if template.ShieldedInstanceConfig != nil {
		shielded := &compute.InstanceTemplateShieldedInstanceConfigArgs{
			EnableSecureBoot:          pulumi.BoolPtr(false),
			EnableVtpm:                pulumi.BoolPtr(true),
			EnableIntegrityMonitoring: pulumi.BoolPtr(true),
		}
		if template.ShieldedInstanceConfig.EnableSecureBoot != nil {
			shielded.EnableSecureBoot = pulumi.BoolPtr(template.ShieldedInstanceConfig.GetEnableSecureBoot())
		}
		if template.ShieldedInstanceConfig.EnableVtpm != nil {
			shielded.EnableVtpm = pulumi.BoolPtr(template.ShieldedInstanceConfig.GetEnableVtpm())
		}
		if template.ShieldedInstanceConfig.EnableIntegrityMonitoring != nil {
			shielded.EnableIntegrityMonitoring = pulumi.BoolPtr(template.ShieldedInstanceConfig.GetEnableIntegrityMonitoring())
		}
		args.ShieldedInstanceConfig = shielded
	}
	if template.ConfidentialInstanceConfig != nil {
		args.ConfidentialInstanceConfig = &compute.InstanceTemplateConfidentialInstanceConfigArgs{
			ConfidentialInstanceType: pulumi.StringPtr(template.ConfidentialInstanceConfig.ConfidentialInstanceType),
		}
	}
	if template.AdvancedMachineFeatures != nil {
		amf := template.AdvancedMachineFeatures
		amfArgs := &compute.InstanceTemplateAdvancedMachineFeaturesArgs{}
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
	if len(template.GuestAccelerators) > 0 {
		accelerators := compute.InstanceTemplateGuestAcceleratorArray{}
		for _, ga := range template.GuestAccelerators {
			accelerators = append(accelerators, &compute.InstanceTemplateGuestAcceleratorArgs{
				Type:  pulumi.String(ga.Type),
				Count: pulumi.Int(int(ga.Count)),
			})
		}
		args.GuestAccelerators = accelerators
	}
	if template.ReservationAffinity != nil {
		ra := &compute.InstanceTemplateReservationAffinityArgs{
			Type: pulumi.String(template.ReservationAffinity.Type),
		}
		if template.ReservationAffinity.SpecificReservation != nil {
			ra.SpecificReservation = &compute.InstanceTemplateReservationAffinitySpecificReservationArgs{
				Key:    pulumi.String(template.ReservationAffinity.SpecificReservation.Key),
				Values: pulumi.ToStringArray(template.ReservationAffinity.SpecificReservation.Values),
			}
		}
		args.ReservationAffinity = ra
	}
	if template.TotalEgressBandwidthTier != "" {
		args.NetworkPerformanceConfig = &compute.InstanceTemplateNetworkPerformanceConfigArgs{
			TotalEgressBandwidthTier: pulumi.String(template.TotalEgressBandwidthTier),
		}
	}
	if len(template.Metadata) > 0 {
		args.Metadata = pulumi.ToStringMap(template.Metadata)
	}
	if template.StartupScript != "" {
		args.MetadataStartupScript = pulumi.StringPtr(template.StartupScript)
	}
	if len(template.Tags) > 0 {
		args.Tags = pulumi.ToStringArray(template.Tags)
	}
	if len(template.ResourceManagerTags) > 0 {
		args.ResourceManagerTags = pulumi.ToStringMap(template.ResourceManagerTags)
	}
	if template.MinCpuPlatform != "" {
		args.MinCpuPlatform = pulumi.StringPtr(template.MinCpuPlatform)
	}
	if template.CanIpForward {
		args.CanIpForward = pulumi.BoolPtr(true)
	}
	if template.KeyRevocationActionType != "" {
		args.KeyRevocationActionType = pulumi.StringPtr(template.KeyRevocationActionType)
	}
	if len(template.ResourcePolicies) > 0 {
		args.ResourcePolicies = pulumi.StringPtr(template.ResourcePolicies[0])
	}
	// Managed workload identity (see the regional builder above).
	if template.WorkloadIdentityConfig != nil {
		args.WorkloadIdentityConfig = &compute.InstanceTemplateWorkloadIdentityConfigArgs{
			Identity:                   pulumi.StringPtr(template.WorkloadIdentityConfig.Identity),
			IdentityCertificateEnabled: pulumi.BoolPtr(template.WorkloadIdentityConfig.IdentityCertificateEnabled),
		}
	}
	createdTemplate, err := compute.NewInstanceTemplate(ctx,
		locals.MigName,
		args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{dependsOn}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create instance template")
	}
	return &templateResult{TemplateRef: createdTemplate.SelfLinkUnique}, nil
}

// zonalTemplateDisks maps the spec's disks to the zonal template's flat
// disk blocks. The template disk shape IS flat (role fields on the disk
// itself) — deliberately unlike the instance resource's
// boot/attached/scratch split.
func zonalTemplateDisks(template *gcpcomputemigv1alpha1.GcpComputeMigTemplate) compute.InstanceTemplateDiskArray {
	disks := compute.InstanceTemplateDiskArray{}
	for _, disk := range template.Disks {
		diskArgs := &compute.InstanceTemplateDiskArgs{}
		if disk.Boot {
			diskArgs.Boot = pulumi.BoolPtr(true)
		}
		if disk.SourceImage != "" {
			diskArgs.SourceImage = pulumi.StringPtr(disk.SourceImage)
		}
		if disk.SourceSnapshot != "" {
			diskArgs.SourceSnapshot = pulumi.StringPtr(disk.SourceSnapshot)
		}
		if disk.Source.GetValue() != "" {
			diskArgs.Source = pulumi.StringPtr(disk.Source.GetValue())
		}
		if disk.SizeGb > 0 {
			diskArgs.DiskSizeGb = pulumi.IntPtr(int(disk.SizeGb))
		}
		if disk.DiskType != "" {
			diskArgs.DiskType = pulumi.StringPtr(disk.DiskType)
		}
		if disk.Type != "" {
			diskArgs.Type = pulumi.StringPtr(disk.Type)
		}
		// Defaults true (matching GCP); explicit-send so a stated false
		// always reaches the engine.
		if disk.AutoDelete != nil {
			diskArgs.AutoDelete = pulumi.BoolPtr(disk.GetAutoDelete())
		} else {
			diskArgs.AutoDelete = pulumi.BoolPtr(true)
		}
		if disk.DeviceName != "" {
			diskArgs.DeviceName = pulumi.StringPtr(disk.DeviceName)
		}
		if disk.DiskName != "" {
			diskArgs.DiskName = pulumi.StringPtr(disk.DiskName)
		}
		if disk.Mode != "" {
			diskArgs.Mode = pulumi.StringPtr(disk.Mode)
		}
		// Google-advice-only lever — sent only when explicitly set.
		if disk.Interface != "" {
			diskArgs.Interface = pulumi.StringPtr(disk.Interface)
		}
		if len(disk.DiskLabels) > 0 {
			diskArgs.Labels = pulumi.ToStringMap(disk.DiskLabels)
		}
		if disk.ProvisionedIops != nil {
			diskArgs.ProvisionedIops = pulumi.IntPtr(int(disk.GetProvisionedIops()))
		}
		if disk.ProvisionedThroughput != nil {
			diskArgs.ProvisionedThroughput = pulumi.IntPtr(int(disk.GetProvisionedThroughput()))
		}
		if disk.Architecture != "" {
			diskArgs.Architecture = pulumi.StringPtr(disk.Architecture)
		}
		if len(disk.GuestOsFeatures) > 0 {
			diskArgs.GuestOsFeatures = pulumi.ToStringArray(disk.GuestOsFeatures)
		}
		// The bridged provider flattens this max-1 list to a single
		// string; the spec's repeated shape (capped at 1) mirrors the
		// provider API.
		if len(disk.ResourcePolicies) > 0 {
			diskArgs.ResourcePolicies = pulumi.StringPtr(disk.ResourcePolicies[0])
		}
		if len(disk.ResourceManagerTags) > 0 {
			diskArgs.ResourceManagerTags = pulumi.ToStringMap(disk.ResourceManagerTags)
		}
		if disk.StoragePool != "" {
			diskArgs.StoragePool = pulumi.StringPtr(disk.StoragePool)
		}
		// CMEK only — CSEK raw-key arms are deliberately not modeled
		// (secure-by-default).
		if disk.DiskEncryption != nil {
			keyArgs := &compute.InstanceTemplateDiskDiskEncryptionKeyArgs{
				KmsKeySelfLink: pulumi.StringPtr(disk.DiskEncryption.KmsKey.GetValue()),
			}
			if disk.DiskEncryption.KmsKeyServiceAccount != "" {
				keyArgs.KmsKeyServiceAccount = pulumi.StringPtr(disk.DiskEncryption.KmsKeyServiceAccount)
			}
			diskArgs.DiskEncryptionKey = keyArgs
		}
		if disk.SourceImageEncryption != nil {
			keyArgs := &compute.InstanceTemplateDiskSourceImageEncryptionKeyArgs{
				KmsKeySelfLink: pulumi.String(disk.SourceImageEncryption.KmsKey.GetValue()),
			}
			if disk.SourceImageEncryption.KmsKeyServiceAccount != "" {
				keyArgs.KmsKeyServiceAccount = pulumi.StringPtr(disk.SourceImageEncryption.KmsKeyServiceAccount)
			}
			diskArgs.SourceImageEncryptionKey = keyArgs
		}
		if disk.SourceSnapshotEncryption != nil {
			keyArgs := &compute.InstanceTemplateDiskSourceSnapshotEncryptionKeyArgs{
				KmsKeySelfLink: pulumi.String(disk.SourceSnapshotEncryption.KmsKey.GetValue()),
			}
			if disk.SourceSnapshotEncryption.KmsKeyServiceAccount != "" {
				keyArgs.KmsKeyServiceAccount = pulumi.StringPtr(disk.SourceSnapshotEncryption.KmsKeyServiceAccount)
			}
			diskArgs.SourceSnapshotEncryptionKey = keyArgs
		}
		disks = append(disks, diskArgs)
	}
	return disks
}

// regionalTemplateDisks mirrors zonalTemplateDisks for the regional
// template's types.
func regionalTemplateDisks(template *gcpcomputemigv1alpha1.GcpComputeMigTemplate) compute.RegionInstanceTemplateDiskArray {
	disks := compute.RegionInstanceTemplateDiskArray{}
	for _, disk := range template.Disks {
		diskArgs := &compute.RegionInstanceTemplateDiskArgs{}
		if disk.Boot {
			diskArgs.Boot = pulumi.BoolPtr(true)
		}
		if disk.SourceImage != "" {
			diskArgs.SourceImage = pulumi.StringPtr(disk.SourceImage)
		}
		if disk.SourceSnapshot != "" {
			diskArgs.SourceSnapshot = pulumi.StringPtr(disk.SourceSnapshot)
		}
		if disk.Source.GetValue() != "" {
			diskArgs.Source = pulumi.StringPtr(disk.Source.GetValue())
		}
		if disk.SizeGb > 0 {
			diskArgs.DiskSizeGb = pulumi.IntPtr(int(disk.SizeGb))
		}
		if disk.DiskType != "" {
			diskArgs.DiskType = pulumi.StringPtr(disk.DiskType)
		}
		if disk.Type != "" {
			diskArgs.Type = pulumi.StringPtr(disk.Type)
		}
		if disk.AutoDelete != nil {
			diskArgs.AutoDelete = pulumi.BoolPtr(disk.GetAutoDelete())
		} else {
			diskArgs.AutoDelete = pulumi.BoolPtr(true)
		}
		if disk.DeviceName != "" {
			diskArgs.DeviceName = pulumi.StringPtr(disk.DeviceName)
		}
		if disk.DiskName != "" {
			diskArgs.DiskName = pulumi.StringPtr(disk.DiskName)
		}
		if disk.Mode != "" {
			diskArgs.Mode = pulumi.StringPtr(disk.Mode)
		}
		if disk.Interface != "" {
			diskArgs.Interface = pulumi.StringPtr(disk.Interface)
		}
		if len(disk.DiskLabels) > 0 {
			diskArgs.Labels = pulumi.ToStringMap(disk.DiskLabels)
		}
		if disk.ProvisionedIops != nil {
			diskArgs.ProvisionedIops = pulumi.IntPtr(int(disk.GetProvisionedIops()))
		}
		if disk.ProvisionedThroughput != nil {
			diskArgs.ProvisionedThroughput = pulumi.IntPtr(int(disk.GetProvisionedThroughput()))
		}
		if disk.Architecture != "" {
			diskArgs.Architecture = pulumi.StringPtr(disk.Architecture)
		}
		if len(disk.GuestOsFeatures) > 0 {
			diskArgs.GuestOsFeatures = pulumi.ToStringArray(disk.GuestOsFeatures)
		}
		if len(disk.ResourcePolicies) > 0 {
			diskArgs.ResourcePolicies = pulumi.StringPtr(disk.ResourcePolicies[0])
		}
		if len(disk.ResourceManagerTags) > 0 {
			diskArgs.ResourceManagerTags = pulumi.ToStringMap(disk.ResourceManagerTags)
		}
		if disk.StoragePool != "" {
			diskArgs.StoragePool = pulumi.StringPtr(disk.StoragePool)
		}
		if disk.DiskEncryption != nil {
			keyArgs := &compute.RegionInstanceTemplateDiskDiskEncryptionKeyArgs{
				KmsKeySelfLink: pulumi.StringPtr(disk.DiskEncryption.KmsKey.GetValue()),
			}
			if disk.DiskEncryption.KmsKeyServiceAccount != "" {
				keyArgs.KmsKeyServiceAccount = pulumi.StringPtr(disk.DiskEncryption.KmsKeyServiceAccount)
			}
			diskArgs.DiskEncryptionKey = keyArgs
		}
		if disk.SourceImageEncryption != nil {
			keyArgs := &compute.RegionInstanceTemplateDiskSourceImageEncryptionKeyArgs{
				KmsKeySelfLink: pulumi.String(disk.SourceImageEncryption.KmsKey.GetValue()),
			}
			if disk.SourceImageEncryption.KmsKeyServiceAccount != "" {
				keyArgs.KmsKeyServiceAccount = pulumi.StringPtr(disk.SourceImageEncryption.KmsKeyServiceAccount)
			}
			diskArgs.SourceImageEncryptionKey = keyArgs
		}
		if disk.SourceSnapshotEncryption != nil {
			keyArgs := &compute.RegionInstanceTemplateDiskSourceSnapshotEncryptionKeyArgs{
				KmsKeySelfLink: pulumi.String(disk.SourceSnapshotEncryption.KmsKey.GetValue()),
			}
			if disk.SourceSnapshotEncryption.KmsKeyServiceAccount != "" {
				keyArgs.KmsKeyServiceAccount = pulumi.StringPtr(disk.SourceSnapshotEncryption.KmsKeyServiceAccount)
			}
			diskArgs.SourceSnapshotEncryptionKey = keyArgs
		}
		disks = append(disks, diskArgs)
	}
	return disks
}

// zonalTemplateNetworkInterfaces maps the spec's NICs to the zonal
// template's types. Absence of access configs keeps the fleet private
// (pair with Cloud NAT for egress).
func zonalTemplateNetworkInterfaces(template *gcpcomputemigv1alpha1.GcpComputeMigTemplate) compute.InstanceTemplateNetworkInterfaceArray {
	interfaces := compute.InstanceTemplateNetworkInterfaceArray{}
	for _, ni := range template.NetworkInterfaces {
		niArgs := &compute.InstanceTemplateNetworkInterfaceArgs{}
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
		if ni.NetworkIp != "" {
			niArgs.NetworkIp = pulumi.StringPtr(ni.NetworkIp)
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
		if ni.Vlan != nil {
			niArgs.Vlan = pulumi.IntPtr(int(ni.GetVlan()))
		}
		if ni.IgmpQuery != "" {
			niArgs.IgmpQuery = pulumi.StringPtr(ni.IgmpQuery)
		}
		if ni.Ipv6Address != "" {
			niArgs.Ipv6Address = pulumi.StringPtr(ni.Ipv6Address)
		}
		if ni.InternalIpv6PrefixLength != nil {
			niArgs.InternalIpv6PrefixLength = pulumi.IntPtr(int(ni.GetInternalIpv6PrefixLength()))
		}
		if len(ni.AccessConfigs) > 0 {
			accessConfigs := compute.InstanceTemplateNetworkInterfaceAccessConfigArray{}
			for _, ac := range ni.AccessConfigs {
				acArgs := &compute.InstanceTemplateNetworkInterfaceAccessConfigArgs{}
				if ac.NatIp != "" {
					acArgs.NatIp = pulumi.StringPtr(ac.NatIp)
				}
				if ac.NetworkTier != "" {
					acArgs.NetworkTier = pulumi.StringPtr(ac.NetworkTier)
				}
				accessConfigs = append(accessConfigs, acArgs)
			}
			niArgs.AccessConfigs = accessConfigs
		}
		if len(ni.Ipv6AccessConfigs) > 0 {
			ipv6Configs := compute.InstanceTemplateNetworkInterfaceIpv6AccessConfigArray{}
			for _, ac := range ni.Ipv6AccessConfigs {
				ipv6Configs = append(ipv6Configs, &compute.InstanceTemplateNetworkInterfaceIpv6AccessConfigArgs{
					NetworkTier: pulumi.String(ac.NetworkTier),
				})
			}
			niArgs.Ipv6AccessConfigs = ipv6Configs
		}
		if len(ni.AliasIpRanges) > 0 {
			aliasRanges := compute.InstanceTemplateNetworkInterfaceAliasIpRangeArray{}
			for _, ar := range ni.AliasIpRanges {
				arArgs := &compute.InstanceTemplateNetworkInterfaceAliasIpRangeArgs{
					IpCidrRange: pulumi.String(ar.IpCidrRange),
				}
				if ar.SubnetworkRangeName != "" {
					arArgs.SubnetworkRangeName = pulumi.StringPtr(ar.SubnetworkRangeName)
				}
				aliasRanges = append(aliasRanges, arArgs)
			}
			niArgs.AliasIpRanges = aliasRanges
		}
		interfaces = append(interfaces, niArgs)
	}
	return interfaces
}

// regionalTemplateNetworkInterfaces mirrors the zonal builder for the
// regional template's types.
func regionalTemplateNetworkInterfaces(template *gcpcomputemigv1alpha1.GcpComputeMigTemplate) compute.RegionInstanceTemplateNetworkInterfaceArray {
	interfaces := compute.RegionInstanceTemplateNetworkInterfaceArray{}
	for _, ni := range template.NetworkInterfaces {
		niArgs := &compute.RegionInstanceTemplateNetworkInterfaceArgs{}
		if ni.Network.GetValue() != "" {
			niArgs.Network = pulumi.StringPtr(ni.Network.GetValue())
		}
		if ni.Subnetwork.GetValue() != "" {
			niArgs.Subnetwork = pulumi.StringPtr(ni.Subnetwork.GetValue())
		}
		if ni.SubnetworkProject != "" {
			niArgs.SubnetworkProject = pulumi.StringPtr(ni.SubnetworkProject)
		}
		if ni.NetworkAttachment != "" {
			niArgs.NetworkAttachment = pulumi.StringPtr(ni.NetworkAttachment)
		}
		if ni.NetworkIp != "" {
			niArgs.NetworkIp = pulumi.StringPtr(ni.NetworkIp)
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
		if ni.Vlan != nil {
			niArgs.Vlan = pulumi.IntPtr(int(ni.GetVlan()))
		}
		if ni.IgmpQuery != "" {
			niArgs.IgmpQuery = pulumi.StringPtr(ni.IgmpQuery)
		}
		if ni.Ipv6Address != "" {
			niArgs.Ipv6Address = pulumi.StringPtr(ni.Ipv6Address)
		}
		if ni.InternalIpv6PrefixLength != nil {
			niArgs.InternalIpv6PrefixLength = pulumi.IntPtr(int(ni.GetInternalIpv6PrefixLength()))
		}
		if len(ni.AccessConfigs) > 0 {
			accessConfigs := compute.RegionInstanceTemplateNetworkInterfaceAccessConfigArray{}
			for _, ac := range ni.AccessConfigs {
				acArgs := &compute.RegionInstanceTemplateNetworkInterfaceAccessConfigArgs{}
				if ac.NatIp != "" {
					acArgs.NatIp = pulumi.StringPtr(ac.NatIp)
				}
				if ac.NetworkTier != "" {
					acArgs.NetworkTier = pulumi.StringPtr(ac.NetworkTier)
				}
				accessConfigs = append(accessConfigs, acArgs)
			}
			niArgs.AccessConfigs = accessConfigs
		}
		if len(ni.Ipv6AccessConfigs) > 0 {
			ipv6Configs := compute.RegionInstanceTemplateNetworkInterfaceIpv6AccessConfigArray{}
			for _, ac := range ni.Ipv6AccessConfigs {
				ipv6Configs = append(ipv6Configs, &compute.RegionInstanceTemplateNetworkInterfaceIpv6AccessConfigArgs{
					NetworkTier: pulumi.String(ac.NetworkTier),
				})
			}
			niArgs.Ipv6AccessConfigs = ipv6Configs
		}
		if len(ni.AliasIpRanges) > 0 {
			aliasRanges := compute.RegionInstanceTemplateNetworkInterfaceAliasIpRangeArray{}
			for _, ar := range ni.AliasIpRanges {
				arArgs := &compute.RegionInstanceTemplateNetworkInterfaceAliasIpRangeArgs{
					IpCidrRange: pulumi.String(ar.IpCidrRange),
				}
				if ar.SubnetworkRangeName != "" {
					arArgs.SubnetworkRangeName = pulumi.StringPtr(ar.SubnetworkRangeName)
				}
				aliasRanges = append(aliasRanges, arArgs)
			}
			niArgs.AliasIpRanges = aliasRanges
		}
		interfaces = append(interfaces, niArgs)
	}
	return interfaces
}

// zonalTemplateScheduling maps the spec's scheduling to the zonal
// template's type. Spot derivations (preemptible flag, no automatic
// restart) are identical to the Terraform module.
func zonalTemplateScheduling(scheduling *gcpcomputemigv1alpha1.GcpComputeMigScheduling) *compute.InstanceTemplateSchedulingArgs {
	isSpot := scheduling.ProvisioningModel == "SPOT"
	args := &compute.InstanceTemplateSchedulingArgs{
		Preemptible: pulumi.BoolPtr(isSpot),
	}
	if scheduling.ProvisioningModel != "" {
		args.ProvisioningModel = pulumi.StringPtr(scheduling.ProvisioningModel)
	}
	automaticRestart := true
	if isSpot || scheduling.ProvisioningModel == "FLEX_START" {
		automaticRestart = false
	} else if scheduling.AutomaticRestart != nil {
		automaticRestart = scheduling.GetAutomaticRestart()
	}
	args.AutomaticRestart = pulumi.BoolPtr(automaticRestart)
	if scheduling.OnHostMaintenance != "" {
		args.OnHostMaintenance = pulumi.StringPtr(scheduling.OnHostMaintenance)
	}
	if scheduling.InstanceTerminationAction != "" {
		args.InstanceTerminationAction = pulumi.StringPtr(scheduling.InstanceTerminationAction)
	}
	if scheduling.TerminationTime != "" {
		args.TerminationTime = pulumi.StringPtr(scheduling.TerminationTime)
	}
	if scheduling.MaxRunDurationSeconds != nil {
		args.MaxRunDuration = &compute.InstanceTemplateSchedulingMaxRunDurationArgs{
			Seconds: pulumi.Int(int(scheduling.GetMaxRunDurationSeconds())),
		}
	}
	if scheduling.DiscardLocalSsdsOnStop != nil {
		args.OnInstanceStopAction = &compute.InstanceTemplateSchedulingOnInstanceStopActionArgs{
			DiscardLocalSsd: pulumi.BoolPtr(scheduling.GetDiscardLocalSsdsOnStop()),
		}
	}
	if scheduling.AvailabilityDomain != nil {
		args.AvailabilityDomain = pulumi.IntPtr(int(scheduling.GetAvailabilityDomain()))
	}
	if scheduling.MinNodeCpus != nil {
		args.MinNodeCpus = pulumi.IntPtr(int(scheduling.GetMinNodeCpus()))
	}
	if len(scheduling.NodeAffinities) > 0 {
		affinities := compute.InstanceTemplateSchedulingNodeAffinityArray{}
		for _, na := range scheduling.NodeAffinities {
			affinities = append(affinities, &compute.InstanceTemplateSchedulingNodeAffinityArgs{
				Key:      pulumi.String(na.Key),
				Operator: pulumi.String(na.Operator),
				Values:   pulumi.ToStringArray(na.Values),
			})
		}
		args.NodeAffinities = affinities
	}
	if scheduling.LocalSsdRecoveryTimeoutSeconds != nil {
		args.LocalSsdRecoveryTimeouts = compute.InstanceTemplateSchedulingLocalSsdRecoveryTimeoutArray{
			&compute.InstanceTemplateSchedulingLocalSsdRecoveryTimeoutArgs{
				Seconds: pulumi.Int(int(scheduling.GetLocalSsdRecoveryTimeoutSeconds())),
			},
		}
	}
	// Host-error detection timeout (90..330 s, steps of 30). Sent only
	// when set so an unset spec keeps Compute Engine's default recovery.
	if scheduling.HostErrorTimeoutSeconds != nil {
		args.HostErrorTimeoutSeconds = pulumi.IntPtr(int(scheduling.GetHostErrorTimeoutSeconds()))
	}
	return args
}

// regionalTemplateScheduling mirrors the zonal builder for the regional
// template's types.
func regionalTemplateScheduling(scheduling *gcpcomputemigv1alpha1.GcpComputeMigScheduling) *compute.RegionInstanceTemplateSchedulingArgs {
	isSpot := scheduling.ProvisioningModel == "SPOT"
	args := &compute.RegionInstanceTemplateSchedulingArgs{
		Preemptible: pulumi.BoolPtr(isSpot),
	}
	if scheduling.ProvisioningModel != "" {
		args.ProvisioningModel = pulumi.StringPtr(scheduling.ProvisioningModel)
	}
	automaticRestart := true
	if isSpot || scheduling.ProvisioningModel == "FLEX_START" {
		automaticRestart = false
	} else if scheduling.AutomaticRestart != nil {
		automaticRestart = scheduling.GetAutomaticRestart()
	}
	args.AutomaticRestart = pulumi.BoolPtr(automaticRestart)
	if scheduling.OnHostMaintenance != "" {
		args.OnHostMaintenance = pulumi.StringPtr(scheduling.OnHostMaintenance)
	}
	if scheduling.InstanceTerminationAction != "" {
		args.InstanceTerminationAction = pulumi.StringPtr(scheduling.InstanceTerminationAction)
	}
	if scheduling.TerminationTime != "" {
		args.TerminationTime = pulumi.StringPtr(scheduling.TerminationTime)
	}
	if scheduling.MaxRunDurationSeconds != nil {
		args.MaxRunDuration = &compute.RegionInstanceTemplateSchedulingMaxRunDurationArgs{
			Seconds: pulumi.Int(int(scheduling.GetMaxRunDurationSeconds())),
		}
	}
	if scheduling.DiscardLocalSsdsOnStop != nil {
		args.OnInstanceStopAction = &compute.RegionInstanceTemplateSchedulingOnInstanceStopActionArgs{
			DiscardLocalSsd: pulumi.BoolPtr(scheduling.GetDiscardLocalSsdsOnStop()),
		}
	}
	if scheduling.AvailabilityDomain != nil {
		args.AvailabilityDomain = pulumi.IntPtr(int(scheduling.GetAvailabilityDomain()))
	}
	if scheduling.MinNodeCpus != nil {
		args.MinNodeCpus = pulumi.IntPtr(int(scheduling.GetMinNodeCpus()))
	}
	if len(scheduling.NodeAffinities) > 0 {
		affinities := compute.RegionInstanceTemplateSchedulingNodeAffinityArray{}
		for _, na := range scheduling.NodeAffinities {
			affinities = append(affinities, &compute.RegionInstanceTemplateSchedulingNodeAffinityArgs{
				Key:      pulumi.String(na.Key),
				Operator: pulumi.String(na.Operator),
				Values:   pulumi.ToStringArray(na.Values),
			})
		}
		args.NodeAffinities = affinities
	}
	if scheduling.LocalSsdRecoveryTimeoutSeconds != nil {
		args.LocalSsdRecoveryTimeouts = compute.RegionInstanceTemplateSchedulingLocalSsdRecoveryTimeoutArray{
			&compute.RegionInstanceTemplateSchedulingLocalSsdRecoveryTimeoutArgs{
				Seconds: pulumi.Int(int(scheduling.GetLocalSsdRecoveryTimeoutSeconds())),
			},
		}
	}
	// Host-error detection timeout (90..330 s, steps of 30). Sent only
	// when set so an unset spec keeps Compute Engine's default recovery.
	if scheduling.HostErrorTimeoutSeconds != nil {
		args.HostErrorTimeoutSeconds = pulumi.IntPtr(int(scheduling.GetHostErrorTimeoutSeconds()))
	}
	return args
}
