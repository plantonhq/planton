# AzureKeyVault Pulumi Module

## Overview

This Pulumi module provisions an Azure Key Vault using the Azure Classic
provider (`pulumi-azure`). It creates a single `keyvault.KeyVault` -- the
tenant-scoped container whose vault-wide controls (authorization mode,
network rules, deletion safety, SKU) everything inside inherits.

The vault authenticates against the deploying credential's Azure AD tenant,
read from the provider's client configuration -- a vault cannot be managed
cross-tenant, so the tenant is never modeled as an input.

What lives inside the vault is composed, never created here: keys are
`AzureKeyVaultKey` resources, certificates are `AzureKeyVaultCertificate`
resources, and secret values belong to a secrets-management workflow rather
than IaC state.

## Resources Created

- `keyvault.KeyVault` -- the vault, with inline legacy access policies when
  the spec carries them (inline so the vault owns its complete grant list)

## Inputs

The module receives an `AzureKeyVaultIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.vault_name` -- the vault's ARM identity (references resolved to literals by the platform); the name is GLOBALLY unique because it becomes `{name}.vault.azure.net`
- `target.spec.sku` -- STANDARD (default) or PREMIUM (HSM-backed key types for the keys inside); updatable in place
- `target.spec.rbac_authorization_enabled` -- Azure RBAC (default true, the recommended posture; grants compose as AzureRoleAssignment) vs legacy access policies
- `target.spec.access_policies` -- legacy grants: principal object id (resolved from a user-assigned identity reference), optional tenant/application ids, and the four permission lists translated through exhaustive enum vocabularies
- `target.spec.enabled_for_*` -- the three resource-manager integration switches (Azure defaults: false)
- `target.spec.public_network_access_enabled` -- false takes the vault private-endpoints-only
- `target.spec.purge_protection_enabled` -- irreversible once on; destroy then schedules deletion instead of purging
- `target.spec.soft_delete_retention_days` -- 7-90 (Azure default 90); fixed at creation
- `target.spec.network_acls` -- default_action + bypass + IP/subnet allowlists; unspecified bypass materializes Azure's default (AzureServices)
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

Optional fields with true/non-zero defaults are presence-guarded: an unset
field explicitly falls back to the proto default so a manifest-built IaC
input (which does not materialize defaults) deploys identically to the
Terraform module.

## Outputs

| Output | Description |
|--------|-------------|
| `key_vault_id` | The vault's ARM resource ID -- what keys, certificates, and vault-scoped role assignments reference |
| `key_vault_name` | The vault's name |
| `vault_uri` | The data-plane URI applications call |
| `tenant_id` | The Azure AD tenant the vault authenticates against |
| `resource_group_name` | The resource group the vault was created in |

## Operational Notes

- **Soft delete is always on.** With the provider's default features, a
  destroy purges the soft-deleted vault so the globally unique name frees
  up immediately -- unless purge protection is on, in which case destroy
  becomes a scheduled deletion at the end of the retention window.
- **A name-colliding create against a soft-deleted vault auto-recovers it**
  (provider default) rather than failing.
- The deployer permissions this module needs (including the extra grant
  required to switch authorization modes on a live vault) are cataloged in
  [`../permissions.yaml`](../permissions.yaml).
