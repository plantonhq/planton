# GcpGcsBucketIamMember

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpGcsBucketIamMemberSpec defines a single ADDITIVE IAM grant ON a Cloud
Storage bucket: one role, to one member, on one bucket.

Use this kind for a grantee that DEPENDS ON the bucket. The canonical case
is a GcpLoggingSink exporting into the bucket: the sink names the bucket as
its destination, and its writer identity needs roles/storage.objectCreator
on that same bucket. Declared in the bucket's own iam_members, that grant
would make the bucket depend on the sink while the sink depends on the
bucket -- a cycle no chart can deploy. This grant depends on both, so the
order is always bucket, then sink, then grant.

For every other grantee (a workload's service account, a group, a service
agent known in advance), GcpGcsBucket's own iam_members field is the
simpler home: the grants travel with the bucket. Both are additive and
never fight; do not declare the same (role, member) pair in both places,
because removing either one removes the grant.

Common roles:
  roles/storage.objectCreator -- write objects (a sink's writer, an
    uploader that must never read)
  roles/storage.objectViewer  -- read objects
  roles/storage.objectUser    -- read, write, and delete objects
  roles/storage.legacyBucketReader -- list the bucket and read its metadata

Additive semantics: the grant merges into the bucket's IAM policy without
touching any other member's bindings on the same role, and removal
subtracts only this exact (role, member, condition) entry. (Authoritative
binding and policy management, which clobbers everything not listed, is
deliberately not modeled.)

Every field is immutable, mirroring the API: an IAM grant has no update --
changing the bucket, role, member, or condition replaces the grant.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGcsBucketIamMember
metadata:
  name: my-sample-bucket-grant
spec:
  # The bucket whose IAM policy receives this grant, by name (not a gs://
  # URL). Reference a GcpGcsBucket -- its `bucket_id` output is exactly
  # this value.
  bucket:
    value: acme-audit-archive

  # The role to grant on the bucket: write-only, what a sink's writer
  # identity needs.
  role:
    value: roles/storage.objectCreator

  # The identity receiving the grant, in GCP IAM member format -- here a
  # project's Cloud Logging writer identity. Reference a GcpLoggingSink's
  # `writer_identity` output to follow the sink's own identity.
  member:
    value: serviceAccount:service-123456789@gcp-sa-logging.iam.gserviceaccount.com

  # Optional IAM Condition (requires uniform bucket-level access)
  # condition:
  #   title: logs-prefix-only
  #   expression: resource.name.startsWith("projects/_/buckets/acme-audit-archive/objects/logs/")
  #   description: The writer may create objects only under logs/
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.bucket` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_id`) |
| `spec.role` | `string \| valueFrom` | yes |  | GcpIamCustomRole (`status.outputs.name`) |
| `spec.member` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.member`), GcpLoggingSink (`status.outputs.writer_identity`) |
| `spec.condition` | `GcpGcsBucketIamMemberCondition` |  |  |  |
| `spec.condition.title` | `string` | yes |  |  |
| `spec.condition.expression` | `string` | yes |  |  |
| `spec.condition.description` | `string` |  |  |  |

## Field Details

### spec.bucket

`string | valueFrom` · required

The bucket whose IAM policy receives this grant, by its globally unique
name (bucket names are not project-scoped, so there is no project
field). Reference a GcpGcsBucket -- its `bucket_id` output is exactly
this value.

- references: GcpGcsBucket (`status.outputs.bucket_id`)
- rule: a literal bucket must be a bucket name: 3-222 lowercase letters, digits, dots, hyphens, or underscores, starting and ending with a letter or digit (not a gs:// URL)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.role

`string | valueFrom` · required

The role to grant on the bucket. A predefined role
("roles/storage.objectCreator", "roles/storage.objectViewer",
"roles/storage.objectUser", "roles/storage.objectAdmin",
"roles/storage.legacyBucketReader", ...) or a custom role's full name
("projects/<project>/roles/<role_id>" or
"organizations/<org>/roles/<role_id>"). Reference a GcpIamCustomRole to
grant a custom role -- its `name` output is exactly this value.

- references: GcpIamCustomRole (`status.outputs.name`)
- rule: a literal role must be roles/<name>, projects/<project>/roles/<id>, or organizations/<org>/roles/<id>
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpIamCustomRole, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.member

`string | valueFrom` · required

The identity receiving the grant, in IAM member format:
  serviceAccount:<email>  -- a service account or Google service agent.
                             Reference GcpServiceAccount (`member`) or a
                             GcpLoggingSink (`writer_identity`, the
                             sink's writer, already in this format).
  user:<email>, group:<email>, domain:<domain>
  principal://... / principalSet://... -- workload identity federation
  allUsers / allAuthenticatedUsers     -- PUBLIC access to the objects
                                          the role covers; refused by
                                          buckets with public access
                                          prevention enforced
Grants to deleted principals ("deleted:...") are not supported.

- references: GcpServiceAccount (`status.outputs.member`), GcpLoggingSink (`status.outputs.writer_identity`)
- rule: a literal member must be <type>:<value> (serviceAccount:, user:, group:, domain:, principal://, principalSet://) or allUsers / allAuthenticatedUsers, and never a deleted: principal
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.member}} -- a bare string does not parse

### spec.condition

`GcpGcsBucketIamMemberCondition`

Optional IAM Condition restricting when this grant applies (an expiry,
or an object-name prefix such as
resource.name.startsWith("projects/_/buckets/<bucket>/objects/logs/")).
Conditions require the bucket to use uniform bucket-level access. The
condition is part of the grant's identity: the same role with and
without a condition are two independent grants.

### spec.condition.title

`string` · required

Short human-readable title naming the condition's intent, e.g.
"logs-prefix-only".

- rule: {"required":true,"string":{"maxLen":"100"}}

### spec.condition.expression

`string` · required

The CEL condition expression.

- rule: {"required":true}

### spec.condition.description

`string`

Optional longer explanation of what the condition does and why.

- rule: {"string":{"maxLen":"256"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpGcsBucketIamMember, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.bucket` | `string` | The bucket whose IAM policy received the grant, after reference resolution. |
| `status.outputs.role` | `string` | The role that was granted, after reference resolution. |
| `status.outputs.member` | `string` | The member the role was granted to, in IAM member format, after reference resolution. |
| `status.outputs.etag` | `string` | The etag of the bucket's IAM policy after this grant was applied -- a fingerprint of the policy version, useful for audit correlation. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.bucket` | GcpGcsBucket | `status.outputs.bucket_id` |
| `spec.role` | GcpIamCustomRole | `status.outputs.name` |
| `spec.member` | GcpServiceAccount | `status.outputs.member` |
| `spec.member` | GcpLoggingSink | `status.outputs.writer_identity` |

## See Also

- [Overview](../README.md)
