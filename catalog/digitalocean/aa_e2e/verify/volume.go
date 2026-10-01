package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// volumeVerifier verifies a DigitalOceanVolume via GET /v2/volumes/{id}.
type volumeVerifier struct{}

func (*volumeVerifier) IDOutputKey() string { return "volume_id" }

func (*volumeVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	exists, err := volumeExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceanvolume verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceanvolume %q not found after deploy", id)
	}
	return nil
}

func (*volumeVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	exists, err := volumeExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceanvolume verify-absent failed for %q", id)
	}
	if exists {
		return &StillExistsError{Component: "digitaloceanvolume", ID: id}
	}
	return nil
}

func volumeExists(ctx context.Context, client *godo.Client, id string) (bool, error) {
	_, _, err := client.Storage.GetVolume(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
