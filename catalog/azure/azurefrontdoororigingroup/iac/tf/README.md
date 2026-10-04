# AzureFrontDoorOriginGroup - Terraform Module

Terraform implementation for the AzureFrontDoorOriginGroup
kind.

## Resources Created

- `azurerm_cdn_frontdoor_origin_group.main` -- the origin group,
  addressed by the parent profile's full ARM id

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.profile_id` | The parent profile's resolved ARM id; ForceNew |
| `spec.origin_group_name` | Required; 2-90 chars; ForceNew -- replaces the group and every origin under it |
| `spec.load_balancing` | Azure requires the block on every group, so the module ALWAYS sends it -- spec values when present, Azure's defaults (4 / 3 / 50 ms) otherwise |
| `spec.health_probe` | Rendered only when set: absent probe settings mean probing DISABLED (all origins assumed healthy) -- a real behavior switch |
| `spec.session_affinity_enabled` / `spec.restore_traffic_time_to_healed_or_new_endpoint_in_minutes` | Sent only when set; Azure defaults to true / 10 minutes |

## Outputs

| Output | Description |
| --- | --- |
| `origin_group_id` | The ARM id -- what origins and routes reference |
| `origin_group_name` | The group's name inside the profile |

## Usage

```hcl
module "front_door_origin_group" {
  source = "./path/to/module"

  metadata = {
    name = "api-backends"
    org  = "mycompany"
  }

  spec = {
    profile_id        = "/subscriptions/.../providers/Microsoft.Cdn/profiles/my-front-door"
    origin_group_name = "api-backends"
    health_probe = {
      protocol            = "HTTPS"
      interval_in_seconds = 30
      path                = "/healthz"
    }
  }
}
```

The group carries no Azure tags: ARM does not support tags on Front
Door origin groups; the platform's identity tags live on the profile.
