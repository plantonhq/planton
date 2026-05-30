package module

import (
	kubernetesreferencegrantv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesreferencegrant/v1"
	gatewayv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildFrom maps the trusted sources (group / kind / namespace) onto the typed
// crd2pulumi from array. All three are required by the upstream CRD: group is a
// non-pointer value type (json:"group", no omitempty), so the API server rejects
// the resource unless the key is present -- even for the core API group, where
// the value is the empty string. We therefore always set group (empty string is
// a valid, meaningful value here), never omit it.
//
// Note: from entries are trust assertions about kinds of resources, not OpenMCF
// foreign keys (DD-009). from[].namespace is the one genuine cross-resource
// reference; infra-chart authors express that edge via metadata.relationships
// (type: uses) when the source namespace is OpenMCF-managed.
func buildFrom(from []*kubernetesreferencegrantv1.KubernetesReferenceGrantFrom) gatewayv1.ReferenceGrantSpecFromArray {
	arr := gatewayv1.ReferenceGrantSpecFromArray{}
	for _, f := range from {
		arr = append(arr, gatewayv1.ReferenceGrantSpecFromArgs{
			Group:     pulumi.String(f.GetGroup()),
			Kind:      pulumi.String(f.GetKind()),
			Namespace: pulumi.String(f.GetNamespace()),
		})
	}
	return arr
}

// buildTo maps the referenceable targets (group / kind / optional name) onto the
// typed crd2pulumi to array. group and kind are required by the upstream CRD
// (both non-pointer value types with no omitempty), so group is always set --
// empty string means the core API group, but the key must still be present or the
// API server rejects the resource ("spec.to[].group: Required value"). name is
// the only optional field: absence means the grant covers all resources of the
// group/kind in this namespace.
func buildTo(to []*kubernetesreferencegrantv1.KubernetesReferenceGrantTo) gatewayv1.ReferenceGrantSpecToArray {
	arr := gatewayv1.ReferenceGrantSpecToArray{}
	for _, t := range to {
		args := gatewayv1.ReferenceGrantSpecToArgs{
			Group: pulumi.String(t.GetGroup()),
			Kind:  pulumi.String(t.GetKind()),
		}
		if name := t.GetName(); name != "" {
			args.Name = pulumi.String(name)
		}
		arr = append(arr, args)
	}
	return arr
}
