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
// Measured live: the API answers 404 within two seconds of the DELETE.
//
// The kind's headline promise is the edge hostname customers point CNAMEs
// at, so the deploy-side check asserts the `endpoint` output against the
// live endpoint whenever the output claims one.
type cdnVerifier struct{}

func (*cdnVerifier) IDOutputKey() string { return "cdn_id" }

func (v *cdnVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	_, err := cdnGet(ctx, client, id)
	return err
}

func (v *cdnVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	_, _, err := client.CDNs.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return pkgerrors.Wrapf(err, "digitaloceancdn verify-absent failed for %q", id)
	}
	return &StillExistsError{Component: "digitaloceancdn", ID: id}
}

func (v *cdnVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "cdn_id")
	if id == "" {
		return pkgerrors.New("cdn_id output missing after deploy")
	}
	cdn, err := cdnGet(ctx, client, id)
	if err != nil {
		return err
	}
	if endpoint := StringOutput(outputs, "endpoint"); endpoint != "" && cdn.Endpoint != endpoint {
		return pkgerrors.Errorf("digitaloceancdn %q endpoint mismatch: output %q, live %q", id, endpoint, cdn.Endpoint)
	}
	return nil
}

func (v *cdnVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "cdn_id")
	if id == "" {
		return pkgerrors.New("cdn_id output missing for destroy verification")
	}
	return v.VerifyAbsent(ctx, client, id)
}

func cdnGet(ctx context.Context, client *godo.Client, id string) (*godo.CDN, error) {
	cdn, _, err := client.CDNs.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, pkgerrors.Errorf("digitaloceancdn %q not found after deploy", id)
		}
		return nil, pkgerrors.Wrapf(err, "digitaloceancdn verify-exists failed for %q", id)
	}
	return cdn, nil
}
