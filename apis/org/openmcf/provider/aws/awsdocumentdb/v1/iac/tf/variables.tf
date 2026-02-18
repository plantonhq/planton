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
  description = "AwsDocumentDbSpec configuration"
  type = object({
    # The AWS region where the resource will be created.
    region = string

    # Subnets for the DB subnet group
    subnets = list(object({
      value = string
    }))

    # Existing DB subnet group (alternative to subnets)
    db_subnet_group = object({
      value = string
    })

    # Security groups to associate with the cluster
    security_groups = list(object({
      value = string
    }))

    # IPv4 CIDRs to allow ingress
    allowed_cidrs = list(string)

    # VPC
    vpc = object({
      value = string
    })

    # DocumentDB engine version (e.g., "4.0.0", "5.0.0")
    engine_version = string

    # Connection port (default: 27017)
    port = number

    # Master username
    master_username = string

    # Master password
    master_password = string

    # Number of instances in the cluster
    instance_count = number

    # Instance class (e.g., "db.r6g.large")
    instance_class = string

    # Enable storage encryption
    storage_encrypted = bool

    # KMS key for storage encryption
    kms_key = object({
      value = string
    })

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

    # CloudWatch logs to export (audit, profiler)
    enabled_cloudwatch_logs_exports = list(string)

    # Apply modifications immediately
    apply_immediately = bool

    # Enable automatic minor version upgrades
    auto_minor_version_upgrade = bool

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
