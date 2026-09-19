package verify

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	orgpolicy "google.golang.org/api/orgpolicy/v2"
)

// orgPolicyVerifier probes an organization policy through the Organization
// Policy v2 API by its full name ({parent}/policies/{constraint}) -- the
// name output. The API exposes one Get per parent type, so the probe
// dispatches on the name's first segment. A deleted policy is gone at once
// (no soft-delete window): 404 is the only destroyed shape.
type orgPolicyVerifier struct{}

func (v *orgPolicyVerifier) IDOutputKey() string { return "name" }

func (v *orgPolicyVerifier) get(ctx context.Context, svc *Services, name string) (*orgpolicy.GoogleCloudOrgpolicyV2Policy, error) {
	switch {
	case strings.HasPrefix(name, "projects/"):
		return svc.OrgPolicy.Projects.Policies.Get(name).Context(ctx).Do()
	case strings.HasPrefix(name, "folders/"):
		return svc.OrgPolicy.Folders.Policies.Get(name).Context(ctx).Do()
	case strings.HasPrefix(name, "organizations/"):
		return svc.OrgPolicy.Organizations.Policies.Get(name).Context(ctx).Do()
	default:
		return nil, errors.Errorf("policy name %q does not start with projects/, folders/, or organizations/", name)
	}
}

func (v *orgPolicyVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	policy, err := v.get(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "organization policy %s not found after deploy", name)
	}
	if policy.Name != name {
		return errors.Errorf("organization policy resolved to name %q, want the name output %q", policy.Name, name)
	}
	if etag := outputs["etag"]; etag != "" && policy.Etag != "" && policy.Etag != etag {
		return errors.Errorf("organization policy %s etag is %q, want the etag output %q (an out-of-band edit landed between apply and verify)", name, policy.Etag, etag)
	}
	return nil
}

func (v *orgPolicyVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, err := v.get(ctx, svc, name)
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing organization policy %s after destroy", name)
	}
	return errors.Errorf("organization policy %s still exists after destroy", name)
}
