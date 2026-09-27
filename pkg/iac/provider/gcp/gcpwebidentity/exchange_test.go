package gcpwebidentity

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAudience = "//iam.googleapis.com/projects/123456/locations/global/workloadIdentityPools/test-pool/providers/test-provider"
	testAccount  = "provisioner@test-project.iam.gserviceaccount.com"
	testToken    = "eyJhbGciOiJSUzI1NiJ9.payload.sig"
)

func webIdentity() *gcpprovider.GcpWebIdentityProviderConfig {
	return &gcpprovider.GcpWebIdentityProviderConfig{
		WebIdentityToken: testToken, Audience: testAudience, ServiceAccountEmail: testAccount,
	}
}

// fakeGoogle stands in for STS and IAM Credentials. It records what each step was sent, and refuses
// the STS step when refuse is set, the way Google answers a token its pool does not trust.
type fakeGoogle struct {
	refuse        bool
	stsForm       url.Values
	impersonated  string
	impersonation string
}

func (g *fakeGoogle) serve(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/v1/token":
			require.NoError(t, r.ParseForm())
			g.stsForm = r.PostForm
			if g.refuse {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"error":"invalid_grant","error_description":"The audience in the token does not match"}`)
				return
			}
			_, _ = io.WriteString(w, `{"access_token":"federated","issued_token_type":"urn:ietf:params:oauth:token-type:access_token","token_type":"Bearer","expires_in":3600}`)
		case strings.HasSuffix(r.URL.Path, ":generateAccessToken"):
			g.impersonated = strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/projects/-/serviceAccounts/"), ":generateAccessToken")
			g.impersonation = r.Header.Get("Authorization")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"accessToken": "ya29.impersonated",
				"expireTime":  time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func endpointsOf(server *httptest.Server) endpoints {
	return endpoints{
		tokenURL:               server.URL + "/v1/token",
		impersonationURLFormat: server.URL + "/v1/projects/-/serviceAccounts/%s:generateAccessToken",
	}
}

func TestExchange_FederatesThenImpersonates(t *testing.T) {
	google := &fakeGoogle{}
	server := google.serve(t)

	token, err := exchange(context.Background(), webIdentity(), endpointsOf(server))
	require.NoError(t, err)

	// The minted token goes to STS as a JWT, for exactly the pool provider it was minted for.
	assert.Equal(t, testToken, google.stsForm.Get("subject_token"))
	assert.Equal(t, jwtSubjectTokenType, google.stsForm.Get("subject_token_type"))
	assert.Equal(t, testAudience, google.stsForm.Get("audience"))
	// The federated token then impersonates the named account, and that account's token is the answer.
	assert.Equal(t, testAccount, google.impersonated)
	assert.Equal(t, "Bearer federated", google.impersonation)
	assert.Equal(t, "ya29.impersonated", token)
}

func TestExchange_RefusedByGoogle_NamesAccountAndPoolProvider(t *testing.T) {
	server := (&fakeGoogle{refuse: true}).serve(t)

	_, err := exchange(context.Background(), webIdentity(), endpointsOf(server))
	require.Error(t, err)
	assert.Contains(t, err.Error(), testAccount)
	assert.Contains(t, err.Error(), testAudience)
	assert.Contains(t, err.Error(), "invalid_grant")
}

func TestValidate(t *testing.T) {
	require.NoError(t, Validate(webIdentity()))
	assert.Error(t, Validate(nil))
	for name, clear := range map[string]func(*gcpprovider.GcpWebIdentityProviderConfig){
		"token":    func(w *gcpprovider.GcpWebIdentityProviderConfig) { w.WebIdentityToken = "" },
		"audience": func(w *gcpprovider.GcpWebIdentityProviderConfig) { w.Audience = "" },
		"account":  func(w *gcpprovider.GcpWebIdentityProviderConfig) { w.ServiceAccountEmail = "" },
	} {
		t.Run(name, func(t *testing.T) {
			wi := webIdentity()
			clear(wi)
			assert.Error(t, Validate(wi))
		})
	}
}

func TestExchange_InvalidConfig_NeverReachesGoogle(t *testing.T) {
	google := &fakeGoogle{}
	server := google.serve(t)
	wi := webIdentity()
	wi.Audience = ""

	_, err := exchange(context.Background(), wi, endpointsOf(server))
	require.Error(t, err)
	assert.Nil(t, google.stsForm, "an invalid config must be refused before any exchange")
}
