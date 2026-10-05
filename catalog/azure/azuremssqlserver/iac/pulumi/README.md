# AzureMssqlServer - Pulumi Module

Pulumi implementation for the AzureMssqlServer kind.

## Architecture

```
mssql.Server
├── mssql.FirewallRule (per firewall_rules entry, Parent: server)
├── mssql.VirtualNetworkRule (per virtual_network_rules entry, Parent: server)
├── mssql.OutboundFirewallRule (per outbound_firewall_rules FQDN, Parent: server)
├── mssql.ServerExtendedAuditingPolicy (when extended_auditing is present, Parent: server)
└── mssql.ServerSecurityAlertPolicy (when security_alert_policy is present, Parent: server)
```

The deploying credential's client configuration (`core.GetClientConfig`)
supplies the tenant fallback for the Entra administrator when the spec
does not pin one.

## Key Design Decisions

- **Enums map through exhaustive vocabularies** in `locals.go`
  (connection policy, identity type, Defender state, detector types --
  the detectors carry ARM's Snake_Pascal wire strings). Unspecified
  connection_policy is not sent at all, matching the Terraform module's
  null.
- **Presence guards on every optional-with-default field** (version, TLS
  floor, public network access, auditing retention/log-monitoring):
  IaC inputs built from a manifest do not materialize proto defaults,
  so unset falls back to the documented default explicitly.
- **The Entra-only contract lives in spec validation** -- the module
  simply omits credentials when they are empty, so an Entra-only server
  deploys without ever materializing a password.
- **`identity_principal_id` is exported conditionally** -- populated only
  when the identity type includes SYSTEM_ASSIGNED, empty otherwise,
  matching the Terraform module's conditional output.
- **The Defender policy addresses the server by name + resource group**
  while everything else uses the server ID -- each child matches its
  azurerm resource's own contract.

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
