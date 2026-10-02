package verify

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// isGkeHubNotFound reports whether a GKE Hub read answered 404.
func isGkeHubNotFound(err error) bool {
	var apiErr *googleapi.Error
	return errors.As(err, &apiErr) && apiErr.Code == 404
}

// gkeFleetVerifier probes the project's fleet through the Fleet API by the
// name output (projects/{p}/locations/global/fleets/default). Posture
// assertions confirm Google's uid matches the exported one. Fleets carry
// no labels (the pinned Pulumi SDK has no fleet labels argument), so the
// attribution canary does not apply.
type gkeFleetVerifier struct{}

func (v *gkeFleetVerifier) IDOutputKey() string { return "name" }

func (v *gkeFleetVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	fleet, err := svc.GkeHub.Projects.Locations.Fleets.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "fleet %s not found after deploy", name)
	}
	if uid := outputs["uid"]; uid != "" && fleet.Uid != uid {
		return errors.Errorf("fleet %s uid output %q does not match live uid %q", name, uid, fleet.Uid)
	}
	return nil
}

func (v *gkeFleetVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	fleet, err := svc.GkeHub.Projects.Locations.Fleets.Get(name).Context(ctx).Do()
	if err != nil {
		if isGkeHubNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing fleet %s after destroy", name)
	}
	if fleet.DeleteTime != "" {
		return nil
	}
	return errors.Errorf("fleet %s still exists after destroy", name)
}

// gkeFleetFeatureVerifier probes a fleet feature by the name output
// (projects/{p}/locations/{l}/features/{feature}). Posture assertions
// confirm the platform attribution labels landed. Google never deletes
// "rbacrolebindingactuation" -- destroy empties its allowlist -- so for
// that feature the absence check accepts an empty allowlist.
type gkeFleetFeatureVerifier struct{}

func (v *gkeFleetFeatureVerifier) IDOutputKey() string { return "name" }

func (v *gkeFleetFeatureVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	feature, err := svc.GkeHub.Projects.Locations.Features.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "fleet feature %s not found after deploy", name)
	}
	if feature.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("fleet feature %s missing the planton-ai_resource attribution label after deploy (labels: %v)", name, feature.Labels)
	}
	return nil
}

func (v *gkeFleetFeatureVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	feature, err := svc.GkeHub.Projects.Locations.Features.Get(name).Context(ctx).Do()
	if err != nil {
		if isGkeHubNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing fleet feature %s after destroy", name)
	}
	if feature.DeleteTime != "" {
		return nil
	}
	if strings.HasSuffix(name, "/features/rbacrolebindingactuation") {
		if feature.Spec == nil || feature.Spec.Rbacrolebindingactuation == nil || len(feature.Spec.Rbacrolebindingactuation.AllowedCustomRoles) == 0 {
			return nil
		}
		return errors.Errorf("fleet feature %s still allows custom roles after destroy: %v", name, feature.Spec.Rbacrolebindingactuation.AllowedCustomRoles)
	}
	return errors.Errorf("fleet feature %s still exists after destroy", name)
}

// gkeFleetScopeVerifier probes a team scope by the name output
// (projects/{p}/locations/global/scopes/{id}). Posture assertions confirm
// the attribution labels landed and the scope ID matches the export.
type gkeFleetScopeVerifier struct{}

func (v *gkeFleetScopeVerifier) IDOutputKey() string { return "name" }

func (v *gkeFleetScopeVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	scope, err := svc.GkeHub.Projects.Locations.Scopes.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "fleet scope %s not found after deploy", name)
	}
	if scope.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("fleet scope %s missing the planton-ai_resource attribution label after deploy (labels: %v)", name, scope.Labels)
	}
	if id := outputs["scope_id"]; id != "" && !strings.HasSuffix(scope.Name, "/scopes/"+id) {
		return errors.Errorf("fleet scope %s does not end in the exported scope_id %q", scope.Name, id)
	}
	return nil
}

func (v *gkeFleetScopeVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	scope, err := svc.GkeHub.Projects.Locations.Scopes.Get(name).Context(ctx).Do()
	if err != nil {
		if isGkeHubNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing fleet scope %s after destroy", name)
	}
	if scope.DeleteTime != "" {
		return nil
	}
	return errors.Errorf("fleet scope %s still exists after destroy", name)
}

// gkeFleetMembershipVerifier probes a fleet membership by the name output
// (projects/{p}/locations/{l}/memberships/{id}). Posture assertions
// confirm the attribution labels landed.
type gkeFleetMembershipVerifier struct{}

func (v *gkeFleetMembershipVerifier) IDOutputKey() string { return "name" }

func (v *gkeFleetMembershipVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	membership, err := svc.GkeHub.Projects.Locations.Memberships.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "fleet membership %s not found after deploy", name)
	}
	if membership.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("fleet membership %s missing the planton-ai_resource attribution label after deploy (labels: %v)", name, membership.Labels)
	}
	return nil
}

func (v *gkeFleetMembershipVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	membership, err := svc.GkeHub.Projects.Locations.Memberships.Get(name).Context(ctx).Do()
	if err != nil {
		if isGkeHubNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing fleet membership %s after destroy", name)
	}
	if membership.DeleteTime != "" {
		return nil
	}
	return errors.Errorf("fleet membership %s still exists after destroy", name)
}
