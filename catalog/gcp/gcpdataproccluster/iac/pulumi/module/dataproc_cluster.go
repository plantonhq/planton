package module

import (
	"github.com/pkg/errors"
	gcpdataprocclusterv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdataproccluster/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/dataproc"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dataprocCluster provisions the Dataproc cluster. Two mutually
// exclusive arms mirror the API: cluster_config provisions dedicated
// Compute Engine VMs; virtual_cluster_config runs Dataproc as pods on
// an existing GKE cluster. Omitting both creates a default GCE cluster
// (2 workers, default machine types).
//
// Mutability: the cluster is create-mostly-immutable. The only in-place
// updates the API supports are labels, primary/secondary worker counts
// (manual scaling), min_num_instances, the autoscaling-policy
// attachment, and the lifecycle TTLs — everything else forces
// recreation. The virtual arm has no update paths at all.
func dataprocCluster(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpDataprocCluster.Spec

	// Enable the Dataproc API — the control plane that owns the cluster.
	// disable_on_destroy stays false: tearing down one cluster must never
	// disable the API for everything else in the project.
	dataprocApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("dataproc.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		dataprocApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdDataprocApi, err := projects.NewService(ctx,
		"dpc-dataproc.googleapis.com", dataprocApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable dataproc.googleapis.com api")
	}

	args := &dataproc.ClusterArgs{
		Name:   pulumi.String(spec.ClusterName),
		Region: pulumi.StringPtr(spec.Region),
	}

	// Honor the spec contract: an empty project_id falls back to the
	// provider's default project (omitting the argument lets the provider
	// resolve its own project).
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}

	// The Dataproc API rejects user labels on virtual (GKE-based)
	// clusters — labels (including the platform attribution set) are sent
	// only for the GCE arm, identically to the Terraform module.
	if spec.VirtualClusterConfig == nil {
		args.Labels = pulumi.ToStringMap(locals.GcpLabels)
	}

	// Applied when worker counts shrink during an update: YARN drains
	// running tasks for up to this window before nodes are removed.
	if spec.GracefulDecommissionTimeout != "" {
		args.GracefulDecommissionTimeout = pulumi.StringPtr(spec.GracefulDecommissionTimeout)
	}

	// Engine-side teardown behavior (DELETE / PREVENT / ABANDON) — the
	// ABANDON lever hands the cluster to out-of-band management.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	if spec.ClusterConfig != nil {
		args.ClusterConfig = buildClusterConfig(spec.ClusterConfig)
	}

	if spec.VirtualClusterConfig != nil {
		args.VirtualClusterConfig = buildVirtualClusterConfig(spec.VirtualClusterConfig)
	}

	createdCluster, err := dataproc.NewCluster(ctx, "dataproc-cluster", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdDataprocApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create dataproc cluster")
	}

	// The resource ID is the fully qualified cluster resource name
	// (projects/{p}/regions/{r}/clusters/{c}) — the bridged provider
	// inherits Terraform's ID format, so both engines export the exact
	// path downstream composition (e.g. spark_history_server_config)
	// consumes.
	ctx.Export(OpClusterId, createdCluster.ID())
	ctx.Export(OpClusterName, createdCluster.Name)

	// The staging bucket actually in use: the user-supplied bucket when
	// one was referenced, otherwise the bucket GCP auto-created. The
	// virtual arm reports its own staging bucket the same way.
	if spec.VirtualClusterConfig != nil {
		ctx.Export(OpStagingBucket, createdCluster.VirtualClusterConfig.StagingBucket().Elem())
	} else {
		ctx.Export(OpStagingBucket, createdCluster.ClusterConfig.Bucket().Elem())
	}

	return nil
}

// buildDiskConfig converts the shared disk-config message for any node
// group into the given block constructor via the supplied setter.
func diskFields(d *gcpdataprocclusterv1alpha1.GcpDataprocClusterDiskConfig) (bootDiskSizeGb *int, bootDiskType *string, numLocalSsds *int, localSsdInterface *string, provisionedIops *int, provisionedThroughput *int) {
	if d.BootDiskSizeGb > 0 {
		v := int(d.BootDiskSizeGb)
		bootDiskSizeGb = &v
	}
	if d.BootDiskType != "" {
		v := d.BootDiskType
		bootDiskType = &v
	}
	if d.NumLocalSsds > 0 {
		v := int(d.NumLocalSsds)
		numLocalSsds = &v
	}
	if d.LocalSsdInterface != "" {
		v := d.LocalSsdInterface
		localSsdInterface = &v
	}
	// Provisioned-performance dials (hyperdisk classes).
	if d.BootDiskProvisionedIops != nil {
		v := int(*d.BootDiskProvisionedIops)
		provisionedIops = &v
	}
	if d.BootDiskProvisionedThroughput != nil {
		v := int(*d.BootDiskProvisionedThroughput)
		provisionedThroughput = &v
	}
	return
}

// attachedDiskFields reads one attached-disk entry into the optional
// scalars every role's attached-disk block shares (0 / "" mean unset).
func attachedDiskFields(d *gcpdataprocclusterv1alpha1.GcpDataprocClusterAttachedDisk) (sizeGb *int, diskType *string, provisionedIops *int, provisionedThroughput *int) {
	if d.DiskSizeGb > 0 {
		v := int(d.DiskSizeGb)
		sizeGb = &v
	}
	if d.DiskType != "" {
		v := d.DiskType
		diskType = &v
	}
	if d.ProvisionedIops != nil {
		v := int(*d.ProvisionedIops)
		provisionedIops = &v
	}
	if d.ProvisionedThroughput != nil {
		v := int(*d.ProvisionedThroughput)
		provisionedThroughput = &v
	}
	return
}

// The bridged SDK types every role's disk blocks separately, so each role
// gets its own pair of builders below: the additional persistent disks on
// every node of the role, and the per-selection disk shape inside its
// instance flexibility policy. The bodies are identical by design.

