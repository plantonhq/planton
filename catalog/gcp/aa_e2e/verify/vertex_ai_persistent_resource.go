package verify

import (
	"context"

	"github.com/pkg/errors"
)

// vertexAiPersistentResourceVerifier probes a Vertex AI persistent
// resource through the REST API: it reads back with the attribution labels
// and in the RUNNING state (every pool provisioned), and the
// persistent_resource_id output is the id at the end of its name.
type vertexAiPersistentResourceVerifier struct{}

// IDOutputKey is the resource's full name
// (projects/{p}/locations/{l}/persistentResources/{id}).
func (v *vertexAiPersistentResourceVerifier) IDOutputKey() string { return "name" }

func (v *vertexAiPersistentResourceVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	resource, err := vertexAiExistsWithLabel(ctx, svc, "vertex ai persistent resource", name)
	if err != nil {
		return err
	}
	if state := stringField(resource, "state"); state != "RUNNING" {
		return errors.Errorf("vertex ai persistent resource %s is in state %q after deploy, want RUNNING", name, state)
	}
	if got := outputs["persistent_resource_id"]; got != lastPathSegment(name) {
		return errors.Errorf("vertex ai persistent resource %s persistent_resource_id output %q does not match its name", name, got)
	}
	return nil
}

func (v *vertexAiPersistentResourceVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return vertexAiAbsent(ctx, svc, "vertex ai persistent resource", outputs["name"])
}
