package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// monitorAlertVerifier verifies a DigitalOceanMonitorAlert via
// GET /v2/monitoring/alerts/{alert_id}. The id is the policy UUID carried
// by the resource id (the provider's own uuid attribute is declared but
// never populated at the pin, which is why the output derives from the id).
type monitorAlertVerifier struct{}

func (*monitorAlertVerifier) IDOutputKey() string { return "alert_id" }

func (*monitorAlertVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	policy, _, err := client.Monitoring.GetAlertPolicy(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return pkgerrors.Errorf("digitaloceanmonitoralert %q not found after deploy", id)
		}
		return pkgerrors.Wrap(err, "digitaloceanmonitoralert verify-exists failed")
	}
	if policy.UUID == "" {
		return pkgerrors.Errorf("digitaloceanmonitoralert %q returned an empty policy", id)
	}
	return nil
}

func (*monitorAlertVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	_, _, err := client.Monitoring.GetAlertPolicy(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return pkgerrors.Wrap(err, "digitaloceanmonitoralert verify-absent failed")
	}
	return &StillExistsError{Component: "digitaloceanmonitoralert", ID: id}
}
