# AWS Analytics Pair: Athena Workgroup and Glue Catalog Database at the Full Provider Surface

**Date**: July 9, 2026
**Type**: Feature
**Components**: API Definitions, Terraform Modules, Pulumi Modules, E2E Harness

## Summary

The two thinnest kinds in the AWS catalog reach the full provider surface. **`AwsAthenaWorkgroup`** grows from a 14-field spec to the complete workgroup model — AWS-managed query results, IAM Identity Center trusted identity propagation, S3 Access Grants, Spark content encryption, the three-arm log-delivery surface, and the enable/disable state dial. **`AwsGlueCatalogDatabase`** grows from three fields to the complete database model — catalog metadata parameters, Lake Formation create-table default grants, cross-account/cross-region resource links, and Redshift-datashare federation. Both kinds move to generator-owned Terraform contracts under the drift guard, converge cross-engine identity tags, and ship first-ever E2E artifacts with live dual-engine runs green. All changes are additive — no manifest breaks.

## What Was Built

### `AwsAthenaWorkgroup` — the complete query-governance surface

- **`managed_query_results`** — AWS-managed result storage: no S3 bucket to create, secure, or lifecycle; results retained 24 hours and retrieved through Athena APIs. Block presence is the enable switch; optional customer KMS key. AWS's own exclusivity rule (managed results cannot combine with an S3 `output_location`) is promoted to a spec CEL, so the conflict fails at manifest validation instead of apply.
- **`identity_center` + `s3_access_grants`** — trusted identity propagation (queries run and audit as the signed-in workforce user) and per-user result access through S3 Access Grants (`DIRECTORY_IDENTITY`, optional per-user prefix). Create-time settings, stated honestly.
- **`monitoring`** — the three log-delivery arms (CloudWatch Logs with worker-type log selection, Athena-managed storage, S3 archive), each enabled by presence, combinable.
- **`customer_content_encryption_kms_key`** — customer KMS encryption for Spark notebook and session content.
- **`description` and `state`** — the TF module previously hardcoded both (`""` / `ENABLED`); `state: DISABLED` is the pause switch that rejects new queries while keeping configuration, history, and saved queries.
- **Cross-engine parity fix**: the Pulumi module silently dropped `enable_minimum_encryption_configuration` behind a stale "not in pinned SDK v7.3.0" deferral comment — the pinned pulumi-aws v7.35.0 carries the full surface; the arm is wired and the deferral class re-checked across the module.

### `AwsGlueCatalogDatabase` — the complete catalog-namespace surface

- **`parameters`** — catalog metadata properties read by engines and governance tooling (distinct from AWS resource tags).
- **`create_table_default_permissions`** — default Lake Formation grants on tables created in the database, including the empty-permissions entry that disables the `IAM_ALLOWED_PRINCIPALS` grant (the hardening step when moving to Lake Formation-managed access). Permission values CEL-validated against the Glue permission set.
- **`target_database`** — resource links: a local pointer to a database shared from another account/region via AWS RAM, exclusive with federation by CEL (a database is exactly one shape).
- **`federated_database`** — Redshift-datashare projection through a Glue connection.
- **`catalog_id`** — cross-account catalog placement (ForceNew, documented).

### Both kinds to the settled engineering bar

- Generator-owned `variables.tf` contracts, enrolled in the drift guard; Terraform floors lifted from `>= 5.0` to `>= 6.34.0` (Athena — S3 Access Grants lands there) and `>= 6.0.0` (Glue).
- Identity tags converged: both TF modules were on stale `planton.dev/*` keys while Pulumi emitted `planton.ai/*` — two engines produced visibly different consoles; now identical key-for-key.
- Missing anatomy created: `iac/hack/manifest.yaml` (the offline `tofu plan` proof input — neither kind had one, so that gate had silently never run for them) and `iac/pulumi/iac-input.yaml`.
- All three pre-existing Athena presets were missing the required `region` — invalid since authoring; fixed, plus new presets for managed results (Athena) and shared-database links (Glue).
- Richly-commented modules in both engines; ApplyT-free chained-accessor export for `effective_engine_version`.

## E2E

- Two new verifiers: Athena `GetWorkGroup` (missing workgroups surface as `InvalidRequestException` "... is not found" — the provider finder's own signal) and Glue `GetDatabase` (typed `EntityNotFoundException`); athena + glue SDK modules added.
- Three scenarios: Glue full-surface (parameters + Lake Formation grant + fixture-bucket location), Athena S3-results full-surface (composed against the shared S3 bucket fixture via the e2e-prerequisites annotation, with managed + S3 logging arms), and Athena managed-results (its own lane because AWS forbids it alongside an S3 output location).
- **Live dual-engine E2E 6/6 green** (2026-07-09): Glue full-surface Pulumi 6m03s (first lane, plugin warm-up; deploy 16s) / Terraform 58s; Athena managed-results 24s/24s; Athena S3-results ~55s/60s incl. the fixture chain. Zero-orphan sweep clean (no workgroups beyond `primary`, no databases, no fixture buckets).
- Live-lane exclusions recorded in the profiles with reasons: Identity Center + Access Grants (need an Identity Center instance), Spark-only arms, Glue resource links + federation (need externally shared infrastructure) — all proven by spec tests and the offline plan gate against the full-surface hack manifests.

## Validation

- Full offline gate: buf lint + stub regen, spec/CEL suites for both kinds (state dial, managed-results exclusivity, access-grants contract, log-type shape, permission set, resource-link requiredness, shape exclusivity), targeted Go builds + both Pulumi entrypoints + the Bazel repo build, foreign-key and secret-coverage guards, drift guard + outputs conformance (two new cases), `tofu init`/`validate` + full-surface `plan` proofs for both modules, all presets/scenarios/hack manifests CLI-validated, site catalog regenerated.
