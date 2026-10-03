# AWS WAF Family Depth Pass

**Date**: July 7, 2026
**Type**: Feature
**Components**: API Definitions, Terraform Generator, Pulumi CLI Integration, E2E Harness

## Summary

Rebuilt `AwsWafWebAcl` to model the full WAFv2 statement language as a typed recursive tree, forged first-class `AwsWafIpSet` (361) and `AwsWafRegexPatternSet` (362), hardened the Terraform variables generator for recursive proto messages, and wired ALB/CloudFront composability via `web_acl_arn`. Both IaC engines serialize rules to canonical AWS `rule_json` for parity. All 8 live dual-engine E2E lanes green (IP set, regex set, full-surface web ACL, and the ALB association re-run); zero-orphan sweep clean.

## Problem Statement

The original WAF web ACL kind used a hybrid four-statement model plus a Struct escape hatch — adequate for demos but not 90/10 coverage. IP sets and regex pattern sets were documentation-only references. The Terraform variables generator hung on recursive proto descriptors. ALB and CloudFront used inconsistent WAF association field names.

## Solution

### Full statement tree on AwsWafWebAcl

- Recursive `AwsWafWebAclStatement` oneof covering managed groups (with ATP/ACFP/Bot Control configs), rate limiting (custom aggregation keys), set references, geo/byte/SQLi/XSS/size/regex/label/ASN matching, and AND/OR/NOT composition
- Top-level CAPTCHA/challenge immunity, association config, data protection, logging
- `application_integration_url` output
- Terraform module builds `rule_json` via three-level statement unroll; Pulumi module mirrors in Go (`rules.go` + unit tests)

### New composable set kinds

- **AwsWafIpSet** — CIDR-only addresses, IPV4/IPV6 family, REGIONAL/CLOUDFRONT scope with CEL coupling to us-east-1
- **AwsWafRegexPatternSet** — shared regex library (min 1 expression), PCRE-subset honesty documented

### Framework: recursive proto → Terraform types

- Path-scoped `visited` cycle detection in `variablestf.go`
- `TFFreeFormList` / `TFFreeFormMap` when heterogeneous `any` leaves would break Terraform list/map typing
- Regression test using `descriptorpb.DescriptorProto`
- Forge rule 013 uplift documenting the pattern

### Composability seams

- `AwsAlb.spec.web_acl_arn` → `aws_wafv2_web_acl_association`
- `AwsCloudFront.spec.web_acl_id` renamed to `web_acl_arn` (breaking)

### Quality-review hardening (post-implementation deep review)

A four-track review (spec vs the canonical provider source, TF↔Pulumi parity walk, new-kind anatomy vs reference siblings, E2E wiring) closed these gaps before the live runs:

- `go_test` Bazel targets added to both new kinds (spec tests were invisible to CI)
- `deferral_reason` → `deferred_reason` in the three WAF profiles (the wrong key was silently discarded by the protojson loader)
- `size_constraint.size` bound tightened to the provider's real int32 cap
- `iac/pulumi/iac-input.yaml` added to both new kinds (sibling anatomy)
- Reciprocal PARITY-EXCEPTION note in the Pulumi `rules.go` for the TF depth-3 ceiling
- `try()` guards on the TF rate-based custom-key text transformations
- Spec-test additions: JA3/JA4/header-order/uri-fragment field-to-match arms, the documented one-level scope-down nesting boundary, regex-set description bound

### Live-caught defects (fixed in-session)

1. **WAF description charset** — AWS restricts descriptions to `[\w+=:#@/-,.\s]` (no parentheses, no em-dashes) with min length 3; every fixture carrying `(E2E)` failed CreateIPSet/CreateRegexPatternSet at apply. The constraint is now a CEL rule on all three specs (with tests), and every fixture/preset/doc example was sanitized.
2. **64-bit integers inside the rules subtree** — protojson stringifies int64/uint64, so the tfvars-rendered statement tree carried `"Size": "16384"` and the provider's `rule_json` unmarshal failed at apply with "cannot unmarshal string into ... of type int64" (Pulumi, marshaling real Go ints, was unaffected — a silent cross-engine divergence). `size` → int32, `asn_list` → uint32, `immunity_time_sec` → int32; the class is documented in forge rule 013 (no 64-bit integers inside `any`-typed subtrees feeding JSON provider arguments).

## Validation

- Unit tests: generator recursion, all three WAF spec tests (incl. new charset + field-to-match cases), Pulumi `rules_test.go`, outputs conformance, drift guard
- New Bazel `go_test` targets pass (`bazel test`)
- `tofu validate` green on all three TF modules (v6 floor)
- `planton validate-refs --check` green
- Pulumi entrypoints build; E2E verify package builds; `go vet -tags e2e` + `make build-go` green
- Site catalog regenerated (`waf-ip-set`, `waf-regex-pattern-set`, updated `waf-web-acl`)
- **Live dual-engine E2E 8/8 green** (2026-07-08): IP set 28s/43s, regex set 26s/37s, full-surface web ACL 1m09s/1m27s (both set fixtures composed via the e2e-prerequisites chain), ALB + WAF association 5m35s/5m41s. Zero-orphan sweep clean (no web ACLs, sets, load balancers, or e2e VPCs left).

## Breaking Changes

- `AwsWafWebAcl` spec reshaped (recursive statement tree; preset YAML updated)
- `AwsCloudFront.web_acl_id` → `web_acl_arn`
