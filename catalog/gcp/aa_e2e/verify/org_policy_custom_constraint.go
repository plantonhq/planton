package verify

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// orgPolicyCustomConstraintVerifier probes a custom constraint through the
// Organization Policy v2 API by its full name
// (organizations/{org}/customConstraints/custom.{name}) -- the name output.
// A deleted constraint is gone at once: 404 is the only destroyed shape.
type orgPolicyCustomConstraintVerifier struct{}

func (v *orgPolicyCustomConstraintVerifier) IDOutputKey() string { return "name" }

func (v *orgPolicyCustomConstraintVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	constraint, err := svc.OrgPolicy.Organizations.CustomConstraints.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "custom constraint %s not found after deploy", name)
	}
	// The constraint output is the custom.{name} handle a policy enforces;
	// it must be the last segment of the full name Google returns.
	if handle := outputs["constraint"]; handle != "" && !strings.HasSuffix(constraint.Name, "/"+handle) {
		return errors.Errorf("custom constraint %s resolved to name %q, want it to end in the constraint output %q", name, constraint.Name, handle)
	}
	return nil
}

func (v *orgPolicyCustomConstraintVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, err := svc.OrgPolicy.Organizations.CustomConstraints.Get(name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing custom constraint %s after destroy", name)
	}
	return errors.Errorf("custom constraint %s still exists after destroy", name)
}
