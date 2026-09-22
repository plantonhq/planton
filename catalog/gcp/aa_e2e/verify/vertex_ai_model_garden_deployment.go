package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// vertexAiModelGardenDeploymentVerifier probes the endpoint a one-step
// Model Garden deployment created, through the Vertex AI REST API on the
// endpoint's regional host. The deployment itself is not a readable
// resource; the proof is the endpoint existing under the exported path with
// the exported deployed model on it.
type vertexAiModelGardenDeploymentVerifier struct{}

// IDOutputKey is the endpoint's full resource path
// (projects/{p}/locations/{l}/endpoints/{id}) -- the same key
// GcpVertexAiEndpoint exports.
func (v *vertexAiModelGardenDeploymentVerifier) IDOutputKey() string { return "endpoint_id" }

// vertexAiDeployedEndpoint is the subset of the API's Endpoint object the
// verifier asserts on.
type vertexAiDeployedEndpoint struct {
	Name           string `json:"name"`
	DeployedModels []struct {
		Id          string `json:"id"`
		DisplayName string `json:"displayName"`
		Model       string `json:"model"`
	} `json:"deployedModels"`
}

func (v *vertexAiModelGardenDeploymentVerifier) get(ctx context.Context, svc *Services, endpointID string) (*vertexAiDeployedEndpoint, int, error) {
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/%s", regionFromVertexResource(endpointID), endpointID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to build model garden endpoint GET request")
	}
	resp, err := svc.RestClient.Do(req)
	if err != nil {
		return nil, 0, errors.Wrap(err, "model garden endpoint GET request failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, errors.Wrap(err, "failed to read model garden endpoint response")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, errors.Errorf("model garden endpoint GET %s returned %d: %s", endpointID, resp.StatusCode, string(body))
	}

	endpoint := &vertexAiDeployedEndpoint{}
	if err := json.Unmarshal(body, endpoint); err != nil {
		return nil, resp.StatusCode, errors.Wrap(err, "failed to decode model garden endpoint")
	}
	return endpoint, resp.StatusCode, nil
}

// VerifyExists confirms the endpoint exists under the exported path, that
// endpoint_name is its numeric last segment, and that the exported
// deployed_model_id names a model deployed on it -- the cross-engine
// determinism contract both modules share.
func (v *vertexAiModelGardenDeploymentVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	endpointID := outputs["endpoint_id"]
	if endpointID == "" {
		return errors.New("endpoint_id output missing after deploy")
	}

	endpoint, _, err := v.get(ctx, svc, endpointID)
	if err != nil {
		return errors.Wrapf(err, "model garden endpoint %s not found after deploy", endpointID)
	}
	if got := outputs["endpoint_name"]; got != "" && got != lastPathSegment(endpoint.Name) {
		return errors.Errorf("model garden endpoint %s endpoint_name output %q does not match live name %q", endpointID, got, lastPathSegment(endpoint.Name))
	}

	deployedModelID := outputs["deployed_model_id"]
	if deployedModelID == "" {
		return errors.Errorf("model garden endpoint %s: deployed_model_id output missing after deploy", endpointID)
	}
	for _, deployed := range endpoint.DeployedModels {
		if deployed.Id == deployedModelID {
			if want := outputs["deployed_model_display_name"]; want != "" && deployed.DisplayName != want {
				return errors.Errorf("model garden endpoint %s deployed model %s display name %q does not match output %q", endpointID, deployedModelID, deployed.DisplayName, want)
			}
			return nil
		}
	}
	return errors.Errorf("model garden endpoint %s has no deployed model %s after deploy (%d deployed)", endpointID, deployedModelID, len(endpoint.DeployedModels))
}

// VerifyAbsent confirms the endpoint is gone after the destroy undeployed
// the model and deleted the endpoint.
func (v *vertexAiModelGardenDeploymentVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	endpointID := outputs["endpoint_id"]
	if endpointID == "" {
		return nil
	}

	_, status, err := v.get(ctx, svc, endpointID)
	if err != nil {
		if status == http.StatusNotFound {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing model garden endpoint %s after destroy", endpointID)
	}
	return errors.Errorf("model garden endpoint %s still exists after destroy", endpointID)
}
