# Azure Cosmos DB SQL Role Assignment

Grants a Cosmos DB data-plane role to a Microsoft Entra principal at a scope inside one SQL (NoSQL) API account. Cosmos DB carries its own RBAC system, separate from ARM RBAC — this assignment is how an Entra identity (a workload's managed identity, a CI principal, an operator) gets data access. With the account's local (key) authentication disabled, grants like this are the ONLY way clients connect: the fully keyless posture. Every grant has three coordinates: WHAT (`roleDefinitionId` — a built-in role by well-known ID, or a custom AzureCosmosdbSqlRoleDefinition by reference), WHO (`principalId` — the Entra OBJECT ID), and WHERE (`scope` — the account, one database, or one container; permissions inherit downward).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cosmos DB SQL Role Assignment** -- a GUID-identified grant record binding the role to the principal at the scope

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Cosmos DB account** speaking the SQL (NoSQL) API.
- **The principal's OBJECT ID** — for a Planton-managed identity, reference the AzureUserAssignedIdentity's `principal_id` output; for users, groups, or external service principals, the object ID from Entra. The object ID is NOT the application (client) ID — an assignment authored with a client ID deploys successfully and grants nothing.
- **For custom roles**: an AzureCosmosdbSqlRoleDefinition whose assignable scopes cover the grant's scope.

## Deploy

### Console

Open the deployment store, find **Azure Cosmos DB SQL Role Assignment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Workload Data Contributor** preset in the [Presets](#presets) tab for the common workload grant.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureCosmosdbSqlRoleAssignment
metadata:
  name: orders-api-data-contributor
  org: acme-corp
  env: prod
spec:
  cosmosdbAccountId:
    valueFrom:
      kind: AzureCosmosdbAccount
      name: orders-db
      fieldPath: status.outputs.cosmosdb_account_id
  roleDefinitionId:
    value: "/subscriptions/.../databaseAccounts/acme-orders-prod/sqlRoleDefinitions/00000000-0000-0000-0000-000000000002"
  principalId:
    valueFrom:
      kind: AzureUserAssignedIdentity
      name: orders-api-identity
      fieldPath: status.outputs.principal_id
  scope:
    valueFrom:
      kind: AzureCosmosdbAccount
      name: orders-db
      fieldPath: status.outputs.cosmosdb_account_id
```

```shell
planton apply -f cosmosdb-sql-role-assignment.yaml
```

This grants the built-in Data Contributor account-wide to the workload's managed identity. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the account, the identity, a custom role, and this grant compose in one InfraPipeline: the pipeline resolves `cosmosdb_account_id`, `principal_id`, and `role_definition_id` into the grant in dependency order.

## Key Configuration

These are the most important decisions when configuring a role assignment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Role** -- Built-in roles (Data Reader `…0001`, Data Contributor `…0002`) exist in every account and are addressed by their well-known GUID composed on the account's ARM ID as a literal. Custom roles reference an AzureCosmosdbSqlRoleDefinition's output with zero translation. Rebinding the role is the grant's ONE in-place update.

**Scope** -- The account (the default reference), one database (`{account-id}/dbs/{db}`), or one container (`{account-id}/dbs/{db}/colls/{container}`). Sub-account scopes are literals composed on the account ID. Prefer the narrowest scope that satisfies the use case; the scope must sit at or below one of the role definition's assignable scopes.

**Principal** -- The Entra OBJECT ID. Referencing a managed identity's `principal_id` output makes the object-ID-vs-client-ID mistake unrepresentable.

**Immutability** -- Account, principal, scope, and the pinned GUID all replace the assignment; only the role binding updates in place — each grant record stays an auditable fact.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| AzureCosmosdbAccount | `cosmosdbAccountId` | `status.outputs.cosmosdb_account_id` |
| AzureCosmosdbSqlRoleDefinition | `roleDefinitionId` | `status.outputs.role_definition_id` |
| AzureUserAssignedIdentity | `principalId` | `status.outputs.principal_id` |
| AzureCosmosdbAccount | `scope` | `status.outputs.cosmosdb_account_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `role_assignment_id` | The fully-scoped ARM ID of the grant record | Audit automation that cites the grant |
| `role_assignment_guid` | The assignment's GUID resource name | Tooling that addresses the grant by name |
| `cosmosdb_account_name` | The parent account's name | Automation that needs the account/grant pair |

Nothing deploys INTO a grant — these outputs exist so automation can audit or reference the grant record without re-reading the spec.

## Common Patterns

**Keyless application grant** — the built-in Data Contributor bound to the workload's managed identity across the whole account: the app authenticates as its identity, with no keys or connection-string secrets. Right when one workload owns the account; too broad when it shares. Start from the **Workload Data Contributor** preset.

**Least-privilege reader on one container** — the built-in Data Reader scoped to exactly one container (`{account-id}/dbs/{db}/colls/{container}`): an analytics job or debugging operator sees the documents it needs and nothing else — no neighboring containers, no writes anywhere. Start from the **Container-Scoped Reader** preset.

**Custom-RBAC composition** — an AzureCosmosdbSqlRoleDefinition's fully-scoped ID flows into the grant by reference, so the definition (WHAT), this grant (WHO and WHERE), the identity, and the account deploy from one manifest set. Use it whenever the built-ins don't express the permission set. Start from the **Custom Role Grant** preset.

## Works With

- [**Azure Cosmos DB Account**](/infra-catalog/azure-cosmosdb-account) — the SQL-API account the grant lives in, and the default scope reference
- [**Azure Cosmos DB SQL Role Definition**](/infra-catalog/azure-cosmosdb-sql-role-definition) — the custom role this grant binds, referenced via `role_definition_id`
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) — the workload identity receiving the grant, referenced via `principal_id`
- [**Azure Cosmos DB SQL Database**](/infra-catalog/azure-cosmosdb-sql-database) / [**Azure Cosmos DB SQL Container**](/infra-catalog/azure-cosmosdb-sql-container) — the narrower scopes grants compose onto (`{account-id}/dbs/{db}` and `{account-id}/dbs/{db}/colls/{container}`)
