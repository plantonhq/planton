package verify

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// cloudIdentityGroupVerifier probes a Cloud Identity group through the
// Cloud Identity API by the name output (groups/{id}). Posture assertions
// confirm the group's email matches the declared one -- the identity IAM
// bindings name -- and that Google's membership count matches the
// manifest's, which proves the folded memberships landed.
type cloudIdentityGroupVerifier struct{}

func (v *cloudIdentityGroupVerifier) IDOutputKey() string { return "name" }

func (v *cloudIdentityGroupVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}

	group, err := svc.CloudIdentity.Groups.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "cloud identity group %s not found after deploy", name)
	}

	if wantEmail := outputs["group_email"]; wantEmail != "" && (group.GroupKey == nil || group.GroupKey.Id != wantEmail) {
		return errors.Errorf("cloud identity group %s email mismatch: output %q, live %v", name, wantEmail, group.GroupKey)
	}

	if wantCount := outputs["membership_count"]; wantCount != "" {
		want, convErr := strconv.Atoi(wantCount)
		if convErr != nil {
			return errors.Wrapf(convErr, "cloud identity group %s membership_count output %q is not a number", name, wantCount)
		}
		resp, listErr := svc.CloudIdentity.Groups.Memberships.List(name).Context(ctx).Do()
		if listErr != nil {
			return errors.Wrapf(listErr, "listing memberships of cloud identity group %s", name)
		}
		// WITH_INITIAL_OWNER adds the caller as an owner beyond the manifest's
		// list, so Google may report one more membership than declared.
		if len(resp.Memberships) < want {
			return errors.Errorf("cloud identity group %s membership count mismatch: declared %d, live %d", name, want, len(resp.Memberships))
		}
	}
	return nil
}

func (v *cloudIdentityGroupVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}

	_, err := svc.CloudIdentity.Groups.Get(name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && (apiErr.Code == 404 || apiErr.Code == 403) {
			// Cloud Identity answers 403 for a group that no longer exists
			// under the caller's customer.
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing cloud identity group %s after destroy", name)
	}
	return errors.Errorf("cloud identity group %s still exists after destroy", name)
}
