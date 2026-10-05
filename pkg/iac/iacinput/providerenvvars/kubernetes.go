package providerenvvars

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/pkg/kubernetes/kubeconfig"
)

// loadKubernetesEnvVars renders the connection's kubeconfig and exports its path to
// the engines. All provider-specific shaping lives in the shared builder
// (pkg/kubernetes/kubeconfig) so this path and the pulumi provider getter cannot
// drift; the exec credential command in the rendered kubeconfig is this very binary
// (the engine re-invokes it whenever a short-lived EKS/GKE token expires).
func loadKubernetesEnvVars(providerConfigYaml []byte, fileCacheLoc string) (map[string]string, error) {
	config := new(kubernetesprovider.KubernetesProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Kubernetes provider config")
	}

	executable, err := os.Executable()
	if err != nil {
		return nil, errors.Wrap(err, "resolving this binary's path for the kubeconfig exec credential")
	}

	kubeConfig, err := kubeconfig.Build(config, executable)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build kubeconfig for provider %v", config.Provider)
	}

	// 0600: the kubeconfig can carry credential material in its exec env entries
	// (same discipline as the iac-input file).
	kubeConfigPath := filepath.Join(fileCacheLoc, uuid.New().String())
	if err := os.WriteFile(kubeConfigPath, []byte(kubeConfig), 0600); err != nil {
		return nil, errors.Wrap(err, "failed to write kube-config to file")
	}

	// The Pulumi and Terraform/OpenTofu Kubernetes providers read different env vars
	// to locate a kubeconfig file: Pulumi honors KUBECONFIG, while the Terraform/OpenTofu
	// hashicorp/kubernetes (and helm) provider honors KUBE_CONFIG_PATH. Both names point
	// at the same generated kubeconfig so either engine resolves the connection; setting
	// the name the active engine ignores is harmless. Omitting KUBE_CONFIG_PATH makes the
	// tofu provider silently fall back to in-cluster auth (the runner pod's own service
	// account), which is the wrong cluster.
	envVars := map[string]string{
		"KUBECONFIG":       kubeConfigPath,
		"KUBE_CONFIG_PATH": kubeConfigPath,
	}

	return envVars, nil
}

// loadHostKubernetesEnvVars is the local-workflow counterpart of
// loadKubernetesEnvVars: no Planton connection, so the operator's own
// kubeconfig, resolved the way kubectl resolves it. The same file list is
// exported under every name an engine reads: KUBECONFIG for Pulumi and
// kubectl-style tooling; KUBE_CONFIG_PATH (one file) or KUBE_CONFIG_PATHS (a
// list) for the Terraform kubernetes and helm providers, which never read
// KUBECONFIG themselves; KUBE_CTX for the context every engine honours. A
// missing kubeconfig is a three-part refusal from the resolver.
func loadHostKubernetesEnvVars(kubeContext string) (map[string]string, error) {
	paths, err := kubeconfig.HostKubeconfigPaths()
	if err != nil {
		return nil, err
	}

	joined := strings.Join(paths, string(os.PathListSeparator))
	envVars := map[string]string{
		"KUBECONFIG": joined,
	}
	if len(paths) == 1 {
		envVars["KUBE_CONFIG_PATH"] = paths[0]
	} else {
		envVars["KUBE_CONFIG_PATHS"] = joined
	}
	if kubeContext != "" {
		envVars[kubeconfig.KubeContextEnvVar] = kubeContext
	}
	return envVars, nil
}
