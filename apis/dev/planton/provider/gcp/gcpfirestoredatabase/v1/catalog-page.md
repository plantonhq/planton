# GCP Firestore Database

Deploys a Google Cloud Firestore database with configurable type (Native or Datastore mode), edition, point-in-time recovery, CMEK encryption, and delete protection. Supports both the project's default database and additional named databases.

## What Gets Created

- **Firestore Database** -- a `google_firestore_database` resource in the specified location with the chosen type and edition
- **CMEK Encryption** -- configured only when `kmsKeyName` is provided, encrypts the database with a customer-managed KMS key

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A GCP project** with the Firestore API enabled
- **A KMS key** in the same location as the database if enabling CMEK encryption (nam5 requires KMS multi-region `us`; eur3 requires KMS multi-region `europe`)

## Quick Start

Create a file `firestore-database.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFirestoreDatabase
metadata:
  name: my-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpFirestoreDatabase.my-db
spec:
  projectId:
    value: my-gcp-project-123
  locationId: nam5
  databaseName: "(default)"
  type: FIRESTORE_NATIVE
```

Deploy:

```shell
planton apply -f firestore-database.yaml
```

This creates the project's default Firestore Native database in the US multi-region with Google-managed encryption and 1-hour version retention.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project ID. Can reference a GcpProject resource via `valueFrom`. | Required |
| `locationId` | `string` | Database location. Multi-region (`nam5`, `eur3`) or single-region (`us-east1`, `europe-west1`). Immutable after creation. | Required |
| `databaseName` | `string` | Database name. Use `(default)` for the primary database or a custom name. Immutable after creation. | `(default)` or 4-63 chars: `^[a-z][a-z0-9-]*[a-z0-9]$` |
| `type` | `string` | Database type. Determines the data model and API surface. | `FIRESTORE_NATIVE` or `DATASTORE_MODE` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `concurrencyMode` | `string` | Per type | Concurrency control. `OPTIMISTIC` (default for Native), `PESSIMISTIC` (default for Datastore), `OPTIMISTIC_WITH_ENTITY_GROUPS` (legacy Datastore only). |
| `pointInTimeRecoveryEnablement` | `string` | `POINT_IN_TIME_RECOVERY_DISABLED` | `POINT_IN_TIME_RECOVERY_ENABLED` retains 7 days of version history for disaster recovery. |
| `deleteProtectionState` | `string` | `DELETE_PROTECTION_DISABLED` | `DELETE_PROTECTION_ENABLED` prevents deletion through any interface until disabled. |
| `databaseEdition` | `string` | `STANDARD` | `STANDARD` or `ENTERPRISE`. ENTERPRISE provides enhanced SLA and advanced features. Requires `type: FIRESTORE_NATIVE`. Immutable. |
| `kmsKeyName` | `StringValueOrRef` | -- | Fully qualified KMS key name for CMEK encryption. Must be in the same location as the database. Immutable. Can reference a GcpKmsKey resource via `valueFrom`. |

## Examples

### Default Firestore Native Database

The simplest starting point -- creates the project's default database that client libraries connect to by default:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFirestoreDatabase
metadata:
  name: default-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpFirestoreDatabase.default-db
spec:
  projectId:
    value: my-gcp-project-123
  locationId: nam5
  databaseName: "(default)"
  type: FIRESTORE_NATIVE
```

### Named Database with PITR and Delete Protection

A production database with 7-day point-in-time recovery and protection against accidental deletion:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFirestoreDatabase
metadata:
  name: orders-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpFirestoreDatabase.orders-db
spec:
  projectId:
    value: my-gcp-project-123
  locationId: us-east1
  databaseName: orders-db
  type: FIRESTORE_NATIVE
  pointInTimeRecoveryEnablement: POINT_IN_TIME_RECOVERY_ENABLED
  deleteProtectionState: DELETE_PROTECTION_ENABLED
```

### Enterprise Edition with CMEK Encryption

Maximum security configuration with Enterprise SLA, customer-managed encryption, and delete protection:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFirestoreDatabase
metadata:
  name: secure-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpFirestoreDatabase.secure-db
spec:
  projectId:
    value: my-gcp-project-123
  locationId: nam5
  databaseName: secure-db
  type: FIRESTORE_NATIVE
  databaseEdition: ENTERPRISE
  pointInTimeRecoveryEnablement: POINT_IN_TIME_RECOVERY_ENABLED
  deleteProtectionState: DELETE_PROTECTION_ENABLED
  kmsKeyName:
    value: projects/my-gcp-project-123/locations/us/keyRings/firestore-ring/cryptoKeys/firestore-key
```

### Using Foreign Key References

Reference other Planton-managed resources instead of hardcoding values:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFirestoreDatabase
metadata:
  name: composed-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpFirestoreDatabase.composed-db
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  locationId: nam5
  databaseName: composed-db
  type: FIRESTORE_NATIVE
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: firestore-key
      field: status.outputs.key_id
  pointInTimeRecoveryEnablement: POINT_IN_TIME_RECOVERY_ENABLED
  deleteProtectionState: DELETE_PROTECTION_ENABLED
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `database_id` | `string` | Fully qualified database path (`projects/{project}/databases/{database}`) |
| `database_name` | `string` | Database name as specified in `databaseName` (e.g., `(default)` or custom name) |
| `uid` | `string` | Server-generated UUID4, unique across all Firestore databases |
| `create_time` | `string` | Timestamp when the database was created (RFC3339 UTC) |
| `earliest_version_time` | `string` | Earliest timestamp for point-in-time recovery reads (RFC3339 UTC). 1 hour without PITR, 7 days with PITR enabled. |

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) -- provides the GCP project that hosts this database
- [GcpKmsKey](/docs/catalog/gcp/gcpkmskey) -- provides the encryption key for CMEK
- [GcpKmsKeyRing](/docs/catalog/gcp/gcpkmskeyring) -- provides the key ring containing the encryption key
