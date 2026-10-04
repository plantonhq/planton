# AzureContainerAppEnvironmentStorage - Terraform Module

Terraform implementation for the AzureContainerAppEnvironmentStorage
kind.

## Resources Created

- `azurerm_container_app_environment_storage.main` -- the file-share
  registration on the environment that app and job volumes mount by
  `storage_name`

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.container_app_environment_id` | The owning environment (ForceNew) |
| `spec.storage_name` | The registration name volumes reference |
| `spec.share_name` | The Azure Files share (or NFS export) being registered |
| `spec.access_mode` | `READ_ONLY` / `READ_WRITE`, mapped to ARM's `ReadOnly` / `ReadWrite` |
| SMB vs NFS | `account_name` + `access_key` (SMB) XOR `nfs_server_url` (NFS); exactly one protocol, spec-enforced |

## Provider Version

`azurerm ~> 5.0`.

## Behavior Notes

- Only the SMB `access_key` updates in place (key rotation); every other
  change recreates the registration, briefly breaking active mounts.
- The NFS path requires a VNet-injected environment.
- No tags: ARM does not support tags on `managedEnvironments/storages`.

## Usage

```hcl
module "env_storage" {
  source = "./path/to/module"

  metadata = { name = "shared-config" }
  spec = {
    container_app_environment_id = "/subscriptions/.../managedEnvironments/apps-env"
    storage_name                 = "shared-config"
    share_name                   = "config"
    access_mode                  = "READ_ONLY"
    account_name                 = "appsstorage"
    access_key                   = var.storage_key
  }
}
```
