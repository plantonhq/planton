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

## Credentials

The Cloudflare provider is configured from the stack-input provider config /
`CLOUDFLARE_API_TOKEN`. The referenced pools/monitors are account-scoped resources.
