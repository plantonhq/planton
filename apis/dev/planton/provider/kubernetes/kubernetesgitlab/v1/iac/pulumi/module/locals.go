package module

import (
	"fmt"

	kubernetesgitlabv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgitlab/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

// Locals holds computed values for the GitLab deployment
type Locals struct {
	Namespace       string
	Labels          map[string]string
	ServiceName     string
	ServiceFQDN     string
	PortForwardCmd  string
	IngressHostname string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	// Users can prefix metadata.name with component type if needed (e.g., "gitlab-prod")
	IngressName   string
	TlsSecretName string
}

// initializeLocals creates and initializes local values from the stack input
func initializeLocals(stackInput *kubernetesgitlabv1.KubernetesGitlabStackInput) *Locals {
	l := &Locals{}

	target := stackInput.Target

	// Extract namespace from spec (required field)
	l.Namespace = target.Spec.Namespace.GetValue()

	// Build standard labels
	l.Labels = getLabels(target.Metadata)

	// Service configuration
	l.ServiceName = target.Metadata.Name
	l.ServiceFQDN = fmt.Sprintf("%s.%s.svc.cluster.local", l.ServiceName, l.Namespace)
	l.PortForwardCmd = fmt.Sprintf("kubectl port-forward -n %s service/%s 8080:80", l.Namespace, l.ServiceName)

	// Ingress configuration (if enabled)
	if target.Spec.Ingress != nil && target.Spec.Ingress.Enabled && target.Spec.Ingress.Hostname != "" {
		l.IngressHostname = target.Spec.Ingress.Hostname
	}

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	l.IngressName = fmt.Sprintf("%s-ingress", target.Metadata.Name)
	l.TlsSecretName = fmt.Sprintf("%s-tls", target.Metadata.Name)

	return l
}

// getLabels returns the standard labels for GitLab resources
func getLabels(metadata *shared.CloudResourceMetadata) map[string]string {
	labels := map[string]string{
		"resource":      "true",
		"resource_kind": "kubernetes_gitlab",
	}

	if metadata.Id != "" {
		labels["resource_id"] = metadata.Id
	} else {
		labels["resource_id"] = metadata.Name
	}

	if metadata.Org != "" {
		labels["organization"] = metadata.Org
	}

	if metadata.Env != "" {
		labels["environment"] = metadata.Env
	}

	return labels
}
