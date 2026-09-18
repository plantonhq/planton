variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsRdsInstance specification"
  type = object({
    # The AWS region the instance is created in. Must match the region of
    # the subnets, security groups, and KMS keys it references.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # Subnets for the instance's DB subnet group. Provide at least two
    # subnets in DISTINCT availability zones -- AWS requires the subnet
    # group to cover two AZs even for a single-AZ instance. Reference
    # AwsSubnet subnet_id outputs or pass literal subnet IDs. The module
    # manages the subnet group itself (pure glue); alternatively point
    # db_subnet_group_name at an existing group.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = optional(list(string), [])

    # Name of an existing DB subnet group to place the instance in,
    # instead of providing subnet_ids.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    db_subnet_group_name = optional(string, "")

    # Security groups attached to the instance. Empty uses the VPC's
    # default security group (the AWS default). Reference AwsSecurityGroup
    # security_group_id outputs or pass literal SG IDs.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # The database engine: "postgres", "mysql", "mariadb", "oracle-ee",
    # "oracle-se2", "sqlserver-ex", "sqlserver-web", "sqlserver-se",
    # "sqlserver-ee", and license-included/CDB variants. Required for a
    # new instance (a read replica inherits it from the source). Changing
    # the engine replaces the instance.
    engine = optional(string, "")

    # The engine version, e.g. "16.4" (postgres) or "8.0.39" (mysql).
    # Leave empty to let AWS pick the engine's current default version --
    # an empty pin never goes stale. Minor upgrades apply in place; major
    # upgrades additionally need allow_major_version_upgrade.
    engine_version = optional(string, "")

    # The instance class (compute size), e.g. "db.t4g.micro",
    # "db.m6g.large". Required.
    instance_class = string

    # Provisioned storage in GiB. Required for a new instance (read
    # replicas and snapshot restores inherit the source's storage).
    # Growing it applies in place; shrinking requires a new instance.
    allocated_storage_gb = optional(number, 0)

    # Storage autoscaling ceiling in GiB. When set above
    # allocated_storage_gb, RDS grows storage automatically as the
    # database approaches capacity -- the cheap insurance against
    # disk-full outages. 0 disables autoscaling.
    max_allocated_storage_gb = optional(number, 0)

    # The EBS storage type: "gp3" (the modern default -- baseline
    # 3000 IOPS/125 MiB/s, independently tunable), "gp2" (legacy
    # burst-credit SSD), "io1"/"io2" (provisioned-IOPS for
    # latency-critical workloads), or "standard" (magnetic, legacy).
    # Empty keeps the AWS default.
    storage_type = optional(string, "")

    # Provisioned IOPS. Required for io1/io2; optional for gp3 to raise
    # it above the 3000 baseline. 0 keeps the storage type's default.
    iops = optional(number, 0)

    # Storage throughput in MiB/s, gp3 only, to raise it above the 125
    # baseline. 0 keeps the default.
    storage_throughput = optional(number, 0)

    # Use a dedicated EBS volume for database logs instead of sharing the
    # data volume -- steadier I/O for audit-heavy or WAL-heavy workloads.
    dedicated_log_volume = optional(bool, false)

    # Encrypt instance storage at rest. Strongly recommended -- and
    # create-time only: an unencrypted instance cannot be encrypted later
    # (requires a snapshot-restore migration). Read replicas inherit the
    # source's encryption.
    storage_encrypted = optional(bool, false)

    # The KMS key for storage encryption when storage_encrypted is true.
    # Empty uses the AWS-managed aws/rds key. Reference an AwsKmsKey
    # key_arn output or pass a literal key ARN. Create-time only.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # The name of the initial database AWS creates. Empty creates no
    # database. Create-time only. Not supported by SQL Server.
    db_name = optional(string, "")

    # The master username. Required for a brand-new instance -- AWS has
    # no default and rejects a blank value at CreateDBInstance. Only
    # instances that inherit credentials from a source (a read replica,
    # a snapshot restore, or a point-in-time restore) leave it empty.
    # Avoid the engine's reserved names (e.g. "rdsadmin"). Create-time
    # only -- changing it replaces the instance.
    username = optional(string, "")

    # Let AWS manage the master password in Secrets Manager: AWS
    # generates it, stores it, rotates it on schedule, and no secret ever
    # touches this manifest or the IaC state. The managed secret's ARN is
    # exported as the master_user_secret_arn output. Mutually exclusive
    # with password -- and the recommended posture.
    manage_master_user_password = optional(bool, false)

    # The KMS key that encrypts the AWS-managed master-user secret (only
    # meaningful with manage_master_user_password). Empty uses the
    # account's default aws/secretsmanager key. Reference an AwsKmsKey
    # key_arn output or pass a literal key ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    master_user_secret_kms_key_id = optional(string, "")

    # The master password, supplied directly. Stored in IaC state --
    # prefer manage_master_user_password, which keeps the secret in
    # Secrets Manager entirely. Mutually exclusive with
    # manage_master_user_password.
    password = optional(string, "")

    # The port the instance accepts connections on. 0 keeps the engine
    # default (5432 postgres, 3306 mysql/mariadb, 1521 oracle, 1433
    # sqlserver).
    port = optional(number, 0)

    # Deploy a synchronous standby replica in a second AZ with automatic
    # failover -- the single most important availability knob on a
    # production instance. (For a Multi-AZ CLUSTER with readable
    # standbys, use AwsRdsCluster with a community engine instead.)
    multi_az = optional(bool, false)

    # Pin a single-AZ instance to one availability zone. Empty lets AWS
    # place it. Cannot be combined with multi_az.
    availability_zone = optional(string, "")

    # Give the instance a public IP. Requires public subnets; keep false
    # for anything production-shaped.
    publicly_accessible = optional(bool, false)

    # The network stack: "IPV4" (AWS default when unset) or "DUAL" for
    # dual-stack IPv4+IPv6.
    network_type = optional(string, "")

    # Make this instance a read replica of the given source: a source
    # instance identifier (same region) or ARN (cross-region). Engine,
    # storage, and credentials are inherited from the source. Promote to
    # a standalone instance by clearing the field.
    replicate_source_db = optional(string, "")

    # Replica behavior, Oracle only: "open-read-only" (a queryable
    # replica) or "mounted" (a running-but-closed disaster-recovery
    # target). Empty keeps the AWS default (open-read-only).
    replica_mode = optional(string, "")

    # Restore from an existing DB snapshot (name or ARN) at create time.
    # Create-time only; mutually exclusive with restore_to_point_in_time
    # and replicate_source_db.
    snapshot_identifier = optional(string, "")

    # Restore from another instance's continuous backup at create time --
    # point-in-time recovery as a first-class create shape. Create-time
    # only; mutually exclusive with snapshot_identifier and
    # replicate_source_db.
    restore_to_point_in_time = optional(object({
      # The source instance identifier. Exactly one source field must be
      # set.
      source_db_instance_identifier = optional(string, "")

      # The source instance's immutable resource ID (resource_id output) --
      # survives identifier renames and points at deleted instances'
      # retained backups.
      source_dbi_resource_id = optional(string, "")

      # The ARN of a retained automated backup to restore from (for
      # instances already deleted).
      source_db_instance_automated_backups_arn = optional(string, "")

      # The UTC timestamp to restore to, RFC3339 (e.g.
      # "2026-07-01T09:45:00Z"). Mutually exclusive with
      # use_latest_restorable_time.
      restore_time = optional(string, "")

      # Restore to the most recent recoverable moment. Mutually exclusive
      # with restore_time.
      use_latest_restorable_time = optional(bool, false)
    }))

    # Days automated backups are retained, 0-35. 0 disables automated
    # backups (and point-in-time recovery) -- production instances want
    # 7+. Note: a MySQL-family read-replica source needs retention > 0.
    backup_retention_period = optional(number, 0)

    # The daily backup window in UTC, format "hh24:mi-hh24:mi" (e.g.
    # "04:00-05:00"). Empty lets AWS assign one. Must not overlap the
    # maintenance window.
    backup_window = optional(string, "")

    # The weekly maintenance window in UTC, format
    # "ddd:hh24:mi-ddd:hh24:mi" (e.g. "sun:05:00-sun:06:00"). Empty lets
    # AWS assign one.
    maintenance_window = optional(string, "")

    # Copy the instance's tags onto automated and manual snapshots.
    copy_tags_to_snapshot = optional(bool, false)

    # Remove automated backups immediately when the instance is deleted.
    # AWS defaults this to true; set false to retain the backups for the
    # remainder of their retention window after deletion -- the last line
    # of defense against a mistaken teardown.
    delete_automated_backups = optional(bool)

    # Skip the final snapshot when the instance is deleted. When false
    # (the safe default), final_snapshot_identifier must be set -- AWS
    # refuses to delete without knowing the snapshot name.
    skip_final_snapshot = optional(bool, false)

    # The name for the final snapshot taken on deletion. Required when
    # skip_final_snapshot is false. Must start with a letter, contain
    # only letters, numbers, and hyphens, with no consecutive or
    # trailing hyphens -- AWS's snapshot-identifier rules.
    final_snapshot_identifier = optional(string, "")

    # Refuse deletion of the instance while enabled. Turn this on for
    # anything holding data you cannot recreate.
    deletion_protection = optional(bool, false)

    # Map IAM identities to database users -- connect with short-lived
    # IAM auth tokens instead of passwords. MySQL and PostgreSQL engines.
    iam_database_authentication_enabled = optional(bool, false)

    # Database log types to export to CloudWatch Logs. Valid types vary
    # by engine: postgres exports "postgresql"/"upgrade"/"iam-db-auth-error";
    # mysql/mariadb export "audit"/"error"/"general"/"slowquery"/
    # "iam-db-auth-error"; oracle exports "alert"/"audit"/"listener"/
    # "trace"/"oemagent"; sqlserver exports "agent"/"error".
    enabled_cloudwatch_logs_exports = optional(list(string), [])

    # Performance Insights: per-query performance telemetry. Free at the
    # default 7-day retention -- worth enabling on almost everything.
    performance_insights_enabled = optional(bool, false)

    # The KMS key encrypting Performance Insights data. Empty uses the
    # AWS default. Reference an AwsKmsKey key_arn output or pass a
    # literal key ARN. Cannot change after Performance Insights is first
    # enabled.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    performance_insights_kms_key_id = optional(string, "")

    # Days of Performance Insights history: 7 (free tier), 731 (2
    # years), or any multiple of 31 in between. 0 keeps the AWS default
    # (7).
    performance_insights_retention_period = optional(number, 0)

    # Enhanced Monitoring granularity in seconds: 1, 5, 10, 15, 30, or
    # 60. 0 disables (the AWS default). OS-level metrics (CPU per
    # process, memory, disk) streamed to CloudWatch Logs -- requires
    # monitoring_role_arn.
    monitoring_interval = optional(number, 0)

    # The IAM role Enhanced Monitoring publishes through (needs the
    # AmazonRDSEnhancedMonitoringRole managed policy). Required by AWS
    # when monitoring_interval is set. Reference an AwsIamRole role_arn
    # output or pass a literal ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    monitoring_role_arn = optional(string, "")

    # CloudWatch Database Insights tier: "standard" (free, included) or
    # "advanced" (paid fleet-level analysis; requires Performance
    # Insights with 465+ day retention). Empty keeps the AWS default
    # (standard).
    database_insights_mode = optional(string, "")

    # The name of an existing DB parameter group to associate. Mutually
    # exclusive with `parameters` -- either bring your own group or let
    # the module manage one from inline parameters. Empty (with
    # `parameters` also empty) keeps the engine default group.
    parameter_group_name = optional(string, "")

    # The name of an existing option group to associate. Mutually
    # exclusive with `options` -- either bring your own group or let the
    # module manage one from inline options. Empty (with `options` also
    # empty) keeps the engine default group.
    option_group_name = optional(string, "")

    # Microsoft Active Directory domain join, for Windows/Kerberos
    # authentication (SQL Server, and Kerberos auth on MySQL/PostgreSQL/
    # Oracle).
    active_directory = optional(object({
      # The directory ID of an AWS Managed Microsoft AD (d-...). Pairs with
      # domain_iam_role_name.
      domain = optional(string, "")

      # The IAM role RDS uses to join the AWS-managed directory (needs the
      # AmazonRDSDirectoryServiceAccess managed policy). Required with
      # domain.
      domain_iam_role_name = optional(string, "")

      # The fully qualified domain name of a self-managed Active Directory
      # (e.g. "corp.example.com").
      domain_fqdn = optional(string, "")

      # The organizational unit DN within the self-managed AD where the
      # computer account is created.
      domain_ou = optional(string, "")

      # The ARN of the Secrets Manager secret holding the self-managed AD
      # join credentials. An ARN reference, not the secret itself.
      domain_auth_secret_arn = optional(string, "")

      # Exactly two DNS server IPs inside the self-managed AD.
      domain_dns_ips = optional(list(string), [])
    }))

    # The license model, for engines that carry one. Per AWS's contract:
    # MySQL/MariaDB use "general-public-license"; PostgreSQL uses
    # "postgresql-license"; Oracle uses "bring-your-own-license" or
    # "license-included"; SQL Server uses "license-included" or
    # "bring-your-own-media"; Db2 uses "bring-your-own-license" or
    # "marketplace-license". Empty keeps the engine default.
    license_model = optional(string, "")

    # The character set for Oracle and SQL Server instances (e.g.
    # "AL32UTF8"). Create-time only. Empty keeps the engine default.
    character_set_name = optional(string, "")

    # The national character set (NCHAR) for Oracle instances. Create-time
    # only. Empty keeps the engine default.
    nchar_character_set_name = optional(string, "")

    # The time zone, SQL Server only (e.g. "GMT Standard Time").
    # Create-time only.
    timezone = optional(string, "")

    # The CA certificate bundle for the instance (e.g.
    # "rds-ca-rsa2048-g1"). Empty keeps the AWS default bundle.
    ca_cert_identifier = optional(string, "")

    # Use RDS Blue/Green Deployments for updates: RDS provisions a
    # synchronized green copy, applies the change there, and switches
    # over in under a minute -- near-zero-downtime engine upgrades and
    # parameter changes. MySQL, MariaDB, and PostgreSQL; not compatible
    # with read replicas of this instance being modified in the same
    # operation.
    blue_green_update_enabled = optional(bool, false)

    # Apply minor engine version patches automatically during the
    # maintenance window. AWS defaults this to true; disable only when
    # patch timing must be controlled manually.
    auto_minor_version_upgrade = optional(bool)

    # Permit engine_version changes that cross a major version. Off (the
    # default) guards against an accidental major upgrade hidden in a
    # version bump.
    allow_major_version_upgrade = optional(bool, false)

    # Apply modifications immediately instead of waiting for the next
    # maintenance window. Immediate changes can interrupt connections;
    # deferred changes wait quietly. AWS defaults to deferred.
    apply_immediately = optional(bool, false)

    # Extended support posture when the engine version leaves standard
    # support: "open-source-rds-extended-support" (AWS default -- paid
    # extended support kicks in automatically) or
    # "open-source-rds-extended-support-disabled" (the instance must be
    # upgraded before end of standard support; opts out of the extra
    # cost).
    engine_lifecycle_support = optional(string, "")

    # Upgrade the storage file system configuration on a read replica or
    # snapshot restore, when the source still runs the older 32-bit file
    # system. One-way and create/restore-scoped.
    upgrade_storage_config = optional(bool, false)

    # Create the instance by restoring a Percona XtraBackup stored in S3
    # -- the on-ramp for migrating a self-managed MySQL database into
    # RDS without a logical dump. Create-time only; mutually exclusive
    # with replicate_source_db, snapshot_identifier, and
    # restore_to_point_in_time.
    s3_import = optional(object({
      # The S3 bucket holding the backup files. Required.
      bucket_name = string

      # The key prefix of the backup files within the bucket. Empty reads
      # from the bucket root.
      bucket_prefix = optional(string, "")

      # The IAM role RDS assumes to read the backup from S3. Reference an
      # AwsIamRole role_arn output or pass a literal role ARN. Required.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      ingestion_role = string

      # The engine of the source backup. AWS accepts only "mysql" for
      # instance S3 restores. Required.
      source_engine = string

      # The version of the source engine the backup was taken from (e.g.
      # "8.0"). Required.
      source_engine_version = string
    }))

    # IAM roles the instance assumes for engine features that reach into
    # other AWS services (e.g. S3 import/export, Lambda invocation).
    # Each entry associates one role to one named engine feature -- the
    # roles own their policies; this instance only associates them. Both
    # IaC modules manage each entry as its own role-association
    # resource, so roles attach and detach without touching the
    # instance.
    iam_roles = optional(list(object({
      # The IAM role to associate. Reference an AwsIamRole role_arn output
      # or pass a literal role ARN. The role owns its policies; the
      # instance only assumes it. The role's trust policy MUST allow
      # rds.amazonaws.com to assume it -- AWS validates that server-side at
      # association time and rejects the call with InvalidParameterValue
      # ("IAM role ARN value is invalid or does not include the required
      # permissions") otherwise; no plan-time check catches it. Required.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role = string

      # The engine feature the role is linked to (e.g. "s3Import",
      # "s3Export", "Lambda", "S3_INTEGRATION" -- the valid set is
      # engine-specific). Required -- AWS rejects an instance role
      # association without a feature name. Changing it replaces the
      # association (the instance is untouched).
      feature_name = string
    })), [])

    # Instance-level engine parameters, managed as a dedicated DB
    # parameter group owned by this instance (the group is glue -- a
    # named parameter list -- so it stays folded). Mutually exclusive
    # with parameter_group_name.
    parameters = optional(list(object({
      # The parameter name (e.g. "max_connections",
      # "rds.force_ssl"). Required.
      name = string

      # The parameter value. Required.
      value = string

      # When the change lands: "immediate" (AWS default -- dynamic
      # parameters apply now) or "pending-reboot" (static parameters wait
      # for the next instance reboot).
      apply_method = optional(string, "")
    })), [])

    # Engine options (e.g. Oracle TDE/OEM, SQL Server native backup),
    # managed as a dedicated option group owned by this instance (the
    # group is glue -- a named option list -- so it stays folded).
    # Mutually exclusive with option_group_name.
    options = optional(list(object({
      # The option name (e.g. "TDE", "OEM", "SQLSERVER_BACKUP_RESTORE").
      # Required.
      option_name = string

      # Settings the option exposes, as name/value pairs (e.g. the
      # IAM_ROLE_ARN setting of SQLSERVER_BACKUP_RESTORE).
      option_settings = optional(list(object({
        # The setting name. Required.
        name = string

        # The setting value. Required.
        value = string
      })), [])

      # The port the option listens on, for options that open one (e.g.
      # OEM's 1158). 0 omits the port.
      port = optional(number, 0)

      # The option version, for options that carry one.
      version = optional(string, "")

      # Security groups granting access to the option's port. Reference
      # AwsSecurityGroup security_group_id outputs or pass literal SG IDs.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      vpc_security_group_memberships = optional(list(string), [])
    })), [])
  })
}
