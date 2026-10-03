# AzureMssqlServer

Azure SQL Database's logical server: the administrative container that
carries the login endpoint, SQL and Microsoft Entra authentication, the
networking posture, the transparent-data-encryption key, auditing, and
Microsoft Defender settings. It has no compute of its own -- databases
(AzureMssqlDatabase) and elastic pools (AzureMssqlElasticPool) are
first-class resources created on it, each carrying its own SKU and
billing.

## When to Use

Use AzureMssqlServer when you need:

- **The container for Azure SQL databases and elastic pools** -- every
  database estate starts with one
- **Entra-only (passwordless) authentication** -- disable SQL logins
  server-wide via the Entra administrator
- **A customer-managed TDE key** composed against `AzureKeyVaultKey`,
  protecting every database on the server
- **Server-wide auditing and Microsoft Defender** threat detection
- **Network governance** -- the public-endpoint dial, IPv4 firewall
  rules, subnet service-endpoint rules, and outbound FQDN restriction

## Key Configuration

### Authentication (at least one mechanism)

- **SQL auth**: `administrator_login` + `administrator_password` (the
  login is fixed once set)
- **Microsoft Entra**: `azuread_administrator` (use a group to admit a
  team); with `azuread_authentication_only` SQL logins are disabled
  server-wide and the SQL credentials must be omitted
- Mixed mode (both) is legal and common during migrations

### Networking

Azure SQL does not support VNet delegation. The dials, matching ARM's
contract:

- `public_network_access_enabled` (default true) + `firewall_rules`
  (IPv4 allowlist; 0.0.0.0-0.0.0.0 admits Azure-internal services) +
  `virtual_network_rules` (subnet service-endpoint allowlist)
- Private connectivity is exclusively **AzurePrivateEndpoint**
  (subresource `sqlServer`) against the `server_id` output
- `outbound_network_restriction_enabled` + `outbound_firewall_rules`
  govern where the server reaches OUT to (elastic queries)

### Encryption, Auditing, Defender

- `transparent_data_encryption_key_vault_key_id` -- a VERSIONED
  `AzureKeyVaultKey.key_id` (ARM pins the version at the server level);
  requires an identity with wrap/unwrap on the vault
- `extended_auditing` -- every database's audit events to blob storage
  and/or Azure Monitor
- `security_alert_policy` + `express_vulnerability_assessment_enabled`
  -- Microsoft Defender for SQL

## Fields That Replace the Server When Changed

`server_name`, `region`, `resource_group`, `version`, and
`administrator_login` once set. Update-time contracts ARM enforces:
`minimum_tls_version` cannot be removed once set, and the password
cannot change while Entra-only auth is on.

## Outputs

| Output | Description |
|--------|-------------|
| `server_id` | ARM ID -- the parent seam for AzureMssqlDatabase, AzureMssqlElasticPool, and AzurePrivateEndpoint |
| `server_name` | Server name |
| `fqdn` | `{name}.database.windows.net` -- the connection-string host |
| `administrator_login` | Admin login (empty on Entra-only servers) |
| `identity_principal_id` | System-assigned identity's principal -- the role-assignment seam |

**Connection string format:**

```text
Server={fqdn},1433;Database={database};User ID={admin};Password={password};Encrypt=True;
```

## Related Resources

- **AzureResourceGroup** -- the server's container
- **AzureMssqlDatabase** -- databases on the server (the unit of compute)
- **AzureMssqlElasticPool** -- shared compute databases can join
- **AzurePrivateEndpoint** -- private connectivity (subresource `sqlServer`)
- **AzureSubnet** -- subnets admitted via service-endpoint rules
- **AzureUserAssignedIdentity** -- the TDE unwrap identity / Entra administrator principal
- **AzureKeyVaultKey** -- the customer-managed TDE key

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
