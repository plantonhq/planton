package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/dataproc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func dataprocCluster(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpDataprocCluster.Spec

	args := &dataproc.ClusterArgs{
		Name:    pulumi.String(spec.ClusterName),
		Region:  pulumi.StringPtr(spec.Region),
		Project: pulumi.StringPtr(spec.ProjectId.GetValue()),
		Labels:  pulumi.ToStringMap(locals.GcpLabels),
	}

	// Graceful decommission timeout.
	if spec.GracefulDecommissionTimeout != "" {
		args.GracefulDecommissionTimeout = pulumi.StringPtr(spec.GracefulDecommissionTimeout)
	}

	// Build cluster_config.
	if spec.ClusterConfig != nil {
		cfg := spec.ClusterConfig
		clusterConfig := &dataproc.ClusterClusterConfigArgs{}

		// ── Staging and temp buckets ──

		if cfg.StagingBucket != nil && cfg.StagingBucket.GetValue() != "" {
			clusterConfig.StagingBucket = pulumi.StringPtr(cfg.StagingBucket.GetValue())
		}
		if cfg.TempBucket != nil && cfg.TempBucket.GetValue() != "" {
			clusterConfig.TempBucket = pulumi.StringPtr(cfg.TempBucket.GetValue())
		}

		// ── GCE cluster config (networking, service account, zone, tags) ──

		if cfg.GceConfig != nil {
			gce := cfg.GceConfig
			gceArgs := &dataproc.ClusterClusterConfigGceClusterConfigArgs{}

			if gce.Network != nil && gce.Network.GetValue() != "" {
				gceArgs.Network = pulumi.StringPtr(gce.Network.GetValue())
			}
			if gce.Subnetwork != nil && gce.Subnetwork.GetValue() != "" {
				gceArgs.Subnetwork = pulumi.StringPtr(gce.Subnetwork.GetValue())
			}
			if gce.ServiceAccount != nil && gce.ServiceAccount.GetValue() != "" {
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

			clusterConfig.GceClusterConfig = gceArgs
		}

		// ── Master config ──

		if cfg.MasterConfig != nil {
			m := cfg.MasterConfig
			masterArgs := &dataproc.ClusterClusterConfigMasterConfigArgs{}

			if m.NumInstances > 0 {
				masterArgs.NumInstances = pulumi.IntPtr(int(m.NumInstances))
			}
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
				diskArgs := &dataproc.ClusterClusterConfigMasterConfigDiskConfigArgs{}
				if m.DiskConfig.BootDiskSizeGb > 0 {
					diskArgs.BootDiskSizeGb = pulumi.IntPtr(int(m.DiskConfig.BootDiskSizeGb))
				}
				if m.DiskConfig.BootDiskType != "" {
					diskArgs.BootDiskType = pulumi.StringPtr(m.DiskConfig.BootDiskType)
				}
				if m.DiskConfig.NumLocalSsds > 0 {
					diskArgs.NumLocalSsds = pulumi.IntPtr(int(m.DiskConfig.NumLocalSsds))
				}
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

			clusterConfig.MasterConfig = masterArgs
		}

		// ── Worker config ──

		if cfg.WorkerConfig != nil {
			w := cfg.WorkerConfig
			workerArgs := &dataproc.ClusterClusterConfigWorkerConfigArgs{}

			if w.NumInstances > 0 {
				workerArgs.NumInstances = pulumi.IntPtr(int(w.NumInstances))
			}
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
				diskArgs := &dataproc.ClusterClusterConfigWorkerConfigDiskConfigArgs{}
				if w.DiskConfig.BootDiskSizeGb > 0 {
					diskArgs.BootDiskSizeGb = pulumi.IntPtr(int(w.DiskConfig.BootDiskSizeGb))
				}
				if w.DiskConfig.BootDiskType != "" {
					diskArgs.BootDiskType = pulumi.StringPtr(w.DiskConfig.BootDiskType)
				}
				if w.DiskConfig.NumLocalSsds > 0 {
					diskArgs.NumLocalSsds = pulumi.IntPtr(int(w.DiskConfig.NumLocalSsds))
				}
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

			clusterConfig.WorkerConfig = workerArgs
		}

		// ── Secondary worker config (preemptible/spot) ──

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
				diskArgs := &dataproc.ClusterClusterConfigPreemptibleWorkerConfigDiskConfigArgs{}
				if s.DiskConfig.BootDiskSizeGb > 0 {
					diskArgs.BootDiskSizeGb = pulumi.IntPtr(int(s.DiskConfig.BootDiskSizeGb))
				}
				if s.DiskConfig.BootDiskType != "" {
					diskArgs.BootDiskType = pulumi.StringPtr(s.DiskConfig.BootDiskType)
				}
				if s.DiskConfig.NumLocalSsds > 0 {
					diskArgs.NumLocalSsds = pulumi.IntPtr(int(s.DiskConfig.NumLocalSsds))
				}
				secondaryArgs.DiskConfig = diskArgs
			}

			clusterConfig.PreemptibleWorkerConfig = secondaryArgs
		}

		// ── Software config ──

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

		// ── Autoscaling policy ──

		if cfg.AutoscalingPolicyUri != "" {
			clusterConfig.AutoscalingConfig = &dataproc.ClusterClusterConfigAutoscalingConfigArgs{
				PolicyUri: pulumi.String(cfg.AutoscalingPolicyUri),
			}
		}

		// ── CMEK encryption for persistent disks ──

		if cfg.EncryptionKmsKeyName != nil && cfg.EncryptionKmsKeyName.GetValue() != "" {
			clusterConfig.EncryptionConfig = &dataproc.ClusterClusterConfigEncryptionConfigArgs{
				KmsKeyName: pulumi.String(cfg.EncryptionKmsKeyName.GetValue()),
			}
		}

		// ── Component Gateway (endpoint config) ──

		if cfg.EndpointConfig != nil {
			clusterConfig.EndpointConfig = &dataproc.ClusterClusterConfigEndpointConfigArgs{
				EnableHttpPortAccess: pulumi.Bool(cfg.EndpointConfig.EnableHttpPortAccess),
			}
		}

		// ── Lifecycle config (auto-delete, idle shutdown) ──

		if cfg.LifecycleConfig != nil {
			lifecycleArgs := &dataproc.ClusterClusterConfigLifecycleConfigArgs{}
			if cfg.LifecycleConfig.IdleDeleteTtl != "" {
				lifecycleArgs.IdleDeleteTtl = pulumi.StringPtr(cfg.LifecycleConfig.IdleDeleteTtl)
			}
			if cfg.LifecycleConfig.AutoDeleteTime != "" {
				lifecycleArgs.AutoDeleteTime = pulumi.StringPtr(cfg.LifecycleConfig.AutoDeleteTime)
			}
			clusterConfig.LifecycleConfig = lifecycleArgs
		}

		args.ClusterConfig = clusterConfig
	}

	createdCluster, err := dataproc.NewCluster(ctx, "dataproc-cluster", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create dataproc cluster")
	}

	// Export outputs.
	ctx.Export(OpClusterName, pulumi.String(spec.ClusterName))
	ctx.Export(OpClusterId, createdCluster.Name)

	// Staging bucket: the computed bucket from cluster config (user-supplied or auto-created).
	ctx.Export(OpStagingBucket, createdCluster.ClusterConfig.Bucket())

	// Cluster UUID is not directly exposed by the Pulumi provider.
	// Export an empty placeholder; downstream consumers should use cluster_id.
	ctx.Export(OpClusterUuid, pulumi.String(""))

	return nil
}
