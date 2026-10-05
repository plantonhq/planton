package module

import (
	"fmt"
	"strconv"

	kubernetesservicev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesservice/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across the module.
type Locals struct {
	Context     *pulumi.Context
	Spec        *kubernetesservicev1alpha1.KubernetesServiceSpec
	Namespace   string
	Name        string
	Labels      map[string]string
	Annotations map[string]string

	// Kubernetes API string forms of the spec enums, resolved once so every
	// consumer (resource args, outputs) agrees on the exact wire value.
	ServiceType           string
	ExternalTrafficPolicy string
	InternalTrafficPolicy string
	TrafficDistribution   string
	SessionAffinity       string
	IpFamilies            []string
	IpFamilyPolicy        string
}

// initializeLocals extracts and transforms spec fields into module-local values.
func initializeLocals(ctx *pulumi.Context, iacInput *kubernetesservicev1alpha1.KubernetesServiceIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	// Resource-identity labels: the kuberneteslabelkeys set, identical to what
	// the Terraform module stamps for the same manifest. User labels merge in
	// afterwards and cannot override the identity keys.
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesService.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}
	for k, v := range spec.GetLabels() {
		if _, isIdentityKey := labels[k]; !isIdentityKey {
			labels[k] = v
		}
	}

	annotations := make(map[string]string)
	for k, v := range spec.GetAnnotations() {
		annotations[k] = v
	}

	// namespace is a StringValueOrRef foreign key. References are resolved to
	// literal strings before the module runs, so GetValue() returns the final
	// namespace name. When omitted entirely, fall back to the cluster's
	// "default" namespace — the same behavior as kubectl without a namespace flag.
	namespace := spec.GetNamespace().GetValue()
	if namespace == "" {
		namespace = "default"
	}

	ipFamilies := make([]string, 0, len(spec.GetIpFamilies()))
	for _, f := range spec.GetIpFamilies() {
		ipFamilies = append(ipFamilies, resolveIpFamily(f))
	}

	return &Locals{
		Context:               ctx,
		Spec:                  spec,
		Namespace:             namespace,
		Name:                  spec.GetName(),
		Labels:                labels,
		Annotations:           annotations,
		ServiceType:           resolveServiceType(spec.GetType()),
		ExternalTrafficPolicy: resolveExternalTrafficPolicy(spec.GetExternalTrafficPolicy()),
		InternalTrafficPolicy: resolveInternalTrafficPolicy(spec.GetInternalTrafficPolicy()),
		TrafficDistribution:   resolveTrafficDistribution(spec.GetTrafficDistribution()),
		SessionAffinity:       resolveSessionAffinity(spec.GetSessionAffinity()),
		IpFamilies:            ipFamilies,
		IpFamilyPolicy:        resolveIpFamilyPolicy(spec.GetIpFamilyPolicy()),
	}
}

// resolveServiceType maps the protobuf enum to the Kubernetes API service type string.
func resolveServiceType(t kubernetesservicev1alpha1.KubernetesServiceSpec_KubernetesServiceType) string {
	switch t {
	case kubernetesservicev1alpha1.KubernetesServiceSpec_node_port:
		return "NodePort"
	case kubernetesservicev1alpha1.KubernetesServiceSpec_load_balancer:
		return "LoadBalancer"
	case kubernetesservicev1alpha1.KubernetesServiceSpec_external_name:
		return "ExternalName"
	default:
		return "ClusterIP"
	}
}

// resolveExternalTrafficPolicy maps the protobuf enum to the Kubernetes API string.
func resolveExternalTrafficPolicy(p kubernetesservicev1alpha1.KubernetesServiceSpec_KubernetesServiceExternalTrafficPolicy) string {
	switch p {
	case kubernetesservicev1alpha1.KubernetesServiceSpec_local:
		return "Local"
	default:
		return "Cluster"
	}
}

// resolveInternalTrafficPolicy maps the protobuf enum to the Kubernetes API string.
func resolveInternalTrafficPolicy(p kubernetesservicev1alpha1.KubernetesServiceSpec_KubernetesServiceInternalTrafficPolicy) string {
	switch p {
	case kubernetesservicev1alpha1.KubernetesServiceSpec_internal_local:
		return "Local"
	default:
		return "Cluster"
	}
}

// resolveTrafficDistribution maps the protobuf enum to the Kubernetes API string.
// Empty string means "unset" — the field is a hint and Kubernetes has no default value.
func resolveTrafficDistribution(d kubernetesservicev1alpha1.KubernetesServiceSpec_KubernetesServiceTrafficDistribution) string {
	switch d {
	case kubernetesservicev1alpha1.KubernetesServiceSpec_prefer_same_zone:
		return "PreferSameZone"
	case kubernetesservicev1alpha1.KubernetesServiceSpec_prefer_same_node:
		return "PreferSameNode"
	default:
		return ""
	}
}

// resolveSessionAffinity maps the protobuf enum to the Kubernetes API string.
func resolveSessionAffinity(s kubernetesservicev1alpha1.KubernetesServiceSpec_KubernetesServiceSessionAffinity) string {
	switch s {
	case kubernetesservicev1alpha1.KubernetesServiceSpec_client_ip:
		return "ClientIP"
	default:
		return "None"
	}
}

// resolveIpFamily maps the protobuf enum to the Kubernetes API string.
func resolveIpFamily(f kubernetesservicev1alpha1.KubernetesServiceSpec_KubernetesServiceIpFamily) string {
	switch f {
	case kubernetesservicev1alpha1.KubernetesServiceSpec_ipv6:
		return "IPv6"
	default:
		return "IPv4"
	}
}

// resolveIpFamilyPolicy maps the protobuf enum to the Kubernetes API string.
// Empty string means "unset" — the cluster then applies SingleStack.
func resolveIpFamilyPolicy(p kubernetesservicev1alpha1.KubernetesServiceSpec_KubernetesServiceIpFamilyPolicy) string {
	switch p {
	case kubernetesservicev1alpha1.KubernetesServiceSpec_single_stack:
		return "SingleStack"
	case kubernetesservicev1alpha1.KubernetesServiceSpec_prefer_dual_stack:
		return "PreferDualStack"
	case kubernetesservicev1alpha1.KubernetesServiceSpec_require_dual_stack:
		return "RequireDualStack"
	default:
		return ""
	}
}

// resolveProtocol maps the protobuf protocol enum to the Kubernetes API string.
func resolveProtocol(p kubernetesservicev1alpha1.KubernetesServicePort_KubernetesServiceProtocol) string {
	switch p {
	case kubernetesservicev1alpha1.KubernetesServicePort_UDP:
		return "UDP"
	case kubernetesservicev1alpha1.KubernetesServicePort_SCTP:
		return "SCTP"
	default:
		return "TCP"
	}
}

// internalDnsName builds the fully qualified in-cluster DNS name of the service.
func internalDnsName(name, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace)
}
