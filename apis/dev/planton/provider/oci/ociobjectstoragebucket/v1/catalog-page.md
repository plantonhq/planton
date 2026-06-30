# OCI Object Storage Bucket

Deploys an Oracle Cloud Infrastructure Object Storage bucket with optional retention rules, lifecycle policies for automatic object transitions and deletions, and cross-region replication. Versioning, auto-tiering, customer-managed encryption, and event emission are configurable at the bucket level.

## What Gets Created

When you deploy an OciObjectStorageBucket resource, Planton provisions:

- **Object Storage Bucket** — an `oci_objectstorage_bucket` resource in the specified compartment and namespace with configurable access type, storage tier, versioning, auto-tiering, and optional KMS encryption. Retention rules are managed inline on the bucket (max 100).
- **Lifecycle Policy** — created only when `lifecycleRules` is non-empty. A single `oci_objectstorage_object_lifecycle_policy` resource containing all lifecycle rules. Rules automate object archival, tiering transitions, deletion, and multipart upload cleanup based on age and name patterns.
- **Replication Policies** — one `oci_objectstorage_replication_policy` per entry in `replicationPolicies`. Each policy asynchronously copies objects to a destination bucket in another OCI region for disaster recovery. All replication policy fields are immutable after creation.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the bucket will be created — either a literal value or a reference to an OciCompartment resource
- **Object Storage namespace** — the tenancy-unique namespace string (retrieve via `oci os ns get` or from the OCI Console)
- **Destination buckets** (for replication) — must already exist in the target region before creating replication policies

## Quick Start

Create a file `bucket.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciObjectStorageBucket
metadata:
  name: my-bucket
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciObjectStorageBucket.my-bucket
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  namespace: "axe1234abc"
  name: "my-bucket"
```

Deploy:

```shell
planton apply -f bucket.yaml
```

This creates a private bucket with Standard storage tier, no versioning, and Oracle-managed encryption. The bucket OCID is exported as a stack output.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the bucket will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `namespace` | `string` | Object Storage namespace for the tenancy. A unique identifier assigned to each tenancy (e.g. `"axe1234abc"`). Retrieve via `oci os ns get`. | Min length 1 |
| `name` | `string` | Bucket name. Must be unique within the namespace. Valid characters: letters, numbers, hyphens, underscores, periods. Changing this forces recreation. | Min length 1 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `accessType` | `enum` | `no_public_access` | Public read access on the bucket. Values: `no_public_access`, `object_read`, `object_read_without_list`. |
| `storageTier` | `enum` | `standard` | Storage class for the bucket. Values: `standard`, `archive`. Immutable after creation. |
| `versioning` | `enum` | — | Object version history. On create: `enabled` or `disabled`. On update: `enabled` or `suspended`. |
| `autoTiering` | `enum` | — | Automatic tier transitions based on access patterns. Values: `auto_tiering_disabled`, `infrequent_access`. |
| `objectEventsEnabled` | `bool` | `false` | When `true`, emits events for object state changes via the OCI Events service. |
| `kmsKeyId` | `StringValueOrRef` | — | OCID of a KMS master encryption key for server-side encryption. When unset, Oracle-managed keys are used. |
| `metadata` | `map<string, string>` | — | User-defined metadata as key-value pairs. Keys must be lowercase. Total size limit is 4 KB. |
| `retentionRules` | `RetentionRule[]` | — | Retention rules enforcing minimum retention periods. Max 100 per bucket. See below. |
| `lifecycleRules` | `LifecycleRule[]` | — | Lifecycle rules automating object transitions and deletions based on age. See below. |
| `replicationPolicies` | `ReplicationPolicy[]` | — | Cross-region replication policies. Each replicates objects to a destination bucket in another region. See below. |

### RetentionRule

| Field | Type | Description |
|-------|------|-------------|
| `displayName` | `string` | Name for the retention rule. Must be unique within the bucket. Changing this forces recreation. |
| `duration` | `Duration` | Retention duration. When omitted, the rule applies indefinitely. |
| `timeRuleLocked` | `string` | RFC 3339 datetime after which this rule becomes locked. Once locked, only duration increases are allowed. |

### Duration

