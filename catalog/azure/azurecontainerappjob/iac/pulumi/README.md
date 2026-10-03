# AzureContainerAppJob - Pulumi Module

Pulumi implementation for the AzureContainerAppJob component.

## Architecture

```
containerapp.Job (one finite-run workload: template + trigger + identity)
```

## Key Design Decisions

- **Exactly one trigger** (manual / schedule / event), spec-enforced;
  switching trigger types replaces the job -- the module keeps that
  ForceNew surface honest.
- **Optional integers ride `intOrDefault`** and optional fields are
  presence-guarded, so unspecified specs deploy Azure's own defaults and
  both engines send identical request bodies (probe delays, trigger
  parallelism/completion counts).
- **Registry credentials reference the job's secret list or a managed
  identity** -- credential material never rides plain configuration.
- **`identity_principal_id` exports empty when no system identity is
  enabled** rather than failing consumers.
- **Identity tags match the Terraform module** key for key and value
  for value: `resource_kind` is the lowercased CatalogKind enum
  name, and `resource_id` is written only when `metadata.id` is set.

## Provider

Built via the shared `pulumiazureprovider.Get` builder -- static client
secret, keyless web identity, or ambient chain, resolved from the stack
input. Never construct the provider inline.
