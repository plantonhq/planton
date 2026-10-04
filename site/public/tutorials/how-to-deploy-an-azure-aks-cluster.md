---
title: "How to Deploy an Azure AKS Cluster"
date: "2026-04-02"
author:
  - name: "Planton Team"
    title: "Platform Engineering"
    bio: "Helping teams deploy infrastructure and services without the DevOps bottleneck"
tags:
  - "azure"
  - "aks"
  - "kubernetes"
  - "infra-catalog"
category: "azure"
excerpt: "Deploy a production-ready Azure Kubernetes Service cluster using a single YAML manifest and the Planton CLI."
---

# How to Deploy an Azure AKS Cluster

This tutorial walks you through deploying a managed Kubernetes cluster on Azure Kubernetes Service (AKS) through Planton. You will write a YAML manifest describing the cluster you want, deploy it with a single CLI command, and connect `kubectl` to the running cluster. By the end, you will have a production-ready AKS cluster with autoscaling node pools, Azure CNI Overlay networking, and Azure AD RBAC -- or a lightweight development cluster, depending on your needs.

> **Note**: The Planton web console provides a guided creation wizard for AKS
> and other Infra Components. This tutorial uses the CLI/YAML approach for stability
> and reproducibility. The console UI evolves frequently — always check it for the
> latest experience.

## What You Will Learn

- How AKS fits in Planton's Infra Catalog as an Azure catalog kind
- How to write an `AzureAksCluster` manifest with system node pools, networking, and security configuration
- How to deploy with `planton apply` and monitor progress in real time
- How to connect `kubectl` to the new cluster
- How production and development configurations differ and when to use each

## Prerequisites

