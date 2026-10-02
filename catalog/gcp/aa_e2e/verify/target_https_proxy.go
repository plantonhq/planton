package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

// getTargetHttpsProxy reads the proxy from the global or regional API
// collection by the region output, the same switch the modules make.
func getTargetHttpsProxy(ctx context.Context, svc *Services, outputs map[string]string) (*compute.TargetHttpsProxy, error) {
	name := outputs["proxy_name"]
	if region := outputs["region"]; region != "" {
		return svc.Compute.RegionTargetHttpsProxies.Get(svc.Project, region, name).Context(ctx).Do()
	}
	return svc.Compute.TargetHttpsProxies.Get(svc.Project, name).Context(ctx).Do()
}

// targetHttpsProxyVerifier probes a target HTTPS proxy by name -- the global
// or regional collection, chosen by the region output -- and confirms both
// wiring dimensions: the URL map (routing) and a TLS input (certificates, a
// certificate map, or a server TLS policy) — a TLS frontend without either
// is not serving.
type targetHttpsProxyVerifier struct{}

func (v *targetHttpsProxyVerifier) IDOutputKey() string { return "self_link" }

func (v *targetHttpsProxyVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["proxy_name"]
	proxy, err := getTargetHttpsProxy(ctx, svc, outputs)
	if err != nil {
		return errors.Wrapf(err, "target https proxy %s not found after deploy", name)
	}
	if proxy.UrlMap == "" {
		return errors.Errorf("target https proxy %s has no url map wired", name)
	}
	if len(proxy.SslCertificates) == 0 && proxy.CertificateMap == "" && proxy.ServerTlsPolicy == "" {
		return errors.Errorf("target https proxy %s has no TLS input (certificates, certificate map, or server TLS policy)", name)
	}
	return nil
}

func (v *targetHttpsProxyVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["proxy_name"]
	_, err := getTargetHttpsProxy(ctx, svc, outputs)
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing target https proxy %s after destroy", name)
	}
	return errors.Errorf("target https proxy %s still exists after destroy", name)
}
