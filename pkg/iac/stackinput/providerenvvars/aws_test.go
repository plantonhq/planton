package providerenvvars

import (
	"context"
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awsprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws"
	"github.com/plantonhq/openmcf/pkg/iac/provider/aws/awswebidentity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

// stubCreds is what the fake resolver returns; the web-identity arm must emit these as env vars.
var stubCreds = awssdk.Credentials{
	AccessKeyID:     "ASIASTUBACCESSKEY",
	SecretAccessKey: "stub-secret-access-key",
	SessionToken:    "stub-session-token",
}

// awsConfigYaml renders an AwsProviderConfig the way the runner injects it (protojson, camelCase);
// loadProviderConfigProto reads it back through YAMLToJSON -> protojson, so this exercises the same
// round-trip the live stack input takes.
func awsConfigYaml(t *testing.T, cfg *awsprovider.AwsProviderConfig) []byte {
	t.Helper()
	b, err := protojson.Marshal(cfg)
	require.NoError(t, err)
	return b
}

// failingResolver fails the test if the STS exchange is reached (region-only, static, pulumi path).
func failingResolver(t *testing.T) awswebidentity.CredentialResolver {
	t.Helper()
	return func(_ context.Context, _ string,
		_ *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error) {
		t.Fatal("credential resolver must not be called for this case")
		return awssdk.Credentials{}, nil
	}
}

func TestLoadAwsEnvVars_WebIdentity_Resolve_EmitsTempCreds(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		Region: "ap-south-1", // connection region (should be overridden by the resource region)
		WebIdentity: &awsprovider.AwsWebIdentityProviderConfig{
			WebIdentityToken: "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			RoleArn:          "arn:aws:iam::123456789012:role/customer-oidc",
			SessionName:      "planton-oidc",
			Duration:         "1h",
		},
	}

	var gotRegion string
	resolve := func(_ context.Context, region string,
		_ *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error) {
		gotRegion = region
		return stubCreds, nil
	}

	env, err := loadAwsEnvVars(awsConfigYaml(t, cfg), true, "us-west-2",
		Options{ResolveAwsWebIdentity: true}, resolve)
	require.NoError(t, err)

	// The resource region wins over the connection region, and is what the exchange runs in.
	assert.Equal(t, "us-west-2", env["AWS_REGION"])
	assert.Equal(t, "us-west-2", gotRegion)
	assert.Equal(t, stubCreds.AccessKeyID, env["AWS_ACCESS_KEY_ID"])
	assert.Equal(t, stubCreds.SecretAccessKey, env["AWS_SECRET_ACCESS_KEY"])
	assert.Equal(t, stubCreds.SessionToken, env["AWS_SESSION_TOKEN"])
}

func TestLoadAwsEnvVars_WebIdentity_TwoHop_PassesChainToResolver(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		WebIdentity: &awsprovider.AwsWebIdentityProviderConfig{
			WebIdentityToken: "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			RoleArn:          "arn:aws:iam::066380525333:role/planton-base",
			ChainedAssumeRoles: []*awsprovider.AwsAssumeRoleConfig{
				{RoleArn: "arn:aws:iam::123456789012:role/customer-cat", ExternalId: "ext-secret-123", Duration: "1h"},
			},
		},
	}

	var gotChain []*awsprovider.AwsAssumeRoleConfig
	resolve := func(_ context.Context, _ string,
		wi *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error) {
		gotChain = wi.GetChainedAssumeRoles()
		return stubCreds, nil
	}

	env, err := loadAwsEnvVars(awsConfigYaml(t, cfg), true, "eu-west-1",
		Options{ResolveAwsWebIdentity: true}, resolve)
	require.NoError(t, err)

	require.Len(t, gotChain, 1)
	assert.Equal(t, "arn:aws:iam::123456789012:role/customer-cat", gotChain[0].GetRoleArn())
	assert.Equal(t, "ext-secret-123", gotChain[0].GetExternalId())
	assert.Equal(t, stubCreds.AccessKeyID, env["AWS_ACCESS_KEY_ID"])
}

