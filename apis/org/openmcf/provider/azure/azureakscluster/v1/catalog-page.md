# Azure AKS Cluster

Deploys an Azure Kubernetes Service cluster with configurable node pools, Azure CNI networking, optional Azure AD RBAC integration, and managed add-ons including Container Insights, Key Vault CSI driver, and Workload Identity. The component uses a system-assigned managed identity and supports both public and private cluster topologies.

## What Gets Created

When you deploy an AzureAksCluster resource, OpenMCF provisions:

- **AKS Managed Cluster** — a `containerservice.ManagedCluster` resource with system-assigned managed identity, the specified Kubernetes version, and a DNS prefix derived from `metadata.name`
- **System Node Pool** — a `System` mode agent pool with autoscaling enabled, placed in the specified availability zones and VNet subnet
- **User Node Pools** — created for each entry in `userNodePools`, each as a separate `containerservice.AgentPool` resource with independent VM size, autoscaling, and optional Spot instance configuration
- **Network Profile** — configures the cluster networking with the selected plugin (Azure CNI or Kubenet), plugin mode (Overlay or Dynamic), a standard SKU load balancer, and service/DNS CIDRs
- **Azure AD RBAC Integration** — enabled by default, provides managed Azure AD authentication and Kubernetes RBAC authorization
- **OIDC Issuer + Workload Identity** — created only when `addons.enableWorkloadIdentity` is `true`, enables pods to authenticate to Azure services via Kubernetes service accounts
- **Add-on Profiles** — conditionally enabled: Container Insights (OMS agent), Key Vault CSI secrets provider, and Azure Policy

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the cluster will be created (can reference an AzureResourceGroup resource)
- **A VNet subnet** for cluster nodes (can reference an AzureVpc resource)
- **A Log Analytics Workspace resource ID** if enabling Container Insights

## Quick Start

Create a file `aks-cluster.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureAksCluster
metadata:
  name: my-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureAksCluster.my-cluster
spec:
  region: eastus
  resourceGroup: my-rg
  vnetSubnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/nodes
  kubernetesVersion: "1.30"
  systemNodePool:
    vmSize: Standard_D4s_v5
    autoscaling:
      minCount: 1
      maxCount: 3
    availabilityZones:
      - "1"
```

Deploy:

```shell
openmcf apply -f aks-cluster.yaml
```

This creates a public AKS cluster with a single system node pool running Kubernetes 1.30 on Azure CNI Overlay networking with Standard SKU control plane.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the cluster (e.g., `eastus`). | Required |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `vnetSubnetId` | `StringValueOrRef` | Azure resource ID of the VNet subnet for cluster nodes. Can reference an AzureVpc resource via `valueFrom`. | Required |
| `systemNodePool` | `object` | System node pool configuration. See nested fields below. | Required |
| `systemNodePool.vmSize` | `string` | Azure VM size for system nodes (e.g., `Standard_D4s_v5`). | Required. Recommended default: `Standard_D4s_v5` |
| `systemNodePool.autoscaling` | `object` | Autoscaling configuration for the system node pool. | Required |
| `systemNodePool.autoscaling.minCount` | `int` | Minimum number of nodes. | >= 1 |
| `systemNodePool.autoscaling.maxCount` | `int` | Maximum number of nodes. | >= 1 |
| `systemNodePool.availabilityZones` | `string[]` | Availability zones for the system node pool (e.g., `["1", "2", "3"]`). | Minimum 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `kubernetesVersion` | `string` | — | Kubernetes version for the control plane (e.g., `1.30`). Recommended default: `1.30`. |
| `controlPlaneSku` | `enum` | `STANDARD` | Control plane SKU tier. Values: `STANDARD` (99.95% SLA, ~$73/month), `FREE` (no SLA, dev/test only). |
| `networkPlugin` | `enum` | `AZURE_CNI` | Networking plugin. Values: `AZURE_CNI` (recommended), `KUBENET` (deprecated, retiring March 2028). |
| `networkPluginMode` | `enum` | `OVERLAY` | Azure CNI mode. Only applies when `networkPlugin` is `AZURE_CNI`. Values: `OVERLAY` (pods use private CIDR), `DYNAMIC` (pods use real VNet IPs). |
| `privateClusterEnabled` | `bool` | `false` | When `true`, the API server endpoint is private and accessible only from within the VNet. |
| `authorizedIpRanges` | `string[]` | `[]` | CIDR blocks allowed to access the API server. Only applies to public clusters. Must match pattern `x.x.x.x/y`. |
| `disableAzureAdRbac` | `bool` | `false` | When `true`, disables Azure AD integration for Kubernetes RBAC. |
| `userNodePools` | `object[]` | `[]` | User node pools for application workloads. See nested fields below. |
| `userNodePools[].name` | `string` | — | Node pool name. Must be lowercase alphanumeric, max 12 characters. Pattern: `^[a-z0-9]{1,12}$`. |
| `userNodePools[].vmSize` | `string` | — | Azure VM size for this pool. Required per pool. |
| `userNodePools[].autoscaling.minCount` | `int` | — | Minimum node count. >= 1. Required per pool. |
| `userNodePools[].autoscaling.maxCount` | `int` | — | Maximum node count. >= 1. Required per pool. |
| `userNodePools[].availabilityZones` | `string[]` | — | Availability zones for the pool. Minimum 1 item. Required per pool. |
| `userNodePools[].spotEnabled` | `bool` | `false` | Enables Azure Spot instances. Eviction policy is `Delete` with max price set to on-demand rate. |
| `addons.enableContainerInsights` | `bool` | `false` | Enables Azure Monitor Container Insights. Requires `addons.logAnalyticsWorkspaceId`. |
| `addons.enableKeyVaultCsiDriver` | `bool` | `false` | Enables the Azure Key Vault CSI driver for mounting secrets as volumes. |
| `addons.enableAzurePolicy` | `bool` | `false` | Enables the Azure Policy add-on for governance. |
| `addons.enableWorkloadIdentity` | `bool` | `false` | Enables Azure AD Workload Identity and OIDC issuer for secret-less pod authentication. |
| `addons.logAnalyticsWorkspaceId` | `string` | — | Azure resource ID of the Log Analytics Workspace. Required when `enableContainerInsights` is `true`. |
| `advancedNetworking.podCidr` | `string` | `10.244.0.0/16` | Pod CIDR for Overlay mode. Only applies when `networkPluginMode` is `OVERLAY`. Must be valid CIDR. |
| `advancedNetworking.serviceCidr` | `string` | `10.0.0.0/16` | Service CIDR for Kubernetes services. Must not overlap with VNet or pod CIDR. |
| `advancedNetworking.dnsServiceIp` | `string` | `10.0.0.10` | DNS service IP. Must be within `serviceCidr` range. |
| `advancedNetworking.customDnsServers` | `string[]` | `[]` | Custom DNS servers for the VNet. Leave empty for Azure-provided DNS. |

