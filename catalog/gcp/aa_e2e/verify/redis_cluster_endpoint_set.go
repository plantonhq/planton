package verify

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// redisClusterEndpointSetVerifier probes the user-created connections
// registered on a Memorystore for Redis Cluster by reading the cluster
// itself: the registration is metadata on the cluster (its
// clusterEndpoints), so existence means the cluster lists at least the
// declared number of connections, each ACTIVE, and absence means the
// cluster lists none.
type redisClusterEndpointSetVerifier struct{}

// IDOutputKey is the bare cluster name the registration is keyed by.
func (v *redisClusterEndpointSetVerifier) IDOutputKey() string { return "cluster_name" }

// clusterPath rebuilds the cluster's full resource path from the harness
// project and the set's cluster_name and region outputs.
func (v *redisClusterEndpointSetVerifier) clusterPath(svc *Services, outputs map[string]string) (string, error) {
	name, region := outputs["cluster_name"], outputs["region"]
	if name == "" || region == "" {
		return "", errors.New("cluster_name or region output missing")
	}
	return fmt.Sprintf("projects/%s/locations/%s/clusters/%s", svc.Project, region, name), nil
}

func (v *redisClusterEndpointSetVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	path, err := v.clusterPath(svc, outputs)
	if err != nil {
		return errors.Wrap(err, "after deploy")
	}

	cluster, err := svc.Redis.Projects.Locations.Clusters.Get(path).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "redis cluster %s not found after registering connections", path)
	}

	wantConnections, _ := strconv.Atoi(outputs["connection_count"])
	live := 0
	for _, ep := range cluster.ClusterEndpoints {
		for _, conn := range ep.Connections {
			if conn.PscConnection == nil {
				continue
			}
			live++
			if conn.PscConnection.PscConnectionStatus != "" && conn.PscConnection.PscConnectionStatus != "ACTIVE" {
				return errors.Errorf("redis cluster %s: connection %s status is %q, want ACTIVE", path, conn.PscConnection.PscConnectionId, conn.PscConnection.PscConnectionStatus)
			}
		}
	}
	if wantConnections > 0 && live < wantConnections {
		return errors.Errorf("redis cluster %s lists %d user-created connections, want at least %d", path, live, wantConnections)
	}
	if wantEndpoints, _ := strconv.Atoi(outputs["endpoint_count"]); wantEndpoints > 0 && len(cluster.ClusterEndpoints) < wantEndpoints {
		return errors.Errorf("redis cluster %s lists %d endpoints, want at least %d", path, len(cluster.ClusterEndpoints), wantEndpoints)
	}
	return nil
}

func (v *redisClusterEndpointSetVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	path, err := v.clusterPath(svc, outputs)
	if err != nil {
		return nil
	}

	cluster, err := svc.Redis.Projects.Locations.Clusters.Get(path).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			// The cluster itself is gone (the chain tore it down after the
			// set); no registration can remain.
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing redis cluster %s after deregistering connections", path)
	}
	for _, ep := range cluster.ClusterEndpoints {
		if len(ep.Connections) > 0 {
			return errors.Errorf("redis cluster %s still lists user-created connections after destroy", path)
		}
	}
	return nil
}
