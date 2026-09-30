package providerdetect

import (
	"os"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/protobufyaml"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"google.golang.org/protobuf/proto"

	auth0provider "github.com/plantonhq/planton/catalog/auth0"
	awsprovider "github.com/plantonhq/planton/catalog/aws"
	azureprovider "github.com/plantonhq/planton/catalog/azure"
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	digitaloceanprovider "github.com/plantonhq/planton/catalog/digitalocean"
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	openfgaprovider "github.com/plantonhq/planton/catalog/openfga"
	stripeprovider "github.com/plantonhq/planton/catalog/stripe"
)

// ValidateProviderConfig validates that the provider config file can be loaded
// as the expected provider type.
func ValidateProviderConfig(providerConfigPath string, provider cloudresourcekind.CloudResourceProvider) error {
	// Read the provider config file
	configBytes, err := os.ReadFile(providerConfigPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read provider config file %s", providerConfigPath)
	}

	// Get the proto message for this provider
	protoMsg, err := ProviderConfigProto(provider)
	if err != nil {
		return err
	}

	// Try to load the config into the proto message
	if err := protobufyaml.LoadYamlBytes(configBytes, protoMsg); err != nil {
		return errors.Wrapf(err, "failed to parse provider config as %s config", ProviderDisplayName(provider))
	}

	return nil
}

// ProviderConfigProto returns a new provider-config proto message for the
// given provider -- the one provider->config-type map (the parity tooling's
// provider-config census and the `-p` validation both read it, so the two
// can never disagree on which type a provider's config is).
func ProviderConfigProto(provider cloudresourcekind.CloudResourceProvider) (proto.Message, error) {
	switch provider {
	case cloudresourcekind.CloudResourceProvider_auth0:
		return new(auth0provider.Auth0ProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_aws:
		return new(awsprovider.AwsProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_azure:
		return new(azureprovider.AzureProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_cloudflare:
		return new(cloudflareprovider.CloudflareProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_digital_ocean:
		return new(digitaloceanprovider.DigitalOceanProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_gcp:
		return new(gcpprovider.GcpProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_kubernetes:
		return new(kubernetesprovider.KubernetesProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_openfga:
		return new(openfgaprovider.OpenFgaProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_stripe:
		return new(stripeprovider.StripeProviderConfig), nil
	default:
		return nil, errors.Errorf("unsupported provider: %s", provider.String())
	}
}

// LoadProviderConfigBytes reads a provider config file and returns its contents.
func LoadProviderConfigBytes(providerConfigPath string) ([]byte, error) {
	configBytes, err := os.ReadFile(providerConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read provider config file %s", providerConfigPath)
	}
	return configBytes, nil
}
