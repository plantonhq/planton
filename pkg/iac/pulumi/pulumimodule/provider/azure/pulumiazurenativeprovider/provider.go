// Package pulumiazurenativeprovider is the convergent place where Azure pulumi-azure-native
// modules build their azurenative.Provider from the stack input's AzureProviderConfig. It mirrors
// pulumiazureprovider (the pulumi-azure "classic" builder) so a coding agent can learn both Azure
// credential-resolution paths from one shape.
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
// SECURITY -- the one structural difference from the classic builder: this azure-native (v3)
// SDK's NewProvider auto-secret-wraps ClientId/ClientSecret/ClientCertificatePassword but NOT
// OidcToken, so an unwrapped token would land in Pulumi state and `preview` output in plaintext.
// This builder therefore wraps the token with pulumi.ToSecret itself. The classic (v6) SDK
// auto-wraps OidcToken, so its builder deliberately does not.
//
// Like the classic builder, there is NO builder-side token exchange: the AWS builders exchange
// the web-identity token themselves only because pulumi-aws's provider-native path is broken
// upstream (pulumi-aws#6228); the pulumi Azure providers consume the inline token natively. Do
// not "converge" this builder toward the AWS workaround shape.
//
// It deliberately never passes empty-string credential fields to azurenative.NewProvider: the
// inline pre-builder modules passed all four fields unconditionally, which conflates "not set"
// with "set to empty" and blocks the ambient credential chain from ever engaging.
package pulumiazurenativeprovider

import (
	"fmt"
	"reflect"

	azureprovider "github.com/plantonhq/planton/apis/dev/planton/provider/azure"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/pulumi/pulumioutput"

	"github.com/pkg/errors"
	azurenative "github.com/pulumi/pulumi-azure-native-sdk/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Get builds an azure-native Provider from the given AzureProviderConfig. There is no region
// argument: the Azure provider is not region-scoped -- each resource carries its own Location.
// nameSuffixes disambiguate the provider resource name when a module needs more than one provider.
func Get(ctx *pulumi.Context, azureProviderConfig *azureprovider.AzureProviderConfig,
	nameSuffixes ...string) (*azurenative.Provider, error) {
	providerArgs, err := buildProviderArgs(azureProviderConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build azure-native provider args")
	}

	azureNativeProvider, err := azurenative.NewProvider(ctx, ProviderResourceName(nameSuffixes), providerArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create azure-native provider")
	}

	return azureNativeProvider, nil
}

// buildProviderArgs is the pure, side-effect-free core of the builder: it maps an
// AzureProviderConfig to azurenative.ProviderArgs. It is split out from Get so the credential
// dispatch (the security-critical part) is unit-testable without a Pulumi context.
func buildProviderArgs(azureProviderConfig *azureprovider.AzureProviderConfig) (*azurenative.ProviderArgs, error) {
	providerArgs := &azurenative.ProviderArgs{}

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
		// This SDK's NewProvider does NOT auto-secret-wrap OidcToken (unlike the classic SDK),
		// so wrap it here to keep the minted JWT out of plaintext Pulumi state.
		providerArgs.UseOidc = pulumi.Bool(true)
		providerArgs.OidcToken = pulumi.ToSecret(pulumi.String(webIdentity.GetWebIdentityToken())).(pulumi.StringInput)

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
func setIdentityCoordinates(providerArgs *azurenative.ProviderArgs, azureProviderConfig *azureprovider.AzureProviderConfig) {
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

// ProviderResourceName returns the Pulumi resource name for the azure-native provider.
//
// The base is intentionally "azure": every Azure module historically created its provider with
// exactly this name (azurenative.NewProvider(ctx, "azure", ...)). Pulumi tracks providers by
// resource name, so keeping it stable lets existing modules adopt this shared builder without
// triggering a provider replacement -- and the resource churn that would follow -- in
// already-provisioned stacks. Do not rename without a state-migration plan. (No module uses both
// the classic and native SDKs in one program, so the shared base name never collides.)
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
