package verify

import (
	"context"
	"strconv"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// databaseReplicaVerifier verifies a DigitalOceanDatabaseReplica via
// GET /v2/databases/{cluster_id}/replicas/{name}. DigitalOcean reads
// replicas by (cluster, name) -- the replica's own UUID exists but has no
// read endpoint of its own -- so the verifier reads both from the stack
// outputs.
type databaseReplicaVerifier struct{}

func (*databaseReplicaVerifier) IDOutputKey() string { return "replica_id" }

func (*databaseReplicaVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandatabasereplica requires the full outputs map (cluster_id + replica_name); " +
		"the harness dispatches through VerifyExistsFromOutputs")
}

func (*databaseReplicaVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandatabasereplica requires the full outputs map (cluster_id + replica_name); " +
		"the harness dispatches through VerifyAbsentFromOutputs")
}

func (v *databaseReplicaVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	replica, err := v.getFromOutputs(ctx, client, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceandatabasereplica verify-exists failed")
	}
	if replica == nil {
		return pkgerrors.Errorf("digitaloceandatabasereplica %q not found after deploy", StringOutput(outputs, "replica_name"))
	}

	if replica.Status != "online" {
		return pkgerrors.Errorf("digitaloceandatabasereplica %q status is %q, want online", replica.Name, replica.Status)
	}

	// The exported replica_id must be the API's UUID, not the legacy
	// composite state id -- assert it against the live object.
	if id := StringOutput(outputs, "replica_id"); id != "" && replica.ID != id {
		return pkgerrors.Errorf("digitaloceandatabasereplica %q replica_id mismatch: output %q, live %q",
			replica.Name, id, replica.ID)
	}

	// Assert connection posture only when the stack outputs claim it.
	if replica.Connection != nil {
		if host := StringOutput(outputs, "host"); host != "" && replica.Connection.Host != host {
			return pkgerrors.Errorf("digitaloceandatabasereplica %q host mismatch: output %q, live %q",
				replica.Name, host, replica.Connection.Host)
		}
		if port := StringOutput(outputs, "port"); port != "" && strconv.Itoa(replica.Connection.Port) != port {
			return pkgerrors.Errorf("digitaloceandatabasereplica %q port mismatch: output %s, live %d",
				replica.Name, port, replica.Connection.Port)
		}
	}

	return nil
}

func (v *databaseReplicaVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	replica, err := v.getFromOutputs(ctx, client, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceandatabasereplica verify-absent failed")
	}
	if replica != nil {
		return &StillExistsError{Component: "digitaloceandatabasereplica", ID: replica.Name}
	}
	return nil
}

// getFromOutputs returns the live replica, or nil when it (or its whole
// primary cluster -- also a valid absence signal on composed teardowns) is
// gone.
func (*databaseReplicaVerifier) getFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) (*godo.DatabaseReplica, error) {
	clusterID := StringOutput(outputs, "cluster_id")
	name := StringOutput(outputs, "replica_name")
	if clusterID == "" || name == "" {
		return nil, pkgerrors.Errorf("outputs must carry cluster_id and replica_name (got cluster_id=%q, replica_name=%q)", clusterID, name)
	}
	replica, _, err := client.Databases.GetReplica(ctx, clusterID, name)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return replica, nil
}
