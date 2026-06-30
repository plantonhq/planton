# AwsRdsCluster

AWS RDS Cluster (Aurora MySQL/PostgreSQL or Multi-AZ DB Cluster) resource. Defines cluster-level configuration such as networking (subnets/DB subnet group), engine/version, encryption, maintenance/backup windows, IAM DB auth, optional Data API, serverless v2 scaling, and cluster parameter group.

## Spec fields (80/20)
- subnetIds: Two or more subnet IDs (usually private) for the cluster. Alternative: set db_subnet_group_name instead.
- dbSubnetGroupName: Existing DB subnet group to use (instead of subnet_ids).
- security_group_ids / associate_security_group_ids: Security groups to attach/use. Accepts literal or foreign-key references.
- database_name: Initial DB name to create.
- manage_master_user_password: Let RDS manage master user password via Secrets Manager (recommended default: true).
- master_user_secret_kms_key_id: KMS key (ARN/alias) for the managed secret.
- username/password: Master user credentials. If manage_master_user_password=true, do not set password.
- engine / engineVersion: Engine family and version (e.g., aurora-mysql, aurora-postgresql).
- storage_encrypted / kms_key_id: Enable storage encryption and optional KMS key.
- enabled_cloudwatch_logs_exports: Log exports; validated by engine family.
- preferred_maintenance_window: ddd:hh:mm–ddd:hh:mm (UTC).
- backup_retention_period / preferred_backup_window: Automated backups config.
- copy_tags_to_snapshot / skip_final_snapshot / final_snapshot_identifier: Snapshot behavior on deletion.
- iam_database_authentication_enabled: Enable IAM auth mappings.
- enable_http_endpoint: Data API for Aurora Serverless (where supported).
- serverless_v2_scaling: Min/max ACUs for Aurora Serverless v2.
- db_cluster_parameter_group_name / parameters: Cluster parameter group and overrides.

## Stack outputs
- rds_cluster_endpoint: Writer endpoint DNS for the cluster.
- rds_cluster_reader_endpoint: Reader endpoint for read replicas.
- rds_cluster_id / rds_cluster_arn: Identifiers of the cluster.
- rds_cluster_engine / rds_cluster_engine_version: Engine info provisioned.
- rds_cluster_port: Port used by the cluster.
- rds_subnet_group: Subnet group name in use.
- rds_security_group: Security group used by the cluster (if managed here).
- rds_cluster_parameter_group: Cluster parameter group name.

## How it works
Planton provisions via Pulumi or Terraform modules defined in this repository. The API contract is protobuf-based (api.proto, spec.proto) and stack execution is orchestrated by the platform using the AwsRdsClusterStackInput (includes provider credentials and IaC info).

## References
- AWS RDS Aurora: https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/CHAP_AuroraOverview.html
- Create DB Cluster API: https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBCluster.html
- Log exports: https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/USER_LogAccess.html
