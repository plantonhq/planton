# AzureFrontDoorEndpoint - Terraform Module

Terraform implementation for the AzureFrontDoorEndpoint
kind.

## Resources Created

- `azurerm_cdn_frontdoor_endpoint.main` -- the endpoint, addressed by
  the parent profile's full ARM id

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.profile_id` | The parent profile's resolved ARM id; ForceNew |
| `spec.endpoint_name` | Required; 2-46 chars; ForceNew -- renaming changes the generated hostname and breaks DNS pointing at it |
| `spec.enabled` | Sent only when set; Azure defaults to enabled |
| `spec.tags` | Merged over the metadata-derived tags in `locals.tf`; user tags win |

## Outputs

| Output | Description |
| --- | --- |
| `endpoint_id` | The ARM id -- what routes reference as their parent |
| `endpoint_name` | The endpoint's name inside the profile |
| `host_name` | The generated `*.azurefd.net` hostname (CNAME target) |

## Usage

```hcl
module "front_door_endpoint" {
  source = "./path/to/module"

  metadata = {
    name = "web-endpoint"
    org  = "mycompany"
  }

  spec = {
    profile_id    = "/subscriptions/.../providers/Microsoft.Cdn/profiles/my-front-door"
    endpoint_name = "web"
  }
}
```
