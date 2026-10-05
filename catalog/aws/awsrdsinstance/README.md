# AwsRdsInstance

A single RDS DB instance: a standalone database server (postgres, mysql, mariadb, oracle-*, sqlserver-*) with its own EBS-backed storage -- optionally Multi-AZ with a synchronous standby, or a read replica of another instance.

This is the classic single-node RDS shape. For Aurora's shared-storage clusters (and Multi-AZ DB clusters of mysql/postgres), use `AwsRdsCluster` instead. Subnets, security groups, KMS keys, and the monitoring role compose by reference.

## Spec highlights

- **Credentials** -- `manageMasterUserPassword` (recommended) keeps the master password in Secrets Manager, generated and rotated by AWS, with the secret's ARN exported; or supply `password` directly (sensitive).
- **Storage** -- gp3/gp2/io1/io2 with IOPS and throughput tuning, autoscaling ceiling (`maxAllocatedStorageGb`), dedicated log volume, encryption (create-time one-way door) with an optional `AwsKmsKey` reference.
- **Availability** -- `multiAz` synchronous standby with automatic failover, or a single-AZ instance optionally pinned to a zone.
- **Replicas and restores** -- `replicateSourceDb` read replicas (same- or cross-region, Oracle `mounted` mode, with `upgradeStorageConfig` for older sources), snapshot restore, point-in-time restore, and Percona XtraBackup restore from S3 (`s3Import`, MySQL) -- each a first-class create shape.
- **Updates and lifecycle** -- RDS Blue/Green Deployments (`blueGreenUpdateEnabled`) for near-zero-downtime engine upgrades, tri-state minor-version auto-upgrade, major-version guard, apply-immediately control, and the extended-support opt-out (`engineLifecycleSupport`).
- **Engine configuration** -- inline `parameters` and `options` (module-managed parameter and option groups with family/version derived from the pinned engine version -- the same own-XOR-existing idiom as the cluster kind) or existing groups by name.
- **Integrations** -- IAM database authentication, feature-scoped engine `iamRoles` (one association per entry; `featureName` is required and engine-specific), CloudWatch log exports, Performance Insights, Enhanced Monitoring, Database Insights, Active Directory join (AWS-managed or self-managed), license models and character sets for Oracle/SQL Server.

## Outputs

`instance_identifier`, `arn`, `resource_id`, `endpoint` (address:port), `address`, `port`, `hosted_zone_id`, `engine_version_actual`, `master_user_secret_arn`, `db_subnet_group_name`, `db_parameter_group_name`, `option_group_name`.

## How it works

Planton provisions via the Pulumi or Terraform module in `iac/`, both implementing the same contract at full parity. The API contract is protobuf-based (`spec.proto`); stack execution is orchestrated using `AwsRdsInstanceIacInput` (provider credentials + IaC info).

## References

- RDS user guide: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Welcome.html
- Blue/Green Deployments: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/blue-green-deployments.html
- Read replicas: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_ReadRepl.html

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
