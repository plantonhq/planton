package providerenvvars

import (
	"github.com/pkg/errors"
	awsprovider "github.com/plantonhq/planton/catalog/aws"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/shared/catalogkind"
	"gopkg.in/yaml.v3"
)

// ExtractAwsProviderConfig decodes the IaC input's provider_config into the typed
// AwsProviderConfig when (and only when) the target resource belongs to the AWS
// provider. Returns (nil, nil) for non-AWS targets and for IaC inputs carrying no
// provider config -- both are ordinary, not errors.
//
// This is the single iac-input -> typed-AWS-config decode point shared by the
// env-var loader above and the tofu provider-override writer (pkg/iac/tofu/tfoverride);
// consumers must never re-implement the target/kind/provider dispatch, so the two
// paths cannot drift on how a provider config is recognized.
func ExtractAwsProviderConfig(iacInputYaml string) (*awsprovider.AwsProviderConfig, error) {
	iacInputMap := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(iacInputYaml), &iacInputMap); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal IaC input yaml")
	}

	targetYaml, err := extractTargetYaml(iacInputMap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract target from IaC input")
	}

	kind, err := catalogkindreflect.ExtractKindFromYaml(targetYaml)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract catalog kind from target")
	}

	if catalogkindreflect.GetProvider(kind) != catalogkind.CatalogProvider_aws {
		return nil, nil
	}

	providerConfigYaml, hasProviderConfig := extractProviderConfigYaml(iacInputMap)
	if !hasProviderConfig {
		return nil, nil
	}

	config := new(awsprovider.AwsProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load AWS provider config")
	}
	return config, nil
}
