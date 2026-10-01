package verify

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// computeImageVerifier probes a Compute Engine custom image via the
// compute API, by the project in its self_link and the name output.
// Posture assertions confirm the platform attribution labels landed (the
// cross-engine label-parity canary) and that the family output matches
// the live image.
type computeImageVerifier struct{}

func (v *computeImageVerifier) IDOutputKey() string { return "name" }

// project reads the project from the image's self link
// (.../projects/{project}/global/images/{name}), falling back to the
// harness project.
func (v *computeImageVerifier) project(svc *Services, outputs map[string]string) string {
	parts := strings.Split(outputs["self_link"], "/")
	for i, part := range parts {
		if part == "projects" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return svc.Project
}

func (v *computeImageVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	image, err := svc.Compute.Images.Get(v.project(svc, outputs), name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "compute image %s not found after deploy", name)
	}
	if image.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("compute image %s missing the planton-ai_resource attribution label after deploy (labels: %v)", name, image.Labels)
	}
	if image.Family != outputs["family"] {
		return errors.Errorf("compute image %s family output %q does not match live family %q", name, outputs["family"], image.Family)
	}
	return nil
}

func (v *computeImageVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, err := svc.Compute.Images.Get(v.project(svc, outputs), name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing compute image %s after destroy", name)
	}
	return errors.Errorf("compute image %s still exists after destroy", name)
}
