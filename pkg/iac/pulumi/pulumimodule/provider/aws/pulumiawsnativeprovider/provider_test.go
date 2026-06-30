package pulumiawsnativeprovider

import (
	"context"
	"errors"
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awsprovider "github.com/plantonhq/planton/apis/dev/planton/provider/aws"
	"github.com/plantonhq/planton/pkg/iac/provider/aws/awswebidentity"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubCreds is what the fake resolver returns; the dispatch must inject these onto the provider.
var stubCreds = awssdk.Credentials{
	AccessKeyID:     "ASIASTUBACCESSKEY",
	SecretAccessKey: "stub-secret-access-key",
	SessionToken:    "stub-session-token",
}

func TestBuildProviderArgs_NilConfig_RegionOnly(t *testing.T) {
	args, err := buildProviderArgs(context.Background(), nil, "us-west-2", failingResolver(t))
	require.NoError(t, err)
	require.NotNil(t, args)

	assert.Equal(t, pulumi.String("us-west-2"), args.Region)
	assert.Nil(t, args.AccessKey)
	assert.Nil(t, args.SecretKey)
	assert.Nil(t, args.Token)
}

func TestBuildProviderArgs_RunnerMode_RegionOnly(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{AccountId: "123456789012", Region: "eu-central-1"}

	args, err := buildProviderArgs(context.Background(), cfg, "eu-central-1", failingResolver(t))
	require.NoError(t, err)

	assert.Equal(t, pulumi.String("eu-central-1"), args.Region)
	assert.Nil(t, args.AccessKey)
}

func TestBuildProviderArgs_StaticCredentials(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		AccountId:       "123456789012",
		AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY123",
		Region:          "us-east-1",
		SessionToken:    "FQoGZXIvYXdzEXAMPLE",
	}

	args, err := buildProviderArgs(context.Background(), cfg, "us-east-1", failingResolver(t))
	require.NoError(t, err)

	assert.Equal(t, pulumi.String("AKIAIOSFODNN7EXAMPLE"), args.AccessKey)
	assert.Equal(t, pulumi.String("wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY123"), args.SecretKey)
	assert.Equal(t, pulumi.String("FQoGZXIvYXdzEXAMPLE"), args.Token)
	assert.Equal(t, pulumi.String("us-east-1"), args.Region)
}

func TestBuildProviderArgs_StaticCredentials_NoSessionToken(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		AccountId:       "123456789012",
		AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY123",
		Region:          "us-east-1",
	}

	args, err := buildProviderArgs(context.Background(), cfg, "us-east-1", failingResolver(t))
	require.NoError(t, err)

	assert.Equal(t, pulumi.String("AKIAIOSFODNN7EXAMPLE"), args.AccessKey)
	// An absent session token must stay nil, never an empty-string pointer.
	assert.Nil(t, args.Token)
}

func TestBuildProviderArgs_WebIdentity_SingleHop_InjectsResolvedCreds(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		AccountId: "123456789012",
		Region:    "us-west-2",
		WebIdentity: &awsprovider.AwsWebIdentityProviderConfig{
			WebIdentityToken: "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			RoleArn:          "arn:aws:iam::123456789012:role/customer-oidc",
			SessionName:      "planton-oidc",
			Duration:         "1h",
		},
	}

	var gotRegion string
	var gotWebIdentity *awsprovider.AwsWebIdentityProviderConfig
	resolve := func(_ context.Context, region string,
		wi *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error) {
		gotRegion = region
		gotWebIdentity = wi
		return stubCreds, nil
	}

	args, err := buildProviderArgs(context.Background(), cfg, "us-west-2", resolve)
	require.NoError(t, err)

	// The resolver received the resource region and the web-identity config.
	assert.Equal(t, "us-west-2", gotRegion)
	assert.Equal(t, "arn:aws:iam::123456789012:role/customer-oidc", gotWebIdentity.GetRoleArn())

	// The resolved temporary credentials were injected statically.
	assert.Equal(t, pulumi.String(stubCreds.AccessKeyID), args.AccessKey)
	assert.Equal(t, pulumi.String(stubCreds.SecretAccessKey), args.SecretKey)
	// The session token is present but secret-wrapped (an Output), so it is not a bare pulumi.String.
	require.NotNil(t, args.Token)
	assert.NotEqual(t, pulumi.String(stubCreds.SessionToken), args.Token)
}

