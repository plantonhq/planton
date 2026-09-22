# GcpCloudArmorPolicy — Pulumi Implementation

This directory contains the Pulumi implementation for provisioning a GCP
Cloud Armor security policy from the Planton spec -- global when
`spec.region` is empty, regional when it is set -- plus, on a regional
network policy that declares it, the region's network edge security
service. It also enables the Compute Engine API on the target project.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, API enablement, and `security_policy` |
| `module/locals.go` | Ambient project fallback, policy name (spec or metadata.name fallback), the `IsRegional` scope selector |
| `module/security_policy.go` | The GLOBAL arm: maps spec to `gcp.compute.SecurityPolicy`; contains the global rule mapping logic |
| `module/region_security_policy.go` | The REGIONAL arm: maps spec to `gcp.compute.RegionSecurityPolicy` (its own rule types, plus `network_match`, `user_defined_fields`, `ddos_protection_config`) and the count-gated `gcp.compute.NetworkEdgeSecurityService` |
| `module/outputs.go` | Output key constants (`policy_id`, `policy_name`, `policy_self_link`, `fingerprint`, `region`, `network_edge_security_service_self_link`) |

## One Kind, Two Scopes

`main.go` branches on `spec.region`: empty calls `securityPolicy` (global),
a region name calls `regionSecurityPolicy`. Pulumi's nested Args types are
per resource, so the regional file carries its own rule, match, rate-limit,
and WAF-exclusion mappers (`mapRegionRules`, `mapRegionMatch`,
`mapNetworkMatch`, `mapRegionRateLimitOptions`,
`mapRegionPreconfiguredWafConfig`) -- the same duplication the Terraform
module carries as two resource blocks. The regional arm wires none of the
global-only levers (labels, Adaptive Protection, reCAPTCHA, request-body
inspection size, redirect, header injection, exceed redirect, token
options) because the spec rejects them when `region` is set; both arms
export the same six outputs, with `region` and the edge-service link empty
on the global one.

## Rule Mapping

The spec uses flattened structures; the Pulumi SDK expects nested types. The mapping logic in `security_policy.go` handles:

- **Match**: `versioned_expr` + `src_ip_ranges` → `SecurityPolicyRuleMatchArgs.Config`; `expression` → `SecurityPolicyRuleMatchArgs.Expr`; `expr_options` → the nested reCAPTCHA site-key options
- **Redirect**: Rule-level `redirect_options` and rate-limit `exceed_redirect_options` share the same spec type (`GcpCloudArmorRedirectConfig`) but map to separate SDK types: `SecurityPolicyRuleRedirectOptionsArgs` vs `SecurityPolicyRuleRateLimitOptionsExceedRedirectOptionsArgs`
- **Composite rate-limit keys**: `enforce_on_key_configs` → `SecurityPolicyRuleRateLimitOptionsEnforceOnKeyConfigArgs` (mutually exclusive with the singular `enforce_on_key`, enforced by the spec)
- **Adaptive protection**: `threshold_configs` with per-granularity traffic units map onto the deeply nested `Layer7DdosDefenseConfigThresholdConfig` SDK types
- **WAF exclusions**: The SDK generates distinct types for each field (`RequestHeader`, `RequestCooky`, `RequestUri`, `RequestQueryParam`). The module uses per-field mappers (`mapWafExclusionHeaders`, `mapWafExclusionCookies`, etc.)

## Default Rule Contract

Every Cloud Armor policy carries a default rule at priority 2147483647.
Creating with NO rules lets the API add a default "allow all" rule
automatically; providing ANY rules requires the set to include that default
explicitly — the spec enforces this before the module ever runs.

## Labels and Destroy Behavior (parity with Terraform)

- **Labels** — user labels from `spec.labels` are merged with the platform
  attribution labels (platform wins on key conflicts), the identical merge
  order the Terraform module uses. Global policies only: the regional
  collection carries no labels.
- **`deletion_policy`** — DELETE (default), PREVENT, or ABANDON decides
  what a destroy does to the policy and to its network edge security
  service; sent only when set on both engines.
- **`rules[].priority`** — read through `GetPriority()`; 0 (Google's
  highest priority) is a legal value and is sent as 0.
- **`advanced_options_config.request_body_inspection_size`** — how much of
  each request body the WAF inspects (8KB default, up to 64KB).
