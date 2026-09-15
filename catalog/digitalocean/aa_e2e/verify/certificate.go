package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// certificateVerifier verifies a DigitalOceanCertificate. The kind's
// certificate_id output carries the certificate resource's id, which at the
// provider pin is the certificate NAME, not its UUID -- a deliberate upstream
// choice, because a lets_encrypt certificate's UUID rotates on every
// auto-renewal while the name is stable. The verifier therefore looks the
// certificate up by name, exactly as the upstream provider's own Read does.
type certificateVerifier struct{}

func (*certificateVerifier) IDOutputKey() string { return "certificate_id" }

func (*certificateVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	exists, err := certificateExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceancertificate verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceancertificate %q not found after deploy", id)
	}
	return nil
}

func (*certificateVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	exists, err := certificateExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceancertificate verify-absent failed for %q", id)
	}
	if exists {
		return &StillExistsError{Component: "digitaloceancertificate", ID: id}
	}
	return nil
}

func certificateExists(ctx context.Context, client *godo.Client, name string) (bool, error) {
	certs, _, err := client.Certificates.ListByName(ctx, name, nil)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return len(certs) > 0, nil
}
