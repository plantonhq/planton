package module

import (
	kubernetesgrpcroutev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesgrpcroute/v1"
	gatewayv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildMatches maps the OpenMCF request matchers (method, headers) onto the typed
// crd2pulumi matches array. Unlike HTTPRoute, GRPCRoute matches on a single
// method object (service/method) plus headers -- there is no path or query-param
// matching.
func buildMatches(matches []*kubernetesgrpcroutev1.KubernetesGrpcRouteMatch) gatewayv1.GRPCRouteSpecRulesMatchesArray {
	arr := gatewayv1.GRPCRouteSpecRulesMatchesArray{}
	for _, m := range matches {
		args := gatewayv1.GRPCRouteSpecRulesMatchesArgs{}
		if method := m.GetMethod(); method != nil {
			args.Method = buildMethodMatch(method)
		}
		if headers := m.GetHeaders(); len(headers) > 0 {
			headerArr := gatewayv1.GRPCRouteSpecRulesMatchesHeadersArray{}
			for _, h := range headers {
				headerArgs := gatewayv1.GRPCRouteSpecRulesMatchesHeadersArgs{
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
		arr = append(arr, args)
	}
	return arr
}

func buildMethodMatch(m *kubernetesgrpcroutev1.KubernetesGrpcRouteMethodMatch) gatewayv1.GRPCRouteSpecRulesMatchesMethodArgs {
	args := gatewayv1.GRPCRouteSpecRulesMatchesMethodArgs{}
	if t := m.GetType(); t != "" {
		args.Type = pulumi.String(t)
	}
	if service := m.GetService(); service != "" {
		args.Service = pulumi.String(service)
	}
	if method := m.GetMethod(); method != "" {
		args.Method = pulumi.String(method)
	}
	return args
}
