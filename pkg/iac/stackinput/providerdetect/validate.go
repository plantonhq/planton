package providerdetect

import (
	"os"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/protobufyaml"
	"google.golang.org/protobuf/proto"

	alicloudprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud"
	atlasprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/atlas"
	auth0provider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/auth0"
	awsprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws"
	azureprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure"
	cloudflareprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare"
	confluentprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/confluent"
	gcpprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
	kubernetesprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	ociprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci"
	openfgaprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openfga"
	openstackprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	scalewayprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway"
	snowflakeprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/snowflake"
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
	case cloudresourcekind.CloudResourceProvider_open_fga:
		return new(openfgaprovider.OpenFgaProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_snowflake:
		return new(snowflakeprovider.SnowflakeProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_openstack:
		return new(openstackprovider.OpenStackProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_scaleway:
		return new(scalewayprovider.ScalewayProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_alicloud:
		return new(alicloudprovider.AlicloudProviderConfig), nil
	case cloudresourcekind.CloudResourceProvider_oci:
		return new(ociprovider.OciProviderConfig), nil
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
