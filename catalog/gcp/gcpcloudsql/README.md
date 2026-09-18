# GCP Cloud SQL

Deploys a Cloud SQL instance (`google_sql_database_instance`) — a fully managed MySQL, PostgreSQL, or SQL Server database server — with the full production surface: edition and HA, disk tuning, explicit connectivity (public IPv4 / private VPC IP / Private Service Connect), backups with point-in-time recovery, maintenance scheduling, Query Insights, password policies, managed connection pooling, CMEK, delete guards, and read replicas.

## What Gets Created

When you deploy a GcpCloudSql resource, Planton provisions:

- **Cloud SQL Admin API enablement** — a `google_project_service` resource that activates `sqladmin.googleapis.com` on the target project
- **Cloud SQL instance** — a `google_sql_database_instance`: a primary, or (with `masterInstanceName`) a read replica of another instance

Databases and users inside the instance are separate composable resources — [GcpCloudSqlDatabase](/docs/catalog/gcp/gcpcloudsqldatabase) and [GcpCloudSqlUser](/docs/catalog/gcp/gcpcloudsqluser) — referencing this instance by its `instance_name` output.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId` (or the provider's default project)
- **IAM permissions** — `roles/cloudsql.admin` on the target project
- **For private IP** — the VPC must already carry private services access: a [GcpGlobalAddress](/docs/catalog/gcp/gcpglobaladdress) with `purpose: VPC_PEERING` plus a [GcpServiceNetworkingConnection](/docs/catalog/gcp/gcpservicenetworkingconnection). GCP rejects instance creation otherwise.
- **For CMEK** — a [GcpKmsKey](/docs/catalog/gcp/gcpkmskey) in the same region as the instance

## Quick Start

Create a file `postgres.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSql
metadata:
  name: my-postgres
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: orders-db
  region: us-central1
  databaseEngine: POSTGRESQL
  databaseVersion: POSTGRES_16
  tier: db-custom-2-7680
  backup:
    enabled: true
    pointInTimeRecoveryEnabled: true
