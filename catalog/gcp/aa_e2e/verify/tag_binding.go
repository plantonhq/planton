package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// tagBindingVerifier probes a Resource Manager tag binding through the v3
// API. The Tag Bindings API has no Get: bindings are LISTED by the full
// resource name of the tagged parent, so the probe lists the parent
// output's bindings and looks for the tag_value output among them. The
// name output is the handle Google deletes by; it is checked against the
// listed binding when present. Location-scoped bindings (a regional or
// zonal parent) are served from the regional endpoint, which this global
// client cannot reach -- those arms are offline-only until the proof lane
// has a regional fixture, and the profile says so.
type tagBindingVerifier struct{}

func (v *tagBindingVerifier) IDOutputKey() string { return "name" }

// find returns the name of the listed binding of tagValue on parent, and
// whether one was found.
func (v *tagBindingVerifier) find(ctx context.Context, svc *Services, parent, tagValue string) (string, bool, error) {
	call := svc.CrmV3.TagBindings.List().Parent(parent).PageSize(300).Context(ctx)
	for {
		resp, err := call.Do()
		if err != nil {
			return "", false, err
		}
		for _, b := range resp.TagBindings {
			if b.TagValue == tagValue {
				return b.Name, true, nil
			}
		}
		if resp.NextPageToken == "" {
			return "", false, nil
		}
		call = call.PageToken(resp.NextPageToken)
	}
}

func (v *tagBindingVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	parent, tagValue := outputs["parent"], outputs["tag_value"]
	if parent == "" || tagValue == "" {
		return errors.New("parent or tag_value output missing after deploy")
	}
	name, found, err := v.find(ctx, svc, parent, tagValue)
	if err != nil {
		return errors.Wrapf(err, "failed to list tag bindings on %s after deploy", parent)
	}
	if !found {
		return errors.Errorf("tag value %s is not bound to %s after deploy", tagValue, parent)
	}
	if want := outputs["name"]; want != "" && name != want {
		return errors.Errorf("tag binding on %s resolved to name %q, want the name output %q", parent, name, want)
	}
	return nil
}

func (v *tagBindingVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	parent, tagValue := outputs["parent"], outputs["tag_value"]
	if parent == "" || tagValue == "" {
		return nil
	}
	_, found, err := v.find(ctx, svc, parent, tagValue)
	if err != nil {
		// A parent that has itself been destroyed (a folder fixture torn
		// down first) answers 404/403; its bindings are gone with it.
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && (apiErr.Code == 404 || apiErr.Code == 403) {
			return nil
		}
		return errors.Wrapf(err, "unexpected error listing tag bindings on %s after destroy", parent)
	}
	if found {
		return errors.Errorf("tag value %s is still bound to %s after destroy", tagValue, parent)
	}
	return nil
}
