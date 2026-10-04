# Messy-account eval suite: the mapping exam with headroom, and the floor a smarter mapper must beat

**Date:** 2026-07-11
**Scope:** `apis/dev/planton/provider/aws/aa_eval/suites/` (second suite + suite-owned members), `pkg/iac/mappingeval` (fingerprint redaction, refs-axis completeness fix, pinned floor test), `e2e/aws` (second live lane)

## Summary

The mapping eval harness gains its second exam. Where `network-staples`
proves the measuring instrument (the deterministic baseline is pinned to a
PERFECT score on it), the new `messy-account` suite measures **headroom**:
an account shaped like a real company's, which the baseline by design
cannot ace. Its pinned imperfect report is **the floor** — any smarter
proposer earns its place by beating it, and the floor's authoritative
record is the pinned offline test (`TestBaselineFloorOnMessyAccount`) plus
each live run's report artifact.

## What makes the account messy, deliberately

- **Look-alike networks**: prod and staging VPCs with IDENTICAL CIDRs and
  settings, each containing a subnet with the same AZ + CIDR —
  distinguishable only by Name tags and containment. The prod subnet is
  routed (a three-resource instance: subnet + owned route table +
  association); its staging twin is bare, the asymmetry a half-mirrored
  staging environment genuinely has.
- **An uncovered tier** the baseline has no mappers for: a security group
  (`vpc_id` edge into the right look-alike), a KMS key + alias (grouping
  needs the alias's TargetKeyId), a DynamoDB table (rich spec + a
  `kms_key_arn` edge), an ECR repository (whose lifecycle policy module
  resource is structurally invisible to any scan).
- **A cross-service edge on a COVERED kind**: the SQS queue is mapped by
  the baseline, but its `kms_key_id` edge into the customer-managed key is
  an inference no deterministic mapper attempts — the refs axis
  discriminates even inside covered territory.
- **A name-derived kind** (S3) so the name-derivability rule keeps applying.

## The floor (pinned offline AND live, both green)

Grouping 9/14 correct (the 5 uncovered-tier resources unclaimed — and
DECLARED unmapped, never silently dropped); spec 21/64 leaves (zero
mismatches; every missing leaf belongs to an instance the baseline never
proposed); refs 4/7 edges (the three missing: the queue's kms edge and the
two owed by unproposed instances); zero unaccounted, zero misassigned,
zero duplicate claims, zero wrong-target edges, zero name breaks. Beating
the floor means strictly higher grouping/spec/refs recall WITHOUT
introducing any of those classes — the baseline's honesty where it is
blind is part of the bar.

## Two harness honesty fixes

- **Seeding-fingerprint redaction** (`RedactSeedFingerprints`): Planton's
  own modules tag everything they deploy (`planton.ai/resource-kind`,
  `planton.ai/environment`, ...) — on a seeded exam account those tags ARE
  the answer key. Both eval lanes now strip exactly the platform's seeding
  fingerprints between scan and proposer (`planton.ai/*`, `e2e-catalog-kind`,
  and `managed-by` only when it carries the e2e marker value); Name tags
  and realistic user tags stay. The redaction is a property of the
  pipeline, never a general tag scrubber, and the tags remain on the provider
  resources for fixture sweeps.
- **Refs-axis completeness**: an unproposed instance's `value_from` edges
  now stay in the denominator as missing edges (mirroring the spec axis) —
  skipping a resource no longer shrinks a proposer's ref debt. New
  discrimination test pins it.

## Suite-composition rule amended

A suite member is a manifest for a **live-proven kind** — either an
existing E2E scenario or a suite-owned fixture (`suites/<name>/members/`).
Either way the deployment code is the same module the kind's own E2E
lane proves; the suite's live lane proves the composition. Suites never
invent parallel deployment paths. (The messy account could not be
assembled from the tidy existing scenarios without polluting every
kind lane with exam-specific fixtures.)

## Exam-fairness rule for members

Every spec field a member sets must be reconstructable from the
scan-visible + enriched surface — a field driving scan-invisible state
would make the exam count a leaf NO scan-driven proposer could ever match,
permanently deflating the ceiling instead of testing judgment. The
DynamoDB member deliberately omits `resourcePolicy`/`contributorInsights`
(their satellites may not surface on Cloud Control's table model); the
live lane confirmed every field the members DO set is visible.

## Verification

- Offline (`make test`, creds-free): 15 mappingeval tests green, including
  the pinned floor, the redaction keep-vs-strip pin, and the
  unproposed-instance-edges discrimination test.
- LIVE lane green (create-and-destroy, e2e account, ambient SSO):
  `TestMappingEval_NetworkStaples` still PERFECT (regression) and
  `TestMappingEval_MessyAccount` matched the pinned floor exactly —
  11 fixtures deployed, scanned blind, redaction proven on the real scan,
  all fixtures destroyed (leak check clean; the KMS key exits via AWS's
  mandatory scheduled-deletion window).
- `make gazelle`, `go build ./pkg/...`, `make e2e-vet`, `make e2e-build`,
  import-map conformance guard — all green.
