# Azure Storage Queue

Deploys a Storage queue inside an Azure Storage Account -- the simple, massive-scale message buffer of Azure storage. Producers enqueue messages up to 64 KB and workers poll-and-delete them: the classic work-queue and load-leveling pattern Functions queue triggers, WebJobs, and worker pools consume. For pub/sub topics, sessions, ordering guarantees, or dead-lettering semantics, Service Bus is the richer sibling -- Storage queues win on cost, capacity, and operational simplicity. Queues are many-per-account with independent lifecycles, which is why they are a first-class kind referencing the account rather than a list folded into the account's spec.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Storage Queue** -- a queue on the referenced storage account (by ARM ID -- the control-plane path), with your data-plane metadata

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureStorageAccount** the queue will live in, referenced through `storageAccountId`. The parent is fixed at creation: a queue cannot move between accounts.

## Deploy

### Console

Open the deployment store, find **Azure Storage Queue**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Work Queue** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureStorageQueue
metadata:
  name: work-items
  org: acme-corp
  env: prod
spec:
  storageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: app-storage
      fieldPath: status.outputs.storage_account_id
  queueName: work-items
```

```shell
planton apply -f queue.yaml
```

This creates an empty queue named `work-items` on the referenced account -- messages hold up to 64 KB, and the queue holds as many as the account's capacity allows. An Infra Job tracks the provisioning in real time.

### InfraChart

When the account and queue deploy in the same InfraChart, wire the account reference with ValueFromRef:

```yaml
spec:
  storageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: app-storage
      fieldPath: status.outputs.storage_account_id
  queueName: work-items
```

The InfraPipeline resolves the dependency graph, deploys the storage account first, then provisions the queue with the resolved ARM ID.

## Key Configuration

These are the most important decisions when configuring a queue. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Queue name** -- `queueName` becomes the URL path segment under the account's queue endpoint. 3-63 lowercase letters, digits, and hyphens; unique within the account. Renaming replaces the queue. The common companion is a second queue named `{name}-poison`, where Functions moves messages that repeatedly fail processing.

**Metadata** -- `metadata` stores free-form key/value pairs on the queue (visible to anyone who can read queue properties -- not for secrets). Keys must be valid C# identifiers; Azure lowercases them on read.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureStorageAccount** | `storageAccountId` | `status.outputs.storage_account_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `queue_id` | Azure Resource Manager ID of the queue | Queue-scoped data-plane role assignments (Storage Queue Data Contributor / Message Processor / Message Sender) |
| `queue_name` | The queue's name | SDK clients, Functions queue triggers, app settings |
| `storage_account_name` | The parent account's name, parsed from the resolved account ID | The account/queue pair without a second reference |

There is deliberately NO URL output: the queue's data-plane URL is the ACCOUNT's queue endpoint plus the queue name, and only the account knows its real endpoint (partitioned-DNS accounts use a different hostname than the classic shared DNS). Compose client URLs from AzureStorageAccount's `primary_queue_endpoint` output + `queue_name`.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Work queue** -- the everyday shape: producers enqueue with Message Sender grants, workers process with Message Processor grants — least privilege per direction, no account keys. Start from the **Work Queue** preset.

**Poison companion** -- the `{name}-poison` sibling Functions moves repeatedly failing messages to. Start from the **Poison-Queue Companion** preset.

**Ingest buffer** -- a load-leveling buffer absorbing bursty producers ahead of steady workers. Start from the **Ingest Buffer Queue** preset.

## Works With

- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the parent account and the source of the queue endpoint client URLs compose from
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- queue-scoped data-plane grants targeting `queue_id`
- [**Azure Function App**](/infra-catalog/azure-function-app) -- queue triggers consume messages by queue name
