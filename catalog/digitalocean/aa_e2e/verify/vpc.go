package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// vpcVerifier verifies a DigitalOceanVpc via GET /v2/vpcs/{id}. A deleted VPC
// returns 404, which is the "absent" signal; any other error surfaces.
type vpcVerifier struct{}

func (*vpcVerifier) IDOutputKey() string { return "vpc_id" }

func (*vpcVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	exists, err := vpcExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceanvpc verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceanvpc %q not found after deploy", id)
	}
	return nil
}

func (*vpcVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	exists, err := vpcExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceanvpc verify-absent failed for %q", id)
	}
	if exists {
		return &StillExistsError{Component: "digitaloceanvpc", ID: id}
	}
	return nil
}

func vpcExists(ctx context.Context, client *godo.Client, id string) (bool, error) {
	_, _, err := client.VPCs.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
