package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// projectVerifier verifies a DigitalOceanProject via
// GET /v2/projects/{project_id}. Destroy relocates member resources to the
// account's default project before deleting, so absence of the project is
// the complete destroy signal (members are expected to survive elsewhere).
type projectVerifier struct{}

func (*projectVerifier) IDOutputKey() string { return "project_id" }

func (*projectVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	project, _, err := client.Projects.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return pkgerrors.Errorf("digitaloceanproject %q not found after deploy", id)
		}
		return pkgerrors.Wrap(err, "digitaloceanproject verify-exists failed")
	}
	if project.ID == "" {
		return pkgerrors.Errorf("digitaloceanproject %q returned an empty project", id)
	}
	return nil
}

func (*projectVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	_, _, err := client.Projects.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return pkgerrors.Wrap(err, "digitaloceanproject verify-absent failed")
	}
	return &StillExistsError{Component: "digitaloceanproject", ID: id}
}
