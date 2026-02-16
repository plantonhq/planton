# GcpCloudArmorPolicy Deployment Component

**Date**: February 16, 2026
**Type**: Feature
**Components**: GCP Provider, API Definitions, Pulumi CLI Integration, Terraform Module

## Summary

Added the GcpCloudArmorPolicy deployment component -- the most structurally complex GCP resource in the expansion project with 13 sub-messages covering WAF rules, DDoS defense, rate limiting, redirect actions, header injection, and preconfigured WAF exclusions. This is the 23rd and final GCP resource in the provider expansion.

## Problem Statement / Motivation

Cloud Armor is a critical security layer for any production GCP deployment. Without it, backend services behind load balancers have no WAF protection, no DDoS defense, no rate limiting, and no IP-based access control.

### Pain Points

- GCP users needed to configure Cloud Armor outside OpenMCF, breaking the single-manifest deployment model
- No way to compose security policies with other OpenMCF-managed resources via infra charts
- Cloud Armor's deeply nested rule structure (5 levels of nesting in TF) is error-prone to configure manually

## Solution / What's New

A complete deployment component at `apis/org/openmcf/provider/gcp/gcpcloudarmorpolicy/v1/` that provisions a Cloud Armor security policy with inline rules supporting all core features.

### Key Design Decisions

- **Flattened match structure**: `match.config.src_ip_ranges` simplified to `match.srcIpRanges` and `match.expr.expression` to `match.expression` -- each wrapper had only one useful field. IaC modules reconstruct the nested structure for the providers.
- **Shared sub-message types**: `GcpCloudArmorRateThreshold` shared between `rateLimitThreshold` and `banThreshold`; `GcpCloudArmorRedirectConfig` shared between rule `redirectOptions` and rate limit `exceedRedirectOptions`.
- **13 sub-messages**: Most complex spec in the project, appropriate for Cloud Armor's domain complexity.
- **Inline rules**: Rules bundled with the policy (not separate resources) because a policy without rules is meaningless.

## Implementation Details

### Proto API

- 4 proto files, 13 sub-messages, 430 lines in `spec.proto`
- 1 `StringValueOrRef` field: `projectId` (GcpProject)
- CEL validations: match mutual exclusion (versioned_expr XOR expression), versioned_expr requires src_ip_ranges, action/type/operator in-list validations
- Flattened match: spec uses `srcIpRanges` and `expression` at the match level; IaC modules reconstruct the provider's nested `config.src_ip_ranges` and `expr.expression`

### Pulumi Module (303 lines in security_policy.go)

- Maps flattened match to nested SDK `Config`/`Expr` structures
- Separate builder functions for WAF exclusion fields: the Pulumi SDK generates distinct types for headers, cookies, URIs, and query params despite identical structures (operator + value)
- Separate redirect type for exceed_redirect_options vs rule redirect_options (SDK has different types)
- Framework GCP labels applied (Pulumi supports `Labels` on `SecurityPolicyArgs`)

### Terraform Module (5 levels of nested dynamic blocks)

- `rule` -> `match`/`rate_limit_options`/`redirect_options`/`header_action`/`preconfigured_waf_config`
- Dynamic blocks within dynamic blocks for WAF exclusions
- Known TF limitation: `labels` and `request_body_inspection_size` not supported in Google provider v6.50.0

### Validation

- 54 tests (30 positive, 24 negative)
- Covers: all action types, all policy types, IP-based and CEL matching, rate limiting (throttle + ban), redirect (EXTERNAL_302 + GOOGLE_RECAPTCHA), header actions, WAF exclusions, adaptive protection, advanced options, preview mode

## Benefits

- Complete Cloud Armor coverage for OpenMCF GCP users
- Composable via `StringValueOrRef` -- `policySelfLink` output enables attachment to backend services in infra charts
- 3 presets for immediate deployment: basic IP allowlist, OWASP WAF protection, API rate limiting

## Impact

- **GCP users**: Can now define Cloud Armor policies declaratively alongside their load balancers and backend services
- **Infra chart authors**: Can compose WAF policies into deployment environments
- **Project milestone**: Completes the GCP resource expansion (23 of 23 resources)

## Related Work

- Part of the 20260215.01.sp.gcp-resource-expansion sub-project
- Complements GcpFirewallRule (R01) for network-level security
- Referenced by GcpCloudCdn backend services

---

**Status**: Production Ready
