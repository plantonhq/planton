package kubetoken

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeEntraTransport answers the two HTTP calls a client-credentials flow makes --
// the tenant's OpenID discovery document and the token request -- capturing the token
// request's form body so tests can assert exactly what was asked of Entra. Fully
// offline: no request ever leaves the process.
type fakeEntraTransport struct {
	t            *testing.T
	capturedForm url.Values
}

func (f *fakeEntraTransport) Do(req *http.Request) (*http.Response, error) {
	jsonResponse := func(body string) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}
	}

	if strings.Contains(req.URL.Path, "openid-configuration") {
		authority := "https://" + req.URL.Host + strings.TrimSuffix(req.URL.Path, "/v2.0/.well-known/openid-configuration")
		return jsonResponse(fmt.Sprintf(
			`{"issuer":"%[1]s/v2.0","authorization_endpoint":"%[1]s/oauth2/v2.0/authorize","token_endpoint":"%[1]s/oauth2/v2.0/token"}`,
			authority)), nil
	}

	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		require.NoError(f.t, err)
		form, err := url.ParseQuery(string(body))
		require.NoError(f.t, err)
		if form.Get("grant_type") != "" {
			f.capturedForm = form
		}
	}

	return jsonResponse(`{"token_type":"Bearer","expires_in":3600,"access_token":"test-aks-token"}`), nil
}

// TestMintAksToken_RequestsAksServerAudience guards the single fact an AKS token
// stands on: the request must be scoped to the AKS AAD server application, because
// the API server rejects tokens minted for any other audience.
func TestMintAksToken_RequestsAksServerAudience(t *testing.T) {
	transport := &fakeEntraTransport{t: t}

	before := time.Now()
	token, err := MintAksToken(context.Background(), AksTokenOptions{
		TenantID:     "11111111-2222-3333-4444-555555555555",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Transport:    transport,
	})
	require.NoError(t, err)

	require.NotNil(t, transport.capturedForm, "a token request must have been sent")
	assert.Contains(t, transport.capturedForm.Get("scope"), aksServerAppID,
		"the token must be requested for the AKS AAD server application")
	assert.Equal(t, "client_credentials", transport.capturedForm.Get("grant_type"))

	assert.Equal(t, "test-aks-token", token.Value)
	assert.WithinDuration(t, before.Add(time.Hour), token.ExpiresAt, 30*time.Second,
		"expiry must come from the token response, never be assumed")
}

// TestMintAksToken_SecretNeverInToken: the client secret authenticates the request
// but must never leak into the bearer token handed to the API server.
func TestMintAksToken_SecretNeverInToken(t *testing.T) {
	token, err := MintAksToken(context.Background(), AksTokenOptions{
		TenantID:     "11111111-2222-3333-4444-555555555555",
		ClientID:     "test-client-id",
		ClientSecret: "super-secret-value",
		Transport:    &fakeEntraTransport{t: t},
	})
	require.NoError(t, err)
	assert.NotContains(t, token.Value, "super-secret-value")
}

// TestMintAksToken_RequiresIdentityWithSecret: a client secret without the identity
// coordinates that give it meaning is a configuration error, reported before any
// network activity.
func TestMintAksToken_RequiresIdentityWithSecret(t *testing.T) {
	for _, opts := range []AksTokenOptions{
		{ClientSecret: "secret", ClientID: "client"},
		{ClientSecret: "secret", TenantID: "tenant"},
	} {
		_, err := MintAksToken(context.Background(), opts)
		assert.Error(t, err, fmt.Sprintf("%+v must be rejected", opts))
	}
}

// fakeTokenCredential satisfies azcore.TokenCredential for callers that resolve
// their credential elsewhere (the MintAksTokenWithCredential seam).
type fakeTokenCredential struct {
	capturedScopes []string
}

func (f *fakeTokenCredential) GetToken(_ context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	f.capturedScopes = opts.Scopes
	return azcore.AccessToken{Token: "external-cred-token", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

// TestMintAksTokenWithCredential_UsesAksScope: the external-credential seam must
// request the same AKS server-application scope as the options-based path -- the
// scope stays private to this one minting seam by construction.
func TestMintAksTokenWithCredential_UsesAksScope(t *testing.T) {
	cred := &fakeTokenCredential{}

	token, err := MintAksTokenWithCredential(context.Background(), cred)
	require.NoError(t, err)

	require.Len(t, cred.capturedScopes, 1)
	assert.Contains(t, cred.capturedScopes[0], aksServerAppID)
	assert.Equal(t, "external-cred-token", token.Value)
}

// TestMintAksToken_SuppliedAccessToken: a token the caller already minted (a keyless
// connection's, exchanged before the engine ran) is handed back as is, with no network
// activity and a re-invoke cadence well inside its real lifetime.
func TestMintAksToken_SuppliedAccessToken(t *testing.T) {
	transport := &fakeEntraTransport{t: t}

	before := time.Now()
	token, err := MintAksToken(context.Background(), AksTokenOptions{
		TenantID:    "11111111-2222-3333-4444-555555555555",
		AccessToken: "eyJ0eXAi.supplied-token",
		Transport:   transport,
	})
	require.NoError(t, err)

	assert.Equal(t, "eyJ0eXAi.supplied-token", token.Value)
	assert.WithinDuration(t, before.Add(suppliedTokenTTL), token.ExpiresAt, 30*time.Second)
	assert.Nil(t, transport.capturedForm, "a supplied token must not trigger a token request")
}

// TestMintAksToken_RefusesTokenWithSecret: a supplied token and a client secret name
// different credentials; the minter refuses to pick one.
func TestMintAksToken_RefusesTokenWithSecret(t *testing.T) {
	_, err := MintAksToken(context.Background(), AksTokenOptions{
		TenantID:     "11111111-2222-3333-4444-555555555555",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		AccessToken:  "eyJ0eXAi.supplied-token",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one credential")
}
