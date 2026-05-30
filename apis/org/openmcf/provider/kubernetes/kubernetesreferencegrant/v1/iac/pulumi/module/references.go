package module

import (
	kubernetesreferencegrantv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesreferencegrant/v1"
	gatewayv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildFrom maps the trusted sources (group / kind / namespace) onto the typed
// crd2pulumi from array. kind and namespace are required by the spec; group is
// only set when non-empty so the empty value (core API group) flows through as
// the upstream default rather than being written explicitly.
//
// Note: from entries are trust assertions about kinds of resources, not OpenMCF
// foreign keys (DD-009). from[].namespace is the one genuine cross-resource
// reference; infra-chart authors express that edge via metadata.relationships
// (type: uses) when the source namespace is OpenMCF-managed.
func buildFrom(from []*kubernetesreferencegrantv1.KubernetesReferenceGrantFrom) gatewayv1.ReferenceGrantSpecFromArray {
	arr := gatewayv1.ReferenceGrantSpecFromArray{}
	for _, f := range from {
		args := gatewayv1.ReferenceGrantSpecFromArgs{
			Kind:      pulumi.String(f.GetKind()),
			Namespace: pulumi.String(f.GetNamespace()),
		}
		if group := f.GetGroup(); group != "" {
			args.Group = pulumi.String(group)
		}
		arr = append(arr, args)
	}
	return arr
}

// buildTo maps the referenceable targets (group / kind / optional name) onto the
// typed crd2pulumi to array. kind is required; group is only set when non-empty
// (empty = core API group). name is only set when present -- absence means the
// grant covers all resources of the group/kind in this namespace.
func buildTo(to []*kubernetesreferencegrantv1.KubernetesReferenceGrantTo) gatewayv1.ReferenceGrantSpecToArray {
	arr := gatewayv1.ReferenceGrantSpecToArray{}
	for _, t := range to {
		args := gatewayv1.ReferenceGrantSpecToArgs{
			Kind: pulumi.String(t.GetKind()),
		}
		if group := t.GetGroup(); group != "" {
			args.Group = pulumi.String(group)
		}
		if name := t.GetName(); name != "" {
			args.Name = pulumi.String(name)
		}
		arr = append(arr, args)
	}
	return arr
}
