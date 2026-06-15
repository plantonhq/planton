package pulumiawsprovider

import (
	"testing"

	awsprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildProviderArgs_NilConfig_RegionOnly(t *testing.T) {
	args, err := buildProviderArgs(nil, "us-west-2")
	require.NoError(t, err)
	require.NotNil(t, args)

	assert.Equal(t, pulumi.String("us-west-2"), args.Region)
	assert.Nil(t, args.AccessKey)
	assert.Nil(t, args.SecretKey)
	assert.Nil(t, args.Token)
	assert.Nil(t, args.AssumeRoleWithWebIdentity)
	assert.Nil(t, args.AssumeRoles)
}

func TestBuildProviderArgs_RunnerMode_RegionOnly(t *testing.T) {
	// No static keys and no web identity -> region only (ambient credential chain).
	cfg := &awsprovider.AwsProviderConfig{AccountId: "123456789012", Region: "eu-central-1"}

	args, err := buildProviderArgs(cfg, "eu-central-1")
	require.NoError(t, err)

	assert.Equal(t, pulumi.String("eu-central-1"), args.Region)
	assert.Nil(t, args.AccessKey)
	assert.Nil(t, args.AssumeRoleWithWebIdentity)
	assert.Nil(t, args.AssumeRoles)
}

func TestBuildProviderArgs_StaticCredentials(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		AccountId:       "123456789012",
		AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY123",
		Region:          "us-east-1",
		SessionToken:    "FQoGZXIvYXdzEXAMPLE",
	}

	args, err := buildProviderArgs(cfg, "us-east-1")
	require.NoError(t, err)

	assert.Equal(t, pulumi.String("AKIAIOSFODNN7EXAMPLE"), args.AccessKey)
	assert.Equal(t, pulumi.String("wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY123"), args.SecretKey)
	assert.Equal(t, pulumi.String("FQoGZXIvYXdzEXAMPLE"), args.Token)
	assert.Equal(t, pulumi.String("us-east-1"), args.Region)
	// Static and keyless are mutually exclusive.
	assert.Nil(t, args.AssumeRoleWithWebIdentity)
	assert.Nil(t, args.AssumeRoles)
}

func TestBuildProviderArgs_StaticCredentials_NoSessionToken(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		AccountId:       "123456789012",
		AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY123",
		Region:          "us-east-1",
	}

	args, err := buildProviderArgs(cfg, "us-east-1")
	require.NoError(t, err)

	assert.Equal(t, pulumi.String("AKIAIOSFODNN7EXAMPLE"), args.AccessKey)
	// An absent session token must stay nil, never an empty-string pointer.
	assert.Nil(t, args.Token)
}

func TestBuildProviderArgs_WebIdentity_SingleHop_Oidc(t *testing.T) {
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

	args, err := buildProviderArgs(cfg, "us-west-2")
	require.NoError(t, err)

	// Keyless path is taken; static keys must remain unset.
	assert.Nil(t, args.AccessKey)
	assert.Nil(t, args.SecretKey)
	// No chained hops for single-hop oidc.
	assert.Nil(t, args.AssumeRoles)

	require.NotNil(t, args.AssumeRoleWithWebIdentity)
	webIdentity, ok := args.AssumeRoleWithWebIdentity.(aws.ProviderAssumeRoleWithWebIdentityArgs)
	require.True(t, ok)
	assert.Equal(t, pulumi.String("arn:aws:iam::123456789012:role/customer-oidc"), webIdentity.RoleArn)
	assert.Equal(t, pulumi.String("eyJhbGciOiJSUzI1NiJ9.payload.sig"), webIdentity.WebIdentityToken)
	assert.Equal(t, pulumi.String("planton-oidc"), webIdentity.SessionName)
	assert.Equal(t, pulumi.String("1h"), webIdentity.Duration)
}

func TestBuildProviderArgs_WebIdentity_TwoHop_CrossAccountTrust(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		AccountId: "123456789012",
		Region:    "us-west-2",
		WebIdentity: &awsprovider.AwsWebIdentityProviderConfig{
			WebIdentityToken: "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			// Hop 1: the Planton base role in account 066380525333.
			RoleArn: "arn:aws:iam::066380525333:role/planton-base",
			ChainedAssumeRoles: []*awsprovider.AwsAssumeRoleConfig{
				{
					// Hop 2: the customer role, gated by external id, capped at 1h.
					RoleArn:    "arn:aws:iam::123456789012:role/customer-cat",
					ExternalId: "ext-secret-123",
					Duration:   "1h",
				},
			},
		},
	}

	args, err := buildProviderArgs(cfg, "us-west-2")
	require.NoError(t, err)

	require.NotNil(t, args.AssumeRoleWithWebIdentity)
	webIdentity, ok := args.AssumeRoleWithWebIdentity.(aws.ProviderAssumeRoleWithWebIdentityArgs)
	require.True(t, ok)
	assert.Equal(t, pulumi.String("arn:aws:iam::066380525333:role/planton-base"), webIdentity.RoleArn)

	require.NotNil(t, args.AssumeRoles)
	chain, ok := args.AssumeRoles.(aws.ProviderAssumeRoleArray)
	require.True(t, ok)
	require.Len(t, chain, 1)
	hop, ok := chain[0].(aws.ProviderAssumeRoleArgs)
	require.True(t, ok)
	assert.Equal(t, pulumi.String("arn:aws:iam::123456789012:role/customer-cat"), hop.RoleArn)
	assert.Equal(t, pulumi.String("ext-secret-123"), hop.ExternalId)
	assert.Equal(t, pulumi.String("1h"), hop.Duration)
}

func TestBuildProviderArgs_WebIdentity_MissingToken_Errors(t *testing.T) {
	cfg := &awsprovider.AwsProviderConfig{
		Region: "us-west-2",
		WebIdentity: &awsprovider.AwsWebIdentityProviderConfig{
			RoleArn: "arn:aws:iam::123456789012:role/customer-oidc",
		},
	}

	_, err := buildProviderArgs(cfg, "us-west-2")
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

	_, err := buildProviderArgs(cfg, "us-west-2")
	assert.Error(t, err)
}

func TestProviderResourceName(t *testing.T) {
	// State continuity: the base name must stay "classic-provider".
	assert.Equal(t, "classic-provider", ProviderResourceName(nil))
	assert.Equal(t, "classic-provider-replica", ProviderResourceName([]string{"replica"}))
	assert.Equal(t, "classic-provider-us-east-1", ProviderResourceName([]string{"us", "east-1"}))
}
