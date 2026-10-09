package kubetoken

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeServiceAccountKeyJSON builds a syntactically valid Google service-account key
// with a freshly generated RSA key, so JWT signing works without any real GCP secret.
// A non-empty tokenURL is embedded as the key's token_uri, which the ADC loader honors
// -- the offline switch for the ambient-chain tests.
func fakeServiceAccountKeyJSON(t *testing.T, tokenURL string) string {
	t.Helper()

	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	})

	fields := map[string]string{
		"type":         "service_account",
		"project_id":   "test-project",
		"client_email": "test-sa@test-project.iam.gserviceaccount.com",
		"private_key":  string(keyPem),
	}
	if tokenURL != "" {
		fields["token_uri"] = tokenURL
	}
	key, err := json.Marshal(fields)
	require.NoError(t, err)
	return string(key)
}

// fakeTokenEndpoint serves the OAuth2 token exchange offline, asserting the JWT
// assertion arrives as the oauth2 jwt flow sends it.
func fakeTokenEndpoint(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		assert.NotEmpty(t, r.Form.Get("assertion"))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"test-access-token","token_type":"Bearer","expires_in":3600}`)
	}))
	t.Cleanup(server.Close)
	return server
}

// shieldAmbientEnv isolates the test from the developer machine's real ambient
// identity: without it, a live GOOGLE_OAUTH_ACCESS_TOKEN or gcloud ADC file would
// change which arm the minter selects.
func shieldAmbientEnv(t *testing.T) {
	t.Helper()
	t.Setenv(GoogleOAuthAccessTokenEnvVar, "")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
}

// TestMintGkeToken_ExchangesKeyForAccessToken proves the full offline loop: parse the
// key, sign the JWT, exchange it at the (fake) token endpoint, and surface the access
// token with the endpoint-reported expiry.
func TestMintGkeToken_ExchangesKeyForAccessToken(t *testing.T) {
	tokenEndpoint := fakeTokenEndpoint(t)

	before := time.Now()
	token, err := MintGkeToken(context.Background(), GkeTokenOptions{
		ServiceAccountKeyJSON: fakeServiceAccountKeyJSON(t, ""),
		TokenURL:              tokenEndpoint.URL,
	})
	require.NoError(t, err)

	assert.Equal(t, "test-access-token", token.Value)
	assert.WithinDuration(t, before.Add(time.Hour), token.ExpiresAt, 30*time.Second,
		"expiry must come from the token response, never be assumed")
}

func TestMintGkeToken_RejectsMalformedKey(t *testing.T) {
	_, err := MintGkeToken(context.Background(), GkeTokenOptions{ServiceAccountKeyJSON: "{not json"})
	assert.Error(t, err, "a malformed key must be rejected at parse time, before any exchange")
}

// TestMintGkeToken_AmbientEnvToken: with no key on the options, a pre-minted
// GOOGLE_OAUTH_ACCESS_TOKEN is the ambient identity (the token a Planton runner mints
// for a connection's named gcloud configuration) and must win over ADC. The reported
// expiry is a refresh cadence, never zero -- the ExecCredential protocol reads a zero
// time as already expired.
func TestMintGkeToken_AmbientEnvToken(t *testing.T) {
	shieldAmbientEnv(t)
	t.Setenv(GoogleOAuthAccessTokenEnvVar, "ya29.ambient-test-token")

	before := time.Now()
	token, err := MintGkeToken(context.Background(), GkeTokenOptions{})
	require.NoError(t, err)

	assert.Equal(t, "ya29.ambient-test-token", token.Value)
	assert.WithinDuration(t, before.Add(suppliedTokenTTL), token.ExpiresAt, 30*time.Second)
}

// TestMintGkeToken_AmbientAdcFallback: with no key and no env token, Application
// Default Credentials mint the token -- proven offline by pointing
// GOOGLE_APPLICATION_CREDENTIALS at a fake key whose token_uri is a local endpoint.
func TestMintGkeToken_AmbientAdcFallback(t *testing.T) {
	shieldAmbientEnv(t)
	tokenEndpoint := fakeTokenEndpoint(t)

	keyPath := filepath.Join(t.TempDir(), "adc-key.json")
	require.NoError(t, os.WriteFile(keyPath,
		[]byte(fakeServiceAccountKeyJSON(t, tokenEndpoint.URL)), 0600))
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", keyPath)

	token, err := MintGkeToken(context.Background(), GkeTokenOptions{})
	require.NoError(t, err)
	assert.Equal(t, "test-access-token", token.Value)
}

// TestMintGkeToken_AmbientUnavailableIsLoud: a job must never silently run without an
// identity -- when the whole ambient chain is unavailable, the error names every arm
// that was tried.
func TestMintGkeToken_AmbientUnavailableIsLoud(t *testing.T) {
	shieldAmbientEnv(t)
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", filepath.Join(t.TempDir(), "does-not-exist.json"))

	_, err := MintGkeToken(context.Background(), GkeTokenOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambient Google credential chain")
}
