package awswebidentity

import (
	"strings"
	"testing"

	awsprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws"
)

func TestValidate(t *testing.T) {
	const (
		roleArn = "arn:aws:iam::123456789012:role/customer-oidc"
		token   = "eyJhbGciOiJSUzI1NiJ9.payload.sig"
	)

	t.Run("inline token with role is valid", func(t *testing.T) {
		if err := Validate(&awsprovider.AwsWebIdentityProviderConfig{
			WebIdentityToken: token,
			RoleArn:          roleArn,
		}); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("token file is rejected with a classic-only explanation", func(t *testing.T) {
		// The builder-side exchange is one-shot, so the re-read file source is not honored here.
		err := Validate(&awsprovider.AwsWebIdentityProviderConfig{
			WebIdentityTokenFile: "/var/run/planton/web-identity-token",
			RoleArn:              roleArn,
		})
		if err == nil {
			t.Fatal("expected an error for web_identity_token_file on a builder-side exchange engine")
		}
		if !strings.Contains(err.Error(), "web_identity_token_file") {
			t.Fatalf("error should name the offending field, got %v", err)
		}
	})

	t.Run("nil web identity is rejected", func(t *testing.T) {
		if err := Validate(nil); err == nil {
			t.Fatal("expected an error for nil web identity")
		}
	})

	t.Run("missing token and role is rejected", func(t *testing.T) {
		if err := Validate(&awsprovider.AwsWebIdentityProviderConfig{}); err == nil {
			t.Fatal("expected an error for missing token and role")
		}
	})

	t.Run("chained hop missing role_arn is rejected", func(t *testing.T) {
		err := Validate(&awsprovider.AwsWebIdentityProviderConfig{
			WebIdentityToken:   token,
			RoleArn:            roleArn,
			ChainedAssumeRoles: []*awsprovider.AwsAssumeRoleConfig{{ExternalId: "ext-only"}},
		})
		if err == nil {
			t.Fatal("expected an error for a chained hop missing role_arn")
		}
	})
}
