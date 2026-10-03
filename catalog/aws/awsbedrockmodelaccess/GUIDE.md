# AwsBedrockModelAccess — Kind Guide

Authored operational judgment for the Bedrock model access component.

## Design decisions

- **The spec asks for the model, never the offer token.** Offer tokens
  are short-lived credentials from ListFoundationModelAgreementOffers —
  both modules look the PUBLIC offer up at deploy time (the Terraform
  data source / the Pulumi invoke), so a manifest stays a timeless
  document. The token is excluded in the parity manifest for exactly this
  reason.
- **The use-case form is modeled but scoped honestly.** It is an
  ACCOUNT-GLOBAL singleton (Put-convergent, no delete, errors on
  differing content). Folding it as an optional arm serves the real
  workflow — Anthropic access needs it once — while the spec comment,
  both docs, and the E2E profile all state the one-owner-instance rule.
  A future AwsBedrockInvocationLogging-style account-settings kind is the
  escape hatch if multi-form contention ever materializes.
- **Region-singleton observability was split out.** The founding
  disposition folded `aws_bedrock_model_invocation_logging_configuration`
  into this kind; schema evidence (a region singleton keyed by the region
  itself) re-judged it to a future AwsBedrockInvocationLogging kind — a
  per-model kind must not own region-wide state (the SES account-settings
  precedent).

## Operational judgment

- **Agreements gate everything downstream.** Wire agents, throughput
  purchases, and profiles to deploy AFTER the access component (the
  `model_id` output exists for chart ordering).
- **Check the offer before automating a new model.**
  `aws bedrock list-foundation-model-agreement-offers --model-id <id>`
  shows whether a public offer exists in the region and its terms.
  Probe-verified classes (2026-08): auto-enabled models (Mistral 7B,
  Llama 3, Amazon first-party) REJECT the call with "Agreement not
  supported for this model" — they need no agreement at all; Anthropic
  models reject it with "account is not authorized" until the use-case
  form is on file; Cohere models list a plain usage-based public offer.
- **Treat destroy as a revocation event**, not cleanup: in-flight
  workloads lose the model immediately.
- **The agreement waiter is real** — create polls to AVAILABLE (commonly
  seconds to a few minutes; the provider allows 30) and delete retries
  through AWS's transient ConflictException window.

## Coverage decisions

- Every configurable argument of `aws_bedrock_foundation_model_agreement`
  and `aws_bedrock_use_case_for_model_access` at the pinned provider is
  modeled, mapped, or excluded with a reason in
  `iac/provider-parity.yaml` (zero findings at forge time).
- The use-case-form arm is offline-proven only (account-global write-once
  state must not be stamped by a test lane; recorded in e2e/profile.yaml
  with its unblock).
