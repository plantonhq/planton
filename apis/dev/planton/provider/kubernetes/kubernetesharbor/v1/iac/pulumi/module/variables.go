package module

type vars struct {
	HarborCorePort                             int
	HarborPortalPort                           int
	HarborRegistryPort                         int
	PostgresPort                               int
	RedisPort                                  int
	IstioIngressNamespace                      string
	GatewayIngressClassName                    string
	GatewayExternalLoadBalancerServiceHostname string
}

var variables = &vars{
	HarborCorePort:          80,
	HarborPortalPort:        80,
	HarborRegistryPort:      5000,
	PostgresPort:            5432,
	RedisPort:               6379,
	IstioIngressNamespace:   "istio-ingress",
	GatewayIngressClassName: "istio",
	GatewayExternalLoadBalancerServiceHostname: "external.istio-ingress.svc.cluster.local",
}
