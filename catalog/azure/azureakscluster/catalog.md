# Azure AKS Cluster

Deploys an Azure Kubernetes Service (AKS) cluster -- the managed control plane plus its built-in system node pool. The cluster is deliberately the CONTROL PLANE'S resource: application capacity attaches as separate AzureAksNodePool Infra Components referencing this cluster's outputs, so pools scale, price (spot), and upgrade independently. The spec covers the full control-plane surface: SKU tier and support plan, the system node pool, CNI networking and egress, API-server exposure, identity and workload identity, and the add-on families from Container Insights to managed Istio.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Managed Control Plane** -- the Kubernetes API server, etcd, and controllers operated by Azure, at the SKU tier you choose (Free / Standard with SLA / Premium with LTS)
- **Default (System) Node Pool** -- the cluster's built-in pool from `defaultNodePool`: VM size, fixed count or autoscaling bounds, availability zones, OS image, disks, and optional kernel/kubelet tuning
- **Networking** -- Azure CNI (flat or Overlay), kubenet, or bring-your-own CNI per `networkProfile`, with the chosen network-policy engine, outbound/egress path, and address ranges
- **Identity Wiring** -- system-assigned or user-assigned control-plane identity, optional kubelet identity, OIDC issuer, and workload identity federation
- **Add-ons** -- only the ones you configure: Container Insights, managed Prometheus, Microsoft Defender, Azure Policy, Key Vault secrets provider, AGIC, web app routing, managed Istio, KEDA/VPA, virtual nodes, image cleaner, KAITO
- **Node Resource Group** -- the AKS-managed resource group (MC_* or your `nodeResourceGroup` name) holding node VMs, disks, and the load balancer

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** -- the cluster object must be created inside a resource group. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **Network planning (bring-your-own-VNet only)** -- when placing nodes into your own AzureSubnet, size it for nodes plus (with flat Azure CNI) max-pods IPs per node, and make sure `serviceCidr`/`podCidr` overlap nothing your VNets or on-premises ranges can route to.
- **Quota** -- the default pool's VM size needs available vCPU quota in the target region (Azure Portal → Quotas → Compute).
- **Entra ID groups (hardened clusters)** -- have the admin group object IDs ready if you plan to disable local accounts.

## Deploy

### Console

