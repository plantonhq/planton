package providerdetect

import (
	"os"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/protobufyaml"
	"google.golang.org/protobuf/proto"

	alicloudprovider "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud"
	atlasprovider "github.com/plantonhq/planton/apis/dev/planton/provider/atlas"
	auth0provider "github.com/plantonhq/planton/apis/dev/planton/provider/auth0"
	awsprovider "github.com/plantonhq/planton/apis/dev/planton/provider/aws"
	azureprovider "github.com/plantonhq/planton/apis/dev/planton/provider/azure"
	cloudflareprovider "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare"
	confluentprovider "github.com/plantonhq/planton/apis/dev/planton/provider/confluent"
	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	hetznercloudprovider "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud"
	kubernetesprovider "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	ociprovider "github.com/plantonhq/planton/apis/dev/planton/provider/oci"
	openfgaprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openfga"
	openstackprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openstack"
	scalewayprovider "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway"
	snowflakeprovider "github.com/plantonhq/planton/apis/dev/planton/provider/snowflake"
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
	protoMsg, err := getProviderConfigProto(provider)
	if err != nil {
		return err
	}

	// Try to load the config into the proto message
	if err := protobufyaml.LoadYamlBytes(configBytes, protoMsg); err != nil {
		return errors.Wrapf(err, "failed to parse provider config as %s config", ProviderDisplayName(provider))
	}

	return nil
}

// getProviderConfigProto returns a new proto message for the given provider.
func getProviderConfigProto(provider cloudresourcekind.CloudResourceProvider) (proto.Message, error) {
	switch provider {
	case cloudresourcekind.CloudResourceProvider_atlas:
		return new(atlasprovider.AtlasProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_auth0:
		return new(auth0provider.Auth0ProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_aws:
		return new(awsprovider.AwsProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_azure:
		return new(azureprovider.AzureProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_cloudflare:
		return new(cloudflareprovider.CloudflareProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_confluent:
		return new(confluentprovider.ConfluentProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_gcp:
		return new(gcpprovider.GcpProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_kubernetes:
		return new(kubernetesprovider.KubernetesProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_openfga:
		return new(openfgaprovider.OpenFgaProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_snowflake:
		return new(snowflakeprovider.SnowflakeProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_openstack:
		return new(openstackprovider.OpenStackProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_scaleway:
		return new(scalewayprovider.ScalewayProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_alicloud:
		return new(alicloudprovider.AliCloudProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_oci:
		return new(ociprovider.OciProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_hetzner_cloud:
		return new(hetznercloudprovider.HetznerCloudProviderConfig), nil
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
