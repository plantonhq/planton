package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// kmsRestGet reads one Cloud KMS Autokey resource (an autokeyConfig or a
// keyHandle) by its full name from https://cloudkms.googleapis.com/v1/ --
// the Autokey surfaces the harness's typed KMS client does not cover.
func kmsRestGet(ctx context.Context, svc *Services, what, name string) (map[string]interface{}, int, error) {
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, what, fmt.Sprintf("https://cloudkms.googleapis.com/v1/%s", name), &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// autokeyModeUnset reports whether an Autokey configuration carries no
// resolution mode -- what Google reads back after the configuration is
// cleared.
func autokeyModeUnset(config map[string]interface{}) bool {
	mode := stringField(config, "keyProjectResolutionMode")
	return mode == "" || mode == "KEY_PROJECT_RESOLUTION_MODE_UNSPECIFIED"
}

// kmsAutokeyConfigVerifier probes an Autokey configuration. Google keeps
// one per folder and project, so destroy (deletion_policy DELETE) clears it
// rather than deleting it: the verdict is that it reads back with no
// resolution mode.
type kmsAutokeyConfigVerifier struct{}

// IDOutputKey is the configuration's full name
// (folders/{id}/autokeyConfig or projects/{id}/autokeyConfig).
func (v *kmsAutokeyConfigVerifier) IDOutputKey() string { return "name" }

func (v *kmsAutokeyConfigVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	config, _, err := kmsRestGet(ctx, svc, "autokey config", name)
	if err != nil {
		return errors.Wrapf(err, "autokey config %s not readable after deploy", name)
	}
	if autokeyModeUnset(config) {
		return errors.Errorf("autokey config %s has no key project resolution mode after deploy", name)
	}
	if parent := outputs["parent"]; parent == "" || name != parent+"/autokeyConfig" {
		return errors.Errorf("autokey config %s parent output %q is not its name's scope", name, parent)
	}
	return nil
}

func (v *kmsAutokeyConfigVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	config, _, err := kmsRestGet(ctx, svc, "autokey config", name)
	if err != nil {
		return errors.Wrapf(err, "autokey config %s not readable after destroy", name)
	}
	if !autokeyModeUnset(config) {
		return errors.Errorf("autokey config %s still has resolution mode %q after destroy; destroy should clear it",
			name, stringField(config, "keyProjectResolutionMode"))
	}
	return nil
}

// kmsKeyHandleVerifier probes a key handle: it reads back with the key
// Autokey assigned equal to the kms_key output, and that key exists.
// Google keeps key handles -- destroy only removes them from state -- so
// the destroy verdict is that the handle still reads.
type kmsKeyHandleVerifier struct{}

// IDOutputKey is the handle's full name
// (projects/{p}/locations/{l}/keyHandles/{name}).
func (v *kmsKeyHandleVerifier) IDOutputKey() string { return "name" }

func (v *kmsKeyHandleVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	handle, _, err := kmsRestGet(ctx, svc, "key handle", name)
	if err != nil {
		return errors.Wrapf(err, "key handle %s not found after deploy", name)
	}
	key := stringField(handle, "kmsKey")
	if key == "" || key != outputs["kms_key"] {
		return errors.Errorf("key handle %s holds key %q, but the kms_key output is %q", name, key, outputs["kms_key"])
	}
	if _, err := svc.CloudKms.Projects.Locations.KeyRings.CryptoKeys.Get(key).Context(ctx).Do(); err != nil {
		return errors.Wrapf(err, "key %s assigned to key handle %s not found", key, name)
	}
	return nil
}

func (v *kmsKeyHandleVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	if _, _, err := kmsRestGet(ctx, svc, "key handle", name); err != nil {
		return errors.Wrapf(err, "key handle %s no longer reads after destroy; Google keeps key handles", name)
	}
	return nil
}
