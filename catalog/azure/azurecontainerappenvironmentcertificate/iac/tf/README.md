# AzureContainerAppEnvironmentCertificate - Terraform Module

Terraform implementation for the AzureContainerAppEnvironmentCertificate
kind.

## Resources Created

- `azurerm_container_app_environment_certificate.main` -- the
  bring-your-own certificate registered on the environment, shared by
  every app in it; custom-domain bindings reference it by
  `certificate_id`

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.certificate_name` | Name within the environment |
| `spec.container_app_environment_id` | The owning environment (ForceNew) |
| `spec.certificate_blob_base64` + `spec.certificate_password` | Inline PFX path; an empty password is still sent (passwordless PFX is legal) |
| `spec.certificate_key_vault` | Key Vault path (`key_vault_secret_id` + optional identity); mutually exclusive with the blob path, unset identity defaults to `"System"` |
| `spec.tags` | The only in-place-updatable field |

## Provider Version

`azurerm ~> 5.0`.

## Behavior Notes

- Blob/password and Key Vault sources are mutually exclusive; the unused
  path's arguments are nulled so both engines send identical bodies.
- Azure never returns the PFX on read, so blob drift is invisible --
  rotation happens by updating the spec, and the `expiration_date`
  output is the alarm to watch.
- Any change other than tags replaces the certificate (a brief rebind
  for the custom domains that reference it).
- The Key Vault path requires the environment's managed identity to
  already read the vault's secrets.

## Usage

```hcl
module "env_cert" {
  source = "./path/to/module"

  metadata = { name = "wildcard-cert" }
  spec = {
    certificate_name             = "wildcard-example-com"
    container_app_environment_id = "/subscriptions/.../managedEnvironments/apps-env"
    certificate_blob_base64      = filebase64("wildcard.pfx")
    certificate_password         = var.pfx_password
  }
}
```
