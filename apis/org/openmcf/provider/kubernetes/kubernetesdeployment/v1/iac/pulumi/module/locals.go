package module

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kubernetesdeploymentv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesdeployment/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/openmcf/pkg/kubernetes/kuberneteslabels"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	IngressCertClusterIssuerName string
	IngressCertSecretName        string
	IngressExternalHostname      string
	IngressHostnames             []string
	IngressInternalHostname      string
	KubePortForwardCommand       string
	KubeServiceFqdn              string
	KubeServiceName              string
	Namespace                    string
	KubernetesDeployment         *kubernetesdeploymentv1.KubernetesDeployment
	ImagePullSecretData          map[string]string
	Labels                       map[string]string
	SelectorLabels               map[string]string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	EnvSecretName                 string
	ImagePullSecretName           string
	IngressCertificateName        string
	ExternalGatewayName           string
	InternalGatewayName           string
	HttpExternalRedirectRouteName string
	HttpsExternalRouteName        string
	HttpInternalRedirectRouteName string
	HttpsInternalRouteName        string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesdeploymentv1.KubernetesDeploymentStackInput) (*Locals, error) {
	locals := &Locals{}

	locals.KubernetesDeployment = stackInput.Target

	target := stackInput.Target

	// Selector labels use metadata.name to ensure each deployment's pods are uniquely identified.
	// This prevents traffic routing conflicts when multiple deployments share a namespace.
	locals.SelectorLabels = map[string]string{
		"app":                            target.Metadata.Name,
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
	}

	// Full labels include both selector labels and additional metadata labels
	locals.Labels = map[string]string{
		"app":                            target.Metadata.Name, // Include selector label
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesDeployment.String(),
	}

	if target.Metadata.Id != "" {
		locals.Labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.Labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.Labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// get namespace from spec, it is required field
	locals.Namespace = target.Spec.Namespace.GetValue()

	// export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	// Users can prefix metadata.name with component type if needed (e.g., "deploy-my-app")
	locals.EnvSecretName = fmt.Sprintf("%s-env-secrets", target.Metadata.Name)
	locals.ImagePullSecretName = fmt.Sprintf("%s-image-pull", target.Metadata.Name)
	locals.IngressCertificateName = fmt.Sprintf("%s-ingress-cert", target.Metadata.Name)
	locals.ExternalGatewayName = fmt.Sprintf("%s-external", target.Metadata.Name)
	locals.InternalGatewayName = fmt.Sprintf("%s-internal", target.Metadata.Name)
	locals.HttpExternalRedirectRouteName = fmt.Sprintf("%s-http-external-redirect", target.Metadata.Name)
	locals.HttpsExternalRouteName = fmt.Sprintf("%s-https-external", target.Metadata.Name)
	locals.HttpInternalRedirectRouteName = fmt.Sprintf("%s-http-internal-redirect", target.Metadata.Name)
	locals.HttpsInternalRouteName = fmt.Sprintf("%s-https-internal", target.Metadata.Name)

	// Priority 1: StackInput (used by Planton - takes precedence)
	// If present, use it and don't check the label at all
	if stackInput.DockerConfigJson != "" {
		locals.ImagePullSecretData = map[string]string{".dockerconfigjson": stackInput.DockerConfigJson}
	} else {
		// Priority 2: Label with file path (for open-source users)
		// Only checked if stackInput.DockerConfigJson is empty
		if dockerConfigFilePath := target.Metadata.Labels[kuberneteslabels.DockerConfigJsonFileLabelKey]; dockerConfigFilePath != "" {
			dockerConfigJson, err := loadDockerConfigFromFile(dockerConfigFilePath)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to load docker config from file specified in label: %s", dockerConfigFilePath)
			}
			locals.ImagePullSecretData = map[string]string{".dockerconfigjson": dockerConfigJson}
		}
		// Priority 3: If neither set, ImagePullSecretData remains nil (no image pull secret)
	}

	// Use metadata.name for service name to avoid conflicts when multiple deployments share a namespace
	locals.KubeServiceName = target.Metadata.Name

	//export kubernetes service name
	ctx.Export(OpService, pulumi.String(locals.KubeServiceName))

	locals.KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.KubeServiceName, locals.Namespace)

	//export kubernetes endpoint
	ctx.Export(OpKubeEndpoint, pulumi.String(locals.KubeServiceFqdn))

	locals.KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s 8080:8080",
		locals.Namespace, locals.KubeServiceName)

	//export kube-port-forward command
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.KubePortForwardCommand))

	if locals.KubernetesDeployment.Spec.Availability == nil {
		locals.KubernetesDeployment.Spec.Availability = &kubernetesdeploymentv1.KubernetesDeploymentAvailability{
			MinReplicas: 1,
		}
	}

	if target.Spec.Ingress == nil ||
		!target.Spec.Ingress.Enabled ||
		target.Spec.Ingress.Hostname == "" {
		return locals, nil
	}

	// Use the hostname directly from spec
	locals.IngressExternalHostname = target.Spec.Ingress.Hostname

	// Internal hostname (private ingress) - prepend internal-
	locals.IngressInternalHostname = fmt.Sprintf("internal-%s", target.Spec.Ingress.Hostname)

	locals.IngressHostnames = []string{
		locals.IngressExternalHostname,
		locals.IngressInternalHostname,
	}

	//export ingress hostnames
	ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))
	ctx.Export(OpInternalHostname, pulumi.String(locals.IngressInternalHostname))

	//note: a ClusterIssuer resource should have already exist on the kubernetes-cluster.
	//this is typically taken care of by the kubernetes cluster administrator.
	//if the kubernetes-cluster is created using Planton, then the cluster-issuer name will be
	//same as the ingress-domain-name as long as the same ingress-domain-name is added to the list of
	//ingress-domain-names for the GkeCluster/EksCluster/AksCluster spec.
	// Extract the domain from hostname for certificate issuer name
	dnsDomain := extractDomainFromHostname(target.Spec.Ingress.Hostname)
	locals.IngressCertClusterIssuerName = dnsDomain

	locals.IngressCertSecretName = locals.Namespace

	if locals.KubernetesDeployment.Spec.Container.App.Image == nil {
		return nil, errors.New("spec.container.app.image is required")
	}

	return locals, nil
}

