package pulumicloudflareprovider

import (
	"fmt"
	"reflect"

	cloudflareprovider "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/pulumi/pulumioutput"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Get builds a pulumi-cloudflare Provider using the supplied credential.
// It supports both API Token and Legacy API Key authentication schemes.
// If the credential is nil, Pulumi's provider will fall back to environment
// variables (CLOUDFLARE_API_TOKEN, CLOUDFLARE_API_KEY, etc.).
func Get(
	ctx *pulumi.Context,
	cloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig,
	nameSuffixes ...string,
) (*cloudflare.Provider, error) {

	providerArgs := &cloudflare.ProviderArgs{}

	// Map credential fields when present; leave them nil to defer to env-vars.
	if cloudflareProviderConfig != nil {
		// Handle authentication based on scheme
		switch cloudflareProviderConfig.AuthScheme {
		case cloudflareprovider.CloudflareAuthScheme_api_token:
			if cloudflareProviderConfig.ApiToken != "" {
				providerArgs.ApiToken = pulumi.StringPtr(cloudflareProviderConfig.ApiToken)
			}
		case cloudflareprovider.CloudflareAuthScheme_legacy_api_key:
			if cloudflareProviderConfig.ApiKey != "" {
				providerArgs.ApiKey = pulumi.StringPtr(cloudflareProviderConfig.ApiKey)
			}
			if cloudflareProviderConfig.Email != "" {
				providerArgs.Email = pulumi.StringPtr(cloudflareProviderConfig.Email)
			}
		}
	}

	provider, err := cloudflare.NewProvider(
		ctx,
		ProviderResourceName(nameSuffixes),
		providerArgs,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare provider")
	}

	return provider, nil
}

// ProviderResourceName builds a deterministic Pulumi resource name such as
// "cloudflare-primary". Mirrors the google helper for naming consistency.
func ProviderResourceName(suffixes []string) string {
	name := "cloudflare"
	for _, s := range suffixes {
		name = fmt.Sprintf("%s-%s", name, s)
	}
	return name
}

// PulumiOutputName produces canonical output names (e.g. "cf_zone_id") to keep
// stack outputs predictable across modules.
func PulumiOutputName(r interface{}, name string, suffixes ...string) string {
	output := fmt.Sprintf("cf_%s", pulumioutput.Name(reflect.TypeOf(r), name))
	for _, s := range suffixes {
		output = fmt.Sprintf("%s_%s", output, s)
	}
	return output
}
