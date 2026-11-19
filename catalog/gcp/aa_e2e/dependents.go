package aa_e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"google.golang.org/api/cloudkms/v1"
	"sigs.k8s.io/yaml"
)

// The classes of provider-created objects a GCP scenario may name with
// planton.dev/e2e-release-dependents (provider.DependentsReleaser).
const (
	// releaseAutokeyKey: a key handle makes Autokey create one Cloud HSM key
	// for it. Google keeps both the handle and the key (Cloud KMS deletes
	// neither), and an enabled HSM key version bills every month, so the
	// release schedules the lane's key's versions for destruction -- the
	// end state Cloud KMS allows, and the one the sweep accepts.
	releaseAutokeyKey = "autokey-key"
)

// ReleaseDependents implements provider.DependentsReleaser.
func (h *Harness) ReleaseDependents(ctx context.Context, tc *provider.KindTestContext, class string) error {
	manifestPath, _ := ctx.Value(provider.ManifestPathKey{}).(string)
	if manifestPath == "" {
		return errors.New("no manifest on the context; the runner passes the first act's")
	}
	switch class {
	case releaseAutokeyKey:
		return h.releaseAutokeyKey(ctx, tc, manifestPath)
	default:
		return errors.Errorf("unknown release class %q: the GCP harness releases %q", class, releaseAutokeyKey)
	}
}

// releaseAutokeyKey finds the key Autokey made for THIS lane's handle -- by
// the handle's run-scoped name in the expanded manifest, never by listing the
// project's shared autokey key ring, so a lane running beside it keeps its
// key -- and schedules every version still able to encrypt for destruction.
// A handle that never got created (a deploy that failed first) means Autokey
// made nothing.
func (h *Harness) releaseAutokeyKey(ctx context.Context, tc *provider.KindTestContext, manifestPath string) error {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return errors.Wrap(err, "reading the key handle manifest")
	}
	var manifest struct {
		Spec struct {
			Location      string `json:"location"`
			KeyHandleName string `json:"keyHandleName"`
		} `json:"spec"`
	}
	if err := yaml.Unmarshal(raw, &manifest); err != nil {
		return errors.Wrap(err, "parsing the key handle manifest")
	}
	if manifest.Spec.Location == "" || manifest.Spec.KeyHandleName == "" {
		return errors.New("the manifest names no key handle (spec.location and spec.keyHandleName)")
	}
	handleName := fmt.Sprintf("projects/%s/locations/%s/keyHandles/%s", h.services.Project, manifest.Spec.Location, manifest.Spec.KeyHandleName)

	var handle struct {
		KmsKey string `json:"kmsKey"`
	}
	status, err := h.restGet(ctx, "https://cloudkms.googleapis.com/v1/"+handleName, &handle)
	if status == http.StatusNotFound {
		lanef(tc, "  [release] key handle %s was never created; Autokey made no key", handleName)
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "reading the key handle %s", handleName)
	}
	if handle.KmsKey == "" {
		return errors.Errorf("key handle %s names no key", handleName)
	}

	versions := h.services.CloudKms.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions
	var destroyed []string
	err = versions.List(handle.KmsKey).Pages(ctx, func(page *cloudkms.ListCryptoKeyVersionsResponse) error {
		for _, version := range page.CryptoKeyVersions {
			if !keyVersionCanStillBeUsed(version.State) {
				continue
			}
			result, err := versions.Destroy(version.Name, &cloudkms.DestroyCryptoKeyVersionRequest{}).Context(ctx).Do()
			if err != nil {
				return errors.Wrapf(err, "scheduling %s for destruction", version.Name)
			}
			if keyVersionCanStillBeUsed(result.State) {
				return errors.Errorf("%s reads %s after its destroy was requested", version.Name, result.State)
			}
			destroyed = append(destroyed, version.Name)
		}
		return nil
	})
	if err != nil {
		return err
	}
	lanef(tc, "  [release] Autokey's key %s for %s: %d version(s) scheduled for destruction", handle.KmsKey, handleName, len(destroyed))
	return nil
}

// keyVersionCanStillBeUsed reports a Cloud KMS key version state that is not
// on its way out: anything but DESTROY_SCHEDULED, DESTROYED and the failed
// import or generation states, which hold no usable key material.
func keyVersionCanStillBeUsed(state string) bool {
	switch state {
	case "DESTROY_SCHEDULED", "DESTROYED", "IMPORT_FAILED", "GENERATION_FAILED":
		return false
	}
	return true
}

// restGet reads one Google REST resource with the harness's credential,
// for APIs the pinned typed clients do not cover. It returns the HTTP status
// so a caller can tell "absent" from a failure.
func (h *Harness) restGet(ctx context.Context, url string, out interface{}) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := h.services.RestClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, errors.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
	}
	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(out)
}
