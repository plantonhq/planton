# GcpFolder

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpFolderSpec creates one Resource Manager folder: a node in the Google
Cloud resource hierarchy that groups projects (and other folders) so
that IAM policy, organization policies, and billing views can be applied
to the whole group at once. A folder lives directly under the
organization or inside another folder, up to ten levels deep; a project
lives inside exactly one folder (or directly under the organization).

The hierarchy is the whole point of this kind. A landing zone is folders
inside folders -- `environments/production`, `teams/payments` -- each
carrying the guardrails (GcpOrgPolicy) and grants (IAM) every project
created beneath it inherits. Declare the parent folder as a reference to
another GcpFolder and the chart builds the tree in dependency order;
place a project inside it with GcpProject's `folder_id` reference.

What is deliberately NOT here: IAM grants on the folder (a
per-principal concern that belongs on its own additive resource, as
GcpProjectIamMember is for projects) and the folder's organization
policies (GcpOrgPolicy scoped to this folder by reference).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFolder
metadata:
  name: production
spec:
  # Where the folder sits: exactly one of organizationId (a top-level folder)
  # or folderId (nested; a literal or a reference to another GcpFolder's
  # folder_id output). Changing the parent MOVES the folder in place.
  parent:
    organizationId: "123456789012"

  # Console name; defaults to metadata.name. 3-30 characters, unique among
  # siblings, reserved for 30 days after deletion.
  displayName: production

  # Client-side destroy guard (Google's default is true). A destroy fails
  # until this is false and applied first. ABANDON bypasses it.
  deletionProtection: true

  # Create-time tags (tagKeys/{id} -> tagValues/{id}). IMMUTABLE -- changing
  # them recreates the folder. Prefer GcpTagBinding after creation.
  # tags:
  #   tagKeys/281475647562788: tagValues/281476102962987

  # DELETE (default; soft-delete, 30-day recovery, refused while non-empty),
  # PREVENT (destroy fails), ABANDON (unmanaged, stays live).
  deletionPolicy: PREVENT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.parent` | `GcpFolderParent` | yes |  |  |
| `spec.parent.organizationId` | `string` |  |  |  |
| `spec.parent.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |
| `spec.displayName` | `string` |  |  |  |
| `spec.deletionProtection` | `bool` |  | `true` |  |
| `spec.tags` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.parent

`GcpFolderParent` · required

