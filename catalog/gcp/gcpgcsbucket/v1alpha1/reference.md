# GcpGcsBucket

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpGcsBucketSpec defines the configuration for a Google Cloud Storage
bucket — the durable object store behind static sites, data lakes, build
artifacts, backups, and every GCP service that stages data (Dataproc,
Cloud Functions sources, Composer DAGs, BigQuery external tables).

Important behavioral notes:

  - bucket_name, location, project, custom placement, hierarchical
    namespace, and enable_object_retention are immutable — changing any
    of them destroys and recreates the bucket (and everything in it).
  - Deleting a non-empty bucket fails unless force_destroy is true, in
    which case every object (including noncurrent versions) is deleted
    first. force_destroy defaults to false — the safe posture for
    data-bearing buckets.
  - A locked retention policy is irreversible: once locked, the policy
    cannot be removed or shortened, and objects cannot be deleted until
    their retention period expires. Unlocking is impossible — GCP forces
    bucket re-creation.
  - Soft delete is on by default in GCP: deleted objects are retained
    (and billed) for 7 days unless soft_delete_policy sets a different
    duration or 0 to disable.
  - Access control is additive IAM: each iam_members entry grants one
    role to one member on this bucket and composes safely with grants
    made elsewhere. Authoritative bindings/policies (which clobber grants
    they don't list) are deliberately not modeled.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGcsBucket
metadata:
  name: test-gcs-bucket
spec:
  bucketName: planton-test-gcs-bucket
  location: us-central1
  uniformBucketLevelAccessEnabled: true
  publicAccessPrevention: enforced
  versioningEnabled: true
  forceDestroy: true
  lifecycleRules:
    - action:
        type: Delete
      condition:
        numNewerVersions: 3
        withState: ARCHIVED
  labels:
    team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.bucketName` | `string` | yes |  |  |
| `spec.location` | `string` | yes |  |  |
| `spec.storageClass` | `string` |  |  |  |
| `spec.forceDestroy` | `bool` |  |  |  |
| `spec.uniformBucketLevelAccessEnabled` | `bool` |  |  |  |
| `spec.publicAccessPrevention` | `string` |  |  |  |
| `spec.versioningEnabled` | `bool` |  |  |  |
| `spec.autoclass` | `GcpGcsBucketAutoclass` |  |  |  |
| `spec.autoclass.enabled` | `bool` | yes |  |  |
| `spec.autoclass.terminalStorageClass` | `string` |  |  |  |
| `spec.lifecycleRules` | `[]GcpGcsBucketLifecycleRule` |  |  |  |
| `spec.lifecycleRules[].action` | `GcpGcsBucketLifecycleAction` | yes |  |  |
| `spec.lifecycleRules[].action.type` | `string` | yes |  |  |
| `spec.lifecycleRules[].action.storageClass` | `string` |  |  |  |
| `spec.lifecycleRules[].condition` | `GcpGcsBucketLifecycleCondition` | yes |  |  |
| `spec.lifecycleRules[].condition.ageDays` | `int32` |  |  |  |
| `spec.lifecycleRules[].condition.createdBefore` | `string` |  |  |  |
| `spec.lifecycleRules[].condition.withState` | `string` |  |  |  |
| `spec.lifecycleRules[].condition.matchesStorageClass` | `[]string` |  |  |  |
| `spec.lifecycleRules[].condition.matchesPrefix` | `[]string` |  |  |  |
| `spec.lifecycleRules[].condition.matchesSuffix` | `[]string` |  |  |  |
| `spec.lifecycleRules[].condition.numNewerVersions` | `int32` |  |  |  |
| `spec.lifecycleRules[].condition.daysSinceNoncurrentTime` | `int32` |  |  |  |
| `spec.lifecycleRules[].condition.noncurrentTimeBefore` | `string` |  |  |  |
| `spec.lifecycleRules[].condition.daysSinceCustomTime` | `int32` |  |  |  |
| `spec.lifecycleRules[].condition.customTimeBefore` | `string` |  |  |  |
| `spec.lifecycleRules[].condition.sizeAboveBytes` | `int64` |  |  |  |
| `spec.lifecycleRules[].condition.sizeBelowBytes` | `int64` |  |  |  |
| `spec.retentionPolicy` | `GcpGcsBucketRetentionPolicy` |  |  |  |
| `spec.retentionPolicy.retentionPeriodSeconds` | `int64` | yes |  |  |
| `spec.retentionPolicy.isLocked` | `bool` |  |  |  |
| `spec.softDeletePolicy` | `GcpGcsBucketSoftDeletePolicy` |  |  |  |
| `spec.softDeletePolicy.retentionDurationSeconds` | `int64` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.requesterPays` | `bool` |  |  |  |
| `spec.defaultEventBasedHold` | `bool` |  |  |  |
| `spec.enableObjectRetention` | `bool` |  |  |  |
| `spec.website` | `GcpGcsBucketWebsite` |  |  |  |
| `spec.website.mainPageSuffix` | `string` |  |  |  |
| `spec.website.notFoundPage` | `string` |  |  |  |
| `spec.corsRules` | `[]GcpGcsBucketCorsRule` |  |  |  |
| `spec.corsRules[].origins` | `[]string` | yes |  |  |
| `spec.corsRules[].methods` | `[]string` | yes |  |  |
| `spec.corsRules[].responseHeaders` | `[]string` |  |  |  |
| `spec.corsRules[].maxAgeSeconds` | `int32` |  |  |  |
| `spec.logging` | `GcpGcsBucketLogging` |  |  |  |
| `spec.logging.logBucket` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_id`) |
| `spec.logging.logObjectPrefix` | `string` |  |  |  |
| `spec.customPlacementConfig` | `GcpGcsBucketCustomPlacementConfig` |  |  |  |
| `spec.customPlacementConfig.dataLocations` | `[]string` | yes |  |  |
| `spec.rpo` | `string` |  |  |  |
| `spec.hierarchicalNamespaceEnabled` | `bool` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.iamMembers` | `[]GcpGcsBucketIamMember` |  |  |  |
| `spec.iamMembers[].role` | `string` | yes |  |  |
| `spec.iamMembers[].member` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.member`) |
| `spec.iamMembers[].condition` | `GcpGcsBucketIamCondition` |  |  |  |
| `spec.iamMembers[].condition.title` | `string` | yes |  |  |
| `spec.iamMembers[].condition.expression` | `string` | yes |  |  |
| `spec.iamMembers[].condition.description` | `string` |  |  |  |
| `spec.ipFilter` | `GcpGcsBucketIpFilter` |  |  |  |
| `spec.ipFilter.mode` | `string` | yes |  |  |
| `spec.ipFilter.publicNetworkSource` | `GcpGcsBucketIpFilterPublicNetworkSource` |  |  |  |
| `spec.ipFilter.publicNetworkSource.allowedIpCidrRanges` | `[]string` | yes |  |  |
| `spec.ipFilter.vpcNetworkSources` | `[]GcpGcsBucketIpFilterVpcNetworkSource` |  |  |  |
| `spec.ipFilter.vpcNetworkSources[].network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_id`) |
| `spec.ipFilter.vpcNetworkSources[].allowedIpCidrRanges` | `[]string` | yes |  |  |
| `spec.ipFilter.allowCrossOrgVpcs` | `bool` |  |  |  |
| `spec.ipFilter.allowAllServiceAgentAccess` | `bool` |  |  |  |
| `spec.encryptionEnforcement` | `GcpGcsBucketEncryptionEnforcement` |  |  |  |
| `spec.encryptionEnforcement.googleManagedRestrictionMode` | `string` |  |  |  |
| `spec.encryptionEnforcement.customerManagedRestrictionMode` | `string` |  |  |  |
| `spec.encryptionEnforcement.customerSuppliedRestrictionMode` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.folders` | `[]GcpGcsBucketFolder` |  |  |  |
| `spec.folders[].name` | `string` | yes |  |  |
| `spec.folders[].forceDestroy` | `bool` |  |  |  |
| `spec.managedFolders` | `[]GcpGcsBucketManagedFolder` |  |  |  |
| `spec.managedFolders[].name` | `string` | yes |  |  |
| `spec.managedFolders[].forceDestroy` | `bool` |  |  |  |
| `spec.notifications` | `[]GcpGcsBucketNotification` |  |  |  |
| `spec.notifications[].topic` | `string \| valueFrom` | yes |  | GcpPubSubTopic (`status.outputs.topic_id`) |
| `spec.notifications[].payloadFormat` | `string` | yes |  |  |
| `spec.notifications[].eventTypes` | `[]string` |  |  |  |
| `spec.notifications[].objectNamePrefix` | `string` |  |  |  |
| `spec.notifications[].customAttributes` | `map<string, string>` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the bucket.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable after creation.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.bucketName

`string` · required

Name of the GCS bucket. Globally unique across ALL of GCP (not just
your project), 3-63 characters: lowercase letters, numbers, hyphens,
dots; must start and end with a letter or number. Deliberately
required: bucket names are a global namespace, so the name deserves an
explicit, stable choice rather than a derived default.
Immutable after creation.

- rule: {"required":true,"string":{"minLen":"3","maxLen":"63","pattern":"^[a-z0-9]([a-z0-9-._]*[a-z0-9])?$"}}

### spec.location

`string` · required

The location for the bucket: a region ("us-east1"), a predefined
dual-region ("NAM4"), or a multi-region ("US", "EU", "ASIA"). For a
custom dual-region, set a multi-region here (e.g. "US") plus
custom_placement_config naming the two regions. Immutable after
creation.

- rule: {"required":true}

### spec.storageClass

`string`

Default storage class for objects written without an explicit class.
Values: "STANDARD" (default; hot data), "NEARLINE" (~monthly access),
"COLDLINE" (~quarterly), "ARCHIVE" (~yearly, cheapest at-rest).
Legacy classes ("MULTI_REGIONAL", "REGIONAL") remain valid API values
for pre-existing buckets but should not be chosen for new ones.
Prefer autoclass over hand-picking a cold class when access patterns
are uncertain. Mutable in place (existing objects keep their class).

### spec.forceDestroy

`bool`

If true, deleting the bucket deletes all contained objects first
(including noncurrent versions) — required to destroy any non-empty
bucket. Defaults to false: destroying a bucket that still holds data
fails instead of silently erasing it. Enable for ephemeral/derived
data; leave off for anything precious.

### spec.uniformBucketLevelAccessEnabled

`bool`

Enable Uniform Bucket-Level Access (UBLA): all access is controlled by
IAM alone and legacy object ACLs are disabled. Strongly recommended —
and required for buckets with hierarchical namespace or managed
folders. GCP's default is false (fine-grained ACLs), and UBLA can be
permanently locked on by GCP 90 days after enablement.

### spec.publicAccessPrevention

`string`

Public access prevention policy:
  ""          -- same as "inherited" (GCP default)
  "inherited" -- inherit the org policy (public access possible unless
                 the org forbids it)
  "enforced"  -- no public access, ever, regardless of IAM grants
Recommended: "enforced" for every bucket not deliberately public.
Mutable in place.

- rule: public_access_prevention must be one of: inherited, enforced

### spec.versioningEnabled

`bool`

Enable object versioning: overwrites and deletes keep the previous
version as a noncurrent object. Pair with a lifecycle rule on
num_newer_versions or days_since_noncurrent_time to bound storage
growth. Cannot be combined with hierarchical namespace. Mutable.

### spec.autoclass

`GcpGcsBucketAutoclass`

Autoclass: GCS automatically transitions each object between storage
classes based on its observed access pattern — the zero-tuning
alternative to hand-written SetStorageClass lifecycle rules.

### spec.autoclass.enabled

`bool` · required · optional (explicit presence)

Enable autoclass. Objects start in STANDARD and transition to colder
classes as they go unread; a read promotes the object back to
STANDARD. Toggling autoclass is allowed but restricted by GCP to
once per 24 hours. Required inside autoclass: declaring the block takes
the feature under management, and the switch says which way. An explicit
false is a real statement (it turns a previously enabled bucket's
autoclass off), which is why the switch carries presence.

- rule: {"required":true}

### spec.autoclass.terminalStorageClass

`string`

The coldest class autoclass may transition objects into:
  ""         -- GCP default ("NEARLINE")
  "NEARLINE" -- stop at NEARLINE
  "ARCHIVE"  -- allow transitions all the way to ARCHIVE (also enables
                COLDLINE as an intermediate step)

- rule: terminal_storage_class must be one of: NEARLINE, ARCHIVE

### spec.lifecycleRules

`[]GcpGcsBucketLifecycleRule`

Lifecycle rules for automatic object management (deletion, storage
class transitions, aborting stale multipart uploads). Up to 100 rules;
all conditions within a rule must match (logical AND).

- rule: {"repeated":{"maxItems":"100"}}

### spec.lifecycleRules[].action

`GcpGcsBucketLifecycleAction` · required

Action to take when the condition matches.

- rule: {"required":true}
- rule: SetStorageClass actions must name the target storage_class
- rule: storage_class is only valid on SetStorageClass actions

### spec.lifecycleRules[].action.type

`string` · required

Action type:
  "Delete"          -- delete the matching object (or its noncurrent
                       version when the condition targets versions)
  "SetStorageClass" -- transition the object to storage_class
  "AbortIncompleteMultipartUpload" -- abort multipart uploads older
                       than the condition's age (reclaims hidden
                       storage from abandoned uploads)

