package verify

import (
	"context"

	"github.com/pkg/errors"
)

// dialogflowCxSecuritySettingsVerifier probes Dialogflow CX security
// settings through the REST API: the settings read back under their name
// with the exported id.
type dialogflowCxSecuritySettingsVerifier struct{}

// IDOutputKey is the settings' full resource name
// (projects/{p}/locations/{l}/securitySettings/{id}).
func (v *dialogflowCxSecuritySettingsVerifier) IDOutputKey() string { return "name" }

func (v *dialogflowCxSecuritySettingsVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	settings, _, err := dialogflowGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "dialogflow cx security settings %s not found after deploy", name)
	}
	if got := outputs["security_settings_id"]; got != lastPathSegment(stringField(settings, "name")) {
		return errors.Errorf("dialogflow cx security settings %s security_settings_id output %q does not match the live id %q", name, got, lastPathSegment(stringField(settings, "name")))
	}
	return nil
}

// VerifyAbsent confirms the settings are gone.
func (v *dialogflowCxSecuritySettingsVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, status, err := dialogflowGet(ctx, svc, name)
	return restAbsent("dialogflow cx security settings", name, status, err)
}
