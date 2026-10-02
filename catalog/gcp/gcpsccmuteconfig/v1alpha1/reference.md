# GcpSccMuteConfig

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpSccMuteConfigSpec is a Security Command Center mute rule at a project,
folder, or organization (`google_scc_v2_project_mute_config`,
`google_scc_v2_folder_mute_config`, `google_scc_v2_organization_mute_config`
-- the scope picks one).

A mute rule hides findings that match its filter -- accepted risks,
known-noisy detectors, test projects -- so triage, notifications that
filter on `mute`, and exports see only what needs action. Muted findings
are not deleted; they stay queryable.

Security Command Center must be activated on the scope.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSccMuteConfig
metadata:
  name: sandbox-public-buckets
spec:
  scope:
    projectId:
      value: my-gcp-project
  muteConfigId: sandbox-public-buckets
  filter: category = "PUBLIC_BUCKET_ACL"
  type: DYNAMIC
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.scope` | `GcpSccMuteConfigScope` |  |  |  |
| `spec.scope.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.scope.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |
| `spec.scope.organizationId` | `string` |  |  |  |
| `spec.muteConfigId` | `string` | yes |  |  |
| `spec.filter` | `string` | yes |  |  |
| `spec.type` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.location` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.scope

`GcpSccMuteConfigScope`

Whose findings the rule mutes. Omit for the provider's default project.

- rule: set at most one of project_id, folder_id, or organization_id (empty means the provider's default project)

### spec.scope.projectId

`string | valueFrom`

A project: a literal project ID or a GcpProject reference.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.scope.folderId

`string | valueFrom`

A folder: the folder's numeric ID -- a literal or a GcpFolder
reference. Covers every project beneath it.

- references: GcpFolder (`status.outputs.folder_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.folder_id}} -- a bare string does not parse

### spec.scope.organizationId

`string`

The organization: the numeric organization ID, without the
organizations/ prefix.

- rule: organization_id must be the numeric organization ID, without the organizations/ prefix

### spec.muteConfigId

`string` · required

The rule's ID, unique within its parent: lowercase letters, digits,
and hyphens, starting with a letter and ending with a letter or digit,
at most 63 characters. Immutable.

- rule: mute_config_id must be 1-63 lowercase letters, digits, or hyphens, starting with a letter and ending with a letter or digit
- rule: {"required":true}

### spec.filter

`string` · required

Which findings are muted. Supported fields: severity, category,
resource.name, resource.project_name, resource.project_display_name,
resource.folders.resource_folder, resource.parent_name,
resource.parent_display_name, resource.type, findingClass,
indicator.ip_addresses, indicator.domains -- each with = or : (substring),
combined with AND and OR. Write it for the scope: a filter naming
project X on a rule scoped to project Y matches nothing. Example:
  category = "PUBLIC_BUCKET_ACL" AND resource.project_display_name = "sandbox"

- rule: {"required":true}

### spec.type

`string` · required

How the rule mutes:
  "DYNAMIC" -- applies to existing and future matching findings, and
               stops applying when the rule changes or is deleted, or a
               finding stops matching (Google's recommendation)
  "STATIC"  -- sets a permanent mute on future matching findings only;
               later rule changes do not unmute them
Google treats the type as immutable after creation.

- rule: {"required":true,"string":{"in":["DYNAMIC","STATIC"]}}

### spec.description

`string`

Why the rule exists -- the accepted risk or the noise it removes.

### spec.location

`string`

Where the rule is stored. "global", the default, unless Security
Command Center data residency was set up at activation (then the
residency location, e.g. "eu" or "us").

- rule: location must be global or a residency location such as eu or us

### spec.deletionPolicy

`string`

What destroying this block does:
  "" / "DELETE" -- the rule is deleted (DYNAMIC mutes are lifted)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the rule leaves management and keeps muting

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpSccMuteConfig, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: {parent}/locations/{location}/muteConfigs/{mute_config_id}. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.scope.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.scope.folderId` | GcpFolder | `status.outputs.folder_id` |

## See Also

- [Overview](../README.md)
