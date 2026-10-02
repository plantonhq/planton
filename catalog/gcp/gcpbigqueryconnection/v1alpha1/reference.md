# GcpBigQueryConnection

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpBigQueryConnectionSpec defines a BigQuery connection
(`google_bigquery_connection`) -- the credential-carrying link BigQuery
uses to reach data outside its own storage: Cloud SQL and Spanner for
federated queries, Cloud Storage and other Google resources through a
Google-managed service account (BigLake tables, remote models and
functions on Vertex AI), AWS and Azure through BigQuery Omni, the
Connector framework for AlloyDB and friends, and Spark stored
procedures.

Set exactly ONE arm: aws, azure, cloud_resource, cloud_spanner,
cloud_sql, configuration, or spark.

Most arms create an identity Google owns that you must then grant
access to -- the cloud_resource and Spark service accounts, the AWS
identity, the Azure application -- all of them outputs.

Immutable: project_id, location, connection_id, the connector id (a
change replaces the connection). Descriptions, the key, and each arm's
settings update in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryConnection
metadata:
  name: orders-postgres
spec:
  projectId:
    value: my-gcp-project
  # Cloud SQL us-central1 pairs with BigQuery US.
  location: US
  friendlyName: Orders database
  description: Federated queries against the orders Postgres
  cloudSql:
    instanceId:
      value: my-gcp-project:us-central1:orders
    database: orders
    type: POSTGRES
    credential:
      username: bigquery_reader
      password: change-me-immediately # replace before applying
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` |  |  |  |
| `spec.connectionId` | `string` |  |  |  |
| `spec.friendlyName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.aws` | `GcpBigQueryConnectionAws` |  |  |  |
| `spec.aws.iamRoleId` | `string` | yes |  |  |
| `spec.azure` | `GcpBigQueryConnectionAzure` |  |  |  |
| `spec.azure.customerTenantId` | `string` | yes |  |  |
| `spec.azure.federatedApplicationClientId` | `string` |  |  |  |
| `spec.cloudResource` | `bool` |  |  |  |
| `spec.cloudSpanner` | `GcpBigQueryConnectionCloudSpanner` |  |  |  |
| `spec.cloudSpanner.database` | `string` | yes |  |  |
| `spec.cloudSpanner.databaseRole` | `string` |  |  |  |
| `spec.cloudSpanner.useParallelism` | `bool` |  |  |  |
| `spec.cloudSpanner.useDataBoost` | `bool` |  |  |  |
| `spec.cloudSpanner.maxParallelism` | `int32` |  |  |  |
| `spec.cloudSql` | `GcpBigQueryConnectionCloudSql` |  |  |  |
| `spec.cloudSql.instanceId` | `string \| valueFrom` | yes |  | GcpCloudSql (`status.outputs.connection_name`) |
| `spec.cloudSql.database` | `string` | yes |  |  |
| `spec.cloudSql.type` | `string` | yes |  |  |
| `spec.cloudSql.credential` | `GcpBigQueryConnectionCloudSqlCredential` | yes |  |  |
| `spec.cloudSql.credential.username` | `string` | yes |  |  |
| `spec.cloudSql.credential.password` | `string` (sensitive) | yes |  |  |
| `spec.configuration` | `GcpBigQueryConnectionConfiguration` |  |  |  |
| `spec.configuration.connectorId` | `string` | yes |  |  |
| `spec.configuration.asset` | `GcpBigQueryConnectionConnectorAsset` | yes |  |  |
| `spec.configuration.asset.database` | `string` |  |  |  |
| `spec.configuration.asset.googleCloudResource` | `string` |  |  |  |
| `spec.configuration.usernamePassword` | `GcpBigQueryConnectionUsernamePassword` |  |  |  |
| `spec.configuration.usernamePassword.username` | `string` | yes |  |  |
| `spec.configuration.usernamePassword.password` | `string` (sensitive) | yes |  |  |
| `spec.configuration.hostPort` | `string` |  |  |  |
| `spec.configuration.networkAttachment` | `string` |  |  |  |
| `spec.spark` | `GcpBigQueryConnectionSpark` |  |  |  |
| `spec.spark.metastoreService` | `string` |  |  |  |
| `spec.spark.historyServerDataprocCluster` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the connection lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string`

