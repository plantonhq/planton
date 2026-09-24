# GcpSecretManagerSecret

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpSecretManagerSecretSpec defines a Secret Manager secret — the
container that stores versioned secret payloads (API keys, passwords,
certificates) with replication, rotation notifications, expiry, CMEK,
and secret-scoped IAM access grants.

One kind covers both scopes. Leave `region` empty for a GLOBAL secret
(payloads replicated per the `replication` message — automatically when
it is omitted); set it for a REGIONAL secret whose payloads never leave
that region (data-residency regimes). The two scopes are separate GCP
API surfaces with one behavioral difference this spec mirrors:
replication is a global-only concept, and regional secrets take their
CMEK directly via `customer_managed_encryption`.

The optional `initial_version` stores the first payload at create time,
and `iam_members` grants secret-SCOPED access (typically
roles/secretmanager.secretAccessor to the workload's service account) —
so one manifest takes a consumer from nothing to a readable secret.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSecretManagerSecret
metadata:
  name: my-sample-secret
spec:
  # GCP project that owns the secret.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # The secret ID; omit to default to metadata.name. Immutable.
  secretId: my-sample-secret

  # Leave region empty for a GLOBAL secret (replication control below);
  # set it for a REGIONAL secret whose payloads never leave the region.

  # GLOBAL secrets only. Omit entirely for automatic replication (the
  # right default when no residency regime applies); shown here for the
  # complete surface.
  replication:
    auto: {}

  # The first payload, stored as version 1 — one manifest yields a
  # READABLE secret. In charts, wire data via valueFrom from a producing
  # resource's sensitive output instead of a literal.
  initialVersion:
    data:
      value: super-secret-value
    enabled: true
    # What destroying THIS resource does to version 1:
    # DELETE (default) / DISABLE (recoverable) / ABANDON.
    deletionPolicy: DELETE

  # Secret-SCOPED additive grants — the access story: each consuming
  # workload's service account gets secretAccessor on exactly this
  # secret, never a project-wide role.
  iamMembers:
    - role: roles/secretmanager.secretAccessor
      member:
        valueFrom:
          kind: GcpServiceAccount
          # The service-account kind's PUBLISHED prerequisite fixture.
          name: planton-oss-e2e-gsa-prereq
          fieldPath: status.outputs.member

  # Friendly alias -> version number; re-point consumers atomically.
  versionAliases:
    current: "1"

  # Delayed version destruction (>= 86400s): destroy first disables,
  # restore window applies. Whole-secret deletion ignores this.
  versionDestroyTtl: 86400s

  # User metadata labels, merged with Planton's platform labels.
  labels:
    team: platform

  # Freeform non-identifying metadata (not queryable in filters).
  annotations:
    owner: platform-team

  # Engine-side destroy guard, evaluated BEFORE deletionPolicy.
  deletionProtection: false

  # What a destroy does to the whole secret: DELETE (default — every
  # version destroyed, unrecoverable), PREVENT, or ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.secretId` | `string` |  |  |  |
| `spec.region` | `string` |  |  |  |
| `spec.replication` | `GcpSecretManagerSecretReplication` |  |  |  |
| `spec.replication.auto` | `GcpSecretManagerSecretReplicationAuto` |  |  |  |
| `spec.replication.auto.customerManagedEncryption` | `GcpSecretManagerSecretCmek` |  |  |  |
| `spec.replication.auto.customerManagedEncryption.kmsKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.replication.userManaged` | `GcpSecretManagerSecretReplicationUserManaged` |  |  |  |
| `spec.replication.userManaged.replicas` | `[]GcpSecretManagerSecretReplica` | yes |  |  |
| `spec.replication.userManaged.replicas[].location` | `string` | yes |  |  |
| `spec.replication.userManaged.replicas[].customerManagedEncryption` | `GcpSecretManagerSecretCmek` |  |  |  |
| `spec.replication.userManaged.replicas[].customerManagedEncryption.kmsKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.customerManagedEncryption` | `GcpSecretManagerSecretCmek` |  |  |  |
| `spec.customerManagedEncryption.kmsKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |
| `spec.expireTime` | `string` |  |  |  |
| `spec.ttl` | `string` |  |  |  |
| `spec.versionAliases` | `map<string, string>` |  |  |  |
| `spec.versionDestroyTtl` | `string` |  |  |  |
| `spec.rotation` | `GcpSecretManagerSecretRotation` |  |  |  |
| `spec.rotation.rotationPeriod` | `string` |  |  |  |
| `spec.rotation.nextRotationTime` | `string` |  |  |  |
| `spec.topics` | `[]string \| valueFrom` |  |  | GcpPubSubTopic (`status.outputs.topic_id`) |
| `spec.initialVersion` | `GcpSecretManagerSecretInitialVersion` |  |  |  |
| `spec.initialVersion.data` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.initialVersion.enabled` | `bool` |  | `true` |  |
| `spec.initialVersion.isBase64` | `bool` |  |  |  |
| `spec.initialVersion.deletionPolicy` | `string` |  |  |  |
| `spec.iamMembers` | `[]GcpSecretManagerSecretIamMember` |  |  |  |
| `spec.iamMembers[].role` | `string` | yes |  |  |
| `spec.iamMembers[].member` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.member`) |
| `spec.iamMembers[].condition` | `GcpSecretManagerSecretIamCondition` |  |  |  |
| `spec.iamMembers[].condition.title` | `string` | yes |  |  |
| `spec.iamMembers[].condition.expression` | `string` | yes |  |  |
| `spec.iamMembers[].condition.description` | `string` |  |  |  |
| `spec.deletionProtection` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the secret. Can be a literal project ID or a
reference to a GcpProject resource. If omitted, the provider's default
project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.secretId

`string`

The secret ID — the last segment of the secret's resource name
(projects/{p}/secrets/{id}). Defaults to metadata.name when left
empty. Letters, numbers, underscores, and hyphens; at most 255
characters. Immutable: changing it destroys and recreates the secret
(and every version with it).

- rule: secret_id may contain only letters, digits, underscores, and hyphens (at most 255 characters)

### spec.region

`string`

Region for a REGIONAL secret (e.g. "us-central1") whose payloads never
leave that region — the data-residency posture. Leave empty for a
GLOBAL secret with replication control. Immutable: a secret cannot
move between scopes.

- rule: region must be a valid GCP region name such as us-central1, or empty for a global secret

### spec.replication

`GcpSecretManagerSecretReplication`

GLOBAL secrets only: where payload replicas live. Omit for automatic
replication (Google chooses placement — the right default when no
residency regime applies); the module then configures the API's `auto`
mode. Set user_managed to pin replicas to specific regions.
Immutable: replication cannot change after create.

- rule: set exactly one of auto or user_managed (or omit replication entirely for automatic placement)

### spec.replication.auto

`GcpSecretManagerSecretReplicationAuto`

Google chooses replica placement. Set this arm explicitly (over
omitting replication) only to attach a CMEK key to automatic
replication.

### spec.replication.auto.customerManagedEncryption

`GcpSecretManagerSecretCmek`

Encrypt payloads with a customer-managed KMS key. Must be a GLOBAL
KMS key for automatic replication; the Secret Manager service agent
needs roles/cloudkms.cryptoKeyEncrypterDecrypter on it.

### spec.replication.auto.customerManagedEncryption.kmsKey

`string | valueFrom` · required

Full KMS crypto key resource path
(projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}) — a literal or
a reference to a GcpKmsKey resource.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.replication.userManaged

`GcpSecretManagerSecretReplicationUserManaged`

Pin payload replicas to specific regions — the residency-controlled
form.

### spec.replication.userManaged.replicas

`[]GcpSecretManagerSecretReplica` · required

The regions holding payload replicas (at least one). Reads are served
from the nearest replica; writes go to all.

- rule: {"repeated":{"minItems":"1"}}

### spec.replication.userManaged.replicas[].location

`string` · required

The replica's region (e.g. "us-east1").

- rule: {"required":true}

### spec.replication.userManaged.replicas[].customerManagedEncryption

`GcpSecretManagerSecretCmek`

Encrypt this replica with a customer-managed KMS key IN THE SAME
region. With multiple replicas, either every replica sets a key or
none does (the API's own rule).

### spec.replication.userManaged.replicas[].customerManagedEncryption.kmsKey

`string | valueFrom` · required

Full KMS crypto key resource path
(projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}) — a literal or
a reference to a GcpKmsKey resource.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.customerManagedEncryption

`GcpSecretManagerSecretCmek`

REGIONAL secrets only: encrypt payloads with a customer-managed KMS
key instead of Google-managed encryption. The key must be in the same
region as the secret, and the Secret Manager service agent needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on it.

### spec.customerManagedEncryption.kmsKey

`string | valueFrom` · required

Full KMS crypto key resource path
(projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}) — a literal or
a reference to a GcpKmsKey resource.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.labels

`map<string, string>`

User labels attached to the secret, merged with Planton's platform
labels (which win on key conflicts).

### spec.annotations

`map<string, string>`

Annotations attached to the secret — freeform non-identifying
metadata (up to 64 entries; keys at most 63 characters). Unlike
labels, annotation values are not queryable in filters.

### spec.tags

`map<string, string>`

Resource manager tags bound at secret creation, as
tagKeys/{id} -> tagValues/{id} pairs — the org-policy surface (IAM
conditions, CMEK enforcement). Immutable: changing tags destroys and
recreates the secret.

### spec.expireTime

`string`

Timestamp when the secret auto-DELETES (RFC3339 UTC, e.g.
"2027-01-01T00:00:00Z"). Deletion removes every version — data
unrecoverable. At most one of expire_time or ttl.

- rule: expire_time must be an RFC3339 UTC timestamp such as 2027-01-01T00:00:00Z

### spec.ttl

`string`

Lifetime after which the secret auto-DELETES, as a seconds duration
(e.g. "7776000s" for 90 days). At most one of expire_time or ttl.

- rule: ttl must be a seconds duration such as 7776000s

### spec.versionAliases

`map<string, string>`

Version aliases: friendly name -> version NUMBER (e.g. "prod" -> "3").
Consumers can then address projects/{p}/secrets/{id}/versions/prod and
re-pointing the alias re-targets every consumer without touching them.
TEMPORAL CONSTRAINT (live API truth): GCP validates aliases against
EXISTING versions at secret create/update ("Aliases cannot be assigned
to versions that don't exist"). A first apply that both seeds
initial_version and aliases it is therefore rejected — the version is
created after the secret. Deploy first, then add the alias on a
subsequent apply once the version exists.

### spec.versionDestroyTtl

`string`

Delayed version destruction: when set (a seconds duration, minimum
"86400s" — 24h), destroying a version first DISABLES it for this
window, during which it can be restored — the undo buffer for fat
fingers. Empty destroys immediately.

- rule: version_destroy_ttl must be a seconds duration of at least 86400s (24 hours)

### spec.rotation

`GcpSecretManagerSecretRotation`

Rotation REMINDERS: GCP publishes a message to `topics` on the
schedule — it does not rotate anything itself; the subscriber (a
Cloud Function, a pipeline) performs the actual rotation. Requires
topics.

### spec.rotation.rotationPeriod

`string`

Time between rotation reminders, as a seconds duration — at least
"3600s" (1 hour), at most "3153600000s" (100 years). Setting it
requires next_rotation_time.

- rule: rotation_period must be a seconds duration of at least 3600s (1 hour)

### spec.rotation.nextRotationTime

`string`

When the FIRST (or next) rotation reminder fires (RFC3339 UTC, e.g.
"2026-09-01T00:00:00Z"). GCP advances it by rotation_period after each
reminder.

- rule: next_rotation_time must be an RFC3339 UTC timestamp such as 2026-09-01T00:00:00Z

### spec.topics

`[]string | valueFrom`

Pub/Sub topics notified on secret lifecycle events (version added,
rotation due, expiry approaching) — at most 10. Each entry is the full
topic path projects/{p}/topics/{t}: a literal or a reference to a
GcpPubSubTopic resource. The Secret Manager service agent needs
roles/pubsub.publisher on each topic.

- references: GcpPubSubTopic (`status.outputs.topic_id`)
- rule: {"repeated":{"maxItems":"10"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.initialVersion

`GcpSecretManagerSecretInitialVersion`

The first secret payload, stored as version 1 at create time — so one
manifest yields a READABLE secret. Omit to create the container only
(versions added later via GCP tooling or rotation pipelines).

### spec.initialVersion.data

`string | valueFrom` · required · sensitive

The secret payload (at most 64KiB). A secret value: the platform
stores it as a managed-secret reference and resolves it just-in-time
at deploy — it never sits in plaintext in the control plane. In
charts, wire it via valueFrom from a producing resource's sensitive
output (e.g. a generated credential) instead of a literal.
Immutable: changing the payload creates a NEW version through GCP
tooling or rotation — this field only seeds version 1.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.initialVersion.enabled

`bool` · optional (explicit presence)

Whether version 1 starts ENABLED (default true). A disabled version
exists but cannot be accessed — staging a payload ahead of cut-over.
Both IaC engines send the value explicitly so behavior is identical
regardless of engine.

- default: `true`

### spec.initialVersion.isBase64

`bool`

Set true when `data` is base64-encoded binary (the API stores the
DECODED bytes). Leave false for text payloads. Immutable.

### spec.initialVersion.deletionPolicy

`string`

What destroying this resource does to version 1 specifically:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the version is destroyed (or disabled-then-destroyed
               when the secret sets version_destroy_ttl)
  "DISABLE" -- the version is disabled but kept — recoverable
  "ABANDON" -- the version is left untouched in GCP
Distinct from the secret-level deletion_policy: deleting the whole
secret always removes every version regardless of this setting.

- rule: initial_version.deletion_policy must be one of: DELETE, DISABLE, ABANDON

### spec.iamMembers

`[]GcpSecretManagerSecretIamMember`

Secret-SCOPED IAM grants — typically roles/secretmanager.secretAccessor
to the service account of each workload that reads this secret. Grants
are additive (iam_member semantics): they compose safely with grants
made elsewhere and never clobber them.

### spec.iamMembers[].role

`string` · required

The role to grant — most commonly roles/secretmanager.secretAccessor
(read payloads); also roles/secretmanager.viewer (metadata only) or
roles/secretmanager.secretVersionManager (add/destroy versions).

- rule: {"required":true}

### spec.iamMembers[].member

`string | valueFrom` · required

The identity receiving the grant, in GCP IAM member format:
  serviceAccount:<email>  -- a workload identity (the most common in
                             IaC; reference a GcpServiceAccount — its
                             `member` output is exactly this value)
  user:<email> / group:<email> / domain:<domain>

- references: GcpServiceAccount (`status.outputs.member`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.member}} -- a bare string does not parse

### spec.iamMembers[].condition

`GcpSecretManagerSecretIamCondition`

Optional IAM Condition restricting when this grant applies. The
condition is part of the grant's identity: the same role with and
without a condition are two independent grants.

### spec.iamMembers[].condition.title

`string` · required

Short title identifying the condition's purpose (shown in the console).

- rule: {"required":true}

### spec.iamMembers[].condition.expression

`string` · required

The CEL condition expression, e.g.
request.time < timestamp("2027-01-01T00:00:00Z").

- rule: {"required":true}

### spec.iamMembers[].condition.description

`string`

What the condition enforces and why — for the operator auditing
access later.

### spec.deletionProtection

`bool`

Engine-side destroy guard (default false). When true in state, any
plan that would delete the secret FAILS before reaching the API —
flip it to false explicitly, then destroy. Independent of (and
evaluated before) deletion_policy.

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed
(evaluated only after deletion_protection allows the plan):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the secret and ALL its versions are deleted; payloads
               unrecoverable (version_destroy_ttl does not apply to
               whole-secret deletion)
  "PREVENT" -- destroy FAILS; belt-and-suspenders with
               deletion_protection for production credentials
  "ABANDON" -- the secret is removed from management but stays
               readable in GCP (IAM grants and versions intact)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `replication_is_global_only`: replication applies only to global secrets — for a regional secret (region set), payloads live in that region by definition
- `regional_cmek_needs_region`: customer_managed_encryption applies only to regional secrets — for a global secret, set CMEK inside replication (auto or per-replica)
- `expire_time_xor_ttl`: set at most one of expire_time or ttl — GCP accepts only one expiry form
- `rotation_requires_topics`: rotation requires at least one entry in topics — rotation reminders are delivered as Pub/Sub messages
- `rotation_period_requires_next_time`: rotation_period requires next_rotation_time — GCP needs the first reminder's timestamp to start the cadence

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpSecretManagerSecret, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.secret_name` | `string` | The full resource name of the secret. Global: projects/{project}/secrets/{secret_id} Regional: projects/{project}/locations/{region}/secrets/{secret_id} The handle consumers use to reference the secret (e.g. a Cloud Run valueFromSecret mount takes this name). |
| `status.outputs.secret_id` | `string` | The short secret ID (the last segment of secret_name). |
| `status.outputs.latest_version_name` | `string` | The full resource name of the version created from initial_version (…/versions/1); empty when no initial_version was configured. Consumers pinning an exact version (instead of the "latest" alias) reference this value. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.replication.auto.customerManagedEncryption.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.replication.userManaged.replicas[].customerManagedEncryption.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.customerManagedEncryption.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.topics` | GcpPubSubTopic | `status.outputs.topic_id` |
| `spec.iamMembers[].member` | GcpServiceAccount | `status.outputs.member` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDatastreamConnectionProfile | `spec.mysqlProfile.secretManagerStoredPassword` | `status.outputs.latest_version_name` |
| GcpDatastreamConnectionProfile | `spec.postgresqlProfile.secretManagerStoredPassword` | `status.outputs.latest_version_name` |
| GcpDatastreamConnectionProfile | `spec.oracleProfile.secretManagerStoredPassword` | `status.outputs.latest_version_name` |
| GcpDatastreamConnectionProfile | `spec.sqlServerProfile.secretManagerStoredPassword` | `status.outputs.latest_version_name` |
| GcpDatastreamConnectionProfile | `spec.mongodbProfile.secretManagerStoredPassword` | `status.outputs.latest_version_name` |
| GcpDatastreamConnectionProfile | `spec.mongodbProfile.sslConfig.secretManagerStoredClientKey` | `status.outputs.latest_version_name` |
| GcpVertexAiAgentEngine | `spec.spec.deploymentSpec.secretEnv[].secretRef.secret` | `status.outputs.secret_id` |

## See Also

- [Overview](../README.md)