- rule: type must be one of: Delete, SetStorageClass, AbortIncompleteMultipartUpload
- rule: {"required":true}

### spec.lifecycleRules[].action.storageClass

`string`

Target storage class for SetStorageClass actions, e.g. "NEARLINE",
"COLDLINE", "ARCHIVE".

### spec.lifecycleRules[].condition

`GcpGcsBucketLifecycleCondition` · required

Condition selecting the objects the action applies to. All specified
criteria must match (logical AND).

- rule: {"required":true}

### spec.lifecycleRules[].condition.ageDays

`int32` · optional (explicit presence)

Minimum age of the object in days. A set 0 matches all objects.

- rule: {"int32":{"gte":0}}

### spec.lifecycleRules[].condition.createdBefore

`string`

Match objects created before this date (RFC 3339 date, "2026-01-01").

### spec.lifecycleRules[].condition.withState

`string`

Match by version state (requires versioning):
  ""        -- any state (default)
  "LIVE"     -- only the current version
  "ARCHIVED" -- only noncurrent versions
  "ANY"      -- both

- rule: with_state must be one of: LIVE, ARCHIVED, ANY

### spec.lifecycleRules[].condition.matchesStorageClass

`[]string`

Match objects currently in any of these storage classes, e.g.
["STANDARD", "NEARLINE"]. Legacy classes ("MULTI_REGIONAL",
"REGIONAL", "DURABLE_REDUCED_AVAILABILITY") are valid here for
matching long-lived objects.

