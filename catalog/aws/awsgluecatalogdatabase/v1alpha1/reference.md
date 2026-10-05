# AwsGlueCatalogDatabase

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsGlueCatalogDatabaseSpec defines the desired configuration for an AWS Glue
Data Catalog database.

A Glue Data Catalog database is a metadata container that organizes table
definitions (schemas) for data stored in Amazon S3, Amazon Redshift, Amazon
RDS, and other data stores. It is the namespace layer that Amazon Athena, AWS
Glue ETL jobs, Glue Crawlers, Amazon Redshift Spectrum, and Amazon EMR use to
discover and query data.

Think of it as a "schema" in a traditional RDBMS but for a serverless data
lake: the database itself holds no data, only metadata about where data lives
and how it is structured.

Three creation shapes:
- A regular database (the common case): optionally set description,
  location_uri, parameters, and Lake Formation create-table permissions.
- A resource link (target_database): a pointer to a database shared from
  another account or region via AWS RAM / Lake Formation -- the linked
  database's tables become queryable locally.
- A federated database (federated_database): projects an external source
  (e.g. a Redshift datashare) into the catalog through a Glue connection.

Notes:
- The database name (from metadata.name) cannot be changed after creation
  (ForceNew). Naming constraints: 1-255 characters, and AWS rejects
  uppercase letters; use lowercase letters, numbers, and underscores.
- Credentials, region, and deployment workflow live outside this spec in
  IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsGlueCatalogDatabase
metadata:
  name: test_glue_db
  org: test-org
  env: dev
  id: test-glue-db-dev
