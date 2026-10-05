package aa_e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
)

// The identities a Cloudflare scenario may declare with planton.dev/e2e-identity.
//
//	connection-with-r2   the credential a Cloudflare connection carrying R2 keys resolves to:
//	                     the harness's API token plus the owner-arranged R2 key pair
//
// A scenario whose kind reads R2 objects through the S3 API (a Worker built from an
// r2_bundle) deploys as this identity, so both engines receive the R2 pair exactly the way
// a console deploy delivers it -- Pulumi through provider_config.r2, OpenTofu through the
// TF_VAR_r2_* variables the provider-environment loader derives from it -- instead of from
// whatever the machine running the lane happens to hold. Cloudflare mints R2 key pairs only
// in its dashboard, so the R2 half is arranged by the owner and named by the environment
// variables below; the scenario lists them in planton.dev/e2e-required-env so a lane
// without the arrangement skips instead of failing.
const identityConnectionWithR2 = "connection-with-r2"

// Environment variables carrying the owner-arranged R2 key pair.
const (
	EnvR2AccessKeyID     = "PLANTON_E2E_CLOUDFLARE_R2_ACCESS_KEY_ID"
	EnvR2SecretAccessKey = "PLANTON_E2E_CLOUDFLARE_R2_SECRET_ACCESS_KEY"
)

// ProvisionIdentity implements provider.IdentityProvisioner: it writes the lane's
// connection-shaped provider configuration and returns its path. Nothing is created at
// Cloudflare, so the cleanup only removes the file.
func (h *Harness) ProvisionIdentity(_ context.Context, _ *provider.KindTestContext, spec string) (string, func(), error) {
	if spec != identityConnectionWithR2 {
		return "", nil, errors.Errorf("unknown Cloudflare identity %q (supported: %q)", spec, identityConnectionWithR2)
	}
	apiToken := os.Getenv(EnvAPIToken)
	accessKeyID := os.Getenv(EnvR2AccessKeyID)
	secretAccessKey := os.Getenv(EnvR2SecretAccessKey)
	if apiToken == "" || accessKeyID == "" || secretAccessKey == "" {
		return "", nil, errors.Errorf("the %q identity needs %s, %s and %s", identityConnectionWithR2,
			EnvAPIToken, EnvR2AccessKeyID, EnvR2SecretAccessKey)
	}

	// JSON is YAML: the iac-input builder reads the file through its YAML path.
	providerConfig, err := json.Marshal(map[string]interface{}{
		"authScheme": "api_token",
		"apiToken":   apiToken,
		"r2": map[string]interface{}{
			"accessKeyId":     accessKeyID,
			"secretAccessKey": secretAccessKey,
		},
	})
	if err != nil {
		return "", nil, errors.Wrap(err, "encoding the lane identity's provider configuration")
	}
	dir, err := os.MkdirTemp("", "planton-e2e-cloudflare-identity-")
	if err != nil {
		return "", nil, errors.Wrap(err, "creating a directory for the lane identity's provider configuration")
	}
	path := filepath.Join(dir, "provider-config.yaml")
	if err := os.WriteFile(path, providerConfig, 0o600); err != nil {
		os.RemoveAll(dir)
		return "", nil, errors.Wrap(err, "writing the lane identity's provider configuration")
	}
	fmt.Printf("  [identity] %s: API token plus the arranged R2 key pair\n", identityConnectionWithR2)
	return path, func() { os.RemoveAll(dir) }, nil
}