### spec.lifecycleRules[].condition.matchesPrefix

`[]string`

Match objects whose name starts with any of these prefixes.

### spec.lifecycleRules[].condition.matchesSuffix

`[]string`

Match objects whose name ends with any of these suffixes.

### spec.lifecycleRules[].condition.numNewerVersions

`int32` · optional (explicit presence)

Match noncurrent versions with at least this many newer versions —
the standard "keep the last N versions" cleanup (requires versioning).
A set 0 matches every noncurrent version.

- rule: {"int32":{"gte":0}}

### spec.lifecycleRules[].condition.daysSinceNoncurrentTime

`int32` · optional (explicit presence)

Match noncurrent versions that became noncurrent at least this many
days ago (requires versioning). A set 0 matches immediately.

- rule: {"int32":{"gte":0}}

### spec.lifecycleRules[].condition.noncurrentTimeBefore

`string`

Match noncurrent versions that became noncurrent before this date
(RFC 3339 date; requires versioning).

### spec.lifecycleRules[].condition.daysSinceCustomTime

`int32` · optional (explicit presence)

Match objects whose Custom-Time metadata is at least this many days
old. A set 0 matches any object with Custom-Time set.

- rule: {"int32":{"gte":0}}

### spec.lifecycleRules[].condition.customTimeBefore

