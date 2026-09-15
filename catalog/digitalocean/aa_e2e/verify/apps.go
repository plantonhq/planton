package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// appVerifier verifies components backed by an App Platform app via
// GET /v2/apps/{id}. Two kinds share it: DigitalOceanApp (the app is the
// component) and DigitalOceanFunction (the provider has no standalone
// Functions resource, so the component deploys an app carrying a functions
// section and its function_id output IS the app id). The struct is
// parameterized the same way the AWS harness reuses one verifier across
// load-balancer kinds.
type appVerifier struct {
	component   string
	idOutputKey string
}

func (v *appVerifier) IDOutputKey() string { return v.idOutputKey }

func (v *appVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	exists, err := appExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "%s verify-exists failed for %q", v.component, id)
	}
	if !exists {
		return pkgerrors.Errorf("%s %q not found after deploy", v.component, id)
	}
	return nil
}

func (v *appVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	exists, err := appExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "%s verify-absent failed for %q", v.component, id)
	}
	if exists {
		return &StillExistsError{Component: v.component, ID: id}
	}
	return nil
}

func appExists(ctx context.Context, client *godo.Client, id string) (bool, error) {
	_, _, err := client.Apps.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