spec:
  region: us-west-2
  description: Test Glue Data Catalog database exercising the regular-database surface
  locationUri: s3://test-data-lake/databases/test/
  parameters:
    classification: parquet
    team: data-platform
  createTableDefaultPermissions:
    - permissions:
        - SELECT
        - ALTER
      principal: arn:aws:iam::123456789012:role/data-lake-admin
    - permissions:
        - ALL
      principal: IAM_ALLOWED_PRINCIPALS
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.catalogId` | `string` |  |  |  |
| `spec.locationUri` | `string` |  |  |  |
| `spec.parameters` | `map<string, string>` |  |  |  |
| `spec.createTableDefaultPermissions` | `[]AwsGlueCatalogDatabasePrincipalPermissions` |  |  |  |
| `spec.createTableDefaultPermissions[].permissions` | `[]string` |  |  |  |
| `spec.createTableDefaultPermissions[].principal` | `string` |  |  |  |
| `spec.targetDatabase` | `AwsGlueCatalogDatabaseTarget` |  |  |  |
| `spec.targetDatabase.catalogId` | `string` | yes |  |  |
| `spec.targetDatabase.databaseName` | `string` | yes |  |  |
| `spec.targetDatabase.region` | `string` |  |  |  |
| `spec.federatedDatabase` | `AwsGlueCatalogDatabaseFederation` |  |  |  |
| `spec.federatedDatabase.identifier` | `string` |  |  |  |
| `spec.federatedDatabase.connectionName` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description of the database. Helps teams understand the
purpose and contents of this catalog namespace.

Maximum 2048 characters (enforced by the AWS API).

Examples:
- "Sales analytics data lake — raw and curated tables from the sales pipeline"
- "Clickstream events from web and mobile applications"

- rule: {"string":{"maxLen":"2048"}}

### spec.catalogId

`string`

ID of the Data Catalog in which to create the database. Defaults to the
AWS account ID of the deploying account; set it only to create the
database inside ANOTHER account's catalog (a cross-account governance
pattern that requires a matching catalog resource policy on that
account). Fixed at creation time (changing it replaces the database).

### spec.locationUri

`string`

Default S3 URI for tables created in this database. When a Glue Crawler or
CREATE TABLE statement does not specify a location, this path is used as the
base directory.

Format: "s3://bucket-name/optional-prefix/"

When omitted, each table must specify its own storage location explicitly.
Setting this is recommended for organized data lakes where all tables in a
database share a common S3 prefix.

One-way in practice: once applied, REMOVING this field does not clear the
location at AWS -- the provider keeps the last-known value. Point it at a
new prefix to change it.

This is a plain string (not StringValueOrRef) because it is an S3 URI with
a user-defined path prefix, not a direct resource identifier.

### spec.parameters

`map<string, string>`

Free-form key-value properties attached to the database. Consumed by
engines and governance tooling that read catalog metadata -- for example
classification hints, team ownership labels, or engine-specific switches.
These are catalog metadata, NOT AWS resource tags (identity tags derive
from metadata.name/org/env automatically).

### spec.createTableDefaultPermissions

`[]AwsGlueCatalogDatabasePrincipalPermissions`

Default Lake Formation permissions granted on tables CREATED in this
database. When omitted, AWS applies its default grant -- ALL permissions
to the virtual group IAM_ALLOWED_PRINCIPALS -- which keeps plain
IAM-policy access working (the compatibility mode most accounts run in).

Lake Formation-governed data lakes typically override this: grant to
specific principals, or supply an entry with an empty permission list to
stop granting IAM_ALLOWED_PRINCIPALS on new tables entirely (the
recommended hardening step when migrating to Lake Formation permissions).

### spec.createTableDefaultPermissions[].permissions

`[]string`

Lake Formation permissions to grant. Valid values: "ALL", "SELECT",
"ALTER", "DROP", "DELETE", "INSERT", "CREATE_DATABASE", "CREATE_TABLE",
"DATA_LOCATION_ACCESS". An entry with an EMPTY list (and no principal)
is meaningful: it disables the default IAM_ALLOWED_PRINCIPALS grant on
newly created tables.

- rule: {"repeated":{"items":{"string":{"in":["ALL","SELECT","ALTER","DROP","DELETE","INSERT","CREATE_DATABASE","CREATE_TABLE","DATA_LOCATION_ACCESS"]}}}}

### spec.createTableDefaultPermissions[].principal

`string`

The principal receiving the grant: an IAM user/role ARN, an AWS account
ID, or the virtual group "IAM_ALLOWED_PRINCIPALS" (grants to any
principal whose IAM policy allows the action -- Lake Formation's
IAM-compatibility mode). 1-255 characters when set.

- rule: {"string":{"maxLen":"255"}}

### spec.targetDatabase

`AwsGlueCatalogDatabaseTarget`

Creates this database as a RESOURCE LINK: a local pointer to a database
that lives in another account or region and was shared via AWS RAM /
Lake Formation. Queries against the link (from Athena, Redshift Spectrum,
EMR) resolve to the target database's tables, subject to the permissions
granted on the share.

A resource link carries no storage or schema of its own, so combining it
with location_uri or create-table permissions has no effect; leave those
unset. Fixed at creation time (changing it replaces the database).

### spec.targetDatabase.catalogId

`string` · required

ID of the Data Catalog that owns the target database -- the sharing
account's AWS account ID.

- rule: {"string":{"minLen":"1"}}

### spec.targetDatabase.databaseName

`string` · required

Name of the target database inside the owning catalog.

- rule: {"string":{"minLen":"1"}}

### spec.targetDatabase.region

`string`

Region of the target database, e.g. "us-east-1". Set for cross-REGION
links; omit when the target lives in the same region as this database.

### spec.federatedDatabase

`AwsGlueCatalogDatabaseFederation`

Creates this database as a FEDERATED database: a projection of an
external data source into the catalog through a Glue connection. The
primary use today is querying Amazon Redshift datashares from Athena and
other catalog consumers without copying data.

### spec.federatedDatabase.identifier

`string`

Unique identifier of the federated source, e.g. the Redshift datashare
ARN being projected into the catalog.

### spec.federatedDatabase.connectionName

`string`

Name of the Glue connection that carries the federation, e.g.
"aws:redshift" for the AWS-managed Redshift federation connection.

## Validation Rules

- `target_xor_federated`: target_database (resource link) and federated_database cannot be combined; a database is exactly one shape

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsGlueCatalogDatabase, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.database_name` | `string` | The name of the Glue Data Catalog database. Used in Athena queries (FROM database.table), Glue crawler configurations, Glue ETL job scripts, and Redshift Spectrum external schema definitions. The database name is unique within a catalog (AWS account + region). |
| `status.outputs.database_arn` | `string` | The Amazon Resource Name (ARN) of the Glue Data Catalog database. Used for IAM policies, Lake Formation permissions, and cross-service authorization. Format: arn:aws:glue:{region}:{account-id}:database/{database-name} |
| `status.outputs.catalog_id` | `string` | The ID of the AWS Glue Data Catalog that contains this database. In most cases this is the AWS Account ID. Useful for downstream resources that require the catalog context (e.g., Glue crawlers, cross-account references). |

## See Also

- [Overview](../README.md)
