package module

// Constants exported by the KubernetesTemporal Pulumi module.  They line
// up one-for-one with fields in KubernetesTemporalStackOutputs so callers
// can rely on stable names when wiring into downstream stacks / jobs.
const (
	OpNamespace                  = "namespace"
	OpFrontendService            = "frontend_service"
	OpUIService                  = "ui_service"
	OpPortForwardFrontendCommand = "port_forward_frontend_command"
	OpPortForwardUICommand       = "port_forward_ui_command"
	OpFrontendEndpoint           = "frontend_endpoint"
	OpWebUiEndpoint              = "web_ui_endpoint"
	OpExternalFrontendHostname   = "external_frontend_hostname"
	OpExternalUIHostname         = "external_ui_hostname"
)
