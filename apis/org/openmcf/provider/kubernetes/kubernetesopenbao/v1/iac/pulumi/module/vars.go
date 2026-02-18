package module

var vars = struct {
	GatewayExternalLoadBalancerServiceHostname string
	GatewayIngressClassName                    string
	HelmChartName                              string
	HelmChartRepoUrl                           string
	HelmChartVersion                           string
	IstioIngressNamespace                      string
	OpenBaoPort                                int
	OpenBaoClusterPort                         int
}{
	GatewayExternalLoadBalancerServiceHostname: "istio-ingress-gateway.istio-ingress.svc.cluster.local",
	GatewayIngressClassName:                    "istio",
	HelmChartName:                              "openbao",
	HelmChartRepoUrl:                           "https://openbao.github.io/openbao-helm",
	HelmChartVersion:                           "0.23.3",
	IstioIngressNamespace:                      "istio-ingress",
	OpenBaoPort:                                8200,
	OpenBaoClusterPort:                         8201,
}