`string`

Match objects whose Custom-Time metadata is before this date
(RFC 3339 date).

### spec.lifecycleRules[].condition.sizeAboveBytes

`int64` · optional (explicit presence)

Match objects LARGER than this many bytes. Combine with
size_below_bytes for a size band (e.g. transition only large
artifacts to cold storage).

- rule: {"int64":{"gte":"0"}}

### spec.lifecycleRules[].condition.sizeBelowBytes

`int64` · optional (explicit presence)

Match objects SMALLER than this many bytes.

- rule: {"int64":{"gte":"0"}}

### spec.retentionPolicy

`GcpGcsBucketRetentionPolicy`

Retention policy for WORM (write once, read many) compliance: objects
cannot be deleted or replaced until they reach the retention age.

### spec.retentionPolicy.retentionPeriodSeconds

`int64` · required

Minimum object retention period in seconds (max ~100 years,
3155760000s). Objects cannot be deleted or overwritten until they are
this old. Mutable while unlocked; can only be increased once locked.

- rule: {"required":true,"int64":{"lt":"3155760000","gt":"0"}}

### spec.retentionPolicy.isLocked

`bool`

Lock the retention policy. IRREVERSIBLE: a locked policy can never be
removed or shortened, and the bucket cannot be deleted until every
object passes its retention period. Attempting to unlock forces bucket
re-creation. Validate the policy against real workloads before locking.

### spec.softDeletePolicy

`GcpGcsBucketSoftDeletePolicy`

How long deleted objects remain recoverable (and billed) before being
permanently removed. GCP's default is 7 days (604800s) even when this
block is omitted. Set 0 to disable soft delete entirely — common for
high-churn scratch buckets where the 7-day tail is pure cost.

### spec.softDeletePolicy.retentionDurationSeconds

`int64` · optional (explicit presence)

How long deleted objects remain recoverable, in seconds. GCP default
is 604800 (7 days); 0 disables soft delete; otherwise the value must
be between 7 and 90 days. Soft-deleted storage is billed at the
object's storage class rate.

- rule: retention_duration_seconds must be 0 (disabled) or between 604800 (7 days) and 7776000 (90 days)

### spec.kmsKeyName

`string | valueFrom`

Default customer-managed encryption key (CMEK) for objects written
without an explicit key. The GCS service agent needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on the key. If omitted,
Google-managed encryption is used. Accepts the fully qualified crypto
key path or a reference to a GcpKmsKey resource. Mutable in place
(existing objects keep the key they were written with).

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.requesterPays

