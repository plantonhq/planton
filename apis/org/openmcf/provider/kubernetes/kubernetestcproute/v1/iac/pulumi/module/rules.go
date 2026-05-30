package module

import (
	kubernetesapis "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	kubernetestcproutev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetestcproute/v1"
	gatewayv1alpha2 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1alpha2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildRules maps the OpenMCF TCP route rules onto the typed crd2pulumi rules
// array. A TCP route rule carries only an optional name and the backend refs
// (no matches, no filters).
func buildRules(rules []*kubernetestcproutev1.KubernetesTcpRouteRule) gatewayv1alpha2.TCPRouteSpecRulesArray {
	arr := gatewayv1alpha2.TCPRouteSpecRulesArray{}
	for _, r := range rules {
		args := gatewayv1alpha2.TCPRouteSpecRulesArgs{}
		if name := r.GetName(); name != "" {
			args.Name = pulumi.String(name)
		}
		if backendRefs := r.GetBackendRefs(); len(backendRefs) > 0 {
			args.BackendRefs = buildBackendRefs(backendRefs)
		}
		arr = append(arr, args)
	}
	return arr
}

// buildBackendRefs maps the shared KubernetesGatewayApiBackendRef (group / kind /
// name / namespace / port / weight) onto the typed crd2pulumi backendRefs array.
// Optional fields are only set when present so controller defaults flow through.
// TCP routes have no per-backend filters.
//
// Note: backend_refs is a plain reference, not an OpenMCF foreign key (DD-009).
// Infra-chart authors express the route -> backend dependency via
// metadata.relationships (type: uses) when the backend is OpenMCF-managed.
func buildBackendRefs(backendRefs []*kubernetesapis.KubernetesGatewayApiBackendRef) gatewayv1alpha2.TCPRouteSpecRulesBackendRefsArray {
	arr := gatewayv1alpha2.TCPRouteSpecRulesBackendRefsArray{}
	for _, b := range backendRefs {
		args := gatewayv1alpha2.TCPRouteSpecRulesBackendRefsArgs{
			Name: pulumi.String(b.GetName()),
		}
		if group := b.GetGroup(); group != "" {
			args.Group = pulumi.String(group)
		}
		if kind := b.GetKind(); kind != "" {
			args.Kind = pulumi.String(kind)
		}
		if namespace := b.GetNamespace(); namespace != "" {
			args.Namespace = pulumi.String(namespace)
		}
		if b.Port != nil {
			args.Port = pulumi.Int(int(b.GetPort()))
		}
		if b.Weight != nil {
			args.Weight = pulumi.Int(int(b.GetWeight()))
		}
		arr = append(arr, args)
	}
	return arr
}
