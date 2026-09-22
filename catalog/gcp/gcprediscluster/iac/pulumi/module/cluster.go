package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/redis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Connection types Google stamps on a cluster's service attachments and
// PSC connections, used to pick the scalar handles the outputs expose.
const (
	connectionTypeDiscovery = "CONNECTION_TYPE_DISCOVERY"
	connectionTypePrimary   = "CONNECTION_TYPE_PRIMARY"
	connectionTypeReader    = "CONNECTION_TYPE_READER"
)

// cluster provisions the Memorystore for Redis Cluster. Connectivity is
// Private Service Connect: with psc_configs set, service connectivity
// automation places the endpoints (a GcpServiceConnectionPolicy for the
// gcp-memorystore-redis class must already exist on that network in this
// region); without it, the cluster only publishes service attachments and
// consumers register their own forwarding rules through
// GcpRedisClusterEndpointSet.
//
// The immutables (ForceNew in the provider): name, region,
// authorizationMode, transitEncryptionMode, zoneDistributionConfig, and
// the seed sources. shardCount, replicaCount, nodeType, redisConfigs,
// kmsKey, pscConfigs, persistence, backups, maintenance, the replication
// role, labels, and deletion protection all update in place.
func cluster(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpRedisCluster.Spec

	// Enable the Memorystore for Redis API (the cluster's control plane)
	// and the Network Connectivity API (the automation that places PSC
	// endpoints). disable_on_destroy stays false: tearing down one cluster
	// must never disable the APIs for everything else in the project.
	redisApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("redis.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	networkConnectivityApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("networkconnectivity.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		redisApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		networkConnectivityApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdRedisApi, err := projects.NewService(ctx,
		"rcl-redis.googleapis.com", redisApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable redis.googleapis.com api")
	}
	createdNetworkConnectivityApi, err := projects.NewService(ctx,
		"rcl-networkconnectivity.googleapis.com", networkConnectivityApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable networkconnectivity.googleapis.com api")
	}

	// Deletion guard (spec default TRUE): always sent explicitly so destroy
	// behavior is identical on both engines -- a manifest that never
	// mentions deletion protection must behave the same everywhere.
	deletionProtection := true
	if spec.DeletionProtectionEnabled != nil {
		deletionProtection = spec.GetDeletionProtectionEnabled()
	}

	// The two immutable security modes are sent explicitly with Google's
	// defaults when the spec leaves them unset, so the cluster's posture
	// never depends on a provider default (identical to the Terraform
	// module).
	authorizationMode := spec.AuthorizationMode
	if authorizationMode == "" {
		authorizationMode = "AUTH_MODE_DISABLED"
	}
	transitEncryptionMode := spec.TransitEncryptionMode
	if transitEncryptionMode == "" {
		transitEncryptionMode = "TRANSIT_ENCRYPTION_MODE_DISABLED"
	}

	args := &redis.ClusterArgs{
		Name:       pulumi.String(locals.ClusterName),
		Region:     pulumi.String(spec.Region),
		ShardCount: pulumi.Int(int(spec.ShardCount)),
		Labels:     pulumi.ToStringMap(locals.GcpLabels),

		// 0 is an explicit "no replicas" -- always sent so the manifest
		// value is authoritative (identical to the Terraform module).
		ReplicaCount: pulumi.IntPtr(int(spec.ReplicaCount)),

		AuthorizationMode:         pulumi.StringPtr(authorizationMode),
		TransitEncryptionMode:     pulumi.StringPtr(transitEncryptionMode),
		DeletionProtectionEnabled: pulumi.BoolPtr(deletionProtection),
	}

	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}

	// Optional+Computed levers are sent only when set so Google's own
	// defaults stay in charge otherwise.
	if spec.NodeType != "" {
		args.NodeType = pulumi.StringPtr(spec.NodeType)
	}
	if len(spec.RedisConfigs) > 0 {
		args.RedisConfigs = pulumi.ToStringMap(spec.RedisConfigs)
	}

	// Google-placed PSC endpoints. network arrives as the VPC's relative
	// resource path (the GcpVpcNetwork network_id output) -- the only form
	// the Service Connectivity API accepts.
	if len(spec.PscConfigs) > 0 {
		pscConfigs := redis.ClusterPscConfigArray{}
		for _, psc := range spec.PscConfigs {
			pscConfigs = append(pscConfigs, &redis.ClusterPscConfigArgs{
				Network: pulumi.String(psc.Network.GetValue()),
			})
		}
		args.PscConfigs = pscConfigs
	}

	// Server certificate authority for the TLS-enabled cluster; a
	// customer-managed CA pool is consumed only in
	// SERVER_CA_MODE_CUSTOMER_MANAGED_CAS_CA mode (spec-enforced pairing).
	if spec.ServerCaMode != "" {
		args.ServerCaMode = pulumi.StringPtr(spec.ServerCaMode)
	}
	if spec.ServerCaPool != "" {
		args.ServerCaPool = pulumi.StringPtr(spec.ServerCaPool)
	}

	if spec.KmsKey.GetValue() != "" {
		args.KmsKey = pulumi.StringPtr(spec.KmsKey.GetValue())
	}

	// Self-service maintenance: a newer available version applies the
	// update now instead of waiting for Google's rollout. Update-only and
	// forward-only at the API.
	if spec.MaintenanceVersion != "" {
		args.MaintenanceVersion = pulumi.StringPtr(spec.MaintenanceVersion)
	}

	// Shared Redis ACL policy attached by full resource name. Sent only
	// when set so a cluster without one keeps its built-in default ACL.
	if spec.AclPolicy != "" {
		args.AclPolicy = pulumi.StringPtr(spec.AclPolicy)
	}

	// Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
	// Sent only when set so the provider default stays in charge otherwise;
	// evaluated only after deletion_protection_enabled allows the destroy.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	// Persistence (RDB snapshots or AOF log); every leaf is
	// Optional+Computed, so each is sent only when set.
	if spec.PersistenceConfig != nil {
		persistenceArgs := &redis.ClusterPersistenceConfigArgs{}
		if spec.PersistenceConfig.Mode != "" {
			persistenceArgs.Mode = pulumi.StringPtr(spec.PersistenceConfig.Mode)
		}
		if spec.PersistenceConfig.RdbConfig != nil {
			rdbArgs := &redis.ClusterPersistenceConfigRdbConfigArgs{}
			if spec.PersistenceConfig.RdbConfig.RdbSnapshotPeriod != "" {
				rdbArgs.RdbSnapshotPeriod = pulumi.StringPtr(spec.PersistenceConfig.RdbConfig.RdbSnapshotPeriod)
			}
			if spec.PersistenceConfig.RdbConfig.RdbSnapshotStartTime != "" {
				rdbArgs.RdbSnapshotStartTime = pulumi.StringPtr(spec.PersistenceConfig.RdbConfig.RdbSnapshotStartTime)
			}
			persistenceArgs.RdbConfig = rdbArgs
		}
		if spec.PersistenceConfig.AofConfig != nil {
			aofArgs := &redis.ClusterPersistenceConfigAofConfigArgs{}
			if spec.PersistenceConfig.AofConfig.AppendFsync != "" {
				aofArgs.AppendFsync = pulumi.StringPtr(spec.PersistenceConfig.AofConfig.AppendFsync)
			}
			persistenceArgs.AofConfig = aofArgs
		}
		args.PersistenceConfig = persistenceArgs
	}

	// Zone distribution (immutable).
	if spec.ZoneDistributionConfig != nil {
		zdcArgs := &redis.ClusterZoneDistributionConfigArgs{}
		if spec.ZoneDistributionConfig.Mode != "" {
			zdcArgs.Mode = pulumi.StringPtr(spec.ZoneDistributionConfig.Mode)
		}
		if spec.ZoneDistributionConfig.Zone != "" {
			zdcArgs.Zone = pulumi.StringPtr(spec.ZoneDistributionConfig.Zone)
		}
		args.ZoneDistributionConfig = zdcArgs
	}

	// Weekly maintenance window. Hours only: Google exposes the Redis
	// Cluster window start by the hour (gcloud has no minute flag), so the
	// TimeOfDay's finer fields are never sent.
	if spec.MaintenancePolicy != nil && spec.MaintenancePolicy.WeeklyMaintenanceWindow != nil {
		args.MaintenancePolicy = &redis.ClusterMaintenancePolicyArgs{
			WeeklyMaintenanceWindows: redis.ClusterMaintenancePolicyWeeklyMaintenanceWindowArray{
				&redis.ClusterMaintenancePolicyWeeklyMaintenanceWindowArgs{
					Day: pulumi.String(spec.MaintenancePolicy.WeeklyMaintenanceWindow.Day),
					StartTime: &redis.ClusterMaintenancePolicyWeeklyMaintenanceWindowStartTimeArgs{
						Hours: pulumi.IntPtr(int(spec.MaintenancePolicy.WeeklyMaintenanceWindow.Hour)),
					},
				},
			},
		}
	}

	// Daily automated backups into the cluster's managed backup
	// collection.
	if spec.AutomatedBackupConfig != nil {
		args.AutomatedBackupConfig = &redis.ClusterAutomatedBackupConfigArgs{
			Retention: pulumi.String(spec.AutomatedBackupConfig.Retention),
			FixedFrequencySchedule: &redis.ClusterAutomatedBackupConfigFixedFrequencyScheduleArgs{
				StartTime: &redis.ClusterAutomatedBackupConfigFixedFrequencyScheduleStartTimeArgs{
					Hours: pulumi.Int(int(spec.AutomatedBackupConfig.StartHour)),
				},
			},
		}
	}

	// Cross-region DR: PRIMARY lists its secondaries; SECONDARY names its
	// primary. Cluster references arrive as full resource paths (the other
	// cluster's name output).
	if spec.CrossClusterReplicationConfig != nil {
		ccrArgs := &redis.ClusterCrossClusterReplicationConfigArgs{}
		if spec.CrossClusterReplicationConfig.ClusterRole != "" {
			ccrArgs.ClusterRole = pulumi.StringPtr(spec.CrossClusterReplicationConfig.ClusterRole)
		}
		if spec.CrossClusterReplicationConfig.PrimaryCluster != nil {
			ccrArgs.PrimaryCluster = &redis.ClusterCrossClusterReplicationConfigPrimaryClusterArgs{
				Cluster: pulumi.StringPtr(spec.CrossClusterReplicationConfig.PrimaryCluster.Cluster.GetValue()),
			}
		}
		if len(spec.CrossClusterReplicationConfig.SecondaryClusters) > 0 {
			secondaries := redis.ClusterCrossClusterReplicationConfigSecondaryClusterArray{}
			for _, s := range spec.CrossClusterReplicationConfig.SecondaryClusters {
				secondaries = append(secondaries, &redis.ClusterCrossClusterReplicationConfigSecondaryClusterArgs{
					Cluster: pulumi.StringPtr(s.Cluster.GetValue()),
				})
			}
			ccrArgs.SecondaryClusters = secondaries
		}
		args.CrossClusterReplicationConfig = ccrArgs
	}

	// Seed sources (mutually exclusive, ForceNew -- seeding happens once).
	if spec.GcsSource != nil {
		uris := make(pulumi.StringArray, 0, len(spec.GcsSource.Uris))
		for _, u := range spec.GcsSource.Uris {
			uris = append(uris, pulumi.String(u))
		}
		args.GcsSource = &redis.ClusterGcsSourceArgs{Uris: uris}
	}
	if spec.ManagedBackupSource != nil {
		args.ManagedBackupSource = &redis.ClusterManagedBackupSourceArgs{
			Backup: pulumi.String(spec.ManagedBackupSource.Backup),
		}
	}

	createdCluster, err := redis.NewCluster(ctx, "redis-cluster", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdRedisApi, createdNetworkConnectivityApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create redis cluster")
	}

	// The Google-placed discovery endpoint: present only when psc_configs
	// asked for automatic connectivity; empty otherwise.
	discoveryAddress := createdCluster.DiscoveryEndpoints.ApplyT(func(endpoints []redis.ClusterDiscoveryEndpoint) string {
		for _, ep := range endpoints {
			if ep.Address != nil {
				return *ep.Address
			}
		}
		return ""
	}).(pulumi.StringOutput)
	discoveryPort := createdCluster.DiscoveryEndpoints.ApplyT(func(endpoints []redis.ClusterDiscoveryEndpoint) int {
		for _, ep := range endpoints {
			if ep.Port != nil {
				return *ep.Port
			}
		}
		return 0
	}).(pulumi.IntOutput)

	// One scalar handle per connection type, so a consumer's forwarding
	// rule (and GcpRedisClusterEndpointSet) can reference the attachment it
	// needs -- a reference cannot index a repeated output.
	serviceAttachmentOfType := func(connectionType string) pulumi.StringOutput {
		return createdCluster.PscServiceAttachments.ApplyT(func(attachments []redis.ClusterPscServiceAttachment) string {
			for _, a := range attachments {
				if a.ConnectionType != nil && *a.ConnectionType == connectionType && a.ServiceAttachment != nil {
					return *a.ServiceAttachment
				}
			}
			return ""
		}).(pulumi.StringOutput)
	}

	ctx.Export(OpName, createdCluster.Name)
	ctx.Export(OpUid, createdCluster.Uid)
	ctx.Export(OpState, createdCluster.State)
	ctx.Export(OpDiscoveryEndpointAddress, discoveryAddress)
	ctx.Export(OpDiscoveryEndpointPort, discoveryPort)
	ctx.Export(OpDiscoveryServiceAttachment, serviceAttachmentOfType(connectionTypeDiscovery))
	ctx.Export(OpPrimaryServiceAttachment, serviceAttachmentOfType(connectionTypePrimary))
	ctx.Export(OpReaderServiceAttachment, serviceAttachmentOfType(connectionTypeReader))
	ctx.Export(OpSizeGb, createdCluster.SizeGb)
	ctx.Export(OpShardCount, createdCluster.ShardCount)
	ctx.Export(OpReplicaCount, createdCluster.ReplicaCount.Elem())
	ctx.Export(OpBackupCollection, createdCluster.BackupCollection)

	return nil
}
