package module

var vars = struct {
	GatewayExternalLoadBalancerServiceHostname string
	GatewayIngressClassName                    string
	HelmChartName                              string
	HelmChartRepoUrl                           string
	HelmChartVersion                           string
	IstioIngressNamespace                      string
	JenkinsAdminPasswordSecretKey              string
	JenkinsAdminUsername                       string
	JenkinsDockerImageTag                      string
}{
	GatewayExternalLoadBalancerServiceHostname: "istio-ingress-gateway.istio-ingress.svc.cluster.local",
	GatewayIngressClassName:                    "istio",
	HelmChartName:                              "jenkins",
	HelmChartRepoUrl:                           "https://charts.jenkins.io",
	HelmChartVersion:                           "5.1.5",
	IstioIngressNamespace:                      "istio-ingress",
	JenkinsAdminPasswordSecretKey:              "jenkins-admin-password",
	JenkinsAdminUsername:                       "admin",
	JenkinsDockerImageTag:                      "2.454-jdk17",
}
