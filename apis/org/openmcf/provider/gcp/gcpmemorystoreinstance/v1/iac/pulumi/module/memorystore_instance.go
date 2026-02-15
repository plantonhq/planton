package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/memorystore"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func memorystoreInstance(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpMemorystoreInstance.Spec

	args := &memorystore.InstanceArgs{
		InstanceId: pulumi.String(spec.InstanceName),
		Location:   pulumi.String(spec.Location),
		ShardCount: pulumi.Int(int(spec.ShardCount)),
		Labels:     pulumi.ToStringMap(locals.GcpLabels),
	}

	// Project.
	if spec.ProjectId != nil && spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}

	// Instance mode.
	if spec.Mode != "" {
		args.Mode = pulumi.StringPtr(spec.Mode)
	}

	// Node type.
	if spec.NodeType != "" {
		args.NodeType = pulumi.StringPtr(spec.NodeType)
	}

	// Engine version.
	if spec.EngineVersion != "" {
		args.EngineVersion = pulumi.StringPtr(spec.EngineVersion)
	}

	// Engine configs.
	if len(spec.EngineConfigs) > 0 {
		args.EngineConfigs = pulumi.ToStringMap(spec.EngineConfigs)
	}

	// Replica count.
	if spec.ReplicaCount > 0 {
		args.ReplicaCount = pulumi.IntPtr(int(spec.ReplicaCount))
	}

	// PSC auto-created endpoints.
	if len(spec.PscAutoConnections) > 0 {
		endpoints := memorystore.InstanceDesiredAutoCreatedEndpointArray{}
		for _, psc := range spec.PscAutoConnections {
			endpoints = append(endpoints, &memorystore.InstanceDesiredAutoCreatedEndpointArgs{
				Network:   pulumi.String(psc.Network.GetValue()),
				ProjectId: pulumi.String(psc.ProjectId.GetValue()),
			})
		}
		args.DesiredAutoCreatedEndpoints = endpoints
	}

	// Authorization mode.
	if spec.AuthorizationMode != "" {
		args.AuthorizationMode = pulumi.StringPtr(spec.AuthorizationMode)
	}

	// Transit encryption mode.
	if spec.TransitEncryptionMode != "" {
		args.TransitEncryptionMode = pulumi.StringPtr(spec.TransitEncryptionMode)
	}

	// CMEK encryption.
	if spec.KmsKey != nil && spec.KmsKey.GetValue() != "" {
		args.KmsKey = pulumi.StringPtr(spec.KmsKey.GetValue())
	}

	// Persistence configuration.
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

	// Maintenance policy.
	if spec.MaintenancePolicy != nil && spec.MaintenancePolicy.WeeklyMaintenanceWindow != nil {
		args.MaintenancePolicy = &memorystore.InstanceMaintenancePolicyArgs{
			WeeklyMaintenanceWindows: memorystore.InstanceMaintenancePolicyWeeklyMaintenanceWindowArray{
				&memorystore.InstanceMaintenancePolicyWeeklyMaintenanceWindowArgs{
					Day: pulumi.String(spec.MaintenancePolicy.WeeklyMaintenanceWindow.Day),
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

	// Deletion protection.
	if spec.DeletionProtectionEnabled {
		args.DeletionProtectionEnabled = pulumi.BoolPtr(true)
	}

	createdInstance, err := memorystore.NewInstance(ctx, "memorystore-instance", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create memorystore instance")
	}

	// Extract discovery endpoint from the PSC endpoint connections.
	// The Endpoints output is a nested structure:
	//   Endpoints -> Connections -> PscAutoConnection -> {IpAddress, Port, ConnectionType}
	// We find the first endpoint with connection_type CONNECTION_TYPE_DISCOVERY,
	// falling back to any available connection.
	discoveryAddress := createdInstance.Endpoints.ApplyT(func(endpoints []memorystore.InstanceEndpoint) string {
		for _, ep := range endpoints {
			for _, conn := range ep.Connections {
				if conn.PscAutoConnection != nil {
					if conn.PscAutoConnection.ConnectionType != nil &&
						*conn.PscAutoConnection.ConnectionType == "CONNECTION_TYPE_DISCOVERY" {
						if conn.PscAutoConnection.IpAddress != nil {
							return *conn.PscAutoConnection.IpAddress
						}
					}
				}
			}
		}
		// Fallback: return the first available IP from any connection.
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
				if conn.PscAutoConnection != nil {
					if conn.PscAutoConnection.ConnectionType != nil &&
						*conn.PscAutoConnection.ConnectionType == "CONNECTION_TYPE_DISCOVERY" {
						if conn.PscAutoConnection.Port != nil {
							return *conn.PscAutoConnection.Port
						}
					}
				}
			}
		}
		// Fallback: return the first available port from any connection.
		for _, ep := range endpoints {
			for _, conn := range ep.Connections {
				if conn.PscAutoConnection != nil && conn.PscAutoConnection.Port != nil {
					return *conn.PscAutoConnection.Port
				}
			}
		}
		return 0
	}).(pulumi.IntOutput)

	// Extract node memory size from node_config.
	nodeSizeGb := createdInstance.NodeConfigs.ApplyT(func(configs []memorystore.InstanceNodeConfig) float64 {
		if len(configs) > 0 && configs[0].SizeGb != nil {
			return *configs[0].SizeGb
		}
		return 0
	}).(pulumi.Float64Output)

	ctx.Export(OpDiscoveryAddress, discoveryAddress)
	ctx.Export(OpDiscoveryPort, pulumi.Sprintf("%d", discoveryPort))
	ctx.Export(OpInstanceUid, createdInstance.Uid)
	ctx.Export(OpNodeSizeGb, pulumi.Sprintf("%v", nodeSizeGb))

	// Log the resource name for debugging.
	ctx.Log.Info(fmt.Sprintf("Created Memorystore instance: %s", spec.InstanceName), nil)

	return nil
}
