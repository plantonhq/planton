package verify

import (
	"context"

	"github.com/pkg/errors"
)

// cloudDeployTargetVerifier probes a deployment target through the Cloud
// Deploy API by the name output
// (projects/{p}/locations/{l}/targets/{id}). Posture assertions: Google's
// uid and target ID match the exported ones, and the platform attribution
// labels landed.
type cloudDeployTargetVerifier struct{}

func (v *cloudDeployTargetVerifier) IDOutputKey() string { return "name" }

func (v *cloudDeployTargetVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	target, err := svc.CloudDeploy.Projects.Locations.Targets.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "deploy target %s not found after deploy", name)
	}
	if uid := outputs["uid"]; uid != "" && target.Uid != uid {
		return errors.Errorf("deploy target %s uid output %q does not match live uid %q", name, uid, target.Uid)
	}
	if id := outputs["target_id"]; id != "" && target.TargetId != id {
		return errors.Errorf("deploy target %s target_id output %q does not match live target ID %q", name, id, target.TargetId)
	}
	if target.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("deploy target %s missing the planton-ai_resource attribution label after deploy (labels: %v)", name, target.Labels)
	}
	return nil
}

func (v *cloudDeployTargetVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, err := svc.CloudDeploy.Projects.Locations.Targets.Get(name).Context(ctx).Do()
	if err != nil {
		if isGoogleNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing deploy target %s after destroy", name)
	}
	return errors.Errorf("deploy target %s still exists after destroy", name)
}
