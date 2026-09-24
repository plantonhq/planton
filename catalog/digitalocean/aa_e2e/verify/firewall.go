package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// firewallVerifier verifies a DigitalOceanFirewall via GET /v2/firewalls/{id}.
type firewallVerifier struct{}

func (*firewallVerifier) IDOutputKey() string { return "firewall_id" }

func (*firewallVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	firewall, _, err := client.Firewalls.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return pkgerrors.Errorf("digitaloceanfirewall %q not found after deploy", id)
		}
		return pkgerrors.Wrapf(err, "digitaloceanfirewall verify-exists failed for %q", id)
	}
	// "waiting" is a legal transient while rules propagate to targets;
	// "failed" is never legal.
	if firewall.Status == "failed" {
		return pkgerrors.Errorf("digitaloceanfirewall %q exists but its live status is failed", id)
	}
	return nil
}

func (*firewallVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	exists, err := firewallExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceanfirewall verify-absent failed for %q", id)
	}
	if exists {
		return &StillExistsError{Component: "digitaloceanfirewall", ID: id}
	}
	return nil
}

func firewallExists(ctx context.Context, client *godo.Client, id string) (bool, error) {
	_, _, err := client.Firewalls.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
