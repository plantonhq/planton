package module

const (
	// OpClusterId is the exported stack output name for the Kapsule
	// cluster's regional ID. Referenced by ScalewayKapsulePool resources.
	OpClusterId = "cluster_id"

	// OpKubeconfig is the exported stack output name for the raw
	// kubeconfig file content. Marked as sensitive in the Pulumi state.
	OpKubeconfig = "kubeconfig"

	// OpApiserverUrl is the exported stack output name for the Kubernetes
	// API server URL (e.g., "https://<uuid>.api.k8s.fr-par.scw.cloud:6443").
	OpApiserverUrl = "apiserver_url"

	// OpClusterCaCertificate is the exported stack output name for the
	// cluster CA certificate (base64-encoded). Used to configure Kubernetes
	// providers in downstream infra chart resources.
	OpClusterCaCertificate = "cluster_ca_certificate"

	// OpWildcardDns is the exported stack output name for the DNS wildcard
	// of ready nodes in the cluster.
	OpWildcardDns = "wildcard_dns"

	// OpDefaultPoolId is the exported stack output name for the default
	// node pool's regional ID.
	OpDefaultPoolId = "default_pool_id"
)
