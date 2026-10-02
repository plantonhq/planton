# GcpBigQueryConnection Guide

The judgment this guide protects: a connection is an identity, not a pipe. What it can reach is exactly what you grant the identity it creates -- so create the connection, then grant, and never hand BigQuery a broader credential than the one query or model needs.

## Choosing the arm

- **`cloudResource`** -- a Google-managed service account BigQuery acts as for BigLake and object tables over Cloud Storage, remote models on Vertex AI, and remote functions on Cloud Run. It has no settings; `true` declares it. Grant `cloud_resource_service_account_id` access to what it reads.
- **`cloudSql`** -- `EXTERNAL_QUERY` against a Cloud SQL for PostgreSQL or MySQL database with a database user and password (stored by BigQuery, never returned).
- **`cloudSpanner`** -- federated reads from Spanner. `useParallelism` splits reads; `useDataBoost` runs them on independent compute so analytics never touch the instance's serving capacity (it requires parallelism, and `maxParallelism` requires both).
- **`aws` / `azure`** -- BigQuery Omni. AWS: an IAM role BigQuery assumes; add `aws_identity` to the role's trust policy. Azure: your tenant, and ideally your own application with a federated credential for `azure_identity`.
- **`configuration`** -- the BigQuery Connector framework, the generic path to AlloyDB and Cloud SQL through a connector id, with an optional Private Service Connect network attachment for private endpoints.
- **`spark`** -- stored procedures in Apache Spark, optionally with a Dataproc Metastore and a Spark History Server cluster; grant `spark_service_account_id` what the code reads and writes.

## Location

A dataset uses only connections in its own location. Cloud SQL must be in the same place, except that Cloud SQL us-central1 pairs with BigQuery `US` and europe-west1 with `EU`; Spanner connections take the instance's region; Omni uses `aws-us-east-1` or `azure-eastus2` style locations. Location, the id, and the connector id never change after creation.
