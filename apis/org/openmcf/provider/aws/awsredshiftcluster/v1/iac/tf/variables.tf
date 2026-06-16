variable "metadata" {
  description = "OpenMCF resource metadata"
  type = object({
    name = string
    org  = optional(string, "")
    env  = optional(string, "")
    id   = optional(string, "")
  })
}

variable "spec" {
  description = "AwsRedshiftCluster spec"
  type = object({
    # The AWS region where the resource will be created.
    region                   = string
    # Core
    node_type                = string
    number_of_nodes          = optional(number, 1)
    database_name            = optional(string, "dev")
    master_username          = optional(string, "admin")
    master_password          = optional(string, "")
    manage_master_password   = optional(bool, false)
    master_password_secret_kms_key_id = optional(string, "")
    port                     = optional(number, 5439)

    # Networking
    subnet_ids                   = optional(list(string), [])
    cluster_subnet_group_name    = optional(string, "")
    security_group_ids           = optional(list(string), [])
    allowed_cidr_blocks          = optional(list(string), [])
    associate_security_group_ids = optional(list(string), [])
    vpc_id                       = optional(string, "")
    publicly_accessible          = optional(bool, false)
    enhanced_vpc_routing         = optional(bool, false)
    multi_az                     = optional(bool, false)

    # Encryption
    encrypted  = optional(bool, true)
    kms_key_id = optional(string, "")

    # IAM
    iam_roles           = optional(list(string), [])
    default_iam_role_arn = optional(string, "")

    # Snapshots
    automated_snapshot_retention_period = optional(number, 1)
    skip_final_snapshot                 = optional(bool, false)
    final_snapshot_identifier           = optional(string, "")

    # Maintenance
    preferred_maintenance_window = optional(string, "")
    allow_version_upgrade        = optional(bool, true)
    maintenance_track_name       = optional(string, "")
    apply_immediately            = optional(bool, false)

    # Logging
    logging = optional(object({
      log_destination_type = string
      s3_bucket_name       = optional(string, "")
      s3_key_prefix        = optional(string, "")
      log_exports          = optional(list(string), [])
    }), null)

    # Parameter Group
    cluster_parameter_group_name = optional(string, "")
    parameters = optional(list(object({
      name  = string
      value = string
    })), [])
  })
}

