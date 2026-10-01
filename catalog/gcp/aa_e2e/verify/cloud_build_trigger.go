package verify

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	cloudbuild "google.golang.org/api/cloudbuild/v1"
)

// cloudBuildTriggerVerifier probes a build trigger through the Cloud Build
// v1 API by the id output. Regional triggers carry the full name
// (projects/{p}/locations/{l}/triggers/{id}); global triggers carry the
// provider's legacy form (projects/{p}/triggers/{id}), read through the
// project-level method. Posture assertion: Google's trigger id matches the
// exported trigger_id. Triggers carry no labels, so the attribution canary
// does not apply.
type cloudBuildTriggerVerifier struct{}

func (v *cloudBuildTriggerVerifier) IDOutputKey() string { return "id" }

func (v *cloudBuildTriggerVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	id := outputs["id"]
	if id == "" {
		return errors.New("id output missing after deploy")
	}
	trigger, err := v.get(ctx, svc, id)
	if err != nil {
		return errors.Wrapf(err, "trigger %s not found after deploy", id)
	}
	if triggerID := outputs["trigger_id"]; triggerID != "" && trigger.Id != triggerID {
		return errors.Errorf("trigger %s trigger_id output %q does not match live id %q", id, triggerID, trigger.Id)
	}
	if name := outputs["name"]; name != "" && trigger.Name != name {
		return errors.Errorf("trigger %s name output %q does not match live name %q", id, name, trigger.Name)
	}
	return nil
}

func (v *cloudBuildTriggerVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	id := outputs["id"]
	if id == "" {
		return nil
	}
	_, err := v.get(ctx, svc, id)
	if err == nil {
		return errors.Errorf("trigger %s still exists after destroy", id)
	}
	if isGoogleNotFound(err) {
		return nil
	}
	return errors.Wrapf(err, "unexpected error probing trigger %s after destroy", id)
}

func (v *cloudBuildTriggerVerifier) get(ctx context.Context, svc *Services, id string) (*cloudbuild.BuildTrigger, error) {
	parts := strings.Split(id, "/")
	switch {
	case len(parts) == 6 && parts[0] == "projects" && parts[2] == "locations" && parts[4] == "triggers":
		return svc.CloudBuild.Projects.Locations.Triggers.Get(id).Context(ctx).Do()
	case len(parts) == 4 && parts[0] == "projects" && parts[2] == "triggers":
		return svc.CloudBuild.Projects.Triggers.Get(parts[1], parts[3]).Context(ctx).Do()
	default:
		return nil, errors.Errorf("trigger id %q is neither projects/{p}/locations/{l}/triggers/{id} nor projects/{p}/triggers/{id}", id)
	}
}
