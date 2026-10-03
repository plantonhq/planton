# AwsBedrockProvisionedThroughput — Kind Guide

Authored operational judgment for the Bedrock Provisioned Throughput
component.

## Design decisions

- **`model_arn` defaults its reference to AwsBedrockCustomModel** — the
  dominant real-world use (fine-tuned models CANNOT serve on-demand), with
  the literal arm open for foundation-model ARNs where AWS allows direct
  provisioning.
- **Omitted `commitment_duration` IS the no-commitment purchase.** The
  provider treats the absent argument as hourly billing; the spec mirrors
  that exactly rather than inventing a NONE value AWS doesn't define.
- **The cost warning lives on the spec.** This is the catalog's most
  financially dangerous small resource — the spec header, the field
  comments, both module headers, and the catalog page all carry the
  billing truth, because the reader who misses it pays for it.

## Operational judgment

- **Quota before purchase.** The account's no-commitment model-unit quota
  is often 0 — a create that fails on quota is the GOOD failure mode;
  raise it via Service Quotas deliberately.
- **Committed terms are financial commitments, not infrastructure.**
  OneMonth/SixMonths bill in full and refuse deletion until the term
  lapses — treat them like reserved instances: a procurement decision
  recorded in a manifest.
- **Watch utilization.** Provisioned capacity bills whether used or not;
  CloudWatch's provisioned-throughput utilization metrics tell you when
  to resize (which is a replacement — plan a capacity overlap).

## Coverage decisions

- Every configurable argument of
  `aws_bedrock_provisioned_model_throughput` at the pinned provider is
  modeled, mapped, or excluded with a reason in
  `iac/provider-parity.yaml` (zero findings at forge time).
- The live E2E lane is a recorded deferral — the upstream provider itself
  annotates the resource "Testing is cost-prohibitive" and skips its own
  acceptance tests (see e2e/profile.yaml for the unblock). Offline plans
  and previews prove both engines' renders arm-for-arm.
