# mappingeval — the import-mapping examination system

Import mapping is the judgment half of bringing existing cloud
infrastructure under management: deciding which discovered provider resources
group into which component instances, what each instance's spec says, and
where `value_from` references run between them. Whatever performs that
judgment — a deterministic mapper, an AI mapping agent — its quality must be
**measured, never hoped for**. This package is the measuring instrument.

## The model

```
seed (answer key)              blind side                    grade
──────────────────             ─────────────────────────     ─────────────────
MappingEvalSuite    deploy →   read-only scan (inventory)    Score(gt, proposal)
(known manifests)              → proposer                    grouping / spec /
= GroundTruth                  → ImportMappingProposal       refs / partition /
                                                             coverage
```

- **`MappingEvalSuite`** (`qa.planton.dev/v1`,
  `{provider}/aa_eval/suites/*.yaml`): an ordered list of fixture manifests
  plus the scan scope. A member is a manifest for a **live-proven
  kind** — either an existing E2E scenario or a suite-owned fixture
  (under `suites/<name>/members/`); either way the deployment code is the
  same module the kind's own E2E lane proves, and the suite's live
  lane proves the composition. Suites never invent parallel deployment
  paths. Deploying one yields the **ground truth**: per instance, the
  manifest as authored (references unresolved) and the scan-visible cloud
  resources it owns (read from its IaC state, translated to scan
  coordinates through the provider import catalog's
  `cloud_control_type_name`).
- **`inventory.Scanner`**: the blind side's only window — Cloud Control
  list/get keyed by CloudFormation type names, mirroring the platform's
  inventory capability shape, with **declared per-type enrichments**
  (S3 bucket regions, route-table routes/associations, internet-gateway
  attachments) closing Cloud Control's model gaps with typed read-only SDK
  calls. Structurally read-only: no mutating code path exists.
- **`ImportMappingProposal`** (`iac.planton.dev/v1`): the proposer's answer
  — proposed **manifests** (not a parallel encoding; validation parses each
  into its typed kind with unknown fields rejected) plus per-instance
  **claims** of the discovered resources it accounts for, and the honest
  `unmapped` remainder. This contract is the seam a future AI mapper must
  emit through; the review surfaces a human approves consume the same
  shape. Nothing is ever created from a proposal without an explicit human
  gate.
- **`baseline.Propose`**: the deterministic reference mapper. Pinned to a
  **perfect score** on network-staples and to a **specific imperfect
  score** on each of the other exams (see "The exams") — any drift in any
  pin is a harness, recipe, or scanner regression, never model variance.
  Its mappers are deliberately bounded to the staples' kinds; generalizing
  hand-written mapping code is exactly the infeasible work the AI proposer
  exists to replace.
- **`Score`**: the grader. Entirely structural — driven by the kinds' own
  proto schemas and the shared `StringValueOrRef` encoding — so it works
  for any kind on any provider with zero per-kind grading code.

## The exams

| Suite | Purpose | Baseline pin |
|-------|---------|--------------|
| `network-staples` | proves the **instrument**: on a clean, well-signaled account, the whole chain (seed, scan, propose, score) works end to end | PERFECT |
| `messy-account` | measures **headroom**: an account shaped like a real company's — look-alike prod/staging networks (identical CIDRs), an uncovered tier (security group, KMS key + alias, DynamoDB, ECR), cross-service `value_from` edges into the KMS key. The ONE suite that grades the **partition axis** (its member names carry recoverable env tokens) | a specific IMPERFECT report — **a floor** |
| `identity-and-egress` | completes **coverage**: the last two import-recipe kinds (IAM role, NAT gateway) under examination — with it, **every kind that carries an import recipe appears in at least one exam**. Adds the first global-service scan (IAM lists the whole account; noise is the norm), the first `nat_gateway` route target, and a second name-derived-identity kind beside S3 | a specific IMPERFECT report — **a floor** |

Each floor's authoritative record is its pinned offline test
(`TestBaselineFloorOnMessyAccount`, `TestBaselineFloorOnIdentityAndEgress`)
and each live run's report artifact. **Beating a floor** means strictly
higher grouping/spec/refs/partition recall without introducing what the
baseline never produces: duplicate claims, misassignments, wrong-target
edges, wrong environments, unaccounted resources, or name-derivability
breaks. The baseline is honest
even where it is blind — everything it cannot map is *declared* unmapped,
never silently dropped — and that honesty is part of the bar. (There is
deliberately no blended numeric score or comparator machinery yet; it
becomes worth building when a second proposer exists.)

### Global services and scan noise

Region scopes nothing on a global service: scanning `AWS::IAM::Role` lists
every role in the account — service-linked roles, other projects, the scan
runner's own identity plumbing. This is not a problem to filter away; it is
what a stranger's account genuinely looks like on that tier, and the
universe-only honesty rules below are exactly what keep the grade
trustworthy through it. The identity-and-egress fixture deliberately
carries service-linked and foreign roles so the offline floor proves it.

### Exam-fairness rule for suite members