// masterAttachedDisks builds the master role's attached-disk list; nil when
// the spec declares none so the block is omitted.
func masterAttachedDisks(d *gcpdataprocclusterv1alpha1.GcpDataprocClusterDiskConfig) dataproc.ClusterClusterConfigMasterConfigDiskConfigAttachedDiskConfigArrayInput {
	if len(d.AttachedDisks) == 0 {
		return nil
	}
	var disks dataproc.ClusterClusterConfigMasterConfigDiskConfigAttachedDiskConfigArray
	for _, ad := range d.AttachedDisks {
		sizeGb, diskType, provIops, provThroughput := attachedDiskFields(ad)
		diskArgs := &dataproc.ClusterClusterConfigMasterConfigDiskConfigAttachedDiskConfigArgs{}
		if sizeGb != nil {
			diskArgs.DiskSizeGb = pulumi.IntPtr(*sizeGb)
		}
		if diskType != nil {
			diskArgs.DiskType = pulumi.StringPtr(*diskType)
		}
		if provIops != nil {
			diskArgs.ProvisionedIops = pulumi.IntPtr(*provIops)
		}
		if provThroughput != nil {
			diskArgs.ProvisionedThroughput = pulumi.IntPtr(*provThroughput)
		}
		disks = append(disks, diskArgs)
	}
	return disks
}

// masterSelectionDiskConfig builds the per-selection disk shape for the
// master role's instance flexibility policy; nil when the selection
// inherits the role's disk_config.
func masterSelectionDiskConfig(d *gcpdataprocclusterv1alpha1.GcpDataprocClusterDiskConfig) dataproc.ClusterClusterConfigMasterConfigInstanceFlexibilityPolicyInstanceSelectionListDiskConfigPtrInput {
	if d == nil {
		return nil
	}
	sizeGb, diskType, ssds, ssdIface, provIops, provThroughput := diskFields(d)
	diskArgs := &dataproc.ClusterClusterConfigMasterConfigInstanceFlexibilityPolicyInstanceSelectionListDiskConfigArgs{}
	if sizeGb != nil {
		diskArgs.BootDiskSizeGb = pulumi.IntPtr(*sizeGb)
	}
	if diskType != nil {
		diskArgs.BootDiskType = pulumi.StringPtr(*diskType)
	}
	if ssds != nil {
		diskArgs.NumLocalSsds = pulumi.IntPtr(*ssds)
	}
	if ssdIface != nil {
		diskArgs.LocalSsdInterface = pulumi.StringPtr(*ssdIface)
	}
	if provIops != nil {
		diskArgs.BootDiskProvisionedIops = pulumi.IntPtr(*provIops)
	}
	if provThroughput != nil {
		diskArgs.BootDiskProvisionedThroughput = pulumi.IntPtr(*provThroughput)
	}
	if len(d.AttachedDisks) > 0 {
		var disks dataproc.ClusterClusterConfigMasterConfigInstanceFlexibilityPolicyInstanceSelectionListDiskConfigAttachedDiskConfigArray
		for _, ad := range d.AttachedDisks {
			adSizeGb, adDiskType, adProvIops, adProvThroughput := attachedDiskFields(ad)
			adArgs := &dataproc.ClusterClusterConfigMasterConfigInstanceFlexibilityPolicyInstanceSelectionListDiskConfigAttachedDiskConfigArgs{}
			if adSizeGb != nil {
				adArgs.DiskSizeGb = pulumi.IntPtr(*adSizeGb)
			}
			if adDiskType != nil {
				adArgs.DiskType = pulumi.StringPtr(*adDiskType)
			}
			if adProvIops != nil {
				adArgs.ProvisionedIops = pulumi.IntPtr(*adProvIops)
			}
			if adProvThroughput != nil {
				adArgs.ProvisionedThroughput = pulumi.IntPtr(*adProvThroughput)
			}
			disks = append(disks, adArgs)
		}
		diskArgs.AttachedDiskConfigs = disks
	}
	return diskArgs
}

// workerAttachedDisks builds the worker role's attached-disk list; nil when
// the spec declares none so the block is omitted.
func workerAttachedDisks(d *gcpdataprocclusterv1alpha1.GcpDataprocClusterDiskConfig) dataproc.ClusterClusterConfigWorkerConfigDiskConfigAttachedDiskConfigArrayInput {
	if len(d.AttachedDisks) == 0 {
		return nil
	}
	var disks dataproc.ClusterClusterConfigWorkerConfigDiskConfigAttachedDiskConfigArray
	for _, ad := range d.AttachedDisks {
		sizeGb, diskType, provIops, provThroughput := attachedDiskFields(ad)
		diskArgs := &dataproc.ClusterClusterConfigWorkerConfigDiskConfigAttachedDiskConfigArgs{}
		if sizeGb != nil {
			diskArgs.DiskSizeGb = pulumi.IntPtr(*sizeGb)
		}
		if diskType != nil {
			diskArgs.DiskType = pulumi.StringPtr(*diskType)
		}
		if provIops != nil {
			diskArgs.ProvisionedIops = pulumi.IntPtr(*provIops)
		}
		if provThroughput != nil {
			diskArgs.ProvisionedThroughput = pulumi.IntPtr(*provThroughput)
		}
		disks = append(disks, diskArgs)
	}
	return disks
}

// workerSelectionDiskConfig builds the per-selection disk shape for the
// worker role's instance flexibility policy; nil when the selection
// inherits the role's disk_config.
func workerSelectionDiskConfig(d *gcpdataprocclusterv1alpha1.GcpDataprocClusterDiskConfig) dataproc.ClusterClusterConfigWorkerConfigInstanceFlexibilityPolicyInstanceSelectionListDiskConfigPtrInput {
	if d == nil {
		return nil
	}
	sizeGb, diskType, ssds, ssdIface, provIops, provThroughput := diskFields(d)
	diskArgs := &dataproc.ClusterClusterConfigWorkerConfigInstanceFlexibilityPolicyInstanceSelectionListDiskConfigArgs{}
	if sizeGb != nil {
		diskArgs.BootDiskSizeGb = pulumi.IntPtr(*sizeGb)
	}
	if diskType != nil {
		diskArgs.BootDiskType = pulumi.StringPtr(*diskType)
	}
	if ssds != nil {
		diskArgs.NumLocalSsds = pulumi.IntPtr(*ssds)
	}
	if ssdIface != nil {
		diskArgs.LocalSsdInterface = pulumi.StringPtr(*ssdIface)
	}
	if provIops != nil {
		diskArgs.BootDiskProvisionedIops = pulumi.IntPtr(*provIops)
	}
	if provThroughput != nil {
		diskArgs.BootDiskProvisionedThroughput = pulumi.IntPtr(*provThroughput)
	}
	if len(d.AttachedDisks) > 0 {
		var disks dataproc.ClusterClusterConfigWorkerConfigInstanceFlexibilityPolicyInstanceSelectionListDiskConfigAttachedDiskConfigArray
		for _, ad := range d.AttachedDisks {
			adSizeGb, adDiskType, adProvIops, adProvThroughput := attachedDiskFields(ad)
			adArgs := &dataproc.ClusterClusterConfigWorkerConfigInstanceFlexibilityPolicyInstanceSelectionListDiskConfigAttachedDiskConfigArgs{}
			if adSizeGb != nil {
				adArgs.DiskSizeGb = pulumi.IntPtr(*adSizeGb)
			}
			if adDiskType != nil {
				adArgs.DiskType = pulumi.StringPtr(*adDiskType)
			}
			if adProvIops != nil {
				adArgs.ProvisionedIops = pulumi.IntPtr(*adProvIops)
			}
			if adProvThroughput != nil {
				adArgs.ProvisionedThroughput = pulumi.IntPtr(*adProvThroughput)
			}
			disks = append(disks, adArgs)
		}
		diskArgs.AttachedDiskConfigs = disks
	}
	return diskArgs
}

