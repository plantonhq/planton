# GcpCloudArmorPolicy — Pulumi Implementation

This directory contains the Pulumi implementation for provisioning a GCP Cloud Armor security policy from the Planton spec.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `security_policy` |
| `module/locals.go` | Initializes policy name (spec or metadata.name fallback), GCP labels, and stack input |
| `module/security_policy.go` | Maps spec to `gcp.compute.SecurityPolicy`; contains rule mapping logic |
| `module/outputs.go` | Output key constants (`policy_id`, `policy_name`, `policy_self_link`, `fingerprint`) |

## Rule Mapping

The spec uses flattened structures; the Pulumi SDK expects nested types. The mapping logic in `security_policy.go` handles:

- **Match**: `versioned_expr` + `src_ip_ranges` → `SecurityPolicyRuleMatchArgs.Config`; `expression` → `SecurityPolicyRuleMatchArgs.Expr`
- **Redirect**: Rule-level `redirect_options` and rate-limit `exceed_redirect_options` share the same spec type (`GcpCloudArmorRedirectConfig`) but map to separate SDK types: `SecurityPolicyRuleRedirectOptionsArgs` vs `SecurityPolicyRuleRateLimitOptionsExceedRedirectOptionsArgs`
- **WAF exclusions**: The SDK generates distinct types for each field (`RequestHeader`, `RequestCooky`, `RequestUri`, `RequestQueryParam`). The module uses per-field mappers (`mapWafExclusionHeaders`, `mapWafExclusionCookies`, etc.)

## Labels

GCP labels are supported via the Pulumi provider. The module applies framework labels (resource kind, name, org, env, id) and passes them to `SecurityPolicyArgs.Labels`. Terraform does not support labels on `google_compute_security_policy`.

## Pulumi-Only Features

- **Labels** — Terraform does not expose labels on the security policy resource.
- **request_body_inspection_size** — Advanced option available in the Pulumi GCP provider; not supported in the Terraform `google_compute_security_policy` resource.
