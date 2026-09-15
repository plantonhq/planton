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
// id -- the same identity imports require. The fingerprint output is
// asserted against the live key too: DigitalOcean computes it from the
// material, so a match proves the key that landed is the key that was sent.
type sshKeyVerifier struct{}

func (*sshKeyVerifier) IDOutputKey() string { return "ssh_key_id" }

func (v *sshKeyVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceansshkey requires the full outputs map (ssh_key_id + fingerprint); " +
		"the harness dispatches through VerifyExistsFromOutputs")
}

func (v *sshKeyVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceansshkey requires the full outputs map (ssh_key_id + fingerprint); " +
		"the harness dispatches through VerifyAbsentFromOutputs")
}

func (v *sshKeyVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "ssh_key_id")
	keyID, err := numericKeyID(id)
	if err != nil {
		return err
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
	if want := StringOutput(outputs, "fingerprint"); want != "" && want != key.Fingerprint {
		return pkgerrors.Errorf("digitaloceansshkey %q fingerprint output %q does not match the live key's %q", id, want, key.Fingerprint)
	}
	return nil
}

func (v *sshKeyVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "ssh_key_id")
	keyID, err := numericKeyID(id)
	if err != nil {
		return err
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

// numericKeyID guards the identity the output claims: a fingerprint
// arriving here would still resolve at the HTTP level, so the assertion
// stays honest about which id the contract carries.
func numericKeyID(id string) (int, error) {
	keyID, err := strconv.Atoi(id)
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "digitaloceansshkey id %q is not the numeric key id", id)
	}
	return keyID, nil
}