// loadDockerConfigFromFile reads docker config JSON from the specified file path.
// Returns error if file doesn't exist or can't be read.
func loadDockerConfigFromFile(filePath string) (string, error) {
	// Expand ~ to home directory if present
	if strings.HasPrefix(filePath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "failed to get user home directory")
		}
		filePath = filepath.Join(homeDir, filePath[2:])
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", errors.Errorf("docker config file does not exist: %s", filePath)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read docker config file: %s", filePath)
	}

	// Validate it's not empty
	if len(content) == 0 {
		return "", errors.Errorf("docker config file is empty: %s", filePath)
	}

	// Optional: Basic JSON validation
	var js json.RawMessage
	if err := json.Unmarshal(content, &js); err != nil {
		return "", errors.Wrapf(err, "docker config file contains invalid JSON: %s", filePath)
	}

	return string(content), nil
}

// extractDomainFromHostname extracts the domain from a hostname
// Example: "myapp.example.com" -> "example.com"
func extractDomainFromHostname(hostname string) string {
	// Split by dots and take everything after the first part
	// This is a simple implementation - assumes standard domain structure
	parts := []rune(hostname)
	firstDotIndex := -1
	for i, char := range parts {
		if char == '.' {
			firstDotIndex = i
			break
		}
	}
	if firstDotIndex > 0 && firstDotIndex < len(hostname)-1 {
		return hostname[firstDotIndex+1:]
	}
	// If no dot found or dot is at the end, return the hostname as-is
	return hostname
}
