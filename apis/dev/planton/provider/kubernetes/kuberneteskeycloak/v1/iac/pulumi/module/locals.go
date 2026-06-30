package module

import (
	"fmt"

	kuberneteskeycloakv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteskeycloak/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	// Keycloak namespace
	Namespace string

	// Resource labels for all resources
	Labels map[string]string

	// Ingress configuration
	IngressEnabled bool
	Hostname       string

	// Keycloak service configuration
	ServiceName string
	ServicePort int

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	PasswordSecretName    string // Keycloak admin password secret
	DbPasswordSecretName  string // PostgreSQL database password secret
	ExternalLbServiceName string // External LoadBalancer service for ingress

	// Stack outputs
	PortForwardCommand string
	KubeEndpoint       string
	ExternalHostname   string
	InternalHostname   string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kuberneteskeycloakv1.KubernetesKeycloakStackInput) *Locals {
	locals := &Locals{}

	// Determine namespace from spec (required StringValueOrRef field)
	locals.Namespace = stackInput.Target.Spec.Namespace.GetValue()

	// Fallback to default pattern if empty
	if locals.Namespace == "" {
		locals.Namespace = "keycloak-" + stackInput.Target.Metadata.Name
	}

	// Set up labels
	locals.Labels = map[string]string{
		"app":      "keycloak",
		"resource": stackInput.Target.Metadata.Name,
	}

	if stackInput.Target.Metadata.Env != "" {
		locals.Labels["env"] = stackInput.Target.Metadata.Env
	}

	if stackInput.Target.Metadata.Org != "" {
		locals.Labels["org"] = stackInput.Target.Metadata.Org
	}

	// Ingress configuration
	if stackInput.Target.Spec != nil && stackInput.Target.Spec.Ingress != nil {
		locals.IngressEnabled = stackInput.Target.Spec.Ingress.Enabled
		locals.Hostname = stackInput.Target.Spec.Ingress.Hostname
	}

	target := stackInput.Target

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	// Users can prefix metadata.name with component type if needed (e.g., "keycloak-my-auth")
	locals.PasswordSecretName = fmt.Sprintf("%s-password", target.Metadata.Name)
	locals.DbPasswordSecretName = fmt.Sprintf("%s-db-password", target.Metadata.Name)
	locals.ExternalLbServiceName = fmt.Sprintf("%s-external-lb", target.Metadata.Name)

	// Service configuration
	// ServiceName uses just the metadata.name; Helm chart handles its own suffixes
	locals.ServiceName = target.Metadata.Name
	locals.ServicePort = 8080

	// Stack outputs
	locals.PortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s svc/%s 8080:8080", locals.Namespace, locals.ServiceName)
	locals.KubeEndpoint = fmt.Sprintf("%s.%s.svc.cluster.local:8080", locals.ServiceName, locals.Namespace)

	if locals.IngressEnabled && locals.Hostname != "" {
		locals.ExternalHostname = "https://" + locals.Hostname
		locals.InternalHostname = "https://internal-" + locals.Hostname
	}

	return locals
}
