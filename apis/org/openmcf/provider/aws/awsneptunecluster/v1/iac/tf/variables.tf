variable "metadata" {
  description = "Resource metadata from the manifest"
  type = object({
    name = string
    id   = string
    org  = string
    env  = string
    labels = object({
      key   = string
      value = string
    })
    annotations = object({
      key   = string
      value = string
    })
    tags = list(string)
  })
}

variable "spec" {
  description = "AwsNeptuneClusterSpec configuration"
  type = object({
    # The AWS region where the Neptune cluster will be created.
    region = string

    # Subnet IDs for the Neptune subnet group
    subnet_ids = list(object({
      value = string
    }))

    # Existing Neptune subnet group (alternative to subnet_ids)
    neptune_subnet_group_name = object({
      value = string
    })

    # Security groups to associate with the cluster
    security_group_ids = list(object({
      value = string
    }))

    # IPv4 CIDRs to allow ingress
    allowed_cidr_blocks = list(string)

    # VPC
    vpc_id = object({
      value = string
    })

    # Neptune engine version (e.g., "1.2.1.0", "1.3.0.0")
    engine_version = string

    # Connection port (default: 8182)
    port = number

    # Storage type: "standard" or "iopt1"
    storage_type = string

    # Number of instances in the cluster
    instance_count = number

    # Instance class (e.g., "db.r6g.large", "db.serverless" for Serverless)
    instance_class = string

    # Serverless v2 scaling configuration (min/max NCUs). Required when instance_class is "db.serverless".
    serverless_v2_scaling = optional(object({
      min_capacity = number
      max_capacity = number
    }))

    # Enable storage encryption
    storage_encrypted = bool

    # KMS key for storage encryption
    kms_key_id = object({
      value = string
    })

    # Enable IAM database authentication
    iam_database_authentication_enabled = bool

    # IAM role ARNs to associate (e.g., for S3 bulk data loading)
    iam_roles = list(object({
      value = string
    }))

    # Backup retention period in days (1-35)
    backup_retention_period = number

    # Daily backup window (hh24:mi-hh24:mi)
    preferred_backup_window = string

    # Weekly maintenance window (ddd:hh24:mi-ddd:hh24:mi)
    preferred_maintenance_window = string

    # Enable deletion protection
    deletion_protection = bool

    # Skip final snapshot on deletion
    skip_final_snapshot = bool

    # Final snapshot identifier
    final_snapshot_identifier = string

    # CloudWatch logs to export (audit, slowquery)
    enabled_cloudwatch_logs_exports = list(string)

    # Apply modifications immediately
    apply_immediately = bool

    # Copy tags to snapshots
    copy_tags_to_snapshot = bool

    # Allow major version upgrade
    allow_major_version_upgrade = bool

    # Cluster parameter group name
    cluster_parameter_group_name = string

    # Cluster parameters
    cluster_parameters = list(object({
      name         = string
      value        = string
      apply_method = string
    }))
  })
}
