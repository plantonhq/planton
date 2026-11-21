package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// subnetworkVerifier probes a subnetwork by name and region via the compute
// API, confirming the primary range matches what the outputs claim.
type subnetworkVerifier struct{}

func (v *subnetworkVerifier) IDOutputKey() string { return "subnetwork_self_link" }

func (v *subnetworkVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["subnetwork_name"]
	region := outputs["region"]
	subnetwork, err := svc.Compute.Subnetworks.Get(projectInSelfLink(outputs["subnetwork_self_link"], svc.Project), region, name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "subnetwork %s/%s not found after deploy", region, name)
	}
	if want := outputs["ip_cidr_range"]; want != "" && subnetwork.IpCidrRange != want {
		return errors.Errorf("subnetwork %s has primary range %q, expected %q",
			name, subnetwork.IpCidrRange, want)
	}
	return nil
}

func (v *subnetworkVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["subnetwork_name"]
	region := outputs["region"]
	_, err := svc.Compute.Subnetworks.Get(projectInSelfLink(outputs["subnetwork_self_link"], svc.Project), region, name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing subnetwork %s after destroy", name)
	}
	return errors.Errorf("subnetwork %s still exists after destroy", name)
}
