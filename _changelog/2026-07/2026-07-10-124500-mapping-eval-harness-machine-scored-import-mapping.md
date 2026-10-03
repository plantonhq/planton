# Mapping eval harness: import mapping quality is now machine-scored

**Date:** 2026-07-10
**Scope:** `apis/dev/planton/iac/importmappingproposal/` (new kind), `apis/dev/planton/qa/mappingevalsuite/` (new kind), `apis/dev/planton/iac/providerimportcatalog/` (+ `cloud_control_type_name`), `apis/dev/planton/provider/aws/aa_eval/` (first suite), `pkg/iac/mappingeval` (contract, scanner, baseline, scorer), `e2e/framework/mappingeval` (suite deployer + ground truth), `e2e/aws` (live lane)

## Summary

Import **mapping** — grouping discovered infra components into kind
instances, reconstructing their specs, and wiring `value_from` references —
now has an examination system: seed an account from known manifests (the
answer key), scan it back blind through a read-only channel, have a
proposer emit its mapping, and machine-score the proposal on grouping,
spec, and reference accuracy. Any proposer takes the same exam and gets the
same impartial grade; the grader itself is entirely structural (driven by
the kinds' proto schemas and the shared `StringValueOrRef` encoding), so it
works for any kind on any provider with zero per-kind grading code.

## The proposal contract (`ImportMappingProposal`, `iac.planton.dev/v1`)

A proposal is **proposed manifests plus accounting**: each proposed
resource carries the full KRM manifest it would create (validated by
parsing into the kind's typed message, unknown fields rejected) and the
discovered account resources it claims to cover; everything unmappable
lands in `unmapped` with a reason. Dangling references fail at the
contract. This one shape serves the eval scorer today and is the exact
output seam a future mapping agent must emit through.

## The scan-side correspondence (`cloud_control_type_name`)

`ProviderImportCatalog` entries gain the CloudFormation type name under
which a read-only scan (Cloud Control) reports the resource — the declared
translation between what a scan sees and what IaC state holds. Declared
ONLY for types empirically verified listable (16 AWS types today); a
covered type may carry none because it is unlistable (`AWS::IAM::RolePolicy`
does not support LIST), needs scoping input (`AWS::EC2::VPCCidrBlock`), or
exists only as properties of a parent model (the S3 satellites). The
conformance guard validates the pattern and that no two terraform types
claim the same scan-side name.

## The harness (`pkg/iac/mappingeval` + `e2e/framework/mappingeval`)

- **`MappingEvalSuite`** (`qa.planton.dev/v1`,
  `{provider}/aa_eval/suites/*.yaml`): ordered fixture members (existing,
  E2E-proven scenarios only) + the scan scope. The deployer runs every
  member through the one terraform arm in listed order (references resolve
  against earlier members, exactly like the E2E prerequisite chain) and
  builds the ground truth from each member's IaC state.
- **Read-only scanner** (`mappingeval/inventory`): Cloud Control list/get
  mirroring the platform inventory shape, with declared per-type
  enrichments (S3 bucket regions, route-table routes/associations,
  internet-gateway attachments) closing model gaps via typed read-only SDK
  calls. No mutating code path exists.
- **Deterministic baseline proposer** (`mappingeval/baseline`): plain-code
  mapping for the suite kinds, pinned to a PERFECT score — any drop is a
  harness/recipe/scanner regression, never model variance — and the floor
  an AI mapper must beat. Deliberately bounded: generalizing hand-written
  mappers is exactly the work an AI proposer exists to replace.
- **Scorer**: grouping (claims vs ground-truth ownership), spec (recall
  over the leaves the ground-truth manifests set), refs (edges at the same
  spec location, targets translated through claim-based instance matching
  — names carry zero score weight, with one declared exception: a
  `from_metadata_name` import recipe makes the proposed name load-bearing),
  coverage (claimed / declared-unmapped / the silent unaccounted gap).
  Honesty rules: scoring is defined over the ground-truth universe only;
  undeclared-scan-side types are structurally invisible; config-only spec
  fields leave the spec axis; proposer-extra fields are informational.

## Proofs

- **Offline** (`make test`, creds-free): the recorded-shape scan fixture
  drives the baseline to a pinned perfect score against the ground truth
  assembled from the real `network-staples` suite manifests, and mutated
  proposals prove every axis discriminates (mis-grouping,
  literal-instead-of-ref, wrong spec value, duplicate claims, silent gaps,
  name-derivability breaks each produce their specific finding).
- **Live** (`PLANTON_E2E_MAPPING_EVAL=1`, `TestMappingEval_NetworkStaples`):
  the full chain against a real account — deploy the 6-member suite
  (VPC, internet gateway, routed subnet, S3, SQS, SNS), scan blind, propose,
  score PERFECT, destroy everything. Create-and-destroy, ambient
  credentials, artifacts recorded.

The live lane earned its keep on its first runs by catching two real
blind-side defects the offline fixture could not see: enriched properties
were stored in Go-typed shapes (`[]map[string]any`) where every other
property is JSON-generic (`[]any`), silently failing consumers' type
assertions on live scans only — now an explicit package invariant — and
the SNS topic access policy is invisible to Cloud Control (CloudFormation
models it as a separate TopicPolicy resource), which became the fourth
declared enrichment. The suite deployer also gained a shared
`TF_PLUGIN_CACHE_DIR` so sequential members stop re-downloading the same
provider (the observed mid-suite download flake).

## Breaking changes

None. Both new kinds are additive; the catalog field is optional and empty
where not declared; no existing package's API changed.
