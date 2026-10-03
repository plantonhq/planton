# Azure MSSQL Server

Deploys an Azure SQL Database logical server — the administrative boundary the whole SQL family hangs off. The server owns authentication (SQL logins, a Microsoft Entra administrator, or Entra-only), network access (the public endpoint, IP firewall rules, VNet rules), encryption (Microsoft-managed or a customer-managed TDE key), auditing, and threat protection. Databases and elastic pools are first-class kinds (AzureMssqlDatabase, AzureMssqlElasticPool) that reference this server's `server_id` output — nothing is embedded in this spec.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SQL Logical Server** -- an administrative endpoint in the specified Azure region and resource group with a globally-unique DNS name (`{serverName}.database.windows.net`), the chosen authentication posture, TLS floor, and connection policy
- **Microsoft Entra Administrator** -- created when `azureadAdministrator` is set; a directory principal (user, group, or managed identity) granted the server's administrator role, optionally with Entra-only authentication
- **Managed Identity** -- created when `identity` is set; the server's own Entra identity (system-assigned, user-assigned, or both) — what unwraps a customer-managed TDE key
- **Firewall, VNet, and Outbound Rules** -- created when `firewallRules` / `virtualNetworkRules` / `outboundFirewallRules` entries are provided; IP ranges and service-endpoint subnets admitted to the public endpoint, and the FQDNs the server may reach out to under outbound restriction
- **Auditing and Threat Protection** -- created when `extendedAuditing` / `securityAlertPolicy` are set; the server-wide audit trail and Defender for SQL alerting
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the server for tracking and governance

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the SQL Server will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A globally unique server name** -- `serverName` becomes the server hostname (`{serverName}.database.windows.net`). 1-63 lowercase letters, digits, and hyphens, starting and ending with a letter or digit.
- **Network access planning** -- decide between public access (firewall/VNet rules) or private-only access (AzurePrivateEndpoint with `publicNetworkAccessEnabled: false`). Unlike PostgreSQL/MySQL Flexible Servers, Azure SQL does not support VNet delegation.

## Deploy

### Console

Open the deployment store, find **Azure MSSQL Server**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard SQL-Auth Server** preset in the [Presets](#presets) tab for the everyday SQL-auth server.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMssqlServer
metadata:
  name: app-sql
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  serverName: acme-app-sql-prod
  administratorLogin: sqladmin
  administratorPassword:
    value: "$secret/sql-admin-password"
  firewallRules:
    - name: allow-azure-services
      startIpAddress: "0.0.0.0"
      endIpAddress: "0.0.0.0"
```

```shell
planton apply -f mssql-server.yaml
```

This creates a SQL logical server with SQL authentication, public access gated by the Azure-services firewall sentinel, and Azure's defaults everywhere else (engine 12.0, TLS 1.2, Default connection policy). Databases attach afterwards as AzureMssqlDatabase resources referencing this server. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the SQL server to a resource group deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the SQL server with the resolved value.

## Key Configuration

These are the most important decisions when configuring a SQL server. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Authentication posture** -- SQL auth (`administratorLogin` + `administratorPassword`, set together), a Microsoft Entra administrator (`azureadAdministrator` with the principal's OBJECT id — for a managed identity that is its `principal_id`), or the passwordless Entra-only posture (`azureadAuthenticationOnly: true`, which forbids the SQL credential pair entirely).

**Public vs. private access** -- Set `publicNetworkAccessEnabled: false` to restrict all access to private endpoints (a separate AzurePrivateEndpoint targeting `server_id`). When public, `firewallRules` admit IPv4 ranges (the `0.0.0.0`-`0.0.0.0` sentinel means "allow Azure services") and `virtualNetworkRules` admit whole subnets carrying the Microsoft.Sql service endpoint.

**Customer-managed TDE** -- `transparentDataEncryptionKeyVaultKeyId` points Transparent Data Encryption at a Key Vault key you own (the VERSIONED key id). Requires an `identity` block — the identity unwraps the key — with `primaryUserAssignedIdentityId` designating which one.

**Connection policy** -- `connectionPolicy` controls how clients connect: unset applies Azure's Default (Redirect inside Azure, Proxy outside); `REDIRECT` gives lowest latency (requires ports 11000-11999 outbound at clients); `PROXY` routes everything through the gateway.

**Auditing and threat protection** -- `extendedAuditing` captures the server-wide audit trail (Azure Monitor via log monitoring, and/or a storage account); `securityAlertPolicy` turns on Defender for SQL anomaly alerts; `expressVulnerabilityAssessmentEnabled` adds the zero-configuration scanner.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureUserAssignedIdentity** | `identity.identityIds`, `primaryUserAssignedIdentityId` | `status.outputs.identity_id` |
| **AzureUserAssignedIdentity** | `azureadAdministrator.objectId` | `status.outputs.principal_id` |
| **AzureKeyVaultKey** | `transparentDataEncryptionKeyVaultKeyId` | `status.outputs.key_id` |
| **AzureSubnet** | `virtualNetworkRules[].subnetId` | `status.outputs.subnet_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `server_id` | Azure resource ID of the SQL Server | AzureMssqlDatabase / AzureMssqlElasticPool / AzureMssqlFailoverGroup `server_id`, AzurePrivateEndpoint targets, diagnostic settings |
| `fqdn` | Fully qualified domain name (`{serverName}.database.windows.net`) | Application connection strings |
| `administrator_login` | SQL administrator login (empty on Entra-only servers) | Application connection strings |
| `identity_principal_id` | Principal ID of the system-assigned identity (when enabled) | RBAC role assignments — the grant target |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard SQL auth** -- The simplest useful server: SQL authentication, the public endpoint with the Azure-services firewall sentinel, and Azure's defaults everywhere else. Start from the **Standard SQL-Auth Server** preset.

**Entra-only** -- The passwordless posture: a Microsoft Entra administrator (grant a group so the DBA team manages itself in the directory) with `azureadAuthenticationOnly: true` and no SQL credential pair. Start from the **Entra-Only Server with Microsoft Defender** preset.

**Private and hardened** -- Public access off (pair an AzurePrivateEndpoint), a user-assigned identity unwrapping a customer-managed TDE key, and the explicit TLS 1.2 pin. Start from the **Private Hardened Server with CMK and Auditing** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the SQL Server is created
- [**Azure MSSQL Database**](/infra-catalog/azure-mssql-database) -- databases attach to this server via its `server_id` output
- [**Azure MSSQL Elastic Pool**](/infra-catalog/azure-mssql-elastic-pool) -- shared-capacity pools attach via `server_id` (same region as the server)
- [**Azure MSSQL Failover Group**](/infra-catalog/azure-mssql-failover-group) -- pairs this server with a partner for cross-region DR
- [**Azure Private Endpoint**](/infra-catalog/azure-private-endpoint) -- private connectivity when public access is off
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the customer-managed TDE key
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the identity that unwraps the TDE key and the Entra administrator principal
