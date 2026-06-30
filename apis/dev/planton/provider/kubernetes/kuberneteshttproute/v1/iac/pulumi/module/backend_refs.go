package module

import (
	kuberneteshttproutev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteshttproute/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildBackendRefs maps the Planton backend references (flattened group / kind /
// name / namespace / port / weight plus optional per-backend filters) onto the
// typed crd2pulumi backendRefs array.
func buildBackendRefs(backendRefs []*kuberneteshttproutev1.KubernetesHttpRouteBackendRef) gatewayv1.HTTPRouteSpecRulesBackendRefsArray {
	arr := gatewayv1.HTTPRouteSpecRulesBackendRefsArray{}
	for _, b := range backendRefs {
		args := gatewayv1.HTTPRouteSpecRulesBackendRefsArgs{
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
		if filters := b.GetFilters(); len(filters) > 0 {
			args.Filters = buildBackendFilters(filters)
		}
		arr = append(arr, args)
	}
	return arr
}

// buildBackendFilters maps the same Planton KubernetesHttpRouteFilter message
// onto the backend-ref-level crd2pulumi filter type tree. This mirrors
// buildRuleFilters in filters.go; the two trees are kept separate because
// crd2pulumi generates distinct (HTTPRouteSpecRulesFilters* vs
// HTTPRouteSpecRulesBackendRefsFilters*) Go types for the same JSON shape.
func buildBackendFilters(filters []*kuberneteshttproutev1.KubernetesHttpRouteFilter) gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersArray {
	arr := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersArray{}
	for _, f := range filters {
		args := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersArgs{
			Type: pulumi.String(f.GetType()),
		}
		if v := f.GetRequestHeaderModifier(); v != nil {
			args.RequestHeaderModifier = buildBackendRequestHeaderModifier(v)
		}
		if v := f.GetResponseHeaderModifier(); v != nil {
			args.ResponseHeaderModifier = buildBackendResponseHeaderModifier(v)
		}
		if v := f.GetRequestMirror(); v != nil {
			args.RequestMirror = buildBackendRequestMirror(v)
		}
		if v := f.GetRequestRedirect(); v != nil {
			args.RequestRedirect = buildBackendRequestRedirect(v)
		}
		if v := f.GetUrlRewrite(); v != nil {
			args.UrlRewrite = buildBackendUrlRewrite(v)
		}
		if v := f.GetCors(); v != nil {
			args.Cors = buildBackendCors(v)
		}
		if v := f.GetExtensionRef(); v != nil {
			args.ExtensionRef = gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersExtensionRefArgs{
				Group: pulumi.String(v.GetGroup()),
				Kind:  pulumi.String(v.GetKind()),
				Name:  pulumi.String(v.GetName()),
			}
		}
		arr = append(arr, args)
	}
	return arr
}

func buildBackendRequestHeaderModifier(m *kuberneteshttproutev1.KubernetesHttpRouteHeaderFilter) gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersRequestHeaderModifierArgs {
	args := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersRequestHeaderModifierArgs{}
	if set := m.GetSet(); len(set) > 0 {
		setArr := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersRequestHeaderModifierSetArray{}
		for _, h := range set {
			setArr = append(setArr, gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersRequestHeaderModifierSetArgs{
				Name:  pulumi.String(h.GetName()),
				Value: pulumi.String(h.GetValue()),
			})
		}
		args.Set = setArr
	}
	if add := m.GetAdd(); len(add) > 0 {
		addArr := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersRequestHeaderModifierAddArray{}
		for _, h := range add {
			addArr = append(addArr, gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersRequestHeaderModifierAddArgs{
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

func buildBackendResponseHeaderModifier(m *kuberneteshttproutev1.KubernetesHttpRouteHeaderFilter) gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersResponseHeaderModifierArgs {
	args := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersResponseHeaderModifierArgs{}
	if set := m.GetSet(); len(set) > 0 {
		setArr := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersResponseHeaderModifierSetArray{}
		for _, h := range set {
			setArr = append(setArr, gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersResponseHeaderModifierSetArgs{
				Name:  pulumi.String(h.GetName()),
				Value: pulumi.String(h.GetValue()),
			})
		}
		args.Set = setArr
	}
	if add := m.GetAdd(); len(add) > 0 {
		addArr := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersResponseHeaderModifierAddArray{}
		for _, h := range add {
			addArr = append(addArr, gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersResponseHeaderModifierAddArgs{
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

func buildBackendRequestMirror(m *kuberneteshttproutev1.KubernetesHttpRouteRequestMirrorFilter) gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersRequestMirrorArgs {
	args := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersRequestMirrorArgs{}
	if ref := m.GetBackendRef(); ref != nil {
		backendRef := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersRequestMirrorBackendRefArgs{
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
		fractionArgs := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersRequestMirrorFractionArgs{
			Numerator: pulumi.Int(int(fraction.GetNumerator())),
		}
		if fraction.Denominator != nil {
			fractionArgs.Denominator = pulumi.Int(int(fraction.GetDenominator()))
		}
		args.Fraction = fractionArgs
	}
	return args
}

func buildBackendRequestRedirect(r *kuberneteshttproutev1.KubernetesHttpRouteRequestRedirectFilter) gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersRequestRedirectArgs {
	args := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersRequestRedirectArgs{}
	if scheme := r.GetScheme(); scheme != "" {
		args.Scheme = pulumi.String(scheme)
	}
	if hostname := r.GetHostname(); hostname != "" {
		args.Hostname = pulumi.String(hostname)
	}
	if path := r.GetPath(); path != nil {
		pathArgs := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersRequestRedirectPathArgs{
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

func buildBackendUrlRewrite(u *kuberneteshttproutev1.KubernetesHttpRouteUrlRewriteFilter) gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersUrlRewriteArgs {
	args := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersUrlRewriteArgs{}
	if hostname := u.GetHostname(); hostname != "" {
		args.Hostname = pulumi.String(hostname)
	}
	if path := u.GetPath(); path != nil {
		pathArgs := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersUrlRewritePathArgs{
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

func buildBackendCors(c *kuberneteshttproutev1.KubernetesHttpRouteCorsFilter) gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersCorsArgs {
	args := gatewayv1.HTTPRouteSpecRulesBackendRefsFiltersCorsArgs{}
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
