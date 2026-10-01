package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// databaseUserVerifier verifies a DigitalOceanDatabaseUser via
// GET /v2/databases/{cluster_id}/users/{name}. The API has no standalone
// user id -- the (cluster, name) pair is the identity -- so the verifier
// reads both from the stack outputs (the OutputsVerifier extension exists
// for exactly this shape).
type databaseUserVerifier struct{}

func (*databaseUserVerifier) IDOutputKey() string { return "user_name" }

func (*databaseUserVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandatabaseuser requires the full outputs map (cluster_id + user_name); " +
		"the harness dispatches through VerifyExistsFromOutputs")
}

func (*databaseUserVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandatabaseuser requires the full outputs map (cluster_id + user_name); " +
		"the harness dispatches through VerifyAbsentFromOutputs")
}

func (v *databaseUserVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	user, err := v.getFromOutputs(ctx, client, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceandatabaseuser verify-exists failed")
	}
	if user == nil {
		return pkgerrors.Errorf("digitaloceandatabaseuser %q not found after deploy", StringOutput(outputs, "user_name"))
	}
	// Assert the role only when the stack outputs claim it (contractually
	// identical across both engines, so one assertion protects both).
	if role := StringOutput(outputs, "role"); role != "" {
		liveRole := user.Role
		if liveRole == "" {
			// The API omits the role for ordinary users; the provider
			// defaults it to "normal", and the outputs follow.
			liveRole = "normal"
		}
		if liveRole != role {
			return pkgerrors.Errorf("digitaloceandatabaseuser %q role mismatch: output %q, live %q",
				user.Name, role, liveRole)
		}
	}
	return nil
}

func (v *databaseUserVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	user, err := v.getFromOutputs(ctx, client, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceandatabaseuser verify-absent failed")
	}
	if user != nil {
		return &StillExistsError{Component: "digitaloceandatabaseuser", ID: user.Name}
	}
	return nil
}

// getFromOutputs returns the live user, or nil when it (or its whole
// cluster -- also a valid absence signal on composed teardowns) is gone.
func (*databaseUserVerifier) getFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) (*godo.DatabaseUser, error) {
	clusterID := StringOutput(outputs, "cluster_id")
	name := StringOutput(outputs, "user_name")
	if clusterID == "" || name == "" {
		return nil, pkgerrors.Errorf("outputs must carry cluster_id and user_name (got cluster_id=%q, user_name=%q)", clusterID, name)
	}
	user, _, err := client.Databases.GetUser(ctx, clusterID, name)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}
