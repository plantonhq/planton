package verify

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

// networkEndpointGroupVerifier probes a network endpoint group through the
// compute API, reading the zonal or global collection by the zone output
// (empty = global), the same switch the modules make. Posture assertions
// confirm the exported self_link matches the live resource -- the value a
// backend service names as its group -- and that Google's member count
// matches the size the manifest declared, which proves the endpoint set
// was written.
type networkEndpointGroupVerifier struct{}

func (v *networkEndpointGroupVerifier) IDOutputKey() string { return "neg_name" }

// getNetworkEndpointGroup reads the group from whichever collection the
// zone output selects.
func getNetworkEndpointGroup(ctx context.Context, svc *Services, outputs map[string]string, name string) (*compute.NetworkEndpointGroup, error) {
	if zone := outputs["zone"]; zone != "" {
		return svc.Compute.NetworkEndpointGroups.Get(svc.Project, zone, name).Context(ctx).Do()
	}
	return svc.Compute.GlobalNetworkEndpointGroups.Get(svc.Project, name).Context(ctx).Do()
}

func (v *networkEndpointGroupVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["neg_name"]
	if name == "" {
		return errors.New("neg_name output missing after deploy")
	}

	group, err := getNetworkEndpointGroup(ctx, svc, outputs, name)
	if err != nil {
		return errors.Wrapf(err, "network endpoint group %s (zone %q) not found after deploy", name, outputs["zone"])
	}

	if wantSelfLink := outputs["self_link"]; wantSelfLink != "" && group.SelfLink != wantSelfLink {
		return errors.Errorf("network endpoint group %s self_link mismatch: output %q, live %q", name, wantSelfLink, group.SelfLink)
	}

	// The declared size is the whole membership; Google's count must agree
	// or the endpoint set did not land.
	if wantSize := outputs["size"]; wantSize != "" {
		want, convErr := strconv.ParseInt(wantSize, 10, 64)
		if convErr != nil {
			return errors.Wrapf(convErr, "network endpoint group %s size output %q is not a number", name, wantSize)
		}
		if group.Size != want {
			return errors.Errorf("network endpoint group %s size mismatch: declared %d, live %d", name, want, group.Size)
		}
	}
	return nil
}

func (v *networkEndpointGroupVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["neg_name"]
	if name == "" {
		return nil
	}

	_, err := getNetworkEndpointGroup(ctx, svc, outputs, name)
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing network endpoint group %s after destroy", name)
	}
	return errors.Errorf("network endpoint group %s still exists after destroy", name)
}
