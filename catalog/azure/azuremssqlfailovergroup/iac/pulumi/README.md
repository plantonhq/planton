# AzureMssqlFailoverGroup - Pulumi Module

Pulumi implementation for the AzureMssqlFailoverGroup deployment
component.

## Architecture

```
mssql.FailoverGroup (one cross-region failover group + listener outputs)
```

## Key Design Decisions

- **Listener FQDNs are composed, not read** --
  `{name}.database.windows.net` and
  `{name}.secondary.database.windows.net` are built from the group name
  because Azure does not return them; the group name is therefore also a
  DNS label.
- **`grace_minutes` pairs only with Automatic failover** (CEL-enforced
  at the spec); Manual sends no grace value.
- **Databases are set only when non-empty** -- an empty failover group
  is legal and deploys cleanly.
- **`readonly_endpoint_failover_policy_enabled` unset deploys the
  provider's Disabled default**, keeping both engines identical.
- **Identity tags match the Terraform module** key for key and value
  for value: `resource_kind` is the lowercased CloudResourceKind enum
  name, and `resource_id` is written only when `metadata.id` is set.

## Provider

Built via the shared `pulumiazureprovider.Get` builder -- static client
secret, keyless web identity, or ambient chain, resolved from the stack
input. Never construct the provider inline.
