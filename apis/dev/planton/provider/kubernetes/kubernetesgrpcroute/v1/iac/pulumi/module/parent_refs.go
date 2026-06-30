package module

import (
	kubernetesapis "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildParentRefs maps the shared parent references (usually Gateways this route
// attaches to) onto the typed crd2pulumi parentRefs array. Optional fields are
// only set when present so controller defaults (group, kind) flow through.
//
// Note: parent_refs is a plain reference, not an Planton foreign key (DD-009).
// Infra-chart authors express the route -> Gateway dependency via
// metadata.relationships (type: depends_on); the names here are literal.
func buildParentRefs(parentRefs []*kubernetesapis.KubernetesGatewayApiParentReference) gatewayv1.GRPCRouteSpecParentRefsArray {
	arr := gatewayv1.GRPCRouteSpecParentRefsArray{}
	for _, ref := range parentRefs {
		args := gatewayv1.GRPCRouteSpecParentRefsArgs{
			Name: pulumi.String(ref.GetName()),
		}
		if group := ref.GetGroup(); group != "" {
			args.Group = pulumi.String(group)
		}
		if kind := ref.GetKind(); kind != "" {
			args.Kind = pulumi.String(kind)
		}
		if namespace := ref.GetNamespace(); namespace != "" {
			args.Namespace = pulumi.String(namespace)
		}
		if sectionName := ref.GetSectionName(); sectionName != "" {
			args.SectionName = pulumi.String(sectionName)
		}
		if ref.Port != nil {
			args.Port = pulumi.Int(int(ref.GetPort()))
		}
		arr = append(arr, args)
	}
	return arr
}
