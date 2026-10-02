package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// tagValueVerifier probes a Resource Manager tag value through the v3 API by
// its resource name (tagValues/{id}) -- the name output. 404 is the
// destroyed shape (the short name stays reserved under the key for 30 days,
// but the value itself is gone).
type tagValueVerifier struct{}

func (v *tagValueVerifier) IDOutputKey() string { return "name" }

func (v *tagValueVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	value, err := svc.CrmV3.TagValues.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "tag value %s not found after deploy", name)
	}
	if ns := outputs["namespaced_name"]; ns != "" && value.NamespacedName != ns {
		return errors.Errorf("tag value %s resolved to namespaced name %q, want the namespaced_name output %q", name, value.NamespacedName, ns)
	}
	if id := outputs["tag_value_id"]; id != "" && value.Name != "tagValues/"+id {
		return errors.Errorf("tag value %s resolved to name %q, want tagValues/%s from the tag_value_id output", name, value.Name, id)
	}
	return nil
}

func (v *tagValueVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, err := svc.CrmV3.TagValues.Get(name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && (apiErr.Code == 404 || apiErr.Code == 403) {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing tag value %s after destroy", name)
	}
	return errors.Errorf("tag value %s still exists after destroy", name)
}
