package verify

import (
	"context"
	"errors"
	"net/http"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// uptimeCheckVerifier verifies a DigitalOceanUptimeCheck via
// GET /v2/uptime/checks/{check_id} plus one GET per composed alert row from
// the alert_ids output (GET /v2/uptime/checks/{check_id}/alerts/{alert_id}).
// The rows are the scenario's whole point when present, so their existence
// is asserted against the API rather than inferred from a clean apply. On
// destroy the check's absence is DigitalOcean's contract for the rows too
// (they cannot outlive their check), and each row is probed anyway so a
// row that somehow survived would fail the lane instead of billing quietly.
//
// The Uptime API is the one DigitalOcean API whose "gone" answer is NOT a
// 404: a check the account does not own -- deleted a moment ago, deleted an
// hour ago, or never created -- answers 403 "you are not authorized to
// access this resource" on both the check and the alert endpoints (measured
// live; the monitoring API's alert policies 404 normally). isUptimeGone
// reads that signal for this verifier alone. It is safe to treat as
// absence here because the lane created the very id it probes: a genuinely
// unauthorized token would have failed DEPLOY, never reached this probe.
type uptimeCheckVerifier struct{}

// isUptimeGone reports whether an Uptime API error means the check or alert
// does not exist for this account: DigitalOcean's 403 forbidden, or the
// 404 every other API would return.
func isUptimeGone(err error) bool {
	if isNotFound(err) {
		return true
	}
	var errResp *godo.ErrorResponse
	return errors.As(err, &errResp) && errResp.Response != nil && errResp.Response.StatusCode == http.StatusForbidden
}

func (*uptimeCheckVerifier) IDOutputKey() string { return "check_id" }

func (*uptimeCheckVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceanuptimecheck requires the full outputs map (check_id + alert_ids); " +
		"the harness dispatches through VerifyExistsFromOutputs")
}

func (*uptimeCheckVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceanuptimecheck requires the full outputs map (check_id + alert_ids); " +
		"the harness dispatches through VerifyAbsentFromOutputs")
}

func (v *uptimeCheckVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	checkID := StringOutput(outputs, "check_id")
	if checkID == "" {
		return pkgerrors.New("digitaloceanuptimecheck outputs carry no check_id")
	}
	check, _, err := client.UptimeChecks.Get(ctx, checkID)
	if err != nil {
		if isUptimeGone(err) {
			return pkgerrors.Errorf("digitaloceanuptimecheck %q not found after deploy", checkID)
		}
		return pkgerrors.Wrap(err, "digitaloceanuptimecheck verify-exists failed")
	}
	if check.ID == "" {
		return pkgerrors.Errorf("digitaloceanuptimecheck %q returned an empty check", checkID)
	}
	for key, alertID := range StringMapOutput(outputs, "alert_ids") {
		alert, _, err := client.UptimeChecks.GetAlert(ctx, checkID, alertID)
		if err != nil {
			if isUptimeGone(err) {
				return pkgerrors.Errorf("digitaloceanuptimecheck alert row %q (%s) not found after deploy", key, alertID)
			}
			return pkgerrors.Wrapf(err, "digitaloceanuptimecheck alert row %q verify-exists failed", key)
		}
		if alert.ID != alertID {
			return pkgerrors.Errorf("digitaloceanuptimecheck alert row %q: API returned id %q for %q", key, alert.ID, alertID)
		}
	}
	return nil
}

func (v *uptimeCheckVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	checkID := StringOutput(outputs, "check_id")
	if checkID == "" {
		return pkgerrors.New("digitaloceanuptimecheck outputs carry no check_id")
	}
	_, _, err := client.UptimeChecks.Get(ctx, checkID)
	if err == nil {
		return &StillExistsError{Component: "digitaloceanuptimecheck", ID: checkID}
	}
	if !isUptimeGone(err) {
		return pkgerrors.Wrap(err, "digitaloceanuptimecheck verify-absent failed")
	}
	for key, alertID := range StringMapOutput(outputs, "alert_ids") {
		_, _, err := client.UptimeChecks.GetAlert(ctx, checkID, alertID)
		if err == nil {
			return &StillExistsError{Component: "digitaloceanuptimecheck", ID: alertID, Detail: "alert row " + key + " outlived its check"}
		}
		if !isUptimeGone(err) {
			return pkgerrors.Wrapf(err, "digitaloceanuptimecheck alert row %q verify-absent failed", key)
		}
	}
	return nil
}
