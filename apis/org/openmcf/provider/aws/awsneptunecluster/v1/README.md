# AwsNeptuneCluster

Provision an AWS Neptune cluster—a fully managed graph database service supporting property-graph (Apache TinkerPop Gremlin) and RDF (SPARQL) query languages. Neptune excels at connected data: social graphs, recommendation engines, fraud detection, and knowledge graphs. Unlike relational databases, Neptune does not use master username/password; access is controlled via IAM database authentication and network-level security (VPC, security groups).

## Spec fields (80/20)
- subnetIds: Subnet IDs for the Neptune subnet group (>=2) or use neptuneSubnetGroupName.
- neptuneSubnetGroupName: Existing Neptune subnet group name (alternative to subnetIds).
- securityGroupIds: Security groups to associate with the cluster.
- allowedCidrBlocks: IPv4 CIDRs to allow ingress to the cluster.
- vpcId: VPC where the cluster will be deployed.
- engineVersion: Neptune engine version (e.g., "1.2.1.0", "1.3.0.0"). Default: "1.3.0.0".
- port: TCP port for connections. Default: 8182.
- storageType: "standard" (default) or "iopt1" (I/O-Optimized for read-heavy workloads).
- instanceCount: Number of instances in the cluster. Default: 1.
- instanceClass: Instance class (e.g., "db.r6g.large", "db.r6g.xlarge"). Use "db.serverless" for Neptune Serverless.
- serverlessV2Scaling: Min/max Neptune Capacity Units (NCUs) when using db.serverless.
- storageEncrypted: Enable storage encryption at rest. Default: true.
- kmsKeyId: KMS key ARN for storage encryption.
- iamDatabaseAuthenticationEnabled: Enable IAM database authentication.
- iamRoles: IAM role ARNs for S3 bulk loading and other service integrations.
- backupRetentionPeriod: Days to retain automated backups (1-35). Default: 7.
- preferredBackupWindow: Daily backup time range in UTC (hh24:mi-hh24:mi).
- preferredMaintenanceWindow: Weekly maintenance window in UTC (ddd:hh24:mi-ddd:hh24:mi).
- deletionProtection: Prevent accidental cluster deletion.
- skipFinalSnapshot: Skip final snapshot on deletion. Default: false.
- finalSnapshotIdentifier: Identifier for final snapshot when not skipping.
- enabledCloudwatchLogsExports: Log types to export ("audit", "slowquery").
- applyImmediately: Apply modifications immediately vs. next maintenance window.
- copyTagsToSnapshot: Copy cluster tags to snapshots.
- clusterParameterGroupName: Custom cluster parameter group name.
- clusterParameters: Custom parameters for the cluster parameter group.

## Stack outputs
- cluster_endpoint: Primary writer endpoint for Gremlin/SPARQL queries.
- cluster_reader_endpoint: Reader endpoint for load-balanced read traffic.
- cluster_id: AWS identifier of the cluster.
- cluster_arn: Amazon Resource Name of the cluster.
- cluster_resource_id: Internal AWS resource identifier.
- cluster_port: Port on which the cluster accepts connections (default 8182).
- db_subnet_group_name: Name of the associated Neptune subnet group.
- security_group_id: Security group ID associated with the cluster.
- cluster_parameter_group_name: Parameter group name in use.
- hosted_zone_id: Route 53 hosted zone ID for the cluster endpoint.

## How it works
The CLI passes a Stack Input with provisioner choice (Pulumi or Terraform), stack info, the target `AwsNeptuneCluster` resource, and AWS credentials to the corresponding module. Neptune clusters are created with optional subnet groups, security groups, and parameter groups; instances are provisioned according to instanceCount and instanceClass (or serverless configuration).

## References
- AWS Neptune: https://docs.aws.amazon.com/neptune/latest/userguide/what-is-neptune.html
- Neptune Engine Versions: https://docs.aws.amazon.com/neptune/latest/userguide/engine-releases.html
- Neptune Instance Classes: https://docs.aws.amazon.com/neptune/latest/userguide/instance-classes.html
- Gremlin: https://tinkerpop.apache.org/gremlin.html
- SPARQL: https://www.w3.org/TR/sparql11-overview/
