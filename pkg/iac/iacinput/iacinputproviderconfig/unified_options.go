package iacinputproviderconfig

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/pkg/iac/iacinput/providerdetect"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/spf13/pflag"
)

// ProviderConfig holds the provider configuration from the unified --provider-config flag.
type ProviderConfig struct {
	// Path is the path to the provider config file
	Path string
	// Provider is the provider type detected from the manifest
	Provider catalogkind.CatalogProvider
}

// GetFromFlags extracts the provider config path from CLI flags.
func GetFromFlags(
	commandFlagSet *pflag.FlagSet,
	detectionResult *providerdetect.DetectionResult,
) (*ProviderConfig, error) {
	providerConfigPath, err := commandFlagSet.GetString(string(flag.ProviderConfig))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s flag", flag.ProviderConfig)
	}

	return &ProviderConfig{
		Path:     providerConfigPath,
		Provider: detectionResult.Provider,
	}, nil
}

// GetFromFlagsSimple extracts the provider config path from CLI flags without detection.
// This is a convenience function for commands that don't perform provider detection.
func GetFromFlagsSimple(commandFlagSet *pflag.FlagSet) (*ProviderConfig, error) {
	providerConfigPath, err := commandFlagSet.GetString(string(flag.ProviderConfig))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s flag", flag.ProviderConfig)
	}

	return &ProviderConfig{
		Path:     providerConfigPath,
		Provider: catalogkind.CatalogProvider_catalog_provider_unspecified,
	}, nil
}
