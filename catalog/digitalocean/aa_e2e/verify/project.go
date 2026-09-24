package verify

import (
	"context"
	"strings"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// projectVerifier verifies a DigitalOceanProject via
// GET /v2/projects/{project_id} and, when the resource_urns output names
// members, GET /v2/projects/{project_id}/resources -- every claimed member
// must really sit in the project after deploy.
//
// Destroy asserts the kind's headline promise, not just absence: the
// provider relocates every member to the account's DEFAULT project before
// deleting (DigitalOcean refuses to delete a non-empty project), so after
// destroy the project must be gone AND every member URN must still exist,
// now under the default project. The dependency fixtures that supplied
// those members are torn down only after this phase, so they are there to
// be found. A member that vanished would mean destroy destroyed something
// it promised never to touch; a member left in limbo would fail the same
// probe.
type projectVerifier struct{}

func (*projectVerifier) IDOutputKey() string { return "project_id" }

func (*projectVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceanproject requires the full outputs map (project_id + resource_urns); " +
		"the harness dispatches through VerifyExistsFromOutputs")
}

func (*projectVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceanproject requires the full outputs map (project_id + resource_urns); " +
		"the harness dispatches through VerifyAbsentFromOutputs")
}

func (v *projectVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "project_id")
	if id == "" {
		return pkgerrors.New("digitaloceanproject outputs carry no project_id")
	}
	project, _, err := client.Projects.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return pkgerrors.Errorf("digitaloceanproject %q not found after deploy", id)
		}
		return pkgerrors.Wrap(err, "digitaloceanproject verify-exists failed")
	}
	if project.ID == "" {
		return pkgerrors.Errorf("digitaloceanproject %q returned an empty project", id)
	}
	claimed := StringSliceOutput(outputs, "resource_urns")
	if len(claimed) == 0 {
		return nil
	}
	members, err := projectMemberURNs(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceanproject verify-exists: listing members failed")
	}
	if missing := missingURNs(claimed, members); len(missing) > 0 {
		return pkgerrors.Errorf("digitaloceanproject %q: claimed members %v are not in the project (live members: %v)", id, missing, members)
	}
	return nil
}

func (v *projectVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "project_id")
	if id == "" {
		return pkgerrors.New("digitaloceanproject outputs carry no project_id")
	}
	_, _, err := client.Projects.Get(ctx, id)
	if err == nil {
		return &StillExistsError{Component: "digitaloceanproject", ID: id}
	}
	if !isNotFound(err) {
		return pkgerrors.Wrap(err, "digitaloceanproject verify-absent failed")
	}
	claimed := StringSliceOutput(outputs, "resource_urns")
	if len(claimed) == 0 {
		return nil
	}
	// The project is gone; its former members must now live in the
	// account's default project. Relocation is what let the delete succeed
	// at all, so by this point the moves have settled -- a miss here is a
	// real defect, never a race to poll through.
	defaultProject, _, err := client.Projects.GetDefault(ctx)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceanproject verify-absent: reading the default project failed")
	}
	members, err := projectMemberURNs(ctx, client, defaultProject.ID)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceanproject verify-absent: listing the default project's members failed")
	}
	if missing := missingURNs(claimed, members); len(missing) > 0 {
		return pkgerrors.Errorf("digitaloceanproject %q destroyed, but former members %v were not relocated to the default project %q (its members: %v)", id, missing, defaultProject.Name, members)
	}
	return nil
}

// projectMemberURNs lists every member URN of a project, following the
// API's pagination.
func projectMemberURNs(ctx context.Context, client *godo.Client, projectID string) ([]string, error) {
	var urns []string
	opts := &godo.ListOptions{PerPage: 200}
	for {
		resources, resp, err := client.Projects.ListResources(ctx, projectID, opts)
		if err != nil {
			return nil, err
		}
		for _, r := range resources {
			urns = append(urns, r.URN)
		}
		if resp == nil || resp.Links == nil || resp.Links.IsLastPage() {
			return urns, nil
		}
		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}
		opts.Page = page + 1
	}
}

// missingURNs returns the claimed URNs absent from the live list. URNs are
// compared case-insensitively: DigitalOcean canonicalizes the type segment
// and a mismatch there is not a membership difference.
func missingURNs(claimed, live []string) []string {
	present := make(map[string]bool, len(live))
	for _, urn := range live {
		present[strings.ToLower(urn)] = true
	}
	var missing []string
	for _, urn := range claimed {
		if !present[strings.ToLower(urn)] {
			missing = append(missing, urn)
		}
	}
	return missing
}
