# GcpCloudArmorPolicy — Terraform Implementation

This directory contains the Terraform implementation for provisioning a GCP
Cloud Armor security policy from the Planton spec -- global when
`spec.region` is empty, regional when it is set -- plus, on a regional
network policy that declares it, the region's network edge security
service. It also enables the Compute Engine API on the target project.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, policy name (spec or metadata.name fallback), the `is_regional` scope selector, the edge-service switch and name, the fanned-out `deletion_policy` |
| `main.tf` | API enablement + `google_compute_security_policy` (global) and `google_compute_region_security_policy` (regional), exactly one created by count guards, plus the count-gated `google_compute_network_edge_security_service` |
| `outputs.tf` | `policy_id`, `policy_name`, `policy_self_link`, `fingerprint` selected by `one(concat(...))` from whichever arm exists; `region`; `network_edge_security_service_self_link` |

## One Kind, Two Scopes

`spec.region` selects the API collection. Empty builds the global policy; a
region name builds the regional one. The regional resource names its rule
list `rules` (the global one says `rule`) and carries no `labels`,
`adaptive_protection_config`, `recaptcha_options_config`,
`request_body_inspection_size`, `redirect_options`, `header_action`,
`exceed_redirect_options`, or `expr_options` -- the spec rejects each of
those when `region` is set, so the regional block simply does not wire
them. The regional block adds `ddos_protection_config`,
`user_defined_fields`, and the per-rule `network_match`; each rule carries
exactly one of `match` / `network_match`. Google compares the regional
rule list as a set, so manifest order never shows as a diff.

The network edge security service is created only when the spec declares
`network_edge_security_service`; its `security_policy` is the regional
policy's self-link and its `deletion_policy` is the kind's own.

## Dynamic Blocks

The spec is mapped to the Terraform resource via nested `dynamic` blocks:

1. **rule** — Each spec rule maps to one `rule` block
2. **match** — `config` (IP-based) or `expr` (CEL-based), plus `expr_options`
   for reCAPTCHA site keys
3. **rate_limit_options** — Thresholds, composite `enforce_on_key_configs`,
   ban, exceed redirect (when action is throttle/rate_based_ban)
4. **header_action** / **preconfigured_waf_config** — Per-rule headers and
   WAF exclusions
4a. **network_match** (regional) — Packet-level match on ranges, ports,
   protocols, country codes, ASNs, and user-defined fields; empty lists are
   sent as null so Google treats the field as unconstrained
4b. **user_defined_fields** / **ddos_protection_config** (regional) — The
   network-policy levers; `offset` and `size` are tri-state (unset sends
   null, an explicit 0 offset is sent as 0)
5. **adaptive_protection_config** — Layer 7 DDoS defense with
   `threshold_configs` and traffic granularity
6. **recaptcha_options_config** — Policy-level reCAPTCHA redirect site key

## Default Rule Contract

Every Cloud Armor policy carries a default rule at priority 2147483647.
Creating with NO rules lets the API add a default "allow all" rule
automatically; providing ANY rules requires the set to include that default
explicitly — the spec enforces this before the module ever runs.

## Labels and Destroy Behavior

- **Labels** — the module merges user labels from `spec.labels` with the
  platform attribution labels (platform wins on key conflicts), the same
  merge order as every labeled GCP kind. Global policies only: the
  regional collection carries no labels, so a regional policy receives
  none.
- **`deletion_policy`** — DELETE (default), PREVENT, or ABANDON decides
  what a destroy does to the policy and to its network edge security
  service.
- **`rules[].priority`** — a required number; 0 (Google's highest
  priority) is sent as 0.
- **`advanced_options_config.request_body_inspection_size`** — how much of
  each request body the WAF inspects (8KB default, up to 64KB).
