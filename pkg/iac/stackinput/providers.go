package stackinput

import (
	"os"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/stackinput/stackinputproviderconfig"
	"github.com/plantonhq/planton/pkg/protobufyaml"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
	sigsyaml "sigs.k8s.io/yaml"
)

// ProviderConfigKey is the key used in stack input for provider configuration.
const ProviderConfigKey = "provider_config"

// addProviderConfig adds the provider config to the stack input map.
// It reads the provider config file and unmarshals it into the stack input.
func addProviderConfig(
	stackInputContentMap map[string]interface{},
	providerConfig *stackinputproviderconfig.ProviderConfig,
) (map[string]interface{}, error) {
	if providerConfig == nil || providerConfig.Path == "" {
		return stackInputContentMap, nil
	}

	providerConfigContent, err := os.ReadFile(providerConfig.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read provider config file: %s", providerConfig.Path)
	}

	var providerConfigContentMap map[string]interface{}
	if err := sigsyaml.Unmarshal(providerConfigContent, &providerConfigContentMap); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal provider config file")
	}

	stackInputContentMap[ProviderConfigKey] = providerConfigContentMap
	return stackInputContentMap, nil
}

// LoadProviderConfig loads a provider config from the stack input map into a proto message.
func LoadProviderConfig(
	stackInputContentMap map[string]interface{},
	providerConfigKey string,
	providerConfigObject proto.Message,
) (isProviderConfigLoaded bool, err error) {
	providerConfigYaml, ok := stackInputContentMap[providerConfigKey]
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
