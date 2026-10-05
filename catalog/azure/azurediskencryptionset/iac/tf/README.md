# AzureDiskEncryptionSet - Terraform Module

Terraform implementation for the AzureDiskEncryptionSet
kind.

## Resources Created

- `azurerm_disk_encryption_set.main` -- the customer-managed-key
  encryption anchor that managed disks, VMs, and scale sets reference by
  ARM ID

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.key_vault_key_id` | Versionless key URL when auto-rotation is on, versioned when off (the provider validates the pairing) |
| `spec.identity` | Required -- the identity that unwraps the key; its crypto grant on the vault is managed out-of-band |
| `spec.encryption_type` | Enum mapped to ARM strings; unset sends null so Azure's default applies |
| `spec.auto_key_rotation_enabled` | Presence-guarded for engine parity |
| `spec.federated_client_id` | Empty means same-tenant (sent as null) |

## Provider Version

`azurerm ~> 5.0`.

## Behavior Notes

- The referenced Key Vault must have purge protection enabled -- an ARM
  requirement for disk encryption sets.
- `identity_principal_id` / `identity_tenant_id` output empty when the
  set uses only user-assigned identities.

## Usage

```hcl
module "des" {
  source = "./path/to/module"

  metadata = { name = "prod-des" }
  spec = {
    region           = "eastus"
    resource_group   = "security-rg"
    name             = "prod-des"
    key_vault_key_id = "https://prod-kv.vault.azure.net/keys/disk-cmk"
    identity         = { type = "SYSTEM_ASSIGNED" }
    auto_key_rotation_enabled = true
  }
}
```
