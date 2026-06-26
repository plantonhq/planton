# Pulumi Module: Cloudflare Load Balancer Pool

Provisions a single account-scoped `cloudflare.LoadBalancerPool` — a group of
origins, health-checked by a referenced monitor, that zone-scoped load balancers
select.

## Layout

```
iac/pulumi/
├── main.go            # entrypoint (loads stack-input, calls module.Resources)
├── Pulumi.yaml
├── Makefile
└── module/
    ├── main.go            # Resources(): provider setup + pool()
    ├── locals.go          # stack-input references
    ├── pool.go            # the cloudflare.LoadBalancerPool (origins, monitor, etc.)
    └── outputs.go         # output constant names
```

## Inputs

A `CloudflareLoadBalancerPoolStackInput` (target + provider config). Required spec
fields: `account_id`, `name`, `origins[]`. Origin `address` and `monitor` arrive
resolved via `StringValueOrRef.GetValue()`; the origin `host_header` is translated
to the provider's origin `Header{ Hosts: [...] }`.

## Outputs

- `pool_id` — referenced by `CloudflareLoadBalancer` pool lists.
- `pool_name` — the pool name.

## Requirements

- **Load Balancing add-on** must be enabled on the account (paid add-on); otherwise
  the Load Balancing API returns `403`.
- The provider is configured from the stack-input provider config /
  `CLOUDFLARE_API_TOKEN`; the token needs
  **Account → Load Balancing: Monitors and Pools → Edit** (pools are account-scoped).
- **Origins must be globally routable** when a `monitor` is attached — Cloudflare
  rejects reserved / non-routable addresses (e.g. RFC 5737 ranges) under monitoring.
- **`check_regions` is capped by plan tier** — exceeding the cap fails validation;
  omit it to probe from every region.
