# GCP SCC BigQuery Export

Continuously exports Security Command Center findings to a BigQuery dataset, for a project, a folder, or the whole organization. The dataset becomes the findings history: dashboards, trend reports, and joins with asset inventories query it with SQL. Security Command Center creates and manages the findings table itself.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `securitycenter.googleapis.com` on a project export's project (never disabled on destroy)
- **BigQuery export** -- one `scc_v2_{project,folder,organization}_scc_big_query_export`, chosen by the scope

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Security Command Center BigQuery-export admin permissions (`roles/securitycenter.bigQueryExportsEditor`) at the scope, and permission to read the dataset.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Security Command Center

Security Command Center must be activated on the scope.

### Required Dependencies

- **`GcpBigQueryDataset`** -- the destination (`dataset`). Grant the `principal` output `roles/bigquery.dataEditor` on it (a dataset access entry), or no rows arrive.

### Optional Dependencies

- **`GcpProject`** / **`GcpFolder`** -- the scope, by reference.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSccBigQueryExport
metadata:
  name: findings-history
spec:
  bigQueryExportId: findings-history
  dataset:
    value: projects/my-gcp-project/datasets/scc_findings
  filter: state = "ACTIVE" AND NOT mute = "MUTED"
```

```shell
planton apply -f scc-bigquery-export.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `bigQueryExportId` | `string` | The export's ID: lowercase letters, digits, hyphens; starts with a letter; at most 63 characters. Immutable. |
| `dataset` | `string` / ref | `projects/{project}/datasets/{dataset}` or a `GcpBigQueryDataset` ref. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `scope` | `object` | provider project | At most one of `projectId` (`GcpProject` ref), `folderId` (`GcpFolder` ref), `organizationId` (numeric). |
| `filter` | `string` | every finding | Which create and update events are exported, e.g. `state = "ACTIVE" AND NOT mute = "MUTED"`. |
| `description` | `string` | none | Up to 1024 characters. |
| `location` | `string` | `global` | Where the configuration is stored; a residency location (`eu`, `us`) only if data residency was set up at activation. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` removes it, `PREVENT` fails destroy, `ABANDON` keeps it in Google. |

### Validation Rules

- `scope` names at most one of project, folder, organization; `organizationId` is numeric without the `organizations/` prefix.
- `location` is `global` or a residency location; `deletionPolicy` takes only Google's values.
- `bigQueryExportId` follows Google's ID rule; `dataset` names a dataset (IDs use letters, digits, and underscores).

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `{parent}/locations/{location}/bigQueryExports/{bigQueryExportId}` |
| `principal` | `string` | The writer Security Command Center uses -- grant it `roles/bigquery.dataEditor` on the dataset |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Grant the writer.** Without `roles/bigquery.dataEditor` for the `principal` output, the export writes nothing.
- **Destroy keeps the data.** Deleting the export stops new rows; the dataset and the rows already written stay.
- **The organization export's name is composed.** Google's organization resource takes its own name as an argument; both modules send Google's value so it never drifts.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpBigQueryDataset** -- the destination, and the access entry that grants the writer
- **GcpSccNotificationConfig** -- real-time findings to Pub/Sub
- **GcpSccMuteConfig** -- keep accepted findings out with `NOT mute = "MUTED"`

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
