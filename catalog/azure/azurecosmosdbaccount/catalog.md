# Azure Cosmos DB Account

Deploys an Azure Cosmos DB account — the globally distributed, multi-model database account that owns regions, consistency, network posture, encryption, and backup for everything stored inside it. Databases and containers are first-class kinds (AzureCosmosdbSqlDatabase / AzureCosmosdbSqlContainer for the SQL API, AzureCosmosdbMongoDatabase / AzureCosmosdbMongoCollection for MongoDB) that reference this account's `cosmosdb_account_id` output — nothing is embedded in this spec. The account is the governance boundary: which regions data lives in, how reads are ordered, who can reach it, how it is encrypted, and how it is backed up are all decided here and inherited by everything inside.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cosmos DB Account** -- a globally distributed database account in the specified Azure region and resource group, with a globally-unique DNS endpoint (`https://{accountName}.documents.azure.com`), the chosen API kind (SQL/NoSQL or MongoDB), consistency policy, geo-locations, capabilities, backup policy, and network access controls
- **Managed Identity** -- created when `identity` is set; the account's own Entra identity (system-assigned, user-assigned, or both) — what unwraps a customer-managed encryption key
- **Virtual Network Rules** -- created when `virtualNetworkRules` entries are provided; restricts account access to specified subnets carrying the Microsoft.AzureCosmosDB service endpoint
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the account for tracking and governance

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the Cosmos DB account will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A globally unique account name** -- `accountName` becomes the endpoint hostname (`https://{accountName}.documents.azure.com`). Must be 3-50 characters, lowercase letters, numbers, and hyphens only.
- **Subnets with service endpoints** (optional) -- required when using virtual network filtering. Each subnet must have the `Microsoft.AzureCosmosDB` service endpoint enabled.
- **A Key Vault key** (optional) -- when using customer-managed encryption, reference an AzureKeyVaultKey's `versionless_id` output. The vault must have purge protection enabled.

## Deploy

### Console

