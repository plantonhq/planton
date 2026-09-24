package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// vpcPeeringVerifier verifies a DigitalOceanVpcPeering via
// GET /v2/vpcs/peerings/{id}. Deletion is asynchronous on DigitalOcean's
// side (the peering passes through DELETING before vanishing), but the
// provider's destroy waits it out, so a 404 here is the settled absence
// signal.
type vpcPeeringVerifier struct{}

func (*vpcPeeringVerifier) IDOutputKey() string { return "peering_id" }

func (*vpcPeeringVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	exists, err := vpcPeeringExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceanvpcpeering verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceanvpcpeering %q not found after deploy", id)
	}
	return nil
}

func (*vpcPeeringVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	exists, err := vpcPeeringExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceanvpcpeering verify-absent failed for %q", id)
	}
	if exists {
		return &StillExistsError{Component: "digitaloceanvpcpeering", ID: id}
	}
	return nil
}

func vpcPeeringExists(ctx context.Context, client *godo.Client, id string) (bool, error) {
	_, _, err := client.VPCs.GetVPCPeering(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
