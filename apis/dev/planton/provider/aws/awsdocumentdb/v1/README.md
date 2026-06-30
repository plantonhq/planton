# AwsDocumentDb

Provision an AWS DocumentDB cluster - a fully managed, MongoDB-compatible document database service. Focuses on essential networking, engine configuration, authentication, encryption, and backup settings.

## Spec fields (summary)
- subnets: Private subnets for the DB subnet group (>=2) or use db_subnet_group.
- dbSubnetGroup: Existing DB subnet group name (alternative to subnets).
- securityGroups: Security groups to associate with the cluster.
- allowedCidrs: IPv4 CIDRs to allow ingress to the cluster.
- vpc: VPC where the cluster will be deployed.
- engineVersion: DocumentDB engine version (e.g., "4.0.0", "5.0.0"). Default: "5.0.0".
- port: TCP port for connections. Default: 27017.
- masterUsername: Master user name. Default: "docdbadmin".
- masterPassword: Master user password (required).
- instanceCount: Number of instances in the cluster. Default: 1.
- instanceClass: Instance class (e.g., "db.r5.large", "db.r6g.large"). Default: "db.r6g.large".
- storageEncrypted: Enable storage encryption at rest. Default: true.
- kmsKey: KMS key ARN for storage encryption.
- backupRetentionPeriod: Days to retain automated backups (1-35). Default: 7.
- preferredBackupWindow: Daily backup time range in UTC (hh24:mi-hh24:mi).
- preferredMaintenanceWindow: Weekly maintenance window in UTC (ddd:hh24:mi-ddd:hh24:mi).
- deletionProtection: Prevent accidental cluster deletion.
- skipFinalSnapshot: Skip final snapshot on deletion. Default: false.
- finalSnapshotIdentifier: Identifier for final snapshot when not skipping.
- enabledCloudwatchLogsExports: Log types to export ("audit", "profiler").
- applyImmediately: Apply modifications immediately vs. next maintenance window.
- autoMinorVersionUpgrade: Enable automatic minor version upgrades. Default: true.
- clusterParameterGroupName: Custom cluster parameter group name.
- clusterParameters: Custom parameters for the cluster parameter group.

## Stack outputs
- cluster_endpoint: Primary writer endpoint for the cluster.
- cluster_reader_endpoint: Reader endpoint for load-balanced read traffic.
- cluster_id: AWS identifier of the cluster.
- cluster_arn: Amazon Resource Name of the cluster.
- cluster_port: Port on which the cluster accepts connections.
- db_subnet_group_name: Name of the associated DB subnet group.
- security_group_id: Security group ID associated with the cluster.
- cluster_parameter_group_name: Parameter group name in use.
- connection_string: MongoDB-compatible connection string.
- cluster_resource_id: Internal AWS resource identifier.

## How it works
The CLI passes a Stack Input with provisioner choice (Pulumi or Terraform), stack info, the target `AwsDocumentDb` resource, and AWS credentials to the corresponding module.

## References
- AWS DocumentDB: https://docs.aws.amazon.com/documentdb/latest/developerguide/what-is.html
- DocumentDB Engine Versions: https://docs.aws.amazon.com/documentdb/latest/developerguide/release-notes.html
- Instance Classes: https://docs.aws.amazon.com/documentdb/latest/developerguide/db-instance-classes.html
- MongoDB Compatibility: https://docs.aws.amazon.com/documentdb/latest/developerguide/mongo-apis.html
