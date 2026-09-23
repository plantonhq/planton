package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// vertexAiGet reads one Vertex AI resource by its full name through the
// REST API on the regional host Google's provider uses
// (https://{region}-aiplatform.googleapis.com/v1/). The decoded body is
// returned as a generic map so each verifier asserts the fields it cares
// about; the status code is returned so callers can tell a 404 from any
// other error. Shared by the Vertex AI kinds whose resources the pinned
// client library has no typed client for at the harness's granularity.
func vertexAiGet(ctx context.Context, svc *Services, name string) (map[string]interface{}, int, error) {
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/%s", regionFromVertexResource(name), name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to build vertex ai GET request")
	}
	resp, err := svc.RestClient.Do(req)
	if err != nil {
		return nil, 0, errors.Wrap(err, "vertex ai GET request failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, errors.Wrap(err, "failed to read vertex ai response")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, errors.Errorf("vertex ai GET %s returned %d: %s", name, resp.StatusCode, string(body))
	}

	obj := map[string]interface{}{}
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil, resp.StatusCode, errors.Wrap(err, "failed to decode vertex ai resource")
	}
	return obj, resp.StatusCode, nil
}

// vertexAiExistsWithLabel reads a Vertex AI resource and asserts the
// platform attribution label landed -- the cross-engine label-parity canary
// every labelled Vertex AI resource shares.
func vertexAiExistsWithLabel(ctx context.Context, svc *Services, what, name string) (map[string]interface{}, error) {
	obj, _, err := vertexAiGet(ctx, svc, name)
	if err != nil {
		return nil, errors.Wrapf(err, "%s %s not found after deploy", what, name)
	}
	labels, _ := obj["labels"].(map[string]interface{})
	if labels["planton-ai_resource"] != "true" {
		return nil, errors.Errorf("%s %s missing the planton-ai_resource attribution label after deploy", what, name)
	}
	return obj, nil
}

// vertexAiAbsent is the destroy verdict every Vertex AI verifier shares: a
// 404 means gone, any other error is unexpected, and a 200 means the
// resource survived.
func vertexAiAbsent(ctx context.Context, svc *Services, what, name string) error {
	if name == "" {
		return nil
	}
	_, status, err := vertexAiGet(ctx, svc, name)
	if err != nil {
		if status == http.StatusNotFound {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing %s %s after destroy", what, name)
	}
	return errors.Errorf("%s %s still exists after destroy", what, name)
}
