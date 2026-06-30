package module

var vars = struct {
	// Helm
	HelmChartName    string
	HelmChartRepoUrl string
	HelmChartVersion string
	// Secret key for external DB password (name is computed in locals)
	DatabasePasswordSecretKey string
	// Service ports
	FrontendGrpcPort int
	FrontendHttpPort int
	UIPort           int
	// Istio / Gateway-API constants for UI ingress
	IstioIngressNamespace                      string
	GatewayIngressClassName                    string
	GatewayExternalLoadBalancerServiceHostname string
	// Temporal defaults and common Helm keys
	DefaultCassandraReplicas     int
	HelmKeyPrometheusEnabled     string
	HelmKeyGrafanaEnabled        string
	HelmKeyElasticsearchEnabled  string
	HelmKeyCassandraReplicaCount string
	// History shards configuration
	// This is an immutable setting that determines cluster scalability
	DefaultNumHistoryShards int
	// Dynamic config defaults (for reference - Temporal uses these if not overridden)
	DefaultHistorySizeLimitError  int64 // 50 MB
	DefaultHistoryCountLimitError int64 // 51200 events
	DefaultHistorySizeLimitWarn   int64 // 10 MB
	DefaultHistoryCountLimitWarn  int64 // 10240 events
}{
	HelmChartName:    "temporal",
	HelmChartRepoUrl: "https://go.temporal.io/helm-charts",
	HelmChartVersion: "0.62.0",

	DatabasePasswordSecretKey: "password",

	FrontendGrpcPort: 7233,
	FrontendHttpPort: 7243,
	UIPort:           8080,

	IstioIngressNamespace:                      "istio-ingress",
	GatewayIngressClassName:                    "istio",
	GatewayExternalLoadBalancerServiceHostname: "istio-ingress-gateway.istio-ingress.svc.cluster.local",

	DefaultCassandraReplicas: 1,

	HelmKeyPrometheusEnabled:     "prometheus.enabled",
	HelmKeyGrafanaEnabled:        "grafana.enabled",
	HelmKeyElasticsearchEnabled:  "elasticsearch.enabled",
	HelmKeyCassandraReplicaCount: "cassandra.replicaCount",

	// Default to 512 shards for production workloads
	// This provides good parallelism without excessive overhead
	DefaultNumHistoryShards: 512,

	// Temporal's default dynamic config values (for documentation)
	DefaultHistorySizeLimitError:  52428800, // 50 MB
	DefaultHistoryCountLimitError: 51200,    // 51200 events
	DefaultHistorySizeLimitWarn:   10485760, // 10 MB
	DefaultHistoryCountLimitWarn:  10240,    // 10240 events
}