Open the deployment store, find **Azure AKS Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the sixteen cluster configuration steps. Start from the **Standard Production AKS Cluster** preset in the [Presets](#presets) tab for a production-ready public cluster, or **Private AKS Cluster with Workload Identity** for a private-endpoint cluster.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureAksCluster
metadata:
  name: prod-aks
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "prod-rg"
  region: eastus2
  name: prod-aks
  kubernetesVersion: "1.32"
  skuTier: STANDARD
  defaultNodePool:
    name: system
    vmSize: Standard_D4s_v5
    autoScalingEnabled: true
    minCount: 3
    maxCount: 5
    zones: ["1", "2", "3"]
    onlyCriticalAddonsEnabled: true
  networkProfile:
    networkPlugin: AZURE_CNI
    networkPluginMode: OVERLAY
  workloadIdentityEnabled: true
```

```shell
planton apply -f azure-aks-cluster.yaml
```

This creates an SLA-backed cluster with a zone-spread autoscaled system pool on Azure CNI Overlay. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the cluster to resources deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  region: eastus2
  name: production-aks
  defaultNodePool:
    name: system
    vmSize: Standard_D4s_v5
    autoScalingEnabled: true
    minCount: 3
    maxCount: 5
    vnetSubnetId:
      valueFrom:
        kind: AzureSubnet
        name: aks-nodes
        fieldPath: status.outputs.subnet_id
  omsAgent:
    logAnalyticsWorkspaceId:
      valueFrom:
        kind: AzureLogAnalyticsWorkspace
        name: prod-logs
        fieldPath: status.outputs.workspace_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group, subnet, and workspace first, then provisions the cluster with the resolved values.

## Key Configuration

These are the most important decisions when configuring an AKS cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SKU tier** -- `skuTier` prices the control plane: `FREE` has no uptime SLA (dev/test), `STANDARD` adds the financially-backed SLA (99.95% with zones) and removes throttling ceilings, `PREMIUM` adds Long-Term Support eligibility (`supportPlan: AKS_LONG_TERM_SUPPORT` requires it).

**The system pool** -- `defaultNodePool` sizes the built-in pool. Exactly one scale mode applies: `nodeCount` (fixed) or `autoScalingEnabled` with `minCount`/`maxCount`. The production pattern: 3-5 nodes across `zones: ["1","2","3"]`, `onlyCriticalAddonsEnabled: true` so applications run in separate AzureAksNodePool resources, and `temporaryNameForRotation` set so future pool changes rotate in place instead of rebuilding the cluster.

**Networking** -- `networkProfile.networkPlugin: AZURE_CNI` with `networkPluginMode: OVERLAY` is the modern default (VNet-frugal pod addressing). The Cilium eBPF data plane (`networkDataPlane: DATA_PLANE_CILIUM`) requires Azure CNI, and Cilium network policy requires the Cilium data plane. `outboundType` picks the egress path -- the Standard LB default, a NAT gateway for SNAT-heavy clusters, or user-defined routing through your firewall.

**API server exposure** -- `privateClusterEnabled: true` replaces the public endpoint with a private endpoint (plan operator/CI network reach and the `privateDnsZoneId` story). Public clusters should set `apiServerAccessProfile.authorizedIpRanges` -- without ranges the endpoint accepts connections from any internet address. Exactly one of `dnsPrefix` (public) / `dnsPrefixPrivateCluster` (private) applies.

**Identity** -- `identity.type: SYSTEM_ASSIGNED` is the simple default; `USER_ASSIGNED` (with `identity.identityIds`) lets you pre-grant RBAC before the cluster exists. `oidcIssuerEnabled` (platform default on) plus `workloadIdentityEnabled: true` gives pods keyless, short-lived Entra ID credentials -- the modern replacement for stored secrets.

**Hardened access** -- `azureActiveDirectoryRoleBasedAccessControl` wires admin groups and optional Azure RBAC for Kubernetes authorization; `localAccountDisabled: true` (which requires Entra ID configured) removes the static admin kubeconfig entirely.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** | `defaultNodePool.vnetSubnetId`, `defaultNodePool.podSubnetId`, `apiServerAccessProfile.subnetId`, `ingressApplicationGateway.subnetId` | `status.outputs.subnet_id` |
| **AzureUserAssignedIdentity** | `identity.identityIds`, `kubeletIdentity.userAssignedIdentityId` | `status.outputs.identity_id` |
| **AzureLogAnalyticsWorkspace** | `omsAgent.logAnalyticsWorkspaceId`, `microsoftDefender.logAnalyticsWorkspaceId` | `status.outputs.workspace_id` |
| **AzurePrivateDnsZone** | `privateDnsZoneId` | `status.outputs.zone_id` |
| **AzureApplicationGateway** | `ingressApplicationGateway.gatewayId` | `status.outputs.application_gateway_id` |
| **AzureDnsZone** | `webAppRouting.dnsZoneIds` | `status.outputs.zone_id` |
| **AzureKeyVault** | `serviceMeshProfile.certificateAuthority.keyVaultId` | `status.outputs.key_vault_id` |
| **AzureKeyVaultKey** | `keyManagementService.keyVaultKeyId` | `status.outputs.key_id` |
| **AzureDiskEncryptionSet** | `diskEncryptionSetId` | `status.outputs.disk_encryption_set_id` |
| **AzurePublicIpPrefix** | `defaultNodePool.nodePublicIpPrefixId`, `networkProfile.loadBalancerProfile.outboundIpPrefixIds` | `status.outputs.public_ip_prefix_id` |
| **AzurePublicIp** | `networkProfile.loadBalancerProfile.outboundIpAddressIds` | `status.outputs.public_ip_id` |
| **AzureContainerRegistry** | `bootstrapProfile.containerRegistryId` | `status.outputs.container_registry_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cluster_id` | Azure Resource Manager ID of the cluster | **AzureAksNodePool** (`kubernetesClusterId`) -- every additional pool attaches through it |
| `cluster_name` | Name of the managed cluster | Scripting, diagnostics settings |
| `fqdn` / `private_fqdn` / `portal_fqdn` | The API server's public / private / portal-facing names | kubeconfig targets, network allow-lists |
| `oidc_issuer_url` | The cluster's OIDC issuer | Federated identity credentials for workload identity |
| `node_resource_group` / `node_resource_group_id` | The AKS-managed infrastructure group | Scoping RBAC grants and policy exemptions |
| `cluster_kubeconfig` | Admin kubeconfig (secret output) | CI bootstrap, GitOps agent installation |
| `cluster_identity_principal_id` | The control-plane identity's principal ID | Role assignments (network, disks, DNS) |
| `kubelet_identity_object_id` / `kubelet_identity_client_id` | The kubelet identity | **AcrPull grants on container registries** |
| `current_kubernetes_version` | The version actually running | Upgrade auditing |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard production cluster** -- Public endpoint, Standard tier, autoscaled zone-spread system pool, CNI Overlay, Container Insights + Key Vault secrets provider + Azure Policy + workload identity. Start from the **Standard Production AKS Cluster** preset.

**Private cluster** -- Private API endpoint with a system-managed private DNS zone, authorized operations via network reach, user-defined or NAT egress. Start from the **Private AKS Cluster with Workload Identity** preset.

**Hardened enterprise cluster** -- Private endpoint, Entra ID admin groups with Azure RBAC, local accounts disabled, KMS etcd encryption, Defender, host encryption. Start from the **Hardened Enterprise AKS Cluster** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- owns the managed-cluster object
- [**Azure AKS Node Pool**](/infra-catalog/azure-aks-node-pool) -- attaches application capacity to this cluster
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- hosts the nodes when you bring your own network
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- pre-created control-plane and kubelet identities
- [**Azure Log Analytics Workspace**](/infra-catalog/azure-log-analytics-workspace) -- receives Container Insights and Defender telemetry
- [**Azure Container Registry**](/infra-catalog/azure-container-registry) -- image source (grant AcrPull to the kubelet identity output)
- [**Azure Private DNS Zone**](/infra-catalog/azure-private-dns-zone) -- resolves a private cluster's API endpoint
