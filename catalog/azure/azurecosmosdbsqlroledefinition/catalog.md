# Azure Cosmos DB SQL Role Definition

Authors a custom DATA-PLANE role for a Cosmos DB SQL (NoSQL) API account. Cosmos DB carries its own RBAC system, separate from ARM RBAC: an ARM role on the account (even Owner) governs managing it and grants no ability to read or write the documents inside. This kind names a reusable bundle of ALLOW data actions; an AzureCosmosdbSqlRoleAssignment binds it to a principal at a scope. Together with the account's local-authentication switch, this is Cosmos DB's keyless story — keys off, workload identities in, every grant explicit and auditable.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cosmos DB SQL Role Definition** -- a named, GUID-identified custom role inside the referenced account, carrying its assignable scopes and permission blocks

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Cosmos DB account** speaking the SQL (NoSQL) API. Reference an AzureCosmosdbAccount Infra Component via ValueFromRef, or provide the account's ARM ID directly.
- **Check the built-ins first** — every SQL-API account already carries two built-in data roles: Data Reader (`00000000-0000-0000-0000-000000000001`) and Data Contributor (`00000000-0000-0000-0000-000000000002`). Assigning one needs NO definition resource: an AzureCosmosdbSqlRoleAssignment references the well-known ID directly. Author a custom definition only when neither fits — "read-only on one container", "write items but never delete", "metadata-only for a monitoring probe".

## Deploy

### Console

Open the deployment store, find **Azure Cosmos DB SQL Role Definition**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Read-Only Role** preset in the [Presets](#presets) tab for the reader shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureCosmosdbSqlRoleDefinition
metadata:
  name: orders-read-only
  org: acme-corp
  env: prod
spec:
  cosmosdbAccountId:
    valueFrom:
      kind: AzureCosmosdbAccount
      name: orders-db
      fieldPath: status.outputs.cosmosdb_account_id
  roleName: orders-read-only
  assignableScopes:
    - valueFrom:
        kind: AzureCosmosdbAccount
        name: orders-db
        fieldPath: status.outputs.cosmosdb_account_id
  permissions:
    - dataActions:
        - Microsoft.DocumentDB/databaseAccounts/readMetadata
        - Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/items/read
        - Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/executeQuery
        - Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/readChangeFeed
```

```shell
planton apply -f cosmosdb-sql-role-definition.yaml
```

This creates a read-only role assignable anywhere in the account. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the account, its roles, and their grants compose in one InfraPipeline: the pipeline deploys the account first, resolves `cosmosdb_account_id` into this definition, then resolves this definition's `role_definition_id` into its assignments.

## Key Configuration

These are the most important decisions when configuring a role definition. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Permissions** -- ALLOW rules over Cosmos data operations, evaluated as a union across blocks. No carve-outs and no control-plane actions exist in this RBAC system; the four write operations (create/replace/upsert/delete) grant individually. Practically every role includes `readMetadata` — SDKs read metadata before any data call. Editing permissions later updates EVERY existing assignment of the role at once.

**Assignable scopes** -- Where assignments of this role may be CREATED: the account, a database (`{account-id}/dbs/{db}`), or a container (`{account-id}/dbs/{db}/colls/{container}`). At least one is required; database/container entries are literal paths composed on the account ID (references cannot append path suffixes).

**Role name and GUID** -- The display name is what audits read and renames in place; assignments track the role by GUID, generated at deploy unless pinned via `roleDefinitionId`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| AzureCosmosdbAccount | `cosmosdbAccountId` | `status.outputs.cosmosdb_account_id` |
| AzureCosmosdbAccount | `assignableScopes[]` | `status.outputs.cosmosdb_account_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `role_definition_id` | The fully-scoped ARM ID of the definition | AzureCosmosdbSqlRoleAssignment `roleDefinitionId` |
| `role_definition_guid` | The definition's GUID resource name | Audit tooling |
| `role_name` | The display name as deployed | Data-plane RBAC listings |
| `cosmosdb_account_name` | The parent account's name | Automation that needs the account/definition pair |

## Common Patterns

**Read-only under your own governance** — the full read surface (point reads, SQL queries, the change feed) plus the `readMetadata` every SDK needs. It mirrors the built-in Data Reader; author it as a custom role when you want your own naming and assignable-scope narrowing rather than Microsoft's. Start from the **Read-Only Role** preset.

**Writer without delete** — the role the built-ins cannot express: full read plus create/replace/upsert, never delete. Ingest pipelines and application workloads write documents all day and have no business destroying them; reserving deletion for operators turns a whole class of bugs and compromises into authorization errors. Start from the **Writer Without Delete** preset.

**Database-scoped boundary** — a single assignable scope of one database's path means an assignment of this role at the account level, or in any other database, is rejected by Azure at apply. The boundary is enforced by the definition itself, not by review discipline on each grant — the shape for giving a team full data access to THEIR database in a shared account. Start from the **Database-Scoped Role** preset.

## Works With

- [**Azure Cosmos DB Account**](/infra-catalog/azure-cosmosdb-account) — the SQL-API account the definition lives in, and the default assignable-scope reference
- [**Azure Cosmos DB SQL Role Assignment**](/infra-catalog/azure-cosmosdb-sql-role-assignment) — the grant that binds this definition to a principal at a scope, consuming `role_definition_id`
- [**Azure Cosmos DB SQL Database**](/infra-catalog/azure-cosmosdb-sql-database) / [**Azure Cosmos DB SQL Container**](/infra-catalog/azure-cosmosdb-sql-container) — the narrower assignable scopes (`{account-id}/dbs/{db}` and `{account-id}/dbs/{db}/colls/{container}`)
