package module

import (
	"github.com/pulumi/pulumi-mongodbatlas/sdk/v3/go/mongodbatlas"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output key constants
const (
	OpId               = "id"
	OpBootstrapUrl     = "bootstrap_endpoint"
	OpCrn              = "crn"
	OpRestEndpoint     = "rest_endpoint"
	OpClusterName      = "cluster_name"
	OpClusterType      = "cluster_type"
	OpStateName        = "state_name"
	OpMongoDBVersion   = "mongo_db_version"
	OpProjectId        = "project_id"
	OpConnectionString = "connection_string"
)

// exportOutputs exports cluster information to Pulumi stack outputs
func exportOutputs(ctx *pulumi.Context, cluster *atlasmongodb.AdvancedCluster, locals *Locals) error {
	// Export required outputs matching stack_outputs.proto
	ctx.Export(OpId, cluster.ID())

	// Export connection strings with safe access
	// The connection string is embedded in the ConnectionStrings array
	ctx.Export(OpBootstrapUrl, cluster.ConnectionStrings.ApplyT(func(connections []atlasmongodb.AdvancedClusterConnectionString) string {
		if len(connections) > 0 && connections[0].StandardSrv != nil {
			return *connections[0].StandardSrv
		}
		return ""
	}).(pulumi.StringOutput))

	// CRN is the cluster ID for Atlas MongoDB
	ctx.Export(OpCrn, cluster.ClusterId)

	ctx.Export(OpRestEndpoint, cluster.ConnectionStrings.ApplyT(func(connections []atlasmongodb.AdvancedClusterConnectionString) string {
		if len(connections) > 0 && connections[0].Standard != nil {
			return *connections[0].Standard
		}
		return ""
	}).(pulumi.StringOutput))

	// Export additional useful outputs
	ctx.Export(OpClusterName, cluster.Name)
	ctx.Export(OpClusterType, cluster.ClusterType)
	ctx.Export(OpStateName, cluster.StateName)
	ctx.Export(OpMongoDBVersion, cluster.MongoDbVersion)
	ctx.Export(OpProjectId, pulumi.String(locals.ProjectId))

	// Export full connection string object as structured data
	ctx.Export(OpConnectionString, cluster.ConnectionStrings)

	return nil
}
