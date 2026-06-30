package module

import (
	kubernetesgrpcroutev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgrpcroute/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildRules maps the Planton route rules onto the typed crd2pulumi rules array.
// Each rule's matches, filters, and backend refs are delegated to their own
// builders; optional collections are only set when present so the controller
// defaults flow through. GRPCRouteRule has no timeouts (unlike HTTPRouteRule).
func buildRules(rules []*kubernetesgrpcroutev1.KubernetesGrpcRouteRule) gatewayv1.GRPCRouteSpecRulesArray {
	arr := gatewayv1.GRPCRouteSpecRulesArray{}
	for _, r := range rules {
		args := gatewayv1.GRPCRouteSpecRulesArgs{}
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
		arr = append(arr, args)
	}
	return arr
}
