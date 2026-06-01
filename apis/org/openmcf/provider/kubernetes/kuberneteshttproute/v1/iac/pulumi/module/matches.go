package module

import (
	kuberneteshttproutev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kuberneteshttproute/v1"
	gatewayv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildMatches maps the OpenMCF request matchers (path, headers, query params,
// method) onto the typed crd2pulumi matches array.
func buildMatches(matches []*kuberneteshttproutev1.KubernetesHttpRouteMatch) gatewayv1.HTTPRouteSpecRulesMatchesArray {
	arr := gatewayv1.HTTPRouteSpecRulesMatchesArray{}
	for _, m := range matches {
		args := gatewayv1.HTTPRouteSpecRulesMatchesArgs{}
		if path := m.GetPath(); path != nil {
			args.Path = buildMatchPath(path)
		}
		if headers := m.GetHeaders(); len(headers) > 0 {
			headerArr := gatewayv1.HTTPRouteSpecRulesMatchesHeadersArray{}
			for _, h := range headers {
				headerArgs := gatewayv1.HTTPRouteSpecRulesMatchesHeadersArgs{
					Name:  pulumi.String(h.GetName()),
					Value: pulumi.String(h.GetValue()),
				}
				if t := h.GetType(); t != "" {
					headerArgs.Type = pulumi.String(t)
				}
				headerArr = append(headerArr, headerArgs)
			}
			args.Headers = headerArr
		}
		if queryParams := m.GetQueryParams(); len(queryParams) > 0 {
			queryArr := gatewayv1.HTTPRouteSpecRulesMatchesQueryParamsArray{}
			for _, q := range queryParams {
				queryArgs := gatewayv1.HTTPRouteSpecRulesMatchesQueryParamsArgs{
					Name:  pulumi.String(q.GetName()),
					Value: pulumi.String(q.GetValue()),
				}
				if t := q.GetType(); t != "" {
					queryArgs.Type = pulumi.String(t)
				}
				queryArr = append(queryArr, queryArgs)
			}
			args.QueryParams = queryArr
		}
		if method := m.GetMethod(); method != "" {
			args.Method = pulumi.String(method)
		}
		arr = append(arr, args)
	}
	return arr
}

func buildMatchPath(path *kuberneteshttproutev1.KubernetesHttpRoutePathMatch) gatewayv1.HTTPRouteSpecRulesMatchesPathArgs {
	args := gatewayv1.HTTPRouteSpecRulesMatchesPathArgs{}
	if t := path.GetType(); t != "" {
		args.Type = pulumi.String(t)
	}
	if value := path.GetValue(); value != "" {
		args.Value = pulumi.String(value)
	}
	return args
}
