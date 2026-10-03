# Import-ID recipes: proto-backed import knowledge with a live round-trip proof

**Date:** 2026-07-09
**Scope:** `apis/dev/planton/iac/` (new), `apis/dev/planton/provider/aws/` (catalog + 4 kind maps), `pkg/iac/importmap` (new), `e2e/framework/runner` (new IMPORT-RT phase)

## Summary

State import has always demanded values only module authors know: the engine
address of each resource and the provider identifier it imports by. This
change makes that knowledge a first-class, machine-proven artifact class.
Addresses are deliberately NOT authored — the engine enumerates them per
spec at import time, which handles constructed names, `for_each`/`count`
instances, and conditional resources by construction. Only the identifier
knowledge is authored, split into a provider tier (ID format per resource
type) and a thin kind tier (which spec field / output / ARN part /
address key supplies each value), both proto-backed KRM YAML in the
`iac.planton.dev/v1` group.

## New API kinds (`apis/dev/planton/iac/`)

- `ProviderImportCatalog` — per provider, the import-ID format of every
  resource type its kinds import by, plus `config_only_attributes`:
  attributes that exist only in IaC configuration (e.g.
  `aws_s3_bucket.force_destroy`) and therefore can never round-trip through
  an import.
- `CatalogKindImportMap` — per kind, the ordered derivations for each
  `{placeholder}` the formats reference, with mandatory "where to find this"
  guidance when nothing can derive a value.

## Authored recipes (AWS pilot)

`aws/aa_import/catalog.yaml` covers the full S3 satellite family, VPC (+
secondary CIDR association), security group, and ECR; kind maps ship
for `awss3bucket` (fully derivable from `metadata.name` + address keys),
`awsecrrepo` (`spec.repository_name`), `awsvpc` and `awssecuritygroup`
(discovered ids via outputs or one pasted ARN).

## `pkg/iac/importmap` (new)

Loader (through `pkg/protobufyaml`), placeholder extraction and strict ID
rendering (a missing value is an error, never an empty substitution),
derivation resolution, and OpenTofu address parsing. See the package README.

## E2E: the IMPORT-RT phase

Opt-in (`PLANTON_E2E_IMPORT_ROUNDTRIP=1`), between VERIFY-RES and DESTROY:
set the deployed state aside, re-import every resource blind through the
recipes, then require the follow-up plan to propose no real change — the
plan JSON is inspected structurally (creates/destroys/replaces always fail;
in-place updates pass only when every changed attribute is declared
config-only). The destroy that follows runs through the re-imported state,
proving the blind import fully owns the resources.

## Guards

- `TestImportMapConformance` (offline, rides `make test`): recipe files
  parse against their protos; every module-declared resource type is
  mapped; every placeholder is declared; spec-field paths and output
  keys resolve against the kind's protos. Allowlist enrollment mirrors
  `TestVariablesTFDrift`.
- Live: `TestAwsS3Bucket_Terraform` with the round-trip enabled — both
  scenarios PASS: 13 resources re-imported blind across minimal +
  full-surface (incl. the intelligent-tiering `{bucket}:{name}` composite
  and `for_each` instance keys), zero real plan changes, one tolerated
  `force_destroy` config-only update per scenario, clean destroy through
  the re-imported state.

## Breaking changes

None. New artifact class, new package, opt-in E2E phase.
