# GcpProjectIamMember

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpProjectIamMemberSpec defines a single ADDITIVE project-level IAM grant:
one role, to one member, on one project — the safe, composable unit of GCP
access control.

Additive semantics: this grant merges into the project's IAM policy without
touching any other member's bindings on the same role, and removal subtracts
only this exact (role, member) pair. Multiple grants — from different charts,
teams, or tools — never fight over the policy. (Authoritative binding/policy
management, which clobbers everything not listed, is deliberately not modeled.)

Every field is immutable, mirroring the underlying API: an IAM grant has no
update — changing role, member, or condition replaces the grant atomically.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpProjectIamMember
metadata:
  name: my-sample-iam-grant
spec:
  # GCP project whose IAM policy receives this grant.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # The role to grant: a predefined role, or a custom role's fully-qualified
  # name (reference a GcpIamCustomRole — its `name` output is exactly this value)
  role:
    value: roles/storage.objectViewer

  # The identity receiving the grant, in GCP IAM member format.
  # Reference a GcpServiceAccount — its `member` output is exactly this value.
  member:
    value: serviceAccount:my-sample-sa@my-gcp-project-123.iam.gserviceaccount.com

  # Optional IAM Condition restricting when the grant applies
  # condition:
  #   title: expires-2027
  #   expression: request.time < timestamp("2027-01-01T00:00:00Z")
  #   description: Temporary access for the migration
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.role` | `string \| valueFrom` | yes |  | GcpIamCustomRole (`status.outputs.name`) |
| `spec.member` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.member`) |
| `spec.condition` | `GcpProjectIamMemberCondition` |  |  |  |
| `spec.condition.title` | `string` | yes |  |  |
| `spec.condition.expression` | `string` | yes |  |  |
| `spec.condition.description` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project whose IAM policy receives this grant.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.role

`string | valueFrom` · required

The role to grant. Either a predefined role ("roles/storage.objectViewer")
or a custom role's fully-qualified name ("projects/<project>/roles/<role_id>").
Reference a GcpIamCustomRole resource to grant a custom role — its `name`
output is exactly this value.

- references: GcpIamCustomRole (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpIamCustomRole, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.member

`string | valueFrom` · required

The identity receiving the grant, in GCP IAM member format:
  serviceAccount:<email>       — a service account (the most common in IaC;
                                 reference a GcpServiceAccount resource —
                                 its `member` output is exactly this value)
  user:<email>                 — a Google account
  group:<email>                — a Google group
  domain:<domain>              — everyone in a Workspace/Cloud Identity domain
  principal:// / principalSet:// — workload identity federation principals
  allUsers / allAuthenticatedUsers — public access (grant with extreme care)
Grants to deleted principals ("deleted:...") are not supported.
Format validation happens at deploy time in the modules rather than here,
because the value usually arrives through a reference that is only
resolved at deploy time.

- references: GcpServiceAccount (`status.outputs.member`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.member}} -- a bare string does not parse

### spec.condition

`GcpProjectIamMemberCondition`

Optional IAM Condition restricting when this grant applies (e.g. only on
resources with a given prefix, or before an expiry date). The condition is
part of the grant's identity: the same role with and without a condition
are two independent grants that do not interfere.

### spec.condition.title

`string` · required

Short human-readable title identifying the condition's intent,
e.g. "expires-2026-12-31" or "prod-buckets-only".

- rule: {"required":true,"string":{"maxLen":"100"}}

### spec.condition.expression

`string` · required

The CEL condition expression, e.g.
request.time < timestamp("2027-01-01T00:00:00Z") or
resource.name.startsWith("projects/_/buckets/prod-").

- rule: {"required":true}

### spec.condition.description

`string`

Optional longer explanation of what the condition does and why it exists.

- rule: {"string":{"maxLen":"256"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpProjectIamMember, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.project_id` | `string` | The project ID whose IAM policy received the grant (after reference resolution and any provider-default fallback). |
| `status.outputs.role` | `string` | The role that was granted (predefined name or fully-qualified custom role name), after reference resolution. |
| `status.outputs.member` | `string` | The member the role was granted to, in IAM member format, after reference resolution. |
| `status.outputs.etag` | `string` | The etag of the project IAM policy after this grant was applied — a fingerprint of the policy version, useful for audit correlation. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.role` | GcpIamCustomRole | `status.outputs.name` |
| `spec.member` | GcpServiceAccount | `status.outputs.member` |

## See Also

- [Overview](../README.md)
