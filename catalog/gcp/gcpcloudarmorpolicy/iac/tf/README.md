# GcpCloudArmorPolicy — Terraform Implementation

This directory contains the Terraform implementation for provisioning a GCP
Cloud Armor security policy from the Planton spec. It also enables the
Compute Engine API on the target project.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, policy name (spec or metadata.name fallback) |
| `main.tf` | API enablement + `google_compute_security_policy` resource with dynamic blocks |
| `outputs.tf` | `policy_id`, `policy_name`, `policy_self_link`, `fingerprint` |

## Dynamic Blocks

The spec is mapped to the Terraform resource via nested `dynamic` blocks:

1. **rule** — Each spec rule maps to one `rule` block
2. **match** — `config` (IP-based) or `expr` (CEL-based), plus `expr_options`
   for reCAPTCHA site keys
3. **rate_limit_options** — Thresholds, composite `enforce_on_key_configs`,
   ban, exceed redirect (when action is throttle/rate_based_ban)
4. **header_action** / **preconfigured_waf_config** — Per-rule headers and
   WAF exclusions
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
  merge order as every labeled GCP kind.
- **`deletion_policy`** — DELETE (default), PREVENT, or ABANDON decides
  what a destroy does to the policy.
- **`advanced_options_config.request_body_inspection_size`** — how much of
  each request body the WAF inspects (8KB default, up to 64KB).
