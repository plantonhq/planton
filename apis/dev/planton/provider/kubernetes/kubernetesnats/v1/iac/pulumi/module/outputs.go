package module

// Constants exported by the KubernetesNats Pulumi module. They map
// one-for-one to fields in KubernetesNatsStackOutputs so callers can
// programmatically fetch them (e.g. CI/CD jobs or downstream stacks).
//
// ⚠️ Keep these string IDs stable – they are consumed by automation that
// marshals the Pulumi outputs into Kubernetes Secrets.
const (
	OpNamespace         = "namespace"
	OpClientUrlInternal = "client_url_internal"
	OpClientUrlExternal = "client_url_external"

	// Auth secret <name>, <key>
	OpAuthSecretName = "auth_token_secret.name"
	OpAuthSecretKey  = "auth_token_secret.key"

	// TLS secret <name>, <key>
	OpTlsSecretName = "tls_secret.name"
	OpTlsSecretKey  = "tls_secret.key"

	// JetStream domain & metrics
	OpJetStreamDomain = "jet_stream_domain"
	OpMetricsEndpoint = "metrics_endpoint"

	// NACK controller outputs
	OpNackControllerEnabled = "nack_controller.enabled"
	OpNackControllerVersion = "nack_controller.version"
	OpStreamsCreated        = "streams_created"
)
