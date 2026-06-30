package module

// Module-level constants and configuration variables
var vars = struct {
	// Tekton namespace (fixed by Tekton manifests)
	TektonNamespace string

	// Tekton Pipelines manifest URL template
	// Version is inserted: https://storage.googleapis.com/tekton-releases/pipeline/{version}/release.yaml
	PipelineReleaseURLTemplate string

	// Tekton Dashboard manifest URL template
	// Version is inserted: https://infra.tekton.dev/tekton-releases/dashboard/{version}/release.yaml
	DashboardReleaseURLTemplate string

	// Dashboard service name (fixed by Tekton Dashboard manifest)
	DashboardServiceName string

	// Dashboard service port
	DashboardServicePort int

	// Gateway API configuration
	IstioIngressNamespace                      string
	GatewayIngressClassName                    string
	GatewayExternalLoadBalancerServiceHostname string
}{
	TektonNamespace:                            "tekton-pipelines",
	PipelineReleaseURLTemplate:                 "https://storage.googleapis.com/tekton-releases/pipeline/%s/release.yaml",
	DashboardReleaseURLTemplate:                "https://infra.tekton.dev/tekton-releases/dashboard/%s/release.yaml",
	DashboardServiceName:                       "tekton-dashboard",
	DashboardServicePort:                       9097,
	IstioIngressNamespace:                      "istio-ingress",
	GatewayIngressClassName:                    "istio",
	GatewayExternalLoadBalancerServiceHostname: "istio-ingress-gateway.istio-ingress.svc.cluster.local",
}
