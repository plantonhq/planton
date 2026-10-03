# Azure Network Watcher Flow Log

Records network traffic metadata for one virtual network, subnet, or network interface into a storage account, optionally enriched by Traffic Analytics in a Log Analytics workspace. Every flow log is a child of its region's Network Watcher -- the singleton Azure auto-creates the moment a region hosts a virtual network -- so this component references the watcher rather than creating one. NSG-targeted flow logs are rejected: Azure stopped accepting new ones in June 2025.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Network Watcher flow log** -- the recorder writing flow records (source, destination, port, protocol, allow/deny verdict) for the target into the storage account, attached to the region's Network Watcher (which Azure auto-creates -- the module never creates a watcher)

Creating a flow log also writes a lifecycle-management rule on the target storage account that overwrites any existing lifecycle rules -- a side effect of Azure's design, not of this module.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **The target** -- an Azure Virtual Network, Azure Subnet, or Azure Network Interface whose traffic is recorded (referenced by `targetResourceId`).
- **An Azure Storage Account** flow-log files land in -- one WITHOUT hand-managed lifecycle rules (creating a flow log overwrites them).
- **Optional: an Azure Log Analytics Workspace** for Traffic Analytics.

### Azure Subscription

- **The flow log must live in the target's region** -- flow logging is regional; ARM rejects cross-region pairings at deploy time.
- **Flow logs are near-free**; the costs that matter are storage (bounded by `retentionPolicy`) and, with Traffic Analytics, workspace ingestion.
- **NSG-targeted flow logs are retired for new creates** (since June 2025; full retirement September 2027) -- validation rejects them before touching ARM; target the network scope instead.

## Deploy

### Console

Open the deployment store, find **Azure Network Watcher Flow Log**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the target and storage references, retention, and the optional Traffic Analytics block. Start from the **Virtual Network Flow Log** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureNetworkWatcherFlowLog
metadata:
  name: platform-vnet-flow-log
  org: acme-corp
  env: prod
spec:
  region: eastus
  name: platform-vnet-flow-log
  targetResourceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network/providers/Microsoft.Network/virtualNetworks/platform-vnet
  storageAccountId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network/providers/Microsoft.Storage/storageAccounts/acmeflowlogs
  version: 2
  retentionPolicy:
    enabled: true
    days: 30
```

```shell
planton apply -f flow-log.yaml
```

This records every flow in the `platform-vnet` virtual network into the `acmeflowlogs` storage account as version-2 records pruned after 30 days, attached to the region's auto-created Network Watcher. An Infra Job tracks the provisioning in real time.

### InfraChart

When the network, storage account, and workspace are Infra Components in the same chart, wire them by reference:

```yaml
spec:
  targetResourceId:
    valueFrom:
      kind: AzureVirtualNetwork
      name: platform-vnet
      fieldPath: status.outputs.virtual_network_id
  storageAccountId:
    valueFrom:
      name: flow-log-storage
  trafficAnalytics:
    workspaceId:
      valueFrom:
        name: platform-logs
    workspaceRegion: eastus
    workspaceResourceId:
      valueFrom:
        name: platform-logs
