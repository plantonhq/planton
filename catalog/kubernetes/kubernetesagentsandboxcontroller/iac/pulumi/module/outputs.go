package module

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output names, one per KubernetesAgentSandboxControllerOutputs field.
const (
	OpNamespace          = "namespace"
	OpMetricsServiceName = "metrics_service_name"
	OpMetricsPort        = "metrics_port"
	OpRouterServiceName  = "router_service_name"
	OpRouterEndpoint     = "router_endpoint"
)

// outputValues resolves the outputs from the settings. Terraform twin:
// iac/tf/outputs.tf.
func outputValues(s *settings) map[string]interface{} {
	out := map[string]interface{}{
		OpNamespace:          vars.Namespace,
		OpMetricsServiceName: vars.ControllerServiceName,
		OpMetricsPort:        int(s.controller.metricsPort),
		OpRouterServiceName:  "",
		OpRouterEndpoint:     "",
	}
	if r := s.router; r != nil {
		out[OpRouterServiceName] = vars.RouterServiceName
		host := fmt.Sprintf("%s.%s.svc.%s", vars.RouterServiceName, vars.Namespace, s.clusterDomain)
		if r.httpEnabled {
			out[OpRouterEndpoint] = fmt.Sprintf("http://%s:%d", host, r.httpPort)
		} else {
			out[OpRouterEndpoint] = fmt.Sprintf("https://%s:%d", host, r.tls.port)
		}
	}
	return out
}

func exportOutputs(ctx *pulumi.Context, s *settings) {
	values := outputValues(s)
	ctx.Export(OpNamespace, pulumi.String(values[OpNamespace].(string)))
	ctx.Export(OpMetricsServiceName, pulumi.String(values[OpMetricsServiceName].(string)))
	ctx.Export(OpMetricsPort, pulumi.Int(values[OpMetricsPort].(int)))
	ctx.Export(OpRouterServiceName, pulumi.String(values[OpRouterServiceName].(string)))
	ctx.Export(OpRouterEndpoint, pulumi.String(values[OpRouterEndpoint].(string)))
}