```

Deploy:

```shell
planton apply -f postgres.yaml
```

This creates a PostgreSQL 16 instance with a public IPv4 address and **no authorized networks** — reachable only through the IAM-authenticated Cloud SQL Auth Proxy or connectors, which is a safe default.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `instanceName` | `string` | Instance name in GCP. Immutable; reserved ~1 week post-delete. | RFC 1035, max 98 chars |
| `region` | `string` | GCP region (e.g. `us-central1`). Immutable. | Valid region name |
| `databaseEngine` | `string` | `MYSQL`, `POSTGRESQL`, or `SQLSERVER`. | Drives engine-specific validation |
| `databaseVersion` | `string` | Exact version, e.g. `POSTGRES_16`. Mutable (in-place upgrade). | Prefix must match the engine |
| `tier` | `string` | Machine type, e.g. `db-custom-4-15360`. Mutable (in-place resize). | Non-empty |

### Optional Fields (grouped)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | GCP project. Can reference a GcpProject. |
| `edition` | `string` | `ENTERPRISE` | `ENTERPRISE` or `ENTERPRISE_PLUS` (data cache, 99.99% HA SLA, 35-day PITR logs). |
| `availabilityType` | `string` | `ZONAL` | `REGIONAL` enables HA with automatic failover (requires backups). |
| `activationPolicy` | `string` | `ALWAYS` | `NEVER` stops the instance while retaining storage. |
| `disk` | object | 10 GB PD_SSD, auto-resize | `type`, `sizeGb`, `autoResize`, `autoResizeLimit`. Disks grow, never shrink. |
| `network` | object | public IPv4, no allowlist | `privateNetwork` (ref → GcpVpcNetwork), `ipv4Enabled`, `authorizedNetworks`, `allocatedIpRange`, `enablePrivatePathForGoogleCloudServices`, `sslMode`, `serverCaMode`/`serverCaPool`/`customSubjectAlternativeNames`, `psc`. |
| `locationPreference` | object | GCP picks | `zone`, `secondaryZone` (REGIONAL only). |
| `backup` | object | disabled | `enabled`, `startTime`, `location`, `binaryLogEnabled` (MySQL PITR), `pointInTimeRecoveryEnabled` (PG/SQL Server PITR), `transactionLogRetentionDays` (1–35), `retainedBackups`. |
| `maintenanceWindow` | object | any time | `day` (1=Mon..7=Sun), `hour` (UTC), `updateTrack` (`canary`/`stable`/`week5`). |
| `denyMaintenancePeriod` | object | — | `startDate`, `endDate`, `time` — a maintenance freeze (max 90 days). |
| `insightsConfig` | object | disabled | Query Insights: `queryInsightsEnabled`, `queryStringLength`, `recordApplicationTags`, `recordClientAddress`, `queryPlansPerMinute`. |
| `passwordValidationPolicy` | object | — | Instance-level password rules for built-in users. |
| `dataCacheEnabled` | `bool` | `false` | Local-SSD read cache. ENTERPRISE_PLUS only. |
| `connectionPooling` | object | disabled | Managed connection pooling (`enabled`, `flags`). |
| `databaseFlags` | `map` | — | Engine flags, e.g. `cloudsql.iam_authentication: "on"`. |
| `threadsPerCore`, `timeZone`, `collation`, `sqlServerAuditConfig`, `activeDirectoryDomain` | — | — | SQL Server-only surface (validated per engine). |
| `connectorEnforcement` | `string` | `NOT_REQUIRED` | `REQUIRED` rejects all direct connections (connectors only). |
| `enableGoogleMlIntegration`, `enableDataplexIntegration` | `bool` | `false` | Vertex AI / Dataplex integrations. |
| `encryptionKeyName` | `StringValueOrRef` | Google-managed | CMEK crypto key (ref → GcpKmsKey `key_id`). Same region; immutable. |
| `deletionProtection` | `bool` | `false` | Engine-side destroy guard (plan-time refusal). |
| `deletionProtectionEnabled` | `bool` | `false` | API-side delete guard (blocks console/gcloud/API deletion too). |
| `retainBackupsOnDelete` | `bool` | `false` | Keep backups (and PITR logs) after instance deletion. |
| `masterInstanceName` | `StringValueOrRef` | — | Makes this instance a read replica of the named primary (ref → GcpCloudSql `instance_name`). Immutable. |
| `replicaConfiguration` | object | — | Replica behavior + external-source replication channel (password and client key are secret). |
| `rootPassword` | `string` (secret) | — | Initial admin password. Required for SQL Server. Write-only in GCP. |

### Validation Rules (enforced pre-deploy)

- `databaseVersion` prefix must match `databaseEngine` (`MYSQL_*` / `POSTGRES_*` / `SQLSERVER_*`).
- SQL Server requires `rootPassword`.
- PITR is PostgreSQL/SQL Server only; MySQL uses `binaryLogEnabled` (MySQL only) instead.
- `REGIONAL` availability requires backups enabled — and binary logs on MySQL.
- `dataCacheEnabled` requires `edition: ENTERPRISE_PLUS`.
- `timeZone`, `collation`, `threadsPerCore`, `sqlServerAuditConfig`, `activeDirectoryDomain` are SQL Server only; `passwordValidationPolicy.passwordChangeInterval` is PostgreSQL only.
- A present `network` block must enable at least one path: `ipv4Enabled`, `privateNetwork`, or `psc.enabled`; `authorizedNetworks` require `ipv4Enabled`; `allocatedIpRange` and private-path require `privateNetwork`; `serverCaPool` pairs exactly with `serverCaMode: CUSTOMER_MANAGED_CAS_CA`.
- `replicaConfiguration` requires `masterInstanceName`; `secondaryZone` requires `REGIONAL`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `instance_name` | `string` | The composition key — databases, users, and replicas reference it |
| `connection_name` | `string` | `project:region:instance` for the Auth Proxy and connectors |
| `private_ip` | `string` | Private IP (empty unless `privateNetwork` is set) |
| `public_ip` | `string` | Public IPv4 (empty unless `ipv4Enabled`) |
| `self_link` | `string` | GCP resource self link |
| `service_account_email` | `string` | The instance's Google-managed service account — grant it GCS access for imports/exports and audit uploads |
| `dns_name` | `string` | DNS name (PSC-enabled instances) |
| `psc_service_attachment_link` | `string` | PSC service attachment consumers target with endpoints |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **Name reservation**: a deleted instance's name stays reserved for about one week and cannot be reused.
- **In-place changes**: `databaseVersion` upgrades and `tier`/`edition` changes apply in place (with a restart) — take a backup before major version upgrades.
- **Private network is one-way**: it can be set or changed in place, but never removed — removing it forces instance replacement.
- **Read replicas**: a replica is its own `GcpCloudSql` node with `masterInstanceName`; promote it to a standalone primary by setting `instanceType: CLOUD_SQL_INSTANCE` and clearing `masterInstanceName`/`replicaConfiguration` in the same change (the instance restarts).
- **Two delete guards**: `deletionProtection` stops IaC destroys; `deletionProtectionEnabled` makes GCP itself reject deletion from any surface. Set both in production.

### Deliberately not modeled (recorded reasons)

Everything else on `google_sql_database_instance` at the pinned provider is representable — including clones, backup-run and Backup-and-DR restores, read pools with auto scaling, DR replica pairing, final backups, Entra ID / customer-managed Active Directory, hyperdisk performance dials, the replica self-recreation threshold (`replicationLagMaxSeconds`), the input-only major-version and transaction-log opt-ins, the new-network-architecture switch, the PSC auto-connection policy, and `deletionPolicy`. The recorded exclusions:

| Excluded Feature | Why |
|---|---|
| `root_password_wo` / `root_password_wo_version` | Write-only variants of the modeled `rootPassword` — same capability through engine-side ergonomics; the spec field is secret-annotated and encrypted in state on both engines. |
| `pricing_plan` | `PER_USE` is the only accepted value on second-generation instances — no reachable capability. |
| `follow_gae_application` | Legacy App Engine zone-following; `locationPreference.zone` is the direct modern placement control. |
| `replication_cluster.psa_write_endpoint` | Documented read-only field; DR pairing is driven through `failoverDrReplicaName`. |
| `user_labels` | Driven by the platform metadata labels on both engines, not spec surface. |

## Related Components

- [GcpCloudSqlDatabase](/docs/catalog/gcp/gcpcloudsqldatabase) — logical databases inside this instance
- [GcpCloudSqlUser](/docs/catalog/gcp/gcpcloudsqluser) — per-application users
- [GcpServiceNetworkingConnection](/docs/catalog/gcp/gcpservicenetworkingconnection) — the private services access peering private IP depends on
- [GcpGlobalAddress](/docs/catalog/gcp/gcpglobaladdress) — the reserved peering range
- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the network for private IP
- [GcpKmsKey](/docs/catalog/gcp/gcpkmskey) — CMEK encryption

## Additional Resources

- [Cloud SQL documentation](https://cloud.google.com/sql/docs)
- [About the Cloud SQL Auth Proxy](https://cloud.google.com/sql/docs/postgres/sql-proxy)
- [Instance settings reference](https://cloud.google.com/sql/docs/postgres/instance-settings)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