```

The InfraPipeline resolves the dependency graph, provisioning the network, the storage account, and the workspace before the flow log that records into them.

## Key Configuration

These are the most important decisions when configuring an Azure Network Watcher Flow Log. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope the target to the question.** A virtual-network flow log records everything in the network -- comprehensive, and correspondingly voluminous in storage and (with Traffic Analytics) ingestion cost. A subnet or NIC target answers narrower questions at a fraction of the volume. Retargeting updates the flow log in place, so widening later is cheap; NSG targets are rejected outright.

**Say `version: 2` explicitly.** Schema version 1 is the provider default and exists for legacy consumers; version 2 adds flow state and byte/packet counters -- the fields capacity analysis and Traffic Analytics actually want. The version is updatable in place, but new flow logs have no reason to start on 1.

**Retention is the single storage dial.** `retentionPolicy` is required on every flow log: `enabled: true` with `days` prunes files automatically; `enabled: false` keeps them forever, making cleanup your problem. Storage is the cost that scales with traffic and retention.

**Dedicate the storage account, or at least its lifecycle policy.** Creating a flow log writes a storage lifecycle-management rule that overwrites existing rules on the account. An account whose lifecycle policy someone hand-tuned loses that tuning silently. The clean posture: a storage account dedicated to flow logs, with `retentionPolicy.days` as the retention knob.

**Let the watcher stay invisible.** Leave `networkWatcherName` and `networkWatcherResourceGroup` unset and the module resolves the region's auto-created singleton (`NetworkWatcher_{region}` in `NetworkWatcherRG`). Set both, together, only when the subscription genuinely runs a self-managed watcher -- half an address silently lands on the wrong watcher, and both fields are fixed at creation.

**Traffic Analytics is an enrichment with two workspace references.** The block wants the workspace GUID (`workspaceId`, from the workspace's `workspace_customer_id` output) AND its ARM ID (`workspaceResourceId`, from `workspace_id`) -- both from the same Azure Log Analytics Workspace -- plus the workspace's own region, which may differ from the flow log's. Leave the block unset to write raw files only.

**The interval is a cost dial.** Traffic Analytics processes on a 60-minute cadence by default; the 10-minute interval buys near-real-time visibility at roughly six times the processing cadence and correspondingly higher workspace ingestion. Pick 10 only when someone is actually watching.

**Pausing is cheaper than deleting.** `enabled: false` stops collection while keeping the configuration and the already-written files. Deleting the flow log ends retention management -- files linger until the lifecycle rule or your cleanup removes them. For temporary quiet, pause.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Azure Virtual Network (or Azure Subnet / Azure Network Interface) | `targetResourceId` | `status.outputs.virtual_network_id` (or `subnet_id` / `network_interface_id`) |
| Azure Storage Account | `storageAccountId` | `status.outputs.storage_account_id` |
| Azure Log Analytics Workspace | `trafficAnalytics.workspaceId` | `status.outputs.workspace_customer_id` |
| Azure Log Analytics Workspace | `trafficAnalytics.workspaceResourceId` | `status.outputs.workspace_id` |
| Azure Resource Group (self-managed watcher only) | `networkWatcherResourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

`status.outputs` carries the flow log's ARM ID (`flow_log_id`), its name within the watcher (`flow_log_name`), and the Network Watcher it attached to (`network_watcher_name` -- the auto-created regional singleton unless the spec addressed a self-managed one). Nothing downstream consumes a flow log by reference -- it is a leaf recorder -- so these outputs exist for identification and for confirming which watcher the flow log landed on.

## Common Patterns

**Audit baseline** -- Record every flow in a production network into storage with 30-day retention, so traffic records exist before the incident that needs them. Start from the **Virtual Network Flow Log** preset.

**Queryable traffic** -- Add Traffic Analytics to turn "we have flow logs" into "we can answer questions about traffic": queryable flows, topology maps, and threat detections in a Log Analytics workspace. Start from the **Flow Log with Traffic Analytics** preset.

**Scoped forensics** -- Target a single subnet or network interface when the question is narrow (one workload's traffic, one suspect NIC). Same spec, narrower `targetResourceId`, a fraction of the storage and ingestion volume.

## Works With

- [**Azure Virtual Network**](/infra-catalog/azure-virtual-network) -- the broadest recording target; reference its `virtual_network_id` output.
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- a narrower target for scoped recording; reference its `subnet_id` output.
- [**Azure Network Interface**](/infra-catalog/azure-network-interface) -- the narrowest target, one NIC's traffic; reference its `network_interface_id` output.
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- where flow-log files land; use one without hand-managed lifecycle rules.
- [**Azure Log Analytics Workspace**](/infra-catalog/azure-log-analytics-workspace) -- the Traffic Analytics destination; supplies both workspace references.
