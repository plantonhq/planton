package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// folderVerifier probes a Resource Manager folder through the v3 API by its
// resource name (folders/{id}) -- the name output is exactly the handle
// the API addresses the folder by, so verifying with it doubles as proof the
// output is honest. Destroyed folders enter a 30-day soft-delete window
// rather than vanishing, so the absence contract accepts DELETE_REQUESTED as
// destroyed -- the same posture as projects.
type folderVerifier struct{}

func (v *folderVerifier) IDOutputKey() string { return "folder_id" }

func (v *folderVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	folder, err := svc.CrmV3.Folders.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "folder %s not found after deploy", name)
	}
	if folder.State != "ACTIVE" {
		return errors.Errorf("folder %s state is %s after deploy, want ACTIVE", name, folder.State)
	}
	if id := outputs["folder_id"]; id != "" && folder.Name != "folders/"+id {
		return errors.Errorf("folder %s resolved to name %q, want folders/%s from the folder_id output", name, folder.Name, id)
	}
	return nil
}

func (v *folderVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	folder, err := svc.CrmV3.Folders.Get(name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && (apiErr.Code == 404 || apiErr.Code == 403) {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing folder %s after destroy", name)
	}
	// Deleted folders linger in DELETE_REQUESTED for the 30-day recovery
	// window -- that IS the destroyed posture for this resource class.
	if folder.State == "DELETE_REQUESTED" {
		return nil
	}
	return errors.Errorf("folder %s still exists after destroy (state %s)", name, folder.State)
}
