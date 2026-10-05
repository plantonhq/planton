# AzureAksCluster

## Overview

`AzureAksCluster` provisions an Azure Kubernetes Service (AKS) managed
cluster: the control plane, its identity and access model, its network
fabric, the mandatory default node pool, and Azure-managed add-ons.

The cluster deliberately carries exactly **one** node pool — the default
(system) pool Azure requires at creation. Every additional pool is its own
composable `AzureAksNodePool` resource referencing this cluster's
`cluster_id` output.

## Key Features

- **Full azurerm v5 surface** — network profile, AAD RBAC, API-server
  access controls, maintenance windows, upgrade channels, OIDC issuer,
  workload identity, add-ons (OMS, Key Vault CSI, Defender, Istio, CSI
  storage drivers, KEDA/VPA, etcd CMK), bootstrap/NAP profiles, and more
- **Managed networking by default** — leave `default_node_pool.vnet_subnet_id`
  unset and AKS provisions its own network; reference an `AzureSubnet` for
  BYO-network advanced paths
- **Workload identity seam** — `oidc_issuer_enabled` (on by default)
  publishes `oidc_issuer_url`, consumed by `AzureFederatedIdentityCredential`
- **Composable** — prerequisite is `AzureResourceGroup` only; subnets are
  optional on the default pool

## Prerequisites

| Kind | Why |
|------|-----|
| `AzureResourceGroup` | The cluster is created inside a referenced resource group |

## Outputs

| Output | Description |
|--------|-------------|
| `cluster_id` | ARM ID of the managed cluster (parent seam for node pools) |
| `cluster_name` | Cluster name within the resource group |
| `fqdn` | Public API-server FQDN (empty for private-only clusters) |
| `private_fqdn` | Private API-server FQDN |
| `portal_fqdn` | Azure Portal FQDN |
| `oidc_issuer_url` | OIDC issuer URL for federated credentials |
| `node_resource_group` | Azure-managed node resource group name |
| `node_resource_group_id` | ARM ID of the node resource group |
| `cluster_kubeconfig` | Base64-encoded kubeconfig (treat as secret) |
| `cluster_identity_principal_id` | Managed identity principal ID |
| `kubelet_identity_object_id` | Kubelet identity object ID |
| `kubelet_identity_client_id` | Kubelet identity client ID |
| `current_kubernetes_version` | Control plane version actually running |

## Related Resources

- **`AzureAksNodePool`** — additional workload-shaped pools (general, spot,
  GPU, Windows) with independent lifecycles
- **`AzureSubnet`** — optional BYO-network placement for the default pool
- **`AzureFederatedIdentityCredential`** — binds workload identity to the
  cluster's OIDC issuer

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