// secondaryAttachedDisks builds the secondary role's attached-disk list; nil when
// the spec declares none so the block is omitted.
func secondaryAttachedDisks(d *gcpdataprocclusterv1alpha1.GcpDataprocClusterDiskConfig) dataproc.ClusterClusterConfigPreemptibleWorkerConfigDiskConfigAttachedDiskConfigArrayInput {
	if len(d.AttachedDisks) == 0 {
		return nil
	}
	var disks dataproc.ClusterClusterConfigPreemptibleWorkerConfigDiskConfigAttachedDiskConfigArray
	for _, ad := range d.AttachedDisks {
		sizeGb, diskType, provIops, provThroughput := attachedDiskFields(ad)
		diskArgs := &dataproc.ClusterClusterConfigPreemptibleWorkerConfigDiskConfigAttachedDiskConfigArgs{}
		if sizeGb != nil {
			diskArgs.DiskSizeGb = pulumi.IntPtr(*sizeGb)
		}
		if diskType != nil {
			diskArgs.DiskType = pulumi.StringPtr(*diskType)
		}
		if provIops != nil {
			diskArgs.ProvisionedIops = pulumi.IntPtr(*provIops)
		}
		if provThroughput != nil {
			diskArgs.ProvisionedThroughput = pulumi.IntPtr(*provThroughput)
		}
		disks = append(disks, diskArgs)
	}
	return disks
}

// secondarySelectionDiskConfig builds the per-selection disk shape for the
// secondary role's instance flexibility policy; nil when the selection
// inherits the role's disk_config.
func secondarySelectionDiskConfig(d *gcpdataprocclusterv1alpha1.GcpDataprocClusterDiskConfig) dataproc.ClusterClusterConfigPreemptibleWorkerConfigInstanceFlexibilityPolicyInstanceSelectionListDiskConfigPtrInput {
	if d == nil {
		return nil
	}
	sizeGb, diskType, ssds, ssdIface, provIops, provThroughput := diskFields(d)
	diskArgs := &dataproc.ClusterClusterConfigPreemptibleWorkerConfigInstanceFlexibilityPolicyInstanceSelectionListDiskConfigArgs{}
	if sizeGb != nil {
		diskArgs.BootDiskSizeGb = pulumi.IntPtr(*sizeGb)
	}
	if diskType != nil {
		diskArgs.BootDiskType = pulumi.StringPtr(*diskType)
	}
	if ssds != nil {
		diskArgs.NumLocalSsds = pulumi.IntPtr(*ssds)
	}
	if ssdIface != nil {
		diskArgs.LocalSsdInterface = pulumi.StringPtr(*ssdIface)
	}
	if provIops != nil {
		diskArgs.BootDiskProvisionedIops = pulumi.IntPtr(*provIops)
	}
	if provThroughput != nil {
		diskArgs.BootDiskProvisionedThroughput = pulumi.IntPtr(*provThroughput)
	}
	if len(d.AttachedDisks) > 0 {
		var disks dataproc.ClusterClusterConfigPreemptibleWorkerConfigInstanceFlexibilityPolicyInstanceSelectionListDiskConfigAttachedDiskConfigArray
		for _, ad := range d.AttachedDisks {
			adSizeGb, adDiskType, adProvIops, adProvThroughput := attachedDiskFields(ad)
			adArgs := &dataproc.ClusterClusterConfigPreemptibleWorkerConfigInstanceFlexibilityPolicyInstanceSelectionListDiskConfigAttachedDiskConfigArgs{}
			if adSizeGb != nil {
				adArgs.DiskSizeGb = pulumi.IntPtr(*adSizeGb)
			}
			if adDiskType != nil {
				adArgs.DiskType = pulumi.StringPtr(*adDiskType)
			}
			if adProvIops != nil {
				adArgs.ProvisionedIops = pulumi.IntPtr(*adProvIops)
			}
			if adProvThroughput != nil {
				adArgs.ProvisionedThroughput = pulumi.IntPtr(*adProvThroughput)
			}
			disks = append(disks, adArgs)
		}
		diskArgs.AttachedDiskConfigs = disks
	}
	return diskArgs
}

