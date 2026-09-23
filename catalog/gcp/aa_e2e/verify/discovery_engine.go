package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// discoveryEngineGet reads one Discovery Engine resource by its full name
// through the REST API (the pinned client library carries no Discovery
// Engine client). The host is location-prefixed exactly as Google's
// provider builds it -- https://{location}-discoveryengine.googleapis.com
// -- for every location including "global". The location is the fourth
// segment of every Discovery Engine resource name
// (projects/{p}/locations/{l}/...). The decoded body is returned as a
// generic map so each verifier asserts the fields it cares about; the
// status code is returned so callers can tell a 404 from any other error.
func discoveryEngineGet(ctx context.Context, svc *Services, name string) (map[string]interface{}, int, error) {
	location := discoveryEngineLocation(name)
	if location == "" {
		return nil, 0, errors.Errorf("discovery engine resource name %q carries no location segment", name)
	}
	url := fmt.Sprintf("https://%s-discoveryengine.googleapis.com/v1/%s", location, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to build discovery engine GET request")
	}
	resp, err := svc.RestClient.Do(req)
	if err != nil {
		return nil, 0, errors.Wrap(err, "discovery engine GET request failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, errors.Wrap(err, "failed to read discovery engine response")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, errors.Errorf("discovery engine GET %s returned %d: %s", name, resp.StatusCode, string(body))
	}

	obj := map[string]interface{}{}
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil, resp.StatusCode, errors.Wrap(err, "failed to decode discovery engine resource")
	}
	return obj, resp.StatusCode, nil
}

// discoveryEngineLocation extracts the location from a Discovery Engine
// resource name (projects/{p}/locations/{l}/...).
func discoveryEngineLocation(name string) string {
	parts := strings.Split(name, "/")
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == "locations" {
			return parts[i+1]
		}
	}
	return ""
}

// discoveryEngineAbsent turns a probe result into the destroy verdict every
// Discovery Engine verifier shares: a 404 means gone, any other error is
// unexpected, and a 200 means the resource survived.
func discoveryEngineAbsent(what, name string, status int, err error) error {
	if err != nil {
		if status == http.StatusNotFound {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing %s %s after destroy", what, name)
	}
	return errors.Errorf("%s %s still exists after destroy", what, name)
}

// listOutput collects a list output as the outputs transformer flattens it
// -- one dot-indexed key per element (key.0, key.1, ...), in manifest
// order.
func listOutput(outputs map[string]string, key string) []string {
	var values []string
	for i := 0; ; i++ {
		value, ok := outputs[fmt.Sprintf("%s.%d", key, i)]
		if !ok {
			break
		}
		if strings.TrimSpace(value) != "" {
			values = append(values, value)
		}
	}
	return values
}

// stringField reads a top-level string field of a decoded resource.
func stringField(obj map[string]interface{}, key string) string {
	if value, ok := obj[key].(string); ok {
		return value
	}
	return ""
}
