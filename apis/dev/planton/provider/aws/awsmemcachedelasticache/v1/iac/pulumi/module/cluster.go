package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/elasticache"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// cluster creates the ElastiCache Memcached cluster and exports outputs.
func cluster(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	createdSubnetGroup *elasticache.SubnetGroup,
	createdParamGroup *elasticache.ParameterGroup,
) error {
	spec := locals.Spec

	args := &elasticache.ClusterArgs{
		ClusterId:        pulumi.String(locals.Target.Metadata.Id),
		Engine:           pulumi.String("memcached"),
		EngineVersion:    pulumi.String(spec.EngineVersion),
		NodeType:         pulumi.String(spec.NodeType),
		NumCacheNodes:    pulumi.Int(spec.NumCacheNodes),
		Tags:             pulumi.ToStringMap(locals.AwsTags),
		ApplyImmediately: pulumi.Bool(spec.ApplyImmediately),
	}

	// Port
	if spec.Port != nil {
		args.Port = pulumi.Int(*spec.Port)
	}

	// AZ mode
	if spec.AzMode != "" {
		args.AzMode = pulumi.String(spec.AzMode)
	}

	// Transit encryption
	if spec.TransitEncryptionEnabled {
		args.TransitEncryptionEnabled = pulumi.Bool(true)
	}

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
	// Parameter group
	// -------------------------------------------------------------------

	if createdParamGroup != nil {
		args.ParameterGroupName = createdParamGroup.Name
	}

	// -------------------------------------------------------------------
	// Maintenance
	// -------------------------------------------------------------------

	if spec.MaintenanceWindow != "" {
		args.MaintenanceWindow = pulumi.String(spec.MaintenanceWindow)
	}
	if spec.AutoMinorVersionUpgrade {
		args.AutoMinorVersionUpgrade = pulumi.String("true")
	}

	// -------------------------------------------------------------------
	// Notifications
	// -------------------------------------------------------------------

	if spec.NotificationTopicArn.GetValue() != "" {
		args.NotificationTopicArn = pulumi.String(spec.NotificationTopicArn.GetValue())
	}

	// -------------------------------------------------------------------
	// Node placement
	// -------------------------------------------------------------------

	if len(spec.PreferredAvailabilityZones) > 0 {
		args.PreferredAvailabilityZones = pulumi.ToStringArray(spec.PreferredAvailabilityZones)
	}

	// -------------------------------------------------------------------
	// Create cluster
	// -------------------------------------------------------------------

	c, err := elasticache.NewCluster(ctx, "cluster", args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create memcached cluster")
	}

	// -------------------------------------------------------------------
	// Export outputs
	// -------------------------------------------------------------------

	ctx.Export(OpClusterId, c.ClusterId)
	ctx.Export(OpClusterAddress, c.ClusterAddress)
	ctx.Export(OpConfigEndpoint, c.ConfigurationEndpoint)
	ctx.Export(OpArn, c.Arn)
	ctx.Export(OpPort, c.Port)

	return nil
}
