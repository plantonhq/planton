package module

import (
	"testing"

	auth0resourceserverv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0resourceserver/v1alpha1"
	"github.com/plantonhq/planton/shared"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/protobuf/proto"
)

func stackInput(spec *auth0resourceserverv1alpha1.Auth0ResourceServerSpec) *auth0resourceserverv1alpha1.Auth0ResourceServerStackInput {
	if spec.Identifier == "" {
		spec.Identifier = "https://api.example.com/"
	}
	return &auth0resourceserverv1alpha1.Auth0ResourceServerStackInput{
		Target: &auth0resourceserverv1alpha1.Auth0ResourceServer{
			Metadata: &shared.CloudResourceMetadata{Name: "api"},
			Spec:     spec,
		},
	}
}

func argsFor(spec *auth0resourceserverv1alpha1.Auth0ResourceServerSpec) *auth0.ResourceServerArgs {
	return resourceServerArgs(initializeLocals(stackInput(spec)))
}

// An API adopted into this kind with only its identifier declared must preview
// no change for anything the spec never declared: none of the settings beyond
// the always-sent token flags may be sent, and no default grant is declared.
func TestEmptySpecSendsNoNewArguments(t *testing.T) {
	locals := initializeLocals(stackInput(&auth0resourceserverv1alpha1.Auth0ResourceServerSpec{}))
	args := resourceServerArgs(locals)

	unsent := map[string]interface{}{
		"allow_online_access":                         args.AllowOnlineAccess,
		"allow_online_access_with_ephemeral_sessions": args.AllowOnlineAccessWithEphemeralSessions,
		"consent_policy":                              args.ConsentPolicy,
		"token_lifetime_for_anonymous_access_tokens":  args.TokenLifetimeForAnonymousAccessTokens,
		"verification_location":                       args.VerificationLocation,
		"signing_secret":                              args.SigningSecret,
		"access_token":                                args.AccessToken,
		"authorization_details":                       args.AuthorizationDetails,
		"authorization_policy":                        args.AuthorizationPolicy,
		"proof_of_possession":                         args.ProofOfPossession,
		"subject_type_authorization":                  args.SubjectTypeAuthorization,
		"token_encryption":                            args.TokenEncryption,
		"signing_alg":                                 args.SigningAlg,
		"token_lifetime":                              args.TokenLifetime,
		"token_lifetime_for_web":                      args.TokenLifetimeForWeb,
		"token_dialect":                               args.TokenDialect,
	}
	for name, value := range unsent {
		if value != nil {
			t.Errorf("%s is sent for an empty spec (%#v); unset must mean unmanaged", name, value)
		}
	}
	if len(locals.DefaultGrants) != 0 {
		t.Errorf("an empty spec declares %d default grants, want none", len(locals.DefaultGrants))
	}
	if args.Name != pulumi.String("api") {
		t.Errorf("name: got %#v, want metadata.name", args.Name)
	}
}

