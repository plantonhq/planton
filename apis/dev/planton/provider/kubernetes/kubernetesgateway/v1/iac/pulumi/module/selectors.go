package module

import (
	kubernetesgatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgateway/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The Gateway CRD generates two structurally identical but distinct Go types
// for the namespace label selector: one under a listener's AllowedRoutes and
// one under the Gateway's AllowedListeners. They are not interchangeable, so a
// single proto KubernetesGatewayLabelSelector is mapped by two dedicated
// builders.

func buildAllowedRoutesSelector(selector *kubernetesgatewayv1.KubernetesGatewayLabelSelector) gatewayv1.GatewaySpecListenersAllowedRoutesNamespacesSelectorArgs {
	args := gatewayv1.GatewaySpecListenersAllowedRoutesNamespacesSelectorArgs{}
	if matchLabels := selector.GetMatchLabels(); len(matchLabels) > 0 {
		args.MatchLabels = pulumi.ToStringMap(matchLabels)
	}
	if expressions := selector.GetMatchExpressions(); len(expressions) > 0 {
		exprArr := gatewayv1.GatewaySpecListenersAllowedRoutesNamespacesSelectorMatchExpressionsArray{}
		for _, e := range expressions {
			exprArgs := gatewayv1.GatewaySpecListenersAllowedRoutesNamespacesSelectorMatchExpressionsArgs{
				Key:      pulumi.String(e.GetKey()),
				Operator: pulumi.String(e.GetOperator()),
			}
			if values := e.GetValues(); len(values) > 0 {
				exprArgs.Values = pulumi.ToStringArray(values)
			}
			exprArr = append(exprArr, exprArgs)
		}
		args.MatchExpressions = exprArr
	}
	return args
}

func buildAllowedListenersSelector(selector *kubernetesgatewayv1.KubernetesGatewayLabelSelector) gatewayv1.GatewaySpecAllowedListenersNamespacesSelectorArgs {
	args := gatewayv1.GatewaySpecAllowedListenersNamespacesSelectorArgs{}
	if matchLabels := selector.GetMatchLabels(); len(matchLabels) > 0 {
		args.MatchLabels = pulumi.ToStringMap(matchLabels)
	}
	if expressions := selector.GetMatchExpressions(); len(expressions) > 0 {
		exprArr := gatewayv1.GatewaySpecAllowedListenersNamespacesSelectorMatchExpressionsArray{}
		for _, e := range expressions {
			exprArgs := gatewayv1.GatewaySpecAllowedListenersNamespacesSelectorMatchExpressionsArgs{
				Key:      pulumi.String(e.GetKey()),
				Operator: pulumi.String(e.GetOperator()),
			}
			if values := e.GetValues(); len(values) > 0 {
				exprArgs.Values = pulumi.ToStringArray(values)
			}
			exprArr = append(exprArr, exprArgs)
		}
		args.MatchExpressions = exprArr
	}
	return args
}
