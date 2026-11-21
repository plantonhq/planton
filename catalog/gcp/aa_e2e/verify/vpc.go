package verify

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// vpcVerifier probes a VPC network by name via the compute API. Registered so
// the VPC can serve as a verified E2E prerequisite for subnet-scoped kinds.
type vpcVerifier struct{}

func (v *vpcVerifier) IDOutputKey() string { return "network_self_link" }

func (v *vpcVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["network_name"]
	net, err := svc.Compute.Networks.Get(projectInSelfLink(outputs["network_self_link"], svc.Project), name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "vpc network %s not found after deploy", name)
	}
	// Posture assertion on the depth surface: the self-link output must match
	// the live network. (No MTU assertion: networks created without an
	// explicit mtu can be returned by the API without one immediately after
	// create, so a non-zero check fails on freshly deployed prerequisites.)
	if selfLink := outputs["network_self_link"]; selfLink != "" && net.SelfLink != selfLink {
		return errors.Errorf("vpc network %s self-link mismatch: output %q, live %q", name, selfLink, net.SelfLink)
	}
	return nil
}

func (v *vpcVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["network_name"]
	_, err := svc.Compute.Networks.Get(projectInSelfLink(outputs["network_self_link"], svc.Project), name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing vpc network %s after destroy", name)
	}
	return errors.Errorf("vpc network %s still exists after destroy", name)
}

// projectInSelfLink returns the project a compute self link names
// (https://www.googleapis.com/compute/v1/projects/P/...), or fallback when the
// link names none: a fixture may deploy into another test project than the
// harness's own (a fleet chain in the second project).
func projectInSelfLink(link, fallback string) string {
	const marker = "projects/"
	i := strings.Index("/"+link, "/"+marker)
	if i < 0 {
		return fallback
	}
	rest := link[i+len(marker):]
	if j := strings.Index(rest, "/"); j > 0 {
		return rest[:j]
	}
	return fallback
}