- [ ] An Azure Provider Connection configured and set as the default for your target environment. This connection provides the `tenant_id`, `subscription_id`, and credentials Planton needs to provision resources in your Azure subscription. The setup follows the same pattern as the [AWS](/tutorials/how-to-connect-your-aws-account-to-planton) and [GCP](/tutorials/how-to-connect-your-gcp-project-to-planton) provider connection tutorials -- create an `AzureProviderConnection` with your credentials and set it as the default.
- [ ] An Azure Resource Group where the AKS cluster will be created. You can use an existing one or create one through Planton (see [Creating Prerequisites via Planton](#creating-prerequisites-via-planton-alternative) at the end of this tutorial).
- [ ] A Virtual Network (VNet) with a subnet for AKS nodes. The subnet's ARM resource ID is required in the manifest. You can use an existing VNet or create one through Planton (see the same section below).
- [ ] A Planton organization and at least one environment created
- [ ] The `planton` CLI installed and authenticated (`planton auth login`)
- [ ] The [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) installed (for connecting `kubectl` after deployment)

## What Is an AKS Infra Component?

`AzureAksCluster` is a catalog kind in the Infra Catalog that provisions a fully managed Azure Kubernetes Service cluster. You define the cluster configuration in a YAML manifest and apply it with `planton apply` -- Planton handles the Azure Resource Manager operations using your Azure provider connection. For more on Infra Components, see the [Infra Components documentation](/docs/infrastructure/infra-components).

## Step 1: Write the AKS Manifest

Create a file named `aks-cluster.yaml` with the following content. This manifest describes a production-grade AKS cluster with multi-zone high availability, autoscaling system and application node pools, and Azure CNI Overlay networking.

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureAksCluster
metadata:
  name: app-aks-cluster
  org: your-org
  env: production
spec:
  region: eastus
  resourceGroup:
    value: "your-resource-group-name"
  vnetSubnetId:
    value: "/subscriptions/your-subscription-id/resourceGroups/your-rg/providers/Microsoft.Network/virtualNetworks/your-vnet/subnets/aks-nodes"
  kubernetesVersion: "1.30"
  controlPlaneSku: STANDARD
  networkPlugin: AZURE_CNI
  networkPluginMode: OVERLAY
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
        minCount: 2
        maxCount: 10
      availabilityZones:
        - "1"
        - "2"
        - "3"
```

Replace these placeholder values with your own:

- **`metadata.name`**: A name for the AKS cluster. Planton generates a URL-safe slug from it.
- **`metadata.org`**: Your Planton organization slug.
- **`metadata.env`**: The environment this cluster belongs to (e.g., `production`, `staging`, `dev`).
- **`spec.resourceGroup.value`**: The name of an existing Azure Resource Group.
- **`spec.vnetSubnetId.value`**: The full ARM resource ID of the subnet for cluster nodes. You can find this in the Azure portal under your VNet's subnet properties, or by running `az network vnet subnet show --resource-group your-rg --vnet-name your-vnet --name aks-nodes --query id -o tsv`.

The key fields in this manifest:

- **`region`**: The Azure region for the cluster. Choose a region that supports AKS availability zones.
- **`resourceGroup` / `vnetSubnetId`**: These use a nested `value` key because they also support `valueFrom` references to other Infra Components. Literal values work for this tutorial; the [Creating Prerequisites via Planton](#creating-prerequisites-via-planton-alternative) section shows the `valueFrom` approach.
- **`kubernetesVersion`**: Pins the cluster to a specific Kubernetes minor version. Azure supports the current version and two previous minor versions.
- **`controlPlaneSku`**: `STANDARD` provides an uptime SLA (99.95% with AZs) for ~$73/month. `FREE` has no SLA -- suitable for development.
- **`networkPlugin` / `networkPluginMode`**: `AZURE_CNI` with `OVERLAY` is recommended. Pods get IPs from a private range (default `10.244.0.0/16`), separate from your VNet address space. Avoids subnet IP exhaustion at scale.
- **`systemNodePool`**: Required. Runs cluster components (CoreDNS, metrics-server). `Standard_D4s_v5` (4 vCPUs, 16 GB) is recommended for production; `Standard_D2s_v3` for development. Spreading across 3 availability zones enables the 99.95% SLA tier.
- **`userNodePools`**: Where application workloads run, separated from system components. Add multiple pools for different workload profiles (compute, memory, Spot instances).

## Step 2: Deploy with `planton apply`

Run the following command to deploy the AKS cluster. The `-t` flag streams the deployment progress to your terminal in real time.

```bash
planton apply -f aks-cluster.yaml -t
```

Planton validates the manifest, creates a deployment job, and begins provisioning the AKS cluster on Azure. The terminal output shows four phases:

1. **init**: Configures the Azure provider using your connection credentials (a few seconds)
2. **refresh**: Checks for any existing state (a few seconds)
3. **preview**: Plans the changes -- shows the Azure resources that will be created (several seconds)
4. **update**: Creates the AKS cluster, system node pool, user node pools, and configures networking (typically 5-10 minutes)

AKS cluster creation takes longer than many other resource types because Azure needs to provision the control plane, set up networking, and boot the node pool VMs across availability zones. Expect the update phase to take 5-10 minutes for a production configuration.

If you prefer to deploy without streaming, omit the `-t` flag:

```bash
planton apply -f aks-cluster.yaml
```

The CLI prints the deployment job ID immediately. You can check on it later with:

```bash
planton follow <infra-job-id>
```

## Step 3: Verify the Deployment

After the deployment completes, retrieve the Infra Component to see its status and outputs:

```bash
planton get AzureAksCluster app-aks-cluster -o yaml
```

The `status.outputs` section contains the key information about your cluster:

| Output | Description | Example |
|---|---|---|
| `fqdn` | The public FQDN of the Kubernetes API server | `app-aks-cluster-dns-abc123.hcp.eastus.azmk8s.io` |
| `clusterId` | The Azure ARM resource ID of the AKS cluster | `/subscriptions/.../managedClusters/app-aks-cluster` |
| `clusterKubeconfig` | A `$secret/` reference to the base64-encoded kubeconfig, which Planton keeps in your secret store | `$secret/@production/azure-aks-cluster-outputs-app-aks-cluster/cluster_kubeconfig` |
| `clusterIdentityPrincipalId` | Principal ID of the cluster's managed identity | `a1b2c3d4-...` |
| `oidcIssuerUrl` | The cluster's OIDC issuer URL, for workload identity federation | `https://eastus.oic.prod-aks.azure.com/...` |

`planton get` prints the resource as you write it, so the keys are camelCase, like every manifest (`status.outputs.clusterKubeconfig`).

To list all deployment jobs for this resource:

```bash
planton infra job list <infra-component-id>
```

The Infra Component ID is in the `metadata.id` field of the `planton get` output.

## Step 4: Connect kubectl to the Cluster

The standard way to connect `kubectl` to an AKS cluster is through the Azure CLI. Run the following command to merge the cluster's credentials into your local kubeconfig:

```bash
az aks get-credentials \
  --resource-group your-resource-group-name \
  --name app-aks-cluster \
  --overwrite-existing
```

Replace `your-resource-group-name` with the name of the resource group you specified in the manifest.

Verify the connection by listing the cluster nodes:

```bash
kubectl get nodes
```

You should see your system pool nodes and user pool nodes listed with a `Ready` status. With the production manifest from Step 1, you will see at least 5 nodes: 3 system nodes (one per availability zone) and at least 2 user nodes.

```text
NAME                              STATUS   ROLES    AGE   VERSION
aks-general-12345678-vmss000000   Ready    <none>   5m    v1.30.x
aks-general-12345678-vmss000001   Ready    <none>   5m    v1.30.x
aks-system-87654321-vmss000000    Ready    <none>   8m    v1.30.x
aks-system-87654321-vmss000001    Ready    <none>   8m    v1.30.x
aks-system-87654321-vmss000002    Ready    <none>   8m    v1.30.x
```

Alternatively, use the kubeconfig the deployment generated. The `clusterKubeconfig` output from Step 3 is a reference to the secret that holds it; read the secret's `cluster_kubeconfig` key and decode it:

```bash
planton get AzureAksCluster app-aks-cluster -o yaml | yq '.status.outputs.clusterKubeconfig'
# $secret/@production/azure-aks-cluster-outputs-app-aks-cluster/cluster_kubeconfig

planton secret get azure-aks-cluster-outputs-app-aks-cluster --env production \
  --reveal --key cluster_kubeconfig -o plain | base64 -d > kubeconfig-aks.yaml
export KUBECONFIG=kubeconfig-aks.yaml
kubectl get nodes
```

The `az aks get-credentials` approach is recommended because it integrates with Azure AD for authentication and handles token refresh automatically.

## Development Configuration

For development and testing environments where cost and speed matter more than resilience, use a lighter configuration:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureAksCluster
metadata:
  name: app-aks-dev
  org: your-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: "your-dev-resource-group"
  vnetSubnetId:
    value: "/subscriptions/your-subscription-id/resourceGroups/your-dev-rg/providers/Microsoft.Network/virtualNetworks/your-dev-vnet/subnets/aks-nodes"
  kubernetesVersion: "1.30"
  controlPlaneSku: FREE
  networkPlugin: AZURE_CNI
  networkPluginMode: OVERLAY
  systemNodePool:
    vmSize: Standard_D2s_v3
    autoscaling:
      minCount: 1
      maxCount: 3
    availabilityZones:
      - "1"
```

Here is what changed from the production configuration and why:

- **`controlPlaneSku: FREE`**: No uptime SLA and no cost for the control plane tier. The Free tier is functionally identical to Standard for most development workflows -- the difference is in availability guarantees.
- **Single availability zone**: `availabilityZones: ["1"]` instead of three zones. This reduces the minimum node count (no need to spread across zones) and may improve scheduling density.
- **Smaller VM size**: `Standard_D2s_v3` (2 vCPUs, 8 GB RAM) is sufficient for development workloads and costs roughly half of `Standard_D4s_v5`.
- **Lower autoscaling**: `minCount: 1` and `maxCount: 3`. A single system node is adequate for development. The cluster autoscaler adds nodes only when pods cannot be scheduled.
- **No user node pools**: Applications run on the system node pool alongside cluster components. This is acceptable for development where workload isolation is not a concern.

Deploy the development configuration the same way:

```bash
planton apply -f aks-dev.yaml -t
```

## Common Patterns and Tips

### Enabling add-ons

AKS offers several Azure-managed add-ons that you can enable through the `addons` block. These are optional -- they are not included in the manifests above to keep the core path focused, but they are recommended for production clusters.

```yaml
addons:
  enableContainerInsights: true
  logAnalyticsWorkspaceId: "/subscriptions/your-sub-id/resourceGroups/your-rg/providers/Microsoft.OperationalInsights/workspaces/your-workspace"
  enableKeyVaultCsiDriver: true
  enableAzurePolicy: true
  enableWorkloadIdentity: true
```

- **Container Insights** (`enableContainerInsights`): Streams container logs, performance metrics, and Kubernetes events to Azure Monitor. Requires a Log Analytics Workspace -- provide its ARM resource ID in `logAnalyticsWorkspaceId`. Container Insights is only enabled when both the flag is `true` and the workspace ID is provided.
- **Key Vault CSI Driver** (`enableKeyVaultCsiDriver`): Allows pods to mount secrets from Azure Key Vault as volumes, eliminating the need to store secrets in Kubernetes Secrets.
- **Azure Policy** (`enableAzurePolicy`): Enforces governance policies on the cluster (pod security standards, resource quotas, allowed registries).
- **Workload Identity** (`enableWorkloadIdentity`): Allows pods to authenticate to Azure services using Kubernetes service accounts instead of storing credentials. This is the recommended approach for applications that access Azure resources like Key Vault, Storage, or SQL Database.

### Restricting API server access

By default, the AKS cluster has a public API server endpoint accessible from any IP address. For production clusters, restrict access to known networks:

```yaml
authorizedIpRanges:
  - "203.0.113.0/24"
  - "198.51.100.0/24"
```

`authorizedIpRanges` accepts a list of CIDR blocks. Only traffic from these ranges can reach the Kubernetes API server. Add your office network, VPN exit points, and CI/CD agent networks.

For maximum security, deploy a fully private cluster:

```yaml
privateClusterEnabled: true
```

When `privateClusterEnabled` is `true`, the API server has no public endpoint. It is accessible only from within the VNet or through peered networks. This requires a VPN, ExpressRoute, or a bastion host for `kubectl` access. The `authorizedIpRanges` field is not applicable to private clusters.

### Spot instance pools

Azure Spot VMs offer 30-90% cost savings over regular VMs, but Azure can evict them when it needs the capacity back. Use Spot pools for fault-tolerant, stateless workloads like batch jobs, background processing, or stateless API replicas that can tolerate interruptions.

```yaml
userNodePools:
  - name: spot
    vmSize: Standard_D4s_v5
    autoscaling:
      minCount: 0
      maxCount: 20
    availabilityZones:
      - "1"
      - "2"
      - "3"
    spotEnabled: true
```

Setting `minCount: 0` allows the pool to scale to zero when there are no pods to schedule, which means you pay nothing when the pool is idle. When Spot nodes are evicted, the cluster autoscaler provisions replacement nodes automatically.

To direct specific workloads to Spot nodes, use Kubernetes tolerations and node affinity in your pod specs. The Spot pool nodes are created with a `kubernetes.azure.com/scalesetpriority: spot` label and a `kubernetes.azure.com/scalesetpriority=spot:NoSchedule` taint.

### Advanced networking

For most clusters, the default networking configuration is correct. Customize these settings only if your network architecture requires specific CIDR ranges to avoid conflicts:

```yaml
advancedNetworking:
  podCidr: "10.244.0.0/16"
  serviceCidr: "10.0.0.0/16"
  dnsServiceIp: "10.0.0.10"
```

- **`podCidr`**: The CIDR range for pod IPs in Overlay mode. The default `10.244.0.0/16` provides over 65,000 pod IPs. Change this only if it conflicts with your VNet or on-premises networks.
- **`serviceCidr`**: The CIDR range for Kubernetes Service cluster IPs. Must not overlap with the VNet address space, pod CIDR, or any peered networks. The default is `10.0.0.0/16`.
- **`dnsServiceIp`**: The IP address for the cluster DNS service (CoreDNS). Must be within the `serviceCidr` range. The default is `10.0.0.10`.

### Disabling Azure AD RBAC

By default, AKS clusters are created with Azure Active Directory RBAC integration enabled. This means cluster access is managed through Azure AD users and groups, and you can assign Kubernetes RBAC roles using Azure role assignments.

If you need to disable this -- for example, in isolated development environments where Azure AD integration is not available -- set:

```yaml
disableAzureAdRbac: true
```

This is not recommended for production. Without Azure AD RBAC, cluster access is managed solely through Kubernetes-native RBAC with client certificates, which is harder to audit and does not integrate with your organization's identity provider.

## Creating Prerequisites via Planton (Alternative)

If you prefer to manage your Azure Resource Group and Virtual Network as Planton Resources rather than creating them externally, you can deploy them through `planton apply` and reference their outputs in the AKS manifest.

### Step A: Create a Resource Group

Create a file named `resource-group.yaml`:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureResourceGroup
metadata:
  name: aks-infrastructure
  org: your-org
  env: production
spec:
  name: rg-aks-production
  region: eastus
```

Deploy it:

```bash
planton apply -f resource-group.yaml -t
```

### Step B: Create a Virtual Network

Create a file named `vnet.yaml`. The `resourceGroup` field references the Resource Group you created in Step A:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVpc
metadata:
  name: aks-network
  org: your-org
  env: production
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: aks-infrastructure
      env: production
      fieldPath: status.outputs.resource_group_name
  addressSpaceCidr: "10.1.0.0/16"
  nodesSubnetCidr: "10.1.0.0/18"
```

Deploy it:

```bash
planton apply -f vnet.yaml -t
```

The `resourceGroup.valueFrom` tells Planton to resolve the resource group name from the `AzureResourceGroup` Infra Component you created in Step A. The `fieldPath` specifies which output to use -- in this case, the `resource_group_name` output.

### Step C: Deploy AKS with resource references

Now create the AKS manifest with `valueFrom` references instead of literal values:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureAksCluster
metadata:
  name: app-aks-cluster
  org: your-org
  env: production
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: aks-infrastructure
      env: production
      fieldPath: status.outputs.resource_group_name
  vnetSubnetId:
    valueFrom:
      kind: AzureVpc
      name: aks-network
      env: production
      fieldPath: status.outputs.nodes_subnet_id
  kubernetesVersion: "1.30"
  controlPlaneSku: STANDARD
  networkPlugin: AZURE_CNI
  networkPluginMode: OVERLAY
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
        minCount: 2
        maxCount: 10
      availabilityZones:
        - "1"
        - "2"
        - "3"
```

The `valueFrom` references allow Planton to resolve the resource group name and subnet ID from the outputs of the resources you deployed in Steps A and B. This approach is particularly useful when deploying through Infra Charts, where multiple resources are orchestrated together and dependencies are resolved automatically through a DAG.

## What to Do Next

Your AKS cluster is running on Azure. From here:

- **Deploy a backend service** to the cluster. See [How to Deploy Your First Service with Zero-Config CI/CD](/tutorials/how-to-deploy-your-first-service-with-zero-config-cicd) to set up a push-to-deploy workflow for your applications.
- **Deploy Redis or other workloads** onto the cluster. See [How to Deploy Redis on Kubernetes](/tutorials/how-to-deploy-redis-on-kubernetes) -- the Kubernetes Infra Component workflow deploys directly to any connected cluster, including the one you created here.
- **Explore other Azure resources** in the Infra Catalog. The same `planton apply` workflow works for Azure SQL Database, Azure Key Vault, Azure Storage Accounts, Azure Container Registry, and other Azure resource types.
