# GcpIamCustomRole

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpIamCustomRoleSpec defines a project-scoped IAM custom role — a named,
reusable bundle of permissions tailored to exactly what a workload or team
needs, instead of granting one of Google's broad predefined roles.

Custom roles are the least-privilege building block: define the role once,
then grant it to identities with GcpProjectIamMember (which references this
role's `name` output). Keeping the role a first-class node means permission
changes happen in one place and every grant picks them up.

Lifecycle note: GCP soft-deletes custom roles. A deleted role stays visible
(and its ID reserved) for up to 14 days before permanent deletion; creating
a role with the ID of a soft-deleted role undeletes and updates it in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpIamCustomRole
metadata:
  name: my-sample-custom-role
spec:
  # Role identifier (3-64 chars; letters, digits, underscores, periods; NO hyphens).
  # Forms the full role name: projects/<project>/roles/<role_id>
  roleId: logBucketWriter

  # GCP project that owns this custom role.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Title shown in the GCP console and IAM policy pickers
  title: Log Bucket Writer

  # What this role is for and who should hold it
  description: Grants exactly the permissions needed to write objects into log buckets

  # The permission bundle (at least one required).
  # Discover valid values with: gcloud iam list-testable-permissions <resource>
  permissions:
    - storage.objects.create
    - storage.objects.get

  # Launch stage label (default GA)
  stage: GA
  # Destroy really destroys in E2E: the live lanes prove the full lifecycle.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.roleId` | `string` | yes |  |  |
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.title` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.permissions` | `[]string` | yes |  |  |
| `spec.stage` | `string` |  | `GA` |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.roleId

`string` · required

The identifier for the role, unique within the project. Forms the full
role name: projects/<project>/roles/<role_id>.
3-64 characters; letters, digits, underscores, and periods only —
hyphens are NOT allowed. Convention is camelCase (e.g. "logBucketWriter").
Immutable: changing it destroys and recreates the role, which breaks every
grant referencing the old role name.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9_.]{3,64}$"}}

### spec.projectId

`string | valueFrom`

The GCP project that owns this custom role. The role can only be granted
on resources within this project.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the role.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.title

`string` · required

Human-readable title shown in the GCP console and IAM policy pickers
(max 100 characters). Mutable.

- rule: {"required":true,"string":{"maxLen":"100"}}

### spec.description

`string`

Human-readable description of what this role is for and who should hold it.
Surfaces in the console next to the title — write it for the operator
auditing IAM policies later. Mutable.

- rule: {"string":{"maxLen":"256"}}

### spec.permissions

`[]string` · required

The IAM permissions this role grants, e.g. ["storage.objects.get",
"storage.objects.list"]. At least one is required. Mutable: adding or
removing permissions updates the role in place, and every existing grant
of the role immediately reflects the change — that is the point of
defining the bundle once.
Permission strings follow <service>.<resource>.<verb>; discover valid
values with `gcloud iam list-testable-permissions <resource>`.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.stage

`string` · optional (explicit presence)

The launch stage of the role, mirroring GCP's role lifecycle labels.
Purely informational — it does not change what the role can do, but
GA (the default) is right for production roles; use DISABLED to keep the
role defined while rejecting all of its grants (an IAM kill switch).

- default: `GA`
- rule: stage must be one of ALPHA, BETA, GA, DEPRECATED, DISABLED, or EAP

### spec.deletionPolicy

`string`

What destroying this resource does to the role. Custom-role deletion
is a SOFT delete in GCP — the role enters a deleted state (recoverable
by undelete for 7 days, purged after 37), so DELETE and ABANDON differ
less here than on hard-delete resources:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the role is soft-deleted; existing bindings to it stop
               granting access
  "PREVENT" -- destroy FAILS; protects a role that live bindings
               depend on
  "ABANDON" -- the role is removed from management but stays active,
               and every binding to it keeps working

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpIamCustomRole, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | The fully-qualified role name: projects/<project>/roles/<role_id>. This is the grantable handle — feed it directly into IAM grants (GcpProjectIamMember's role field) exactly as IAM policies expect it. |
| `status.outputs.role_id` | `string` | The bare role ID within the project (the spec's role_id, echoed for convenience when composing tooling that addresses the role by short ID). |
| `status.outputs.deleted` | `bool` | Whether the role is currently soft-deleted. GCP retains deleted custom roles for up to 14 days; a true value means the role exists but rejects all grants until undeleted or permanently purged. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpGcsBucketIamMember | `spec.role` | `status.outputs.name` |
| GcpKmsKeyIamMember | `spec.role` | `status.outputs.name` |
| GcpProjectIamMember | `spec.role` | `status.outputs.name` |
| GcpPubSubTopicIamMember | `spec.role` | `status.outputs.name` |
| GcpServiceAccountIamMember | `spec.role` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
