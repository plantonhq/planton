package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

// getTargetHttpProxy reads the proxy from the global or regional API
// collection by the region output, the same switch the modules make.
func getTargetHttpProxy(ctx context.Context, svc *Services, outputs map[string]string) (*compute.TargetHttpProxy, error) {
	name := outputs["proxy_name"]
	if region := outputs["region"]; region != "" {
		return svc.Compute.RegionTargetHttpProxies.Get(svc.Project, region, name).Context(ctx).Do()
	}
	return svc.Compute.TargetHttpProxies.Get(svc.Project, name).Context(ctx).Do()
}

// targetHttpProxyVerifier probes a target HTTP proxy by name -- the global or
// regional collection, chosen by the region output -- and confirms the
// URL-map wiring — a proxy without a routing table is not a working
// frontend.
type targetHttpProxyVerifier struct{}

func (v *targetHttpProxyVerifier) IDOutputKey() string { return "self_link" }

func (v *targetHttpProxyVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["proxy_name"]
	proxy, err := getTargetHttpProxy(ctx, svc, outputs)
	if err != nil {
		return errors.Wrapf(err, "target http proxy %s not found after deploy", name)
	}
	if proxy.UrlMap == "" {
		return errors.Errorf("target http proxy %s has no url map wired", name)
	}
	return nil
}

func (v *targetHttpProxyVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["proxy_name"]
	_, err := getTargetHttpProxy(ctx, svc, outputs)
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing target http proxy %s after destroy", name)
	}
	return errors.Errorf("target http proxy %s still exists after destroy", name)
}
