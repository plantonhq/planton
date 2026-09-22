package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/redis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// endpointSet registers the consumer-built Private Service Connect
// connections on a Memorystore for Redis Cluster -- Google's user-created
// connections. Google replaces the cluster's whole user-created endpoint
// list with this set on every apply, so the manifest's list IS the set.
//
// Every field of a connection is an output of the block that built it (the
// forwarding rule twice, on two output paths; the reserved address; the
// network; the cluster's service-attachment handle), so the registration
// carries no literals a consumer has to copy by hand. Only the project is
// immutable; the set updates in place.
func endpointSet(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpRedisClusterEndpointSet.Spec

	endpoints := redis.ClusterUserCreatedConnectionsClusterEndpointArray{}
	connectionCount := 0
	for _, endpoint := range spec.Endpoints {
		connections := redis.ClusterUserCreatedConnectionsClusterEndpointConnectionArray{}
		for _, connection := range endpoint.Connections {
			pscConnection := &redis.ClusterUserCreatedConnectionsClusterEndpointConnectionPscConnectionArgs{
				ForwardingRule:    pulumi.String(connection.ForwardingRule.GetValue()),
				PscConnectionId:   pulumi.String(connection.PscConnectionId.GetValue()),
				Address:           pulumi.String(connection.Address.GetValue()),
				Network:           pulumi.String(connection.Network.GetValue()),
				ServiceAttachment: pulumi.String(connection.ServiceAttachment.GetValue()),
			}
			// Optional+Computed: Google records the rule's own project when
			// the spec leaves it unset, so it is sent only when set.
			if connection.ProjectId.GetValue() != "" {
				pscConnection.ProjectId = pulumi.StringPtr(connection.ProjectId.GetValue())
			}
			connections = append(connections, &redis.ClusterUserCreatedConnectionsClusterEndpointConnectionArgs{
				PscConnection: pscConnection,
			})
			connectionCount++
		}
		endpoints = append(endpoints, &redis.ClusterUserCreatedConnectionsClusterEndpointArgs{
			Connections: connections,
		})
	}

	args := &redis.ClusterUserCreatedConnectionsArgs{
		Name:             pulumi.String(locals.ClusterName),
		Region:           pulumi.String(spec.Region),
		ClusterEndpoints: endpoints,
	}
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}
	// Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
	// Sent only when set so the provider default stays in charge otherwise.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdSet, err := redis.NewClusterUserCreatedConnections(ctx, "redis-cluster-endpoint-set", args,
		pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to register redis cluster user-created connections")
	}

	ctx.Export(OpClusterName, createdSet.Name)
	ctx.Export(OpEndpointCount, pulumi.Int(len(spec.Endpoints)))
	ctx.Export(OpConnectionCount, pulumi.Int(connectionCount))
	ctx.Export(OpRegion, createdSet.Region)

	return nil
}
