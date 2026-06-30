# AWS Glue Catalog Database Deployment Component

**Date**: February 16, 2026
**Type**: Feature
**Components**: API Definitions, AWS Provider, Pulumi CLI Integration

## Summary

Added the AwsGlueCatalogDatabase deployment component — a metadata namespace for organizing table definitions in AWS's serverless data lake architecture. This is the twenty-third new AWS resource kind in the cloud provider expansion project, providing the foundational catalog layer that Athena, Glue Crawlers, Glue ETL, and Redshift Spectrum depend on.

## Problem Statement / Motivation

Data engineers building analytics pipelines on AWS need a Data Catalog database to organize table definitions before they can run Athena queries, configure Glue Crawlers, or set up ETL jobs. Without an Planton component for Glue Catalog Database, this foundational resource had to be created manually or outside the declarative infrastructure workflow.

### Pain Points

- No declarative way to provision Glue Data Catalog databases alongside other analytics resources
- Missing link in the data lake infrastructure chain: S3 → **Glue Database** → Athena Workgroup
- Inability to compose data analytics stacks in infra charts without the catalog database component

## Solution / What's New

### AwsGlueCatalogDatabase Component

A lean, focused component with 2 spec fields that accurately represents the simplicity of a Glue Catalog Database:

- `description` — Human-readable documentation of the database's purpose
- `locationUri` — Default S3 storage location for tables in this database

The design deliberately keeps the spec minimal because a Glue Database is a metadata container — the complexity belongs in downstream resources (Glue Tables, Crawlers, ETL Jobs), not the namespace itself.

## Implementation Details

### Proto API

Four proto files following established Planton patterns:
- `spec.proto` — 2 optional fields, no nested messages, no CEL validations
- `api.proto` — Standard KRM envelope (apiVersion, kind, metadata, spec, status)
- `stack_input.proto` — Target resource + AWS provider config
- `stack_outputs.proto` — 3 outputs (database_name, database_arn, catalog_id)

### IaC Modules

**Pulumi (Go):** 4 files — `main.go`, `locals.go`, `outputs.go`, `catalog_database.go`. Creates a single `glue.NewCatalogDatabase` resource with conditional description and location URI.

**Terraform (HCL):** 5 files — `main.tf`, `variables.tf`, `outputs.tf`, `locals.tf`, `provider.tf`. Feature parity with Pulumi module.

### Validation Tests

12 tests covering:
- 7 happy path scenarios (empty spec, description only, location only, both, long description, no trailing slash, production-ready)
- 5 API envelope tests (wrong apiVersion, wrong kind, missing metadata, missing spec, valid complete)

### Documentation and Presets

- README.md with spec reference and omitted features table
- examples.md with 6 progressive examples (minimal → multi-environment → infra chart)
- docs/README.md with architecture (catalog hierarchy, cost model, security, patterns)
- 2 presets: basic-data-catalog, s3-data-lake
- Catalog page audited to zero Critical issues

### Design Decision: Lean v1

Four features were discovered in the Terraform provider that were not mentioned in the planning phase. All were deferred to v2:

| Feature | Reason |
|---------|--------|
| `create_table_default_permission` | Lake Formation governance; default IAM behavior covers >80% |
| `federated_database` | Redshift Data Share, ~5% adoption, requires Lake Formation |
| `target_database` | Cross-region/cross-account, ~5% adoption |
| `parameters` | Generic metadata map, rarely set by users |

## Benefits

- Completes the Athena analytics chain: S3 → Glue Database → Athena Workgroup
- Enables data lake infra charts that compose storage, catalog, and query resources
- Exports database_name, database_arn, and catalog_id for downstream resource wiring
- Proportionally simple spec for a simple resource — no unnecessary complexity

## Impact

- **Data Engineers**: Can now declaratively manage Glue Catalog Databases in their infrastructure-as-code workflow
- **Infra Chart Authors**: Can compose data analytics stacks with the full S3 → Glue → Athena pipeline
- **AWS Coverage**: 23 of ~32 new AWS resources complete (72%)

## Related Work

- Previous: AwsAthenaWorkgroup (R18) — queries data described by tables in Glue databases
- Next: AwsRedshiftCluster (R20) — Redshift Spectrum queries the Glue catalog for external tables

---

**Status**: Production Ready
