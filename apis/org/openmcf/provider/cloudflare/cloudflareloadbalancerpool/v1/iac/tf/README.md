# Terraform Module: Cloudflare Load Balancer Pool

Provisions a single account-scoped `cloudflare_load_balancer_pool` — a group of
origins, health-checked by a referenced monitor, that zone-scoped load balancers
select.

## Inputs

- `metadata` — name/labels.
- `spec` — see [variables.tf](./variables.tf). Required: `account_id`, `name`,
  `origins[]` (each with `name` + `address`). `StringValueOrRef` fields (origin
  `address`, `monitor`) flatten to plain strings; `check_regions` enums flatten to
  their string names; unset `optional` scalars (weight/enabled/flatten_cname/
  latitude/longitude) arrive as null so the provider applies its defaults. The
  origin `host_header` is translated to the provider's `header { host = [...] }`.

## Outputs

| Output | Description |
|---|---|
| `pool_id` | The pool ID (referenced by a load balancer's pool lists) |
| `pool_name` | The pool name |

## Requirements

- **Load Balancing add-on** must be enabled on the account (paid add-on); otherwise
  the Load Balancing API returns `403`.
- The provider reads `CLOUDFLARE_API_TOKEN`; the token needs
  **Account → Load Balancing: Monitors and Pools → Edit** (pools are account-scoped).
- **Origins must be globally routable** when a `monitor` is attached — Cloudflare
  rejects reserved / non-routable addresses (e.g. RFC 5737 ranges) under monitoring.
- **`check_regions` is capped by plan tier** — exceeding the cap fails validation;
  omit it to probe from every region.
