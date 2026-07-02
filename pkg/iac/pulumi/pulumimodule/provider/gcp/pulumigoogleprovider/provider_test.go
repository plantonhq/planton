package pulumigoogleprovider

import (
	"testing"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
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

func TestBuildProviderArgs_NilConfig_Ambient(t *testing.T) {
	args, err := buildProviderArgs(nil)
	require.NoError(t, err)
	require.NotNil(t, args)

	assert.Nil(t, args.Credentials)
	assert.Nil(t, args.ExternalCredentials)
}

func TestBuildProviderArgs_EmptyConfig_Ambient(t *testing.T) {
	// No service-account key and no web identity -> ambient ADC chain.
	args, err := buildProviderArgs(&gcpprovider.GcpProviderConfig{})
	require.NoError(t, err)

	assert.Nil(t, args.Credentials)
	assert.Nil(t, args.ExternalCredentials)
}

func TestBuildProviderArgs_ServiceAccountKey(t *testing.T) {
	cfg := &gcpprovider.GcpProviderConfig{
		ServiceAccountKey: validServiceAccountKey,
	}

	args, err := buildProviderArgs(cfg)
	require.NoError(t, err)

	assert.Equal(t, pulumi.String(validServiceAccountKey), args.Credentials)
	// Static and keyless are mutually exclusive.
	assert.Nil(t, args.ExternalCredentials)
}

func TestBuildProviderArgs_ServiceAccountKey_InvalidJson_Errors(t *testing.T) {
	_, err := buildProviderArgs(&gcpprovider.GcpProviderConfig{
		ServiceAccountKey: "not-json",
	})
	assert.Error(t, err)
}

func TestBuildProviderArgs_ServiceAccountKey_MissingField_Errors(t *testing.T) {
	_, err := buildProviderArgs(&gcpprovider.GcpProviderConfig{
		ServiceAccountKey: `{"type": "service_account", "project_id": "p"}`,
	})
	assert.Error(t, err)
}

func TestBuildProviderArgs_ServiceAccountKey_NonPemPrivateKey_Errors(t *testing.T) {
	_, err := buildProviderArgs(&gcpprovider.GcpProviderConfig{
		ServiceAccountKey: `{
			"type": "service_account",
			"project_id": "test-project",
			"private_key": "MIIEvQIBADANBgkqhkiG9w0BAQEFAASC",
			"client_email": "test@test-project.iam.gserviceaccount.com"
		}`,
	})
	assert.Error(t, err)
}

func TestBuildProviderArgs_WebIdentity_SetsSecretWrappedExternalCredentials(t *testing.T) {
	cfg := &gcpprovider.GcpProviderConfig{
		WebIdentity: &gcpprovider.GcpWebIdentityProviderConfig{
			WebIdentityToken:    "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			Audience:            testWifProvider,
			ServiceAccountEmail: "provisioner@test-project.iam.gserviceaccount.com",
		},
	}

	args, err := buildProviderArgs(cfg)
	require.NoError(t, err)
	require.NotNil(t, args.ExternalCredentials)

	external, ok := args.ExternalCredentials.(*gcp.ProviderExternalCredentialsArgs)
	require.True(t, ok)

	// The audience must be passed through verbatim (byte-identity with the token's `aud`).
	assert.Equal(t, pulumi.String(testWifProvider), external.Audience)
	assert.Equal(t,
		pulumi.String("provisioner@test-project.iam.gserviceaccount.com"),
		external.ServiceAccountEmail)

	// The SDK does NOT auto-secret-wrap identity_token, so the builder must: the wrapped
	// value is a secret Output, no longer the plain pulumi.String.
	require.NotNil(t, external.IdentityToken)
	assert.NotEqual(t, pulumi.String("eyJhbGciOiJSUzI1NiJ9.payload.sig"), external.IdentityToken)

	// Keyless must never carry a static credential.
	assert.Nil(t, args.Credentials)
}

func TestBuildProviderArgs_WebIdentity_TakesPrecedenceOverStaleKey(t *testing.T) {
	// A config carrying both dispatches to keyless: web identity is the deliberate mode
	// switch, a lingering service_account_key must not win.
	cfg := &gcpprovider.GcpProviderConfig{
		ServiceAccountKey: validServiceAccountKey,
		WebIdentity: &gcpprovider.GcpWebIdentityProviderConfig{
			WebIdentityToken:    "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			Audience:            testWifProvider,
			ServiceAccountEmail: "provisioner@test-project.iam.gserviceaccount.com",
		},
	}

	args, err := buildProviderArgs(cfg)
	require.NoError(t, err)

	assert.NotNil(t, args.ExternalCredentials)
	assert.Nil(t, args.Credentials)
}

func TestBuildProviderArgs_WebIdentity_MissingToken_Errors(t *testing.T) {
	_, err := buildProviderArgs(&gcpprovider.GcpProviderConfig{
		WebIdentity: &gcpprovider.GcpWebIdentityProviderConfig{
			Audience:            testWifProvider,
			ServiceAccountEmail: "provisioner@test-project.iam.gserviceaccount.com",
		},
	})
	assert.Error(t, err)
}

func TestBuildProviderArgs_WebIdentity_MissingAudience_Errors(t *testing.T) {
	_, err := buildProviderArgs(&gcpprovider.GcpProviderConfig{
		WebIdentity: &gcpprovider.GcpWebIdentityProviderConfig{
			WebIdentityToken:    "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			ServiceAccountEmail: "provisioner@test-project.iam.gserviceaccount.com",
		},
	})
	assert.Error(t, err)
}

func TestBuildProviderArgs_WebIdentity_MissingServiceAccountEmail_Errors(t *testing.T) {
	_, err := buildProviderArgs(&gcpprovider.GcpProviderConfig{
		WebIdentity: &gcpprovider.GcpWebIdentityProviderConfig{
			WebIdentityToken: "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			Audience:         testWifProvider,
		},
	})
	assert.Error(t, err)
}

func TestProviderResourceName(t *testing.T) {
	// State continuity: the base name must stay "google".
	assert.Equal(t, "google", ProviderResourceName(nil))
	assert.Equal(t, "google-replica", ProviderResourceName([]string{"replica"}))
	assert.Equal(t, "google-dns-zone", ProviderResourceName([]string{"dns", "zone"}))
}
