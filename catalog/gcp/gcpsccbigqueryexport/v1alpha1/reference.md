# GcpSccBigQueryExport

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpSccBigQueryExportSpec continuously exports Security Command Center
findings to a BigQuery dataset, at a project, folder, or organization
(`google_scc_v2_project_scc_big_query_export`,
`google_scc_v2_folder_scc_big_query_export`,
`google_scc_v2_organization_scc_big_query_export` -- the scope picks one).

The dataset becomes the findings history: dashboards, trend reports, and
joins with asset inventories query it with SQL. New and updated findings
arrive within minutes; Security Command Center creates and manages the
findings table itself.

Two things must be true before rows arrive:
  - Security Command Center is activated on the scope.
  - The principal output can create tables and write data in the dataset
    (roles/bigquery.dataEditor through the dataset's access entries).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSccBigQueryExport
metadata:
  name: findings-history
spec:
  scope:
    projectId:
      value: my-gcp-project
  bigQueryExportId: findings-history
  dataset:
    value: projects/my-gcp-project/datasets/scc_findings
  filter: state = "ACTIVE" AND NOT mute = "MUTED"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.scope` | `GcpSccBigQueryExportScope` |  |  |  |
| `spec.scope.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.scope.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |
| `spec.scope.organizationId` | `string` |  |  |  |
| `spec.bigQueryExportId` | `string` | yes |  |  |
| `spec.dataset` | `string \| valueFrom` | yes |  | GcpBigQueryDataset (`status.outputs.self_link`) |
| `spec.filter` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.location` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.scope

`GcpSccBigQueryExportScope`

Whose findings are exported. Omit for the provider's default project.

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

### spec.bigQueryExportId

`string` · required

The export's ID, unique within its parent: lowercase letters, digits,
and hyphens, starting with a letter and ending with a letter or digit,
at most 63 characters. Immutable.

- rule: big_query_export_id must be 1-63 lowercase letters, digits, or hyphens, starting with a letter and ending with a letter or digit
- rule: {"required":true}

### spec.dataset

`string | valueFrom` · required

The dataset findings are written to: a literal
projects/{project}/datasets/{dataset} or a GcpBigQueryDataset
reference (the module trims the dataset's self link to Google's form).
Dataset IDs use letters, digits, and underscores only. Required: an
export without a dataset writes nothing.

- references: GcpBigQueryDataset (`status.outputs.self_link`)
- rule: dataset must be projects/{project}/datasets/{dataset} or a BigQuery dataset self link
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBigQueryDataset, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.filter

`string`

Which finding create and update events are exported, in the same
syntax as notification filters:
  state = "ACTIVE" AND NOT mute = "MUTED"
Empty exports every finding.

### spec.description

`string`

What the export is for, up to 1024 characters.

- rule: {"string":{"maxLen":"1024"}}

### spec.location

`string`

Where the export configuration is stored. "global", the default,
unless Security Command Center data residency was set up at activation
(then the residency location, e.g. "eu" or "us").

- rule: location must be global or a residency location such as eu or us

### spec.deletionPolicy

`string`

What destroying this block does:
  "" / "DELETE" -- the export is deleted; the dataset and the rows
                   already written stay
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the export leaves management and keeps writing

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpSccBigQueryExport, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: {parent}/locations/{location}/bigQueryExports/{big_query_export_id}. |
| `status.outputs.principal` | `string` | The Security Command Center service account that writes the findings. Grant it roles/bigquery.dataEditor on the dataset (a GcpBigQueryDataset access entry), or no rows arrive. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.scope.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.scope.folderId` | GcpFolder | `status.outputs.folder_id` |
| `spec.dataset` | GcpBigQueryDataset | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)