`bool`

Requester pays: the caller's project (not the bucket owner) is billed
for data access and egress. Useful for widely shared public datasets.
Mutable in place.

### spec.defaultEventBasedHold

`bool`

Automatically place an event-based hold on every new object — the
object cannot be deleted until the hold is released AND its retention
period (measured from release) expires. For event-driven compliance
like "retain 3 years after account closure". Mutable in place.

### spec.enableObjectRetention

`bool`

Enable per-object retention: individual objects can carry their own
retention configuration, independent of the bucket-level policy.
Can only be set at creation — immutable.

### spec.website

`GcpGcsBucketWebsite`

Static website serving configuration (main page and 404 page) for
requests via the bucket's website endpoint or a load balancer backend
bucket. For production HTTPS sites, front the bucket with the L7 load
balancer family (GcpBackendBucket + URL map + HTTPS proxy).

### spec.website.mainPageSuffix

`string`

Object served for directory requests, e.g. "index.html".

### spec.website.notFoundPage

`string`

Object served when the requested path does not exist, e.g. "404.html".

### spec.corsRules

`[]GcpGcsBucketCorsRule`

CORS rules for direct cross-origin browser access (fonts, direct
uploads, XHR downloads). Not needed when all access goes through a
load balancer or same-origin paths.

### spec.corsRules[].origins

`[]string` · required

Origins allowed to make cross-origin requests, e.g.
"https://example.com". "*" allows any origin.

- rule: {"required":true}

### spec.corsRules[].methods

`[]string` · required

HTTP methods allowed, e.g. ["GET", "HEAD"]. "*" allows any method.

- rule: {"required":true}

### spec.corsRules[].responseHeaders

`[]string`

Response headers browsers are allowed to read.

### spec.corsRules[].maxAgeSeconds

`int32`

How long (seconds) browsers may cache the preflight response.

- rule: {"int32":{"gte":0}}

### spec.logging

`GcpGcsBucketLogging`

Usage/access log delivery to another GCS bucket (classic storage
access logs). Most observability needs are better served by Cloud
Audit Logs; use this for legacy tooling that parses access-log files.

### spec.logging.logBucket

`string | valueFrom` · required

Destination bucket that receives the log objects. Accepts a literal
bucket name or a reference to another GcpGcsBucket resource.

- references: GcpGcsBucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.logging.logObjectPrefix

`string`

Prefix for log object names. Defaults to this bucket's name.

### spec.customPlacementConfig

`GcpGcsBucketCustomPlacementConfig`

Custom dual-region placement: exactly two regions (e.g. ["US-EAST1",
"US-WEST1"]) that must belong to the multi-region set in `location`.
Gives dual-region durability with region pinning. Immutable after
creation.

### spec.customPlacementConfig.dataLocations

`[]string` · required

Exactly two regions forming the custom dual-region, e.g.
["US-EAST1", "US-WEST1"]. Both must belong to the multi-region set in
`location`. Immutable after creation.

- rule: {"repeated":{"minItems":"2","maxItems":"2"}}

### spec.rpo

`string`

Recovery point objective for dual- and multi-region buckets:
  ""            -- provider default ("DEFAULT")
  "DEFAULT"     -- asynchronous replication with no SLA on lag
  "ASYNC_TURBO" -- turbo replication (15-minute RPO SLA; dual-region
                   only, additional cost)
Mutable in place.

- rule: rpo must be one of: DEFAULT, ASYNC_TURBO

### spec.hierarchicalNamespaceEnabled

`bool`

Enable hierarchical namespace (HNS): the bucket gets real folder
semantics with atomic folder renames — required for Hadoop/Spark-style
workloads that rename directories. Requires uniform bucket-level
access and cannot be combined with object versioning. Immutable —
only settable at creation.

### spec.labels

`map<string, string>`

User-defined labels attached to the bucket, for cost attribution and
fleet queries. Merged with Planton's platform labels (which win on key
conflicts). Mutable in place.

### spec.iamMembers

`[]GcpGcsBucketIamMember`

Additive IAM grants on this bucket. Each entry grants one role to one
member and composes safely with grants made by other tools or charts —
removal subtracts only that exact (role, member) pair.

Common roles:
  roles/storage.objectViewer  -- read objects
  roles/storage.objectAdmin   -- full object control (no bucket admin)
  roles/storage.admin         -- full bucket + object control

Public access: grant roles/storage.objectViewer to "allUsers" (also
requires public_access_prevention to be "inherited" and the org policy
to allow it).

### spec.iamMembers[].role

`string` · required

The role to grant, e.g. "roles/storage.objectViewer",
"roles/storage.objectAdmin", "roles/storage.admin", or a custom
role's fully-qualified name.

- rule: {"required":true}

### spec.iamMembers[].member

