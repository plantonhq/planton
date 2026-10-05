# Azure Event Grid System Topic

Deploys an Azure Event Grid system topic -- the subscription surface for events Azure itself publishes about one of your resources: a storage account announcing blob creations, a resource group announcing resource writes, a Key Vault announcing secret expiries. Where a custom topic receives events your application posts, a system topic exposes a stream an Azure service already emits; attach event subscriptions to route it to handlers. Azure allows exactly one system topic per source resource per topic type, and a system topic is free at rest -- billing is per operation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Event Grid system topic** -- the source binding that makes an Azure service's built-in event stream subscribable: the `(sourceResourceId, topicType)` pair is the topic's identity, with an optional managed identity for secured delivery
- **Azure Tags** -- Planton-derived metadata tags merged with the manifest's `tags` (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Azure Subscription

- **A resource group for the topic** -- this is where the TOPIC resource sits; the source resource may live in a different group. Choose a platform-owned group: the topic is shared infrastructure (see Key Configuration).
- **The source resource** -- the storage account, Key Vault, or other resource whose events you want; reference its ID output with an explicit valueFrom or pass a literal ARM ID. No system topic may already exist on that source for the same topic type -- a second create fails with a conflict.
- **The region must match the source's region** -- global sources (Azure subscriptions via `Microsoft.Resources.Subscriptions`, resource groups via `Microsoft.Resources.ResourceGroups`) require `Global`.

## Deploy

### Console

Open the deployment store, find **Azure Event Grid System Topic**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields: the resource group, region, source binding, and identity. Start from the **Storage Account Events** or **Resource Group Events** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventgridSystemTopic
metadata:
  name: storage-events
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-events
      fieldPath: status.outputs.resource_group_name
  name: app-storage-events
  region: eastus
  sourceResourceId:
    valueFrom:
      kind: AzureStorageAccount
      name: app-storage
      fieldPath: status.outputs.storage_account_id
  topicType: Microsoft.Storage.StorageAccounts
```

```shell
planton apply -f system-topic.yaml
```

This binds a storage account's built-in event stream (blob and queue events) to a subscribable topic; until a subscription is attached, the source's events are evaluated and dropped. An Infra Job tracks the provisioning in real time.

### InfraChart

When the source, the topic, and its subscriptions deploy in the same InfraChart, wire them with ValueFromRef:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-events
      fieldPath: status.outputs.resource_group_name
  name: app-storage-events
  region: eastus
  sourceResourceId:
    valueFrom:
      kind: AzureStorageAccount
      name: app-storage
      fieldPath: status.outputs.storage_account_id
  topicType: Microsoft.Storage.StorageAccounts
```

The InfraPipeline resolves the dependency graph, deploys the resource group and storage account first, then provisions the topic -- and an Azure Event Grid Event Subscription downstream can reference this topic's `system_topic_id` in the same chart.

## Key Configuration

These are the most important decisions when configuring a system topic. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The topic is a singleton claim on its source** -- Azure allows exactly one system topic per source resource per topic type, which makes the topic SHARED INFRASTRUCTURE in a multi-team subscription: the first team to create it owns its placement and lifecycle, and every other team attaches subscriptions to it. Place it in a platform-owned resource group rather than an app team's group that might get deleted, and never let two pipelines race to create topics on the same source -- the loser fails with a conflict, not a merge.

**Every identity field is create-only** -- `sourceResourceId`, `topicType`, `name`, `region`, and `resourceGroup` are all ForceNew, and a replaced topic silently drops every subscription attached to it, including other teams'. Treat topic replacement like a coordinated migration: enumerate subscriptions first (`az eventgrid system-topic event-subscription list`), recreate them after.

**Region is inherited, not chosen** -- the topic must sit in its source's region; there is no placement decision, just a contract to satisfy. Global sources (subscriptions, resource groups) emit from Azure's control plane, so their topics take `Global`. A mismatch fails at deploy time.

**topicType names a catalog entry, not a promise** -- the value must match the source's service (`Microsoft.Storage.StorageAccounts`, `Microsoft.Resources.ResourceGroups`, `Microsoft.KeyVault.vaults`, ...; list them with `az eventgrid topic-type list`), and not every resource type emits every event -- storage accounts emit blob and queue events but not table events. Check the type's event schema (`az eventgrid topic-type list-event-types --name <type>`) before designing a pipeline around an event that does not exist.

**Enable identity before the subscriptions need it** -- subscriptions deliver AS the topic's identity when targeting locked-down destinations (a storage queue with shared keys disabled). The identity must exist on the topic BEFORE such a subscription is created, and its data-plane grants on the destination must land first too. A system topic supports both flavors at once -- the combined mode exists precisely for migrations: keep the old identity granted while the new one rolls out.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureStorageAccount** (or any source kind, via explicit valueFrom) | `sourceResourceId` | `status.outputs.storage_account_id` |
| **AzureUserAssignedIdentity** (optional, per identity) | `identity.identityIds` | `status.outputs.identity_id` |

Sources span dozens of kinds, so `sourceResourceId` carries no default -- reference the owning kind's ID output explicitly or pass a literal ARM ID.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `system_topic_id` | The system topic's ARM ID | An Azure Event Grid Event Subscription's `systemTopicId` reference |
| `metric_resource_id` | The GUID-style identifier Azure Monitor uses for the topic's metrics (not an ARM ID) | Wiring metric alerts to the topic's delivery/drop counters |
| `identity_principal_id` | The system-assigned identity's principal ID (empty when none is configured) | Role assignments granting the topic identity-based access on delivery targets |

`system_topic_name` is also exported but only echoes the manifest's `name` back.

## Common Patterns

**Storage account events** -- the most common source: expose BlobCreated/BlobDeleted (and queue) events for ingest pipelines, thumbnail generation, or audit without polling the account. Start from the **Storage Account Events** preset.

**Resource group governance** -- expose resource-lifecycle events (write, delete, action success/failure) for audit trails and drift automation, with region `Global` and subscriptions filtering `includedEventTypes` down to the operations that matter. Start from the **Resource Group Events** preset.

**Shared topic, per-team subscriptions** -- one platform-owned topic per source, with each consuming team attaching its own filtered subscription; the singleton constraint makes any other shape a deploy-time conflict.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the group the topic resource sits in
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the most common event source (blob and queue events)
- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- a source for secret and certificate lifecycle events
- [**Azure Event Grid Event Subscription**](/infra-catalog/azure-eventgrid-event-subscription) -- routes the topic's events to handlers via `systemTopicId`
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- identities attached for identity-based delivery
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- grants the topic's principal data-plane access on delivery targets
- [**Azure Monitor Metric Alert**](/infra-catalog/azure-monitor-metric-alert) -- alerts on the topic's delivery and drop counters