func TestDeclaredSettingsAreSent(t *testing.T) {
	args := argsFor(&auth0resourceserverv1alpha1.Auth0ResourceServerSpec{
		AllowOnlineAccess:                     proto.Bool(false),
		ConsentPolicy:                         proto.String("null"),
		TokenLifetimeForAnonymousAccessTokens: proto.Int32(86400),
		VerificationLocation:                  proto.String("https://api.example.com/jwks.json"),
		SigningSecret:                         proto.String("a-shared-secret-of-32-characters"),
		AuthorizationDetails: []*auth0resourceserverv1alpha1.Auth0ResourceServerAuthorizationDetail{
			{Type: proto.String("payment")},
		},
		AuthorizationPolicy: &auth0resourceserverv1alpha1.Auth0ResourceServerAuthorizationPolicy{PolicyId: proto.String("pol_1")},
		ProofOfPossession: &auth0resourceserverv1alpha1.Auth0ResourceServerProofOfPossession{
			Mechanism: proto.String("dpop"), Required: proto.Bool(true),
		},
		TokenEncryption: &auth0resourceserverv1alpha1.Auth0ResourceServerTokenEncryption{
			Format: proto.String("compact-nested-jwe"),
			EncryptionKey: &auth0resourceserverv1alpha1.Auth0ResourceServerTokenEncryptionKey{
				Algorithm: "RSA-OAEP-256", Pem: "-----BEGIN PUBLIC KEY-----\n-----END PUBLIC KEY-----\n",
			},
		},
	})

	for name, value := range map[string]interface{}{
		"allow_online_access (declared false)":       args.AllowOnlineAccess,
		"consent_policy":                             args.ConsentPolicy,
		"token_lifetime_for_anonymous_access_tokens": args.TokenLifetimeForAnonymousAccessTokens,
		"verification_location":                      args.VerificationLocation,
		"signing_secret":                             args.SigningSecret,
		"authorization_details":                      args.AuthorizationDetails,
		"authorization_policy":                       args.AuthorizationPolicy,
		"proof_of_possession":                        args.ProofOfPossession,
		"token_encryption":                           args.TokenEncryption,
	} {
		if value == nil {
			t.Errorf("%s is declared but not sent", name)
		}
	}
	if args.AllowOnlineAccessWithEphemeralSessions != nil {
		t.Errorf("allow_online_access_with_ephemeral_sessions is sent though undeclared")
	}

	pop := args.ProofOfPossession.(*auth0.ResourceServerProofOfPossessionArgs)
	if pop.Disable != nil || pop.RequiredFor != nil {
		t.Errorf("proof_of_possession sends undeclared settings: %+v", pop)
	}
	key := args.TokenEncryption.(*auth0.ResourceServerTokenEncryptionArgs).EncryptionKey.(*auth0.ResourceServerTokenEncryptionEncryptionKeyArgs)
	if key.Kid != nil || key.Name != nil {
		t.Errorf("encryption_key sends an undeclared kid or name: %+v", key)
	}
}

func TestAccessToken(t *testing.T) {
	t.Run("an empty block sends no claims mapping", func(t *testing.T) {
		args := argsFor(&auth0resourceserverv1alpha1.Auth0ResourceServerSpec{
			AccessToken: &auth0resourceserverv1alpha1.Auth0ResourceServerAccessToken{},
		})
		if got := args.AccessToken.(*auth0.ResourceServerAccessTokenArgs).ClaimsMapping; got != nil {
			t.Errorf("claims_mapping: got %#v, want nil", got)
		}
	})

	t.Run("a declared claims mapping without claims sends an empty list, which clears them", func(t *testing.T) {
		args := argsFor(&auth0resourceserverv1alpha1.Auth0ResourceServerSpec{
			AccessToken: &auth0resourceserverv1alpha1.Auth0ResourceServerAccessToken{
				ClaimsMapping: &auth0resourceserverv1alpha1.Auth0ResourceServerClaimsMapping{},
			},
		})
		mapping := args.AccessToken.(*auth0.ResourceServerAccessTokenArgs).ClaimsMapping.(*auth0.ResourceServerAccessTokenClaimsMappingArgs)
		claims, ok := mapping.CustomClaims.(auth0.ResourceServerAccessTokenClaimsMappingCustomClaimArray)
		if !ok || claims == nil || len(claims) != 0 {
			t.Errorf("custom_claims: got %#v, want an empty list", mapping.CustomClaims)
		}
	})

	t.Run("claims are sent in order", func(t *testing.T) {
		args := argsFor(&auth0resourceserverv1alpha1.Auth0ResourceServerSpec{
			AccessToken: &auth0resourceserverv1alpha1.Auth0ResourceServerAccessToken{
				ClaimsMapping: &auth0resourceserverv1alpha1.Auth0ResourceServerClaimsMapping{
					CustomClaims: []*auth0resourceserverv1alpha1.Auth0ResourceServerCustomClaim{
						{Name: "country", Expression: "anonymous_session.metadata.country"},
						{Name: "city", Expression: "anonymous_session.metadata.city"},
					},
				},
			},
		})
		mapping := args.AccessToken.(*auth0.ResourceServerAccessTokenArgs).ClaimsMapping.(*auth0.ResourceServerAccessTokenClaimsMappingArgs)
		claims := mapping.CustomClaims.(auth0.ResourceServerAccessTokenClaimsMappingCustomClaimArray)
		if len(claims) != 2 {
			t.Fatalf("custom_claims: got %d, want 2", len(claims))
		}
		first := claims[0].(*auth0.ResourceServerAccessTokenClaimsMappingCustomClaimArgs)
		if first.Name != pulumi.String("country") || first.Expression != pulumi.String("anonymous_session.metadata.country") {
			t.Errorf("first claim: got %+v", first)
		}
	})
}

