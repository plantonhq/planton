# AzureStorageLocalUser - Terraform Module

Terraform implementation for the AzureStorageLocalUser deployment
component.

## Resources Created

- `azurerm_storage_account_local_user.main` -- the SFTP credential
  identity, addressed by the parent account's ARM id (a pure
  management-plane resource)

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.storage_account_id` | The parent account's resolved ARM id; the account NAME and the composed `sftp_username` output are derived from it in `locals.tf` |
| `spec.user_name` | 3-64 lowercase letters and digits; ForceNew -- renaming replaces the user and regenerates credentials |
| `spec.ssh_authorized_keys` | Paired with `ssh_key_enabled` (enforced in the spec so it fails at validate time, not apply time) |
| `spec.permission_scopes` | service as the spec enum's name string (BLOB/FILE) mapped to lowercase wire values; the five grant booleans pass through into the provider's `permissions` block |

## Usage

```hcl
module "storage_local_user" {
  source = "./path/to/module"

  metadata = {
    name = "partner01"
    org  = "mycompany"
  }

  spec = {
    storage_account_id   = "/subscriptions/.../providers/Microsoft.Storage/storageAccounts/mysftpstorage"
    user_name            = "partner01"
    ssh_password_enabled = true
    permission_scopes = [{
      service       = "BLOB"
      resource_name = "partner-inbound"
      read          = true
      write         = true
      list          = true
      create        = true
    }]
  }
}
```

The `password` output is marked sensitive (it is returned by Azure
exactly once, at the creation that enabled password auth); `sid` is an
identifier and is unwrapped from the provider's sensitive mark. Local users carry no Azure tags; the platform's identity tags
live on the parent account.