// buildClusterConfig assembles the standard Compute Engine arm.
func buildClusterConfig(cfg *gcpdataprocclusterv1alpha1.GcpDataprocClusterConfig) *dataproc.ClusterClusterConfigArgs {
	clusterConfig := &dataproc.ClusterClusterConfigArgs{}

	// Bucket names arrive as resolved references. GCP auto-creates
	// staging/temp buckets when these are unset.
	if cfg.StagingBucket.GetValue() != "" {
		clusterConfig.StagingBucket = pulumi.StringPtr(cfg.StagingBucket.GetValue())
	}
	if cfg.TempBucket.GetValue() != "" {
		clusterConfig.TempBucket = pulumi.StringPtr(cfg.TempBucket.GetValue())
	}
	if cfg.ClusterTier != "" {
		clusterConfig.ClusterTier = pulumi.StringPtr(cfg.ClusterTier)
	}
	// Structural type (STANDARD / SINGLE_NODE / ZERO_SCALE) and execution
	// engine (DEFAULT / LIGHTNING). Both immutable.
	if cfg.ClusterType != "" {
		clusterConfig.ClusterType = pulumi.StringPtr(cfg.ClusterType)
	}
	if cfg.Engine != "" {
		clusterConfig.Engine = pulumi.StringPtr(cfg.Engine)
	}

	// ── GCE environment: networking, identity, hardening, placement ──

	if cfg.GceConfig != nil {
		gce := cfg.GceConfig
		gceArgs := &dataproc.ClusterClusterConfigGceClusterConfigArgs{}

		if gce.Network.GetValue() != "" {
			gceArgs.Network = pulumi.StringPtr(gce.Network.GetValue())
		}
		if gce.Subnetwork.GetValue() != "" {
			gceArgs.Subnetwork = pulumi.StringPtr(gce.Subnetwork.GetValue())
		}
		if gce.ServiceAccount.GetValue() != "" {
			gceArgs.ServiceAccount = pulumi.StringPtr(gce.ServiceAccount.GetValue())
		}
		if len(gce.ServiceAccountScopes) > 0 {
			gceArgs.ServiceAccountScopes = pulumi.ToStringArray(gce.ServiceAccountScopes)
		}
		if gce.Zone != "" {
			gceArgs.Zone = pulumi.StringPtr(gce.Zone)
		}
		if gce.InternalIpOnly {
			gceArgs.InternalIpOnly = pulumi.BoolPtr(true)
		}
		if len(gce.Tags) > 0 {
			gceArgs.Tags = pulumi.ToStringArray(gce.Tags)
		}
		if len(gce.Metadata) > 0 {
			gceArgs.Metadata = pulumi.ToStringMap(gce.Metadata)
		}
		// IAM-governed secure tags (distinct from network tags).
		if len(gce.ResourceManagerTags) > 0 {
			gceArgs.ResourceManagerTags = pulumi.ToStringMap(gce.ResourceManagerTags)
		}

		if gce.ShieldedInstanceConfig != nil {
			gceArgs.ShieldedInstanceConfig = &dataproc.ClusterClusterConfigGceClusterConfigShieldedInstanceConfigArgs{
				EnableSecureBoot:          pulumi.BoolPtr(gce.ShieldedInstanceConfig.EnableSecureBoot),
				EnableVtpm:                pulumi.BoolPtr(gce.ShieldedInstanceConfig.EnableVtpm),
				EnableIntegrityMonitoring: pulumi.BoolPtr(gce.ShieldedInstanceConfig.EnableIntegrityMonitoring),
			}
		}

		if gce.ReservationAffinity != nil {
			reservationArgs := &dataproc.ClusterClusterConfigGceClusterConfigReservationAffinityArgs{}
			if gce.ReservationAffinity.ConsumeReservationType != "" {
				reservationArgs.ConsumeReservationType = pulumi.StringPtr(gce.ReservationAffinity.ConsumeReservationType)
			}
			if gce.ReservationAffinity.Key != "" {
				reservationArgs.Key = pulumi.StringPtr(gce.ReservationAffinity.Key)
			}
			if len(gce.ReservationAffinity.Values) > 0 {
				reservationArgs.Values = pulumi.ToStringArray(gce.ReservationAffinity.Values)
			}
			gceArgs.ReservationAffinity = reservationArgs
		}

		if gce.NodeGroupAffinity != nil {
			gceArgs.NodeGroupAffinity = &dataproc.ClusterClusterConfigGceClusterConfigNodeGroupAffinityArgs{
				NodeGroupUri: pulumi.String(gce.NodeGroupAffinity.NodeGroupUri),
			}
		}

		if gce.ConfidentialInstanceConfig != nil {
			confidentialArgs := &dataproc.ClusterClusterConfigGceClusterConfigConfidentialInstanceConfigArgs{}
			// The provider-deprecated boolean is sent only when a manifest
			// still sets it, so a manifest on confidential_instance_type
			// alone never trips the deprecation warning.
			if gce.ConfidentialInstanceConfig.EnableConfidentialCompute {
				confidentialArgs.EnableConfidentialCompute = pulumi.BoolPtr(true)
			}
			if gce.ConfidentialInstanceConfig.ConfidentialInstanceType != "" {
				confidentialArgs.ConfidentialInstanceType = pulumi.StringPtr(gce.ConfidentialInstanceConfig.ConfidentialInstanceType)
			}
			gceArgs.ConfidentialInstanceConfig = confidentialArgs
		}

		clusterConfig.GceClusterConfig = gceArgs
	}

	// ── Master node group ──

	if cfg.MasterConfig != nil {
		m := cfg.MasterConfig
		masterArgs := &dataproc.ClusterClusterConfigMasterConfigArgs{}

		if m.NumInstances > 0 {
			masterArgs.NumInstances = pulumi.IntPtr(int(m.NumInstances))
		}
		// MachineType never coexists with InstanceFlexibilityPolicy (spec
		// CEL): the API drops a paired machineTypeUri from the stored
		// config, and the create-only argument then re-plans as cluster
		// replacement forever.
		if m.MachineType != "" {
			masterArgs.MachineType = pulumi.StringPtr(m.MachineType)
		}
		if m.MinCpuPlatform != "" {
			masterArgs.MinCpuPlatform = pulumi.StringPtr(m.MinCpuPlatform)
		}
		if m.ImageUri != "" {
			masterArgs.ImageUri = pulumi.StringPtr(m.ImageUri)
		}
		if m.DiskConfig != nil {
			sizeGb, diskType, ssds, ssdIface, provIops, provThroughput := diskFields(m.DiskConfig)
			diskArgs := &dataproc.ClusterClusterConfigMasterConfigDiskConfigArgs{}
			if sizeGb != nil {
				diskArgs.BootDiskSizeGb = pulumi.IntPtr(*sizeGb)
			}
			if diskType != nil {
				diskArgs.BootDiskType = pulumi.StringPtr(*diskType)
			}
			if ssds != nil {
				diskArgs.NumLocalSsds = pulumi.IntPtr(*ssds)
			}
			if ssdIface != nil {
				diskArgs.LocalSsdInterface = pulumi.StringPtr(*ssdIface)
			}
			if provIops != nil {
				diskArgs.BootDiskProvisionedIops = pulumi.IntPtr(*provIops)
			}
			if provThroughput != nil {
				diskArgs.BootDiskProvisionedThroughput = pulumi.IntPtr(*provThroughput)
			}
			diskArgs.AttachedDiskConfigs = masterAttachedDisks(m.DiskConfig)
			masterArgs.DiskConfig = diskArgs
		}
		if len(m.Accelerators) > 0 {
			var accels dataproc.ClusterClusterConfigMasterConfigAcceleratorArray
			for _, a := range m.Accelerators {
				accels = append(accels, &dataproc.ClusterClusterConfigMasterConfigAcceleratorArgs{
					AcceleratorType:  pulumi.String(a.AcceleratorType),
					AcceleratorCount: pulumi.Int(int(a.AcceleratorCount)),
				})
			}
			masterArgs.Accelerators = accels
		}

		// Ranked machine-type fallbacks (masters carry no provisioning
		// mix — that argument exists on secondary workers only; spec CEL
		// enforces it).
		if m.InstanceFlexibilityPolicy != nil && len(m.InstanceFlexibilityPolicy.InstanceSelectionList) > 0 {
			var selections dataproc.ClusterClusterConfigMasterConfigInstanceFlexibilityPolicyInstanceSelectionListArray
			for _, sel := range m.InstanceFlexibilityPolicy.InstanceSelectionList {
				selections = append(selections, &dataproc.ClusterClusterConfigMasterConfigInstanceFlexibilityPolicyInstanceSelectionListArgs{
					MachineTypes: pulumi.ToStringArray(sel.MachineTypes),
					Rank:         pulumi.IntPtr(int(sel.Rank)),
					DiskConfig:   masterSelectionDiskConfig(sel.DiskConfig),
				})
			}
			masterArgs.InstanceFlexibilityPolicy = &dataproc.ClusterClusterConfigMasterConfigInstanceFlexibilityPolicyArgs{
				InstanceSelectionLists: selections,
			}
		}

		clusterConfig.MasterConfig = masterArgs
	}

	// ── Primary worker node group ──
	// num_instances / min_num_instances are the manual-scaling levers —
	// the only node counts that update in place (with
	// graceful_decommission_timeout honored on shrink).

	if cfg.WorkerConfig != nil {
		w := cfg.WorkerConfig
		workerArgs := &dataproc.ClusterClusterConfigWorkerConfigArgs{}

		if w.NumInstances > 0 {
			workerArgs.NumInstances = pulumi.IntPtr(int(w.NumInstances))
		}
		// MachineType never coexists with InstanceFlexibilityPolicy (spec
		// CEL) — see the master-config note.
		if w.MachineType != "" {
			workerArgs.MachineType = pulumi.StringPtr(w.MachineType)
		}
		if w.MinCpuPlatform != "" {
			workerArgs.MinCpuPlatform = pulumi.StringPtr(w.MinCpuPlatform)
		}
		if w.ImageUri != "" {
			workerArgs.ImageUri = pulumi.StringPtr(w.ImageUri)
		}
		if w.MinNumInstances > 0 {
			workerArgs.MinNumInstances = pulumi.IntPtr(int(w.MinNumInstances))
		}
		if w.DiskConfig != nil {
			sizeGb, diskType, ssds, ssdIface, provIops, provThroughput := diskFields(w.DiskConfig)
			diskArgs := &dataproc.ClusterClusterConfigWorkerConfigDiskConfigArgs{}
			if sizeGb != nil {
				diskArgs.BootDiskSizeGb = pulumi.IntPtr(*sizeGb)
			}
			if diskType != nil {
				diskArgs.BootDiskType = pulumi.StringPtr(*diskType)
			}
			if ssds != nil {
				diskArgs.NumLocalSsds = pulumi.IntPtr(*ssds)
			}
			if ssdIface != nil {
				diskArgs.LocalSsdInterface = pulumi.StringPtr(*ssdIface)
			}
			if provIops != nil {
				diskArgs.BootDiskProvisionedIops = pulumi.IntPtr(*provIops)
			}
			if provThroughput != nil {
				diskArgs.BootDiskProvisionedThroughput = pulumi.IntPtr(*provThroughput)
			}
			diskArgs.AttachedDiskConfigs = workerAttachedDisks(w.DiskConfig)
			workerArgs.DiskConfig = diskArgs
		}
		if len(w.Accelerators) > 0 {
			var accels dataproc.ClusterClusterConfigWorkerConfigAcceleratorArray
			for _, a := range w.Accelerators {
				accels = append(accels, &dataproc.ClusterClusterConfigWorkerConfigAcceleratorArgs{
					AcceleratorType:  pulumi.String(a.AcceleratorType),
					AcceleratorCount: pulumi.Int(int(a.AcceleratorCount)),
				})
			}
			workerArgs.Accelerators = accels
		}

		// Ranked machine-type fallbacks (primary workers carry no
		// provisioning mix — that argument exists on secondary workers
		// only; spec CEL enforces it).
		if w.InstanceFlexibilityPolicy != nil && len(w.InstanceFlexibilityPolicy.InstanceSelectionList) > 0 {
			var selections dataproc.ClusterClusterConfigWorkerConfigInstanceFlexibilityPolicyInstanceSelectionListArray
			for _, sel := range w.InstanceFlexibilityPolicy.InstanceSelectionList {
				selections = append(selections, &dataproc.ClusterClusterConfigWorkerConfigInstanceFlexibilityPolicyInstanceSelectionListArgs{
					MachineTypes: pulumi.ToStringArray(sel.MachineTypes),
					Rank:         pulumi.IntPtr(int(sel.Rank)),
					DiskConfig:   workerSelectionDiskConfig(sel.DiskConfig),
				})
			}
			workerArgs.InstanceFlexibilityPolicy = &dataproc.ClusterClusterConfigWorkerConfigInstanceFlexibilityPolicyArgs{
				InstanceSelectionLists: selections,
			}
		}

		clusterConfig.WorkerConfig = workerArgs
	}

	// ── Secondary (preemptible/spot) worker group ──
	// The count updates in place; preemptibility is immutable. Machine
	// shape is inherited from the primary workers unless the flexibility
	// policy overrides it.

	if cfg.SecondaryWorkerConfig != nil {
		s := cfg.SecondaryWorkerConfig
		secondaryArgs := &dataproc.ClusterClusterConfigPreemptibleWorkerConfigArgs{}

		if s.NumInstances > 0 {
			secondaryArgs.NumInstances = pulumi.IntPtr(int(s.NumInstances))
		}
		if s.Preemptibility != "" {
			secondaryArgs.Preemptibility = pulumi.StringPtr(s.Preemptibility)
		}
		if s.DiskConfig != nil {
			sizeGb, diskType, ssds, ssdIface, provIops, provThroughput := diskFields(s.DiskConfig)
			diskArgs := &dataproc.ClusterClusterConfigPreemptibleWorkerConfigDiskConfigArgs{}
			if sizeGb != nil {
				diskArgs.BootDiskSizeGb = pulumi.IntPtr(*sizeGb)
			}
			if diskType != nil {
				diskArgs.BootDiskType = pulumi.StringPtr(*diskType)
			}
			if ssds != nil {
				diskArgs.NumLocalSsds = pulumi.IntPtr(*ssds)
			}
			if ssdIface != nil {
				diskArgs.LocalSsdInterface = pulumi.StringPtr(*ssdIface)
			}
			if provIops != nil {
				diskArgs.BootDiskProvisionedIops = pulumi.IntPtr(*provIops)
			}
			if provThroughput != nil {
				diskArgs.BootDiskProvisionedThroughput = pulumi.IntPtr(*provThroughput)
			}
			diskArgs.AttachedDiskConfigs = secondaryAttachedDisks(s.DiskConfig)
			secondaryArgs.DiskConfig = diskArgs
		}

		if s.InstanceFlexibilityPolicy != nil {
			flexArgs := &dataproc.ClusterClusterConfigPreemptibleWorkerConfigInstanceFlexibilityPolicyArgs{}

			if len(s.InstanceFlexibilityPolicy.InstanceSelectionList) > 0 {
				var selections dataproc.ClusterClusterConfigPreemptibleWorkerConfigInstanceFlexibilityPolicyInstanceSelectionListArray
				for _, sel := range s.InstanceFlexibilityPolicy.InstanceSelectionList {
					selections = append(selections, &dataproc.ClusterClusterConfigPreemptibleWorkerConfigInstanceFlexibilityPolicyInstanceSelectionListArgs{
						MachineTypes: pulumi.ToStringArray(sel.MachineTypes),
						Rank:         pulumi.IntPtr(int(sel.Rank)),
						DiskConfig:   secondarySelectionDiskConfig(sel.DiskConfig),
					})
				}
				flexArgs.InstanceSelectionLists = selections
			}

			if s.InstanceFlexibilityPolicy.ProvisioningModelMix != nil {
				flexArgs.ProvisioningModelMix = &dataproc.ClusterClusterConfigPreemptibleWorkerConfigInstanceFlexibilityPolicyProvisioningModelMixArgs{
					StandardCapacityBase:             pulumi.IntPtr(int(s.InstanceFlexibilityPolicy.ProvisioningModelMix.StandardCapacityBase)),
					StandardCapacityPercentAboveBase: pulumi.IntPtr(int(s.InstanceFlexibilityPolicy.ProvisioningModelMix.StandardCapacityPercentAboveBase)),
				}
			}

			secondaryArgs.InstanceFlexibilityPolicy = flexArgs
		}

		clusterConfig.PreemptibleWorkerConfig = secondaryArgs
	}

	// ── Software config ──
	// The spec's `properties` map feeds the provider's
	// override_properties — the API's writable surface (the provider's
	// `properties` attribute is the computed resolved set).

	if cfg.SoftwareConfig != nil {
		sw := cfg.SoftwareConfig
		softwareArgs := &dataproc.ClusterClusterConfigSoftwareConfigArgs{}

		if sw.ImageVersion != "" {
			softwareArgs.ImageVersion = pulumi.StringPtr(sw.ImageVersion)
		}
		if len(sw.OptionalComponents) > 0 {
			softwareArgs.OptionalComponents = pulumi.ToStringArray(sw.OptionalComponents)
		}
		if len(sw.Properties) > 0 {
			softwareArgs.OverrideProperties = pulumi.ToStringMap(sw.Properties)
		}

		clusterConfig.SoftwareConfig = softwareArgs
	}

	// ── Initialization actions ──

	if len(cfg.InitializationActions) > 0 {
		var initActions dataproc.ClusterClusterConfigInitializationActionArray
		for _, action := range cfg.InitializationActions {
			initArgs := &dataproc.ClusterClusterConfigInitializationActionArgs{
				Script: pulumi.String(action.Script),
			}
			if action.TimeoutSec > 0 {
				initArgs.TimeoutSec = pulumi.IntPtr(int(action.TimeoutSec))
			}
			initActions = append(initActions, initArgs)
		}
		clusterConfig.InitializationActions = initActions
	}

	// ── Autoscaling policy attachment ──
	// A first-class resource referenced by its full resource name;
	// attach/swap/detach updates in place.

	if cfg.AutoscalingPolicyUri.GetValue() != "" {
		clusterConfig.AutoscalingConfig = &dataproc.ClusterClusterConfigAutoscalingConfigArgs{
			PolicyUri: pulumi.String(cfg.AutoscalingPolicyUri.GetValue()),
		}
	}

	// ── CMEK for all persistent disks (key change forces recreation) ──

	if cfg.EncryptionKmsKeyName.GetValue() != "" {
		clusterConfig.EncryptionConfig = &dataproc.ClusterClusterConfigEncryptionConfigArgs{
			KmsKeyName: pulumi.String(cfg.EncryptionKmsKeyName.GetValue()),
		}
	}

	// ── Kerberos XOR personal-cluster identity mapping ──
	// Kerberos secret fields are GCS URIs of KMS-encrypted files — never
	// inline material (the API's own contract).

	if cfg.SecurityConfig != nil {
		securityArgs := &dataproc.ClusterClusterConfigSecurityConfigArgs{}

		if k := cfg.SecurityConfig.KerberosConfig; k != nil {
			kerberosArgs := &dataproc.ClusterClusterConfigSecurityConfigKerberosConfigArgs{
				RootPrincipalPasswordUri: pulumi.String(k.RootPrincipalPasswordUri),
				KmsKeyUri:                pulumi.String(k.KmsKeyUri.GetValue()),
			}
			if k.EnableKerberos {
				kerberosArgs.EnableKerberos = pulumi.BoolPtr(true)
			}
			if k.Realm != "" {
				kerberosArgs.Realm = pulumi.StringPtr(k.Realm)
			}
			if k.TgtLifetimeHours > 0 {
				kerberosArgs.TgtLifetimeHours = pulumi.IntPtr(int(k.TgtLifetimeHours))
			}
			if k.KdcDbKeyUri != "" {
				kerberosArgs.KdcDbKeyUri = pulumi.StringPtr(k.KdcDbKeyUri)
			}
			if k.KeystoreUri != "" {
				kerberosArgs.KeystoreUri = pulumi.StringPtr(k.KeystoreUri)
			}
			if k.KeystorePasswordUri != "" {
				kerberosArgs.KeystorePasswordUri = pulumi.StringPtr(k.KeystorePasswordUri)
			}
			if k.KeyPasswordUri != "" {
				kerberosArgs.KeyPasswordUri = pulumi.StringPtr(k.KeyPasswordUri)
			}
			if k.TruststoreUri != "" {
				kerberosArgs.TruststoreUri = pulumi.StringPtr(k.TruststoreUri)
			}
			if k.TruststorePasswordUri != "" {
				kerberosArgs.TruststorePasswordUri = pulumi.StringPtr(k.TruststorePasswordUri)
			}
			if k.CrossRealmTrustRealm != "" {
				kerberosArgs.CrossRealmTrustRealm = pulumi.StringPtr(k.CrossRealmTrustRealm)
			}
			if k.CrossRealmTrustKdc != "" {
				kerberosArgs.CrossRealmTrustKdc = pulumi.StringPtr(k.CrossRealmTrustKdc)
			}
			if k.CrossRealmTrustAdminServer != "" {
				kerberosArgs.CrossRealmTrustAdminServer = pulumi.StringPtr(k.CrossRealmTrustAdminServer)
			}
			if k.CrossRealmTrustSharedPasswordUri != "" {
				kerberosArgs.CrossRealmTrustSharedPasswordUri = pulumi.StringPtr(k.CrossRealmTrustSharedPasswordUri)
			}
			securityArgs.KerberosConfig = kerberosArgs
		}

		if id := cfg.SecurityConfig.IdentityConfig; id != nil {
			securityArgs.IdentityConfig = &dataproc.ClusterClusterConfigSecurityConfigIdentityConfigArgs{
				UserServiceAccountMapping: pulumi.ToStringMap(id.UserServiceAccountMapping),
			}
		}

		clusterConfig.SecurityConfig = securityArgs
	}

	// ── Component Gateway (authenticated web UIs) ──

	if cfg.EndpointConfig != nil {
		clusterConfig.EndpointConfig = &dataproc.ClusterClusterConfigEndpointConfigArgs{
			EnableHttpPortAccess: pulumi.Bool(cfg.EndpointConfig.EnableHttpPortAccess),
		}
	}

	// ── Cost-control TTLs — both update in place ──

	if cfg.LifecycleConfig != nil {
		lifecycleArgs := &dataproc.ClusterClusterConfigLifecycleConfigArgs{}
		if cfg.LifecycleConfig.IdleDeleteTtl != "" {
			lifecycleArgs.IdleDeleteTtl = pulumi.StringPtr(cfg.LifecycleConfig.IdleDeleteTtl)
		}
		if cfg.LifecycleConfig.AutoDeleteTime != "" {
			lifecycleArgs.AutoDeleteTime = pulumi.StringPtr(cfg.LifecycleConfig.AutoDeleteTime)
		}
		// Stop (not delete): VMs shut down, cluster stays restartable.
		if cfg.LifecycleConfig.IdleStopTtl != "" {
			lifecycleArgs.IdleStopTtl = pulumi.StringPtr(cfg.LifecycleConfig.IdleStopTtl)
		}
		if cfg.LifecycleConfig.AutoStopTime != "" {
			lifecycleArgs.AutoStopTime = pulumi.StringPtr(cfg.LifecycleConfig.AutoStopTime)
		}
		clusterConfig.LifecycleConfig = lifecycleArgs
	}

	// ── Persistent shared Hive metastore ──

	if cfg.MetastoreConfig != nil {
		clusterConfig.MetastoreConfig = &dataproc.ClusterClusterConfigMetastoreConfigArgs{
			DataprocMetastoreService: pulumi.String(cfg.MetastoreConfig.DataprocMetastoreService.GetValue()),
		}
	}

	// ── OSS metric collection into Cloud Monitoring ──

	if cfg.DataprocMetricConfig != nil {
		var metrics dataproc.ClusterClusterConfigDataprocMetricConfigMetricArray
		for _, m := range cfg.DataprocMetricConfig.Metrics {
			metricArgs := &dataproc.ClusterClusterConfigDataprocMetricConfigMetricArgs{
				MetricSource: pulumi.String(m.MetricSource),
			}
			if len(m.MetricOverrides) > 0 {
				metricArgs.MetricOverrides = pulumi.ToStringArray(m.MetricOverrides)
			}
			metrics = append(metrics, metricArgs)
		}
		clusterConfig.DataprocMetricConfig = &dataproc.ClusterClusterConfigDataprocMetricConfigArgs{
			Metrics: metrics,
		}
	}

	// ── Dedicated DRIVER node groups ──

	if len(cfg.AuxiliaryNodeGroups) > 0 {
		var groups dataproc.ClusterClusterConfigAuxiliaryNodeGroupArray
		for _, g := range cfg.AuxiliaryNodeGroups {
			nodeGroupArgs := &dataproc.ClusterClusterConfigAuxiliaryNodeGroupNodeGroupArgs{
				Roles: pulumi.ToStringArray(g.Roles),
			}

			if g.NodeGroupConfig != nil {
				ngc := g.NodeGroupConfig
				configArgs := &dataproc.ClusterClusterConfigAuxiliaryNodeGroupNodeGroupNodeGroupConfigArgs{}
				if ngc.NumInstances > 0 {
					configArgs.NumInstances = pulumi.IntPtr(int(ngc.NumInstances))
				}
				if ngc.MachineType != "" {
					configArgs.MachineType = pulumi.StringPtr(ngc.MachineType)
				}
				if ngc.MinCpuPlatform != "" {
					configArgs.MinCpuPlatform = pulumi.StringPtr(ngc.MinCpuPlatform)
				}
				if ngc.DiskConfig != nil {
					sizeGb, diskType, ssds, ssdIface, provIops, provThroughput := diskFields(ngc.DiskConfig)
					diskArgs := &dataproc.ClusterClusterConfigAuxiliaryNodeGroupNodeGroupNodeGroupConfigDiskConfigArgs{}
					if sizeGb != nil {
						diskArgs.BootDiskSizeGb = pulumi.IntPtr(*sizeGb)
					}
					if diskType != nil {
						diskArgs.BootDiskType = pulumi.StringPtr(*diskType)
					}
					if ssds != nil {
						diskArgs.NumLocalSsds = pulumi.IntPtr(*ssds)
					}
					if ssdIface != nil {
						diskArgs.LocalSsdInterface = pulumi.StringPtr(*ssdIface)
					}
					if provIops != nil {
						diskArgs.BootDiskProvisionedIops = pulumi.IntPtr(*provIops)
					}
					if provThroughput != nil {
						diskArgs.BootDiskProvisionedThroughput = pulumi.IntPtr(*provThroughput)
					}
					configArgs.DiskConfig = diskArgs
				}
				if len(ngc.Accelerators) > 0 {
					var accels dataproc.ClusterClusterConfigAuxiliaryNodeGroupNodeGroupNodeGroupConfigAcceleratorArray
					for _, a := range ngc.Accelerators {
						accels = append(accels, &dataproc.ClusterClusterConfigAuxiliaryNodeGroupNodeGroupNodeGroupConfigAcceleratorArgs{
							AcceleratorType:  pulumi.String(a.AcceleratorType),
							AcceleratorCount: pulumi.Int(int(a.AcceleratorCount)),
						})
					}
					configArgs.Accelerators = accels
				}
				nodeGroupArgs.NodeGroupConfig = configArgs
			}

			groupArgs := &dataproc.ClusterClusterConfigAuxiliaryNodeGroupArgs{
				NodeGroups: dataproc.ClusterClusterConfigAuxiliaryNodeGroupNodeGroupArray{nodeGroupArgs},
			}
			if g.NodeGroupId != "" {
				groupArgs.NodeGroupId = pulumi.StringPtr(g.NodeGroupId)
			}
			groups = append(groups, groupArgs)
		}
		clusterConfig.AuxiliaryNodeGroups = groups
	}

	return clusterConfig
}

