# AzureFrontDoorRoute - Terraform Module

Terraform implementation for the AzureFrontDoorRoute
kind.

## Resources Created

- `azurerm_cdn_frontdoor_route.main` -- the route, addressed by the
  parent endpoint's full ARM id and forwarding to the referenced origin
  group

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.endpoint_id` | The parent endpoint's resolved ARM id; ForceNew |
| `spec.route_name` | Required; 2-90 chars; ForceNew |
| `spec.origin_group_id` | The destination pool; updatable in place |
| `spec.origin_ids` | Never sent to Azure -- pure provisioning-order references, because ARM rejects a route whose origin group has no origins yet |
| `spec.rule_set_ids` | Attached AzureFrontDoorRuleSet delivery policies; empty lists normalize to null (Front Door treats empty as "disassociate" on update) |
| `spec.custom_domain_ids` | The custom domains the route serves; same empty-to-null normalization; `link_to_default_domain: false` requires at least one (spec CEL) |
| `spec.supported_protocols` / `spec.forwarding_protocol` | Spec enum value names mapped to ARM's casing in `locals.tf`; absent forwarding protocol deploys MatchRequest |
| `spec.cache` | Rendered only when set: absent cache settings mean caching DISABLED (the provider transmits an explicit null) -- a real behavior switch |

## Outputs

| Output | Description |
| --- | --- |
| `route_id` | The ARM id of the route |
| `route_name` | The route's name inside its endpoint |

## Usage

```hcl
module "front_door_route" {
  source = "./path/to/module"

  metadata = {
    name = "default-route"
    org  = "mycompany"
  }

  spec = {
    endpoint_id         = "/subscriptions/.../providers/Microsoft.Cdn/profiles/my-fd/afdEndpoints/web"
    route_name          = "default"
    origin_group_id     = "/subscriptions/.../providers/Microsoft.Cdn/profiles/my-fd/originGroups/api-backends"
    patterns_to_match   = ["/*"]
    supported_protocols = ["HTTP", "HTTPS"]
  }
}
```

The route carries no Azure tags: ARM does not support tags on Front
Door routes; the platform's identity tags live on the profile.
