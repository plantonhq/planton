# AzureKeyVaultSecret Pulumi Module

## Overview

Stores a secret in an Azure Key Vault -- a versioned data-plane object the vault guards -- on the classic Pulumi Azure SDK (`pulumi-azure/sdk/v6`), wire-identical to the Terraform module. The value arrives already reference-resolved (the spec's sensitive field); it is never written to outputs.

## Resources Created

- `keyvault.Secret` -- the secret

## Outputs

- `secret_id` -- the versioned data-plane ID (pins consumers to this version)
- `versionless_id` -- the versionless data-plane ID (the reference consumers should use; value updates propagate automatically)
- `secret_name` -- the secret's name within the vault
- `version` -- the current version identifier
- `resource_id` / `resource_versionless_id` -- the versioned/versionless ARM resource IDs (control-plane identity)

## Behavior Notes

- **A value change creates a NEW version** -- the versioned outputs move with it; `versionless_id` keeps resolving to latest.
- **The provider's write-only `value_wo`/`value_wo_version` pair is deliberately not wired** -- it duplicates `value` for a plaintext-in-config problem this module does not have (values are reference-resolved at deploy, never stored in manifests).
- **Data-plane permissions are separate from ARM**: creation fails with a 403 when the deploying credential lacks secret permissions on the vault, even if it owns the subscription.
- **Destroy purges the soft-deleted secret** (the provider's default feature behavior) so the name frees immediately -- skipped automatically when the vault has purge protection, where the name stays reserved for the retention window by design.
- **Key Vault strips raw newlines** -- multi-line values should be base64-encoded, with the encoding recorded in `content_type`.
- **Engine parity**: the classic SDK v6.38.0 carries the FULL azurerm v5 surface for this kind -- zero parity exceptions.

## Required Permissions

The deployer permissions this module needs are cataloged in [`../permissions.yaml`](../permissions.yaml).
