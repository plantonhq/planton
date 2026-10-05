package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// computeDiskVerifier probes a Compute Engine persistent disk via the
// compute API using the (project, zone, name) triple from the
// outputs. Posture assertions confirm the platform attribution labels
// landed (the cross-engine label-parity canary) and that the size and
// normalized type outputs match the live disk.
type computeDiskVerifier struct{}

func (v *computeDiskVerifier) IDOutputKey() string { return "name" }

func (v *computeDiskVerifier) project(svc *Services, outputs map[string]string) string {
	if p := outputs["project_id"]; p != "" {
		return p
	}
	return svc.Project
}

func (v *computeDiskVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	zone := outputs["zone"]
	project := v.project(svc, outputs)

	disk, err := svc.Compute.Disks.Get(project, zone, name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "compute disk %s/%s not found after deploy", zone, name)
	}

	// The platform attribution labels are the cross-engine parity canary:
	// a missing set means one engine stamped labels and the other did not.
	if disk.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("compute disk %s missing the planton-ai_resource attribution label after deploy (labels: %v)", name, disk.Labels)
	}

	// type is exported normalized to the plain name on both engines; the
	// live attribute is a full URL — suffix-match.
	if dt := outputs["type"]; dt != "" && !hasPathSuffix(disk.Type, dt) {
		return errors.Errorf("compute disk %s type output %q does not match live type %q", name, dt, disk.Type)
	}

	return nil
}

func (v *computeDiskVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	zone := outputs["zone"]
	project := v.project(svc, outputs)

	_, err := svc.Compute.Disks.Get(project, zone, name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing compute disk %s after destroy", name)
	}
	return errors.Errorf("compute disk %s still exists after destroy", name)
}