`string | valueFrom` · required

The identity receiving the grant, in GCP IAM member format:
  serviceAccount:<email>  -- a service account (the most common in IaC;
                             reference a GcpServiceAccount resource —
                             its `member` output is exactly this value)
  user:<email> / group:<email> / domain:<domain>
  allUsers / allAuthenticatedUsers -- public access (grant with care)

- references: GcpServiceAccount (`status.outputs.member`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.member}} -- a bare string does not parse

### spec.iamMembers[].condition

`GcpGcsBucketIamCondition`

Optional IAM Condition restricting when this grant applies (e.g. only
objects under a prefix, or before an expiry date). The condition is
part of the grant's identity: the same role with and without a
condition are two independent grants.

### spec.iamMembers[].condition.title

`string` · required

Short human-readable title identifying the condition's intent,
e.g. "reports-prefix-only".

- rule: {"required":true,"string":{"maxLen":"100"}}

### spec.iamMembers[].condition.expression

`string` · required

The CEL condition expression, e.g.
resource.name.startsWith("projects/_/buckets/b/objects/reports/").

- rule: {"required":true}

### spec.iamMembers[].condition.description

`string`

Optional longer explanation of what the condition does.

- rule: {"string":{"maxLen":"256"}}

### spec.ipFilter

`GcpGcsBucketIpFilter`

Network-layer IP filtering: restrict which public CIDR ranges and
which VPC networks may reach the bucket at all, before IAM is even
evaluated. Defense-in-depth for data-exfiltration control — IAM
decides WHO, the IP filter decides FROM WHERE. Mutable in place.

- rule: an Enabled ip_filter needs public_network_source and/or vpc_network_sources to define allowed origins

### spec.ipFilter.mode

`string` · required

The filter mode:
  "Enabled"  -- only the listed sources may reach the bucket
  "Disabled" -- filter retained but inactive (all sources allowed)

- rule: mode must be one of: Enabled, Disabled
- rule: {"required":true}

### spec.ipFilter.publicNetworkSource

`GcpGcsBucketIpFilterPublicNetworkSource`

Public internet sources allowed to access the bucket, as IPv4/IPv6
CIDR ranges.

### spec.ipFilter.publicNetworkSource.allowedIpCidrRanges

`[]string` · required

Public IPv4/IPv6 CIDR ranges allowed to access the bucket,
e.g. "203.0.113.0/24".

- rule: {"repeated":{"minItems":"1"}}

### spec.ipFilter.vpcNetworkSources

`[]GcpGcsBucketIpFilterVpcNetworkSource`

VPC networks allowed to access the bucket, each with its own CIDR
allowlist.

### spec.ipFilter.vpcNetworkSources[].network

`string | valueFrom` · required

The VPC network, in the form projects/{project}/global/networks/{name}.
Accepts a literal path or a reference to a GcpVpcNetwork resource
(its network_id output is exactly this value).

