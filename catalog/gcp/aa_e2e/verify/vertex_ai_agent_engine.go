package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// vertexAiAgentEngineVerifier probes an Agent Engine instance (Google's
// ReasoningEngine) through the Vertex AI REST API on its regional host (the
// pinned client library has no typed surface for it). Existence, the
// platform attribution labels, and the numeric id both engines export are
// asserted from the JSON body.
type vertexAiAgentEngineVerifier struct{}

// IDOutputKey is the agent's full resource name
// (projects/{p}/locations/{l}/reasoningEngines/{id}).
func (v *vertexAiAgentEngineVerifier) IDOutputKey() string { return "name" }

// vertexAiReasoningEngine is the subset of the API's ReasoningEngine object
// the verifier asserts on.
type vertexAiReasoningEngine struct {
	Name        string            `json:"name"`
	DisplayName string            `json:"displayName"`
	Labels      map[string]string `json:"labels"`
	CreateTime  string            `json:"createTime"`
}

func (v *vertexAiAgentEngineVerifier) get(ctx context.Context, svc *Services, name string) (*vertexAiReasoningEngine, int, error) {
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/%s", regionFromVertexResource(name), name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to build agent engine GET request")
	}
	resp, err := svc.RestClient.Do(req)
	if err != nil {
		return nil, 0, errors.Wrap(err, "agent engine GET request failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, errors.Wrap(err, "failed to read agent engine response")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, errors.Errorf("agent engine GET %s returned %d: %s", name, resp.StatusCode, string(body))
	}

	engine := &vertexAiReasoningEngine{}
	if err := json.Unmarshal(body, engine); err != nil {
		return nil, resp.StatusCode, errors.Wrap(err, "failed to decode agent engine")
	}
	return engine, resp.StatusCode, nil
}

// VerifyExists confirms the agent exists under the exported name with the
// attribution labels, and that reasoning_engine_id is its numeric last
// segment on both engines.
func (v *vertexAiAgentEngineVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}

	engine, _, err := v.get(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "agent engine %s not found after deploy", name)
	}
	if engine.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("agent engine %s missing the planton-ai_resource attribution label after deploy", name)
	}
	if got := outputs["reasoning_engine_id"]; got != "" && got != lastPathSegment(engine.Name) {
		return errors.Errorf("agent engine %s reasoning_engine_id output %q does not match live name %q", name, got, lastPathSegment(engine.Name))
	}
	if location := outputs["location"]; location != "" && regionFromVertexResource(name) != location {
		return errors.Errorf("agent engine %s location output %q does not match the resource path", name, location)
	}
	return nil
}

// VerifyAbsent confirms the agent is gone after destroy.
func (v *vertexAiAgentEngineVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}

	_, status, err := v.get(ctx, svc, name)
	if err != nil {
		if status == http.StatusNotFound {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing agent engine %s after destroy", name)
	}
	return errors.Errorf("agent engine %s still exists after destroy", name)
}
