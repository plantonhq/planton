package module

import (
	kuberneteshttproutev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kuberneteshttproute/v1"
	gatewayv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// This file maps the single OpenMCF KubernetesHttpRouteFilter message onto the
// rule-level crd2pulumi filter type tree (HTTPRouteSpecRulesFilters*). The same
// proto message is also valid on a backend ref, but crd2pulumi generates a
// structurally identical yet distinct type tree there
// (HTTPRouteSpecRulesBackendRefsFilters*), so backend_refs.go carries a parallel
// set of builders. The duplication is forced by the generated SDK, not by the
// OpenMCF model (the same pattern the Gateway component uses for its two
// non-interchangeable namespace-selector types).

// buildRuleFilters maps rule-level filters. Each filter is a discriminated union
// whose populated field matches its type (enforced by spec validation), so only
// the field for the given type is read.
func buildRuleFilters(filters []*kuberneteshttproutev1.KubernetesHttpRouteFilter) gatewayv1.HTTPRouteSpecRulesFiltersArray {
	arr := gatewayv1.HTTPRouteSpecRulesFiltersArray{}
	for _, f := range filters {
		args := gatewayv1.HTTPRouteSpecRulesFiltersArgs{
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
		if v := f.GetRequestRedirect(); v != nil {
			args.RequestRedirect = buildRuleRequestRedirect(v)
		}
		if v := f.GetUrlRewrite(); v != nil {
			args.UrlRewrite = buildRuleUrlRewrite(v)
		}
		if v := f.GetCors(); v != nil {
			args.Cors = buildRuleCors(v)
		}
		if v := f.GetExtensionRef(); v != nil {
			args.ExtensionRef = gatewayv1.HTTPRouteSpecRulesFiltersExtensionRefArgs{
				Group: pulumi.String(v.GetGroup()),
				Kind:  pulumi.String(v.GetKind()),
				Name:  pulumi.String(v.GetName()),
			}
		}
		arr = append(arr, args)
	}
	return arr
}

func buildRuleRequestHeaderModifier(m *kuberneteshttproutev1.KubernetesHttpRouteHeaderFilter) gatewayv1.HTTPRouteSpecRulesFiltersRequestHeaderModifierArgs {
	args := gatewayv1.HTTPRouteSpecRulesFiltersRequestHeaderModifierArgs{}
	if set := m.GetSet(); len(set) > 0 {
		setArr := gatewayv1.HTTPRouteSpecRulesFiltersRequestHeaderModifierSetArray{}
		for _, h := range set {
			setArr = append(setArr, gatewayv1.HTTPRouteSpecRulesFiltersRequestHeaderModifierSetArgs{
				Name:  pulumi.String(h.GetName()),
				Value: pulumi.String(h.GetValue()),
			})
		}
		args.Set = setArr
	}
	if add := m.GetAdd(); len(add) > 0 {
		addArr := gatewayv1.HTTPRouteSpecRulesFiltersRequestHeaderModifierAddArray{}
		for _, h := range add {
			addArr = append(addArr, gatewayv1.HTTPRouteSpecRulesFiltersRequestHeaderModifierAddArgs{
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

func buildRuleResponseHeaderModifier(m *kuberneteshttproutev1.KubernetesHttpRouteHeaderFilter) gatewayv1.HTTPRouteSpecRulesFiltersResponseHeaderModifierArgs {
	args := gatewayv1.HTTPRouteSpecRulesFiltersResponseHeaderModifierArgs{}
	if set := m.GetSet(); len(set) > 0 {
		setArr := gatewayv1.HTTPRouteSpecRulesFiltersResponseHeaderModifierSetArray{}
		for _, h := range set {
			setArr = append(setArr, gatewayv1.HTTPRouteSpecRulesFiltersResponseHeaderModifierSetArgs{
				Name:  pulumi.String(h.GetName()),
				Value: pulumi.String(h.GetValue()),
			})
		}
		args.Set = setArr
	}
	if add := m.GetAdd(); len(add) > 0 {
		addArr := gatewayv1.HTTPRouteSpecRulesFiltersResponseHeaderModifierAddArray{}
		for _, h := range add {
			addArr = append(addArr, gatewayv1.HTTPRouteSpecRulesFiltersResponseHeaderModifierAddArgs{
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

func buildRuleRequestMirror(m *kuberneteshttproutev1.KubernetesHttpRouteRequestMirrorFilter) gatewayv1.HTTPRouteSpecRulesFiltersRequestMirrorArgs {
	args := gatewayv1.HTTPRouteSpecRulesFiltersRequestMirrorArgs{}
	if ref := m.GetBackendRef(); ref != nil {
		backendRef := gatewayv1.HTTPRouteSpecRulesFiltersRequestMirrorBackendRefArgs{
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
		fractionArgs := gatewayv1.HTTPRouteSpecRulesFiltersRequestMirrorFractionArgs{
			Numerator: pulumi.Int(int(fraction.GetNumerator())),
		}
		if fraction.Denominator != nil {
			fractionArgs.Denominator = pulumi.Int(int(fraction.GetDenominator()))
		}
		args.Fraction = fractionArgs
	}
	return args
}

func buildRuleRequestRedirect(r *kuberneteshttproutev1.KubernetesHttpRouteRequestRedirectFilter) gatewayv1.HTTPRouteSpecRulesFiltersRequestRedirectArgs {
	args := gatewayv1.HTTPRouteSpecRulesFiltersRequestRedirectArgs{}
	if scheme := r.GetScheme(); scheme != "" {
		args.Scheme = pulumi.String(scheme)
	}
	if hostname := r.GetHostname(); hostname != "" {
		args.Hostname = pulumi.String(hostname)
	}
	if path := r.GetPath(); path != nil {
		pathArgs := gatewayv1.HTTPRouteSpecRulesFiltersRequestRedirectPathArgs{
			Type: pulumi.String(path.GetType()),
		}
		if v := path.GetReplaceFullPath(); v != "" {
			pathArgs.ReplaceFullPath = pulumi.String(v)
		}
		if v := path.GetReplacePrefixMatch(); v != "" {
			pathArgs.ReplacePrefixMatch = pulumi.String(v)
		}
		args.Path = pathArgs
	}
	if r.Port != nil {
		args.Port = pulumi.Int(int(r.GetPort()))
	}
	if r.StatusCode != nil {
		args.StatusCode = pulumi.Int(int(r.GetStatusCode()))
	}
	return args
}

func buildRuleUrlRewrite(u *kuberneteshttproutev1.KubernetesHttpRouteUrlRewriteFilter) gatewayv1.HTTPRouteSpecRulesFiltersUrlRewriteArgs {
	args := gatewayv1.HTTPRouteSpecRulesFiltersUrlRewriteArgs{}
	if hostname := u.GetHostname(); hostname != "" {
		args.Hostname = pulumi.String(hostname)
	}
	if path := u.GetPath(); path != nil {
		pathArgs := gatewayv1.HTTPRouteSpecRulesFiltersUrlRewritePathArgs{
			Type: pulumi.String(path.GetType()),
		}
		if v := path.GetReplaceFullPath(); v != "" {
			pathArgs.ReplaceFullPath = pulumi.String(v)
		}
		if v := path.GetReplacePrefixMatch(); v != "" {
			pathArgs.ReplacePrefixMatch = pulumi.String(v)
		}
		args.Path = pathArgs
	}
	return args
}

func buildRuleCors(c *kuberneteshttproutev1.KubernetesHttpRouteCorsFilter) gatewayv1.HTTPRouteSpecRulesFiltersCorsArgs {
	args := gatewayv1.HTTPRouteSpecRulesFiltersCorsArgs{}
	if v := c.GetAllowOrigins(); len(v) > 0 {
		args.AllowOrigins = pulumi.ToStringArray(v)
	}
	if c.AllowCredentials != nil {
		args.AllowCredentials = pulumi.Bool(c.GetAllowCredentials())
	}
	if v := c.GetAllowMethods(); len(v) > 0 {
		args.AllowMethods = pulumi.ToStringArray(v)
	}
	if v := c.GetAllowHeaders(); len(v) > 0 {
		args.AllowHeaders = pulumi.ToStringArray(v)
	}
	if v := c.GetExposeHeaders(); len(v) > 0 {
		args.ExposeHeaders = pulumi.ToStringArray(v)
	}
	if c.MaxAge != nil {
		args.MaxAge = pulumi.Int(int(c.GetMaxAge()))
	}
	return args
}
