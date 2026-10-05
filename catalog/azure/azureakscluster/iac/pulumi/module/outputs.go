package module

// Output keys, matching AzureAksClusterOutputs field for field.
const (
	OpClusterId                  = "cluster_id"
	OpClusterName                = "cluster_name"
	OpFqdn                       = "fqdn"
	OpPrivateFqdn                = "private_fqdn"
	OpPortalFqdn                 = "portal_fqdn"
	OpOidcIssuerUrl              = "oidc_issuer_url"
	OpNodeResourceGroup          = "node_resource_group"
	OpNodeResourceGroupId        = "node_resource_group_id"
	OpClusterKubeconfig          = "cluster_kubeconfig"
	OpClusterIdentityPrincipalId = "cluster_identity_principal_id"
	OpKubeletIdentityObjectId    = "kubelet_identity_object_id"
	OpKubeletIdentityClientId    = "kubelet_identity_client_id"
	OpCurrentKubernetesVersion   = "current_kubernetes_version"
	OpClusterCaCertificate       = "cluster_ca_certificate"
	OpEntraIntegrationEnabled    = "entra_integration_enabled"
)
