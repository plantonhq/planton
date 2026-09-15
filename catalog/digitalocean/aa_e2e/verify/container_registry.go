package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// containerRegistryVerifier verifies a DigitalOceanContainerRegistry via
// GET /v2/registries/{name}. Registries are identified by name, which the
// kind exports as registry_name.
type containerRegistryVerifier struct{}

func (*containerRegistryVerifier) IDOutputKey() string { return "registry_name" }

func (*containerRegistryVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	exists, err := containerRegistryExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceancontainerregistry verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceancontainerregistry %q not found after deploy", id)
	}
	return nil
}

func (*containerRegistryVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	exists, err := containerRegistryExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceancontainerregistry verify-absent failed for %q", id)
	}
	if exists {
		return &StillExistsError{Component: "digitaloceancontainerregistry", ID: id}
	}
	return nil
}

func containerRegistryExists(ctx context.Context, client *godo.Client, name string) (bool, error) {
	_, _, err := client.Registries.Get(ctx, name)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
