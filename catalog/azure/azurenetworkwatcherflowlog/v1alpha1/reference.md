# AzureNetworkWatcherFlowLog

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureNetworkWatcherFlowLogSpec** defines a Network Watcher flow
log -- the recorder that writes network traffic metadata (source,
destination, port, protocol, allow/deny verdict) for ONE target (a
virtual network, subnet, or network interface) into a storage
account, optionally enriched by Traffic Analytics in a Log Analytics
workspace.

**The regional Network Watcher**: every flow log is a child of the
region's Network Watcher. Azure creates that watcher AUTOMATICALLY
(named "NetworkWatcher_{region}" in the "NetworkWatcherRG" resource
group) the moment the region hosts a virtual network, and allows
exactly one per region per subscription -- so this spec references
the watcher rather than creating one. Leave network_watcher_name and
network_watcher_resource_group unset to use the auto-created
singleton (derived from region); set both only when the subscription
runs a self-managed watcher.

**Targets**: virtual network (records ALL flows in the network),
subnet, or network interface -- pick the narrowest scope that
answers your audit question. NSG-targeted flow logs are NOT modeled:
Azure stopped accepting new NSG flow logs on 2025-06-30 (full
retirement 2027-09-30) -- target the network scope instead.

**The storage lifecycle-rule side effect**: creating a flow log
writes a lifecycle-management rule on the target storage account
that OVERWRITES any existing lifecycle rules -- point flow logs at a
storage account that carries no hand-managed lifecycle policy.

## Example

```yaml
# Offline-plan test manifest. Exercises the full surface: a
# virtual-network target, schema version 2, bounded retention, the
# complete Traffic Analytics block at the 10-minute interval, and user
# tags merged over the derived ones. The watcher fields stay UNSET so
# the plan renders the module-derived defaults (NetworkWatcher_eastus
# in NetworkWatcherRG -- the auto-created singleton addressing).
apiVersion: azure.planton.dev/v1alpha1
kind: AzureNetworkWatcherFlowLog
metadata:
  name: test-network-watcher-flow-log
  org: test-org
  env: dev
spec:
  region: eastus
  name: test-vnet-flow-log
  targetResourceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/virtualNetworks/platform-vnet
  storageAccountId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Storage/storageAccounts/platformflowlogs
  version: 2
  retentionPolicy:
    enabled: true
    days: 30
  trafficAnalytics:
    workspaceId:
      value: 11111111-2222-3333-4444-555555555555
    workspaceRegion: eastus
    workspaceResourceId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.OperationalInsights/workspaces/platform-law
    intervalInMinutes: 10
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.targetResourceId` | `string \| valueFrom` | yes |  |  |
| `spec.storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.version` | `int32` |  | `1` |  |
| `spec.retentionPolicy` | `AzureNetworkWatcherFlowLogRetentionPolicy` | yes |  |  |
| `spec.retentionPolicy.enabled` | `bool` |  |  |  |
| `spec.retentionPolicy.days` | `int32` |  |  |  |
| `spec.trafficAnalytics` | `AzureNetworkWatcherFlowLogTrafficAnalytics` |  |  |  |
| `spec.trafficAnalytics.enabled` | `bool` |  | `true` |  |
| `spec.trafficAnalytics.workspaceId` | `string \| valueFrom` | yes |  | AzureLogAnalyticsWorkspace (`status.outputs.workspace_customer_id`) |
| `spec.trafficAnalytics.workspaceRegion` | `string` | yes |  |  |
| `spec.trafficAnalytics.workspaceResourceId` | `string \| valueFrom` | yes |  | AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`) |
| `spec.trafficAnalytics.intervalInMinutes` | `int32` |  | `60` |  |
| `spec.networkWatcherName` | `string` |  |  |  |
| `spec.networkWatcherResourceGroup` | `string \| valueFrom` |  |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region, e.g. "eastus" -- the region of the flow log AND
of the regional Network Watcher it attaches to. Must be the region
the TARGET resource lives in (flow logging is regional; ARM
rejects a cross-region pairing at deploy time). Changing the
region replaces the flow log.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.name

`string` · required

The flow log's name, unique within the Network Watcher. 1-80
characters of letters, numbers, underscores, periods, and hyphens;
starts with a letter or number, ends with a letter, number, or
underscore. Changing the name replaces the flow log.

- rule: Flow log names start with a letter or number, end with a letter, number, or underscore, and may contain alphanumerics, underscores, periods, and hyphens (1-80 characters)
- rule: {"required":true,"string":{"maxLen":"80"}}

### spec.targetResourceId

`string | valueFrom` · required

The resource whose traffic is recorded: a virtual network, subnet,
or network interface, by ARM resource ID. Reference the owning
kind's id output (AzureVirtualNetwork's virtual_network_id,
AzureSubnet's subnet_id, or AzureNetworkInterface's
network_interface_id) or pass a literal ID. Retargeting updates
the flow log in place. NSG targets are rejected -- Azure stopped
accepting new NSG flow logs on 2025-06-30.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.storageAccountId

`string | valueFrom` · required

The storage account flow-log files land in -- references an
AzureStorageAccount's ARM id. Creating the flow log writes a
lifecycle-management rule on this account that OVERWRITES existing
rules (see the spec note). Updatable in place. Containment-exempt: the
flow log WRITES to the account and lives in its Network Watcher's
resource group, so on a diagram it is drawn beside the account with a
line to it, never inside it.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.enabled

`bool` · optional (explicit presence)

Whether the flow log actively records. Unspecified applies true --
a flow log exists to record; set false to pause collection while
keeping the configuration (and the already-written logs) in place.

- default: `true`

### spec.version

`int32` · optional (explicit presence)

The flow-log schema version: 2 adds flow state and byte/packet
counters per flow (and is what Traffic Analytics consumes best);
1 is the legacy record shape. Unspecified applies 1, the
provider's default -- prefer 2 for anything new. Updatable in
place.

- default: `1`
- rule: {"int32":{"lte":2,"gte":1}}

### spec.retentionPolicy

`AzureNetworkWatcherFlowLogRetentionPolicy` · required

How long flow-log files stay in the storage account before Azure
deletes them. Required by the service on every flow log.

- rule: {"required":true}

### spec.retentionPolicy.enabled

`bool`

Whether Azure prunes flow-log files after `days`. False keeps
files forever (delete them yourself or via storage lifecycle
policy).

### spec.retentionPolicy.days

`int32`

Days each file is kept before deletion. 0 with enabled=false
means keep forever. Meaningful (non-zero) only when enabled.

- rule: {"int32":{"gte":0}}

### spec.trafficAnalytics

`AzureNetworkWatcherFlowLogTrafficAnalytics`

Traffic Analytics: pipe the flow records into a Log Analytics
workspace for query, topology, and threat views. Leave unset to
write raw files only.

### spec.trafficAnalytics.enabled

`bool` · optional (explicit presence)

Whether Traffic Analytics processing is on. Unspecified applies
true -- a configured block exists to analyze; set false to pause
processing while keeping the wiring.

- default: `true`

### spec.trafficAnalytics.workspaceId

`string | valueFrom` · required

The Log Analytics workspace's GUID (its "customer id", NOT the ARM
resource ID) -- references an AzureLogAnalyticsWorkspace's
workspace_customer_id output.

- references: AzureLogAnalyticsWorkspace (`status.outputs.workspace_customer_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLogAnalyticsWorkspace, name: <that resource's name>, fieldPath: status.outputs.workspace_customer_id}} -- a bare string does not parse

