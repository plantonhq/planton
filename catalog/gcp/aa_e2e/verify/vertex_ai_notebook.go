package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// vertexAiNotebookVerifier probes a Vertex AI Workbench instance via the
// Notebooks REST API. The pinned google.golang.org/api line has no typed
// Workbench client, so the probe is a plain authenticated GET on the
// instance's documented v2 resource path. Posture assertions confirm the
// platform attribution labels landed (the label-parity proof) and that the
// instance is in an operational state.
type vertexAiNotebookVerifier struct{}

func (v *vertexAiNotebookVerifier) IDOutputKey() string { return "instance_id" }

type vertexAiNotebook struct {
	Name   string            `json:"name"`
	State  string            `json:"state"`
	Labels map[string]string `json:"labels"`
}

func (v *vertexAiNotebookVerifier) get(ctx context.Context, svc *Services, instanceID string) (*vertexAiNotebook, int, error) {
	instance := &vertexAiNotebook{}
	status, err := googleRestGet(ctx, svc, "workbench instance", fmt.Sprintf("https://notebooks.googleapis.com/v2/%s", instanceID), instance)
	if err != nil {
		return nil, status, err
	}
	return instance, status, nil
}

func (v *vertexAiNotebookVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	instanceID := outputs["instance_id"]
	if instanceID == "" {
		return errors.New("instance_id output missing after deploy")
	}

	instance, _, err := v.get(ctx, svc, instanceID)
	if err != nil {
		return errors.Wrapf(err, "workbench instance %s not found after deploy", instanceID)
	}

	if instance.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("workbench instance %s missing the planton-ai_resource attribution label after deploy", instanceID)
	}

	// A freshly created notebook should be running or still initializing;
	// anything terminal means the deploy left a broken posture.
	switch instance.State {
	case "ACTIVE", "INITIALIZING", "STARTING":
	default:
		return errors.Errorf("workbench instance %s in state %s after deploy (expected ACTIVE, INITIALIZING, or STARTING)", instanceID, instance.State)
	}

	if got := outputs["instance_name"]; got != "" {
		liveName := lastPathSegment(instance.Name)
		if got != liveName {
			return errors.Errorf("workbench instance %s instance_name output %q does not match live name %q", instanceID, got, liveName)
		}
	}
	return nil
}

func (v *vertexAiNotebookVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	instanceID := outputs["instance_id"]
	if instanceID == "" {
		return nil
	}

	_, status, err := v.get(ctx, svc, instanceID)
	return restAbsent("workbench instance", instanceID, status, err)
}
