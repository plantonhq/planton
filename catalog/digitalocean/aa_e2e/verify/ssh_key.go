package verify

import (
	"context"
	"strconv"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// sshKeyVerifier verifies a DigitalOceanSshKey via
// GET /v2/account/keys/{ssh_key_id}. The id is the key's numeric id (the
// stack output is string-typed by contract); the API also accepts
// fingerprints on this endpoint, but the output always carries the numeric
// id -- the same identity imports require.
type sshKeyVerifier struct{}

func (*sshKeyVerifier) IDOutputKey() string { return "ssh_key_id" }

func (v *sshKeyVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	keyID, err := strconv.Atoi(id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceansshkey id %q is not the numeric key id", id)
	}
	key, _, err := client.Keys.GetByID(ctx, keyID)
	if err != nil {
		if isNotFound(err) {
			return pkgerrors.Errorf("digitaloceansshkey %q not found after deploy", id)
		}
		return pkgerrors.Wrap(err, "digitaloceansshkey verify-exists failed")
	}
	if key.Fingerprint == "" {
		return pkgerrors.Errorf("digitaloceansshkey %q returned an empty fingerprint", id)
	}
	return nil
}

func (v *sshKeyVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	keyID, err := strconv.Atoi(id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceansshkey id %q is not the numeric key id", id)
	}
	_, _, err = client.Keys.GetByID(ctx, keyID)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return pkgerrors.Wrap(err, "digitaloceansshkey verify-absent failed")
	}
	return &StillExistsError{Component: "digitaloceansshkey", ID: id}
}
