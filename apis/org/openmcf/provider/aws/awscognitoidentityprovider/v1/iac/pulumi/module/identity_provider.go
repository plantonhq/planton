package module

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	cogidpv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscognitoidentityprovider/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cognito"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func identityProvider(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	providerDetails, err := buildProviderDetails(spec)
	if err != nil {
		return errors.Wrap(err, "failed to build provider details")
	}

	createdIdp, err := cognito.NewIdentityProvider(ctx, locals.Target.Metadata.Name, &cognito.IdentityProviderArgs{
		UserPoolId:      pulumi.String(spec.UserPoolId.GetValue()),
		ProviderName:    pulumi.String(spec.ProviderName),
		ProviderType:    pulumi.String(spec.ProviderType.String()),
		ProviderDetails: pulumi.ToStringMap(providerDetails),
		AttributeMapping: func() pulumi.StringMap {
			if len(spec.AttributeMapping) == 0 {
				return nil
			}
			return pulumi.ToStringMap(spec.AttributeMapping)
		}(),
		IdpIdentifiers: func() pulumi.StringArrayInput {
			if len(spec.IdpIdentifiers) == 0 {
				return nil
			}
			return pulumi.ToStringArray(spec.IdpIdentifiers)
		}(),
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create cognito identity provider")
	}

	ctx.Export(OpProviderName, createdIdp.ProviderName)
	ctx.Export(OpProviderType, createdIdp.ProviderType)

	return nil
}

// buildProviderDetails converts the typed oneof provider configuration into
// the flat map[string]string expected by the AWS Cognito API.
//
// Key naming conventions vary by provider type:
//   - OAuth providers (Google, Facebook, Amazon, Apple, OIDC): snake_case keys
//   - SAML: PascalCase keys (MetadataFile, IDPSignout, etc.)
func buildProviderDetails(spec *cogidpv1.AwsCognitoIdentityProviderSpec) (map[string]string, error) {
	details := make(map[string]string)

	switch spec.ProviderType {
	case cogidpv1.AwsCognitoIdentityProviderType_Google:
		cfg := spec.GetGoogle()
		details["client_id"] = cfg.ClientId
		details["client_secret"] = cfg.ClientSecret
		details["authorize_scopes"] = cfg.AuthorizeScopes

	case cogidpv1.AwsCognitoIdentityProviderType_Facebook:
		cfg := spec.GetFacebook()
		details["client_id"] = cfg.ClientId
		details["client_secret"] = cfg.ClientSecret
		details["authorize_scopes"] = cfg.AuthorizeScopes
		if cfg.ApiVersion != "" {
			details["api_version"] = cfg.ApiVersion
		}

	case cogidpv1.AwsCognitoIdentityProviderType_LoginWithAmazon:
		cfg := spec.GetLoginWithAmazon()
		details["client_id"] = cfg.ClientId
		details["client_secret"] = cfg.ClientSecret
		details["authorize_scopes"] = cfg.AuthorizeScopes

	case cogidpv1.AwsCognitoIdentityProviderType_SignInWithApple:
		cfg := spec.GetSignInWithApple()
		details["client_id"] = cfg.ClientId
		details["team_id"] = cfg.TeamId
		details["key_id"] = cfg.KeyId
		details["private_key"] = cfg.PrivateKey
		details["authorize_scopes"] = cfg.AuthorizeScopes

	case cogidpv1.AwsCognitoIdentityProviderType_OIDC:
		cfg := spec.GetOidc()
		details["client_id"] = cfg.ClientId
		details["oidc_issuer"] = cfg.OidcIssuer
		if cfg.AuthorizeScopes != "" {
			details["authorize_scopes"] = cfg.AuthorizeScopes
		}
		if cfg.ClientSecret != "" {
			details["client_secret"] = cfg.ClientSecret
		}
		if cfg.AttributesRequestMethod != "" {
			details["attributes_request_method"] = cfg.AttributesRequestMethod
		}
		if cfg.AuthorizeUrl != "" {
			details["authorize_url"] = cfg.AuthorizeUrl
		}
		if cfg.TokenUrl != "" {
			details["token_url"] = cfg.TokenUrl
		}
		if cfg.AttributesUrl != "" {
			details["attributes_url"] = cfg.AttributesUrl
		}
		if cfg.JwksUri != "" {
			details["jwks_uri"] = cfg.JwksUri
		}

	case cogidpv1.AwsCognitoIdentityProviderType_SAML:
		cfg := spec.GetSaml()
		if cfg.MetadataFile != "" {
			details["MetadataFile"] = cfg.MetadataFile
		}
		if cfg.MetadataUrl != "" {
			details["MetadataURL"] = cfg.MetadataUrl
		}
		if cfg.IdpSignOut {
			details["IDPSignout"] = strconv.FormatBool(cfg.IdpSignOut)
		}
		if cfg.IdpInit {
			details["IDPInit"] = strconv.FormatBool(cfg.IdpInit)
		}
		if cfg.EncryptedResponses {
			details["EncryptedResponses"] = strconv.FormatBool(cfg.EncryptedResponses)
		}
		if cfg.RequestSigningAlgorithm != "" {
			details["RequestSigningAlgorithm"] = cfg.RequestSigningAlgorithm
		}

	default:
		return nil, fmt.Errorf("unsupported provider type: %s", spec.ProviderType.String())
	}

	return details, nil
}
