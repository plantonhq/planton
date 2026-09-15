package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// loadBalancerVerifier verifies a DigitalOceanLoadBalancer via
// GET /v2/load_balancers/{id}. Beyond existence, it asserts the live
// balancer is active and checks the IPv4 address the module CLAIMS in its
// stack outputs against the live balancer -- outputs are contractually
// identical across both engines, so one assertion protects both, and an
// absent output simply means "not claimed" and is skipped. Status is always
// read live, never from an output: an apply-time snapshot goes stale
// immediately.
type loadBalancerVerifier struct{}

func (*loadBalancerVerifier) IDOutputKey() string { return "load_balancer_id" }

func (*loadBalancerVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	_, err := getLoadBalancer(ctx, client, id)
	return err
}

func (*loadBalancerVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	_, _, err := client.LoadBalancers.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return pkgerrors.Wrapf(err, "digitaloceanloadbalancer verify-absent failed for %q", id)
	}
	return &StillExistsError{Component: "digitaloceanloadbalancer", ID: id}
}

func (v *loadBalancerVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "load_balancer_id")
	if id == "" {
		return pkgerrors.New("load_balancer_id output missing after deploy")
	}

	lb, err := getLoadBalancer(ctx, client, id)
	if err != nil {
		return err
	}

	if lb.Status != "active" {
		return pkgerrors.Errorf("digitaloceanloadbalancer %q status is %q, want active", id, lb.Status)
	}

	if ip := StringOutput(outputs, "ip"); ip != "" && lb.IP != ip {
		return pkgerrors.Errorf("digitaloceanloadbalancer %q ip mismatch: output %q, live %q",
			id, ip, lb.IP)
	}

	return nil
}

func (v *loadBalancerVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, "load_balancer_id")
	if id == "" {
		return pkgerrors.New("load_balancer_id output missing for destroy verification")
	}
	return v.VerifyAbsent(ctx, client, id)
}

func getLoadBalancer(ctx context.Context, client *godo.Client, id string) (*godo.LoadBalancer, error) {
	lb, _, err := client.LoadBalancers.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, pkgerrors.Errorf("digitaloceanloadbalancer %q not found after deploy", id)
		}
		return nil, pkgerrors.Wrapf(err, "digitaloceanloadbalancer verify-exists failed for %q", id)
	}
	return lb, nil
}
