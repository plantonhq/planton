# AwsDocumentDb

An Amazon DocumentDB (with MongoDB compatibility) cluster -- a managed document database speaking the MongoDB wire protocol over shared cluster storage, provisioned or Serverless.

The cluster is the shared-storage brain -- endpoints, credentials, backups, encryption, and engine lifecycle. The compute that serves queries is the folded `instances` list: each entry materializes as its own DB instance inside the cluster (a writer plus any readers), managed per-name so scaling readers never touches the cluster. Subnets, security groups, and KMS keys compose by reference.

## Spec highlights

- **Cluster shapes** -- provisioned (`instances` with provisioned classes) or DocumentDB Serverless (`serverlessV2Scaling` DCU bounds + `db.serverless` instances).
- **Credentials** -- `manageMasterUserPassword` (recommended) keeps the master password in Secrets Manager, generated and rotated by AWS, with the secret's ARN exported; or supply `masterPassword` directly (sensitive).
- **Data protection** -- storage encryption (create-time one-way door), continuous backups with `backupRetentionPeriod`, deletion protection, and final-snapshot enforcement.
- **Restore shapes** -- create from a snapshot (`snapshotIdentifier`) or from another cluster's continuous backup (`restoreToPointInTime`, including copy-on-write fast clones), and join a DocumentDB global cluster (`globalClusterIdentifier`).
- **Observability** -- `audit` and `profiler` CloudWatch log exports (paired with their cluster parameters) and per-instance Performance Insights.
- **Per-instance maintenance control** -- each `instances` entry can stagger its maintenance window, pin its CA bundle, defer the CA-rotation restart (`certificateRotationRestart: false` -- unset keeps AWS's restart-on-rotation default), and apply its own changes immediately (`applyImmediately`).
- **Parameters** -- inline `parameters` (a module-managed cluster parameter group with the family derived from the pinned engine version) or an existing `dbClusterParameterGroupName`.

## Outputs

`cluster_identifier`, `arn`, `cluster_resource_id`, `endpoint` (writer), `reader_endpoint`, `port`, `hosted_zone_id`, `engine_version_actual`, `master_user_secret_arn`, `db_subnet_group_name`, `db_cluster_parameter_group_name`, `instance_endpoints`.

## How it works

Planton provisions via the Pulumi or Terraform module in `iac/`, both implementing the same contract at full parity. The API contract is protobuf-based (`spec.proto`); stack execution is orchestrated using `AwsDocumentDbIacInput` (provider credentials + IaC info).

## References

- DocumentDB developer guide: https://docs.aws.amazon.com/documentdb/latest/developerguide/what-is.html
- DocumentDB Serverless: https://docs.aws.amazon.com/documentdb/latest/developerguide/docdb-serverless.html
- Secrets Manager integration: https://docs.aws.amazon.com/documentdb/latest/developerguide/secrets-manager.html

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
