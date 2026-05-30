package module

import (
	kuberneteshttproutev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kuberneteshttproute/v1"
	gatewayv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildRules maps the OpenMCF route rules onto the typed crd2pulumi rules array.
// Each rule's matches, filters, and backend refs are delegated to their own
// builders; optional collections are only set when present so the controller
// defaults (for example, the implicit "match all on /" rule) flow through.
func buildRules(rules []*kuberneteshttproutev1.KubernetesHttpRouteRule) gatewayv1.HTTPRouteSpecRulesArray {
	arr := gatewayv1.HTTPRouteSpecRulesArray{}
	for _, r := range rules {
		args := gatewayv1.HTTPRouteSpecRulesArgs{}
		if name := r.GetName(); name != "" {
			args.Name = pulumi.String(name)
		}
		if matches := r.GetMatches(); len(matches) > 0 {
			args.Matches = buildMatches(matches)
		}
		if filters := r.GetFilters(); len(filters) > 0 {
			args.Filters = buildRuleFilters(filters)
		}
		if backendRefs := r.GetBackendRefs(); len(backendRefs) > 0 {
			args.BackendRefs = buildBackendRefs(backendRefs)
		}
		if timeouts := r.GetTimeouts(); timeouts != nil {
			args.Timeouts = buildTimeouts(timeouts)
		}
		arr = append(arr, args)
	}
	return arr
}

func buildTimeouts(timeouts *kuberneteshttproutev1.KubernetesHttpRouteTimeouts) gatewayv1.HTTPRouteSpecRulesTimeoutsArgs {
	args := gatewayv1.HTTPRouteSpecRulesTimeoutsArgs{}
	if request := timeouts.GetRequest(); request != "" {
		args.Request = pulumi.String(request)
	}
	if backendRequest := timeouts.GetBackendRequest(); backendRequest != "" {
		args.BackendRequest = pulumi.String(backendRequest)
	}
	return args
}
