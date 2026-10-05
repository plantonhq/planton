# Pulumi Module: Cloudflare Load Balancer

Provisions a single zone-scoped `cloudflare.LoadBalancer` that references
account-scoped pools. Pools and monitors are separate modules
(`cloudflareloadbalancerpool`, `cloudflareloadbalancermonitor`).

## Layout

```
iac/pulumi/
├── main.go            # entrypoint (loads iac-input, calls module.Resources)
├── Pulumi.yaml
└── module/
    ├── main.go            # Resources(): provider setup + load_balancer()
    ├── locals.go          # iac-input references
    ├── load_balancer.go   # the cloudflare.LoadBalancer + rules/geoPoolMap helpers
    └── outputs.go         # output constant names
```

## Inputs

A `CloudflareLoadBalancerIacInput` (target + provider config). Required spec
fields: `hostname`, `zoneId`, `defaultPools`, `fallbackPool`. Pool/zone references
arrive resolved via `StringValueOrRef.GetValue()`.

`spec.rules[]` is converted by `ruleArray`/`ruleOverridesArgs` in
[module/load_balancer.go](./module/load_balancer.go). Rule OVERRIDES keep real
presence: an unset override inherits the load balancer's setting, while an
explicit value — including `none`/`off`/`0` for the presence-carrying
`session_affinity`, `steering_policy`, and `priority` fields — is sent as a
real override. `terminates`/`disabled` are sent only when true (a
`fixed_response` rule is auto-marked terminating server-side).

## Outputs

- `load_balancer_id` — the load balancer ID.
- `load_balancer_dns_record_name` — the hostname.
- `load_balancer_cname_target` — the hostname clients point their DNS at.
- `zone_id` — the owning zone (the API identity is `zones/{zone_id}/load_balancers/{id}`).

## Requirements

- **Load Balancing add-on** must be enabled on the account (paid add-on); otherwise
  the Load Balancing API returns `403`.
- The Cloudflare provider is configured from the iac-input provider config /
  `CLOUDFLARE_API_TOKEN`. The token needs **Zone → Load Balancers → Edit** for the
  zone-scoped load balancer (distinct from the account-level "Load Balancers Account"
  permission), plus **Account → Load Balancing: Monitors and Pools → Edit** for the
  pools/monitors it references, and the zone in its Zone Resources scope.
