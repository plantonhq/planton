package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// redisClusterVerifier probes a Memorystore for Redis Cluster through the
// typed Redis API client: existence, READY state, the uid the module
// exported, the Google-placed discovery endpoint when the manifest asked
// for one, and the per-connection-type service attachment handles the
// outputs expose for consumer forwarding rules.
type redisClusterVerifier struct{}

// IDOutputKey is the cluster's full resource path -- the composition key
// every reference to a cluster resolves to.
func (v *redisClusterVerifier) IDOutputKey() string { return "name" }

func (v *redisClusterVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}

	cluster, err := svc.Redis.Projects.Locations.Clusters.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "redis cluster %s not found after deploy", name)
	}
	if cluster.State != "READY" {
		return errors.Errorf("redis cluster %s state is %q, want READY", name, cluster.State)
	}
	if uid := outputs["uid"]; uid != "" && cluster.Uid != uid {
		return errors.Errorf("redis cluster %s uid mismatch: output %q, live %q", name, uid, cluster.Uid)
	}

	// The discovery address in the outputs must be one of the live
	// Google-placed endpoints -- the address clients will actually connect
	// to. Empty when the manifest carried no psc_configs.
	if addr := outputs["discovery_endpoint_address"]; addr != "" {
		found := false
		for _, ep := range cluster.DiscoveryEndpoints {
			if ep.Address == addr {
				found = true
			}
		}
		if !found {
			return errors.Errorf("redis cluster %s: discovery_endpoint_address output %q matches no live discovery endpoint", name, addr)
		}
	}

	// Each service attachment handle the module exported must be one Google
	// publishes under the matching connection type -- what a consumer's
	// forwarding rule and a GcpRedisClusterEndpointSet reference.
	for output, connectionType := range map[string]string{
		"discovery_service_attachment": "CONNECTION_TYPE_DISCOVERY",
		"primary_service_attachment":   "CONNECTION_TYPE_PRIMARY",
		"reader_service_attachment":    "CONNECTION_TYPE_READER",
	} {
		want := outputs[output]
		if want == "" {
			continue
		}
		found := false
		for _, a := range cluster.PscServiceAttachments {
			if a.ConnectionType == connectionType && a.ServiceAttachment == want {
				found = true
			}
		}
		if !found {
			return errors.Errorf("redis cluster %s: %s output %q matches no live %s attachment", name, output, want, connectionType)
		}
	}
	return nil
}

func (v *redisClusterVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}

	_, err := svc.Redis.Projects.Locations.Clusters.Get(name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing redis cluster %s after destroy", name)
	}
	return errors.Errorf("redis cluster %s still exists after destroy", name)
}
