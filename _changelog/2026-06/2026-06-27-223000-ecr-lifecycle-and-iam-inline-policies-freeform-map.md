# ECR lifecycle policy correctness + free-form JSON map variables (`any`, not `map(any)`)

**Date**: June 27, 2026
**Type**: Bug Fix
**Components**: AWS Provider (AwsEcrRepo, AwsIamRole), Provider Framework (variables.tf generator)

## Summary

Two real-cloud deploy failures from the `aws-ecs-environment` chart are fixed at their source:

1. **`AwsEcrRepo`** generated an invalid ECR lifecycle policy (two rules both selecting untagged
   images) and ignored its own `spec.lifecycle_policy`. The module now builds the policy from the
   spec, emits at most one untagged rule, and creates the policy only when configured.
2. **`AwsIamRole`** could not be applied with more than one inline policy: the generated
   `variables.tf` typed `inline_policies` (`map<string, google.protobuf.Struct>`) as
   `optional(map(any), {})`, and Terraform's `map(any)` rejects heterogeneous entries with
   `attribute "inline_policies": all map elements must have the same type`. The generator now
   types free-form JSON maps as `any`, and the module encodes each policy to JSON at the boundary.

## Problem Statement / Motivation

A real (post-`tofu apply`) deploy of the `aws-ecs-environment` chart failed on two kinds (confirmed
from the stack jobs' diagnostics):

- `AwsEcrRepo` / `ecr-repo`:
  ```
  InvalidParameterException: Invalid parameter at 'LifecyclePolicyText' failed to satisfy constraint:
  'Lifecycle policy validation failure: Only one rule can select Untagged images per storage class.'
  ```
  The module hard-coded two lifecycle rules, both with `tagStatus = "untagged"` — which AWS forbids —
  and never read `spec.lifecycle_policy` (`expire_untagged_after_days`, `max_image_count`), which the
  proto and `variables.tf` already exposed.

- `AwsIamRole` / `default-ecs-task-execution-role`:
  ```
  error: Invalid value for input variable
  ... attribute "inline_policies": all map elements must have the same type.
  ```
  `inline_policies` is `map<string, google.protobuf.Struct>` — each entry is an arbitrary policy
  document. `map(any)` forces every entry to a single common type, so two differently-shaped policies
  (e.g. one with a `Sid` and one statement, another with two statements and none) fail input
  validation. (`trust_policy`, a single `Struct`, already worked because it was `any`.)

## What Changed

### AwsEcrRepo module (`apis/dev/planton/provider/aws/awsecrrepo/v1/iac/tf`)

- `locals.tf` assembles `lifecycle_rules` from `spec.lifecycle_policy`: an untagged expire-by-age rule
  (`sinceImagePushed`, `expire_untagged_after_days`) and an `any` keep-last-N rule
  (`imageCountMoreThan`, `max_image_count`, highest `rulePriority`). Each rule is included only when
  its value is `> 0`, so at most one rule ever selects untagged images.
- `main.tf` creates `aws_ecr_lifecycle_policy` conditionally
  (`count = lifecycle_policy != null && length(lifecycle_rules) > 0`). The failing chart, which sets
  no `lifecycle_policy`, now creates no policy and succeeds.
- `outputs.tf` `lifecycle_policy` output is null-safe for the `count` resource.

### variables.tf generator (`pkg/iac/tofu/generators`)

- New `TFFreeFormMap` type (`tftype.go`) renders as the bare `any` keyword with a `{}` zero default.
- `mapFieldToTFType` (`variablestf.go`) returns it for `map<string, Struct/Value/ListValue>` instead
  of `TFMap{any}`.
- Regenerated `awsiamrole/v1/iac/tf/variables.tf` (drift guard): `inline_policies = optional(any, {})`
  (the only migrated kind with a free-form JSON map; no other kind's `variables.tf` changes).
- `doc.go` documents the rule and the consuming-module contract.

### AwsIamRole module body

- `locals.tf` encodes at the boundary into a homogeneous `map(string)`:
  `inline_policies_json = { for k, v in try(var.spec.inline_policies, {}) : k => jsonencode(v) }`.
- `main.tf` iterates that map (`for_each = local.inline_policies_json`, `policy = each.value`).

## Verification

- `go test ./pkg/iac/tofu/generators/...` — drift guard green; new free-form-map generator tests pass.
- Isolated `tofu plan` with two differently-shaped policies: `optional(map(any), {})` reproduces the
  exact `all map elements must have the same type` error; `optional(any, {})` + the boundary
  comprehension plans cleanly (one resource per policy).
- Isolated `tofu plan` of the ECR lifecycle locals: a configured spec yields exactly one untagged
  rule; an unset spec creates no policy.
- `tofu validate` passes for both modules.
- `awsiamrole` hack manifest extended with a second, differently-shaped inline policy so the fixture
  reproduces this class (the prior single-policy fixture could not).