- references: GcpVpcNetwork (`status.outputs.network_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_id}} -- a bare string does not parse

### spec.ipFilter.vpcNetworkSources[].allowedIpCidrRanges

`[]string` · required

IPv4/IPv6 CIDR ranges within this network allowed to access the
bucket.

- rule: {"repeated":{"minItems":"1"}}

### spec.ipFilter.allowCrossOrgVpcs

`bool`

Allow VPC network sources that belong to a different organization
than the bucket.

### spec.ipFilter.allowAllServiceAgentAccess

`bool`

Exempt Google service agents (the identities GCP services act as —
e.g. the Storage transfer agent) from the IP filter, so managed
integrations keep working when the filter is Enabled.

### spec.encryptionEnforcement

`GcpGcsBucketEncryptionEnforcement`

Encryption-type enforcement for NEW objects: restrict which encryption
mechanisms (Google-managed, customer-managed KMS, customer-supplied)
may be used when writing objects into this bucket. Applies to new
objects only — existing objects keep their encryption. Mutable in
place.

### spec.encryptionEnforcement.googleManagedRestrictionMode

`string`

Restriction for Google-managed encryption keys (GMEK) — the default
encryption objects get when no KMS key applies.

- rule: google_managed_restriction_mode must be one of: NotRestricted, FullyRestricted

### spec.encryptionEnforcement.customerManagedRestrictionMode

`string`

Restriction for customer-managed encryption keys (CMEK, Cloud KMS).

- rule: customer_managed_restriction_mode must be one of: NotRestricted, FullyRestricted

### spec.encryptionEnforcement.customerSuppliedRestrictionMode

`string`

Restriction for customer-supplied encryption keys (CSEK — raw keys
provided per request).

- rule: customer_supplied_restriction_mode must be one of: NotRestricted, FullyRestricted

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the bucket is deleted (subject to force_destroy)
  "PREVENT" -- destroy FAILS; a guard rail for buckets that must
               never be removed by automation
  "ABANDON" -- the bucket is removed from management but left
               running in GCP (an orphan by design — reserve for
               deliberate hand-offs)
Mutable in place.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

### spec.folders

`[]GcpGcsBucketFolder`

Folders to create inside the bucket — REAL directories with atomic
rename semantics, available only on hierarchical-namespace buckets
(hierarchical_namespace_enabled, which itself requires uniform
bucket-level access and is create-time only). Parent folders must be
listed explicitly: creating "logs/2026/" requires a "logs/" entry too
— the Storage API does not auto-create missing parents.
The kind-level deletion_policy applies to each folder resource as
well as the bucket (PREVENT guards them; ABANDON leaves them behind).

### spec.folders[].name

`string` · required

The folder path, WITH the trailing slash the Storage API requires:
"logs/", "logs/2026/", "a-b/d-f/". Renaming is atomic on HNS buckets,
but through this spec a name change is destroy-and-recreate (the
path is the resource's identity).

The API does NOT auto-create missing parents, so nested paths need
every ancestor listed as its own entry ("logs/2026/" needs "logs/"
too) — the modules then create parents before children and delete
children before parents. Nesting is capped at 5 levels.

- rule: {"required":true,"string":{"pattern":"^([^/]+/){1,5}$"}}

### spec.folders[].forceDestroy

`bool`

If true, destroying this folder first deletes every object under it
(the provider sweeps the prefix client-side, in parallel), then any
sub-folders. Defaults to false: destroying a non-empty folder fails
instead of silently erasing data — same safe posture as the bucket's
own force_destroy, decided per folder.

### spec.managedFolders

`[]GcpGcsBucketManagedFolder`

Managed folders — permission anchor points inside the bucket. Unlike
plain folders they exist on ANY bucket with uniform bucket-level
access (hierarchical namespace not required) and their purpose is
prefix-scoped IAM: grant a role on "reports/" without granting it on
the whole bucket. Grants on a managed folder are made through the
managed-folder IAM surface (composition), not through this kind's
bucket-level iam_members. The kind-level deletion_policy applies to
each managed folder as well as the bucket.

### spec.managedFolders[].name

`string` · required

The managed-folder path, WITH the trailing slash the Storage API
requires: "reports/", "reports/2026/". The path is the resource's
identity — changing it is destroy-and-recreate. Managed folders are
independent prefix anchors, not a hierarchy — "reports/2026/" does
not need a "reports/" managed folder to exist.

- rule: {"required":true,"string":{"pattern":"^([^/]+/)+$"}}

### spec.managedFolders[].forceDestroy

`bool`

If true, the managed folder can be destroyed even while objects live
under its prefix. Unlike a plain folder's force_destroy this is
SERVER-side and non-destructive to data: the objects survive and
simply stop being covered by the managed folder's IAM. Defaults to
false (destroy fails while the prefix is non-empty).

### spec.notifications

`[]GcpGcsBucketNotification`

Pub/Sub notification configurations: GCS publishes an event to the
named topic whenever objects change in this bucket — the standard
trigger surface for event-driven pipelines (Cloud Run, Cloud
Functions, Eventarc all consume the topic downstream).

HARD PREREQUISITE the API enforces at create time: the project's GCS
service agent — service-{project_number}@gs-project-accounts.iam.gserviceaccount.com,
where {project_number} is this kind's project_number output — must
hold roles/pubsub.publisher on the topic (or project-wide). The agent
is NOT granted automatically; compose the grant alongside this
resource (GcpProjectIamMember for project scope, or the topic's own
IAM surface) exactly like the CMEK key grant for kms_key_name.

### spec.notifications[].topic

`string | valueFrom` · required

The Pub/Sub topic that receives the events, as the FULLY-QUALIFIED
resource path "projects/{project}/topics/{name}" — the API accepts
only this form (a bare topic name is rejected). Reference a
GcpPubSubTopic — its topic_id output is exactly this value. The
topic may live in a different project than the bucket.

Before creation the project's GCS service agent must hold
roles/pubsub.publisher on this topic — see the notifications field
comment on the spec for the composition pattern.

- references: GcpPubSubTopic (`status.outputs.topic_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.notifications[].payloadFormat

`string` · required

The payload attached to each event message:
  "JSON_API_V1" -- the object's full JSON representation (the common
                   choice; consumers read metadata without a GET)
  "NONE"        -- no payload; event attributes only (cheapest)

- rule: payload_format must be one of: JSON_API_V1, NONE
- rule: {"required":true}

### spec.notifications[].eventTypes

`[]string`

