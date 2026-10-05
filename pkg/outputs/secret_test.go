//go:build !codegen
// +build !codegen

package outputs

import (
	"testing"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretOutputs_ReadsTheSchemasMarks(t *testing.T) {
	secrets, err := SecretOutputs(catalogkindreflect.KindFromString("CloudflareZeroTrustAccessServiceToken"))
	require.NoError(t, err)
	assert.True(t, secrets["client_secret"], "client_secret is marked sensitive in the schema")
	assert.Contains(t, secrets, "client_id")
	assert.False(t, secrets["client_id"], "client_id is public")
}

func TestCaptureResult_IsSensitive(t *testing.T) {
	result := &CaptureResult{Secrets: map[string]bool{"password": true, "host": false, "auth_keys": true}}

	assert.True(t, result.IsSensitive("password"))
	assert.True(t, result.IsSensitive("auth_keys.primary"), "a dotted key inherits its top-level output's mark")
	assert.True(t, result.IsSensitive("auth-keys.primary"), "engine names spell underscores as hyphens")
	assert.False(t, result.IsSensitive("host"))
	assert.True(t, result.IsSensitive("passwordless"), "a key the schema does not declare is never printed")

	var none *CaptureResult
	assert.True(t, none.IsSensitive("host"), "with no capture, nothing is known to be safe")
}
