package tofumodule

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/stackinput/providerenvvars"
)

// GetProviderConfigEnvVars returns provider-specific environment variables for the stack.
// It delegates to the IaC-agnostic providerenvvars package which determines the correct provider
// based on the target's api_version/kind and loads only the relevant provider configuration.
func GetProviderConfigEnvVars(stackInputYaml, fileCacheLoc, kubeContext string) ([]string, error) {
	// EngineReadsEnvironment is named here because this is the tofu/terraform execution boundary:
	// the catalog modules' provider blocks are empty, so a keyless connection's credential must
	// arrive as environment variables (exchanged for AWS and Google Cloud, the federated token for
	// Azure). The pulumi path calls GetEnvVarsWithOptions directly and names EngineBuildsProviders.
	// KubeContext rides the same options: the loader exports it as KUBE_CTX
	// beside the kubeconfig it resolves (a connection's rendered file, or the
	// operator's own kubeconfig when there is no connection), so one seam owns
	// every name the kubernetes and helm providers read.
	providerConfigEnvVars, err := providerenvvars.GetEnvVarsWithOptions(stackInputYaml, providerenvvars.Options{
		FileCacheLoc: fileCacheLoc,
		Engine:       providerenvvars.EngineReadsEnvironment,
		KubeContext:  kubeContext,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get provider env vars from stack input")
	}

	return mapToSlice(providerConfigEnvVars), nil
}

// mapToSlice converts a map of string to string into a slice of string slices by joining key-value pairs with an equals sign.
func mapToSlice(inputMap map[string]string) []string {
	var result []string
	for key, value := range inputMap {
		result = append(result, key+"="+value)
	}
	return result
}
