# Terraform Module: Cloudflare Load Balancer

Provisions a single zone-scoped `cloudflare_load_balancer` that references
account-scoped pools. Pools and monitors are separate modules
(`cloudflareloadbalancerpool`, `cloudflareloadbalancermonitor`).

## Resources

- `cloudflare_load_balancer` (zone-scoped) — attaches the hostname to the given
  `default_pools` / `fallback_pool` and applies steering, session affinity,
  geo-pool maps, adaptive routing, location strategy, and random steering.

## Inputs

- `metadata` — name/labels.
- `spec` — see [variables.tf](./variables.tf). Required: `hostname`, `zone_id`,
  `default_pools`, `fallback_pool`. Pool references flatten from `StringValueOrRef`
  to plain strings (and lists to `list(string)`); enums flatten to their string
  names (`none`/`off` are omitted so the provider applies its default).

## Outputs

| Output | Description |
|---|---|
| `load_balancer_id` | The load balancer ID |
| `load_balancer_dns_record_name` | The hostname |
| `load_balancer_cname_target` | The hostname clients point their DNS at |

## Credentials

The provider reads `CLOUDFLARE_API_TOKEN` from the environment (Load Balancers:
Edit on the zone; the pools/monitors it references are account-scoped).
