package pulumiazurenativeprovider

import (
	"testing"

	azureprovider "github.com/plantonhq/planton/apis/dev/planton/provider/azure"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildProviderArgs_NilConfig_Ambient(t *testing.T) {
	args, err := buildProviderArgs(nil)
	require.NoError(t, err)
	require.NotNil(t, args)

	assert.Nil(t, args.ClientId)
	assert.Nil(t, args.ClientSecret)
	assert.Nil(t, args.TenantId)
	assert.Nil(t, args.SubscriptionId)
	assert.Nil(t, args.UseOidc)
	assert.Nil(t, args.OidcToken)
}

func TestBuildProviderArgs_RunnerMode_IdentityCoordinatesOnly(t *testing.T) {
	// No client secret and no web identity -> identity coordinates only (ambient chain).
	cfg := &azureprovider.AzureProviderConfig{
		ClientId:       "11111111-1111-1111-1111-111111111111",
		TenantId:       "22222222-2222-2222-2222-222222222222",
		SubscriptionId: "33333333-3333-3333-3333-333333333333",
	}

	args, err := buildProviderArgs(cfg)
	require.NoError(t, err)

	assert.Equal(t, pulumi.String("11111111-1111-1111-1111-111111111111"), args.ClientId)
	assert.Equal(t, pulumi.String("22222222-2222-2222-2222-222222222222"), args.TenantId)
	assert.Equal(t, pulumi.String("33333333-3333-3333-3333-333333333333"), args.SubscriptionId)
	assert.Nil(t, args.ClientSecret)
	assert.Nil(t, args.UseOidc)
	assert.Nil(t, args.OidcToken)
}

func TestBuildProviderArgs_StaticCredentials(t *testing.T) {
	cfg := &azureprovider.AzureProviderConfig{
		ClientId:       "11111111-1111-1111-1111-111111111111",
		ClientSecret:   "static-client-secret",
		TenantId:       "22222222-2222-2222-2222-222222222222",
		SubscriptionId: "33333333-3333-3333-3333-333333333333",
	}

	args, err := buildProviderArgs(cfg)
	require.NoError(t, err)

	assert.Equal(t, pulumi.String("11111111-1111-1111-1111-111111111111"), args.ClientId)
	assert.Equal(t, pulumi.String("static-client-secret"), args.ClientSecret)
	assert.Equal(t, pulumi.String("22222222-2222-2222-2222-222222222222"), args.TenantId)
	assert.Equal(t, pulumi.String("33333333-3333-3333-3333-333333333333"), args.SubscriptionId)
	// Static and keyless are mutually exclusive.
	assert.Nil(t, args.UseOidc)
	assert.Nil(t, args.OidcToken)
}

func TestBuildProviderArgs_WebIdentity_SetsSecretWrappedOidcToken(t *testing.T) {
	cfg := &azureprovider.AzureProviderConfig{
		ClientId:       "11111111-1111-1111-1111-111111111111",
		TenantId:       "22222222-2222-2222-2222-222222222222",
		SubscriptionId: "33333333-3333-3333-3333-333333333333",
		WebIdentity: &azureprovider.AzureWebIdentityProviderConfig{
			WebIdentityToken: "eyJhbGciOiJSUzI1NiJ9.payload.sig",
		},
	}

	args, err := buildProviderArgs(cfg)
	require.NoError(t, err)

	assert.Equal(t, pulumi.Bool(true), args.UseOidc)
	// This SDK's NewProvider does NOT auto-secret-wrap OidcToken, so the builder must: the
	// token is present but secret-wrapped (an Output), never a bare pulumi.String.
	require.NotNil(t, args.OidcToken)
	assert.NotEqual(t, pulumi.String("eyJhbGciOiJSUzI1NiJ9.payload.sig"), args.OidcToken)
	assert.Equal(t, pulumi.String("11111111-1111-1111-1111-111111111111"), args.ClientId)
	assert.Equal(t, pulumi.String("22222222-2222-2222-2222-222222222222"), args.TenantId)
	assert.Equal(t, pulumi.String("33333333-3333-3333-3333-333333333333"), args.SubscriptionId)
	// Keyless must never carry a client secret.
	assert.Nil(t, args.ClientSecret)
}

func TestBuildProviderArgs_WebIdentity_TakesPrecedenceOverStaleSecret(t *testing.T) {
	// A config carrying both dispatches to keyless: web identity is the deliberate mode
	// switch, a lingering client_secret must not win.
	cfg := &azureprovider.AzureProviderConfig{
		ClientId:       "11111111-1111-1111-1111-111111111111",
		ClientSecret:   "stale-client-secret",
		TenantId:       "22222222-2222-2222-2222-222222222222",
		SubscriptionId: "33333333-3333-3333-3333-333333333333",
		WebIdentity: &azureprovider.AzureWebIdentityProviderConfig{
			WebIdentityToken: "eyJhbGciOiJSUzI1NiJ9.payload.sig",
		},
	}

	args, err := buildProviderArgs(cfg)
	require.NoError(t, err)

	assert.Equal(t, pulumi.Bool(true), args.UseOidc)
	assert.NotNil(t, args.OidcToken)
	assert.Nil(t, args.ClientSecret)
}

func TestBuildProviderArgs_WebIdentity_MissingToken_Errors(t *testing.T) {
	cfg := &azureprovider.AzureProviderConfig{
		ClientId:       "11111111-1111-1111-1111-111111111111",
		TenantId:       "22222222-2222-2222-2222-222222222222",
		SubscriptionId: "33333333-3333-3333-3333-333333333333",
		WebIdentity:    &azureprovider.AzureWebIdentityProviderConfig{},
	}

	_, err := buildProviderArgs(cfg)
	assert.Error(t, err)
}

func TestProviderResourceName(t *testing.T) {
	// State continuity: the base name must stay "azure".
	assert.Equal(t, "azure", ProviderResourceName(nil))
	assert.Equal(t, "azure-replica", ProviderResourceName([]string{"replica"}))
	assert.Equal(t, "azure-aks-nodepool", ProviderResourceName([]string{"aks", "nodepool"}))
}
