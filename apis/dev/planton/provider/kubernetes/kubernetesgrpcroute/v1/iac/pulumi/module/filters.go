package module

import (
	kubernetesgrpcroutev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgrpcroute/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// This file maps the single Planton KubernetesGrpcRouteFilter message onto the
// rule-level crd2pulumi filter type tree (GRPCRouteSpecRulesFilters*). The same
// proto message is also valid on a backend ref, but crd2pulumi generates a
// structurally identical yet distinct type tree there
// (GRPCRouteSpecRulesBackendRefsFilters*), so backend_refs.go carries a parallel
// set of builders. The duplication is forced by the generated SDK, not by the
// Planton model (the same pattern HTTPRoute and Gateway use). The GRPC filter set
// is a subset of HTTP's: only header modifiers, request mirror, and extension ref
// -- no redirect, URL rewrite, or CORS.

// buildRuleFilters maps rule-level filters. Each filter is a discriminated union
// whose populated field matches its type (enforced by spec validation), so only
// the field for the given type is read.
func buildRuleFilters(filters []*kubernetesgrpcroutev1.KubernetesGrpcRouteFilter) gatewayv1.GRPCRouteSpecRulesFiltersArray {
	arr := gatewayv1.GRPCRouteSpecRulesFiltersArray{}
	for _, f := range filters {
		args := gatewayv1.GRPCRouteSpecRulesFiltersArgs{
			Type: pulumi.String(f.GetType()),
		}
		if v := f.GetRequestHeaderModifier(); v != nil {
			args.RequestHeaderModifier = buildRuleRequestHeaderModifier(v)
		}
		if v := f.GetResponseHeaderModifier(); v != nil {
			args.ResponseHeaderModifier = buildRuleResponseHeaderModifier(v)
		}
		if v := f.GetRequestMirror(); v != nil {
			args.RequestMirror = buildRuleRequestMirror(v)
		}
		if v := f.GetExtensionRef(); v != nil {
			args.ExtensionRef = gatewayv1.GRPCRouteSpecRulesFiltersExtensionRefArgs{
				Group: pulumi.String(v.GetGroup()),
				Kind:  pulumi.String(v.GetKind()),
				Name:  pulumi.String(v.GetName()),
			}
		}
		arr = append(arr, args)
	}
	return arr
}

func buildRuleRequestHeaderModifier(m *kubernetesgrpcroutev1.KubernetesGrpcRouteHeaderFilter) gatewayv1.GRPCRouteSpecRulesFiltersRequestHeaderModifierArgs {
	args := gatewayv1.GRPCRouteSpecRulesFiltersRequestHeaderModifierArgs{}
	if set := m.GetSet(); len(set) > 0 {
		setArr := gatewayv1.GRPCRouteSpecRulesFiltersRequestHeaderModifierSetArray{}
		for _, h := range set {
			setArr = append(setArr, gatewayv1.GRPCRouteSpecRulesFiltersRequestHeaderModifierSetArgs{
				Name:  pulumi.String(h.GetName()),
				Value: pulumi.String(h.GetValue()),
			})
		}
		args.Set = setArr
	}
	if add := m.GetAdd(); len(add) > 0 {
		addArr := gatewayv1.GRPCRouteSpecRulesFiltersRequestHeaderModifierAddArray{}
		for _, h := range add {
			addArr = append(addArr, gatewayv1.GRPCRouteSpecRulesFiltersRequestHeaderModifierAddArgs{
				Name:  pulumi.String(h.GetName()),
				Value: pulumi.String(h.GetValue()),
			})
		}
		args.Add = addArr
	}
	if remove := m.GetRemove(); len(remove) > 0 {
		args.Remove = pulumi.ToStringArray(remove)
	}
	return args
}

func buildRuleResponseHeaderModifier(m *kubernetesgrpcroutev1.KubernetesGrpcRouteHeaderFilter) gatewayv1.GRPCRouteSpecRulesFiltersResponseHeaderModifierArgs {
	args := gatewayv1.GRPCRouteSpecRulesFiltersResponseHeaderModifierArgs{}
	if set := m.GetSet(); len(set) > 0 {
		setArr := gatewayv1.GRPCRouteSpecRulesFiltersResponseHeaderModifierSetArray{}
		for _, h := range set {
			setArr = append(setArr, gatewayv1.GRPCRouteSpecRulesFiltersResponseHeaderModifierSetArgs{
				Name:  pulumi.String(h.GetName()),
				Value: pulumi.String(h.GetValue()),
			})
		}
		args.Set = setArr
	}
	if add := m.GetAdd(); len(add) > 0 {
		addArr := gatewayv1.GRPCRouteSpecRulesFiltersResponseHeaderModifierAddArray{}
		for _, h := range add {
			addArr = append(addArr, gatewayv1.GRPCRouteSpecRulesFiltersResponseHeaderModifierAddArgs{
				Name:  pulumi.String(h.GetName()),
				Value: pulumi.String(h.GetValue()),
			})
		}
		args.Add = addArr
	}
	if remove := m.GetRemove(); len(remove) > 0 {
		args.Remove = pulumi.ToStringArray(remove)
	}
	return args
}

func buildRuleRequestMirror(m *kubernetesgrpcroutev1.KubernetesGrpcRouteRequestMirrorFilter) gatewayv1.GRPCRouteSpecRulesFiltersRequestMirrorArgs {
	args := gatewayv1.GRPCRouteSpecRulesFiltersRequestMirrorArgs{}
	if ref := m.GetBackendRef(); ref != nil {
		backendRef := gatewayv1.GRPCRouteSpecRulesFiltersRequestMirrorBackendRefArgs{
			Name: pulumi.String(ref.GetName()),
		}
		if group := ref.GetGroup(); group != "" {
			backendRef.Group = pulumi.String(group)
		}
		if kind := ref.GetKind(); kind != "" {
			backendRef.Kind = pulumi.String(kind)
		}
		if namespace := ref.GetNamespace(); namespace != "" {
			backendRef.Namespace = pulumi.String(namespace)
		}
		if ref.Port != nil {
			backendRef.Port = pulumi.Int(int(ref.GetPort()))
		}
		args.BackendRef = backendRef
	}
	if m.Percent != nil {
		args.Percent = pulumi.Int(int(m.GetPercent()))
	}
	if fraction := m.GetFraction(); fraction != nil {
		fractionArgs := gatewayv1.GRPCRouteSpecRulesFiltersRequestMirrorFractionArgs{
			Numerator: pulumi.Int(int(fraction.GetNumerator())),
		}
		if fraction.Denominator != nil {
			fractionArgs.Denominator = pulumi.Int(int(fraction.GetDenominator()))
		}
		args.Fraction = fractionArgs
	}
	return args
}
