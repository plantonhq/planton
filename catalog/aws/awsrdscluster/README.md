# AwsRdsCluster

An RDS DB cluster: an Aurora MySQL/PostgreSQL cluster (provisioned or Serverless v2), an Aurora Serverless v1 cluster, or a Multi-AZ RDS cluster for the community mysql/postgres engines.

The cluster is the shared-storage brain -- endpoints, credentials, backups, encryption, and engine lifecycle. The compute that serves queries is the folded `instances` list: each entry materializes as its own DB instance inside the cluster (a writer plus any readers), managed per-name so scaling readers never touches the cluster. Subnets, security groups, KMS keys, and IAM roles compose by reference.

## Spec highlights

- **Cluster shapes** -- Aurora provisioned (`instances` with provisioned classes), Aurora Serverless v2 (`serverlessV2Scaling` + `db.serverless` instances, scale-to-zero with `minCapacity: 0`), legacy Aurora Serverless v1 (`engineMode: serverless` + `serverlessV1Scaling`), and Multi-AZ RDS clusters (`engine: mysql|postgres` + `dbClusterInstanceClass` + storage sizing).
- **Credentials** -- `manageMasterUserPassword` (recommended) keeps the master password in Secrets Manager, generated and rotated by AWS, with the secret's ARN exported; or supply `masterPassword` directly (sensitive).
- **Data protection** -- storage encryption (create-time one-way door), continuous backups with `backupRetentionPeriod`, deletion protection, final-snapshot enforcement, `deleteAutomatedBackups: false` retention, Aurora MySQL backtrack.
- **Restore shapes** -- create from a snapshot (`snapshotIdentifier`), from another cluster's continuous backup (`restoreToPointInTime`, including Aurora copy-on-write fast clones), or from a Percona XtraBackup in S3 (`s3Import`, the Aurora MySQL migration on-ramp).
- **Integrations** -- IAM database authentication, feature-scoped engine `iamRoles` (each entry one role association with an optional `featureName` -- s3Import/s3Export/Lambda/SageMaker), Kerberos through an AWS Managed Microsoft AD (`domain` + `domainIamRoleName`), the Data API (`enableHttpEndpoint`), CloudWatch log exports, Performance Insights, Enhanced Monitoring, Database Insights, Aurora Global Database membership, and local/global write forwarding.
- **Audit and access shapes** -- `activityStream` streams every database event to a KMS-encrypted Kinesis stream (Aurora only; the stream name is an output), and `customEndpoints` carve stable READER/ANY DNS names over chosen instances (member names reference `instances` entries).
- **Per-instance overrides** -- each `instances` entry can set its own Performance Insights key and retention, backup/maintenance windows, monitoring role, snapshot tag copy, and `applyImmediately`, on top of class, tier, AZ, and parameter group.
- **Parameters** -- inline `parameters` (a module-managed cluster parameter group with the family derived from the pinned engine version) or an existing `dbClusterParameterGroupName`.

## Outputs

`cluster_identifier`, `arn`, `cluster_resource_id`, `endpoint` (writer), `reader_endpoint`, `port`, `hosted_zone_id`, `engine_version_actual`, `master_user_secret_arn`, `db_subnet_group_name`, `db_cluster_parameter_group_name`, `instance_endpoints`, `custom_endpoints` (name + DNS per entry), `activity_stream_kinesis_stream_name`.

## How it works

Planton provisions via the Pulumi or Terraform module in `iac/`, both implementing the same contract at full parity. The API contract is protobuf-based (`spec.proto`); stack execution is orchestrated using `AwsRdsClusterIacInput` (provider credentials + IaC info).

## References

- Aurora user guide: https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/CHAP_AuroraOverview.html
- Multi-AZ DB clusters: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/multi-az-db-clusters-concepts.html
- Aurora Serverless v2: https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless-v2.html

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
