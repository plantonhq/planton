# GCP Firestore Database

Deploys a Firestore database in a GCP project with configurable database type (Native or Datastore mode), location (single-region or multi-region), concurrency mode, point-in-time recovery, delete protection, edition tier (Standard or Enterprise), CMEK encryption, ENTERPRISE data-access modes (classic Firestore API, MongoDB-compatible API, realtime updates), resource-manager tags, and teardown policy. A project holds one `(default)` database plus any number of named databases; this kind models one database per Infra Component.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Firestore Database** -- a named database resource in the specified GCP project and location, configured with the chosen database type, concurrency mode, and edition
- **Point-in-Time Recovery** -- created only when `pointInTimeRecoveryEnablement` is set to `POINT_IN_TIME_RECOVERY_ENABLED`; retains 7 days of version history for timestamp-based reads and recovery
- **Delete Protection** -- enabled when `deleteProtectionState` is set to `DELETE_PROTECTION_ENABLED` (GCP's default leaves the database deletable); prevents accidental database deletion through any interface until explicitly disabled
- **CMEK Encryption** -- created only when `kmsKeyName` is provided; encrypts database data at rest using a customer-managed Cloud KMS key in the same location as the database
- **Data-Access Modes** -- on ENTERPRISE databases, three switches control which protocols can read and write: the classic Firestore API (`firestoreDataAccessMode`), MongoDB drivers (`mongodbCompatibleDataAccessMode`), and realtime query snapshots (`realtimeUpdatesMode`)
- **Teardown Policy** -- `deletionPolicy` defaults to DELETE so destroys manage the full lifecycle; PREVENT makes a destroy fail, ABANDON unmanages the database while it keeps serving
- **Firestore API enablement** -- `firestore.googleapis.com` enabled in the target project (never disabled on destroy)

Firestore databases do not support GCP labels. Resource Manager tags (`resourceManagerTags`) are applied at create time only; day-to-day resource tracking relies on the Planton metadata (organization, environment, resource kind) stored in the Infra Component record.

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the Firestore database will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **Cloud KMS key** in the same location as the database (only for CMEK). For multi-region databases: `nam5` requires a key in the `us` multi-region; `eur3` requires a key in the `europe` multi-region.

## Deploy

### Console

Open the deployment store, find **GCP Firestore Database**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Default Firestore Native Database** preset in the [Presets](#presets) tab to create the project's default Firestore Native database.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirestoreDatabase
metadata:
  name: app-database
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  locationId: nam5
  databaseName: "(default)"
  type: FIRESTORE_NATIVE
```

```shell
planton apply -f firestore-database.yaml
```

This creates the project's default Firestore Native database in the US multi-region with optimistic concurrency, no PITR, delete protection disabled, and Google-managed encryption. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the database to a GCP project and KMS key deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: firestore-encryption-key
      fieldPath: status.outputs.key_id
```

The InfraPipeline resolves the dependency graph, deploys the project and KMS key first, then provisions the Firestore database with CMEK encryption.

## Key Configuration

These are the most important decisions when configuring a Firestore database. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Database type** -- Choose `FIRESTORE_NATIVE` for the modern Firestore API with real-time listeners and offline support, or `DATASTORE_MODE` for the legacy Datastore API with entity-group transactions. The type can be changed after creation, but this is a significant operational change affecting client library compatibility.

**Location** -- Set `locationId` to a multi-region (`nam5` for US, `eur3` for Europe) for higher availability, or a single region (e.g., `us-east1`) for lower latency and cost. Immutable after creation.

**Database name** -- Use `"(default)"` for the project's primary database that client libraries connect to by default. Use a custom name (4-63 characters) for additional databases when isolating workloads. Immutable after creation.

**Point-in-time recovery** -- Set `pointInTimeRecoveryEnablement` to `POINT_IN_TIME_RECOVERY_ENABLED` to retain 7 days of version history. Enables reads at any timestamp within the past hour or any 1-minute snapshot within 7 days. Essential for production data protection.

**Edition and encryption** -- Set `databaseEdition` to `ENTERPRISE` for enhanced SLA and advanced security (requires `FIRESTORE_NATIVE` type). Add `kmsKeyName` for CMEK encryption. Both are immutable after creation.

**App Engine integration** -- Leave `appEngineIntegrationMode` unset (or `DISABLED`) unless the database serves a legacy App Engine application: `ENABLED` couples the database's lifecycle to the project's App Engine app, so disabling the app disables the database with it. Immutable after creation.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpKmsKey** (optional) | `kmsKeyName` | `status.outputs.key_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `database_id` | Fully qualified database ID (`projects/{project}/databases/{database}`) | Application configuration, IAM bindings |
| `database_name` | Database name (`(default)` or custom name) | Client library database selection, `GcpFirestoreIndex`/`GcpFirestoreBackupSchedule` attachment |
| `earliest_version_time` | Earliest timestamp for version reads (RFC3339 UTC) | Point-in-time recovery planning |
| `version_retention_period` | How long past versions are retained (`3600s` without PITR, `604800s` with PITR) | Recovery-window verification |
| `key_prefix` | Datastore Mode app-identifier prefix (empty without a legacy App Engine app) | Legacy Datastore key construction |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Default Native** -- The project's default Firestore Native database in the US multi-region with delete protection enabled. The starting point for most applications using the Firestore client libraries. Start from the **Default Firestore Native Database** preset.

**Named Native with PITR** -- A custom-named Firestore Native database with point-in-time recovery enabled and delete protection. Suitable for production workloads that need data recovery capabilities and workload isolation from the default database. Start from the **Named Firestore Native Database with PITR** preset.

**Enterprise with CMEK** -- Enterprise edition Firestore Native database with PITR, delete protection, and customer-managed encryption. Suitable for regulated industries requiring enhanced SLA, advanced security, and encryption key control. Start from the **Enterprise Firestore Database with CMEK** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the database is created
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides the CMEK encryption key for data at rest
- [**GCP Firestore Index**](/infra-catalog/gcp-firestore-index) -- composite and vector indexes attached to this database
- [**GCP Firestore Backup Schedule**](/infra-catalog/gcp-firestore-backup-schedule) -- daily and weekly managed backups of this database