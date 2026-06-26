# Terraform Module: Cloudflare Load Balancer Monitor

Provisions a single account-scoped `cloudflare_load_balancer_monitor` — a health
check that `CloudflareLoadBalancerPool`s reference to probe their origins.

## Inputs

- `metadata` — name/labels.
- `spec` — see [variables.tf](./variables.tf). Required: `account_id`. The `type`
  enum flattens to its string name (`monitor_type_unspecified`/`""` → `http`);
  numeric tuning knobs left at 0 are omitted so the provider applies its defaults.

## Outputs

| Output | Description |
|---|---|
| `monitor_id` | The monitor ID (referenced by a pool's `monitor`) |
| `monitor_type` | The health-check protocol |

## Requirements

- **Load Balancing add-on** must be enabled on the account (paid add-on); otherwise
  the Load Balancing API returns `403`.
- The provider reads `CLOUDFLARE_API_TOKEN`; the token needs
  **Account → Load Balancing: Monitors and Pools → Edit** (monitors are account-scoped).
