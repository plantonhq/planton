# AwsBedrockInferenceProfile — Kind Guide

Authored operational judgment for the Bedrock application inference
profile component.

## Design decisions

- **`source_arn` is a leaf, not a block.** The provider nests the source
  under a one-element `model_source { copy_from }` block; the spec carries
  the ARN directly (the parity manifest records the flattening). The CEL
  pattern admits both legal shapes — a foundation-model ARN and a
  system-defined inference-profile ARN.
- **The foundation-model ARN shape works only for ON_DEMAND-capable
  models** (live-verified 2026-08-13): CreateInferenceProfile rejects a
  direct reference to a profile-only model — the Nova family and most
  2025+ releases — with "The provided foundation model does not support
  On Demand inference". Reference those through their system-defined
  inference-profile ARN (`arn:...:<account>:inference-profile/us.<model>`)
  instead; check a model's arms with
  `aws bedrock list-foundation-models --by-inference-type ON_DEMAND`.
- **AWS never echoes the source back.** GetInferenceProfile reports the
  RESOLVED models list instead; the provider pins the configured value in
  state (and ignores it on import). Taught on the spec field.
- **The kind creates APPLICATION profiles only.** SYSTEM_DEFINED profiles
  are AWS-owned surface — consumed by reference (as a `source_arn`),
  never managed.

## Operational judgment

- **Name profiles after their consumer** (service, team, tenant) — the
  name is the id, and the whole point is per-consumer attribution.
- **Scope IAM to the profile ARN.** Granting `bedrock:InvokeModel` on the
  profile (not the model) is what makes usage attribution enforceable
  rather than advisory.
- **Plan replacements as consumer migrations.** Any spec change replaces
  the profile and its ARN; roll the new ARN to consumers like a
  credential rotation.

## Coverage decisions

- Every configurable argument of `aws_bedrock_inference_profile` at the
  pinned provider is modeled, mapped, or excluded with a reason in
  `iac/provider-parity.yaml` (zero findings at forge time).
- The cross-region source shape is not a separate live lane: `copy_from`
  is one string forwarded verbatim, and the system-profile ARN embeds the
  caller's account id, which a committed manifest cannot carry honestly
  (recorded in e2e/profile.yaml).
