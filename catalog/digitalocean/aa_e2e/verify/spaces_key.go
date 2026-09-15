package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// spacesKeyVerifier verifies a DigitalOceanSpacesKey via
// GET /v2/spaces/keys/{access_key}. Keys are control-plane objects reached
// with the account token (no Spaces credentials involved), and the access
// key ID is the resource identity. The secret key is write-once and never
// part of verification -- it exists only in the create response.
type spacesKeyVerifier struct{}

func (*spacesKeyVerifier) IDOutputKey() string { return "access_key" }

func (*spacesKeyVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	exists, err := spacesKeyExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceanspaceskey verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceanspaceskey %q not found after deploy", id)
	}
	return nil
}

func (*spacesKeyVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	exists, err := spacesKeyExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceanspaceskey verify-absent failed for %q", id)
	}
	if exists {
		return &StillExistsError{Component: "digitaloceanspaceskey", ID: id}
	}
	return nil
}

func spacesKeyExists(ctx context.Context, client *godo.Client, id string) (bool, error) {
	_, _, err := client.SpacesKeys.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
