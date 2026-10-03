# AzureDiskEncryptionSet - Pulumi Module

Pulumi implementation for the AzureDiskEncryptionSet deployment
component.

## Architecture

```
compute.DiskEncryptionSet (one CMK encryption anchor)
```

## Key Design Decisions

- **`auto_key_rotation_enabled` and `encryption_type` are sent only when
  the spec sets them explicitly** -- unset lets Azure apply its own
  defaults, keeping both engines' request bodies identical.
- **The key URL shape is rotation-dependent** -- versionless with
  auto-rotation on, versioned with it off; the provider validates the
  pairing at plan.
- **Identity is the key-unwrap principal** -- its crypto grant on the
  vault is managed out-of-band, and the vault must carry purge
  protection.
- **Principal/tenant outputs resolve via `ApplyT` after create** and
  export empty for user-assigned-only identity sets.
- **Identity tags match the Terraform module** key for key and value
  for value: `resource_kind` is the lowercased CatalogKind enum
  name, and `resource_id` is written only when `metadata.id` is set.

## Provider

Built via the shared `pulumiazureprovider.Get` builder -- static client
secret, keyless web identity, or ambient chain, resolved from the stack
input. Never construct the provider inline.
