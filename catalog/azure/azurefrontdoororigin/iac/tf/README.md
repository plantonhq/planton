# AzureFrontDoorOrigin - Terraform Module

Terraform implementation for the AzureFrontDoorOrigin
kind.

## Resources Created

- `azurerm_cdn_frontdoor_origin.main` -- the origin, addressed by the
  parent origin group's full ARM id

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.origin_group_id` | The parent group's resolved ARM id; ForceNew |
| `spec.origin_name` | Required; 2-90 chars; ForceNew |
| `spec.host_name` | Required; the backend's hostname or IP |
| `spec.certificate_name_check_enabled` | Always sent (the provider requires it); absent means the spec default `true` -- the secure posture, and required with Private Link |
| `spec.origin_host_header` | Sent only when set; unset lets Azure use the origin hostname (right for App Service/Container Apps) |
| `spec.private_link` | Rendered only when set; PREMIUM profiles only (Azure enforces at apply); `target_type` maps spec enum names to azurerm's snake_case dialect (`blob_secondary`, `web_secondary`) |

## Outputs

| Output | Description |
| --- | --- |
| `origin_id` | The ARM id -- what routes list in `origin_ids` |
| `origin_name` | The origin's name inside its group |

## Usage

```hcl
module "front_door_origin" {
  source = "./path/to/module"

  metadata = {
    name = "primary-app"
    org  = "mycompany"
  }

  spec = {
    origin_group_id = "/subscriptions/.../providers/Microsoft.Cdn/profiles/my-fd/originGroups/api-backends"
    origin_name     = "primary-app"
    host_name       = "myapp.azurewebsites.net"
  }
}
```

The origin carries no Azure tags: ARM does not support tags on Front
Door origins; the platform's identity tags live on the profile.
