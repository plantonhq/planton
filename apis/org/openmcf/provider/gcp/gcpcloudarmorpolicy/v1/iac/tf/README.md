# GcpCloudArmorPolicy — Terraform Implementation

This directory contains the Terraform implementation for provisioning a GCP Cloud Armor security policy from the OpenMCF spec.

## Provider

- **Provider**: `hashicorp/google` `~> 6.0`
- Credentials supplied via `var.provider_config.service_account_key_base64` (base64-decoded)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata`, `spec`, and `provider_config` variable definitions |
| `locals.tf` | Project ID, policy name (spec or metadata.name fallback), framework labels |
| `main.tf` | `google_compute_security_policy` resource with dynamic blocks |
| `outputs.tf` | `policy_id`, `policy_name`, `policy_self_link`, `fingerprint` |

## Dynamic Blocks

The spec is mapped to the Terraform resource via nested `dynamic` blocks. Structure (5 levels):

1. **rule** — Each spec rule maps to one `rule` block
2. **match** — Contains either `config` (IP-based) or `expr` (CEL-based)
3. **rate_limit_options** — Thresholds, ban, exceed redirect (when action is throttle/rate_based_ban)
4. **header_action** / **preconfigured_waf_config** — Per-rule headers and WAF exclusions
5. **exclusion** → **request_header**, **request_cookie**, **request_uri**, **request_query_param** — WAF exclusion fields

## Known Limitations

- **Labels** — The `google_compute_security_policy` Terraform resource does not support a `labels` argument. Use the Pulumi implementation if labels are required.
- **request_body_inspection_size** — The Terraform provider does not expose this advanced option. Use Pulumi for JSON body inspection size configuration.

## Feature Parity

With the above exceptions, the Terraform module supports the same spec fields as the Pulumi implementation: adaptive protection, advanced options (json_parsing, log_level, user_ip_request_headers), all rule actions, rate limiting, redirects, header injection, and preconfigured WAF exclusions.
