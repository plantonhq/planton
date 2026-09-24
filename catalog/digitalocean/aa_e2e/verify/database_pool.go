package verify

import (
	"context"
	"strconv"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// databaseConnectionPoolVerifier verifies a
// DigitalOceanDatabaseConnectionPool via
// GET /v2/databases/{cluster_id}/pools/{name}. The API has no standalone
// pool id -- the (cluster, name) pair is the identity -- so the verifier
// reads both from the stack outputs.
//
// Connection URIs are never asserted: the provider assembles them from
// state credentials, so byte equality with the live API is not a contract
// (host/port are).
type databaseConnectionPoolVerifier struct{}

func (*databaseConnectionPoolVerifier) IDOutputKey() string { return "pool_name" }

func (*databaseConnectionPoolVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandatabaseconnectionpool requires the full outputs map (cluster_id + pool_name); " +
		"the harness dispatches through VerifyExistsFromOutputs")
}

func (*databaseConnectionPoolVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandatabaseconnectionpool requires the full outputs map (cluster_id + pool_name); " +
		"the harness dispatches through VerifyAbsentFromOutputs")
}

func (v *databaseConnectionPoolVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	pool, err := v.getFromOutputs(ctx, client, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceandatabaseconnectionpool verify-exists failed")
	}
	if pool == nil {
		return pkgerrors.Errorf("digitaloceandatabaseconnectionpool %q not found after deploy", StringOutput(outputs, "pool_name"))
	}

	// Assert connection posture only when the stack outputs claim it.
	if pool.Connection != nil {
		if host := StringOutput(outputs, "host"); host != "" && pool.Connection.Host != host {
			return pkgerrors.Errorf("digitaloceandatabaseconnectionpool %q host mismatch: output %q, live %q",
				pool.Name, host, pool.Connection.Host)
		}
		if port := StringOutput(outputs, "port"); port != "" && strconv.Itoa(pool.Connection.Port) != port {
			return pkgerrors.Errorf("digitaloceandatabaseconnectionpool %q port mismatch: output %s, live %d",
				pool.Name, port, pool.Connection.Port)
		}
	}

	return nil
}

func (v *databaseConnectionPoolVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	pool, err := v.getFromOutputs(ctx, client, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceandatabaseconnectionpool verify-absent failed")
	}
	if pool != nil {
		return &StillExistsError{Component: "digitaloceandatabaseconnectionpool", ID: pool.Name}
	}
	return nil
}

// getFromOutputs returns the live pool, or nil when it (or its whole
// cluster -- also a valid absence signal on composed teardowns) is gone.
func (*databaseConnectionPoolVerifier) getFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) (*godo.DatabasePool, error) {
	clusterID := StringOutput(outputs, "cluster_id")
	name := StringOutput(outputs, "pool_name")
	if clusterID == "" || name == "" {
		return nil, pkgerrors.Errorf("outputs must carry cluster_id and pool_name (got cluster_id=%q, pool_name=%q)", clusterID, name)
	}
	pool, _, err := client.Databases.GetPool(ctx, clusterID, name)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return pool, nil
}
