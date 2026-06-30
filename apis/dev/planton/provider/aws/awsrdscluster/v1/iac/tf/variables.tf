variable "metadata" {
  description = "metadata"
  type = object({

    # name of the resource
    name = string

    # id of the resource
    id = string

    # id of the organization to which the api-resource belongs to
    org = string

    # environment to which the resource belongs to
    env = string

    # labels for the resource
    labels = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # annotations for the resource
    annotations = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # tags for the resource
    tags = list(string)
  })
}

variable "spec" {
  description = "spec"
  type = object({

    # The AWS region where the resource will be created.
    region = string

    # Name of the database engine to be used for this DB cluster. Valid Values: `aurora-mysql`,
    # `aurora-postgresql`, `mysql`, `postgres`. (Note that `mysql` and `postgres` are Multi-AZ RDS clusters).
    engine = string

    # Database engine version. Updating this argument results in an outage.
    # See the [Aurora MySQL](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/AuroraMySQL.Updates.html) and
    # [Aurora Postgres](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/AuroraPostgreSQL.Updates.html)
    # documentation for your configured engine to determine this value, or by running
    # `aws rds describe-db-engine-versions`. For example with Aurora MySQL 2, a potential value for this
    # argument is `5.7.mysql_aurora.2.03.2`. The value can contain a partial version where supported by the API.
    # The actual engine version used is returned in the attribute `engineVersionActual`
    engine_version = string

    # Database engine mode. Valid values: `global` (only valid for Aurora MySQL 1.21 and earlier),
    # `parallelquery`, `provisioned`, `serverless`. Defaults to: `provisioned`.
    # See the [RDS User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless.html)
    # for limitations when using `serverless`.
    engine_mode = string

    # Family of the DB parameter group.
    cluster_family = string

    # Instance class to use. For details on CPU and memory, see [Scaling Aurora DB Instances](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Aurora.Managing.html).
    # Aurora uses `db.*` instance classes/types. Please see [AWS Documentation](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html)
    # for currently available instance classes and complete details. For Aurora Serverless v2 use `db.serverless`.
    # EC2 instance type for aws rds cluster
    # https://aws.amazon.com/rds/aurora/pricing
    instance_type = string

    # aws rds cluster size
    cluster_size = number

    # Set to true to allow RDS to manage the master user password in Secrets Manager. Cannot be set if master_password is provided
    manage_master_user_password = bool

    # Amazon Web Services KMS key identifier is the key ARN, key ID, alias ARN, or alias name for the KMS key.
    # To use a KMS key in a different Amazon Web Services account, specify the key ARN or alias ARN.
    # If not specified, the default KMS key for your Amazon Web Services account is used.
    master_user_secret_kms_key_id = string

    # Username for the master DB user. Ignored if snapshot_identifier or replication_source_identifier is provided
    master_user = string

    # Password for the master DB user. Ignored if snapshot_identifier or replication_source_identifier is provided
    master_password = string

    # Database name (default is not to create a database)
    database_name = string

    # Set true to make this database accessible from the public internet
    is_publicly_accessible = bool

    # database port
    database_port = number

    # VPC ID to create the cluster in (e.g. `vpc-a22222ee`). Defaults to the region's default VPC.
    vpc_id = string

    # List of subnet IDs for the DB. DB instance will be created in the VPC associated with the DB subnet group
    # provisioned using the subnet IDs. Specify one of `subnet_ids`, `db_subnet_group_name`
    subnet_ids = list(string)

    # Name of DB subnet group. DB instance will be created in the VPC associated with the DB subnet group.
    # Specify one of `subnet_ids`, `db_subnet_group_name`
    db_subnet_group_name = string

    # The IDs of the security groups from which to allow `ingress` traffic to the DB instance
    security_group_ids = list(string)

    # Whether to allow traffic between resources inside the database's security group.
    intra_security_group_traffic_enabled = bool

    # List of CIDRs allowed to access the database (in addition to security groups and subnets)
    allowed_cidr_blocks = list(string)

    # The IDs of the existing security groups to associate with the DB instance
    associate_security_group_ids = list(string)

    # Specifies whether or mappings of AWS Identity and Access Management (IAM) accounts to database accounts is enabled
    iam_database_authentication_enabled = bool

    # Specifies whether the DB cluster is encrypted
    storage_encrypted = bool

    # The ARN for the KMS encryption key. When specifying `kms_key_arn`, `storage_encrypted` needs to be set to `true`
    storage_kms_key_arn = string

    # Whether to enable Performance Insights
    is_performance_insights_enabled = bool

    # The ARN for the KMS encryption key. When specifying `kms_key_arn`, `is_performance_insights_enabled` needs to be set to `true`
    performance_insights_kms_key_id = string

    # Weekly time range during which system maintenance can occur, in UTC
    maintenance_window = string

    # List of log types to enable for exporting to CloudWatch logs. If omitted, no logs will be exported.
    # Valid values (depending on engine): alert, audit, error, general, listener, slowquery, trace, postgresql (PostgreSQL),
    # upgrade (PostgreSQL).
    enabled_cloudwatch_logs_exports = list(string)

    # A boolean flag to enable/disable the creation of the enhanced monitoring IAM role.
    enhanced_monitoring_role_enabled = bool

    # Attributes used to format the Enhanced Monitoring IAM role.
    # If this role hits IAM role length restrictions (max 64 characters),
    # consider shortening these strings.
    enhanced_monitoring_attributes = list(string)

    # Description for rds_monitoring_interval
    rds_monitoring_interval = number

    # Normally AWS makes a snapshot of the database before deleting it. Set this to `true` in order to skip this.
    # NOTE: The final snapshot has a name derived from the cluster name. If you delete a cluster, get a final snapshot,
    # then create a cluster of the same name, its final snapshot will fail with a name collision unless you delete
    # the previous final snapshot first.
    skip_final_snapshot = bool

    # Specifies whether the Cluster should have deletion protection enabled. The database can't be deleted when this value is set to `true`
    deletion_protection = bool

    # Specifies whether or not to create this cluster from a snapshot
    snapshot_identifier = string

    # Enable to allow major engine version upgrades when changing engine versions. Defaults to false.
    allow_major_version_upgrade = bool

    # The identifier of the CA certificate for the DB instance
    ca_cert_identifier = string

    # Number of days to retain backups for
    retention_period = number

    # Daily time range during which the backups happen, UTC
    backup_window = string

    # AwsRdsClusterAutoScaling defines the auto-scaling settings for the RDS cluster, allowing dynamic scaling of instances
    # based on specified metrics and policies.
    auto_scaling = object({

      # Whether to enable cluster autoscaling
      is_enabled = bool

      # Autoscaling policy type. `TargetTrackingScaling` and `StepScaling` are supported
      policy_type = string

      # The metrics type to use. If this value isn't provided the default is CPU utilization
      target_metrics = string

      # The target value to scale with respect to target metrics
      target_value = number

      # The amount of time, in seconds, after a scaling activity completes and before the next scaling down activity can start. Default is 300s
      scale_in_cooldown = number

      # The amount of time, in seconds, after a scaling activity completes and before the next scaling up activity can start. Default is 300s
      scale_out_cooldown = number

      # Minimum number of instances to be maintained by the autoscaler
      min_capacity = number

      # Maximum number of instances to be maintained by the autoscaler
      max_capacity = number
    })

    # List of nested attributes with scaling properties. Only valid when `engine_mode` is set to `serverless`. This is required for Serverless v1
    scaling_configuration = object({

      # Whether to enable automatic pause. A DB cluster can be paused only when it's idle (it has no connections).
      # If a DB cluster is paused for more than seven days, the DB cluster might be backed up with a snapshot.
      # In this case, the DB cluster is restored when there is a request to connect to it.
      auto_pause = bool

      # Maximum capacity for an Aurora DB cluster in `serverless` DB engine mode.
      # The maximum capacity must be greater than or equal to the minimum capacity.
      # Valid Aurora PostgreSQL capacity values are (`2`, `4`, `8`, `16`, `32`, `64`, `192`, and `384`). Defaults to `16`.
      max_capacity = number

      # Minimum capacity for an Aurora DB cluster in `serverless` DB engine mode.
      # The minimum capacity must be lesser than or equal to the maximum capacity.
      # Valid Aurora PostgreSQL capacity values are (`2`, `4`, `8`, `16`, `32`, `64`, `192`, and `384`). Defaults to `2`.
      min_capacity = number

      # Time, in seconds, before an Aurora DB cluster in serverless mode is paused. Valid values are `300` through `86400`. Defaults to `300`.
      seconds_until_auto_pause = number

      # Action to take when the timeout is reached. Valid values: `ForceApplyCapacityChange`, `RollbackCapacityChange`.
      # Defaults to `RollbackCapacityChange`.
      # See [documentation](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless-v1.how-it-works.html#aurora-serverless.how-it-works.timeout-action).
      timeout_action = string
    })

    # Nested attribute with scaling properties for ServerlessV2. Only valid when `engine_mode` is set to `provisioned.` This is required for Serverless v2
    serverlessv2_scaling_configuration = object({

      # Minimum capacity for an Aurora DB cluster in `provisioned` DB engine mode. The minimum capacity must be
      # lesser than or equal to the maximum capacity. Valid capacity values are in a range of `0.5` up to `128` in steps of `0.5`.
      max_capacity = number

      # Maximum capacity for an Aurora DB cluster in `provisioned` DB engine mode. The maximum capacity must be
      # greater than or equal to the minimum capacity. Valid capacity values are in a range of `0.5` up to `128` in steps of `0.5`.
      min_capacity = number
    })

    # Name of the DB cluster parameter group to associate.
    cluster_parameter_group_name = string

    # List of DB cluster parameters to apply
    cluster_parameters = list(object({

      # "immediate" (default), or "pending-reboot". Some
      # engines can't apply some parameters without a reboot, and you will need to
      # specify "pending-reboot" here.
      apply_method = string

      # The name of the DB parameter.
      name = string

      # The value of the DB parameter.
      value = string
    }))
  })
}