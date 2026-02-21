# OciObjectStorageBucket Examples

## Minimal Private Bucket

A bucket with default settings — no public access, Standard tier, Oracle-managed encryption:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciObjectStorageBucket
metadata:
  name: app-data
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciObjectStorageBucket.app-data
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  namespace: "axe1234abc"
  name: "app-data"
```

## Public Read Bucket with Metadata

A bucket for static assets with public object read access and custom metadata:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciObjectStorageBucket
metadata:
  name: static-assets
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciObjectStorageBucket.static-assets
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  namespace: "axe1234abc"
  name: "static-assets"
  accessType: object_read
  metadata:
    team: "frontend"
    purpose: "cdn-origin"
```

## Versioned Bucket with Retention Rules

A compliance-oriented bucket with version history and two retention rules — one with a time-lock:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciObjectStorageBucket
metadata:
  name: audit-logs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciObjectStorageBucket.audit-logs
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  namespace: "axe1234abc"
  name: "audit-logs"
  versioning: enabled
  objectEventsEnabled: true
  kmsKeyId:
    value: "ocid1.key.oc1..example"
  retentionRules:
    - displayName: "regulatory-hold"
      duration:
        timeAmount: 7
        timeUnit: years
      timeRuleLocked: "2027-01-01T00:00:00Z"
    - displayName: "short-term-hold"
      duration:
        timeAmount: 90
        timeUnit: days
```

## Lifecycle Rules with Object Name Filters

A data lake bucket with lifecycle rules that archive logs after 90 days, delete old object versions after one year, and abort stale multipart uploads:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciObjectStorageBucket
metadata:
  name: data-lake
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciObjectStorageBucket.data-lake
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  namespace: "axe1234abc"
  name: "data-lake"
  autoTiering: infrequent_access
  versioning: enabled
  lifecycleRules:
    - name: "archive-logs"
      action: lifecycle_archive
      isEnabled: true
      timeAmount: 90
      timeUnit: days
      target: "objects"
      objectNameFilter:
        inclusionPrefixes:
          - "logs/"
    - name: "tier-down-reports"
      action: lifecycle_infrequent_access
      isEnabled: true
      timeAmount: 30
      timeUnit: days
      target: "objects"
      objectNameFilter:
        inclusionPatterns:
          - "reports/*.csv"
        exclusionPatterns:
          - "reports/latest-*"
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

## Cross-Region Replication

A production bucket with replication to two regions for disaster recovery. Destination buckets must already exist:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciObjectStorageBucket
metadata:
  name: prod-store
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciObjectStorageBucket.prod-store
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  namespace: "axe1234abc"
  name: "prod-store"
  versioning: enabled
  objectEventsEnabled: true
  replicationPolicies:
    - name: "dr-to-phoenix"
      destinationBucketName: "prod-store-dr-phx"
      destinationRegionName: "us-phoenix-1"
    - name: "dr-to-london"
      destinationBucketName: "prod-store-dr-lhr"
      destinationRegionName: "uk-london-1"
```

## Common Operations

### Enable versioning on an existing bucket

Set `versioning: enabled` in the spec. If the bucket previously had versioning disabled, change to `enabled`. If it was enabled and you want to pause, use `suspended` (previous versions are preserved but new versions stop being created).

### Add a lifecycle rule to an existing bucket

Append a new entry to the `lifecycleRules` list and re-apply. All lifecycle rules are managed as a single policy resource — existing rules are preserved as long as they remain in the list.

### Add a replication policy

Append a new entry to `replicationPolicies`. The destination bucket must already exist in the target region. Replication policy fields are immutable — to change the destination, delete the policy entry and create a new one.

### Lock a retention rule

Set `timeRuleLocked` to an RFC 3339 datetime in the future. After that datetime passes, the rule can no longer be shortened or deleted (only duration increases are allowed).

## Best Practices

1. **Use versioning for production data** — protects against accidental overwrites and deletions.
2. **Set lifecycle rules for cost control** — archive or delete objects that are no longer actively accessed.
3. **Abort stale multipart uploads** — add a lifecycle rule with `lifecycle_abort` targeting `"multipart-uploads"` to clean up incomplete uploads.
4. **Use customer-managed KMS keys** for data subject to regulatory encryption requirements.
5. **Enable events** (`objectEventsEnabled: true`) when you need audit trails or event-driven workflows.
6. **Use `valueFrom` references** for `compartmentId` to avoid hardcoding OCIDs and to maintain dependency ordering between OpenMCF resources.
