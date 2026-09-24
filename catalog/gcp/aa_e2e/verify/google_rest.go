package verify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// googleRestGet reads one Google Cloud resource with a plain authenticated
// GET on its REST URL and decodes the body into out -- a typed struct
// holding the fields a verifier asserts on, or a map for a generic read.
// It is the one probe for every service the pinned client library has no
// typed client for; per-service helpers only build the URL (the host
// Google's provider uses plus the resource name). The status code is
// returned so callers can tell a 404 (see restAbsent) from any other
// error; what names the resource in error messages.
func googleRestGet(ctx context.Context, svc *Services, what, url string, out interface{}) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to build %s GET request", what)
	}
	resp, err := svc.RestClient.Do(req)
	if err != nil {
		return 0, errors.Wrapf(err, "%s GET request failed", what)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, errors.Wrapf(err, "failed to read %s response", what)
	}
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, errors.Errorf("%s GET %s returned %d: %s", what, url, resp.StatusCode, string(body))
	}
	if err := json.Unmarshal(body, out); err != nil {
		return resp.StatusCode, errors.Wrapf(err, "failed to decode %s", what)
	}
	return resp.StatusCode, nil
}

// resourceLocation extracts the location segment from a Google resource
// name (projects/{p}/locations/{l}/...) -- the part location-prefixed hosts
// such as Discovery Engine's and Dialogflow's are built from. Empty when
// the name carries no location.
func resourceLocation(name string) string {
	parts := strings.Split(name, "/")
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == "locations" {
			return parts[i+1]
		}
	}
	return ""
}

// restAbsent turns a googleRestGet result into the destroy verdict every
// REST-probed verifier shares: a 404 means gone, any other error is
// unexpected, and a 200 means the resource survived.
func restAbsent(what, name string, status int, err error) error {
	if err != nil {
		if status == http.StatusNotFound {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing %s %s after destroy", what, name)
	}
	return errors.Errorf("%s %s still exists after destroy", what, name)
}
