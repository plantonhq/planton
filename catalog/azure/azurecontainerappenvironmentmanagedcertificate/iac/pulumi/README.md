# AzureContainerAppEnvironmentManagedCertificate - Pulumi Module

Pulumi implementation for the
AzureContainerAppEnvironmentManagedCertificate kind.

## Architecture

```
containerapp.EnvironmentManagedCertificate (one Azure-managed certificate)
```

## Key Design Decisions

- **The validation enum matches Azure's wire values verbatim**
  (`HTTP` / `CNAME`); unspecified deploys `HTTP`, sent explicitly so
  both engines send identical request bodies.
- **Create blocks on domain-validation proof** -- the `asuid` TXT record
  and the CNAME (or HTTP routing) must exist before deploy; Azure polls
  validation for up to ~30 minutes.
- **The issued certificate attaches to the matching custom-domain
  binding asynchronously** -- the binding module tolerates that drift by
  design; this module only owns issuance.
- **Identity tags match the Terraform module** key for key and value
  for value: `resource_kind` is the lowercased CatalogKind enum
  name, and `resource_id` is written only when `metadata.id` is set.

## Provider

Built via the shared `pulumiazureprovider.Get` builder -- static client
secret, keyless web identity, or ambient chain, resolved from the IaC
input. Never construct the provider inline.
