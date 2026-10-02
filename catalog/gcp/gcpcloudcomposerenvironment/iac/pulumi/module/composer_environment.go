package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/envsecrets"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/composer"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// composerEnvironment provisions a Cloud Composer environment (managed
// Airflow). Composer assembles a GKE cluster, Cloud SQL metadata
// database, web server, and DAG bucket behind this one resource —
// creation takes 25-45 minutes.
//
// Mutability: environment name, region, node networking, encryption,
// and the private environment configuration are immutable (recreate on
// change). Workloads sizing, environment size, resilience mode,
// software configuration, maintenance window, web-server access
// control, and labels update in place.
func composerEnvironment(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpCloudComposerEnvironment.Spec

	// Enable the Cloud Composer API — the control plane that owns the
	// environment. disable_on_destroy stays false: tearing down one
	// environment must never disable the API for everything else in the
	// project.
	composerApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("composer.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		composerApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdComposerApi, err := projects.NewService(ctx,
		"cce-composer.googleapis.com", composerApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable composer.googleapis.com api")
	}

	// Determine environment name: explicit spec field, else metadata name.
	envName := spec.EnvironmentName
	if envName == "" && locals.GcpCloudComposerEnvironment.Metadata != nil {
		envName = locals.GcpCloudComposerEnvironment.Metadata.Name
	}

	// Secret values are stored in Secret Manager before the environment
	// exists; Airflow receives each version's resource name as the variable
	// and reads the value at run time. The environment is created after the
	// grants so its first DAG run can read them.
	storedSecrets, err := envsecrets.Store(ctx, secretPlacement(locals, envName), secretVariables(spec), gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to store the environment's secret values")
	}

	// -- Build config args ------------------------------------------------

	configArgs := &composer.EnvironmentConfigArgs{}
	hasConfig := false

	// -- Node config (networking) -----------------------------------------

	if spec.NodeConfig != nil {
		nc := spec.NodeConfig
		nodeConfigArgs := &composer.EnvironmentConfigNodeConfigArgs{}
		hasNodeConfig := false

		if nc.Network != nil && nc.Network.GetValue() != "" {
			nodeConfigArgs.Network = pulumi.StringPtr(nc.Network.GetValue())
			hasNodeConfig = true
		}
		if nc.Subnetwork != nil && nc.Subnetwork.GetValue() != "" {
			nodeConfigArgs.Subnetwork = pulumi.StringPtr(nc.Subnetwork.GetValue())
			hasNodeConfig = true
		}
		if nc.ServiceAccount != nil && nc.ServiceAccount.GetValue() != "" {
			nodeConfigArgs.ServiceAccount = pulumi.StringPtr(nc.ServiceAccount.GetValue())
			hasNodeConfig = true
		}
		if len(nc.Tags) > 0 {
			nodeConfigArgs.Tags = pulumi.ToStringArray(nc.Tags)
			hasNodeConfig = true
		}
		if nc.ComposerNetworkAttachment != "" {
			nodeConfigArgs.ComposerNetworkAttachment = pulumi.StringPtr(nc.ComposerNetworkAttachment)
			hasNodeConfig = true
		}
		if nc.ComposerInternalIpv4CidrBlock != "" {
			nodeConfigArgs.ComposerInternalIpv4CidrBlock = pulumi.StringPtr(nc.ComposerInternalIpv4CidrBlock)
			hasNodeConfig = true
		}
		if nc.EnableIpMasqAgent {
			nodeConfigArgs.EnableIpMasqAgent = pulumi.BoolPtr(true)
			hasNodeConfig = true
		}

		// VPC-native range assignment: a named secondary range on the
		// subnetwork XOR a CIDR for GKE to carve, per range.
		if nc.IpAllocationPolicy != nil {
			ipAllocArgs := &composer.EnvironmentConfigNodeConfigIpAllocationPolicyArgs{}
			if nc.IpAllocationPolicy.ClusterSecondaryRangeName != "" {
				ipAllocArgs.ClusterSecondaryRangeName = pulumi.StringPtr(nc.IpAllocationPolicy.ClusterSecondaryRangeName)
			}
			if nc.IpAllocationPolicy.ClusterIpv4CidrBlock != "" {
				ipAllocArgs.ClusterIpv4CidrBlock = pulumi.StringPtr(nc.IpAllocationPolicy.ClusterIpv4CidrBlock)
			}
			if nc.IpAllocationPolicy.ServicesSecondaryRangeName != "" {
				ipAllocArgs.ServicesSecondaryRangeName = pulumi.StringPtr(nc.IpAllocationPolicy.ServicesSecondaryRangeName)
			}
			if nc.IpAllocationPolicy.ServicesIpv4CidrBlock != "" {
				ipAllocArgs.ServicesIpv4CidrBlock = pulumi.StringPtr(nc.IpAllocationPolicy.ServicesIpv4CidrBlock)
			}
			nodeConfigArgs.IpAllocationPolicy = ipAllocArgs
			hasNodeConfig = true
		}

		if hasNodeConfig {
			configArgs.NodeConfig = nodeConfigArgs
			hasConfig = true
		}
	}

	// -- Software config --------------------------------------------------

	if spec.SoftwareConfig != nil {
		sc := spec.SoftwareConfig
		softwareConfigArgs := &composer.EnvironmentConfigSoftwareConfigArgs{}
		hasSoftwareConfig := false

		if sc.ImageVersion != "" {
			softwareConfigArgs.ImageVersion = pulumi.StringPtr(sc.ImageVersion)
			hasSoftwareConfig = true
		}
		if len(sc.AirflowConfigOverrides) > 0 {
			softwareConfigArgs.AirflowConfigOverrides = pulumi.ToStringMap(sc.AirflowConfigOverrides)
			hasSoftwareConfig = true
		}
		if len(sc.PypiPackages) > 0 {
			softwareConfigArgs.PypiPackages = pulumi.ToStringMap(sc.PypiPackages)
			hasSoftwareConfig = true
		}
		if len(sc.EnvVariables)+len(sc.SecretEnvVariables) > 0 {
			softwareConfigArgs.EnvVariables = envVariables(sc, storedSecrets.Refs)
			hasSoftwareConfig = true
		}
		if sc.WebServerPluginsMode != "" {
			softwareConfigArgs.WebServerPluginsMode = pulumi.StringPtr(sc.WebServerPluginsMode)
			hasSoftwareConfig = true
		}

		// Automatic dataset lineage reporting into Dataplex.
		if sc.CloudDataLineageIntegration != nil {
			softwareConfigArgs.CloudDataLineageIntegration = &composer.EnvironmentConfigSoftwareConfigCloudDataLineageIntegrationArgs{
				Enabled: pulumi.Bool(sc.CloudDataLineageIntegration.Enabled),
			}
			hasSoftwareConfig = true
		}

		if hasSoftwareConfig {
			configArgs.SoftwareConfig = softwareConfigArgs
			hasConfig = true
		}
	}

	// -- Private environment config (Composer 2.x) ------------------------

	if spec.PrivateEnvironmentConfig != nil {
		pec := spec.PrivateEnvironmentConfig
		privateArgs := &composer.EnvironmentConfigPrivateEnvironmentConfigArgs{}

		if pec.EnablePrivateEndpoint {
			privateArgs.EnablePrivateEndpoint = pulumi.BoolPtr(true)
		}
		if pec.ConnectionType != "" {
			privateArgs.ConnectionType = pulumi.StringPtr(pec.ConnectionType)
		}
		if pec.MasterIpv4CidrBlock != "" {
			privateArgs.MasterIpv4CidrBlock = pulumi.StringPtr(pec.MasterIpv4CidrBlock)
		}
		if pec.CloudSqlIpv4CidrBlock != "" {
			privateArgs.CloudSqlIpv4CidrBlock = pulumi.StringPtr(pec.CloudSqlIpv4CidrBlock)
		}
		if pec.CloudComposerNetworkIpv4CidrBlock != "" {
			privateArgs.CloudComposerNetworkIpv4CidrBlock = pulumi.StringPtr(pec.CloudComposerNetworkIpv4CidrBlock)
		}
		if pec.CloudComposerConnectionSubnetwork != "" {
			privateArgs.CloudComposerConnectionSubnetwork = pulumi.StringPtr(pec.CloudComposerConnectionSubnetwork)
		}
		if pec.EnablePrivatelyUsedPublicIps {
			privateArgs.EnablePrivatelyUsedPublicIps = pulumi.BoolPtr(true)
		}

		configArgs.PrivateEnvironmentConfig = privateArgs
		hasConfig = true
	}

	// -- Workloads config -------------------------------------------------

	if spec.WorkloadsConfig != nil {
		wc := spec.WorkloadsConfig
		workloadsArgs := &composer.EnvironmentConfigWorkloadsConfigArgs{}
		hasWorkloads := false

		if wc.Scheduler != nil {
			schedulerArgs := &composer.EnvironmentConfigWorkloadsConfigSchedulerArgs{}
			if wc.Scheduler.Cpu > 0 {
				schedulerArgs.Cpu = pulumi.Float64Ptr(wc.Scheduler.Cpu)
			}
			if wc.Scheduler.MemoryGb > 0 {
				schedulerArgs.MemoryGb = pulumi.Float64Ptr(wc.Scheduler.MemoryGb)
			}
			if wc.Scheduler.StorageGb > 0 {
				schedulerArgs.StorageGb = pulumi.Float64Ptr(wc.Scheduler.StorageGb)
			}
			if wc.Scheduler.Count > 0 {
				schedulerArgs.Count = pulumi.IntPtr(int(wc.Scheduler.Count))
			}
			workloadsArgs.Scheduler = schedulerArgs
			hasWorkloads = true
		}

		if wc.WebServer != nil {
			webServerArgs := &composer.EnvironmentConfigWorkloadsConfigWebServerArgs{}
			if wc.WebServer.Cpu > 0 {
				webServerArgs.Cpu = pulumi.Float64Ptr(wc.WebServer.Cpu)
			}
			if wc.WebServer.MemoryGb > 0 {
				webServerArgs.MemoryGb = pulumi.Float64Ptr(wc.WebServer.MemoryGb)
			}
			if wc.WebServer.StorageGb > 0 {
				webServerArgs.StorageGb = pulumi.Float64Ptr(wc.WebServer.StorageGb)
			}
			workloadsArgs.WebServer = webServerArgs
			hasWorkloads = true
		}

		if wc.Worker != nil {
			workerArgs := &composer.EnvironmentConfigWorkloadsConfigWorkerArgs{}
			if wc.Worker.Cpu > 0 {
				workerArgs.Cpu = pulumi.Float64Ptr(wc.Worker.Cpu)
			}
			if wc.Worker.MemoryGb > 0 {
				workerArgs.MemoryGb = pulumi.Float64Ptr(wc.Worker.MemoryGb)
			}
			if wc.Worker.StorageGb > 0 {
				workerArgs.StorageGb = pulumi.Float64Ptr(wc.Worker.StorageGb)
			}
			if wc.Worker.MinCount > 0 {
				workerArgs.MinCount = pulumi.IntPtr(int(wc.Worker.MinCount))
			}
			if wc.Worker.MaxCount > 0 {
				workerArgs.MaxCount = pulumi.IntPtr(int(wc.Worker.MaxCount))
			}
			workloadsArgs.Worker = workerArgs
			hasWorkloads = true
		}

		if wc.Triggerer != nil {
			triggererArgs := &composer.EnvironmentConfigWorkloadsConfigTriggererArgs{
				Count:    pulumi.Int(int(wc.Triggerer.Count)),
				Cpu:      pulumi.Float64(wc.Triggerer.Cpu),
				MemoryGb: pulumi.Float64(wc.Triggerer.MemoryGb),
			}
			workloadsArgs.Triggerer = triggererArgs
			hasWorkloads = true
		}

		if wc.DagProcessor != nil {
			dagProcessorArgs := &composer.EnvironmentConfigWorkloadsConfigDagProcessorArgs{}
			if wc.DagProcessor.Cpu > 0 {
				dagProcessorArgs.Cpu = pulumi.Float64Ptr(wc.DagProcessor.Cpu)
			}
			if wc.DagProcessor.MemoryGb > 0 {
				dagProcessorArgs.MemoryGb = pulumi.Float64Ptr(wc.DagProcessor.MemoryGb)
			}
			if wc.DagProcessor.StorageGb > 0 {
				dagProcessorArgs.StorageGb = pulumi.Float64Ptr(wc.DagProcessor.StorageGb)
			}
			if wc.DagProcessor.Count > 0 {
				dagProcessorArgs.Count = pulumi.IntPtr(int(wc.DagProcessor.Count))
			}
			workloadsArgs.DagProcessor = dagProcessorArgs
			hasWorkloads = true
		}

		if hasWorkloads {
			configArgs.WorkloadsConfig = workloadsArgs
			hasConfig = true
		}
	}

	// -- Environment size -------------------------------------------------

	if spec.EnvironmentSize != "" {
		configArgs.EnvironmentSize = pulumi.StringPtr(spec.EnvironmentSize)
		hasConfig = true
	}

	// -- Resilience mode --------------------------------------------------

	if spec.ResilienceMode != "" {
		configArgs.ResilienceMode = pulumi.StringPtr(spec.ResilienceMode)
		hasConfig = true
	}

	// -- Encryption config (CMEK) -----------------------------------------

	if spec.KmsKeyName != nil && spec.KmsKeyName.GetValue() != "" {
		configArgs.EncryptionConfig = &composer.EnvironmentConfigEncryptionConfigArgs{
			KmsKeyName: pulumi.String(spec.KmsKeyName.GetValue()),
		}
		hasConfig = true
	}

	// -- Maintenance window -----------------------------------------------

	if spec.MaintenanceWindow != nil {
		configArgs.MaintenanceWindow = &composer.EnvironmentConfigMaintenanceWindowArgs{
			StartTime:  pulumi.String(spec.MaintenanceWindow.StartTime),
			EndTime:    pulumi.String(spec.MaintenanceWindow.EndTime),
			Recurrence: pulumi.String(spec.MaintenanceWindow.Recurrence),
		}
		hasConfig = true
	}

	// -- Recovery config --------------------------------------------------

	if spec.RecoveryConfig != nil {
		snapshotArgs := &composer.EnvironmentConfigRecoveryConfigScheduledSnapshotsConfigArgs{
			Enabled: pulumi.Bool(spec.RecoveryConfig.Enabled),
		}
		if spec.RecoveryConfig.SnapshotLocation != "" {
			snapshotArgs.SnapshotLocation = pulumi.StringPtr(spec.RecoveryConfig.SnapshotLocation)
		}
		if spec.RecoveryConfig.SnapshotCreationSchedule != "" {
			snapshotArgs.SnapshotCreationSchedule = pulumi.StringPtr(spec.RecoveryConfig.SnapshotCreationSchedule)
		}
		if spec.RecoveryConfig.TimeZone != "" {
			snapshotArgs.TimeZone = pulumi.StringPtr(spec.RecoveryConfig.TimeZone)
		}
		configArgs.RecoveryConfig = &composer.EnvironmentConfigRecoveryConfigArgs{
			ScheduledSnapshotsConfig: snapshotArgs,
		}
		hasConfig = true
	}

	// -- Web server network access control --------------------------------

	if spec.WebServerNetworkAccessControl != nil && len(spec.WebServerNetworkAccessControl.AllowedIpRanges) > 0 {
		var allowedRanges composer.EnvironmentConfigWebServerNetworkAccessControlAllowedIpRangeArray
		for _, ipRange := range spec.WebServerNetworkAccessControl.AllowedIpRanges {
			rangeArgs := &composer.EnvironmentConfigWebServerNetworkAccessControlAllowedIpRangeArgs{
				Value: pulumi.String(ipRange.Value),
			}
			if ipRange.Description != "" {
				rangeArgs.Description = pulumi.StringPtr(ipRange.Description)
			}
			allowedRanges = append(allowedRanges, rangeArgs)
		}
		configArgs.WebServerNetworkAccessControl = &composer.EnvironmentConfigWebServerNetworkAccessControlArgs{
			AllowedIpRanges: allowedRanges,
		}
		hasConfig = true
	}

	// -- Master authorized networks (GKE control-plane allowlist) ----------

	if spec.MasterAuthorizedNetworksConfig != nil {
		masterAuthArgs := &composer.EnvironmentConfigMasterAuthorizedNetworksConfigArgs{
			Enabled: pulumi.Bool(spec.MasterAuthorizedNetworksConfig.Enabled),
		}
		if len(spec.MasterAuthorizedNetworksConfig.CidrBlocks) > 0 {
			var cidrBlocks composer.EnvironmentConfigMasterAuthorizedNetworksConfigCidrBlockArray
			for _, block := range spec.MasterAuthorizedNetworksConfig.CidrBlocks {
				blockArgs := &composer.EnvironmentConfigMasterAuthorizedNetworksConfigCidrBlockArgs{
					CidrBlock: pulumi.String(block.CidrBlock),
				}
				if block.DisplayName != "" {
					blockArgs.DisplayName = pulumi.StringPtr(block.DisplayName)
				}
				cidrBlocks = append(cidrBlocks, blockArgs)
			}
			masterAuthArgs.CidrBlocks = cidrBlocks
		}
		configArgs.MasterAuthorizedNetworksConfig = masterAuthArgs
		hasConfig = true
	}

	// -- Data retention (task logs + Airflow metadata) ----------------------

	if spec.DataRetentionConfig != nil {
		retentionArgs := &composer.EnvironmentConfigDataRetentionConfigArgs{}
		hasRetention := false

		if spec.DataRetentionConfig.TaskLogsStorageMode != "" {
			retentionArgs.TaskLogsRetentionConfigs = composer.EnvironmentConfigDataRetentionConfigTaskLogsRetentionConfigArray{
				&composer.EnvironmentConfigDataRetentionConfigTaskLogsRetentionConfigArgs{
					StorageMode: pulumi.StringPtr(spec.DataRetentionConfig.TaskLogsStorageMode),
				},
			}
			hasRetention = true
		}

		if spec.DataRetentionConfig.AirflowMetadataRetentionMode != "" {
			metadataArgs := &composer.EnvironmentConfigDataRetentionConfigAirflowMetadataRetentionConfigArgs{
				RetentionMode: pulumi.StringPtr(spec.DataRetentionConfig.AirflowMetadataRetentionMode),
			}
			if spec.DataRetentionConfig.AirflowMetadataRetentionDays > 0 {
				metadataArgs.RetentionDays = pulumi.IntPtr(int(spec.DataRetentionConfig.AirflowMetadataRetentionDays))
			}
			retentionArgs.AirflowMetadataRetentionConfigs = composer.EnvironmentConfigDataRetentionConfigAirflowMetadataRetentionConfigArray{metadataArgs}
			hasRetention = true
		}

		if hasRetention {
			configArgs.DataRetentionConfig = retentionArgs
			hasConfig = true
		}
	}

	// -- Composer 3 private environment flags -----------------------------

	if spec.EnablePrivateEnvironment {
		configArgs.EnablePrivateEnvironment = pulumi.BoolPtr(true)
		hasConfig = true
	}
	if spec.EnablePrivateBuildsOnly {
		configArgs.EnablePrivateBuildsOnly = pulumi.BoolPtr(true)
		hasConfig = true
	}

	// -- Create the Composer environment ----------------------------------

	args := &composer.EnvironmentArgs{
		Name:   pulumi.StringPtr(envName),
		Region: pulumi.StringPtr(spec.Region),
		Labels: pulumi.ToStringMap(locals.GcpLabels),
	}

	// Honor the spec contract: an empty project_id falls back to the
	// provider's default project (omitting the argument lets the provider
	// resolve its own project).
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}

	// Client-side destroy behavior: DELETE (default), PREVENT (destroy
	// fails), or ABANDON (drop from state, keep the environment — and its
	// meaningful idle bill — running in GCP). The auto-created DAG bucket
	// survives a DELETE either way.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	// Existing bucket for DAGs/plugins/data instead of the auto-created one.
	if spec.StorageBucket.GetValue() != "" {
		args.StorageConfig = &composer.EnvironmentStorageConfigArgs{
			Bucket: pulumi.String(spec.StorageBucket.GetValue()),
		}
	}

	if hasConfig {
		args.Config = configArgs
	}

	createdEnv, err := composer.NewEnvironment(ctx, "composer-environment", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn(append([]pulumi.Resource{createdComposerApi}, storedSecrets.Grants...)))
	if err != nil {
		return errors.Wrap(err, "failed to create cloud composer environment")
	}

	// -- Outputs ----------------------------------------------------------

	ctx.Export(OpEnvironmentId, createdEnv.ID())
	ctx.Export(OpEnvironmentName, createdEnv.Name)
	ctx.Export(OpAirflowUri, createdEnv.Config.AirflowUri())
	ctx.Export(OpDagGcsPrefix, createdEnv.Config.DagGcsPrefix())
	ctx.Export(OpGkeCluster, createdEnv.Config.GkeCluster())

	return nil
}