Where the connection lives, and so which datasets can use it: a
BigQuery multi-region (US, EU) or a region (us-central1,
europe-west1). Each arm has its own rule -- Cloud SQL must match (with
us-central1 -> US and europe-west1 -> EU allowed), Spanner matches the
instance's region, AWS uses aws-{region} (e.g. aws-us-east-1) and
Azure azure-{region} (e.g. azure-eastus2). Empty leaves Google's
default. Immutable.

### spec.connectionId

`string`

The connection's ID. Defaults to metadata.name. Immutable.

### spec.friendlyName

`string`

A human-readable name shown in the console.

### spec.description

`string`

A free-text description.

### spec.kmsKeyName

`string | valueFrom`

A Cloud KMS key encrypting the connection's stored credential (CMEK):
a GcpKmsKey reference or a literal
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}.
Empty uses Google-managed keys.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.aws

`GcpBigQueryConnectionAws`

AWS through BigQuery Omni.

### spec.aws.iamRoleId

`string` · required

The ARN of the AWS IAM role BigQuery assumes, e.g.
arn:aws:iam::123456789012:role/bigquery-omni. The role must trust the
Google-owned identity the connection outputs.

- rule: {"required":true}

### spec.azure

`GcpBigQueryConnectionAzure`

Azure through BigQuery Omni.

### spec.azure.customerTenantId

`string` · required

Your Azure AD (Entra ID) tenant ID -- the directory that holds the
data.

- rule: {"required":true}

### spec.azure.federatedApplicationClientId

`string`

The client ID of YOUR Azure application that hosts a federated
credential for the connection's Google identity (the recommended,
secretless setup). Empty makes Google create a multi-tenant Azure
application you grant consent to (see the azure_* outputs).

### spec.cloudResource

`bool`

A Google-managed service account BigQuery acts as when it reads Cloud
Storage (BigLake and object tables), calls Vertex AI (remote models),
or invokes Cloud Run functions (remote functions). Google's arm has no
settings: true declares it. Grant the cloud_resource_service_account_id
output access to what it reads.

### spec.cloudSpanner

`GcpBigQueryConnectionCloudSpanner`

A Cloud Spanner database.

### spec.cloudSpanner.database

`string` · required

The database, as project/instance/database (slashes, no
"projects/" prefix).

- rule: {"required":true,"string":{"pattern":"^[^/]+/[^/]+/[^/]+$"}}

### spec.cloudSpanner.databaseRole

`string`

A Spanner fine-grained access control role the queries run as (it
must start with a letter; letters, digits, underscores). Empty reads
with the caller's database-level permissions.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-zA-Z][a-zA-Z0-9_]*$"}}

### spec.cloudSpanner.useParallelism

`bool`

Read with Spanner parallelism. Required by use_data_boost and
max_parallelism.

### spec.cloudSpanner.useDataBoost

`bool`

Run the queries on Spanner Data Boost -- independent compute that
leaves the instance's serving capacity untouched (billed separately
by Spanner). Requires use_parallelism.

### spec.cloudSpanner.maxParallelism

`int32`

The most parallel reads per query on Data Boost. Requires both
use_parallelism and use_data_boost. Empty lets Spanner pick from the
instance configuration.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.cloudSql

`GcpBigQueryConnectionCloudSql`

A Cloud SQL database.

### spec.cloudSql.instanceId

`string | valueFrom` · required

The Cloud SQL instance, as project:region:instance -- a GcpCloudSql
reference (its connection_name output) or a literal in that form.

