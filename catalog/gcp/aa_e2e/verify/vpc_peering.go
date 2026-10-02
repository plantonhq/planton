package verify

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

// vpcPeeringVerifier probes this side's peering entry on the network object
// (the API has no standalone peering GET) by the peering_name output, and
// -- because a peering goes ACTIVE only once the other side exists, which
// the fixture chain deploys before the instance under test -- polls the
// entry for ACTIVE for a bounded time. The routes-config form exports an
// empty state and is not asserted on state (it does not own the entry).
type vpcPeeringVerifier struct{}

func (v *vpcPeeringVerifier) IDOutputKey() string { return "peering_name" }

// peeringEntry returns the network's peering entry with the given name, or
// nil when the network has none by that name.
func peeringEntry(ctx context.Context, svc *Services, networkSelfLink, peeringName string) (*compute.NetworkPeering, error) {
	networkName := networkNameFromSelfLink(networkSelfLink)
	network, err := svc.Compute.Networks.Get(svc.Project, networkName).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	for _, peering := range network.Peerings {
		if peering != nil && peering.Name == peeringName {
			return peering, nil
		}
	}
	return nil, nil
}

func (v *vpcPeeringVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	peeringName, networkSelfLink := outputs["peering_name"], outputs["network"]
	if peeringName == "" || networkSelfLink == "" {
		return errors.New("peering_name / network outputs missing after deploy")
	}

	// Poll for ACTIVE: the pair is up once both sides exist, which can lag
	// the second side's create by a few seconds.
	deadline := time.Now().Add(2 * time.Minute)
	for {
		peering, err := peeringEntry(ctx, svc, networkSelfLink, peeringName)
		if err != nil {
			return errors.Wrapf(err, "network %s not readable after deploy", networkNameFromSelfLink(networkSelfLink))
		}
		if peering == nil {
			return errors.Errorf("network %s has no peering named %s after deploy", networkNameFromSelfLink(networkSelfLink), peeringName)
		}
		// The routes-config form exports an empty state (it does not own the
		// entry); only the create form asserts the pair's state.
		if outputs["state"] == "" {
			return nil
		}
		if peering.State == "ACTIVE" {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.Errorf("peering %s on %s is %s after deploy (%s), want ACTIVE",
				peeringName, networkNameFromSelfLink(networkSelfLink), peering.State, peering.StateDetails)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
}

func (v *vpcPeeringVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	peeringName, networkSelfLink := outputs["peering_name"], outputs["network"]
	if peeringName == "" || networkSelfLink == "" {
		return nil
	}
	// The routes-config form's destroy is a no-op in GCP: the peering it
	// tuned still exists by design, so nothing is asserted absent.
	if outputs["state"] == "" {
		return nil
	}
	peering, err := peeringEntry(ctx, svc, networkSelfLink, peeringName)
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && (apiErr.Code == 404 || apiErr.Code == 403) {
			// The network itself is gone (torn down after the peering).
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing network %s after destroy", networkNameFromSelfLink(networkSelfLink))
	}
	if peering != nil {
		return errors.Errorf("peering %s still exists on %s after destroy", peeringName, networkNameFromSelfLink(networkSelfLink))
	}
	return nil
}
