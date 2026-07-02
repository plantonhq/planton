// Package pulumiazureprovider is the single, convergent place where every Azure pulumi-azure
// "classic" module builds its azure.Provider from the stack input's AzureProviderConfig. It
// mirrors the sibling per-cloud builders (e.g. pulumiazurenativeprovider, pulumiawsprovider,
// pulumigoogleprovider) so a coding agent can learn the Azure credential-resolution path by
// reading one file.
//
// It dispatches on which fields of AzureProviderConfig are populated, supporting every auth mode
// with a single seam:
//   - web_identity set  -> keyless OIDC federation. The minted JWT is handed to the provider
//     inline (UseOidc + OidcToken) together with the identity coordinates; the provider plugin
//     performs the AAD client-assertion exchange itself, out of our process.
//   - client_secret set -> static service-principal credentials (today's four-field mode).
//   - neither           -> identity coordinates only when present, no credential. The provider
//     falls back to the SDK's ambient credential chain (e.g. a self-hosted runner's managed
//     identity or `az` CLI login).
//
// Why NO builder-side token exchange (deliberate contrast with the AWS builders): the AWS
// builders exchange the web-identity token themselves only because pulumi-aws's provider-native
// path is broken upstream (pulumi-aws#6228). The pulumi-azure providers consume the inline token
// natively -- the plugin sends it as the OAuth client_assertion and gets the AAD access token --
// so a builder-side exchange here would add a dependency and put credentials in our process for
// zero benefit. Do not "converge" this builder toward the AWS workaround shape.
//
// Freshness: the inline token is consumed at provider configure time and the resulting AAD access
// token is provider-managed for the run. Each pulumi operation re-runs the module program with a
// freshly minted JWT, so the exchange always sees a fresh token whose validity covers that one
// operation. No token is ever written to disk.
//
// Secrecy: this classic (v6) SDK's NewProvider auto-secret-wraps OidcToken (along with ClientId,
// ClientSecret, etc.), so the builder passes the plain string and must NOT double-wrap it. The
// azure-native (v3) SDK does NOT auto-wrap OidcToken -- that difference is owned by the sibling
// pulumiazurenativeprovider, which wraps it explicitly.
//
// It deliberately never passes empty-string credential fields to azure.NewProvider: the inline
// pre-builder modules passed all four fields unconditionally, which conflates "not set" with
// "set to empty" and blocks the ambient credential chain from ever engaging.
package pulumiazureprovider

import (
	"fmt"
	"reflect"

	azureprovider "github.com/plantonhq/planton/apis/dev/planton/provider/azure"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/pulumi/pulumioutput"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Get builds an azure.Provider from the given AzureProviderConfig. There is no region argument:
// unlike AWS, the Azure provider is not region-scoped -- each resource carries its own Location.
// nameSuffixes disambiguate the provider resource name when a module needs more than one provider.
func Get(ctx *pulumi.Context, azureProviderConfig *azureprovider.AzureProviderConfig,
	nameSuffixes ...string) (*azure.Provider, error) {
	providerArgs, err := buildProviderArgs(azureProviderConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build azure provider args")
	}

	azureProvider, err := azure.NewProvider(ctx, ProviderResourceName(nameSuffixes), providerArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create azure provider")
	}

	return azureProvider, nil
}

// buildProviderArgs is the pure, side-effect-free core of the builder: it maps an
// AzureProviderConfig to azure.ProviderArgs. It is split out from Get so the credential dispatch
// (the security-critical part) is unit-testable without a Pulumi context. Unlike the AWS builders
// it needs no injectable exchange seam -- the provider plugin does the token exchange itself.
func buildProviderArgs(azureProviderConfig *azureprovider.AzureProviderConfig) (*azure.ProviderArgs, error) {
	providerArgs := &azure.ProviderArgs{}

	// No config -> ambient credential chain.
	if azureProviderConfig == nil {
		return providerArgs, nil
	}

	setIdentityCoordinates(providerArgs, azureProviderConfig)

	switch {
	case azureProviderConfig.GetWebIdentity() != nil:
		webIdentity := azureProviderConfig.GetWebIdentity()
		if webIdentity.GetWebIdentityToken() == "" {
			return nil, errors.New("web_identity is set but web_identity_token is empty")
		}

		// The provider plugin exchanges the inline token via the AAD client-assertion flow.
		// This SDK's NewProvider auto-secret-wraps OidcToken, so pass the plain string here.
		providerArgs.UseOidc = pulumi.Bool(true)
		providerArgs.OidcToken = pulumi.String(webIdentity.GetWebIdentityToken())

	case azureProviderConfig.GetClientSecret() != "":
		// Static service-principal credentials.
		providerArgs.ClientSecret = pulumi.String(azureProviderConfig.GetClientSecret())

	default:
		// No explicit credential: the provider resolves credentials from the SDK's ambient
		// chain (e.g. a self-hosted runner's managed identity or `az` CLI login).
	}

	return providerArgs, nil
}

// setIdentityCoordinates copies the non-credential identity fields (client, tenant,
// subscription), skipping empty values so an absent field never reaches the provider as an
// empty string.
func setIdentityCoordinates(providerArgs *azure.ProviderArgs, azureProviderConfig *azureprovider.AzureProviderConfig) {
	if azureProviderConfig.GetClientId() != "" {
		providerArgs.ClientId = pulumi.String(azureProviderConfig.GetClientId())
	}
	if azureProviderConfig.GetTenantId() != "" {
		providerArgs.TenantId = pulumi.String(azureProviderConfig.GetTenantId())
	}
	if azureProviderConfig.GetSubscriptionId() != "" {
		providerArgs.SubscriptionId = pulumi.String(azureProviderConfig.GetSubscriptionId())
	}
}

// ProviderResourceName returns the Pulumi resource name for the Azure provider.
//
// The base is intentionally "azure": every Azure module historically created its provider with
// exactly this name (azure.NewProvider(ctx, "azure", ...)). Pulumi tracks providers by resource
// name, so keeping it stable lets existing modules adopt this shared builder without triggering
// a provider replacement -- and the resource churn that would follow -- in already-provisioned
// stacks. Do not rename without a state-migration plan.
func ProviderResourceName(suffixes []string) string {
	name := "azure"
	for _, s := range suffixes {
		name = fmt.Sprintf("%s-%s", name, s)
	}
	return name
}

// PulumiOutputName builds a stable, prefixed output name for Azure resources, mirroring the
// helper exposed by the other per-cloud provider builders.
func PulumiOutputName(r interface{}, name string, suffixes ...string) string {
	outputName := fmt.Sprintf("azure_%s", pulumioutput.Name(reflect.TypeOf(r), name))
	for _, s := range suffixes {
		outputName = fmt.Sprintf("%s_%s", outputName, s)
	}
	return outputName
}
