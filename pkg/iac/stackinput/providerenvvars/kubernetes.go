package providerenvvars

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
)

const (
	gcpExecPluginPath = "/usr/local/bin/kube-client-go-gcp-exec-plugin"
)

// gcpExecPluginKubeConfigTemplate requires the following inputs for rendering a kubeconfig:
// 1. cluster endpoint ip
// 2. cluster cert-authority data
// 3. path to the exec plugin
// 4. base64 encoded google service account key
const gcpExecPluginKubeConfigTemplate = `apiVersion: v1
kind: Config
current-context: kube-context
contexts:
- name: kube-context
  context: {cluster: gke-cluster, user: kube-user}
clusters:
- name: gke-cluster
  cluster:
    server: https://%s
    certificate-authority-data: %s
users:
- name: kube-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1
      interactiveMode: Never
      command: %s
      args:
        - %s
`

// loadKubernetesEnvVars loads Kubernetes provider config and returns environment variables.
// It writes a kubeconfig file to the specified cache location.
func loadKubernetesEnvVars(providerConfigYaml []byte, fileCacheLoc string) (map[string]string, error) {
	config := new(kubernetesprovider.KubernetesProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Kubernetes provider config")
	}

	var kubeConfig string
	var err error

	switch config.Provider {
	case kubernetesprovider.KubernetesProvider_gcp_gke:
		kubeConfig, err = buildGcpGkeKubeConfig(config.GcpGke)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build kube-config for GCP GKE")
		}
	case kubernetesprovider.KubernetesProvider_aws_eks:
		kubeConfig, err = buildAwsEksKubeConfig(config.AwsEks)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build kube-config for AWS EKS")
		}
	case kubernetesprovider.KubernetesProvider_azure_aks:
		kubeConfig, err = buildAzureAksKubeConfig(config.AzureAks)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build kube-config for Azure AKS")
		}
	default:
		return nil, errors.Errorf("unsupported kubernetes provider: %v", config.Provider)
	}

	// Write kubeconfig to a file
	kubeConfigPath := filepath.Join(fileCacheLoc, uuid.New().String())
	if err := os.WriteFile(kubeConfigPath, []byte(kubeConfig), 0644); err != nil {
		return nil, errors.Wrap(err, "failed to write kube-config to file")
	}

	envVars := map[string]string{
		"KUBECONFIG": kubeConfigPath,
	}

	return envVars, nil
}

func buildGcpGkeKubeConfig(c *kubernetesprovider.KubernetesProviderConfigGcpGke) (string, error) {
	if c == nil {
		return "", errors.New("GCP GKE config is nil")
	}
	return fmt.Sprintf(gcpExecPluginKubeConfigTemplate,
		c.ClusterEndpoint,
		c.ClusterCaData,
		gcpExecPluginPath,
		base64.StdEncoding.EncodeToString([]byte(c.ServiceAccountKey))), nil
}

func buildAwsEksKubeConfig(c *kubernetesprovider.KubernetesProviderConfigAwsEks) (string, error) {
	// TODO: Implement AWS EKS kubeconfig generation
	return "", nil
}

func buildAzureAksKubeConfig(c *kubernetesprovider.KubernetesProviderConfigAzureAks) (string, error) {
	// TODO: Implement Azure AKS kubeconfig generation
	return "", nil
}
