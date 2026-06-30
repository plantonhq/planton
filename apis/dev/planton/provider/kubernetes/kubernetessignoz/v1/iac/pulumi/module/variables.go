package module

var vars = struct {
	HelmChartName                              string
	HelmChartRepoUrl                           string
	HelmChartVersion                           string
	SignozUIPort                               int
	SignozFrontendPort                         int
	OtelGrpcPort                               int
	OtelHttpPort                               int
	GatewayExternalLoadBalancerServiceHostname string
	GatewayIngressClassName                    string
	IstioIngressNamespace                      string
}{
	HelmChartName:      "signoz",
	HelmChartRepoUrl:   "https://charts.signoz.io",
	HelmChartVersion:   "0.52.0",
	SignozUIPort:       8080,
	SignozFrontendPort: 3301,
	OtelGrpcPort:       4317,
	OtelHttpPort:       4318,
	GatewayExternalLoadBalancerServiceHostname: "istio-ingress-gateway.istio-ingress.svc.cluster.local",
	GatewayIngressClassName:                    "istio",
	IstioIngressNamespace:                      "istio-ingress",
}
