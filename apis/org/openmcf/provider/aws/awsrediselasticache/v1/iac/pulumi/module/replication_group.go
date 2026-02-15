package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/elasticache"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// replicationGroup creates the ElastiCache replication group and exports outputs.
func replicationGroup(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	createdSubnetGroup *elasticache.SubnetGroup,
	createdParamGroup *elasticache.ParameterGroup,
) error {
	spec := locals.Spec

	args := &elasticache.ReplicationGroupArgs{
		ReplicationGroupId: pulumi.String(locals.Target.Metadata.Id),
		Description:        pulumi.String(spec.Description),
		Engine:             pulumi.String(spec.Engine),
		NodeType:           pulumi.String(spec.NodeType),
		Tags:               pulumi.ToStringMap(locals.AwsTags),
		ApplyImmediately:   pulumi.Bool(spec.ApplyImmediately),
	}

	// Engine version
	if spec.EngineVersion != "" {
		args.EngineVersion = pulumi.String(spec.EngineVersion)
	}

	// Port
	if spec.Port != nil {
		args.Port = pulumi.Int(*spec.Port)
	}

	// -------------------------------------------------------------------
	// Topology
	// -------------------------------------------------------------------

	if spec.NumCacheClusters > 0 {
		// Non-clustered mode
		args.NumCacheClusters = pulumi.Int(spec.NumCacheClusters)
	} else if spec.NumNodeGroups > 0 {
		// Clustered mode
		args.NumNodeGroups = pulumi.Int(spec.NumNodeGroups)
		if spec.ReplicasPerNodeGroup > 0 {
			args.ReplicasPerNodeGroup = pulumi.Int(spec.ReplicasPerNodeGroup)
		}
	}

	// -------------------------------------------------------------------
	// High availability
	// -------------------------------------------------------------------

	args.AutomaticFailoverEnabled = pulumi.Bool(spec.AutomaticFailoverEnabled)
	args.MultiAzEnabled = pulumi.Bool(spec.MultiAzEnabled)

	// -------------------------------------------------------------------
	// Networking
	// -------------------------------------------------------------------

	if createdSubnetGroup != nil {
		args.SubnetGroupName = createdSubnetGroup.Name
	}

	var sgIds pulumi.StringArray
	for _, sg := range spec.SecurityGroupIds {
		if sg.GetValue() != "" {
			sgIds = append(sgIds, pulumi.String(sg.GetValue()))
		}
	}
	if len(sgIds) > 0 {
		args.SecurityGroupIds = sgIds
	}

	// -------------------------------------------------------------------
	// Encryption
	// -------------------------------------------------------------------

	args.AtRestEncryptionEnabled = pulumi.Bool(spec.AtRestEncryptionEnabled)
	args.TransitEncryptionEnabled = pulumi.Bool(spec.TransitEncryptionEnabled)

	if spec.TransitEncryptionMode != "" {
		args.TransitEncryptionMode = pulumi.String(spec.TransitEncryptionMode)
	}

	if spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.String(spec.KmsKeyId.GetValue())
	}

	// -------------------------------------------------------------------
	// Authentication
	// -------------------------------------------------------------------

	if spec.AuthToken.GetValue() != "" {
		args.AuthToken = pulumi.String(spec.AuthToken.GetValue())
	}

	if len(spec.UserGroupIds) > 0 {
		args.UserGroupIds = pulumi.ToStringArray(spec.UserGroupIds)
	}

	// -------------------------------------------------------------------
	// Maintenance and snapshots
	// -------------------------------------------------------------------

	if spec.MaintenanceWindow != "" {
		args.MaintenanceWindow = pulumi.String(spec.MaintenanceWindow)
	}
	if spec.SnapshotRetentionLimit > 0 {
		args.SnapshotRetentionLimit = pulumi.Int(spec.SnapshotRetentionLimit)
	}
	if spec.SnapshotWindow != "" {
		args.SnapshotWindow = pulumi.String(spec.SnapshotWindow)
	}
	if spec.FinalSnapshotIdentifier != "" {
		args.FinalSnapshotIdentifier = pulumi.String(spec.FinalSnapshotIdentifier)
	}

	// -------------------------------------------------------------------
	// Parameter group
	// -------------------------------------------------------------------

	if createdParamGroup != nil {
		args.ParameterGroupName = createdParamGroup.Name
	}

	// -------------------------------------------------------------------
	// Logging
	// -------------------------------------------------------------------

	if len(spec.LogDeliveryConfigurations) > 0 {
		var logConfigs elasticache.ReplicationGroupLogDeliveryConfigurationArray
		for _, lc := range spec.LogDeliveryConfigurations {
			logConfigs = append(logConfigs, &elasticache.ReplicationGroupLogDeliveryConfigurationArgs{
				DestinationType: pulumi.String(lc.DestinationType),
				Destination:     pulumi.String(lc.Destination.GetValue()),
				LogFormat:       pulumi.String(lc.LogFormat),
				LogType:         pulumi.String(lc.LogType),
			})
		}
		args.LogDeliveryConfigurations = logConfigs
	}

	// -------------------------------------------------------------------
	// Advanced
	// -------------------------------------------------------------------

	if spec.NotificationTopicArn.GetValue() != "" {
		args.NotificationTopicArn = pulumi.String(spec.NotificationTopicArn.GetValue())
	}
	if spec.AutoMinorVersionUpgrade {
		args.AutoMinorVersionUpgrade = pulumi.Bool(true)
	}
	if spec.DataTieringEnabled {
		args.DataTieringEnabled = pulumi.Bool(true)
	}

	// -------------------------------------------------------------------
	// Create replication group
	// -------------------------------------------------------------------

	rg, err := elasticache.NewReplicationGroup(ctx, "replication-group", args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create replication group")
	}

	// -------------------------------------------------------------------
	// Export outputs
	// -------------------------------------------------------------------

	ctx.Export(OpReplicationGroupId, rg.ID())
	ctx.Export(OpPrimaryEndpointAddress, rg.PrimaryEndpointAddress)
	ctx.Export(OpReaderEndpointAddress, rg.ReaderEndpointAddress)
	ctx.Export(OpConfigurationEndpointAddress, rg.ConfigurationEndpointAddress)
	ctx.Export(OpArn, rg.Arn)
	ctx.Export(OpPort, rg.Port)

	return nil
}