Where the folder sits in the hierarchy: directly under the
organization or inside another folder. Exactly one arm. Changing the
parent MOVES the folder in place (Google's folders.move) -- projects and
sub-folders travel with it and nothing is recreated -- but every IAM
grant and organization policy the folder inherited from its old parent
stops applying and the new parent's start applying at once.

- rule: {"required":true}
- rule: set exactly one of organization_id or folder_id -- a folder lives directly under the organization or inside one other folder

### spec.parent.organizationId

`string`

A top-level folder: the numeric organization ID (from `gcloud
organizations list`), without the `organizations/` prefix.

- rule: organization_id must be the numeric organization ID, without the organizations/ prefix

### spec.parent.folderId

`string | valueFrom`

A nested folder: the parent folder's numeric ID -- a literal, or a
reference to another GcpFolder resource (its folder_id output). The
reference is how a chart builds a hierarchy: the child waits for the
parent to exist and nests inside it.

- references: GcpFolder (`status.outputs.folder_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.folder_id}} -- a bare string does not parse

### spec.displayName

`string`

The name shown in the console and in `gcloud resource-manager folders
list`. Defaults to metadata.name when empty. Google's rules: 3-30
characters; letters, digits, spaces, hyphens, and underscores; must
start and end with a letter or digit; and UNIQUE among the parent's
direct children (two sibling folders cannot share a display name).
Mutable: a rename is an in-place update.

A deleted folder keeps its display name reserved for the 30-day
soft-delete window (folders are recoverable via undelete in that
window), so a fresh folder with the same name under the same parent
fails until the window closes or the old one is purged.

- rule: display_name must be 3-30 characters of letters, digits, spaces, hyphens, or underscores, starting and ending with a letter or digit -- e.g. production or team-payments

### spec.deletionProtection

`bool` · optional (explicit presence)

Client-side destroy guard. While true (Google's provider default), any
destroy -- including the platform's own teardown flows -- FAILS until
this field is set to false AND applied first; only then does a second
destroy delete the folder. Both engines always send the value
explicitly, so the spec is the single source of truth. The guard exists
because deleting a folder is a hierarchy-wide act: Google refuses to
delete a folder that still holds projects or folders, but a folder that
was just emptied is one accidental destroy away from a 30-day recovery
exercise.

Ordering quirk in the provider: deletion_policy ABANDON is evaluated
BEFORE this guard, so abandoning a protected folder still works;
PREVENT is evaluated before both.

- default: `true`

### spec.tags

`map<string, string>`

Resource Manager tags bound to the folder at CREATE TIME only, as
`tagKeys/{numeric_id}` -> `tagValues/{numeric_id}` (the `name` outputs
of GcpTagKey and GcpTagValue). Tags are what organization policies
(`resource.matchTag`), IAM conditions, and firewall policies key on.

Changing this map after creation RECREATES the folder, which for a
folder holding projects is not a move but a failure (Google refuses to
delete a non-empty folder). Use it only when the tag must exist at
create time -- typically so an organization policy conditioned on the
tag governs the folder from its first second. For every other case,
bind tags after creation with GcpTagBinding, which attaches and
detaches without touching the folder.

- rule: {"map":{"keys":{"string":{"pattern":"^tagKeys/[0-9]+$"}},"values":{"string":{"pattern":"^tagValues/[0-9]+$"}}}}

### spec.deletionPolicy

`string`

What destroying this resource does to the folder in GCP:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the folder is deleted (Google soft-deletes it for 30
               days, during which it can be undeleted and its display
               name stays reserved); fails while deletion_protection is
               true or the folder still holds projects or folders
  "PREVENT" -- destroy FAILS; the guard for the folders a landing zone
               is built on
  "ABANDON" -- the folder is removed from management but keeps existing
               in GCP with everything inside it; bypasses
               deletion_protection

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpFolder, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.folder_id` | `string` | The folder's numeric ID -- the value every child needs: a nested GcpFolder's parent.folder_id, a GcpProject's folder_id, a GcpOrgPolicy's scope.folder_id, a GcpTagBinding's parent.folder_id. |
| `status.outputs.name` | `string` | The folder's resource name, `folders/{folder_id}` -- the form Google's APIs and IAM address the folder by. |
| `status.outputs.lifecycle_state` | `string` | The folder's lifecycle state as Google reports it: ACTIVE for a live folder, DELETE_REQUESTED during the 30-day soft-delete window. |
| `status.outputs.create_time` | `string` | When the folder was created (RFC 3339 UTC). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.parent.folderId` | GcpFolder | `status.outputs.folder_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpBillingBudget | `spec.budgetFilter.resourceAncestors` | `status.outputs.name` |
| GcpFolder | `spec.parent.folderId` | `status.outputs.folder_id` |
| GcpHierarchicalFirewallPolicy | `spec.parent.folderId` | `status.outputs.folder_id` |
| GcpHierarchicalFirewallPolicy | `spec.associations[].target.folderId` | `status.outputs.folder_id` |
| GcpModelArmorFloorSetting | `spec.scope.folderId` | `status.outputs.folder_id` |
| GcpOrgPolicy | `spec.scope.folderId` | `status.outputs.folder_id` |
| GcpProject | `spec.folderId` | `status.outputs.folder_id` |
| GcpTagBinding | `spec.parent.folderId` | `status.outputs.folder_id` |

## See Also

- [Overview](../README.md)
