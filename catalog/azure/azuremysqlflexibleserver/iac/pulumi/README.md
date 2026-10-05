# AzureMysqlFlexibleServer - Pulumi Module

Pulumi implementation for the AzureMysqlFlexibleServer
kind.

## Architecture

```
mysql.FlexibleServer
├── mysql.FlexibleDatabase (per databases entry, Parent: server)
├── mysql.FlexibleServerFirewallRule (per firewall_rules entry, Parent: server)
├── mysql.FlexibleServerConfiguration (per server_parameters entry, Parent: server)
└── mysql.FlexibleServerActiveDirectoryAdministratory (the single aad_administrator, Parent: server)
```

The deploying credential's client configuration (`core.GetClientConfig`)
supplies the tenant fallback for the Entra administrator grant when the
spec does not pin one.

## Key Design Decisions

- **Enums map through exhaustive vocabularies** in `locals.go` (create
  mode, HA mode, public network access) -- unspecified create_mode and
  public_network_access are not sent at all, matching the Terraform
  module's null and letting Azure derive the network posture.
- **`version` is only sent for a fresh (DEFAULT) server** -- replicas and
  restores inherit the source's version, so materializing the spec default
  onto them would fight the service.
- **Presence guards on every optional-with-default field** (version,
  backup_retention_days, storage auto_grow_enabled, database
  charset/collation): IaC inputs built from a manifest do not
  materialize proto defaults, so unset falls back to the documented
  default explicitly.
- **`replication_role` is day-2 only** -- Azure rejects it at creation; the
  only legal update value ("None") promotes a replica in place.
- **The identity block renders only when identities are attached** --
  MySQL supports user-assigned identities only, so the block is always
  `UserAssigned` with the spec's resolved ARM IDs.
- **The database and firewall-rule children address the server by name +
  resource group** (MySQL's azurerm contract), while the Entra
  administrator addresses it by server ID -- each matches its azurerm
  resource exactly. (The SDK resource name
  `FlexibleServerActiveDirectoryAdministratory` carries an upstream
  typo; it is the correct symbol for
  `azurerm_mysql_flexible_server_active_directory_administrator`.)

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless (web identity), and ambient
credential chains. Never construct the provider inline.

## Running Locally

```bash
# Build
make build

# Run with Pulumi
make run
```
