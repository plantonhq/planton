package verify

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// modelArmorTemplateGet reads a Model Armor template on the regional
// endpoint Google's provider uses
// (https://modelarmor.{location}.rep.googleapis.com/v1/).
func modelArmorTemplateGet(ctx context.Context, svc *Services, name string) (map[string]interface{}, int, error) {
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, "model armor template",
		fmt.Sprintf("https://modelarmor.%s.rep.googleapis.com/v1/%s", regionFromVertexResource(name), name), &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// modelArmorFloorSettingGet reads a floor setting on Model Armor's global
// endpoint (https://modelarmor.googleapis.com/v1/), where Google's provider
// manages floor settings.
func modelArmorFloorSettingGet(ctx context.Context, svc *Services, name string) (map[string]interface{}, int, error) {
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, "model armor floor setting",
		"https://modelarmor.googleapis.com/v1/"+name, &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// modelArmorTemplateVerifier probes a Model Armor template: it reads back
// under its exported name with the attribution labels and a filter
// configuration, and template_id is the last segment of that name.
type modelArmorTemplateVerifier struct{}

// IDOutputKey is the template's full resource name
// (projects/{p}/locations/{l}/templates/{id}).
func (v *modelArmorTemplateVerifier) IDOutputKey() string { return "name" }

func (v *modelArmorTemplateVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	template, _, err := modelArmorTemplateGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "model armor template %s not found after deploy", name)
	}
	labels, _ := template["labels"].(map[string]interface{})
	if labels["planton-ai_resource"] != "true" {
		return errors.Errorf("model armor template %s missing the planton-ai_resource attribution label after deploy", name)
	}
	if _, ok := template["filterConfig"].(map[string]interface{}); !ok {
		return errors.Errorf("model armor template %s has no filter configuration after deploy", name)
	}
	if got := outputs["template_id"]; got != lastPathSegment(name) {
		return errors.Errorf("model armor template %s template_id output %q does not match its name", name, got)
	}
	return nil
}

func (v *modelArmorTemplateVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, status, err := modelArmorTemplateGet(ctx, svc, name)
	return restAbsent("model armor template", name, status, err)
}

// modelArmorFloorSettingVerifier probes a Model Armor floor setting: it
// reads back under its exported name with a filter configuration, and the
// parent output is the name's prefix. Google cannot delete a floor
// setting, so the destroy verdict is the opposite of every other verifier:
// the setting must still read after destroy.
type modelArmorFloorSettingVerifier struct{}

// IDOutputKey is the floor setting's full resource name
// ({parent}/locations/{location}/floorSetting).
func (v *modelArmorFloorSettingVerifier) IDOutputKey() string { return "name" }

func (v *modelArmorFloorSettingVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	floor, _, err := modelArmorFloorSettingGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "model armor floor setting %s not found after deploy", name)
	}
	if _, ok := floor["filterConfig"].(map[string]interface{}); !ok {
		return errors.Errorf("model armor floor setting %s has no filter configuration after deploy", name)
	}
	if parent := outputs["parent"]; parent == "" || !strings.HasPrefix(name, parent+"/") {
		return errors.Errorf("model armor floor setting %s parent output %q is not its name's prefix", name, parent)
	}
	return nil
}

func (v *modelArmorFloorSettingVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	if _, _, err := modelArmorFloorSettingGet(ctx, svc, name); err != nil {
		return errors.Wrapf(err, "model armor floor setting %s no longer reads after destroy; Google keeps floor settings", name)
	}
	return nil
}
