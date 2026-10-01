package module

import (
	"reflect"
	"testing"

	auth0clientfrommetadatadocumentv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0clientfrommetadatadocument/v1alpha1"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"google.golang.org/protobuf/proto"
)

const documentURL = "https://mcp-client.example.com/.well-known/oauth-client-metadata"

// setArgs names every argument clientArgs set (a non-nil Input). Walking the
// struct by reflection keeps the test honest when the SDK grows an argument:
// a new field is nil unless the module sets it.
func setArgs(args *auth0.ClientCimdArgs) []string {
	var set []string
	v := reflect.ValueOf(args).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Interface && !field.IsNil() {
			set = append(set, v.Type().Field(i).Name)
		}
	}
	return set
}

func TestClientArgsUnsetMeansUnmanaged(t *testing.T) {
	t.Run("a spec with only the document's URL sends only the URL", func(t *testing.T) {
		args := clientArgs(&auth0clientfrommetadatadocumentv1alpha1.Auth0ClientFromMetadataDocumentSpec{
			ExternalClientId: documentURL,
		})
		if got, want := setArgs(args), []string{"ExternalClientId"}; !reflect.DeepEqual(got, want) {
			t.Errorf("set arguments: got %v, want %v", got, want)
		}
	})

	t.Run("empty lists and an empty map are not sent", func(t *testing.T) {
		args := clientArgs(&auth0clientfrommetadatadocumentv1alpha1.Auth0ClientFromMetadataDocumentSpec{
			ExternalClientId:             documentURL,
			GrantTypes:                   []string{},
			AllowedOrigins:               []string{},
			WebOrigins:                   []string{},
			OrganizationDiscoveryMethods: []string{},
			ClientMetadata:               map[string]string{},
		})
		if got, want := setArgs(args), []string{"ExternalClientId"}; !reflect.DeepEqual(got, want) {
			t.Errorf("set arguments: got %v, want %v", got, want)
		}
	})

	t.Run("an explicit false or zero is sent", func(t *testing.T) {
		spec := &auth0clientfrommetadatadocumentv1alpha1.Auth0ClientFromMetadataDocumentSpec{
			ExternalClientId:         documentURL,
			ExternalClientIdVersion:  proto.Int32(0),
			RequireProofOfPossession: proto.Bool(false),
		}
		spec.SkipNonVerifiableCallbackUriConfirmationPrompt = proto.Bool(false)
		args := clientArgs(spec)
		want := []string{
			"ExternalClientId",
			"ExternalClientIdVersion",
			"RequireProofOfPossession",
			"SkipNonVerifiableCallbackUriConfirmationPrompt",
		}
		if got := setArgs(args); !reflect.DeepEqual(got, want) {
			t.Errorf("set arguments: got %v, want %v", got, want)
		}
	})
}

func TestClientArgsDeclaredSettings(t *testing.T) {
	spec := &auth0clientfrommetadatadocumentv1alpha1.Auth0ClientFromMetadataDocumentSpec{
		ExternalClientId:             documentURL,
		ExternalClientIdVersion:      proto.Int32(2),
		AppType:                      proto.String("native"),
		GrantTypes:                   []string{"authorization_code", "refresh_token"},
		Description:                  proto.String("MCP client"),
		AllowedOrigins:               []string{"https://mcp-client.example.com"},
		WebOrigins:                   []string{"https://mcp-client.example.com"},
		OidcConformant:               proto.Bool(true),
		RedirectionPolicy:            proto.String("open_redirect_protection"),
		OrganizationDiscoveryMethods: []string{"email"},
		DefaultOrganization: &auth0clientfrommetadatadocumentv1alpha1.Auth0ClientFromMetadataDocumentDefaultOrganization{
			OrganizationId: "org_abc123",
			Flows:          []string{"client_credentials"},
		},
		ClientMetadata: map[string]string{"owner": "platform"},
		JwtConfiguration: &auth0clientfrommetadatadocumentv1alpha1.Auth0ClientFromMetadataDocumentJwtConfiguration{
			Alg: proto.String("RS256"),
		},
		RefreshToken: &auth0clientfrommetadatadocumentv1alpha1.Auth0ClientFromMetadataDocumentRefreshToken{
			RotationType:   proto.String("rotating"),
			ExpirationType: proto.String("expiring"),
			TokenLifetime:  proto.Int32(2592000),
		},
		TokenQuota: &auth0clientfrommetadatadocumentv1alpha1.Auth0ClientFromMetadataDocumentTokenQuota{
			ClientCredentials: &auth0clientfrommetadatadocumentv1alpha1.Auth0ClientFromMetadataDocumentTokenQuotaClientCredentials{
				PerHour: proto.Int32(100),
			},
		},
	}
	args := clientArgs(spec)

	want := []string{
		"AllowedOrigins",
		"AppType",
		"ClientMetadata",
		"DefaultOrganization",
		"Description",
		"ExternalClientId",
		"ExternalClientIdVersion",
		"GrantTypes",
		"JwtConfiguration",
		"OidcConformant",
		"OrganizationDiscoveryMethods",
		"RedirectionPolicy",
		"RefreshToken",
		"TokenQuota",
		"WebOrigins",
	}
	if got := setArgs(args); !reflect.DeepEqual(got, want) {
		t.Errorf("set arguments: got %v, want %v", got, want)
	}

	// Inside a declared block, only the declared fields are sent.
	jwt := args.JwtConfiguration.(*auth0.ClientCimdJwtConfigurationArgs)
	if jwt.Alg == nil || jwt.LifetimeInSeconds != nil {
		t.Errorf("jwt_configuration: got alg %v and lifetime %v, want only alg", jwt.Alg, jwt.LifetimeInSeconds)
	}
	rt := args.RefreshToken.(*auth0.ClientCimdRefreshTokenArgs)
	if rt.RotationType == nil || rt.ExpirationType == nil || rt.TokenLifetime == nil ||
		rt.Leeway != nil || rt.IdleTokenLifetime != nil || rt.InfiniteTokenLifetime != nil || rt.InfiniteIdleTokenLifetime != nil {
		t.Errorf("refresh_token: got %+v, want only rotation_type, expiration_type and token_lifetime", rt)
	}
	credentials := args.TokenQuota.(*auth0.ClientCimdTokenQuotaArgs).ClientCredentials.(auth0.ClientCimdTokenQuotaClientCredentialsArgs)
	if credentials.PerHour == nil || credentials.PerDay != nil || credentials.Enforce != nil {
		t.Errorf("token_quota.client_credentials: got %+v, want only per_hour (enforce left to the provider's default)", credentials)
	}
}

func TestSummarizeValidation(t *testing.T) {
	valid := true
	cases := []struct {
		name        string
		validations []auth0.ClientCimdValidation
		want        documentValidation
	}{
		{
			name:        "no preview reads as not valid with nothing to report",
			validations: nil,
			want:        documentValidation{Warnings: []string{}, Violations: []string{}},
		},
		{
			name: "a valid document carries its warnings",
			validations: []auth0.ClientCimdValidation{{
				Valid:    &valid,
				Warnings: []string{"Grant type not supported: 'implicit'"},
			}},
			want: documentValidation{
				Valid:      true,
				Warnings:   []string{"Grant type not supported: 'implicit'"},
				Violations: []string{},
			},
		},
		{
			name: "an entry without a verdict reads as not valid",
			validations: []auth0.ClientCimdValidation{{
				Violations: []string{"client_id does not match the document URL"},
			}},
			want: documentValidation{
				Warnings:   []string{},
				Violations: []string{"client_id does not match the document URL"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := summarizeValidation(tc.validations); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
