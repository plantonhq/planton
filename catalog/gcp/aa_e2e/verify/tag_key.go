package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// tagKeyVerifier probes a Resource Manager tag key through the v3 API by its
// resource name (tagKeys/{id}) -- the name output. A deleted key is gone at
// once from the API's point of view (Google reserves the short name for 30
// days, but the key itself answers 404), so 404 is the destroyed shape.
type tagKeyVerifier struct{}

func (v *tagKeyVerifier) IDOutputKey() string { return "name" }

func (v *tagKeyVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	key, err := svc.CrmV3.TagKeys.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "tag key %s not found after deploy", name)
	}
	if ns := outputs["namespaced_name"]; ns != "" && key.NamespacedName != ns {
		return errors.Errorf("tag key %s resolved to namespaced name %q, want the namespaced_name output %q", name, key.NamespacedName, ns)
	}
	if id := outputs["tag_key_id"]; id != "" && key.Name != "tagKeys/"+id {
		return errors.Errorf("tag key %s resolved to name %q, want tagKeys/%s from the tag_key_id output", name, key.Name, id)
	}
	return nil
}

func (v *tagKeyVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, err := svc.CrmV3.TagKeys.Get(name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && (apiErr.Code == 404 || apiErr.Code == 403) {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing tag key %s after destroy", name)
	}
	return errors.Errorf("tag key %s still exists after destroy", name)
}
