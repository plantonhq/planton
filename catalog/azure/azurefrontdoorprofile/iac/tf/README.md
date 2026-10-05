# AzureFrontDoorProfile - Terraform Module

Terraform implementation for the AzureFrontDoorProfile
kind.

## Resources Created

- `azurerm_cdn_frontdoor_profile.main` -- the profile container (SKU,
  response timeout, identity, log scrubbing, tags)

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.profile_name` | Required; unique within the resource group; ForceNew -- replaces the profile AND everything nested under it |
| `spec.sku` | Spec enum value name (STANDARD/PREMIUM); absent deploys Standard (tfvars drops zero-valued proto fields); ForceNew, and Azure refuses Premium -> Standard outright |
| `spec.identity` | Rendered as a dynamic `identity` block; the spec CEL guarantees identity ids are present exactly when the type includes UserAssigned |
| `spec.log_scrubbing_variables` | Each entry renders one `log_scrubbing_rule` block; presence enables scrubbing (the service supports only the match-everything operator on profiles) |
| `spec.response_timeout_seconds` | Sent only when set; Azure applies its 120 s default when omitted |

## Outputs

| Output | Description |
| --- | --- |
| `profile_id` | The ARM id -- what endpoints and origin groups reference |
| `profile_name` | The ARM namespace every child nests under |
| `resource_guid` | Front Door's own GUID (afdverify traffic validation) |
| `identity_principal_id` | The system-assigned principal (empty without one) |

## Usage

```hcl
module "front_door_profile" {
  source = "./path/to/module"

  metadata = {
    name = "my-front-door"
    org  = "mycompany"
  }

  spec = {
    resource_group = "my-rg"
    profile_name   = "my-front-door"
    sku            = "PREMIUM"
    identity = {
      type = "SYSTEM_ASSIGNED"
    }
  }
}
```

Front Door is global -- the provider forces location, so there is no
region variable.