// buildVirtualClusterConfig assembles the Dataproc-on-GKE arm. All
// references arrive resolved to the fully qualified resource names the
// Dataproc API requires. The whole arm is immutable — changes replace
// the virtual cluster without touching the underlying GKE resources.
func buildVirtualClusterConfig(vcc *gcpdataprocclusterv1alpha1.GcpDataprocClusterVirtualClusterConfig) *dataproc.ClusterVirtualClusterConfigArgs {
	args := &dataproc.ClusterVirtualClusterConfigArgs{}

	if vcc.StagingBucket.GetValue() != "" {
		args.StagingBucket = pulumi.StringPtr(vcc.StagingBucket.GetValue())
	}

	kcc := vcc.KubernetesClusterConfig
	kccArgs := &dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigArgs{}

	if kcc.KubernetesNamespace.GetValue() != "" {
		kccArgs.KubernetesNamespace = pulumi.StringPtr(kcc.KubernetesNamespace.GetValue())
	}

	gkeArgs := &dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigGkeClusterConfigArgs{
		GkeClusterTarget: pulumi.StringPtr(kcc.GkeClusterConfig.GkeClusterTarget.GetValue()),
	}

	if len(kcc.GkeClusterConfig.NodePoolTarget) > 0 {
		var targets dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigGkeClusterConfigNodePoolTargetArray
		for _, t := range kcc.GkeClusterConfig.NodePoolTarget {
			targetArgs := &dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigGkeClusterConfigNodePoolTargetArgs{
				NodePool: pulumi.String(t.NodePool.GetValue()),
				Roles:    pulumi.ToStringArray(t.Roles),
			}

			if t.NodePoolConfig != nil {
				npc := t.NodePoolConfig
				npcArgs := &dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigGkeClusterConfigNodePoolTargetNodePoolConfigArgs{
					Locations: pulumi.ToStringArray(npc.Locations),
				}

				if npc.Autoscaling != nil {
					npcArgs.Autoscaling = &dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigGkeClusterConfigNodePoolTargetNodePoolConfigAutoscalingArgs{
						MinNodeCount: pulumi.IntPtr(int(npc.Autoscaling.MinNodeCount)),
						MaxNodeCount: pulumi.IntPtr(int(npc.Autoscaling.MaxNodeCount)),
					}
				}

				// Sizing for a Dataproc-created pool; ignored when the
				// referenced pool already exists.
				if npc.MachineType != "" || npc.LocalSsdCount > 0 || npc.MinCpuPlatform != "" || npc.Preemptible || npc.Spot {
					configArgs := &dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigGkeClusterConfigNodePoolTargetNodePoolConfigConfigArgs{}
					if npc.MachineType != "" {
						configArgs.MachineType = pulumi.StringPtr(npc.MachineType)
					}
					if npc.LocalSsdCount > 0 {
						configArgs.LocalSsdCount = pulumi.IntPtr(int(npc.LocalSsdCount))
					}
					if npc.MinCpuPlatform != "" {
						configArgs.MinCpuPlatform = pulumi.StringPtr(npc.MinCpuPlatform)
					}
					if npc.Preemptible {
						configArgs.Preemptible = pulumi.BoolPtr(true)
					}
					if npc.Spot {
						configArgs.Spot = pulumi.BoolPtr(true)
					}
					npcArgs.Config = configArgs
				}

				targetArgs.NodePoolConfig = npcArgs
			}

			targets = append(targets, targetArgs)
		}
		gkeArgs.NodePoolTargets = targets
	}

	kccArgs.GkeClusterConfig = gkeArgs

	softwareArgs := &dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigKubernetesSoftwareConfigArgs{
		ComponentVersion: pulumi.ToStringMap(kcc.KubernetesSoftwareConfig.ComponentVersion),
	}
	if len(kcc.KubernetesSoftwareConfig.Properties) > 0 {
		softwareArgs.Properties = pulumi.ToStringMap(kcc.KubernetesSoftwareConfig.Properties)
	}
	kccArgs.KubernetesSoftwareConfig = softwareArgs

	args.KubernetesClusterConfig = kccArgs

	if vcc.AuxiliaryServicesConfig != nil {
		auxArgs := &dataproc.ClusterVirtualClusterConfigAuxiliaryServicesConfigArgs{}

		if vcc.AuxiliaryServicesConfig.MetastoreConfig != nil {
			auxArgs.MetastoreConfig = &dataproc.ClusterVirtualClusterConfigAuxiliaryServicesConfigMetastoreConfigArgs{
				DataprocMetastoreService: pulumi.StringPtr(vcc.AuxiliaryServicesConfig.MetastoreConfig.DataprocMetastoreService.GetValue()),
			}
		}

		if shs := vcc.AuxiliaryServicesConfig.SparkHistoryServerConfig; shs != nil && shs.DataprocCluster.GetValue() != "" {
			auxArgs.SparkHistoryServerConfig = &dataproc.ClusterVirtualClusterConfigAuxiliaryServicesConfigSparkHistoryServerConfigArgs{
				DataprocCluster: pulumi.StringPtr(shs.DataprocCluster.GetValue()),
			}
		}

		args.AuxiliaryServicesConfig = auxArgs
	}

	return args
}
