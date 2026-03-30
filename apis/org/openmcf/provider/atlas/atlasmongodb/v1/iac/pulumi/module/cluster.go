package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-mongodbatlas/sdk/v3/go/mongodbatlas"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createCluster creates a Atlas MongoDB advanced cluster with all configured parameters
func createCluster(ctx *pulumi.Context, locals *Locals, provider *mongodbatlas.Provider) (*mongodbatlas.AdvancedCluster, error) {
	// Determine region based on provider
	// In a production implementation, this would come from the spec
	// For now, we'll use default regions per provider
	regionName := getDefaultRegion(locals.ProviderName)

	// Build region configuration
	regionConfig := &mongodbatlas.AdvancedClusterReplicationSpecRegionConfigArgs{
		ProviderName: pulumi.String(locals.ProviderName),
		RegionName:   pulumi.String(regionName),
		Priority:     pulumi.Int(int(locals.Priority)),

		// Electable nodes configuration
		ElectableSpecs: &mongodbatlas.AdvancedClusterReplicationSpecRegionConfigElectableSpecsArgs{
			InstanceSize: pulumi.String(locals.InstanceSize),
			NodeCount:    pulumi.Int(int(locals.ElectableNodes)),
		},

		// Auto-scaling configuration
		AutoScaling: &mongodbatlas.AdvancedClusterReplicationSpecRegionConfigAutoScalingArgs{
			DiskGbEnabled:           pulumi.Bool(locals.AutoScalingEnabled),
			ComputeEnabled:          pulumi.Bool(false),
			ComputeScaleDownEnabled: pulumi.Bool(false),
		},
	}

	// Add read-only specs if configured
	if locals.ReadOnlyNodes > 0 {
		regionConfig.ReadOnlySpecs = &mongodbatlas.AdvancedClusterReplicationSpecRegionConfigReadOnlySpecsArgs{
			InstanceSize: pulumi.String(locals.InstanceSize),
			NodeCount:    pulumi.Int(int(locals.ReadOnlyNodes)),
		}
	}

	// Build replication specs based on cluster configuration
	replicationSpecs := mongodbatlas.AdvancedClusterReplicationSpecArray{
		&mongodbatlas.AdvancedClusterReplicationSpecArgs{
			// Number of shards - 1 for REPLICASET, more for SHARDED/GEOSHARDED
			NumShards: pulumi.Int(getNumShards(locals.ClusterType)),

			// Region configuration
			RegionConfigs: mongodbatlas.AdvancedClusterReplicationSpecRegionConfigArray{regionConfig},
		},
	}

	// Build cluster arguments
	clusterArgs := &mongodbatlas.AdvancedClusterArgs{
		ProjectId:           pulumi.String(locals.ProjectId),
		Name:                pulumi.String(locals.ClusterName),
		ClusterType:         pulumi.String(locals.ClusterType),
		MongoDbMajorVersion: pulumi.String(locals.MongoDBVersion),
		BackupEnabled:       pulumi.Bool(locals.CloudBackup),
		ReplicationSpecs:    replicationSpecs,
	}

	// Create the cluster resource
	cluster, err := mongodbatlas.NewAdvancedCluster(ctx, locals.ClusterName, clusterArgs, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Atlas MongoDB cluster %s", locals.ClusterName)
	}

	return cluster, nil
}

// getDefaultRegion returns a default region for the given cloud provider
func getDefaultRegion(providerName string) string {
	switch providerName {
	case "AWS":
		return "US_EAST_1"
	case "GCP":
		return "CENTRAL_US"
	case "AZURE":
		return "US_EAST_2"
	default:
		return "US_EAST_1"
	}
}

// getNumShards returns the number of shards based on cluster type
func getNumShards(clusterType string) int {
	switch clusterType {
	case "SHARDED", "GEOSHARDED":
		return 2
	default:
		return 1
	}
}
