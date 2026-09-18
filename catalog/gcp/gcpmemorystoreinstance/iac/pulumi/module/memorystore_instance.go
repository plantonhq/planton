package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/memorystore"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// memorystoreInstance provisions the Memorystore (Valkey) instance.
// Connectivity is PSC-only and driven by service connectivity
// automation: a service connection policy for the gcp-memorystore class
// must already exist on each endpoint's network in this region, or
// creation fails with a connectivity error — the policy is a separate
// first-class resource, deployed before this one.
//
// The immutables (ForceNew in the provider): instanceId, location, mode,
// authorizationMode, transitEncryptionMode, kmsKey,
// zoneDistributionConfig, the PSC endpoints, and the seed sources.
// shardCount and replicaCount resize in place; engineConfigs,
// persistence, maintenance, backups, labels, and the DR role update in
// place too.
func memorystoreInstance(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpMemorystoreInstance.Spec

	// Enable the Memorystore API (the instance's control plane) and the
	// Network Connectivity API (the automation that places the PSC
	// endpoints). disable_on_destroy stays false: tearing down one
	// instance must never disable the APIs for everything else in the
	// project.
	memorystoreApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("memorystore.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	networkConnectivityApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("networkconnectivity.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		memorystoreApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		networkConnectivityApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdMemorystoreApi, err := projects.NewService(ctx,
		"msi-memorystore.googleapis.com", memorystoreApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable memorystore.googleapis.com api")
	}
	createdNetworkConnectivityApi, err := projects.NewService(ctx,
		"msi-networkconnectivity.googleapis.com", networkConnectivityApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable networkconnectivity.googleapis.com api")
	}

	// Deletion guard (spec default TRUE): always sent explicitly so
	// destroy behavior is identical on both engines — omitting it would
	// let the provider default decide, and a manifest that never
	// mentions deletion protection must behave the same everywhere.
	deletionProtection := true
	if spec.DeletionProtectionEnabled != nil {
		deletionProtection = spec.GetDeletionProtectionEnabled()
	}

	args := &memorystore.InstanceArgs{
		InstanceId: pulumi.String(spec.InstanceName),
		Location:   pulumi.String(spec.Location),
		ShardCount: pulumi.Int(int(spec.ShardCount)),
		Labels:     pulumi.ToStringMap(locals.GcpLabels),

		// 0 is an explicit "no replicas" — always sent so the manifest
		// value is authoritative (identical to the Terraform module).
		ReplicaCount: pulumi.IntPtr(int(spec.ReplicaCount)),

		DeletionProtectionEnabled: pulumi.BoolPtr(deletionProtection),
	}

	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}
	if spec.Mode != "" {
		args.Mode = pulumi.StringPtr(spec.Mode)
	}
	if spec.NodeType != "" {
		args.NodeType = pulumi.StringPtr(spec.NodeType)
	}
	if spec.EngineVersion != "" {
		args.EngineVersion = pulumi.StringPtr(spec.EngineVersion)
	}
	if len(spec.EngineConfigs) > 0 {
		args.EngineConfigs = pulumi.ToStringMap(spec.EngineConfigs)
	}

	// PSC auto-created endpoints. network arrives as the VPC's relative
	// resource path — the only format the Service Connectivity API
	// accepts (full https:// self-links are rejected). An entry that
	// omits its consumer project rides the provider's effective project
	// (the common same-project case) — resolved once, mirroring the
	// Terraform module's data.google_project lookup.
	if len(spec.PscAutoConnections) > 0 {
		effectiveProject := spec.ProjectId.GetValue()
		needsAmbientProject := false
		for _, psc := range spec.PscAutoConnections {
			if psc.ProjectId.GetValue() == "" {
				needsAmbientProject = true
			}
		}
		if effectiveProject == "" && needsAmbientProject {
			clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
			if err != nil {
				return errors.Wrap(err, "failed to resolve ambient project from provider config")
			}
			effectiveProject = clientConfig.Project
		}

		endpoints := memorystore.InstanceDesiredAutoCreatedEndpointArray{}
		for _, psc := range spec.PscAutoConnections {
			entryProject := psc.ProjectId.GetValue()
			if entryProject == "" {
				entryProject = effectiveProject
			}
			endpoints = append(endpoints, &memorystore.InstanceDesiredAutoCreatedEndpointArgs{
				Network:   pulumi.String(psc.Network.GetValue()),
				ProjectId: pulumi.String(entryProject),
			})
		}
		args.DesiredAutoCreatedEndpoints = endpoints
	}

	if spec.AuthorizationMode != "" {
		args.AuthorizationMode = pulumi.StringPtr(spec.AuthorizationMode)
	}
	if spec.TransitEncryptionMode != "" {
		args.TransitEncryptionMode = pulumi.StringPtr(spec.TransitEncryptionMode)
	}
	if spec.KmsKey.GetValue() != "" {
		args.KmsKey = pulumi.StringPtr(spec.KmsKey.GetValue())
	}

	// Server certificate authority for the TLS-enabled instance: which CA
	// signs the certificate clients verify. A customer-managed CA pool is
	// consumed only in CUSTOMER_MANAGED_CAS_CA mode (spec-enforced
	// pairing). Both are fixed at creation (ForceNew).
	if spec.ServerCaMode != "" {
		args.ServerCaMode = pulumi.StringPtr(spec.ServerCaMode)
	}
	if spec.ServerCaPool != "" {
		args.ServerCaPool = pulumi.StringPtr(spec.ServerCaPool)
	}

	// Self-service maintenance: setting a newer available version applies
	// the update now instead of waiting for GCP's rollout. Update-only —
	// the API rejects it at create and rejects downgrades.
	if spec.MaintenanceVersion != "" {
		args.MaintenanceVersion = pulumi.StringPtr(spec.MaintenanceVersion)
	}

	// Shared Valkey ACL policy attached by full resource name. Sent only
	// when set so an instance without one keeps its built-in default ACL
	// (identical to the Terraform module); swapping is an in-place update.
	if spec.AclPolicy != "" {
		args.AclPolicy = pulumi.StringPtr(spec.AclPolicy)
	}

	// Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
	// Sent only when set so the provider default stays in charge otherwise.
	// Evaluated only after deletion_protection_enabled allows the destroy.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	// Persistence configuration (RDB or AOF).
	if spec.PersistenceConfig != nil {
		persistenceArgs := &memorystore.InstancePersistenceConfigArgs{
			Mode: pulumi.StringPtr(spec.PersistenceConfig.Mode),
		}
		if spec.PersistenceConfig.RdbConfig != nil {
			rdbArgs := &memorystore.InstancePersistenceConfigRdbConfigArgs{
				RdbSnapshotPeriod: pulumi.StringPtr(spec.PersistenceConfig.RdbConfig.RdbSnapshotPeriod),
			}
			if spec.PersistenceConfig.RdbConfig.RdbSnapshotStartTime != "" {
				rdbArgs.RdbSnapshotStartTime = pulumi.StringPtr(spec.PersistenceConfig.RdbConfig.RdbSnapshotStartTime)
			}
			persistenceArgs.RdbConfig = rdbArgs
		}
		if spec.PersistenceConfig.AofConfig != nil {
			persistenceArgs.AofConfig = &memorystore.InstancePersistenceConfigAofConfigArgs{
				AppendFsync: pulumi.StringPtr(spec.PersistenceConfig.AofConfig.AppendFsync),
			}
		}
		args.PersistenceConfig = persistenceArgs
	}

	// Zone distribution configuration.
	if spec.ZoneDistributionConfig != nil {
		zdcArgs := &memorystore.InstanceZoneDistributionConfigArgs{
			Mode: pulumi.StringPtr(spec.ZoneDistributionConfig.Mode),
		}
		if spec.ZoneDistributionConfig.Zone != "" {
			zdcArgs.Zone = pulumi.StringPtr(spec.ZoneDistributionConfig.Zone)
		}
		args.ZoneDistributionConfig = zdcArgs
	}

	// Maintenance policy with weekly maintenance window.
	if spec.MaintenancePolicy != nil && spec.MaintenancePolicy.WeeklyMaintenanceWindow != nil {
		args.MaintenancePolicy = &memorystore.InstanceMaintenancePolicyArgs{
			WeeklyMaintenanceWindows: memorystore.InstanceMaintenancePolicyWeeklyMaintenanceWindowArray{
				&memorystore.InstanceMaintenancePolicyWeeklyMaintenanceWindowArgs{
					Day: pulumi.String(spec.MaintenancePolicy.WeeklyMaintenanceWindow.Day),
					// Hours only: the Memorystore API rejects a start_time
					// carrying minutes (400 "Invalid start time, only hours
					// are supported" — live-verified; narrower than Redis).
					StartTime: &memorystore.InstanceMaintenancePolicyWeeklyMaintenanceWindowStartTimeArgs{
						Hours: pulumi.IntPtr(int(spec.MaintenancePolicy.WeeklyMaintenanceWindow.Hour)),
					},
				},
			},
		}
	}

	// Automated backup configuration.
	if spec.AutomatedBackupConfig != nil {
		args.AutomatedBackupConfig = &memorystore.InstanceAutomatedBackupConfigArgs{
			Retention: pulumi.String(spec.AutomatedBackupConfig.Retention),
			FixedFrequencySchedule: &memorystore.InstanceAutomatedBackupConfigFixedFrequencyScheduleArgs{
				StartTime: &memorystore.InstanceAutomatedBackupConfigFixedFrequencyScheduleStartTimeArgs{
					Hours: pulumi.Int(int(spec.AutomatedBackupConfig.StartHour)),
				},
			},
		}
	}

	// Cross-region DR: PRIMARY lists its secondaries; SECONDARY points
	// at its primary. Roles are exchanged in place during a planned
	// switchover. Instance references arrive as full resource paths (the
	// other instance's name output).
	if spec.CrossInstanceReplicationConfig != nil {
		circArgs := &memorystore.InstanceCrossInstanceReplicationConfigArgs{
			InstanceRole: pulumi.StringPtr(spec.CrossInstanceReplicationConfig.InstanceRole),
		}
		if spec.CrossInstanceReplicationConfig.PrimaryInstance != nil {
			circArgs.PrimaryInstance = &memorystore.InstanceCrossInstanceReplicationConfigPrimaryInstanceArgs{
				Instance: pulumi.StringPtr(spec.CrossInstanceReplicationConfig.PrimaryInstance.Instance.GetValue()),
			}
		}
		if len(spec.CrossInstanceReplicationConfig.SecondaryInstances) > 0 {
			secondaries := memorystore.InstanceCrossInstanceReplicationConfigSecondaryInstanceArray{}
			for _, s := range spec.CrossInstanceReplicationConfig.SecondaryInstances {
				secondaries = append(secondaries, &memorystore.InstanceCrossInstanceReplicationConfigSecondaryInstanceArgs{
					Instance: pulumi.StringPtr(s.Instance.GetValue()),
				})
			}
			circArgs.SecondaryInstances = secondaries
		}
		args.CrossInstanceReplicationConfig = circArgs
	}

	// Seed sources (mutually exclusive, ForceNew — seeding only happens
	// at creation).
	if spec.GcsSource != nil {
		uris := make(pulumi.StringArray, 0, len(spec.GcsSource.Uris))
		for _, u := range spec.GcsSource.Uris {
			uris = append(uris, pulumi.String(u))
		}
		args.GcsSource = &memorystore.InstanceGcsSourceArgs{Uris: uris}
	}
	if spec.ManagedBackupSource != nil {
		args.ManagedBackupSource = &memorystore.InstanceManagedBackupSourceArgs{
			Backup: pulumi.String(spec.ManagedBackupSource.Backup),
		}
	}

	createdInstance, err := memorystore.NewInstance(ctx, "memorystore-instance", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdMemorystoreApi, createdNetworkConnectivityApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create memorystore instance")
	}

	// Extract the discovery endpoint from the PSC endpoint connections:
	// prefer the connection GCP marks CONNECTION_TYPE_DISCOVERY, falling
	// back to any available connection (single-connection instances
	// report only the primary).
	discoveryAddress := createdInstance.Endpoints.ApplyT(func(endpoints []memorystore.InstanceEndpoint) string {
		for _, ep := range endpoints {
			for _, conn := range ep.Connections {
				if conn.PscAutoConnection != nil &&
					conn.PscAutoConnection.ConnectionType != nil &&
					*conn.PscAutoConnection.ConnectionType == "CONNECTION_TYPE_DISCOVERY" &&
					conn.PscAutoConnection.IpAddress != nil {
					return *conn.PscAutoConnection.IpAddress
				}
			}
		}
		for _, ep := range endpoints {
			for _, conn := range ep.Connections {
				if conn.PscAutoConnection != nil && conn.PscAutoConnection.IpAddress != nil {
					return *conn.PscAutoConnection.IpAddress
				}
			}
		}
		return ""
	}).(pulumi.StringOutput)

	discoveryPort := createdInstance.Endpoints.ApplyT(func(endpoints []memorystore.InstanceEndpoint) int {
		for _, ep := range endpoints {
			for _, conn := range ep.Connections {
				if conn.PscAutoConnection != nil &&
					conn.PscAutoConnection.ConnectionType != nil &&
					*conn.PscAutoConnection.ConnectionType == "CONNECTION_TYPE_DISCOVERY" &&
					conn.PscAutoConnection.Port != nil {
					return *conn.PscAutoConnection.Port
				}
			}
		}
		for _, ep := range endpoints {
			for _, conn := range ep.Connections {
				if conn.PscAutoConnection != nil && conn.PscAutoConnection.Port != nil {
					return *conn.PscAutoConnection.Port
				}
			}
		}
		return 0
	}).(pulumi.IntOutput)

	// Memory per node is a consequence of node_type — reported for
	// capacity planning rather than configured.
	nodeSizeGb := createdInstance.NodeConfigs.ApplyT(func(configs []memorystore.InstanceNodeConfig) float64 {
		if len(configs) > 0 && configs[0].SizeGb != nil {
			return *configs[0].SizeGb
		}
		return 0
	}).(pulumi.Float64Output)

	ctx.Export(OpDiscoveryAddress, discoveryAddress)
	ctx.Export(OpDiscoveryPort, discoveryPort)
	ctx.Export(OpInstanceUid, createdInstance.Uid)
	ctx.Export(OpNodeSizeGb, nodeSizeGb)
	ctx.Export(OpName, createdInstance.Name)
	ctx.Export(OpBackupCollection, createdInstance.BackupCollection)

	return nil
}
