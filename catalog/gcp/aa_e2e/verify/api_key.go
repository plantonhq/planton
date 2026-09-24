package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// apiKeyVerifier probes an API key by its full resource name
// (projects/{p}/locations/global/keys/{key_id}) through the API Keys API —
// the name output is exactly the handle the API addresses the key by, so
// verifying with it doubles as proof the output is honest. The key STRING
// is never read here: the metadata GET deliberately omits it, and the
// verifier has no business handling the credential.
type apiKeyVerifier struct{}

func (v *apiKeyVerifier) IDOutputKey() string { return "name" }

func (v *apiKeyVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	key, err := svc.ApiKeys.Projects.Locations.Keys.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "api key %s not found after deploy", name)
	}
	if key.DeleteTime != "" {
		return errors.Errorf("api key %s is soft-deleted (deleteTime %s) right after deploy", name, key.DeleteTime)
	}
	if uid := outputs["uid"]; uid != "" && key.Uid != uid {
		return errors.Errorf("api key %s resolved to uid %q, want the uid output %q", name, key.Uid, uid)
	}
	return nil
}

func (v *apiKeyVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	// Deleting a key is a SOFT delete: Google keeps it recoverable for 30
	// days, so "gone" has two honest shapes -- the API answers 404 for the
	// name, or returns the key carrying a deleteTime. Either is the
	// destroyed state; a live key with no deleteTime is the failure.
	key, err := svc.ApiKeys.Projects.Locations.Keys.Get(name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing api key %s after destroy", name)
	}
	if key.DeleteTime == "" {
		return errors.Errorf("api key %s still exists and is not soft-deleted after destroy", name)
	}
	return nil
}
