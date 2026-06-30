package module

var vars = struct {
	GatewayExternalLoadBalancerServiceHostname string
	GatewayIngressClassName                    string
	IstioIngressNamespace                      string
	ElasticsearchVersion                       string
	ElasticsearchPort                          int
	KibanaPort                                 int
}{
	GatewayExternalLoadBalancerServiceHostname: "istio-ingress-gateway.istio-ingress.svc.cluster.local",
	GatewayIngressClassName:                    "istio",
	IstioIngressNamespace:                      "istio-ingress",
	ElasticsearchVersion:                       "8.15.0",
	ElasticsearchPort:                          9200,
	KibanaPort:                                 5601,
}