func TestSubjectTypeAuthorization(t *testing.T) {
	t.Run("only the policies declared are sent", func(t *testing.T) {
		args := argsFor(&auth0resourceserverv1alpha1.Auth0ResourceServerSpec{
			SubjectTypeAuthorization: &auth0resourceserverv1alpha1.Auth0ResourceServerSubjectTypeAuthorization{
				User:   &auth0resourceserverv1alpha1.Auth0ResourceServerUserAuthorization{Policy: proto.String("require_client_grant")},
				Client: &auth0resourceserverv1alpha1.Auth0ResourceServerClientAuthorization{},
			},
		})
		sta := args.SubjectTypeAuthorization.(*auth0.ResourceServerSubjectTypeAuthorizationArgs)
		user, ok := sta.User.(*auth0.ResourceServerSubjectTypeAuthorizationUserArgs)
		if !ok || user.Policy != pulumi.String("require_client_grant") {
			t.Errorf("user: got %#v, want require_client_grant", sta.User)
		}
		if sta.Client != nil {
			t.Errorf("client: a block without its policy must not be sent, got %#v", sta.Client)
		}
		if sta.AnonymousUser != nil {
			t.Errorf("anonymous_user: undeclared, got %#v", sta.AnonymousUser)
		}
	})
}

func TestDefaultGrantArgs(t *testing.T) {
	audience := pulumi.String("https://api.example.com/")

	t.Run("a user grant sends its scopes and names no application", func(t *testing.T) {
		args := defaultGrantArgs(&auth0resourceserverv1alpha1.Auth0ResourceServerThirdPartyClientDefaultGrant{
			SubjectType: "user",
			Scopes:      []string{"tools:read", "tools:call"},
		}, audience)
		if args.DefaultFor != pulumi.String("third_party_clients") {
			t.Errorf("default_for: got %#v", args.DefaultFor)
		}
		if args.ClientId != nil {
			t.Errorf("client_id: a default grant names no application, got %#v", args.ClientId)
		}
		if args.SubjectType != pulumi.String("user") {
			t.Errorf("subject_type: got %#v", args.SubjectType)
		}
		if scopes := args.Scopes.(pulumi.StringArray); len(scopes) != 2 {
			t.Errorf("scopes: got %#v", scopes)
		}
		for name, value := range map[string]interface{}{
			"allow_all_scopes":            args.AllowAllScopes,
			"authorization_details_types": args.AuthorizationDetailsTypes,
			"organization_usage":          args.OrganizationUsage,
			"allow_any_organization":      args.AllowAnyOrganization,
		} {
			if value != nil {
				t.Errorf("%s is sent though undeclared", name)
			}
		}
	})

	t.Run("allow_all_scopes sends no scopes list", func(t *testing.T) {
		args := defaultGrantArgs(&auth0resourceserverv1alpha1.Auth0ResourceServerThirdPartyClientDefaultGrant{
			SubjectType:       "client",
			AllowAllScopes:    proto.Bool(true),
			OrganizationUsage: proto.String("allow"),
		}, audience)
		if args.Scopes != nil {
			t.Errorf("scopes: got %#v, want nil beside allow_all_scopes", args.Scopes)
		}
		if args.AllowAllScopes == nil || args.OrganizationUsage == nil {
			t.Errorf("allow_all_scopes and organization_usage are declared but not sent")
		}
	})

	t.Run("authorization_details_types are sent when declared", func(t *testing.T) {
		args := defaultGrantArgs(&auth0resourceserverv1alpha1.Auth0ResourceServerThirdPartyClientDefaultGrant{
			SubjectType:               "user",
			Scopes:                    []string{"read:payments"},
			AuthorizationDetailsTypes: []string{"payment"},
		}, audience)
		if types := args.AuthorizationDetailsTypes.(pulumi.StringArray); len(types) != 1 {
			t.Errorf("authorization_details_types: got %#v", types)
		}
	})
}
