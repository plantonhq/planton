# GCP Cloud SQL Depth Rebuild + Database & User Kinds

**Date**: July 4, 2026
**Type**: Feature
**Components**: API Definitions, GCP Provider, IAC Modules, E2E Framework, InfraCharts

## Summary

Deep-rebuilt `GcpCloudSql` (604) to the released-provider floor with full dual-engine parity, forged `GcpCloudSqlDatabase` (637) and `GcpCloudSqlUser` (638) as first-class composable kinds, extended the GCP E2E harness with a `sqladmin` client and three verifiers, and reworked both consuming Infra Charts onto the new spec shape — including fixing the `cloud-run-environment` chart's missing private-services-access chain.

## Problem Statement / Motivation

`GcpCloudSql` carried the sharpest verified correctness defect on the baseline ledger: spec fields silently dropped by **both** engines (`edition`, `query_insights_enabled`, `maintenance_window`, `high_availability.zone`, `backup.point_in_time_recovery_enabled`), stale TF ref typing, `root_password` not secret in Pulumi state, hardcoded `6.19.0` pin, no API enablement, and legacy hack-manifest location. Databases and users lived inside the instance module as hidden resources with no independent lifecycle.

The data wave's first consumer of the session-011 PSA pair needed a deep instance rebuild plus first-class database and user nodes to prove real end-to-end composition.

## Solution / What's New

### `GcpCloudSql` deep-rebuild (604)

- Full released-floor `settings` surface: engine enum incl. SQL Server, disk, `ip_configuration` (private network ref → `GcpVpc.network_id`, SSL, PSC, authorized networks), backup/PITR, maintenance + deny window, insights, password policy, data cache, flags, CMEK ref, replica arm, dual deletion protection.
- `root_password` marked `(sensitive)` with Pulumi `ToSecret`; TF on `google ~> 6.0`; `sqladmin.googleapis.com` enablement; converter-contract plain-string refs; ambient-project fallback.
- Registry `prerequisites: [GcpServiceNetworkingConnection]` for private-IP composition.
- Three rewritten presets (Postgres private-IP production, MySQL HA, Postgres read replica); extended outputs (`service_account_email`, `dns_name`, `psc_service_attachment_link`).

### `GcpCloudSqlDatabase` (637, `gcpsqldb`)

Logical database inside an instance; instance ref → `GcpCloudSql.instance_name`; charset/collation engine semantics documented; no redundant API enablement.

### `GcpCloudSqlUser` (638, `gcpsqlusr`)

Database user with `(sensitive)` password + ToSecret; BUILT_IN and CLOUD_IAM_* types with CEL coherence; `password_policy` block; outputs include `user_name` + `instance_name`.

### E2E harness

- `sqladmin/v1` client in `aa_e2e`; verifiers for instance (connectivity posture from outputs), database (self-link instance parse), user.
- Scenarios: public-IP Postgres leaf, private-IP PSA flagship, database on instance prerequisite, user on instance prerequisite.
- Consumer-scoped prerequisite overrides for the full PSA chain with `${E2E_RUN_ID}` in cloud-side VPC/address names (fixed metadata for FK resolution).
- Dependency destroy retry (6×60s) and SNC peering poll (2 min) for asynchronous producer cleanup.

### Chart rework

- **`serverless-api-backend`**: new spec shape (`disk.sizeGb`, `network.privateNetwork` → `network_id`, private-IP only); added `GcpCloudSqlDatabase` + `GcpCloudSqlUser` nodes.
- **`cloud-run-environment`**: added explicit PSA pair (`psa.yaml`); fixed database template; added app database + user nodes; fixed storage bucket missing `bucketName`.

## Validation

Offline green: `make protos` + gazelle; spec tests (all three kinds); `secret-coverage --check`; `validate-refs --check`; `pkg/outputs` conformance (3 cases); `tofu validate` + offline `planton tofu plan` ×3; hack manifests + presets validated with locally built `planton`; release-equivalent Pulumi builds ×3; `make build-go`.

Parity audits: `GcpCloudSql` **100% Fully Complete — PARITY ✅**; database and user **PARITY ✅** with minor docs-band gaps closed in-session.

Live E2E (project `planton-e2e`):

| Scenario | Pulumi | Terraform |
|---|---|---|
| `gcpcloudsql/private-postgres` (PSA chain → private-IP instance) | ✅ PASS (23m) | ✅ PASS (33m) |
| `gcpcloudsql/public-postgres` | ✅ PASS (25m) | ✅ PASS (30m) |
| `gcpcloudsqldatabase/minimal` (full PSA chain → instance → database) | ✅ PASS (16m) | ✅ PASS (13m) |
| `gcpcloudsqluser/minimal` (full PSA chain → instance → user) | ✅ PASS (16m) | ✅ PASS (17m) |

8/8 scenario-runs green across both engines, zero orphans after final teardown. The database/user chains exercise all three kinds composed through the full private services access chain in one scenario each.

## Learn-once uplifts

- E2E README: Cloud SQL chain consumer-scoped prerequisites with run-scoped VPC/address names; dependency lifecycle robustness (destroy retry, peering poll, `pulumi up --refresh`).
- `load-tfvars` stdout fix (`fmt.Println` not `println`) so `> file.tfvars` captures rendered content.
