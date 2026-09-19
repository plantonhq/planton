package verify

import (
	"context"

	"github.com/pkg/errors"
)

// sharedVpcServiceProjectVerifier probes the attachment through the HOST's
// xpnResources list (the API has no per-attachment GET): the service project
// appears there by ID (or number) while attached and disappears on detach.
type sharedVpcServiceProjectVerifier struct{}

func (v *sharedVpcServiceProjectVerifier) IDOutputKey() string { return "service_project_id" }

// attached reports whether the service project is listed among the host's
// service resources. Google returns each as a PROJECT resource id that may
// be the project id or its number; the outputs carry the id, which is what
// the module sent.
func (v *sharedVpcServiceProjectVerifier) attached(ctx context.Context, svc *Services, host, service string) (bool, error) {
	resources, err := svc.Compute.Projects.GetXpnResources(host).Context(ctx).Do()
	if err != nil {
		return false, errors.Wrapf(err, "cannot list service projects of host %s", host)
	}
	for _, resource := range resources.Resources {
		if resource != nil && resource.Type == "PROJECT" && resource.Id == service {
			return true, nil
		}
	}
	return false, nil
}

func (v *sharedVpcServiceProjectVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	host, service := outputs["host_project_id"], outputs["service_project_id"]
	if host == "" || service == "" {
		return errors.New("host_project_id / service_project_id outputs missing after deploy")
	}
	ok, err := v.attached(ctx, svc, host, service)
	if err != nil {
		return errors.Wrap(err, "after deploy")
	}
	if !ok {
		return errors.Errorf("service project %s is not attached to host %s after deploy", service, host)
	}
	return nil
}

func (v *sharedVpcServiceProjectVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	host, service := outputs["host_project_id"], outputs["service_project_id"]
	if host == "" || service == "" {
		return nil
	}
	ok, err := v.attached(ctx, svc, host, service)
	if err != nil {
		// The host may itself have been torn down after the attachment; a
		// host the identity can no longer read has no attachments to report.
		return nil
	}
	if ok {
		return errors.Errorf("service project %s is still attached to host %s after destroy", service, host)
	}
	return nil
}