func TestBuildProviderArgs_WebIdentity_TwoHop_PassesChainToResolver(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		Region: "us-west-2",
		WebIdentity: &awsprovider.AwsWebIdentityProviderConfig{
			WebIdentityToken: "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			RoleArn:          "arn:aws:iam::066380525333:role/planton-base",
			ChainedAssumeRoles: []*awsprovider.AwsAssumeRoleConfig{
				{
					RoleArn:    "arn:aws:iam::123456789012:role/customer-cat",
					ExternalId: "ext-secret-123",
					Duration:   "1h",
				},
			},
		},
	}

	var gotWebIdentity *awsprovider.AwsWebIdentityProviderConfig
	resolve := func(_ context.Context, _ string,
		wi *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error) {
		gotWebIdentity = wi
		return stubCreds, nil
	}

	args, err := buildProviderArgs(context.Background(), cfg, "us-west-2", resolve)
	require.NoError(t, err)

	require.Len(t, gotWebIdentity.GetChainedAssumeRoles(), 1)
	assert.Equal(t, "arn:aws:iam::123456789012:role/customer-cat",
		gotWebIdentity.GetChainedAssumeRoles()[0].GetRoleArn())
	assert.Equal(t, "ext-secret-123", gotWebIdentity.GetChainedAssumeRoles()[0].GetExternalId())
	assert.Equal(t, pulumi.String(stubCreds.AccessKeyID), args.AccessKey)
}

func TestBuildProviderArgs_WebIdentity_MissingToken_Errors(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		Region: "us-west-2",
		WebIdentity: &awsprovider.AwsWebIdentityProviderConfig{
			RoleArn: "arn:aws:iam::123456789012:role/customer-oidc",
		},
	}

	_, err := buildProviderArgs(context.Background(), cfg, "us-west-2", failingResolver(t))
	assert.Error(t, err)
}

func TestBuildProviderArgs_WebIdentity_ChainedHopMissingRoleArn_Errors(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		Region: "us-west-2",
		WebIdentity: &awsprovider.AwsWebIdentityProviderConfig{
			WebIdentityToken:   "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			RoleArn:            "arn:aws:iam::066380525333:role/planton-base",
			ChainedAssumeRoles: []*awsprovider.AwsAssumeRoleConfig{{ExternalId: "ext-only"}},
		},
	}

	_, err := buildProviderArgs(context.Background(), cfg, "us-west-2", failingResolver(t))
	assert.Error(t, err)
}

func TestBuildProviderArgs_WebIdentity_ResolverError_Propagates(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		Region: "us-west-2",
		WebIdentity: &awsprovider.AwsWebIdentityProviderConfig{
			WebIdentityToken: "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			RoleArn:          "arn:aws:iam::123456789012:role/customer-oidc",
		},
	}
	resolve := func(_ context.Context, _ string,
		_ *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error) {
		return awssdk.Credentials{}, errors.New("sts exchange failed")
	}

	_, err := buildProviderArgs(context.Background(), cfg, "us-west-2", resolve)
	assert.Error(t, err)
}

func TestProviderResourceName(t *testing.T) {
	// State continuity: the base name must stay "native-provider".
	assert.Equal(t, "native-provider", ProviderResourceName(nil))
	assert.Equal(t, "native-provider-replica", ProviderResourceName([]string{"replica"}))
}

// failingResolver returns a resolver that fails the test if invoked -- used by cases where the
// dispatch must never reach the STS exchange (region-only, static keys, validation errors).
func failingResolver(t *testing.T) awswebidentity.CredentialResolver {
	t.Helper()
	return func(_ context.Context, _ string,
		_ *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error) {
		t.Fatal("credential resolver must not be called for this case")
		return awssdk.Credentials{}, nil
	}
}
