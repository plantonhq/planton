# Import recipes cover the AWS staples, every one proven by a live round-trip

**Date:** 2026-07-10
**Scope:** `apis/dev/planton/provider/aws/` (catalog + 8 new kind maps + ECR e2e fixture), `apis/dev/planton/iac/` (recipe vocabulary), `pkg/iac/importmap` (auto-discovery + optional segments), `e2e/framework/runner` (round-trip tolerance classes + account-level ARN parts), `_rules/catalog-kind/` (forge/update carry the recipe step)

## Summary

The import-recipe library grows from 4 to 12 AWS kinds — the resources a
console-managed account is actually full of — and, unlike the pilot (where
only S3 had run the live proof), **every mapped kind now has a green
round-trip lane**: deploy a real fixture, set its state aside, re-import
every resource blind through the recipes, and require the follow-up plan to
propose no real change. Enrollment discipline is restructured so recipe
coverage can scale without silent rot, and the kind forge/update
workflows now carry the recipe step by default.

## New recipes (all live-proven 2026-07-10; ledger in `pkg/iac/importmap/README.md`)

- **awssubnet** — subnet, its owned route table, and the
  `{subnet_id}/{route_table_id}` association composite. A new `routed`
  scenario deploys all three so the composite is proven, not assumed.
- **awsinternetgateway**, **awsnatgateway** — AWS-assigned gateway ids from
  outputs.
- **awsiamrole** — the role by name (`metadata.name` basis), inline policies
  by `{role_name}:{inline_policy_name}` and managed attachments by
  `{role_name}/{managed_policy_arn}` — both second segments ride the
  module's `for_each` keys via `from_address_key`.
- **awsdynamodb** — table by name; resource policy by table ARN;
  contributor insights by
  `name:{table_name}/index:{index_name?}/{account_id}` (table-level AND
  GSI-level variants).
- **awssqsqueue** (queue URL), **awssnstopic** (topic ARN), **awskmskey**
  (key UUID + `alias/...` names via `for_each` keys).

## Pilot debt closed

- `awsvpc` and `awssecuritygroup` ran their (previously unrun) lanes; the SG
  lane surfaced `revoke_rules_on_delete` as a config-only attribute, now
  declared.
- `awsecrrepo` had no E2E fixture at all: it gained `e2e/profile.yaml` +
  `scenarios/full-surface.yaml` (repository + lifecycle policy), a
  DescribeRepositories verifier, and `TestAwsEcrRepo_*` entry points — then
  ran green on BOTH engines. Two real defects fixed on the way: the module's
  provider pin (`= 5.82.0`) contradicted its committed lock file (6.53.0),
  and `concat()` of the two lifecycle-rule shapes unified their objects to
  `map(string)`, silently stringifying `countNumber` so the ECR API rejected
  every lifecycle policy — rules now compose as a null-filtered tuple, with
  the why documented inline.

## Recipe vocabulary (small, provider-documented extensions)

- **Optional ID segments**: `{name?}` marks a segment the provider documents
  as legitimately empty in some variants (DynamoDB table-level contributor
  insights: `name:tbl/index:/123456789012`). Required segments still abort
  rather than render empty.
- **`from_arn_part: arn`**: the full pasted ARN, for types whose import ID
  IS the ARN (SNS topics, DynamoDB resource policies).
- **`write_normalized_attributes`** (provider catalog, beside
  `config_only_attributes`): attributes the cloud normalizes on write —
  policy documents read back semantically identical but textually different
  (the provider's own import tests ignore them). The round-trip tolerates
  in-place updates on either declared class; everything else still fails.
- **Count indexes are not identity**: `ParseTofuAddress` (and the platform
  wizard's mirror) now yield an address key only for QUOTED `for_each` keys;
  a numeric `[0]` is positional and must never leak into an import ID.

## Enrollment discipline: the file IS the signal

`TestImportMapConformance` now auto-discovers every
`{kind}/v1/iac/import-map.yaml` (`DiscoverCatalogKindImportMaps`) instead
of a hand-maintained allowlist — previously a brand-new map got ZERO offline
validation until someone remembered to enroll it, while the platform's
catalog bundler shipped it to users regardless. File presence is now the
single enrollment signal for the offline guard, the live round-trip gate,
and the catalog bundler alike. "Live-proven" is a README ledger plus the
no-unproven-merges rule; the round-trip lane itself is the only honest
enforcement. Drift errors now tell the next kind author exactly which
file to fix.

## Round-trip harness

- Populates the ACCOUNT-LEVEL ARN parts (`account_id`, `region`) from the
  deployment's own ARN-shaped outputs — deployment-level facts the platform
  derives from the connection in the real flow — so recipes embedding the
  account id stay blind-derivable. Per-resource parts are deliberately never
  synthesized.
- Only REQUIRED placeholders block a blind import; optional segments render
  empty, per the provider-documented variant shapes.

## Workflow rules

`forge-catalog-kind.mdc` gains Phase 6c (import recipes: catalog
entries + kind map + the live round-trip as a merge gate) and the
success-criteria line; `update-catalog-kind.mdc`'s final validation
covers the module-resource-set-changed case with the conformance guard as
its tripwire.

## Breaking changes

None. Recipes are additive data; the conformance guard's auto-discovery
validates strictly more than the allowlist did; the vocabulary extensions
are backward-compatible (existing formats carry no optional markers, and the
new catalog field is empty everywhere it is not declared).