Which object events publish to the topic. Empty means ALL event
types. Values:
  "OBJECT_FINALIZE"        -- a new object (or new version) is written
  "OBJECT_METADATA_UPDATE" -- an object's metadata changed
  "OBJECT_DELETE"          -- an object (or version) was deleted or
                              overwritten
  "OBJECT_ARCHIVE"         -- a live version became noncurrent
                              (versioned buckets only)

- rule: {"repeated":{"items":{"cel":[{"id":"valid_event_type","message":"event_types entries must be one of: OBJECT_FINALIZE, OBJECT_METADATA_UPDATE, OBJECT_DELETE, OBJECT_ARCHIVE","expression":"this in ['OBJECT_FINALIZE', 'OBJECT_METADATA_UPDATE', 'OBJECT_DELETE', 'OBJECT_ARCHIVE']"}]}}}

### spec.notifications[].objectNamePrefix

`string`

Only objects whose names start with this prefix generate events —
scope a busy bucket's feed to one subtree (e.g. "uploads/").

### spec.notifications[].customAttributes

`map<string, string>`

Custom key:value attributes stamped onto every event message the
topic receives — routing hints for subscribers (e.g. env: prod).

## Validation Rules

- `autoclass_conflicts_with_lifecycle_storage_class`: autoclass and SetStorageClass lifecycle rules both manage storage classes — enable only one transition mechanism
- `folders_require_hierarchical_namespace`: folders exist only on hierarchical-namespace buckets — enable hierarchical_namespace_enabled (create-time only, and it requires uniform bucket-level access)
- `managed_folders_require_uniform_access`: managed folders require uniform bucket-level access — enable uniform_bucket_level_access_enabled
- `folder_names_unique`: folders must not repeat the same path
- `managed_folder_names_unique`: managed_folders must not repeat the same path

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpGcsBucket, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.bucket_id` | `string` | ID of the bucket. For GCS this equals the globally unique bucket name — the value every consumer (backend buckets, function sources, Dataproc staging, Pub/Sub sinks) references. |
| `status.outputs.bucket_name` | `string` | Name of the bucket (identical to bucket_id; exported under both keys so consumers can use whichever reads naturally). |
| `status.outputs.url` | `string` | The base URI of the bucket, in the form gs://<bucket_name>. |
| `status.outputs.self_link` | `string` | The API self link of the bucket, e.g. https://www.googleapis.com/storage/v1/b/<bucket_name>. |
| `status.outputs.location` | `string` | Location of the bucket as reported by GCS (upper-cased region, dual-region, or multi-region, e.g. "US-EAST1", "US"). |
| `status.outputs.project_number` | `int64` | Numeric project number of the project that owns the bucket. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.logging.logBucket` | GcpGcsBucket | `status.outputs.bucket_id` |
| `spec.iamMembers[].member` | GcpServiceAccount | `status.outputs.member` |
| `spec.ipFilter.vpcNetworkSources[].network` | GcpVpcNetwork | `status.outputs.network_id` |
| `spec.notifications[].topic` | GcpPubSubTopic | `status.outputs.topic_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpBackendBucket | `spec.bucketName` | `status.outputs.bucket_id` |
| GcpCloudComposerEnvironment | `spec.storageBucket` | `status.outputs.bucket_id` |
| GcpCloudFunction | `spec.buildConfig.source.storageSource.bucket` | `status.outputs.bucket_id` |
| GcpCloudRun | `spec.volumes[].gcs.bucket` | `status.outputs.bucket_id` |
| GcpCloudRunJob | `spec.template.volumes[].gcs.bucket` | `status.outputs.bucket_id` |
| GcpCloudRunWorkerPool | `spec.volumes[].gcs.bucket` | `status.outputs.bucket_id` |
| GcpColabSchedule | `spec.notebookExecutionJob.gcsOutputUri` | `status.outputs.url` |
| GcpDataprocCluster | `spec.clusterConfig.stagingBucket` | `status.outputs.bucket_id` |
| GcpDataprocCluster | `spec.clusterConfig.tempBucket` | `status.outputs.bucket_id` |
| GcpDataprocCluster | `spec.virtualClusterConfig.stagingBucket` | `status.outputs.bucket_id` |
| GcpGcsBucket | `spec.logging.logBucket` | `status.outputs.bucket_id` |
| GcpLoggingSink | `spec.destination.gcsBucket` | `status.outputs.bucket_id` |
| GcpPubSubSubscription | `spec.cloudStorageConfig.bucket` | `status.outputs.bucket_id` |
| GcpPubSubTopic | `spec.ingestionDataSourceSettings.cloudStorage.bucket` | `status.outputs.bucket_id` |
| KubernetesOpenBao | `spec.backup.objectStore.gcs.bucket` | `status.outputs.bucket_name` |

## See Also

- [Overview](../README.md)