- references: GcpCloudSql (`status.outputs.connection_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudSql, name: <that resource's name>, fieldPath: status.outputs.connection_name}} -- a bare string does not parse

### spec.cloudSql.database

`string` · required

The database name inside the instance.

- rule: {"required":true}

### spec.cloudSql.type

`string` · required

The engine: POSTGRES or MYSQL.

- rule: {"required":true,"string":{"in":["POSTGRES","MYSQL"]}}

### spec.cloudSql.credential

`GcpBigQueryConnectionCloudSqlCredential` · required

The database user and password.

- rule: {"required":true}

### spec.cloudSql.credential.username

`string` · required

The database username.

- rule: {"required":true}

### spec.cloudSql.credential.password

`string` · required · sensitive

The database user's password. Stored by BigQuery and never returned.

- rule: {"required":true}

### spec.configuration

`GcpBigQueryConnectionConfiguration`

The BigQuery Connector framework.

### spec.configuration.connectorId

`string` · required

The connector, e.g. google-alloydb, google-cloudsql-mysql,
google-cloudsql-postgres. Immutable.

- rule: {"required":true}

### spec.configuration.asset

`GcpBigQueryConnectionConnectorAsset` · required

The data source the connector reads.

- rule: {"required":true}

### spec.configuration.asset.database

`string`

The database name.

### spec.configuration.asset.googleCloudResource

`string`

The full resource name of the Google Cloud resource -- for AlloyDB,
//alloydb.googleapis.com/projects/{project}/locations/{region}/clusters/{cluster}/instances/{instance}.

### spec.configuration.usernamePassword

`GcpBigQueryConnectionUsernamePassword`

Database-user authentication. The connection outputs the service
account the connector uses (connector_service_account).

### spec.configuration.usernamePassword.username

`string` · required

The database username.

- rule: {"required":true}

### spec.configuration.usernamePassword.password

`string` · required · sensitive

The database user's password, sent as Google's plaintext secret form
and stored by BigQuery; never returned.

- rule: {"required":true}

### spec.configuration.hostPort

`string`

The endpoint to connect to, as host:port, when the connector needs one
spelled out.

### spec.configuration.networkAttachment

`string`

A Private Service Connect network attachment the connector reaches a
private endpoint through, as
projects/{project}/regions/{region}/networkAttachments/{attachment}.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/regions/[^/]+/networkAttachments/[^/]+$"}}

### spec.spark

`GcpBigQueryConnectionSpark`

Stored procedures for Apache Spark.

### spec.spark.metastoreService

`string`

An existing Dataproc Metastore service the Spark code uses as its Hive
metastore, as projects/{project}/locations/{region}/services/{service}.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/locations/[^/]+/services/[^/]+$"}}

### spec.spark.historyServerDataprocCluster

`string`

An existing Dataproc cluster that serves as the Spark History Server
for the procedures' runs, as
projects/{project}/regions/{region}/clusters/{cluster}.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[^/]+/regions/[^/]+/clusters/[^/]+$"}}

### spec.deletionPolicy

`string`

What happens to the connection when this resource is destroyed:
  "" / "DELETE" -- deleted (tables and routines using it stop working)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- it leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.exactly_one_arm`: set exactly one of aws, azure, cloud_resource, cloud_spanner, cloud_sql, configuration, spark
- `spec.spanner_data_boost_needs_parallelism`: cloud_spanner.use_data_boost requires cloud_spanner.use_parallelism
- `spec.spanner_max_parallelism_needs_data_boost`: cloud_spanner.max_parallelism requires both cloud_spanner.use_data_boost and cloud_spanner.use_parallelism

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpBigQueryConnection, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/connections/{connection_id}. |
| `status.outputs.connection_id` | `string` | The connection's ID -- what a BigQuery table, routine, or remote model names as {project}.{location}.{connection_id}. |
| `status.outputs.location` | `string` | The connection's location. |
| `status.outputs.cloud_resource_service_account_id` | `string` | cloud_resource: the service account BigQuery acts as. Grant it access to the buckets, models, or functions it reaches. |
| `status.outputs.spark_service_account_id` | `string` | spark: the service account the Spark procedures run as. |
| `status.outputs.cloud_sql_service_account_id` | `string` | cloud_sql: the service account BigQuery connects to the instance as. |
| `status.outputs.connector_service_account` | `string` | configuration: the service account the connector authenticates with. |
| `status.outputs.aws_identity` | `string` | aws: the Google-owned identity your IAM role must trust. |
| `status.outputs.azure_identity` | `string` | azure: the Google-owned identity your federated credential trusts. |
| `status.outputs.azure_application` | `string` | azure: the Azure AD application Google created (when no federated application was given). |
| `status.outputs.azure_client_id` | `string` | azure: that application's client ID. |
| `status.outputs.azure_object_id` | `string` | azure: that application's object ID. |
| `status.outputs.azure_redirect_uri` | `string` | azure: the consent redirect URL for that application. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.cloudSql.instanceId` | GcpCloudSql | `status.outputs.connection_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDatastreamStream | `spec.destinationConfig.bigqueryDestinationConfig.blmtConfig.connectionName` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