func TestLoadAwsEnvVars_WebIdentity_NotResolved_RegionOnly(t *testing.T) {
	// The pulumi path (ResolveAwsWebIdentity=false): the in-program builder owns the exchange,
	// so the loader must NOT call STS and must emit region only.
	cfg := &awsprovider.AwsProviderConfig{
		WebIdentity: &awsprovider.AwsWebIdentityProviderConfig{
			WebIdentityToken: "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			RoleArn:          "arn:aws:iam::123456789012:role/customer-oidc",
		},
	}

	env, err := loadAwsEnvVars(awsConfigYaml(t, cfg), true, "us-east-1",
		Options{ResolveAwsWebIdentity: false}, failingResolver(t))
	require.NoError(t, err)

	assert.Equal(t, "us-east-1", env["AWS_REGION"])
	_, hasKey := env["AWS_ACCESS_KEY_ID"]
	assert.False(t, hasKey)
}

func TestLoadAwsEnvVars_StaticCredentials_WithSessionToken(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY123",
		SessionToken:    "FQoGZXIvYXdzEXAMPLE",
	}

	env, err := loadAwsEnvVars(awsConfigYaml(t, cfg), true, "us-east-2",
		Options{ResolveAwsWebIdentity: true}, failingResolver(t))
	require.NoError(t, err)

	assert.Equal(t, "us-east-2", env["AWS_REGION"])
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", env["AWS_ACCESS_KEY_ID"])
	assert.Equal(t, "wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY123", env["AWS_SECRET_ACCESS_KEY"])
	// AWS_SESSION_TOKEN must be emitted for temporary static (ASIA) credentials.
	assert.Equal(t, "FQoGZXIvYXdzEXAMPLE", env["AWS_SESSION_TOKEN"])
}

func TestLoadAwsEnvVars_StaticCredentials_NoSessionToken(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY123",
	}

	env, err := loadAwsEnvVars(awsConfigYaml(t, cfg), true, "us-east-2",
		Options{}, failingResolver(t))
	require.NoError(t, err)

	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", env["AWS_ACCESS_KEY_ID"])
	_, hasToken := env["AWS_SESSION_TOKEN"]
	assert.False(t, hasToken)
}

func TestLoadAwsEnvVars_RegionOnly_NoEmptyCredKeys(t *testing.T) {
	// runner auth mode: provider_config carries region/account but no credentials. The loader
	// must emit ONLY region -- an empty AWS_ACCESS_KEY_ID would poison the SDK ambient chain.
	cfg := &awsprovider.AwsProviderConfig{AccountId: "123456789012", Region: "ca-central-1"}

	env, err := loadAwsEnvVars(awsConfigYaml(t, cfg), true, "",
		Options{ResolveAwsWebIdentity: true}, failingResolver(t))
	require.NoError(t, err)

	// resourceRegion was empty, so the connection region is the documented fallback.
	assert.Equal(t, "ca-central-1", env["AWS_REGION"])
	_, hasKey := env["AWS_ACCESS_KEY_ID"]
	assert.False(t, hasKey)
	_, hasSecret := env["AWS_SECRET_ACCESS_KEY"]
	assert.False(t, hasSecret)
}

func TestLoadAwsEnvVars_NoProviderConfig_RegionFromTarget(t *testing.T) {
	// Standalone-CLI ambient case: no provider_config, but the empty provider block still needs
	// a region, so AWS_REGION must come from the resource's spec.region.
	env, err := loadAwsEnvVars(nil, false, "us-west-1",
		Options{ResolveAwsWebIdentity: true}, failingResolver(t))
	require.NoError(t, err)

	assert.Equal(t, "us-west-1", env["AWS_REGION"])
	assert.Len(t, env, 1)
}

func TestLoadAwsEnvVars_WebIdentity_ResolverError_Propagates(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		WebIdentity: &awsprovider.AwsWebIdentityProviderConfig{
			WebIdentityToken: "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			RoleArn:          "arn:aws:iam::123456789012:role/customer-oidc",
		},
	}
	resolve := func(_ context.Context, _ string,
		_ *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error) {
		return awssdk.Credentials{}, assert.AnError
	}

	_, err := loadAwsEnvVars(awsConfigYaml(t, cfg), true, "us-west-2",
		Options{ResolveAwsWebIdentity: true}, resolve)
	assert.Error(t, err)
}
