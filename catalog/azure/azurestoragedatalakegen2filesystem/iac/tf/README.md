# AzureStorageDataLakeGen2Filesystem - Terraform Module

Terraform implementation for the AzureStorageDataLakeGen2Filesystem
kind.

## Resources Created

- `azurerm_storage_data_lake_gen2_filesystem.main` -- the filesystem,
  addressed by the parent account's ARM id but CREATED through the
  account's dfs data plane (shared-key auth by default)

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.storage_account_id` | The parent account's resolved ARM id; the account NAME and the container-proxy output id are derived from it in `locals.tf` |
| `spec.filesystem_name` | 3-63 lowercase letters/digits/hyphens (or `$root`); ForceNew -- renaming replaces the filesystem and its data |
| `spec.aces` | scope/type as spec enum name strings mapped to the data plane's lowercase values; `object_id` only on USER/GROUP entries |
| `spec.owner` / `spec.group` | Sent only when set (Computed on the provider) so Azure's `$superuser` defaults stand |
| `spec.properties` | Metadata map; Azure requires base64-encoded VALUES |

## Usage

```hcl
module "data_lake_filesystem" {
  source = "./path/to/module"

  metadata = {
    name = "raw-zone"
    org  = "mycompany"
  }

  spec = {
    storage_account_id = "/subscriptions/.../providers/Microsoft.Storage/storageAccounts/mylakestorage"
    filesystem_name    = "raw-zone"
    aces = [
      { type = "USER", permissions = "rwx" },
      { type = "GROUP", permissions = "r-x" },
      { type = "OTHER", permissions = "---" },
    ]
  }
}
```

POSIX access control requires hierarchical namespace on the ACCOUNT --
Azure rejects owner/group/ace on flat-namespace accounts at apply time.
Filesystems carry no Azure tags (the properties map is the
filesystem-level metadata surface); the platform's identity tags live
on the parent account.
