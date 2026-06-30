# AwsRdsInstance

Provision a single AWS RDS DB instance (PostgreSQL, MySQL, MariaDB, Oracle, SQL Server). Focuses on essential networking, engine selection, sizing, and credentials.

## Spec fields (summary)
- subnetIds: Private subnets for the DB subnet group (>=2) or use db_subnet_group_name.
- dbSubnetGroupName: Existing DB subnet group name (alternative to subnet_ids).
- securityGroupIds: Security groups to associate with the instance.
- engine: Database engine (e.g., postgres, mysql, mariadb, oracle-se2, sqlserver-ex).
- engineVersion: Desired engine version (e.g., 14.10 for Postgres).
- instanceClass: DB instance class (e.g., db.t3.micro, db.m6g.large).
- allocatedStorageGb: Allocated storage for the instance in GiB (>0).
- storageEncrypted: Enable storage encryption.
- kms_key_id: KMS key ARN/alias for encryption when enabled.
- username: Master username.
- password: Master user password.
- port: Database port (0–65535).
- publiclyAccessible: Whether to allocate a public IP.
- multiAz: Enable Multi-AZ deployment.
- parameter_group_name: Optional DB parameter group name.
- option_group_name: Optional option group name.

## Stack outputs
- rds_instance_id: RDS instance identifier.
- rds_instance_arn: RDS instance ARN.
- rds_instance_endpoint: Endpoint hostname.
- rds_instance_port: Database port.
- rds_subnet_group: Subnet group name.
- rds_security_group: Associated security group ID.
- rds_parameter_group: Parameter group name.

## How it works
- The CLI passes a Stack Input with provisioner choice (Pulumi or Terraform), stack info, the target `AwsRdsInstance` resource, and AWS credentials to the corresponding module.

## References
- AWS RDS Instances: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.html
- Engine versions (PostgreSQL): https://docs.aws.amazon.com/AmazonRDS/latest/PostgreSQLReleaseNotes/Welcome.html
- Engine versions (MySQL): https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/MySQL.Concepts.VersionMgmt.html
