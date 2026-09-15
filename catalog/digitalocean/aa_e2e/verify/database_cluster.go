package verify

import (
	"context"
	"strconv"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// databaseClusterVerifier verifies a DigitalOceanDatabaseCluster via
// GET /v2/databases/{id}. Beyond existence, it asserts the connection
// details the module CLAIMS in its stack outputs (host, port) against the
// live cluster -- outputs are contractually identical across both engines,
// so one assertion protects both, and an absent output simply means "not
// claimed" and is skipped.
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
