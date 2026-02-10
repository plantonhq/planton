package module

import (
	"fmt"
	"strings"

	kubernetesservicev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesservice/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the stack input for use across the module.
type Locals struct {
	Context   *pulumi.Context
	Spec      *kubernetesservicev1.KubernetesServiceSpec
	Namespace string
	Name      string
	Labels    map[string]string
	Annotations map[string]string
	Selector  map[string]string
	Ports     []*kubernetesservicev1.KubernetesServicePort
	ServiceType string
	Headless    bool
	ExternalDnsName string
	ExternalTrafficPolicy string
	SessionAffinity       string
	LoadBalancerSourceRanges []string
}

// initializeLocals extracts and transforms spec fields into module-local values.
func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesservicev1.KubernetesServiceStackInput) *Locals {
	spec := stackInput.Target.Spec

	// Build combined labels: standard labels merged with user-provided labels.
	labels := map[string]string{
		"managed-by":    "openmcf",
		"resource":      stackInput.Target.Metadata.GetName(),
		"resource-kind": "KubernetesService",
	}
	for k, v := range spec.GetLabels() {
		labels[k] = v
	}

	// Build annotations from user-provided annotations.
	annotations := make(map[string]string)
	for k, v := range spec.GetAnnotations() {
		annotations[k] = v
	}

	// Resolve service type. Use getter since it's an optional field.
	serviceType := resolveServiceType(spec.GetType())

	// Resolve external traffic policy.
	externalTrafficPolicy := resolveExternalTrafficPolicy(spec.GetExternalTrafficPolicy())

	// Resolve session affinity.
	sessionAffinity := resolveSessionAffinity(spec.GetSessionAffinity())

	return &Locals{
		Context:                  ctx,
		Spec:                     spec,
		Namespace:                spec.GetNamespace(),
		Name:                     spec.GetName(),
		Labels:                   labels,
		Annotations:              annotations,
		Selector:                 spec.GetSelector(),
		Ports:                    spec.GetPorts(),
		ServiceType:              serviceType,
		Headless:                 spec.GetHeadless(),
		ExternalDnsName:          spec.GetExternalDnsName(),
		ExternalTrafficPolicy:    externalTrafficPolicy,
		SessionAffinity:          sessionAffinity,
		LoadBalancerSourceRanges: spec.GetLoadBalancerSourceRanges(),
	}
}

// resolveServiceType maps the protobuf enum to the Kubernetes API service type string.
func resolveServiceType(t kubernetesservicev1.KubernetesServiceSpec_KubernetesServiceType) string {
	switch t {
	case kubernetesservicev1.KubernetesServiceSpec_node_port:
		return "NodePort"
	case kubernetesservicev1.KubernetesServiceSpec_load_balancer:
		return "LoadBalancer"
	case kubernetesservicev1.KubernetesServiceSpec_external_name:
		return "ExternalName"
	default:
		return "ClusterIP"
	}
}

// resolveExternalTrafficPolicy maps the protobuf enum to the Kubernetes API string.
func resolveExternalTrafficPolicy(p kubernetesservicev1.KubernetesServiceSpec_KubernetesServiceExternalTrafficPolicy) string {
	switch p {
	case kubernetesservicev1.KubernetesServiceSpec_local:
		return "Local"
	default:
		return "Cluster"
	}
}

// resolveSessionAffinity maps the protobuf enum to the Kubernetes API string.
func resolveSessionAffinity(s kubernetesservicev1.KubernetesServiceSpec_KubernetesServiceSessionAffinity) string {
	switch s {
	case kubernetesservicev1.KubernetesServiceSpec_client_ip:
		return "ClientIP"
	default:
		return "None"
	}
}

// resolveProtocol maps the protobuf protocol enum to the Kubernetes API string.
func resolveProtocol(p kubernetesservicev1.KubernetesServicePort_KubernetesServiceProtocol) string {
	switch p {
	case kubernetesservicev1.KubernetesServicePort_UDP:
		return "UDP"
	case kubernetesservicev1.KubernetesServicePort_SCTP:
		return "SCTP"
	default:
		return "TCP"
	}
}

// internalDnsName builds the fully qualified internal DNS name for the service.
func internalDnsName(name, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", strings.ToLower(name), strings.ToLower(namespace))
}
