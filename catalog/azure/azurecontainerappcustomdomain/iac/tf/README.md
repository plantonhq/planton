# AzureContainerAppCustomDomain - Terraform Module

Terraform implementation for the AzureContainerAppCustomDomain
kind.

## Resources Created

Exactly one of two `count`-dispatched variants of the same binding:

- `azurerm_container_app_custom_domain.byo` -- when the spec references
  an environment certificate (bring-your-own TLS)
- `azurerm_container_app_custom_domain.managed` -- otherwise; Azure
  attaches its managed certificate to the binding asynchronously, so the
  variant ignores drift on `certificate_binding_type` and
  `container_app_environment_certificate_id`

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.domain_name` | The hostname to bind |
| `spec.container_app_id` | The app carrying the binding; it must have ingress enabled |
| `spec.container_app_environment_certificate_id` | Presence selects the BYO variant; absence selects managed |
| `spec.certificate_binding_type` | Enum mapped to ARM wire values (`SNI_ENABLED` -> `SniEnabled`, `DISABLED` -> `Disabled`, `AUTO` -> `Auto`); paired with the certificate id by spec validation |

## Provider Version

`azurerm ~> 5.0`.

## Behavior Notes

- Create BLOCKS on DNS ownership proof: publish the `asuid` TXT record
  and the CNAME/A record BEFORE deploying, or the apply times out.
- Every field change replaces the binding -- there is no update surface.
- No tags: the binding is ingress configuration, not a taggable ARM
  resource.
- Outputs `coalesce` across the two count variants, so consumers never
  care which path deployed.

## Usage

```hcl
module "custom_domain" {
  source = "./path/to/module"

  metadata = { name = "www-binding" }
  spec = {
    domain_name      = "www.example.com"
    container_app_id = "/subscriptions/.../containerApps/web"
  }
}
```
