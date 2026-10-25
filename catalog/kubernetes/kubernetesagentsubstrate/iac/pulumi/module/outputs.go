package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output names, one per KubernetesAgentSubstrateOutputs field.
const (
	OpNamespace               = "namespace"
	OpAPIAddress              = "api_address"
	OpAPIServerName           = "api_server_name"
	OpAPITokenAudience        = "api_token_audience"
	OpAPITrustBundleName      = "api_trust_bundle_name"
	OpAPITrustBundleSigner    = "api_trust_bundle_signer"
	OpAPIAccessNamespaceLabel = "api_access_namespace_label"
	OpRouterEndpoint          = "router_endpoint"
	OpRouterTLSEndpoint       = "router_tls_endpoint"
	OpStorageLocation         = "storage_location"
	OpActorJWTIssuer          = "actor_jwt_issuer"
	OpGvisorWorkerImage       = "gvisor_worker_image"
	OpMicrovmWorkerImage      = "microvm_worker_image"
	OpVersion                 = "version"
	OpRouterClientSignerName  = "router_client_signer_name"
)

// outputValues resolves the outputs from the settings and the release's
// image list. Terraform twin: iac/tf/outputs.tf.
//
// The router's plaintext endpoint exists only with agentgateway, which
// still serves plaintext listeners; the Envoy router serves only TLS.
func outputValues(s *settings, images map[string]string) map[string]string {
	worker := func(component string) string {
		image := images[component]
		if s.registry != "" && image != "" {
			return mirrored(image, s.registry)
		}
		return image
	}
	return map[string]string{
		OpNamespace:               vars.Namespace,
		OpAPIAddress:              vars.APIServerHost + ":443",
		OpAPIServerName:           vars.APIServerHost,
		OpAPITokenAudience:        vars.APIServerHost,
		OpAPITrustBundleName:      vars.APITrustBundleName,
		OpAPITrustBundleSigner:    vars.APITrustBundleSign,
		OpRouterClientSignerName:  vars.RouterClientSigner,
		OpAPIAccessNamespaceLabel: vars.APIAccessLabelKey + "=" + vars.APIAccessLabelValue,
		OpRouterEndpoint:          plaintextRouterEndpoint(s),
		OpRouterTLSEndpoint:       "https://" + vars.RouterName + "." + vars.Namespace + ".svc:443",
		OpStorageLocation:         s.storage.location(),
		OpActorJWTIssuer:          s.apiServer.actorIssuer,
		OpGvisorWorkerImage:       worker("ateworker-gvisor"),
		OpMicrovmWorkerImage:      worker("ateworker-microvm"),
		OpVersion:                 s.version,
	}
}

func plaintextRouterEndpoint(s *settings) string {
	if !s.router.agentgateway {
		return ""
	}
	return "http://" + vars.RouterName + "." + vars.Namespace + ".svc:80"
}

func exportOutputs(ctx *pulumi.Context, s *settings, images map[string]string) {
	for name, value := range outputValues(s, images) {
		ctx.Export(name, pulumi.String(value))
	}
}
