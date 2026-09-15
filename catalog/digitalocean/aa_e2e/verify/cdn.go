package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// cdnVerifier verifies a DigitalOceanCdn via GET /v2/cdn/endpoints/{id}.
// This verifier is the kind's trustworthy destroy signal: the provider's own
// read-after-destroy ERRORS instead of settling (it retries 404s for 30
// seconds expecting create-side eventual consistency, then still returns the
// error after clearing state), so an IaC refresh cannot prove absence.
type cdnVerifier struct{}

func (*cdnVerifier) IDOutputKey() string { return "cdn_id" }

func (*cdnVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	exists, err := cdnExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceancdn verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceancdn %q not found after deploy", id)
	}
	return nil
}

func (*cdnVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	exists, err := cdnExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceancdn verify-absent failed for %q", id)
	}
	if exists {
		return &StillExistsError{Component: "digitaloceancdn", ID: id}
	}
	return nil
}

func cdnExists(ctx context.Context, client *godo.Client, id string) (bool, error) {
	_, _, err := client.CDNs.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
