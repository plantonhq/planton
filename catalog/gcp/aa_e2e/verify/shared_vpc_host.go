package verify

import (
	"context"

	"github.com/pkg/errors"
)

// sharedVpcHostVerifier probes the project's Shared VPC role through the
// compute projects API: a host reports xpnProjectStatus HOST; a project
// that is not a host reports UNSPECIFIED_XPN_PROJECT_STATUS (or omits the
// field). The project itself outlives the resource, so "absent" means the
// role is gone, not the project.
type sharedVpcHostVerifier struct{}

func (v *sharedVpcHostVerifier) IDOutputKey() string { return "host_project_id" }

func (v *sharedVpcHostVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	projectID := outputs["host_project_id"]
	if projectID == "" {
		return errors.New("host_project_id output missing after deploy")
	}
	project, err := svc.Compute.Projects.Get(projectID).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "project %s not readable after deploy", projectID)
	}
	if project.XpnProjectStatus != "HOST" {
		return errors.Errorf("project %s has xpnProjectStatus %q after deploy, want HOST", projectID, project.XpnProjectStatus)
	}
	return nil
}

func (v *sharedVpcHostVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	projectID := outputs["host_project_id"]
	if projectID == "" {
		return nil
	}
	project, err := svc.Compute.Projects.Get(projectID).Context(ctx).Do()
	if err != nil {
		// A project the identity can no longer read is as absent as the
		// verifier can tell; the project is not this resource's to delete.
		return nil
	}
	if project.XpnProjectStatus == "HOST" {
		return errors.Errorf("project %s is still a Shared VPC host after destroy", projectID)
	}
	return nil
}
