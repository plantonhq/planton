package verify

import (
	"context"
	"strconv"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// databaseClusterVerifier verifies a DigitalOceanDatabaseCluster via
// GET /v2/databases/{id}. Beyond existence, it asserts the connection
// details the module CLAIMS in its stack outputs (host, port,
// database_name, and the VPC-only private_host) against the live cluster
// -- outputs are contractually identical across both engines, so one
// assertion protects both, and an absent output simply means "not claimed"
// and is skipped.
type databaseClusterVerifier struct{}

func (*databaseClusterVerifier) IDOutputKey() string { return "cluster_id" }

func (*databaseClusterVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	_, err := getDatabaseCluster(ctx, client, id)
	if err != nil {
		return err
	}
	return nil
}

func (*databaseClusterVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	_, _, err := client.Databases.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return pkgerrors.Wrapf(err, "digitaloceandatabasecluster verify-absent failed for %q", id)
	}
	return &StillExistsError{Component: "digitaloceandatabasecluster", ID: id}
}

func (v *databaseClusterVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "cluster_id")
	if id == "" {
		return pkgerrors.New("cluster_id output missing after deploy")
	}

	database, err := getDatabaseCluster(ctx, client, id)
	if err != nil {
		return err
	}

	if database.Status != "online" {
		return pkgerrors.Errorf("digitaloceandatabasecluster %q status is %q, want online", id, database.Status)
	}

	// Assert connection posture only when the stack outputs claim it.
	if database.Connection != nil {
		if host := StringOutput(outputs, "host"); host != "" && database.Connection.Host != host {
			return pkgerrors.Errorf("digitaloceandatabasecluster %q host mismatch: output %q, live %q",
				id, host, database.Connection.Host)
		}
		if port := StringOutput(outputs, "port"); port != "" && strconv.Itoa(database.Connection.Port) != port {
			return pkgerrors.Errorf("digitaloceandatabasecluster %q port mismatch: output %s, live %d",
				id, port, database.Connection.Port)
		}
		if name := StringOutput(outputs, "database_name"); name != "" && database.Connection.Database != name {
			return pkgerrors.Errorf("digitaloceandatabasecluster %q database_name mismatch: output %q, live %q",
				id, name, database.Connection.Database)
		}
	}

	// The private endpoint exists only for a VPC-attached cluster, so a
	// claimed private_host is the VPC arm's own proof: it must match the
	// private connection DigitalOcean reports, never merely be non-empty.
	if privateHost := StringOutput(outputs, "private_host"); privateHost != "" {
		if database.PrivateConnection == nil {
			return pkgerrors.Errorf("digitaloceandatabasecluster %q claims private_host %q but the live cluster reports no private connection",
				id, privateHost)
		}
		if database.PrivateConnection.Host != privateHost {
			return pkgerrors.Errorf("digitaloceandatabasecluster %q private_host mismatch: output %q, live %q",
				id, privateHost, database.PrivateConnection.Host)
		}
	}

	return nil
}

func (v *databaseClusterVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "cluster_id")
	if id == "" {
		return pkgerrors.New("cluster_id output missing for destroy verification")
	}
	return v.VerifyAbsent(ctx, client, id)
}

func getDatabaseCluster(ctx context.Context, client *godo.Client, id string) (*godo.Database, error) {
	database, _, err := client.Databases.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, pkgerrors.Errorf("digitaloceandatabasecluster %q not found after deploy", id)
		}
		return nil, pkgerrors.Wrapf(err, "digitaloceandatabasecluster verify-exists failed for %q", id)
	}
	return database, nil
}
