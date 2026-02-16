# GcpCloudArmorPolicy — Pulumi Architecture Overview

## Module Entry Point

The Pulumi module is invoked via `module.Resources()`. Flow:

1. `initializeLocals()` — Resolve policy name, labels, and stack input
2. `pulumigoogleprovider.Get()` — Obtain GCP provider from provider config
3. `securityPolicy()` — Create the `SecurityPolicy` resource and export outputs

## Policy Name Resolution

`locals.go` resolves the policy name:

- If `spec.policy_name` is set, use it.
- Otherwise fall back to `metadata.name`.

This matches the Terraform locals logic and allows users to omit `policy_name` when the resource name is sufficient.

## Rule Mapping: Flattened Spec → Nested SDK

### Match Conditions

The spec flattens match into `versioned_expr` + `src_ip_ranges` (IP-based) or `expression` (CEL). The SDK expects:

- IP-based: `Match.VersionedExpr` + `Match.Config.SrcIpRanges`
- CEL-based: `Match.Expr.Expression`

`mapMatch()` reconstructs the nested structure from the flattened spec.

### Redirect Types

Rule-level `redirect_options` and rate-limit `exceed_redirect_options` both use `GcpCloudArmorRedirectConfig` in the spec. The Pulumi SDK separates them:

- `SecurityPolicyRuleRedirectOptionsArgs` (rule action)
- `SecurityPolicyRuleRateLimitOptionsExceedRedirectOptionsArgs` (rate limit exceed action)

Both are populated from the shared config (`type`, `target`).

### WAF Exclusions

Exclusion field params (operator + value) are identical across headers, cookies, URIs, and query params. The Pulumi SDK generates distinct types per field. The module uses:

- `mapWafExclusionHeaders`
- `mapWafExclusionCookies`
- `mapWafExclusionUris`
- `mapWafExclusionQueryParams`

## Outputs

Outputs are exported via `ctx.Export()`:

| Key | Source |
|-----|--------|
| `policy_id` | `createdPolicy.ID()` — fully qualified resource ID |
| `policy_name` | `createdPolicy.Name` |
| `policy_self_link` | `createdPolicy.SelfLink` — used when attaching to backends |
| `fingerprint` | `createdPolicy.Fingerprint` — for concurrency control |
