// Package pulumigoogleprovider is the single, convergent place where every GCP pulumi module
// builds its gcp.Provider from the IaC input's GcpProviderConfig. Every GCP module already
// routes through Get, so extending the credential dispatch here extends it for all of them at
// once. It mirrors the sibling per-cloud builders (e.g. pulumiazureprovider,
// pulumiazurenativeprovider, pulumiawsprovider) so a coding agent can learn the GCP
// credential-resolution path by reading one file.
//
// It dispatches on which fields of GcpProviderConfig are populated, supporting every auth mode
// with a single seam:
//   - web_identity set        -> keyless OIDC federation. The minted JWT is handed to the
//     provider inline (external_credentials) together with the workload identity pool provider
//     audience and the service account to impersonate; the provider plugin performs the STS
//     exchange + impersonation itself, out of our process.
//   - access_token set        -> a pre-minted short-lived Google OAuth2 access token (e.g.
//     issued by a credential broker at deploy time), passed through the typed AccessToken
//     field. This is the one credential field the SDK's NewProvider auto-secret-wraps itself,
//     so no explicit pulumi.ToSecret is needed here.
//   - service_account_key set -> static service-account key JSON (today's mode).
//   - none                    -> no credential. The provider falls back to Google's ambient
//     Application Default Credentials chain (e.g. a self-hosted runner's attached service
//     account or `gcloud auth application-default`).
//
// Why NO builder-side token exchange (deliberate contrast with the AWS builders): the AWS
// builders exchange the web-identity token themselves only because pulumi-aws's provider-native
// path is broken upstream (pulumi-aws#6228). The pulumi-gcp provider consumes the inline token
// natively -- the plugin exchanges it at GCP STS and impersonates the target service account --
// so a builder-side exchange here would add a dependency and put credentials in our process for
// zero benefit. Do not "converge" this builder toward the AWS workaround shape.
//
// Why the keyless arm passes external_credentials as the typed OBJECT
// (gcp.ProviderExternalCredentialsArgs): that is the field's schema shape, and the shape the
// provider-config encoder demands from pulumi-gcp v9.37 on. Before that, at v9.29, the encoder
// rejected the object ("Expected an Array PropertyValue", pulumi-gcp#3869), so this arm
// registered raw properties with a single-element LIST instead. When the catalog moved to v9.37
// (2026-10-02) the encoder flipped and the list itself became the failure -- `objectEncoder
// failed on property "external_credentials": Expected an Object PropertyValue, found []` at
// ValidateProviderConfig, before any GCP call -- which broke every keyless deploy on Pulumi.
// The typed form is now the only form; TestBuildProviderInputs_WebIdentity_TypedObjectShape pins
// it, and a keyless preview with a fake token is the live check: it must get past provider
// configuration and fail only at the STS exchange.
//
// Freshness: the inline token is consumed at provider configure time and the resulting
// impersonated credentials are provider-managed for the run. Each pulumi operation re-runs the
// module program with a freshly minted JWT, so the exchange always sees a fresh token whose
// validity covers that one operation. No token is ever written to disk.
//
// SECURITY: this SDK's NewProvider auto-secret-wraps only the plain AccessToken field, NOT
// external_credentials.identity_token -- an unwrapped token would land in Pulumi state and
// `preview` output in plaintext. This builder therefore wraps the token with pulumi.ToSecret
// itself (the same obligation pulumiazurenativeprovider carries for its OidcToken).
package pulumigoogleprovider

