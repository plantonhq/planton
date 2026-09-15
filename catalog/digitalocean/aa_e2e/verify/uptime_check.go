package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// uptimeCheckVerifier verifies a DigitalOceanUptimeCheck via
// GET /v2/uptime/checks/{check_id}. The composed alert rows live under the
// check and are destroyed with it, so the check's absence is the complete
// destroy signal.
type uptimeCheckVerifier struct{}

func (*uptimeCheckVerifier) IDOutputKey() string { return "check_id" }

func (*uptimeCheckVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	check, _, err := client.UptimeChecks.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return pkgerrors.Errorf("digitaloceanuptimecheck %q not found after deploy", id)
		}
		return pkgerrors.Wrap(err, "digitaloceanuptimecheck verify-exists failed")
	}
	if check.ID == "" {
		return pkgerrors.Errorf("digitaloceanuptimecheck %q returned an empty check", id)
	}
	return nil
}

func (*uptimeCheckVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	_, _, err := client.UptimeChecks.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return pkgerrors.Wrap(err, "digitaloceanuptimecheck verify-absent failed")
	}
	return &StillExistsError{Component: "digitaloceanuptimecheck", ID: id}
}
