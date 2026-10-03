# GCP SCC BigQuery Export

Keeps a live copy of your Security Command Center findings in BigQuery, for a project, a folder, or the whole organization. New and changed findings land in the dataset within minutes, ready for SQL dashboards, trend reports, and joins with asset inventories.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `securitycenter.googleapis.com` on a project export's project
- **BigQuery export** -- the scope's `securitycenter.V2*SccBigQueryExport`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Security Command Center BigQuery-export admin permissions at the scope. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP SCC BigQuery Export**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Project Active Findings** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSccBigQueryExport
metadata:
  name: findings-history
  org: acme-corp
  env: prod
spec:
  bigQueryExportId: findings-history
  dataset:
    value: projects/security-analytics/datasets/scc_findings
  filter: state = "ACTIVE" AND NOT mute = "MUTED"
```

```shell
planton apply -f scc-bigquery-export.yaml
```

This keeps every active, unmuted finding in the project in BigQuery. An Infra Job tracks the provisioning in real time.

### InfraChart

Reference a `GcpBigQueryDataset` from `dataset` and grant the export's `status.outputs.principal` write access through the dataset's access entries; reference a `GcpFolder` from `scope.folderId` to cover a folder.

## Key Configuration

These are the most important decisions when configuring an export. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Dataset** -- where the findings table lives; the export's principal needs write access there.

**Filter** -- which findings are exported; empty exports everything.

**Scope** -- a project, a folder, or the organization.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `scope.projectId` | `status.outputs.project_id` |
| **GcpFolder** | `scope.folderId` | `status.outputs.folder_id` |
| **GcpBigQueryDataset** | `dataset` | `status.outputs.self_link` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `principal` | The writer Security Command Center uses | The dataset's access entry |
| `name` | The export's resource name | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Project active findings** -- active, unmuted findings of one project. Start from the **Project Active Findings** preset.

**Organization findings history** -- every finding in the organization into the security data lake. Start from the **Organization Findings History** preset.

## Works With

- [**GCP BigQuery Dataset**](/infra-catalog/gcp-bigquery-dataset) -- the destination
- [**GCP SCC Notification Config**](/infra-catalog/gcp-scc-notification-config) -- real-time findings to Pub/Sub
- [**GCP SCC Mute Config**](/infra-catalog/gcp-scc-mute-config) -- keep accepted findings out