| Field | Type | Description |
|-------|------|-------------|
| `timeAmount` | `int64` | Time amount (>= 1). |
| `timeUnit` | `enum` | Unit for `timeAmount`. Values: `days`, `years`. |

### LifecycleRule

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Rule name. Must be unique within the lifecycle policy. |
| `action` | `enum` | Action to perform. Values: `lifecycle_archive`, `lifecycle_infrequent_access`, `lifecycle_delete`, `lifecycle_abort`. |
| `isEnabled` | `bool` | Whether this rule is active. |
| `timeAmount` | `int64` | Age threshold (>= 1). Objects older than this are acted upon. |
| `timeUnit` | `enum` | Unit for `timeAmount`. Values: `days`, `years`. |
| `target` | `string` | Target object type. Values: `"objects"` (default), `"multipart-uploads"`, `"previous-object-versions"`. |
| `objectNameFilter` | `ObjectNameFilter` | Filter by name pattern. Not valid when target is `"multipart-uploads"`. |

### ObjectNameFilter

| Field | Type | Description |
|-------|------|-------------|
| `inclusionPatterns` | `string[]` | Glob patterns to include. Empty list includes all objects. |
| `inclusionPrefixes` | `string[]` | Object name prefixes to include. Prefer `inclusionPatterns`. |
| `exclusionPatterns` | `string[]` | Glob patterns to exclude. Takes precedence over inclusions. |

### ReplicationPolicy

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Policy name. Immutable after creation. |
| `destinationBucketName` | `string` | Name of the destination bucket. Must already exist in the destination region. Immutable after creation. |
| `destinationRegionName` | `string` | OCI region identifier for the destination (e.g. `"us-ashburn-1"`). Immutable after creation. |

## Examples

### Minimal Private Bucket

A bucket with default settings — suitable for development or application data:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciObjectStorageBucket
metadata:
  name: dev-data
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciObjectStorageBucket.dev-data
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  namespace: "axe1234abc"
  name: "dev-data"
```

### Versioned Bucket with Retention

A bucket with versioning enabled and a 90-day retention rule for compliance:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciObjectStorageBucket
metadata:
  name: compliance-store
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.OciObjectStorageBucket.compliance-store
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  namespace: "axe1234abc"
  name: "compliance-store"
  versioning: enabled
  objectEventsEnabled: true
  retentionRules:
    - displayName: "90-day-hold"
      duration:
        timeAmount: 90
        timeUnit: days
```

### Lifecycle and Auto-Tiering

A bucket with auto-tiering for cost optimization and lifecycle rules to archive old data and clean up incomplete multipart uploads:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciObjectStorageBucket
metadata:
  name: data-lake
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciObjectStorageBucket.data-lake
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  namespace: "axe1234abc"
  name: "data-lake"
  autoTiering: infrequent_access
  lifecycleRules:
    - name: "archive-after-180-days"
      action: lifecycle_archive
      isEnabled: true
      timeAmount: 180
      timeUnit: days
      target: "objects"
    - name: "delete-old-versions"
      action: lifecycle_delete
      isEnabled: true
      timeAmount: 365
      timeUnit: days
      target: "previous-object-versions"
    - name: "abort-stale-uploads"
      action: lifecycle_abort
      isEnabled: true
      timeAmount: 7
      timeUnit: days
      target: "multipart-uploads"
```

### Cross-Region Replication with KMS Encryption

A production bucket with customer-managed encryption and cross-region disaster recovery:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciObjectStorageBucket
metadata:
  name: prod-artifacts
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciObjectStorageBucket.prod-artifacts
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  namespace: "axe1234abc"
  name: "prod-artifacts"
  versioning: enabled
  objectEventsEnabled: true
  kmsKeyId:
    value: "ocid1.key.oc1..example"
  replicationPolicies:
    - name: "dr-to-phoenix"
      destinationBucketName: "prod-artifacts-dr"
      destinationRegionName: "us-phoenix-1"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `bucket_id` | `string` | OCID of the created Object Storage bucket |

## Related Components

- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciVcn](/docs/catalog/oci/ocivcn) — if using private endpoints for bucket access (future scope)
