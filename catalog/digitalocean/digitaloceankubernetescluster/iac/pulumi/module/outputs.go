package module

// Output keys — exactly the DigitalOceanKubernetesClusterOutputs
// contract, identical across both provisioners.
const (
	// OpClusterId is the exported output containing the cluster UUID.
	OpClusterId = "cluster_id"
	// OpKubeconfig is the exported output containing the raw kubeconfig
	// YAML (not base64-encoded; contains admin credentials).
	OpKubeconfig = "kubeconfig"
	// OpApiServerEndpoint is the exported output with the API server URL.
	OpApiServerEndpoint = "api_server_endpoint"
	// OpUrn is the exported output with the cluster's uniform resource
	// name ("do:kubernetes:<cluster_id>").
	OpUrn = "urn"
	// OpIpv4Address is the exported output with the control plane's
	// public IPv4 (empty on HA clusters).
	OpIpv4Address = "ipv4_address"
	// OpDefaultNodePoolId is the exported output with the inline
	// default node pool's UUID.
	OpDefaultNodePoolId = "default_node_pool_id"
	// OpClusterSubnet is the exported output with the pod CIDR block.
	OpClusterSubnet = "cluster_subnet"
	// OpServiceSubnet is the exported output with the service CIDR block.
	OpServiceSubnet = "service_subnet"
)
