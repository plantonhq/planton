# Pulumi Module: Cloudflare Load Balancer Monitor

Provisions a single account-scoped `cloudflare.LoadBalancerMonitor` — a health
check that `CloudflareLoadBalancerPool`s reference to probe their origins.

## Layout

```
iac/pulumi/
├── main.go            # entrypoint (loads stack-input, calls module.Resources)
├── Pulumi.yaml
├── Makefile
└── module/
    ├── main.go            # Resources(): provider setup + monitor()
    ├── locals.go          # stack-input references
    ├── monitor.go         # the cloudflare.LoadBalancerMonitor
    └── outputs.go         # output constant names
```

## Inputs

A `CloudflareLoadBalancerMonitorStackInput` (target + provider config). Required
spec field: `account_id`. The `type` enum's unspecified zero value maps to `http`.

## Outputs

- `monitor_id` — referenced by `CloudflareLoadBalancerPool.monitor`.
- `monitor_type` — the configured protocol.

## Requirements

- **Load Balancing add-on** must be enabled on the account (paid add-on); otherwise
  the Load Balancing API returns `403`.
- The provider is configured from the stack-input provider config /
  `CLOUDFLARE_API_TOKEN`; the token needs
  **Account → Load Balancing: Monitors and Pools → Edit** (monitors are account-scoped).
