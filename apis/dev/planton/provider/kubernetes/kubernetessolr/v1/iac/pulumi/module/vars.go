package module

var vars = struct {
	GatewayExternalLoadBalancerServiceHostname string
	GatewayIngressClassName                    string
	GatewayInternalLoadBalancerServiceHostname string
	IstioIngressNamespace                      string
	SolrCloudSolrModules                       []string
}{
	GatewayExternalLoadBalancerServiceHostname: "istio-ingress-gateway.istio-ingress.svc.cluster.local",
	GatewayIngressClassName:                    "istio",
	GatewayInternalLoadBalancerServiceHostname: "ingress-internal.istio-ingress.svc.cluster.local",
	IstioIngressNamespace:                      "istio-ingress",
	SolrCloudSolrModules: []string{
		"jaegertracer-configurator",
		"ltr",
	},
}
