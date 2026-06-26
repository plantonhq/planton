# Pulumi Module: Cloudflare Load Balancer

Provisions a single zone-scoped `cloudflare.LoadBalancer` that references
account-scoped pools. Pools and monitors are separate modules
(`cloudflareloadbalancerpool`, `cloudflareloadbalancermonitor`).

## Layout

```
iac/pulumi/
├── main.go            # entrypoint (loads stack-input, calls module.Resources)
├── Pulumi.yaml
├── Makefile
└── module/
    ├── main.go            # Resources(): provider setup + load_balancer()
    ├── locals.go          # stack-input references
    ├── load_balancer.go   # the cloudflare.LoadBalancer + geoPoolMap helper
    └── outputs.go         # output constant names
```

## Inputs

A `CloudflareLoadBalancerStackInput` (target + provider config). Required spec
fields: `hostname`, `zoneId`, `defaultPools`, `fallbackPool`. Pool/zone references
arrive resolved via `StringValueOrRef.GetValue()`.

## Outputs

- `load_balancer_id` — the load balancer ID.
- `load_balancer_dns_record_name` — the hostname.
- `load_balancer_cname_target` — the hostname clients point their DNS at.

## Requirements

- **Load Balancing add-on** must be enabled on the account (paid add-on); otherwise
  the Load Balancing API returns `403`.
- The Cloudflare provider is configured from the stack-input provider config /
  `CLOUDFLARE_API_TOKEN`. The token needs **Zone → Load Balancers → Edit** for the
  zone-scoped load balancer (distinct from the account-level "Load Balancers Account"
  permission), plus **Account → Load Balancing: Monitors and Pools → Edit** for the
  pools/monitors it references, and the zone in its Zone Resources scope.
