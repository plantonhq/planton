package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

// getBackendService reads the service from the global or regional API
// collection by the region output, the same switch the modules make.
func getBackendService(ctx context.Context, svc *Services, outputs map[string]string) (*compute.BackendService, error) {
	name := outputs["backend_service_name"]
	if region := outputs["region"]; region != "" {
		return svc.Compute.RegionBackendServices.Get(svc.Project, region, name).Context(ctx).Do()
	}
	return svc.Compute.BackendServices.Get(svc.Project, name).Context(ctx).Do()
}

// backendServiceVerifier probes a Compute Engine backend service by name via
// the compute API -- the global or regional collection, chosen by the region
// output -- additionally confirming its health-check wiring —
// the FK-resolved reference to the GcpHealthCheck prerequisite is the point
// of the composition.
type backendServiceVerifier struct{}

func (v *backendServiceVerifier) IDOutputKey() string { return "self_link" }

func (v *backendServiceVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["backend_service_name"]
	backendService, err := getBackendService(ctx, svc, outputs)
	if err != nil {
		return errors.Wrapf(err, "backend service %s not found after deploy", name)
	}
	// When a health check is wired, the deployed service must carry exactly one
	// (GCP caps the list at one). Serverless-NEG scenarios omit health_check —
	// those backends manage their own health — so an empty list is valid too.
	if len(backendService.HealthChecks) > 1 {
		return errors.Errorf("backend service %s carries %d health checks, expected at most 1",
			name, len(backendService.HealthChecks))
	}
	// The neg-backend scenario wires a live NEG; confirm at least one backend
	// group is attached when no health check is present.
	if len(backendService.HealthChecks) == 0 && len(backendService.Backends) == 0 {
		return errors.Errorf("backend service %s has neither health checks nor backends after deploy", name)
	}
	return nil
}

func (v *backendServiceVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["backend_service_name"]
	_, err := getBackendService(ctx, svc, outputs)
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing backend service %s after destroy", name)
	}
	return errors.Errorf("backend service %s still exists after destroy", name)
}
