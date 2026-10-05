# AzurePostgresqlFlexibleServer - Pulumi Module

Pulumi implementation for the AzurePostgresqlFlexibleServer
kind.

## Architecture

```
postgresql.FlexibleServer
├── postgresql.FlexibleServerDatabase (per databases entry, Parent: server)
├── postgresql.FlexibleServerFirewallRule (per firewall_rules entry, Parent: server)
├── postgresql.FlexibleServerConfiguration (per server_parameters entry, Parent: server)
└── postgresql.FlexibleServerActiveDirectoryAdministrator (per aad_administrators entry, Parent: server)
```

The deploying credential's client configuration (`core.GetClientConfig`)
supplies the tenant fallback for Entra authentication and administrator
grants when the spec does not pin one.

## Key Design Decisions

- **Create-mode enums map through exhaustive vocabularies** in `locals.go`
  (create mode, HA mode, identity type, principal type) -- unspecified
  create_mode is not sent at all, matching the Terraform module's null.
- **`version` is only sent for a fresh (DEFAULT) server** -- replicas and
  restores inherit the source's version, so materializing the spec default
  onto them would fight the service.
- **Presence guards on every optional-with-default field** (version,
  public_network_access_enabled, backup_retention_days,
  authentication.password_auth_enabled, database charset/collation): IaC
  inputs built from a manifest do not materialize proto defaults, so unset
  falls back to the documented default explicitly.
- **`replication_role` is day-2 only** -- Azure rejects it at creation; the
  only legal update value ("None") promotes a replica in place.
- **`identity_principal_id` is exported conditionally** -- populated only
  when the identity type includes SYSTEM_ASSIGNED, empty otherwise,
  matching the Terraform module's conditional output.

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
