package module

import (
	kubernetesapis "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	kubernetestlsroutev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetestlsroute/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildRules maps the Planton TLS route rules onto the typed crd2pulumi rules
// array. A TLS route rule carries only an optional name and the backend refs
// (no matches, no filters). Upstream permits exactly one rule.
func buildRules(rules []*kubernetestlsroutev1.KubernetesTlsRouteRule) gatewayv1.TLSRouteSpecRulesArray {
	arr := gatewayv1.TLSRouteSpecRulesArray{}
	for _, r := range rules {
		args := gatewayv1.TLSRouteSpecRulesArgs{}
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
// TLS routes have no per-backend filters.
//
// Note: backend_refs is a plain reference, not an Planton foreign key (DD-009).
// Infra-chart authors express the route -> backend dependency via
// metadata.relationships (type: uses) when the backend is Planton-managed.
func buildBackendRefs(backendRefs []*kubernetesapis.KubernetesGatewayApiBackendRef) gatewayv1.TLSRouteSpecRulesBackendRefsArray {
	arr := gatewayv1.TLSRouteSpecRulesBackendRefsArray{}
	for _, b := range backendRefs {
		args := gatewayv1.TLSRouteSpecRulesBackendRefsArgs{
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
