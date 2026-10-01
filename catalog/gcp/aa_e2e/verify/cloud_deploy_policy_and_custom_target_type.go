package verify

import (
	"context"

	"github.com/pkg/errors"
)

// cloudDeployPolicyVerifier probes a deploy policy through the Cloud Deploy
// API by the name output (projects/{p}/locations/{l}/deployPolicies/{id}).
// Posture assertions: Google's uid matches the exported one, and the
// attribution canary label is present.
type cloudDeployPolicyVerifier struct{}

func (v *cloudDeployPolicyVerifier) IDOutputKey() string { return "name" }

func (v *cloudDeployPolicyVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	policy, err := svc.CloudDeploy.Projects.Locations.DeployPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "deploy policy %s not found after deploy", name)
	}
	if uid := outputs["uid"]; uid != "" && policy.Uid != uid {
		return errors.Errorf("deploy policy %s uid output %q does not match live uid %q", name, uid, policy.Uid)
	}
	if policy.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("deploy policy %s missing the planton-ai_resource attribution label after deploy (labels: %v)", name, policy.Labels)
	}
	return nil
}

func (v *cloudDeployPolicyVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, err := svc.CloudDeploy.Projects.Locations.DeployPolicies.Get(name).Context(ctx).Do()
	if err == nil {
		return errors.Errorf("deploy policy %s still exists after destroy", name)
	}
	if isGoogleNotFound(err) {
		return nil
	}
	return errors.Wrapf(err, "unexpected error probing deploy policy %s after destroy", name)
}

// cloudDeployCustomTargetTypeVerifier probes a custom target type through
// the Cloud Deploy API by the name output
// (projects/{p}/locations/{l}/customTargetTypes/{id}). Posture assertions:
// Google's uid matches the exported one, and the attribution canary label
// is present.
type cloudDeployCustomTargetTypeVerifier struct{}

func (v *cloudDeployCustomTargetTypeVerifier) IDOutputKey() string { return "name" }

func (v *cloudDeployCustomTargetTypeVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	targetType, err := svc.CloudDeploy.Projects.Locations.CustomTargetTypes.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "custom target type %s not found after deploy", name)
	}
	if uid := outputs["uid"]; uid != "" && targetType.Uid != uid {
		return errors.Errorf("custom target type %s uid output %q does not match live uid %q", name, uid, targetType.Uid)
	}
	if targetType.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("custom target type %s missing the planton-ai_resource attribution label after deploy (labels: %v)", name, targetType.Labels)
	}
	return nil
}

func (v *cloudDeployCustomTargetTypeVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, err := svc.CloudDeploy.Projects.Locations.CustomTargetTypes.Get(name).Context(ctx).Do()
	if err == nil {
		return errors.Errorf("custom target type %s still exists after destroy", name)
	}
	if isGoogleNotFound(err) {
		return nil
	}
	return errors.Wrapf(err, "unexpected error probing custom target type %s after destroy", name)
}
