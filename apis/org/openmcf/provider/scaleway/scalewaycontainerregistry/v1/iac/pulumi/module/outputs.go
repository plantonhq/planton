package module

const (
	// OpNamespaceId is the exported stack output name for the registry
	// namespace's unique identifier (format: "{region}/{uuid}").
	OpNamespaceId = "namespace_id"

	// OpEndpoint is the exported stack output name for the Docker
	// endpoint URL (e.g., "rg.fr-par.scw.cloud/my-registry").
	// Used by Docker clients, Kubernetes imagePullSecrets, and
	// serverless function/container image configuration.
	OpEndpoint = "endpoint"

	// OpNamespaceName is the exported stack output name for the
	// registry namespace's name as it exists in Scaleway.
	OpNamespaceName = "namespace_name"

	// OpRegion is the exported stack output name for the region where
	// the registry namespace is deployed.
	OpRegion = "region"
)
