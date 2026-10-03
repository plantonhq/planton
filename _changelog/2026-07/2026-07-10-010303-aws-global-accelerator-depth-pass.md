# AWS Global Accelerator at the Full Provider Surface

**Date**: July 10, 2026
**Type**: Enhancement | Breaking Change
**Components**: AwsGlobalAccelerator, aa_e2e verify, pkg/outputs, pkg/iac/tofu/generators, e2e/aws

## Summary

`AwsGlobalAccelerator` (enum 283) is rebuilt to the complete
`terraform-provider-aws` Global Accelerator surface with a presence-honest
spec, two live cross-engine defects fixed, a generator-owned Terraform
contract under the drift guard, and the kind's first-ever E2E coverage —
live dual-engine lanes 4/4 green with a zero-orphan sweep.

## Spec (breaking, presence-honesty restructure)

Four fields moved to presence-honest `optional` shapes because the proto3
zero-value forms made legitimate AWS values unrepresentable:

- `endpoints[].weight` → `optional int32` (an explicit `0` drains one
  endpoint — AWS's documented maneuver; omitted lets AWS default to 128).
  Previously `0` was silently rewritten to 128 by the TF contract and
  dropped by the Pulumi module.
- `traffic_dial_percentage` → `optional double` (an explicit `0.0` drains a
  region; omitted means 100). Previously `0` silently became 100.
- `health_check_port` → `optional int32` 1–65535 (was `gte: 0` fake
  presence).
- `client_ip_preservation_enabled` → `optional bool` (tri-state: omitted
  lets AWS apply its per-endpoint-type default — the provider marks the
  attribute Optional+Computed).

New surface and rule fixes:

- `attachment_arn` on endpoint configurations — cross-account endpoints
  join through a Global Accelerator cross-account attachment by literal ARN.
- `threshold_count` corrected to the provider's real 1–10 (previously
  allowed 0, which AWS rejects).
- The missing `to_port >= from_port` CEL added on port ranges (the comment
  had promised it); `health_check_path` capped at 255 chars; a
  `flow_logs.s3_bucket`-required-when-enabled CEL added (presence-only —
  message CEL cannot dereference `StringValueOrRef` sub-fields);
  `ip_addresses`/`port_ranges`/`port_overrides` limits moved to field-level
  `repeated` bounds.
- `health_check_interval_seconds` carries the API's REAL contract:
  `in: [10, 30]`. The live lane proved the provider's `IntBetween(10, 30)`
  validator is looser than the service — AWS rejects interval 20 with
  "expected values are: 10 or 30". The rule's comment records the rejection.
- `region` documented honestly: Global Accelerator is a global service homed
  in us-west-2; the spec's region acts as the default
  `endpoint_group_region`.

## Both engines at parity

- **Pulumi output defect fixed**: `accelerator_ip_addresses` was a
  `return nil` stub — now a typed flatten across all IP sets, matching the
  Terraform module's export exactly.
- **Live-caught Pulumi defect fixed**: the SDK's
  `AcceleratorAttributesPtr(...)` wrapper trips a runtime input-marshaling
  assertion (compiles clean, panics at deploy); the value
  `AcceleratorAttributesArgs` is assigned directly.
- **Flow-logs disable path fixed in BOTH engines**: the accelerator
  `attributes` block is always materialized with an explicit
  `flow_logs_enabled` value — previously an omitted block was
  diff-suppressed, so disabling flow logs after enabling them silently left
  AWS logging forever.
- **Identity tags converged**: the TF module's bare tag keys (`Resource`,
  `Organization`, …) now match the Pulumi module's `planton.ai/*` set
  key-for-key.
- The hand-written `variables.tf` replaced by the generator-owned contract
  (drift-guard enrolled); provider floor lifted `>= 5.0` → `>= 6.0.0`
  (the newest modeled argument, `attachment_arn`, landed on the v5 line).
- Presets and the hack manifest carried bare-scalar `StringValueOrRef`
  literals (`endpointId:`, `s3Bucket:`) that could never load through the
  proto-YAML layer — converted to the `value:` form; all presets validate.

## First-ever E2E

- `globalaccelerator` AWS SDK added; a state-aware verifier
  (DescribeAccelerator keyed on the ARN, pinned to the us-west-2 global
  control plane; `AcceleratorNotFoundException` = absent).
- Two scenarios: dependency-free `minimal` (endpoint group with no
  endpoints — a valid AWS shape) and `eip-endpoint-composed` (an
  `AwsElasticIp` fixture resolved into the polymorphic `endpoint_id`
  reference via the e2e-prerequisites annotation, plus SOURCE_IP affinity,
  10s interval, dedicated health-check port, explicit weight/dial, and a
  port override). Excluded arms recorded in the scenarios: flow logs,
  BYOIP, DUAL_STACK, attachment_arn.
- `pkg/outputs` conformance case (repeated ip_addresses + the name-keyed
  listener/endpoint-group ARN maps incl. "listener/group" composite keys);
  `TestAwsGlobalAccelerator_Pulumi`/`_Terraform` entrypoints.
- **Live dual-engine E2E 4/4 green** (2026-07-10): Pulumi composed
  2m59s deploy / 1m19s destroy, minimal 2m37s / 2m11s; Terraform composed
  3m02s / 1m16s, minimal 3m13s / 1m26s. Zero-orphan sweep verified
  (0 accelerators, 0 EIPs). Profile green with observed timings.

## Rule uplifts

- `forge/flow/009-pulumi-module`: nested-block args — assign the value
  `XArgs{...}` to `XPtrInput` fields directly; the `XPtr(...)` wrapper is a
  compiles-clean, panics-at-deploy trap.
- `update/update-catalog-kind`: provider schema validators can be
  LOOSER than the cloud API's real contract — never loosen an existing spec
  rule to match a provider validator without a live-API proof; promote the
  API's observed rejection into the spec rule.

## Deliberately deferred (recorded in the kind docs)

Custom-routing accelerator family (distinct AWS resource family, ~5%
adoption — its own candidate kind), the cross-account attachment object
(two-account handshake plane; the consuming `attachment_arn` arm is
modeled), DNS alias records (compose via the exported DNS name + hosted
zone ID with Route 53 records), WAF (attaches to ALB endpoints, the honest
owner), and custom user tags (platform-wide concern).
