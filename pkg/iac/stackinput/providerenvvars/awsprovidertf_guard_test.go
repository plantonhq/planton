package providerenvvars

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// canonicalAwsProviderBlock is the single shape every AWS tofu module's provider.tf must use.
// The AWS provider block is intentionally empty: region and credentials are injected by the
// runtime as environment variables (see loadAwsEnvVars), and keyless connections have their
// web-identity JWT exchanged for temporary credentials before the tofu run. If you are adding a
// new AWS kind, copy this block verbatim into its iac/tf/provider.tf -- do NOT wire region or
// static keys into the HCL.
const canonicalAwsProviderBlock = `provider "aws" {
  # Region and credentials are injected by the runtime as environment variables
  # (AWS_REGION + AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY / AWS_SESSION_TOKEN), resolved
  # from the stack input's provider_config. For keyless (oidc / cross_account_trust)
  # connections the runtime performs the STS web-identity exchange and injects the resulting
  # short-lived credentials. Keep this block empty -- do not wire region or static keys here.
}`

// TestAwsProviderTfConvergence enforces that every AWS tofu module ships the canonical empty,
// injection-driven provider block -- the HCL half of the keyless AWS path (D6). It guards
// against (a) a new module reintroducing static-key / region wiring, and (b) drift in the
// shared block. The count assertion catches a new AWS kind that forgets the canonical shape.
func TestAwsProviderTfConvergence(t *testing.T) {
	// This guard reads the apis/ source tree, which is not present in the Bazel test sandbox
	// (Bazel sets TEST_SRCDIR). It runs under `go test` / `make test` and CI go-test instead.
	if os.Getenv("TEST_SRCDIR") != "" {
		t.Skip("convergence guard reads the apis source tree; skipped in the bazel sandbox")
	}

	root := repoRoot(t)
	matches, err := filepath.Glob(filepath.Join(root,
		"apis", "dev", "planton", "provider", "aws", "*", "v1", "iac", "tf", "provider.tf"))
	require.NoError(t, err)

	// Sized assertion: a new AWS tofu kind must adopt the canonical block (bump this with intent).
	assert.Len(t, matches, 66, "unexpected number of AWS tofu provider.tf files")

	forbidden := []string{
		"region =", "access_key", "secret_key", "session_token", "var.provider_config",
	}
	for _, path := range matches {
		b, err := os.ReadFile(path)
		require.NoError(t, err)
		content := string(b)

		assert.Containsf(t, content, canonicalAwsProviderBlock,
			"%s does not contain the canonical empty provider block", rel(root, path))
		for _, bad := range forbidden {
			assert.NotContainsf(t, content, bad,
				"%s reintroduces provider wiring (%q); region+credentials must flow via env injection",
				rel(root, path), bad)
		}
	}
}

// repoRoot walks up from this test file until it finds the module go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		require.NotEqualf(t, parent, dir, "reached filesystem root without finding go.mod")
		dir = parent
	}
}

func rel(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return r
}
