package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/redis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func redisInstance(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpRedisInstance.Spec

	args := &redis.InstanceArgs{
		Name:         pulumi.String(spec.InstanceName),
		Project:      pulumi.StringPtr(spec.ProjectId.GetValue()),
		Region:       pulumi.StringPtr(spec.Region),
		Tier:         pulumi.StringPtr(spec.Tier),
		MemorySizeGb: pulumi.Int(int(spec.MemorySizeGb)),
		Labels:       pulumi.ToStringMap(locals.GcpLabels),
	}

	// Redis version.
	if spec.RedisVersion != "" {
		args.RedisVersion = pulumi.StringPtr(spec.RedisVersion)
	}

	// Display name.
	if spec.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(spec.DisplayName)
	}

	// Zone placement.
	if spec.LocationId != "" {
		args.LocationId = pulumi.StringPtr(spec.LocationId)
	}

	// VPC network.
	if spec.AuthorizedNetwork != nil && spec.AuthorizedNetwork.GetValue() != "" {
		args.AuthorizedNetwork = pulumi.StringPtr(spec.AuthorizedNetwork.GetValue())
	}

	// Connection mode.
	if spec.ConnectMode != "" {
		args.ConnectMode = pulumi.StringPtr(spec.ConnectMode)
	}

	// Reserved IP range.
	if spec.ReservedIpRange != "" {
		args.ReservedIpRange = pulumi.StringPtr(spec.ReservedIpRange)
	}

	// Redis AUTH.
	if spec.AuthEnabled {
		args.AuthEnabled = pulumi.BoolPtr(true)
	}

	// Transit encryption.
	if spec.TransitEncryptionMode != "" {
		args.TransitEncryptionMode = pulumi.StringPtr(spec.TransitEncryptionMode)
	}

	// Redis configuration parameters.
	if len(spec.RedisConfigs) > 0 {
		args.RedisConfigs = pulumi.ToStringMap(spec.RedisConfigs)
	}

	// Maintenance policy.
	if spec.MaintenanceWindow != nil {
		args.MaintenancePolicy = &redis.InstanceMaintenancePolicyArgs{
			WeeklyMaintenanceWindows: redis.InstanceMaintenancePolicyWeeklyMaintenanceWindowArray{
				&redis.InstanceMaintenancePolicyWeeklyMaintenanceWindowArgs{
					Day: pulumi.String(spec.MaintenanceWindow.Day),
					StartTime: &redis.InstanceMaintenancePolicyWeeklyMaintenanceWindowStartTimeArgs{
						Hours: pulumi.IntPtr(int(spec.MaintenanceWindow.Hour)),
					},
				},
			},
		}
	}

	// Read replicas.
	if spec.ReadReplicasMode != "" {
		args.ReadReplicasMode = pulumi.StringPtr(spec.ReadReplicasMode)
	}
	if spec.ReplicaCount > 0 {
		args.ReplicaCount = pulumi.IntPtr(int(spec.ReplicaCount))
	}

	// Persistence configuration.
	if spec.PersistenceConfig != nil {
		persistenceArgs := &redis.InstancePersistenceConfigArgs{
			PersistenceMode: pulumi.StringPtr(spec.PersistenceConfig.PersistenceMode),
		}
		if spec.PersistenceConfig.RdbSnapshotPeriod != "" {
			persistenceArgs.RdbSnapshotPeriod = pulumi.StringPtr(spec.PersistenceConfig.RdbSnapshotPeriod)
		}
		args.PersistenceConfig = persistenceArgs
	}

	// CMEK encryption.
	if spec.CustomerManagedKey != nil && spec.CustomerManagedKey.GetValue() != "" {
		args.CustomerManagedKey = pulumi.StringPtr(spec.CustomerManagedKey.GetValue())
	}

	// Deletion protection.
	if spec.DeletionProtection {
		args.DeletionProtection = pulumi.BoolPtr(true)
	}

	createdInstance, err := redis.NewInstance(ctx, "redis-instance", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create redis instance")
	}

	ctx.Export(OpHost, createdInstance.Host)
	ctx.Export(OpPort, createdInstance.Port)
	ctx.Export(OpReadEndpoint, createdInstance.ReadEndpoint)
	ctx.Export(OpReadEndpointPort, createdInstance.ReadEndpointPort)
	ctx.Export(OpCurrentLocationId, createdInstance.CurrentLocationId)
	ctx.Export(OpAuthString, createdInstance.AuthString)

	return nil
}
