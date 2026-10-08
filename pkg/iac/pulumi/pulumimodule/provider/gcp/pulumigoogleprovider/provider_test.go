package pulumigoogleprovider

import (
	"testing"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validServiceAccountKey carries the four fields validateServiceAccountKey requires plus a
// PEM-shaped private key. Not a real credential.
const validServiceAccountKey = `{
	"type": "service_account",
	"project_id": "test-project",
	"private_key": "-----BEGIN PRIVATE KEY-----\nfake\n-----END PRIVATE KEY-----\n",
	"client_email": "test@test-project.iam.gserviceaccount.com"
}`

const testWifProvider = "//iam.googleapis.com/projects/123456/locations/global/workloadIdentityPools/test-pool/providers/test-provider"

const testIdentityToken = "eyJhbGciOiJSUzI1NiJ9.payload.sig"

const testServiceAccountEmail = "provisioner@test-project.iam.gserviceaccount.com"

func webIdentityConfig() *gcpprovider.GcpProviderConfig {
	return &gcpprovider.GcpProviderConfig{
		WebIdentity: &gcpprovider.GcpWebIdentityProviderConfig{
			WebIdentityToken:    testIdentityToken,
			Audience:            testWifProvider,
			ServiceAccountEmail: testServiceAccountEmail,
		},
	}
}

// requireWebIdentityCredentials asserts the keyless dispatch's shape and returns its
// external_credentials block: the typed OBJECT (gcp.ProviderExternalCredentialsArgs), the shape
// the provider-config encoder demands from pulumi-gcp v9.37 on. The single-element list this arm
// once sent (the pulumi-gcp#3869 workaround) fails at ValidateProviderConfig with "Expected an
// Object PropertyValue, found []" -- see the package doc.
func requireWebIdentityCredentials(t *testing.T, args *gcp.ProviderArgs) *gcp.ProviderExternalCredentialsArgs {
	t.Helper()

	require.NotNil(t, args)
	// Keyless never also carries a static or pre-minted credential.
	require.Nil(t, args.Credentials)
	require.Nil(t, args.AccessToken)

	credentials, ok := args.ExternalCredentials.(*gcp.ProviderExternalCredentialsArgs)
	require.True(t, ok, "externalCredentials must be the typed object, never a list")
	require.NotNil(t, credentials)
	return credentials
}

func TestBuildProviderInputs_NilConfig_Ambient(t *testing.T) {
	args, err := buildProviderInputs(nil)
	require.NoError(t, err)
	require.NotNil(t, args)

	assert.Nil(t, args.Credentials)
	assert.Nil(t, args.ExternalCredentials)
}

func TestBuildProviderInputs_EmptyConfig_Ambient(t *testing.T) {
	// No service-account key and no web identity -> ambient ADC chain.
	args, err := buildProviderInputs(&gcpprovider.GcpProviderConfig{})
	require.NoError(t, err)
	require.NotNil(t, args)

	assert.Nil(t, args.Credentials)
	assert.Nil(t, args.ExternalCredentials)
}

func TestBuildProviderInputs_ServiceAccountKey(t *testing.T) {
	cfg := &gcpprovider.GcpProviderConfig{
		ServiceAccountKey: validServiceAccountKey,
	}

	args, err := buildProviderInputs(cfg)
	require.NoError(t, err)
	require.NotNil(t, args)

	assert.Equal(t, pulumi.String(validServiceAccountKey), args.Credentials)
	// Static and keyless are mutually exclusive.
	assert.Nil(t, args.ExternalCredentials)
}

func TestBuildProviderInputs_ServiceAccountKey_InvalidJson_Errors(t *testing.T) {
	_, err := buildProviderInputs(&gcpprovider.GcpProviderConfig{
		ServiceAccountKey: "not-json",
	})
	assert.Error(t, err)
}

func TestBuildProviderInputs_ServiceAccountKey_MissingField_Errors(t *testing.T) {
	_, err := buildProviderInputs(&gcpprovider.GcpProviderConfig{
		ServiceAccountKey: `{"type": "service_account", "project_id": "p"}`,
	})
	assert.Error(t, err)
}

func TestBuildProviderInputs_ServiceAccountKey_NonPemPrivateKey_Errors(t *testing.T) {
	_, err := buildProviderInputs(&gcpprovider.GcpProviderConfig{
		ServiceAccountKey: `{
			"type": "service_account",
			"project_id": "test-project",
			"private_key": "MIIEvQIBADANBgkqhkiG9w0BAQEFAASC",
			"client_email": "test@test-project.iam.gserviceaccount.com"
		}`,
	})
	assert.Error(t, err)
}

func TestBuildProviderInputs_WebIdentity_TypedObjectShape(t *testing.T) {
	args, err := buildProviderInputs(webIdentityConfig())
	require.NoError(t, err)

	credentials := requireWebIdentityCredentials(t, args)

	// The audience must be passed through verbatim (byte-identity with the token's `aud`).
	assert.Equal(t, pulumi.String(testWifProvider), credentials.Audience)
	assert.Equal(t, pulumi.String(testServiceAccountEmail), credentials.ServiceAccountEmail)
}

func TestBuildProviderInputs_WebIdentity_IdentityTokenIsSecretWrapped(t *testing.T) {
	args, err := buildProviderInputs(webIdentityConfig())
	require.NoError(t, err)

	credentials := requireWebIdentityCredentials(t, args)

	// The SDK does NOT auto-secret-wrap identity_token, so the builder must: the wrapped
	// value is a secret Output, no longer the plain pulumi.String.
	require.NotNil(t, credentials.IdentityToken)
	_, isPlainString := credentials.IdentityToken.(pulumi.String)
	assert.False(t, isPlainString, "identityToken must be a ToSecret-wrapped Output, not a plain string")
}

func TestBuildProviderInputs_AccessToken_TypedArg(t *testing.T) {
	args, err := buildProviderInputs(&gcpprovider.GcpProviderConfig{
		AccessToken: "ya29.test-token",
	})
	require.NoError(t, err)
	require.NotNil(t, args)

	// The token rides the typed AccessToken field (the one field the SDK's NewProvider
	// auto-secret-wraps itself, so the builder passes it plain), never the raw keyless map.
	assert.Equal(t, pulumi.String("ya29.test-token"), args.AccessToken)
	assert.Nil(t, args.Credentials)
	assert.Nil(t, args.ExternalCredentials)
}

func TestBuildProviderInputs_AccessToken_WinsOverStaleKey(t *testing.T) {
	// An explicitly supplied short-lived token is the deliberate credential for this run;
	// a lingering service_account_key must not win.
	args, err := buildProviderInputs(&gcpprovider.GcpProviderConfig{
		AccessToken:       "ya29.test-token",
		ServiceAccountKey: validServiceAccountKey,
	})
	require.NoError(t, err)
	require.NotNil(t, args)

	assert.Equal(t, pulumi.String("ya29.test-token"), args.AccessToken)
	assert.Nil(t, args.Credentials)
}

func TestBuildProviderInputs_WebIdentity_WinsOverAccessToken(t *testing.T) {
	// Dispatch precedence is explicit: web_identity > access_token > service_account_key.
	// A config carrying both keyless federation and a token dispatches to federation.
	cfg := webIdentityConfig()
	cfg.AccessToken = "ya29.test-token"

	args, err := buildProviderInputs(cfg)
	require.NoError(t, err)

	requireWebIdentityCredentials(t, args)
}

func TestBuildProviderInputs_WebIdentity_TakesPrecedenceOverStaleKey(t *testing.T) {
	// A config carrying both dispatches to keyless: web identity is the deliberate mode
	// switch, a lingering service_account_key must not win.
	cfg := webIdentityConfig()
	cfg.ServiceAccountKey = validServiceAccountKey

	args, err := buildProviderInputs(cfg)
	require.NoError(t, err)

	requireWebIdentityCredentials(t, args)
}

func TestBuildProviderInputs_WebIdentity_MissingToken_Errors(t *testing.T) {
	cfg := webIdentityConfig()
	cfg.WebIdentity.WebIdentityToken = ""

	_, err := buildProviderInputs(cfg)
	assert.Error(t, err)
}

func TestBuildProviderInputs_WebIdentity_MissingAudience_Errors(t *testing.T) {
	cfg := webIdentityConfig()
	cfg.WebIdentity.Audience = ""

	_, err := buildProviderInputs(cfg)
	assert.Error(t, err)
}

func TestBuildProviderInputs_WebIdentity_MissingServiceAccountEmail_Errors(t *testing.T) {
	cfg := webIdentityConfig()
	cfg.WebIdentity.ServiceAccountEmail = ""

	_, err := buildProviderInputs(cfg)
	assert.Error(t, err)
}

func TestProviderResourceName(t *testing.T) {
	// State continuity: the base name must stay "google".
	assert.Equal(t, "google", ProviderResourceName(nil))
	assert.Equal(t, "google-replica", ProviderResourceName([]string{"replica"}))
	assert.Equal(t, "google-dns-zone", ProviderResourceName([]string{"dns", "zone"}))
}

// The project the component names reaches the provider, so an import records it (a bucket
// imported by its name otherwise lands in state with no project, and its preview replaces it).
func TestSetProject_ComponentProject(t *testing.T) {
	args := &gcp.ProviderArgs{}
	setProject(args, "pure-lantern-360309")
	assert.Equal(t, pulumi.String("pure-lantern-360309"), args.Project)
}

func TestSetProject_ResourceNameFormIsBare(t *testing.T) {
	args := &gcp.ProviderArgs{}
	setProject(args, "projects/pure-lantern-360309")
	assert.Equal(t, pulumi.String("pure-lantern-360309"), args.Project)
}

func TestSetProject_EmptyKeepsAmbientDefault(t *testing.T) {
	args := &gcp.ProviderArgs{}
	setProject(args, "")
	assert.Nil(t, args.Project)
}
