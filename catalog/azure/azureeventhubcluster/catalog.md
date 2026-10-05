# Azure Event Hub Cluster

Provisions a dedicated Event Hubs cluster -- single-tenant capacity units (CUs) of guaranteed, isolated throughput that namespaces are placed on via their `dedicatedClusterId` reference. The cluster is the top of the Event Hubs capacity ladder, above PREMIUM's shared infrastructure: sitting on one unlocks up to 1024 partitions per hub, 90-day retention, and namespace-level customer-managed-key encryption. Many namespaces share one cluster, which is why the cluster is a first-class Infra Component rather than a namespace property. Dedicated clusters bill per capacity unit per hour at enterprise rates whether traffic flows or not -- the most expensive resource in the Event Hubs family.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Event Hubs Cluster** -- in the referenced resource group, with the ARM sku composed as `Dedicated_{n}` from your capacity-unit count (Dedicated is the ONLY sku family Azure sells for clusters -- the tier is a constant, the count is the choice)
- **Governance tags** -- your tags merged over the Planton-derived resource tags (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureResourceGroup** the cluster will live in. Reference its `resource_group_name` output via ValueFromRef.
- **Budget sign-off** -- dedicated clusters bill per capacity unit per hour at enterprise rates whether traffic flows or not: this is the most expensive resource in the Event Hubs family. Provision one deliberately; PREMIUM processing units cover most workloads below dedicated scale.

## Deploy

### Console

Open the deployment store, find **Azure Event Hub Cluster**, and click **Deploy**. The creation wizard walks you through placement and naming (with the 4-hour deletion moratorium taught up front), the capacity-unit dial, and governance tags. Start from the **Dedicated Cluster** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventHubCluster
metadata:
  name: streaming-cluster
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: streaming-rg
      fieldPath: status.outputs.resource_group_name
  clusterName: myorg-streaming-dedicated
  capacityUnits: 1
```

```shell
planton apply -f cluster.yaml
```

This creates a one-capacity-unit dedicated cluster named `myorg-streaming-dedicated` in `eastus`, ready for namespaces to be placed on it. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire namespaces onto the cluster deployed in the same InfraPipeline:

```yaml
spec:
  dedicatedClusterId:
    valueFrom:
      kind: AzureEventHubCluster
      name: streaming-cluster
      fieldPath: status.outputs.cluster_id
```

The InfraPipeline resolves the dependency graph, deploys the cluster first, then provisions the namespaces placed on it.

## Key Configuration

These are the most important decisions when configuring a cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Capacity units** -- `capacityUnits` is the cluster's ONLY dial: each CU is a slice of guaranteed, single-tenant ingest/egress capacity, and the modules compose the ARM sku `Dedicated_{n}` from the count. Unset deploys 1 CU (Azure's entry size). Scaling updates in place with no downtime and no data movement -- start small, measure, and grow with real load.

**Placement** -- `region` decides where every namespace on the cluster must live (a namespace joins a cluster in its own region). The `clusterName` serves many teams' namespaces, so name it for the platform capability, not any one workload.

**Tags** -- because many namespaces share one cluster, the cluster's bill is a PLATFORM cost. Tag the owning platform team and cost center here; per-team usage attribution belongs on each namespace's own tags.

**One-way doors and the deletion moratorium** -- `region`, `resourceGroup`, and `clusterName` are fixed at creation: changing any of them replaces the cluster. And Azure forbids deleting a cluster for 4 HOURS after creation -- a destroy inside that window retries until Azure permits it, so expect a destroy of a young cluster to take hours by the service's own rule.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cluster_id` | Azure Resource Manager ID of the cluster | The namespace placement seam: an AzureEventHubNamespace's `dedicatedClusterId` consumes it at namespace CREATION (placement is fixed for the namespace's life) |
| `cluster_name` | The cluster's name within the resource group | Operational tooling and portal navigation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dedicated streaming platform** -- one entry-size cluster owned by the platform team, with product teams' namespaces placed on it: isolation and outer limits centrally funded, streams locally owned. Start from the **Dedicated Cluster** preset.

**CMK-eligible capacity** -- namespaces that must encrypt with customer-managed keys need single-tenant capacity: place them on this cluster (or PREMIUM), then apply an AzureEventHubNamespaceCustomerManagedKey.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- where the cluster lives
- [**Azure Event Hub Namespace**](/infra-catalog/azure-event-hub-namespace) -- placed on the cluster via `dedicatedClusterId` at creation
- [**Azure Event Hub Namespace Customer Managed Key**](/infra-catalog/azure-event-hub-namespace-customer-managed-key) -- BYOK encryption, which requires the single-tenant capacity this cluster provides
- [**Azure Event Hub**](/infra-catalog/azure-event-hub) -- hubs on clustered namespaces may use up to 1024 partitions and 90-day retention
