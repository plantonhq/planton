package verify

import (
	"context"
	"strconv"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// dropletVerifier verifies a DigitalOceanDroplet via GET /v2/droplets/{id}.
// Droplet ids are integers in the API; the stack output carries the decimal
// string form. Beyond existence, it asserts the live droplet is active and
// checks the public IPv4 address the module CLAIMS in its stack outputs
// against the live droplet -- outputs are contractually identical across
// both engines, so one assertion protects both, and an absent output simply
// means "not claimed" and is skipped. Status is always read live, never
// from an output: an apply-time snapshot goes stale immediately.
type dropletVerifier struct{}

func (*dropletVerifier) IDOutputKey() string { return "droplet_id" }

func (*dropletVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	_, err := getDroplet(ctx, client, id)
	return err
}

func (*dropletVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	numericID, err := strconv.Atoi(id)
	if err != nil {
		return pkgerrors.Wrapf(err, "droplet id %q is not the API's integer id", id)
	}
	_, _, err = client.Droplets.Get(ctx, numericID)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return pkgerrors.Wrapf(err, "digitaloceandroplet verify-absent failed for %q", id)
	}
	return &StillExistsError{Component: "digitaloceandroplet", ID: id}
}

func (v *dropletVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "droplet_id")
	if id == "" {
		return pkgerrors.New("droplet_id output missing after deploy")
	}

	droplet, err := getDroplet(ctx, client, id)
	if err != nil {
		return err
	}

	if droplet.Status != "active" {
		return pkgerrors.Errorf("digitaloceandroplet %q status is %q, want active", id, droplet.Status)
	}

	if ipv4 := StringOutput(outputs, "ipv4_address"); ipv4 != "" {
		livePublicIPv4, err := droplet.PublicIPv4()
		if err != nil {
			return pkgerrors.Wrapf(err, "digitaloceandroplet %q public ipv4 read failed", id)
		}
		if livePublicIPv4 != ipv4 {
			return pkgerrors.Errorf("digitaloceandroplet %q ipv4 mismatch: output %q, live %q",
				id, ipv4, livePublicIPv4)
		}
	}

	return nil
}

func (v *dropletVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "droplet_id")
	if id == "" {
		return pkgerrors.New("droplet_id output missing for destroy verification")
	}
	return v.VerifyAbsent(ctx, client, id)
}

func getDroplet(ctx context.Context, client *godo.Client, id string) (*godo.Droplet, error) {
	numericID, err := strconv.Atoi(id)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "droplet id %q is not the API's integer id", id)
	}
	droplet, _, err := client.Droplets.Get(ctx, numericID)
	if err != nil {
		if isNotFound(err) {
			return nil, pkgerrors.Errorf("digitaloceandroplet %q not found after deploy", id)
		}
		return nil, pkgerrors.Wrapf(err, "digitaloceandroplet verify-exists failed for %q", id)
	}
	return droplet, nil
}