import (
	"encoding/json"
	"fmt"
	"reflect"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/pulumi/pulumioutput"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Get builds a gcp.Provider from the given GcpProviderConfig. There is no region argument: the
// GCP provider is not region-scoped -- each resource carries its own location. nameSuffixes
// disambiguate the provider resource name when a module needs more than one provider.
func Get(ctx *pulumi.Context, gcpProviderConfig *gcpprovider.GcpProviderConfig,
	nameSuffixes ...string) (*gcp.Provider, error) {
	return get(ctx, gcpProviderConfig, false, "", nameSuffixes)
}

// GetWithUserProjectOverride builds the provider like Get, additionally arming
// user_project_override so every API call attributes quota and billing to the RESOURCE's
// project (the X-Goog-User-Project header). Kinds whose APIs REQUIRE a quota project on
// user-credential calls use this instead of Get — the Identity Toolkit API was the first:
// without the override, a deploy under plain `gcloud auth application-default login` fails at
// create with 403 "The identitytoolkit.googleapis.com API requires a quota project, which is
// not set by default" (live-verified). Service-account and keyless credentials are unaffected
// by the header, so arming it is safe across every credential mode.
func GetWithUserProjectOverride(ctx *pulumi.Context, gcpProviderConfig *gcpprovider.GcpProviderConfig,
	nameSuffixes ...string) (*gcp.Provider, error) {
	return get(ctx, gcpProviderConfig, true, "", nameSuffixes)
}

// GetWithQuotaProject builds the provider like GetWithUserProjectOverride and additionally NAMES
// the project every call attributes quota to (billing_project). The override alone attributes a
// RESOURCE call to the resource's own project, but a DATA SOURCE read carries no project the
// header can borrow -- the Firebase app-config and Admin SDK reads are the case -- so under a
// user's gcloud sign-in the header goes unset and Google attributes the call to its shared ADC
// project, where the API is disabled: 403 "requires a quota project" on the read-back after a
// create that succeeded (live-verified 2026-09-18 on GcpFirebaseProject under a runner-mode
// user credential). Service-account and keyless credentials never need it and are unaffected.
// Kinds whose modules READ through a data source pass their resource's project here; an empty
// quotaProject degrades to the override alone.
func GetWithQuotaProject(ctx *pulumi.Context, gcpProviderConfig *gcpprovider.GcpProviderConfig,
	quotaProject string, nameSuffixes ...string) (*gcp.Provider, error) {
	return get(ctx, gcpProviderConfig, true, quotaProject, nameSuffixes)
}

func get(ctx *pulumi.Context, gcpProviderConfig *gcpprovider.GcpProviderConfig,
	userProjectOverride bool, quotaProject string, nameSuffixes []string) (*gcp.Provider, error) {
	args, err := buildProviderInputs(gcpProviderConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build google provider args")
	}
	if userProjectOverride {
		args.UserProjectOverride = pulumi.Bool(true)
	}
	if quotaProject != "" {
		args.BillingProject = pulumi.String(quotaProject)
	}

	googleProvider, err := gcp.NewProvider(ctx, ProviderResourceName(nameSuffixes), args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create google provider")
	}

	return googleProvider, nil
}

// buildProviderInputs is the pure, side-effect-free core of the builder: it maps a
// GcpProviderConfig to the provider's registration inputs. It is split out from Get so the
// credential dispatch (the security-critical part) is unit-testable without a Pulumi context.
// Unlike the AWS builders it needs no injectable exchange seam -- the provider plugin does the
// token exchange itself.
func buildProviderInputs(gcpProviderConfig *gcpprovider.GcpProviderConfig) (*gcp.ProviderArgs, error) {
	// No config -> ambient Application Default Credentials chain.
	if gcpProviderConfig == nil {
		return &gcp.ProviderArgs{}, nil
	}

	switch {
	case gcpProviderConfig.GetWebIdentity() != nil:
		// Web identity is the deliberate mode switch: it wins even if a stale
		// service_account_key is still present on the config.
		webIdentity := gcpProviderConfig.GetWebIdentity()
		if webIdentity.GetWebIdentityToken() == "" {
			return nil, errors.New("web_identity is set but web_identity_token is empty")
		}
		if webIdentity.GetAudience() == "" {
			return nil, errors.New("web_identity is set but audience is empty")
		}
		if webIdentity.GetServiceAccountEmail() == "" {
			return nil, errors.New("web_identity is set but service_account_email is empty")
		}

		// The provider plugin exchanges the inline token at GCP STS and impersonates the
		// service account. The audience is passed through verbatim: it must stay
		// byte-identical to the token's `aud` claim and the pool provider's allowed
		// audiences, or GCP denies the exchange. The SDK does NOT auto-secret-wrap
		// identity_token (only the plain AccessToken field), so wrap it here to keep the
		// minted JWT out of plaintext Pulumi state.
		return &gcp.ProviderArgs{
			ExternalCredentials: &gcp.ProviderExternalCredentialsArgs{
				Audience:            pulumi.String(webIdentity.GetAudience()),
				IdentityToken:       pulumi.ToSecret(pulumi.String(webIdentity.GetWebIdentityToken())).(pulumi.StringOutput),
				ServiceAccountEmail: pulumi.String(webIdentity.GetServiceAccountEmail()),
			},
		}, nil

	case gcpProviderConfig.GetAccessToken() != "":
		// A pre-minted short-lived OAuth2 access token. Like web_identity above, an
		// explicitly supplied short-lived credential wins over a stale service_account_key.
		// AccessToken is the one field the SDK's NewProvider auto-secret-wraps (and registers
		// as an additional secret output), so it is passed plain here -- wrapping it again
		// would be redundant, not wrong.
		return &gcp.ProviderArgs{
			AccessToken: pulumi.String(gcpProviderConfig.GetAccessToken()),
		}, nil

	case gcpProviderConfig.GetServiceAccountKey() != "":
		// Static service-account key JSON.
		if err := validateServiceAccountKey(gcpProviderConfig.GetServiceAccountKey()); err != nil {
			return nil, err
		}
		return &gcp.ProviderArgs{
			Credentials: pulumi.String(gcpProviderConfig.GetServiceAccountKey()),
		}, nil

	default:
		// No explicit credential: the provider resolves credentials from the ambient
		// Application Default Credentials chain.
		return &gcp.ProviderArgs{}, nil
	}
}

// validateServiceAccountKey fails fast on malformed service-account key JSON so the error
// surfaces as a clear message instead of an opaque provider-plugin authentication failure.
func validateServiceAccountKey(serviceAccountKey string) error {
	var serviceAccountKeyMap map[string]interface{}
	if err := json.Unmarshal([]byte(serviceAccountKey), &serviceAccountKeyMap); err != nil {
		return errors.Wrap(err, "failed to parse service account key JSON. "+
			"Ensure the value is a valid GCP Service Account key file containing fields: "+
			"type, project_id, private_key_id, private_key, client_email, client_id, auth_uri, token_uri")
	}

	requiredFields := []string{"type", "project_id", "private_key", "client_email"}
	for _, field := range requiredFields {
		if _, ok := serviceAccountKeyMap[field]; !ok {
			return errors.Errorf("service account key JSON is missing required field: %s", field)
		}
	}

	privateKey, ok := serviceAccountKeyMap["private_key"].(string)
	if !ok {
		return errors.New("service account key 'private_key' field must be a string")
	}
	if len(privateKey) > 11 && privateKey[:11] != "-----BEGIN " {
		return errors.New("service account key 'private_key' field must be a PEM-encoded key " +
			"(starting with '-----BEGIN PRIVATE KEY-----'). " +
			"Ensure you're using a JSON key file from GCP, not a P12/PKCS12 key")
	}

	return nil
}

// ProviderResourceName returns the Pulumi resource name for the google provider.
//
// The base is intentionally "google": every GCP module historically created its provider with
// exactly this name. Pulumi tracks providers by resource name, so keeping it stable lets
// existing stacks keep their provider identity -- renaming it would trigger a provider
// replacement and the resource churn that follows in already-provisioned stacks. Do not rename
// without a state-migration plan.
func ProviderResourceName(suffixes []string) string {
	name := "google"
	for _, s := range suffixes {
		name = fmt.Sprintf("%s-%s", name, s)
	}
	return name
}

func PulumiOutputName(r interface{}, name string, suffixes ...string) string {
	outputName := fmt.Sprintf("gcp_%s", pulumioutput.Name(reflect.TypeOf(r), name))
	for _, s := range suffixes {
		outputName = fmt.Sprintf("%s_%s", outputName, s)
	}
	return outputName
}
