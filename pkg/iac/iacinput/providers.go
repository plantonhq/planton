package iacinput

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/iacinput/iacinputproviderconfig"
	"github.com/plantonhq/planton/pkg/protobufyaml"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

// ProviderConfigKey is the key used in IaC input for provider configuration.
const ProviderConfigKey = "provider_config"

// addProviderConfig adds the provider config to the IaC input map.
// It reads the provider config file and unmarshals it into the IaC input.
func addProviderConfig(
	iacInputContentMap map[string]interface{},
	providerConfig *iacinputproviderconfig.ProviderConfig,
) (map[string]interface{}, error) {
	if providerConfig == nil || providerConfig.Path == "" {
		return iacInputContentMap, nil
	}

	providerConfigContent, err := os.ReadFile(providerConfig.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read provider config file: %s", providerConfig.Path)
	}

	providerConfigJson, err := protobufyaml.YAMLToJSON(providerConfigContent)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal provider config file")
	}
	var providerConfigContentMap map[string]interface{}
	if err := json.Unmarshal(providerConfigJson, &providerConfigContentMap); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal provider config file")
	}

	iacInputContentMap[ProviderConfigKey] = providerConfigContentMap
	return iacInputContentMap, nil
}

// LoadProviderConfig loads a provider config from the IaC input map into a proto message.
func LoadProviderConfig(
	iacInputContentMap map[string]interface{},
	providerConfigKey string,
	providerConfigObject proto.Message,
) (isProviderConfigLoaded bool, err error) {
	providerConfigYaml, ok := iacInputContentMap[providerConfigKey]
	if !ok {
		return false, nil
	}

	providerConfigBytes, err := yaml.Marshal(providerConfigYaml)
	if err != nil {
		return false, errors.Wrap(err, "failed to marshal provider config yaml content")
	}

	if err := protobufyaml.LoadYamlBytes(providerConfigBytes, providerConfigObject); err != nil {
		return false, errors.Wrap(err, "failed to load yaml bytes into provider config")
	}

	return true, nil
}
