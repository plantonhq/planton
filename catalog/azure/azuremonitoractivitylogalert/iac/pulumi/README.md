# AzureMonitorActivityLogAlert - Pulumi Module

Pulumi implementation for the AzureMonitorActivityLogAlert deployment
component.

## Architecture

```
monitoring.ActivityLogAlert (one subscription-plane Activity Log alert)
```

## Key Design Decisions

- **Criteria and actions are built with presence guards** -- fields are
  sent only when the spec carries them, so unspecified filters match
  everything in the chosen category and both engines send identical
  request bodies.
- **`enabled` is sent only when explicit** -- unset leaves Azure's own
  default (true).
- **Enum mappers mirror the Terraform module's locals row-by-row**
  (categories, levels, health statuses/reasons, recommendation types) --
  a vocabulary drift fails loudly at preview.
- **The recommendation / resource-health / service-health sub-criteria
  are mutually exclusive**, enforced by spec CELs before either engine
  runs.
- **Identity tags match the Terraform module** key for key and value
  for value: `resource_kind` is the lowercased CloudResourceKind enum
  name, and `resource_id` is written only when `metadata.id` is set.

## Provider

Built via the shared `pulumiazureprovider.Get` builder -- static client
secret, keyless web identity, or ambient chain, resolved from the stack
input. Never construct the provider inline.
