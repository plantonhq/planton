# Azure Storage Table

Deploys a Storage table inside an Azure Storage Account -- the serverless NoSQL key/value store of Azure storage. Applications store schemaless entities addressed by partition key + row key -- device state, user profiles, audit trails, IoT telemetry -- at petabyte scale with single-digit-millisecond point reads and no capacity provisioning. Cosmos DB's Table API is the premium sibling (global distribution, throughput SLAs) at a very different price point. Tables are many-per-account with independent lifecycles, which is why they are a first-class kind referencing the account rather than a list folded into the account's spec.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Storage Table** -- a table on the referenced storage account (by ARM ID -- the control-plane path), with optional stored access policies (signed identifiers) anchoring revocable SAS tokens

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureStorageAccount** the table will live in, referenced through `storageAccountId`. The parent is fixed at creation: a table cannot move between accounts.
- **Shared-key access on the account** -- the provider drives table creation and stored access policies through the table DATA PLANE with shared-key authorization, so the account must keep `sharedAccessKeyEnabled: true` (Azure's default) for deploys to work.

## Deploy

### Console

Open the deployment store, find **Azure Storage Table**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Application Entities Table** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureStorageTable
metadata:
  name: device-state
  org: acme-corp
  env: prod
spec:
  storageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: app-storage
      fieldPath: status.outputs.storage_account_id
  tableName: DeviceState
```

```shell
planton apply -f table.yaml
```

This creates a table with no stored access policies -- applications reach it with data-plane RBAC or account keys. An Infra Job tracks the provisioning in real time.

### InfraChart

When the account and table deploy in the same InfraChart, wire the account reference with ValueFromRef:

```yaml
spec:
  storageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: app-storage
      fieldPath: status.outputs.storage_account_id
  tableName: DeviceState
```

The InfraPipeline resolves the dependency graph, deploys the storage account first, then provisions the table with the resolved ARM ID.

## Key Configuration

These are the most important decisions when configuring a table. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Table name** -- `tableName` is 3-63 alphanumeric characters starting with a letter -- no hyphens or underscores (stricter than blob containers), and never the literal word "table" (Azure reserves it). Unique within the account. Renaming replaces the table.

**Stored access policies** -- `acls` holds up to five signed identifiers (Azure's limit). Each policy anchors shared-access-signature tokens: revoking or shortening the policy immediately revokes every SAS issued against it -- the operational reason to prefer policy-anchored SAS over ad-hoc SAS. Table policies carry the TABLE grammar (`r` read/query, `a` add, `u` update, `d` delete -- in strict order), and unlike blob/share policies the window is ALL-REQUIRED: start, expiry, and permissions are all mandatory.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureStorageAccount** | `storageAccountId` | `status.outputs.storage_account_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `table_id` | Azure Resource Manager ID of the table | Table-scoped data-plane role assignments (Storage Table Data Reader/Contributor) |
| `table_name` | The table's name | SDK clients, function bindings, app settings |
| `storage_account_name` | The parent account's name, parsed from the resolved account ID | The account/table pair without a second reference |

There is deliberately NO URL output: the table's data-plane address is the ACCOUNT's table endpoint plus the table name, and only the account knows its real endpoint (partitioned-DNS accounts use a different hostname than the classic shared DNS).

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Application entities** -- the default shape: a table per entity domain (DeviceState, UserProfiles), applications reaching it with role assignments scoped to `table_id`. Start from the **Application Entities Table** preset.

**Append-mostly audit trail** -- a table capturing who-did-what rows keyed by time; consumers query, writers add. Start from the **Audit Trail Table** preset.

**Policy-anchored partner access** -- a query-only stored access policy (`r`) with a bounded window; ending the engagement is shortening one expiry. Start from the **Policy-Anchored Read-Only Table** preset.

## Works With

- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the parent account and the source of the table endpoint clients compose addresses from
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- table-scoped data-plane grants targeting `table_id`
- [**Azure Cosmos DB Account**](/infra-catalog/azure-cosmosdb-account) -- the premium sibling when the workload outgrows Table Storage's SLAs
