package module

import (
	kubernetesapis "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	gatewayv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildParentRefs maps the shared parent references (usually Gateways this route
// attaches to) onto the typed crd2pulumi parentRefs array. Optional fields are
// only set when present so controller defaults (group, kind) flow through.
func buildParentRefs(parentRefs []*kubernetesapis.KubernetesGatewayApiParentReference) gatewayv1.HTTPRouteSpecParentRefsArray {
	arr := gatewayv1.HTTPRouteSpecParentRefsArray{}
	for _, ref := range parentRefs {
		args := gatewayv1.HTTPRouteSpecParentRefsArgs{
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
