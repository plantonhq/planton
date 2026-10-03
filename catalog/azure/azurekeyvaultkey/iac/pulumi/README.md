# AzureKeyVaultKey Pulumi Module

## Overview

This Pulumi module provisions a cryptographic key inside an Azure Key Vault
using the Azure Classic provider (`pulumi-azure`). It creates a single
`keyvault.Key` -- a DATA-PLANE object: the provider talks to the vault's
`{name}.vault.azure.net` endpoint, not ARM, so creation fails with a 403
when the deploying credential lacks key permissions on the vault, even if
it owns the subscription.

Key material is immutable by design: type, size, and curve are fixed at
creation; changing any of them replaces the key and every consumer
re-encrypts through the new key on its next unwrap.

## Resources Created

- `keyvault.Key` -- the key, with its optional rotation policy (Azure
  models the policy as a sub-resource updated in place)

## Inputs

The module receives an `AzureKeyVaultKeyIacInput` containing:

- `target.spec.name` -- 1-127 letters/digits/hyphens, unique among the vault's keys
- `target.spec.key_vault_id` -- the vault's ARM ID (resolved from an AzureKeyVault reference by the platform)
- `target.spec.key_type` -- RSA/RSA_HSM/EC/EC_HSM, translated to Azure's hyphenated wire values through an exhaustive vocabulary; the `_HSM` variants need the vault's PREMIUM SKU
- `target.spec.key_size` / `target.spec.curve` -- RSA size XOR EC curve (spec validation enforces the pairing); an unset curve lets Azure default to P-256, identically on both engines
- `target.spec.key_opts` -- the capability boundary, translated to Azure's camelCase operation strings
- `target.spec.not_before_date` / `target.spec.expiration_date` -- RFC 3339 activation/expiry instants
- `target.spec.rotation_policy` -- expire_after + notify_before_expiry (paired) and/or the automatic rotation trigger, as ISO 8601 durations
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `key_id` | Versioned data-plane ID (pins consumers) |
| `versionless_id` | Versionless data-plane ID -- the CMK reference that follows rotation |
| `key_name` / `version` | Name and current version |
| `resource_id` / `resource_versionless_id` | ARM (control-plane) IDs |
| `public_key_pem` / `public_key_openssh` | The public half |

## Operational Notes

- A deleted key's name stays reserved in the vault for the soft-delete
  retention window; the provider's default features purge soft-deleted
  keys on destroy so the name frees up immediately.
- Once `expiration_date` is set it cannot be fully unset on the underlying
  key, even across delete/recreate (Azure restores purged names' state).

## Required Permissions

The deployer permissions this module needs are cataloged in
[`../permissions.yaml`](../permissions.yaml).