Open the deployment store, find **Azure Cosmos DB Account**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Production SQL API Account** preset in the [Presets](#presets) tab for the most common configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureCosmosdbAccount
metadata:
  name: orders-db
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  accountName: acme-orders-prod
  consistencyPolicy:
    consistencyLevel: SESSION
  geoLocations:
    - location: eastus
      failoverPriority: 0
```

```shell
planton apply -f cosmosdb-account.yaml
```

This creates a Cosmos DB account with the SQL API (default), Session consistency, a single geo-location, and Azure's defaults everywhere else. Databases attach afterwards as AzureCosmosdbSqlDatabase resources referencing this account. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Cosmos DB account to a resource group deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the Cosmos DB account with the resolved value.

## Key Configuration

These are the most important decisions when configuring a Cosmos DB account. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**API kind** -- The `kind` field selects the wire protocol: unset / `GLOBAL_DOCUMENT_DB` for SQL-like queries over JSON documents (and the base for Cassandra, Gremlin, and Table via capabilities), or `MONGO_DB` for MongoDB driver compatibility. This is a ForceNew field — changing it destroys and recreates the account.

**Consistency** -- Five well-defined levels from Strong (linearizable) to Eventual. Session — read-your-writes within a session — is Azure's recommendation for most applications and the default when unset.

**Global distribution** -- `geoLocations` declares the regions data lives in. Priority 0 is the write region; higher numbers promote on failover. Adding and removing regions is an in-place update. Set `automaticFailoverEnabled` for unattended promotion, or `multipleWriteLocationsEnabled` for active-active writes (with per-container conflict resolution).

**Capabilities** -- Customize what the account can do: serverless billing (`ENABLE_SERVERLESS`), extra APIs on a SQL account, MongoDB feature switches, vector/full-text search. Most capability changes recreate the account — settle them before production.

**Network posture** -- `publicNetworkAccessEnabled: false` restricts all access to private endpoints (AzurePrivateEndpoint targeting `cosmosdb_account_id`, subresource `"Sql"` or `"MongoDB"`). When public, `isVirtualNetworkFilterEnabled` + `virtualNetworkRules` and `ipRangeFilter` scope who can reach the account.

**Customer-managed encryption** -- `keyVaultKeyId` points encryption at a Key Vault key you own (the VERSIONLESS key id so rotation propagates). Fixed at creation. Pair with a user-assigned `defaultIdentity` so the key can be unwrapped before the account's own identity exists.

**Backup** -- Unset means Periodic (every 4 hours, retained 8 hours, geo-redundant). Continuous enables point-in-time restore and is required for `createMode: RESTORE`. Periodic → Continuous upgrades in place; Continuous → Periodic recreates the account.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| AzureResourceGroup | `resourceGroup` | `status.outputs.resource_group_name` |
| AzureSubnet | `virtualNetworkRules[].subnetId` | `status.outputs.subnet_id` |
| AzureKeyVaultKey | `keyVaultKeyId` | `status.outputs.versionless_id` |
| AzureUserAssignedIdentity | `identity.identityIds[]`, `defaultIdentity.userAssignedIdentityId` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cosmosdb_account_id` | The ARM ID of the account | AzureCosmosdbSqlDatabase / AzureCosmosdbMongoDatabase `cosmosdbAccountId`, AzurePrivateEndpoint targets |
| `cosmosdb_account_name` | The globally unique DNS label | Connection string composition, scripting |
| `endpoint` | The document endpoint SDKs connect to | Application configuration |
| `read_endpoints` | Per-region read endpoints, ordered by failover priority | Latency-sensitive readers pinning to a region |
| `write_endpoints` | Per-region write endpoints | Multi-write clients |
| `primary_key` | The primary read-write account key (secret-bearing) | Legacy key-auth clients; `secondary_key` is the rotation partner |
| `primary_sql_connection_string` | Ready-made SQL-API connection string (secret-bearing; secondary and read-only variants also exported) | SQL SDK clients |
| `primary_mongodb_connection_string` | Ready-made MongoDB connection string (secret-bearing; secondary and read-only variants also exported) | MongoDB drivers on MONGO_DB accounts |
| `identity_principal_id` | The system-assigned identity's principal ID | Role assignments the account needs against other services |

When `localAuthenticationEnabled` is false, the keys and connection strings stop authenticating — data-plane access rides Entra ID instead.

## Common Patterns

**Production document store that survives a regional outage** — two `geoLocations` with `automaticFailoverEnabled`, Session consistency, and Continuous backup for point-in-time restore. Regions can be added in place later, but Continuous cannot go back to Periodic without recreating the account — the one-way door to decide up front. Start from the **Production SQL API Account** preset.

**MongoDB migration without driver changes** — `kind: MONGO_DB` with `ENABLE_MONGO` declared explicitly and `mongoServerVersion` pinned to what your drivers support. The API kind is fixed at creation; the account exports ready-made MongoDB connection strings applications consume directly. Start from the **MongoDB API Account** preset.

**Serverless, keyless, private** — `ENABLE_SERVERLESS` bills per request (databases and containers must not declare throughput, and serverless accounts are single-region), `publicNetworkAccessEnabled: false` restricts access to private endpoints, and `localAuthenticationEnabled: false` forces every data-plane caller through Entra ID. Right for dev/staging with idle periods and for regulated workloads; wrong for steady high-throughput production, where provisioned RU/s is cheaper. Start from the **Serverless, Entra-Only Account** preset.

**Customer-managed encryption** — `keyVaultKeyId` with a user-assigned `defaultIdentity` so the key can be unwrapped before the account's own identity exists. The key choice is fixed at creation; the vault needs purge protection, and the identity needs get/wrapKey/unwrapKey on the key.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) — the resource group the account is created in
- [**Azure Cosmos DB SQL Database**](/infra-catalog/azure-cosmosdb-sql-database) / [**Azure Cosmos DB SQL Container**](/infra-catalog/azure-cosmosdb-sql-container) — the SQL-API data plane, referencing `cosmosdb_account_id`
- [**Azure Cosmos DB Mongo Database**](/infra-catalog/azure-cosmosdb-mongo-database) / [**Azure Cosmos DB Mongo Collection**](/infra-catalog/azure-cosmosdb-mongo-collection) — the MongoDB-API data plane on MONGO_DB accounts
- [**Azure Cosmos DB SQL Role Definition**](/infra-catalog/azure-cosmosdb-sql-role-definition) / [**Azure Cosmos DB SQL Role Assignment**](/infra-catalog/azure-cosmosdb-sql-role-assignment) — data-plane RBAC for the keyless posture
- [**Azure Private Endpoint**](/infra-catalog/azure-private-endpoint) — private connectivity when public network access is disabled (subresource `Sql` or `MongoDB`)
- [**Azure Subnet**](/infra-catalog/azure-subnet) — subnets admitted through the virtual-network filter
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) — the customer-managed encryption key, referenced by versionless ID
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) — the identity that unwraps the customer-managed key