Every spec field a member sets must be reconstructable from the
scan-visible + enriched surface. A field driving scan-invisible state (the
staples lane's SNS-policy lesson) would make the exam count a leaf NO
scan-driven proposer could ever match — permanently deflating the ceiling
instead of testing judgment. When a new member's field turns out invisible
on the live lane, either drop the field from the member or add a declared
enrichment.

### Seeding-fingerprint redaction

Planton's own modules tag everything they deploy
(`planton.ai/resource-kind`, `planton.ai/environment`, ...) — on a seeded
exam account those tags ARE the answer key. Both eval lanes (offline and
live) apply `RedactSeedFingerprints` between scan and proposer, so the
account presents exactly as a stranger's: `Name` tags and realistic user
tags stay; only the deploy machinery's identity tags leave. The redaction
is a property of the PIPELINE, not of any recorded fixture, and it is never
a general tag scrubber.

## The axes

| Axis | Question | Unit |
|------|----------|------|
| grouping | did each discovered resource land in the right instance? | claims vs ground-truth ownership (precision + recall) |
| spec | do the manifests reconstruct the declared settings? | recall over the leaves the ground-truth spec sets |
| refs | are dependencies wired as references, not frozen literals? | edges at the same spec location, targets translated through instance matching; an UNPROPOSED instance's edges stay in the denominator (skipping a resource never shrinks the ref debt) |
| partition | did each manifest land in the right environment? | each proposed manifest's `metadata.env` vs the ground truth's, over the ground-truth instances that set one (unproposed instances owe theirs). Wrong-env is the worst class (it poisons env-scoped references); honest unassignment reads as missing; proposed-env-where-none is informational |
| coverage | is everything accounted for? | claimed / declared-unmapped / **unaccounted** (the silent gap) |

The partition axis is **suite-declared** (`grade_environment_partition` on
the suite spec) — an exam-fairness call, never inferred: fixture manifests
also carry operational environments (e2e bookkeeping) that leave no
scan-visible trace once seeding fingerprints are redacted, and owing those
would put debt in the denominator no proposer could ever pay. The baseline
recovers environments through the deterministic partition engine
(`pkg/iac/envpartition`, untaught default rule — name tokens and
containment; the authoritative-tag tier is exactly what redaction removes
on a seeded account) and stamps `metadata.env` only where the engine
assigned. It never guesses an environment, exactly as it never guesses a
grouping.

Plus one declared name rule: names are never scored (instances match by
kind + claim overlap — a blind proposer cannot know internal names), except
where a kind's import recipe derives its import id `from_metadata_name`
(S3: the manifest name IS the bucket name; IAM roles: the manifest name IS
the role name) — breaking that derivation breaks the downstream zero-typing
import and is flagged.

One encoding note on the spec axis: a repeated `StringValueOrRef` field
(e.g. a role's managed policy ARNs) participates exactly like its singular
form — its literal arms are ONE whole-list spec leaf (mirroring scalar
lists, so element count never inflates the denominator) and its
`value_from` arms are edges, the refs axis's business.

## Honesty rules (what keeps scores trustworthy on a shared account)

- **The scored universe is the ground truth only.** A real region also
  contains AWS-implicit resources and unrelated infrastructure; proposals
  about those are reported for information, never rewarded or penalized.
- **Structural invisibility is declared, not discovered.** A resource type
  with no `cloud_control_type_name` in the provider catalog cannot be seen
  by any scan, so it never enters the grouping denominator (the S3
  satellites are properties of the bucket's model, not scannable
  resources).
- **Config-only spec fields leave the spec axis** (declared
  `config_only_attributes`): they exist only in IaC configuration, so
  expecting a scan-driven proposer to produce them would penalize physics.
- **Proposer-extra spec fields are informational**: materializing a
  cloud-observed default explicitly is honest, not wrong.

## Proofs

- **Offline** (`harness_test.go` + `messy_floor_test.go`, run in
  `make test`, creds-free): recorded-shape scan fixtures drive the
  baseline, whose proposal must score perfect on the staples fixture and
  exactly the pinned floor on the messy fixture — and hand-mutated
  proposals prove every axis **discriminates** (mis-grouping,
  literal-instead-of-ref, wrong spec value, duplicate claims, silent gaps,
  dropped ref-carrying instances, name-derivability breaks, wrong and
  missing environments each produce their specific finding). An exam
  nothing can fail is not an exam.
- **Live** (`TestMappingEval_NetworkStaples` +
  `TestMappingEval_MessyAccount` + `TestMappingEval_IdentityAndEgress`,
  opt-in via `PLANTON_E2E_MAPPING_EVAL=1`): the same chain against a real
  account — deploy the suite, scan blind, redact fingerprints, propose,
  score, assert the suite's pin (perfect / its floor), destroy everything.
  Create-and-destroy with ambient credentials; artifacts (scan, proposal,
  report per suite) recorded via `PLANTON_E2E_MAPPING_EVAL_ARTIFACTS`.

## Plugging in a proposer

A proposer is anything that turns a `Scan` into an `ImportMappingProposal`.
The contract is enforced at `ParseProposal` (typed-kind parsing, non-empty
claims, no dangling references); the grader consumes only the validated
form. An AI mapper slots into exactly the baseline's seat and is graded by
exactly the same machinery — that symmetry is the point.
