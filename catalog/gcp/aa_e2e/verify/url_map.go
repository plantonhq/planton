package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

// getUrlMap reads the map from the global or regional API collection by the
// region output, the same switch the modules make.
func getUrlMap(ctx context.Context, svc *Services, outputs map[string]string) (*compute.UrlMap, error) {
	name := outputs["url_map_name"]
	if region := outputs["region"]; region != "" {
		return svc.Compute.RegionUrlMaps.Get(svc.Project, region, name).Context(ctx).Do()
	}
	return svc.Compute.UrlMaps.Get(svc.Project, name).Context(ctx).Do()
}

// urlMapVerifier probes a URL map by name -- the global or regional
// collection, chosen by the region output -- and confirms default-service
// wiring when the scenario sets one.
type urlMapVerifier struct{}

func (v *urlMapVerifier) IDOutputKey() string { return "self_link" }

func (v *urlMapVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["url_map_name"]
	urlMap, err := getUrlMap(ctx, svc, outputs)
	if err != nil {
		return errors.Wrapf(err, "url map %s not found after deploy", name)
	}
	if urlMap.DefaultService == "" && urlMap.DefaultUrlRedirect == nil && urlMap.DefaultRouteAction == nil {
		return errors.Errorf("url map %s has no default target configured", name)
	}
	return nil
}

func (v *urlMapVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["url_map_name"]
	_, err := getUrlMap(ctx, svc, outputs)
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing url map %s after destroy", name)
	}
	return errors.Errorf("url map %s still exists after destroy", name)
}
