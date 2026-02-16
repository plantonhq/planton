package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/memorydb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// cluster creates the MemoryDB cluster and exports outputs.
func cluster(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	createdSubnetGroup *memorydb.SubnetGroup,
	createdParamGroup *memorydb.ParameterGroup,
) error {
	spec := locals.Spec

	args := &memorydb.ClusterArgs{
		AclName:  pulumi.String(spec.GetAclName()),
		NodeType: pulumi.String(spec.NodeType),
		Tags:     pulumi.ToStringMap(locals.AwsTags),
	}

	// Engine
	if spec.Engine != "" {
		args.Engine = pulumi.String(spec.Engine)
	}

	// Engine version
	if spec.EngineVersion != "" {
		args.EngineVersion = pulumi.String(spec.EngineVersion)
	}

	// Description
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}

	// Port
	if spec.Port != nil {
		args.Port = pulumi.Int(*spec.Port)
	}

	// Topology
	if spec.NumShards != nil {
		args.NumShards = pulumi.Int(*spec.NumShards)
	}
	if spec.NumReplicasPerShard != nil {
		args.NumReplicasPerShard = pulumi.Int(*spec.NumReplicasPerShard)
	}

	// Networking
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

	// Encryption
	if spec.TlsEnabled != nil {
		args.TlsEnabled = pulumi.Bool(*spec.TlsEnabled)
	}
	if spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyArn = pulumi.String(spec.KmsKeyId.GetValue())
	}

	// Maintenance and snapshots
	if spec.MaintenanceWindow != "" {
		args.MaintenanceWindow = pulumi.String(spec.MaintenanceWindow)
	}
	if spec.SnapshotRetentionLimit > 0 {
		args.SnapshotRetentionLimit = pulumi.Int(spec.SnapshotRetentionLimit)
	}
	if spec.SnapshotWindow != "" {
		args.SnapshotWindow = pulumi.String(spec.SnapshotWindow)
	}
	if spec.FinalSnapshotName != "" {
		args.FinalSnapshotName = pulumi.String(spec.FinalSnapshotName)
	}

	// Restore from snapshot
	if len(spec.SnapshotArns) > 0 {
		args.SnapshotArns = pulumi.ToStringArray(spec.SnapshotArns)
	}
	if spec.SnapshotName != "" {
		args.SnapshotName = pulumi.String(spec.SnapshotName)
	}

	// Parameter group
	if createdParamGroup != nil {
		args.ParameterGroupName = createdParamGroup.Name
	}

	// Notifications
	if spec.SnsTopicArn.GetValue() != "" {
		args.SnsTopicArn = pulumi.String(spec.SnsTopicArn.GetValue())
	}

	// Advanced
	if spec.AutoMinorVersionUpgrade != nil {
		args.AutoMinorVersionUpgrade = pulumi.Bool(*spec.AutoMinorVersionUpgrade)
	}
	if spec.DataTiering {
		args.DataTiering = pulumi.Bool(true)
	}

	// Create cluster
	c, err := memorydb.NewCluster(ctx, "memorydb-cluster", args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create memorydb cluster")
	}

	// Export outputs
	ctx.Export(OpClusterArn, c.Arn)
	ctx.Export(OpClusterName, c.Name)
	ctx.Export(OpEnginePatchVersion, c.EnginePatchVersion)

	// Cluster endpoint is a nested object with address and port
	ctx.Export(OpClusterEndpointAddress, c.ClusterEndpoints.Index(pulumi.Int(0)).Address())
	ctx.Export(OpClusterEndpointPort, c.ClusterEndpoints.Index(pulumi.Int(0)).Port())

	return nil
}
