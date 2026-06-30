package module

// vars groups all operator-level constants so version bumps or repo changes
// remain one-liner edits.
var vars = struct {
	OperatorNamespace        string
	ComponentsNamespace      string
	OperatorReleaseURLFormat string
	TektonConfigName         string

	// Dashboard service configuration (fixed by Tekton operator)
	DashboardServiceName string
	DashboardServicePort int

	// Gateway API configuration for dashboard ingress
	IstioIngressNamespace                      string
	GatewayIngressClassName                    string
	GatewayExternalLoadBalancerServiceHostname string
}{
	OperatorNamespace:   "tekton-operator",
	ComponentsNamespace: "tekton-pipelines",
	// Release URL format: %s is replaced with the version (e.g., v0.78.1)
	// https://github.com/tektoncd/operator/releases
	// Using infra.tekton.dev as per current Tekton operator documentation
	// https://tekton.dev/docs/operator/install/
	OperatorReleaseURLFormat: "https://infra.tekton.dev/tekton-releases/operator/previous/%s/release.yaml",
	TektonConfigName:         "config",

	// Dashboard service (created by Tekton operator when dashboard is enabled)
	DashboardServiceName: "tekton-dashboard",
	DashboardServicePort: 9097,

	// Gateway API settings for Istio ingress
	IstioIngressNamespace:                      "istio-ingress",
	GatewayIngressClassName:                    "istio",
	GatewayExternalLoadBalancerServiceHostname: "istio-ingress-gateway.istio-ingress.svc.cluster.local",
}
