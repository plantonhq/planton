# GCP BigQuery Connection

Lets BigQuery reach data it does not store: query Cloud SQL and Spanner in place, read Cloud Storage through BigLake, call Vertex AI models and Cloud Run functions from SQL, query AWS and Azure data with BigQuery Omni, or run Spark procedures -- each through a connection whose Google-owned identity you grant exactly the access it needs.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `bigqueryconnection.googleapis.com` on the project (never disabled on destroy)
- **Connection** -- a `bigquery_connection` with its one arm

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with BigQuery connection admin permissions (`roles/bigquery.connectionAdmin`) on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpCloudSql`** -- the instance a `cloudSql` connection queries (`cloudSql.instanceId`, its `connection_name`).
- **`GcpKmsKey`** -- a key encrypting the stored credential (`kmsKeyName`).
- **The resources the connection's identity reads** -- grant the exported service account or identity access after creation.

## Deploy

### Console

Open the deployment store, find **GCP BigQuery Connection**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **BigLake and Remote Models** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryConnection
metadata:
  name: lake
  org: acme-corp
  env: prod
spec:
  location: US
  cloudResource: true
```

```shell
planton apply -f bigquery-connection.yaml
```

This creates a cloud-resource connection in the US multi-region and outputs the service account BigQuery acts as through it. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a `GcpCloudSql` from `cloudSql.instanceId` for federated queries; grant the exported service account (for example with a bucket IAM member) in the same chart.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The arm** -- what the connection reaches decides everything else -- pick exactly one.

**The identity** -- `cloudResource`, `spark`, `cloudSql`, `aws`, and `azure` produce identities Google owns; grant them access to the data, never the other way round.

**Location** -- a dataset uses only connections in its own location; choose `US`/`EU` or the region that matches the source.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpCloudSql** | `cloudSql.instanceId` | `status.outputs.connection_name` |
| **GcpKmsKey** | `kmsKeyName` | `status.outputs.key_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `connection_id` | The connection's id | External tables, remote models, and routines |
| `cloud_resource_service_account_id` | The service account BigQuery acts as | Bucket and Vertex AI IAM grants |
| `aws_identity` | Google's identity for Omni on AWS | The IAM role's trust policy |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**BigLake and remote models** -- A cloud-resource connection whose service account reads Cloud Storage and calls Vertex AI. Start from the **BigLake and Remote Models** preset.

**Federated queries** -- EXTERNAL_QUERY against a Cloud SQL for PostgreSQL database. Start from the **Cloud SQL Federation** preset.

**Spanner analytics** -- Federated Spanner reads on independent compute. Start from the **Spanner with Data Boost** preset.

## Works With

- [**GCP Cloud SQL**](/cloud-catalog/gcp-cloud-sql) -- federated-query source
- [**GCP BigQuery Dataset**](/cloud-catalog/gcp-bigquery-dataset) -- where external tables live
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- credential encryption