### spec.trafficAnalytics.workspaceRegion

`string` · required

The region the Log Analytics workspace lives in (which may differ
from the flow log's region), e.g. "eastus".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.trafficAnalytics.workspaceResourceId

`string | valueFrom` · required

The Log Analytics workspace's ARM resource ID -- references an
AzureLogAnalyticsWorkspace's workspace_id output.

- references: AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLogAnalyticsWorkspace, name: <that resource's name>, fieldPath: status.outputs.workspace_id}} -- a bare string does not parse

### spec.trafficAnalytics.intervalInMinutes

`int32` · optional (explicit presence)

How often Traffic Analytics processes accumulated flow logs: every
60 minutes (the default) or every 10 (faster insight, higher
workspace ingestion). Unspecified applies 60.

- default: `60`
- rule: {"int32":{"in":[10,60]}}

### spec.networkWatcherName

`string`

The Network Watcher the flow log attaches to, by name. Leave unset
to use the region's AUTO-CREATED watcher ("NetworkWatcher_{region}"
-- the shape virtually every subscription runs); set it (with
network_watcher_resource_group) only for a self-managed watcher.
Fixed at creation.

### spec.networkWatcherResourceGroup

`string | valueFrom`

The resource group the Network Watcher lives in (NOT the target's
resource group). Leave unset to use "NetworkWatcherRG", the group
Azure auto-creates its watchers in; set it together with
network_watcher_name for a self-managed watcher. Fixed at
creation.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Free-form tags applied to the flow log, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins.

## Validation Rules

- `nsg_targets_are_retired`: Azure no longer accepts NEW NSG-targeted flow logs (retired 2025-06-30; NSG flow logs retire fully 2027-09-30) -- target the virtual network, subnet, or network interface instead
- `custom_watcher_addressed_completely`: network_watcher_name and network_watcher_resource_group are set together (leave BOTH unset to use the region's auto-created watcher)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureNetworkWatcherFlowLog, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.flow_log_id` | `string` | The flow log's ARM resource ID (.../networkWatchers/{watcher}/flowLogs/{name}). |
| `status.outputs.flow_log_name` | `string` | The flow log's name within its Network Watcher. |
| `status.outputs.network_watcher_name` | `string` | The Network Watcher the flow log attached to -- the auto-created regional singleton unless the spec addressed a self-managed one. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.trafficAnalytics.workspaceId` | AzureLogAnalyticsWorkspace | `status.outputs.workspace_customer_id` |
| `spec.trafficAnalytics.workspaceResourceId` | AzureLogAnalyticsWorkspace | `status.outputs.workspace_id` |
| `spec.networkWatcherResourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |

## See Also

- [Overview](../README.md)
