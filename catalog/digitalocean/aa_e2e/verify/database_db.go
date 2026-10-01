package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// databaseDbVerifier verifies a DigitalOceanDatabaseDb via
// GET /v2/databases/{cluster_id}/dbs/{name}. The API has no standalone
// database id -- the (cluster, name) pair is the identity -- so the
// verifier reads both from the stack outputs.
type databaseDbVerifier struct{}

func (*databaseDbVerifier) IDOutputKey() string { return "database_name" }

func (*databaseDbVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandatabasedb requires the full outputs map (cluster_id + database_name); " +
		"the harness dispatches through VerifyExistsFromOutputs")
}

func (*databaseDbVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandatabasedb requires the full outputs map (cluster_id + database_name); " +
		"the harness dispatches through VerifyAbsentFromOutputs")
}

func (v *databaseDbVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	exists, err := v.existsFromOutputs(ctx, client, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceandatabasedb verify-exists failed")
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceandatabasedb %q not found after deploy", StringOutput(outputs, "database_name"))
	}
	return nil
}

func (v *databaseDbVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	exists, err := v.existsFromOutputs(ctx, client, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceandatabasedb verify-absent failed")
	}
	if exists {
		return &StillExistsError{Component: "digitaloceandatabasedb", ID: StringOutput(outputs, "database_name")}
	}
	return nil
}

// existsFromOutputs reports the logical database's live existence. The
// cluster vanishing is also a valid absence signal on composed teardowns.
func (*databaseDbVerifier) existsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) (bool, error) {
	clusterID := StringOutput(outputs, "cluster_id")
	name := StringOutput(outputs, "database_name")
	if clusterID == "" || name == "" {
		return false, pkgerrors.Errorf("outputs must carry cluster_id and database_name (got cluster_id=%q, database_name=%q)", clusterID, name)
	}
	_, _, err := client.Databases.GetDB(ctx, clusterID, name)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
