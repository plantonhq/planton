package module

const (
	OpNamespace                       = "namespace"
	OpElasticsearchService            = "elasticsearch.service"
	OpElasticsearchPortForwardCommand = "elasticsearch.port_forward_command"
	OpElasticsearchKubeEndpoint       = "elasticsearch.kube_endpoint"
	OpElasticsearchExternalHostname   = "elasticsearch.external_hostname"
	OpElasticsearchUsername           = "elasticsearch.username"
	OpElasticsearchPasswordSecretName = "elasticsearch.password_secret.name"
	OpElasticsearchPasswordSecretKey  = "elasticsearch.password_secret.key"
	OpKibanaService                   = "kibana.service"
	OpKibanaPortForwardCommand        = "kibana.port_forward_command"
	OpKibanaKubeEndpoint              = "kibana.kube_endpoint"
	OpKibanaExternalHostname          = "kibana.external_hostname"
)
