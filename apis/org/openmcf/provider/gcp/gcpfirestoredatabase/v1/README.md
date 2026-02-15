# GcpFirestoreDatabase

Provisions a [Google Cloud Firestore](https://cloud.google.com/firestore) database with configurable database type, edition, PITR, CMEK encryption, and delete protection.

## What It Does

A Firestore database is the top-level container for collections, documents, and indexes in Google Cloud Firestore. This component creates and manages the database itself, including its type (Native or Datastore mode), edition, encryption configuration, and point-in-time recovery settings.

Each GCP project can have multiple named databases plus one special `(default)` database. The default database is what client libraries connect to when no database ID is specified.

## When to Use

- You need a managed NoSQL document database for your application
- You want to manage the database lifecycle (creation, encryption, PITR) through OpenMCF
- You need to enforce CMEK encryption or delete protection for compliance
- You want a named database separate from the project's default database

## Key Configuration

### Database Type (`type`)

Choose the database type at creation time:

| Type | When to Use |
|---|---|
| `FIRESTORE_NATIVE` | Modern Firestore with real-time listeners, offline support, and mobile SDKs. Recommended for new applications. |
| `DATASTORE_MODE` | Legacy Datastore-compatible mode. Use for existing Datastore applications or server-side workloads. |

### Database Name (`database_name`)

- Use `(default)` for the project's primary database (one per project)
- Use a custom name (4-63 chars, lowercase letters/digits/hyphens) for additional databases

### Database Edition (`database_edition`)

| Edition | When to Use |
|---|---|
| `STANDARD` | Default. Suitable for most workloads. |
| `ENTERPRISE` | Enhanced SLA, advanced security. Requires `FIRESTORE_NATIVE` type. |

### Point-in-Time Recovery (`point_in_time_recovery_enablement`)

When enabled, retains 7 days of version history for disaster recovery. When disabled (default), retains 1 hour.

### Labels

Firestore databases do **not** support GCP labels.

## Outputs

| Output | Description |
|---|---|
| `database_id` | Fully qualified path (`projects/{project}/databases/{name}`) |
| `database_name` | Database name (e.g., `(default)` or custom name) |
| `uid` | Server-generated UUID4 |
| `create_time` | Creation timestamp (RFC3339) |
| `earliest_version_time` | Earliest PITR recovery timestamp (RFC3339) |

## Relationships

- **Depends on**: GcpProject (project_id), optionally GcpKmsKey (kms_key_name)
- **Referenced by**: Application connection strings, Firebase client libraries

## Deployment

```shell
openmcf apply -f firestore-database.yaml
```

For copy-paste ready manifests, see [examples.md](examples.md).
