# GCP BigQuery Connection

A BigQuery connection -- the credential-carrying link BigQuery uses to reach data outside its own storage. Exactly one arm: `cloudSql` or `cloudSpanner` for federated queries, `cloudResource` for a Google-managed service account that reads Cloud Storage (BigLake), calls Vertex AI (remote models), or invokes Cloud Run (remote functions), `aws` or `azure` for BigQuery Omni, `configuration` for the BigQuery Connector framework (AlloyDB and more), or `spark` for stored procedures in Apache Spark.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `bigqueryconnection.googleapis.com` on the project (never disabled on destroy)
- **Connection** -- a `bigquery_connection` with its one arm

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with BigQuery connection admin permissions (`roles/bigquery.connectionAdmin`) on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpCloudSql`** -- the instance a `cloudSql` connection queries (`cloudSql.instanceId`, its `connection_name`).
- **`GcpKmsKey`** -- a key encrypting the stored credential (`kmsKeyName`).
- **The resources the connection's identity reads** -- grant the exported service account or identity access after creation.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryConnection
metadata:
  name: lake
spec:
  location: US
  cloudResource: true
```

```shell
planton apply -f bigquery-connection.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| one arm | -- | Exactly one of `aws`, `azure`, `cloudResource`, `cloudSpanner`, `cloudSql`, `configuration`, `spark`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project. |
| `location` | `string` | Google's default | `US`, `EU`, a region, `aws-*`, or `azure-*` -- per arm. Immutable. |
| `connectionId` | `string` | `metadata.name` | Immutable. |
| `friendlyName` / `description` | `string` | none | Console text. |
| `kmsKeyName` | `StringValueOrRef` | Google-managed | CMEK for the stored credential. |
| `cloudSql` | `object` | -- | `instanceId` (`GcpCloudSql` reference), `database`, `type` (`POSTGRES`/`MYSQL`), `credential.username`/`password`. |
| `cloudSpanner` | `object` | -- | `database` (`project/instance/database`), `databaseRole`, `useParallelism`, `useDataBoost`, `maxParallelism`. |
| `cloudResource` | `bool` | `false` | Declares the Google-managed service account arm. |
| `aws` | `object` | -- | `iamRoleId` -- the IAM role BigQuery Omni assumes. |
| `azure` | `object` | -- | `customerTenantId`, `federatedApplicationClientId`. |
| `configuration` | `object` | -- | `connectorId`, `asset`, `usernamePassword`, `hostPort`, `networkAttachment`. |
| `spark` | `object` | -- | `metastoreService`, `historyServerDataprocCluster`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- Exactly one arm is set.
- Spanner: `useDataBoost` requires `useParallelism`; `maxParallelism` requires both; `databaseRole` starts with a letter.
- Cloud SQL `type` is `POSTGRES` or `MYSQL`; resource paths (Spanner database, network attachment, Metastore service, Dataproc cluster) take their documented form.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/connections/{connection_id}` |
| `connection_id` | `string` | What tables and routines name as `{project}.{location}.{connection_id}` |
| `location` | `string` | The connection's location |
| `cloud_resource_service_account_id` | `string` | The `cloudResource` service account to grant access to |
| `spark_service_account_id` | `string` | The `spark` service account |
| `cloud_sql_service_account_id` | `string` | The `cloudSql` service account |
| `connector_service_account` | `string` | The `configuration` connector's service account |
| `aws_identity` | `string` | The identity your AWS role trusts |
| `azure_identity` / `azure_application` / `azure_client_id` / `azure_object_id` / `azure_redirect_uri` | `string` | The Azure setup values |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **The identity is the point.** Most arms create a Google-owned identity; the connection works only after you grant it access -- storage.objectViewer on a bucket, Vertex AI user for remote models, the AWS role's trust policy.
- **Location rules differ by arm.** Cloud SQL must match (with us-central1 pairing with `US` and europe-west1 with `EU`), Spanner matches its instance region, Omni uses `aws-{region}` or `azure-{region}`.
- **Passwords are managed secrets.** `cloudSql.credential.password` and `configuration.usernamePassword.password` are sensitive; BigQuery stores them and never returns them.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpCloudSql** -- a federated-query source
- **GcpBigQueryDataset** -- where external tables and remote models live
- **GcpKmsKey** -- CMEK for the credential

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