## Examples

### Dev/Test Cluster with Free Tier

A minimal cluster for development with free control plane SKU and a single availability zone:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureAksCluster
metadata:
  name: dev-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureAksCluster.dev-cluster
spec:
  region: eastus
  resourceGroup: dev-rg
  vnetSubnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/dev-rg/providers/Microsoft.Network/virtualNetworks/dev-vnet/subnets/nodes
  kubernetesVersion: "1.30"
  controlPlaneSku: FREE
  systemNodePool:
    vmSize: Standard_D2s_v3
    autoscaling:
      minCount: 1
      maxCount: 3
    availabilityZones:
      - "1"
```

### Production Cluster with User Node Pools

A production cluster with Standard SKU, multi-AZ system pool, a general-purpose user pool, and Azure AD RBAC:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureAksCluster
metadata:
  name: prod-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureAksCluster.prod-cluster
spec:
  region: eastus
  resourceGroup: prod-rg
  vnetSubnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/nodes
  kubernetesVersion: "1.30"
  systemNodePool:
    vmSize: Standard_D4s_v5
    autoscaling:
      minCount: 3
      maxCount: 5
    availabilityZones:
      - "1"
      - "2"
      - "3"
  userNodePools:
    - name: general
      vmSize: Standard_D8s_v5
      autoscaling:
        minCount: 3
        maxCount: 10
      availabilityZones:
        - "1"
        - "2"
        - "3"
```

### Private Cluster with Add-ons and Spot Pools

A private cluster with no public API endpoint, full add-on suite, a general user pool, and a cost-optimized Spot pool for batch workloads:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureAksCluster
metadata:
  name: secure-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureAksCluster.secure-cluster
spec:
  region: westeurope
  resourceGroup: secure-rg
  vnetSubnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/secure-rg/providers/Microsoft.Network/virtualNetworks/secure-vnet/subnets/nodes
  kubernetesVersion: "1.30"
  privateClusterEnabled: true
  systemNodePool:
    vmSize: Standard_D4s_v5
    autoscaling:
      minCount: 3
      maxCount: 5
    availabilityZones:
      - "1"
      - "2"
      - "3"
  userNodePools:
    - name: general
      vmSize: Standard_D8s_v5
      autoscaling:
        minCount: 3
        maxCount: 15
      availabilityZones:
        - "1"
        - "2"
        - "3"
    - name: spot
      vmSize: Standard_D8s_v5
      autoscaling:
        minCount: 1
        maxCount: 20
      availabilityZones:
        - "1"
        - "2"
        - "3"
      spotEnabled: true
  addons:
    enableContainerInsights: true
    enableKeyVaultCsiDriver: true
    enableAzurePolicy: true
    enableWorkloadIdentity: true
    logAnalyticsWorkspaceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/secure-rg/providers/Microsoft.OperationalInsights/workspaces/secure-logs
  advancedNetworking:
    podCidr: "10.244.0.0/16"
    serviceCidr: "10.0.0.0/16"
    dnsServiceIp: "10.0.0.10"
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureAksCluster
metadata:
  name: ref-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureAksCluster.ref-cluster
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  vnetSubnetId:
    valueFrom:
      kind: AzureVpc
      name: my-vpc
      field: status.outputs.nodes_subnet_id
  kubernetesVersion: "1.30"
  systemNodePool:
    vmSize: Standard_D4s_v5
    autoscaling:
      minCount: 3
      maxCount: 5
    availabilityZones:
      - "1"
      - "2"
      - "3"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `api_server_endpoint` | `string` | URL of the Kubernetes API server endpoint for the AKS cluster |
| `cluster_resource_id` | `string` | Azure Resource ID of the AKS cluster |
| `cluster_kubeconfig` | `string` | Base64-encoded kubeconfig file contents for cluster access |
| `managed_identity_principal_id` | `string` | Azure AD principal ID of the cluster's kubelet managed identity |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) — provides the resource group for cluster placement
- [AzureVpc](/docs/catalog/azure/azurevpc) — provides the VNet and subnet for cluster nodes
- [AzureKeyVault](/docs/catalog/azure/azurekeyvault) — stores secrets that pods can mount via the Key Vault CSI driver
- [AzureLogAnalyticsWorkspace](/docs/catalog/azure/azureloganalyticsworkspace) — provides the workspace for Container Insights monitoring
- [AzureContainerRegistry](/docs/catalog/azure/azurecontainerregistry) — hosts container images for cluster workloads
