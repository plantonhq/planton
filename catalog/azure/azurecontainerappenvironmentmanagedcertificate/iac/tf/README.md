# AzureContainerAppEnvironmentManagedCertificate - Terraform Module

Terraform implementation for the
AzureContainerAppEnvironmentManagedCertificate kind.

## Resources Created

- `azurerm_container_app_environment_managed_certificate.main` -- the
  free, Azure-managed, domain-validated certificate on the environment,
  auto-renewed by Azure

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.certificate_name` | Name within the environment |
| `spec.container_app_environment_id` | The owning environment (ForceNew) |
| `spec.subject_name` | The hostname Azure validates and issues for |
| `spec.domain_control_validation` | `HTTP` or `CNAME`, matching Azure's wire values verbatim; unset deploys `HTTP` (sent explicitly for engine parity) |
| `spec.tags` | The only in-place-updatable field |

## Provider Version

`azurerm ~> 5.0`.

## Behavior Notes

- Create BLOCKS on domain-validation proof and polls up to ~30 minutes:
  publish the `asuid` TXT record plus the CNAME (or HTTP routing) BEFORE
  deploying.
- Azure attaches the issued certificate to the matching custom-domain
  binding asynchronously -- the binding module ignores that drift by
  design.
- Any change other than tags re-issues the certificate.

## Usage

```hcl
module "managed_cert" {
  source = "./path/to/module"

  metadata = { name = "www-managed-cert" }
  spec = {
    certificate_name             = "www-example-com"
    container_app_environment_id = "/subscriptions/.../managedEnvironments/apps-env"
    subject_name                 = "www.example.com"
  }
}
```
