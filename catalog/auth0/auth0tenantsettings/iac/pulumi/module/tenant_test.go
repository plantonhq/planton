package module

import (
	"reflect"
	"sort"
	"testing"

	auth0tenantsettingsv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0tenantsettings/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/protobuf/proto"
)

func stackInput(spec *auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec) *auth0tenantsettingsv1alpha1.Auth0TenantSettingsStackInput {
	return &auth0tenantsettingsv1alpha1.Auth0TenantSettingsStackInput{
		Target: &auth0tenantsettingsv1alpha1.Auth0TenantSettings{
			Metadata: &shared.CloudResourceMetadata{Name: "tenant-settings"},
			Spec:     spec,
		},
	}
}

func argsFor(spec *auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec) *auth0.TenantArgs {
	return tenantArgs(initializeLocals(stackInput(spec)))
}

// setArgs lists the names of the struct's fields that are not nil -- the
// arguments the provider will be sent.
func setArgs(args interface{}) []string {
	value := reflect.ValueOf(args).Elem()
	var set []string
	for i := 0; i < value.NumField(); i++ {
		if !value.Field(i).IsNil() {
			set = append(set, value.Type().Field(i).Name)
		}
	}
	sort.Strings(set)
	return set
}

func TestTenantArgsSendOnlyWhatTheSpecSets(t *testing.T) {
	cases := []struct {
		name string
		spec *auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec
		want []string
	}{
		{
			name: "an empty spec sends no argument, so an adopted tenant previews no change",
			spec: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec{},
			want: nil,
		},
		{
			name: "a spec managing only the default domain sends no tenant argument",
			spec: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec{
				DefaultCustomDomain: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "id.example.com"},
				},
			},
			want: nil,
		},
		{
			name: "a presentation setting sends only itself",
			spec: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec{FriendlyName: "Acme"},
			want: []string{"FriendlyName"},
		},
		{
			name: "empty lists are not managed",
			spec: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec{
				EnabledLocales:     []string{},
				AcrValuesSupported: []string{},
				AllowedLogoutUrls:  []string{},
				SessionLifetime:    proto.Float64(24),
			},
			want: []string{"SessionLifetime"},
		},
		{
			name: "the OAuth settings MCP clients depend on send only themselves",
			spec: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec{
				ClientIdMetadataDocumentSupported: proto.Bool(true),
				ResourceParameterProfile:          proto.String("compatibility"),
				Flags: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsFlags{
					EnableDynamicClientRegistration: proto.Bool(true),
				},
			},
			want: []string{"ClientIdMetadataDocumentSupported", "Flags", "ResourceParameterProfile"},
		},
		{
			name: "an explicit false is sent, never mistaken for unset",
			spec: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec{
				PushedAuthorizationRequestsSupported:           proto.Bool(false),
				SkipNonVerifiableCallbackUriConfirmationPrompt: proto.Bool(false),
			},
			want: []string{"PushedAuthorizationRequestsSupported", "SkipNonVerifiableCallbackUriConfirmationPrompt"},
		},
		{
			name: "resolved references are sent as their values",
			spec: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec{
				DefaultAudience: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "https://api.example.com"},
				},
				DefaultDirectory: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "Username-Password-Authentication"},
				},
			},
			want: []string{"DefaultAudience", "DefaultDirectory"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := setArgs(argsFor(tc.spec)); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTenantArgsNilSpecSendsNothing(t *testing.T) {
	if got := setArgs(argsFor(nil)); got != nil {
		t.Errorf("got %v, want no arguments", got)
	}
}

func TestDeclaredFlagsSendOnlyTheFlagsSet(t *testing.T) {
	args := argsFor(&auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec{
		Flags: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsFlags{
			EnableClientConnections: proto.Bool(false),
			EnableSso:               proto.Bool(true),
		},
	})
	flags, ok := args.Flags.(*auth0.TenantFlagsArgs)
	if !ok {
		t.Fatalf("flags: got %T, want *auth0.TenantFlagsArgs", args.Flags)
	}
	if got, want := setArgs(flags), []string{"EnableClientConnections", "EnableSso"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSkipPromptIsSentAsTheProvidersString(t *testing.T) {
	for _, value := range []bool{true, false} {
		args := argsFor(&auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec{
			SkipNonVerifiableCallbackUriConfirmationPrompt: proto.Bool(value),
		})
		got, ok := args.SkipNonVerifiableCallbackUriConfirmationPrompt.(pulumi.String)
		want := map[bool]string{true: "true", false: "false"}[value]
		if !ok || string(got) != want {
			t.Errorf("value %v: got %#v, want %q", value, args.SkipNonVerifiableCallbackUriConfirmationPrompt, want)
		}
	}
}

func TestDeclaredBlocksCarryOnlyTheirSetFields(t *testing.T) {
	args := argsFor(&auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec{
		Sessions: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsSessions{OidcLogoutPromptEnabled: proto.Bool(false)},
		ErrorPage: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsErrorPage{
			Url: proto.String("https://example.com/error"),
		},
		DefaultTokenQuota: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsDefaultTokenQuota{
			Clients: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsTokenQuota{
				ClientCredentials: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsClientCredentialsQuota{PerDay: proto.Int32(1000)},
			},
		},
	})

	sessions := args.Sessions.(*auth0.TenantSessionsArgs)
	if got, want := setArgs(sessions), []string{"OidcLogoutPromptEnabled"}; !reflect.DeepEqual(got, want) {
		t.Errorf("sessions: got %v, want %v (anonymous is sent only when declared)", got, want)
	}

	errorPage := args.ErrorPage.(*auth0.TenantErrorPageArgs)
	if got, want := setArgs(errorPage), []string{"Url"}; !reflect.DeepEqual(got, want) {
		t.Errorf("error page: got %v, want %v", got, want)
	}

	quota := args.DefaultTokenQuota.(*auth0.TenantDefaultTokenQuotaArgs)
	if got, want := setArgs(quota), []string{"Clients"}; !reflect.DeepEqual(got, want) {
		t.Errorf("token quota: got %v, want %v", got, want)
	}
	credentials := quota.Clients.(*auth0.TenantDefaultTokenQuotaClientsArgs).ClientCredentials.(auth0.TenantDefaultTokenQuotaClientsClientCredentialsArgs)
	if perDay, ok := credentials.PerDay.(pulumi.Int); !ok || int(perDay) != 1000 {
		t.Errorf("per_day: got %#v, want 1000", credentials.PerDay)
	}
	if credentials.PerHour != nil || credentials.Enforce != nil {
		t.Errorf("unset quota fields must not be sent: per_hour %#v, enforce %#v", credentials.PerHour, credentials.Enforce)
	}
}
