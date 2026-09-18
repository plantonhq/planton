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
  description = "AwsRdsCluster specification"
  type = object({
    # The AWS region the cluster is created in. Must match the region of
    # the subnets, security groups, and KMS keys it references.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # Subnets for the cluster's DB subnet group. Provide at least two
    # subnets in DISTINCT availability zones -- AWS rejects a subnet group
    # that covers fewer than two AZs. Reference AwsSubnet subnet_id outputs
    # or pass literal subnet IDs. The module manages the subnet group
    # itself (pure glue: a named list of subnets); alternatively point
    # db_subnet_group_name at an existing group.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = optional(list(string), [])

    # Name of an existing DB subnet group to place the cluster in, instead
    # of providing subnet_ids. Changing the subnet group replaces the
    # cluster.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    db_subnet_group_name = optional(string, "")

    # Security groups attached to the cluster. Empty uses the VPC's
    # default security group (the AWS default). Reference AwsSecurityGroup
    # security_group_id outputs or pass literal SG IDs -- database ingress
    # rules belong on the referenced AwsSecurityGroup node, never inside
    # this cluster.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # Availability zones Aurora replicates storage across. Leave empty and
    # AWS picks three zones automatically -- the right call almost always,
    # because this list is create-time-only and a later change replaces the
    # cluster. Not applicable to Multi-AZ RDS clusters (AWS chooses).
    availability_zones = optional(list(string), [])

    # The network stack of the cluster: "IPV4" (AWS default when unset) or
    # "DUAL" for dual-stack IPv4+IPv6. Requires subnets with IPv6 CIDRs
    # for "DUAL".
    network_type = optional(string, "")

    # The port the cluster accepts connections on. 0 keeps the engine
    # default (3306 for MySQL-family, 5432 for PostgreSQL-family).
    port = optional(number, 0)

    # The database engine. Aurora engines ("aurora-mysql",
    # "aurora-postgresql") use shared cluster storage with `instances`
    # compute; community engines ("mysql", "postgres") create a Multi-AZ
    # RDS cluster and require db_cluster_instance_class +
    # allocated_storage_gb + iops. Changing the engine replaces the
    # cluster.
    engine = string

    # The engine version, e.g. "8.0.mysql_aurora.3.08.0" (Aurora MySQL)
    # or "16.4" (Aurora PostgreSQL). Leave empty to let AWS pick the
    # engine's current default version -- an empty pin never goes stale.
    # Minor upgrades apply in place; major upgrades additionally need
    # allow_major_version_upgrade.
    engine_version = optional(string, "")

    # How the engine provisions compute. Empty or "provisioned" (the AWS
    # default) covers both classic provisioned instances AND Aurora
    # Serverless v2 (Serverless v2 is provisioned mode + a
    # serverless_v2_scaling block + "db.serverless" instances).
    # "serverless" selects the legacy Aurora Serverless v1 engine mode,
    # where AWS owns the compute and serverless_v1_scaling applies.
    # Changing the mode replaces the cluster.
    engine_mode = optional(string, "")

    # Extended support posture when the engine version leaves standard
    # support: "open-source-rds-extended-support" (AWS default -- paid
    # extended support kicks in automatically) or
    # "open-source-rds-extended-support-disabled" (the cluster must be
    # upgraded before end of standard support; opts out of the extra
    # cost).
    engine_lifecycle_support = optional(string, "")

    # The DB instances that serve this cluster's queries -- one writer
    # (lowest promotion tier) plus any number of readers. Each entry is
    # managed as its own provider resource keyed by `name`, so scaling
    # readers in and out never touches the cluster. Empty is only valid
    # for the two shapes where AWS owns the compute: Aurora Serverless v1
    # (engine_mode "serverless") and Multi-AZ RDS clusters
    # (db_cluster_instance_class set) -- an Aurora provisioned cluster
    # with no instances stores data but cannot serve a single query.
    instances = optional(list(object({
      # Instance name, unique within the cluster. Becomes part of the AWS
      # instance identifier and the key both IaC engines manage the
      # provider resource by -- renaming an entry replaces that instance
      # (the others are untouched). Required.
      name = string

      # The instance class: a provisioned class like "db.r6g.large", or
      # "db.serverless" for an Aurora Serverless v2 instance that scales
      # within the cluster's serverless_v2_scaling bounds. Required.
      instance_class = string

      # Failover priority, 0-15 (lower is promoted first). Give the
      # largest readers tier 0/1 so a failover lands on capacity that can
      # absorb the write load. 0 is the AWS default.
      promotion_tier = optional(number, 0)

      # Pin the instance to one availability zone. Empty lets AWS place it
      # -- preferred, since AWS spreads instances across the cluster's
      # zones automatically. Create-time only.
      availability_zone = optional(string, "")

      # Give the instance a public IP. Requires public subnets; keep false
      # for anything production-shaped.
      publicly_accessible = optional(bool, false)

      # The DB (instance-level) parameter group for this instance. Empty
      # keeps the engine default group.
      db_parameter_group_name = optional(string, "")

      # Apply minor engine version patches automatically during the
      # maintenance window. AWS defaults this to true; disable only when
      # patch timing must be controlled manually.
      auto_minor_version_upgrade = optional(bool)

      # Per-instance Performance Insights override. Unset inherits the
      # cluster-level setting.
      performance_insights_enabled = optional(bool)

      # Enhanced Monitoring granularity for this instance in seconds: 1,
      # 5, 10, 15, 30, or 60. 0 disables. Uses the cluster's
      # monitoring_role_arn.
      monitoring_interval = optional(number, 0)

      # The CA certificate bundle for this instance (e.g.
      # "rds-ca-rsa2048-g1"). Empty keeps the AWS default.
      ca_cert_identifier = optional(string, "")

      # Copy this instance's tags onto its snapshots.
      copy_tags_to_snapshot = optional(bool, false)

      # The KMS key encrypting this instance's Performance Insights data.
      # Empty uses the AWS default. Reference an AwsKmsKey key_arn output
      # or pass a literal key ARN. Cannot change after Performance
      # Insights is first enabled on the instance.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      performance_insights_kms_key_id = optional(string, "")

      # Days of Performance Insights history for THIS instance: 7 (free
      # tier), 731 (2 years), or any multiple of 31 in between. 0 inherits
      # the cluster-level setting (or the AWS default of 7).
      performance_insights_retention_period = optional(number, 0)

      # The daily backup window for THIS instance in UTC, format
      # "hh24:mi-hh24:mi". Empty inherits the cluster's window -- stagger
      # per-instance windows when backups must not contend.
      preferred_backup_window = optional(string, "")

      # The weekly maintenance window for THIS instance in UTC, format
      # "ddd:hh24:mi-ddd:hh24:mi". Empty inherits scheduling from AWS --
      # stagger per-instance windows so readers never patch
      # simultaneously.
      preferred_maintenance_window = optional(string, "")

      # The IAM role Enhanced Monitoring publishes through for THIS
      # instance (needs the AmazonRDSEnhancedMonitoringRole managed
      # policy). Empty uses the cluster's monitoring_role_arn. Reference
      # an AwsIamRole role_arn output or pass a literal ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      monitoring_role_arn = optional(string, "")

      # Apply modifications to THIS instance immediately instead of
      # waiting for its next maintenance window. AWS defaults to deferred.
      apply_immediately = optional(bool, false)
    })), [])

    # Aurora Serverless v2 capacity bounds. Requires provisioned engine
    # mode and instances of class "db.serverless" -- each such instance
    # scales independently within these bounds. min_capacity 0 enables
    # automatic pause (scale-to-zero) after seconds_until_auto_pause of
    # idleness.
    serverless_v2_scaling = optional(object({
      # Minimum ACUs, 0-256 in 0.5 steps. 0 enables automatic pause: the
      # instance suspends after seconds_until_auto_pause of idleness and
      # costs nothing while paused (storage still billed) -- resumed on the
      # next connection in ~15 seconds. Dev/test clusters want 0; latency-
      # sensitive production wants >= 0.5 to never pause.
      min_capacity = optional(number, 0)

      # Maximum ACUs, 1-256. The hard spend/performance ceiling per
      # instance. Required.
      max_capacity = number

      # Seconds of idleness before an instance auto-pauses, 300-86400.
      # Only meaningful when min_capacity is 0. 0 keeps the AWS default
      # (300 -- five minutes).
      seconds_until_auto_pause = optional(number, 0)
    }))

    # Aurora Serverless v1 autoscaling configuration. Only valid with
    # engine_mode "serverless" (the legacy serverless offering) -- prefer
    # Serverless v2 for new designs.
    serverless_v1_scaling = optional(object({
      # Pause compute after seconds_until_auto_pause of idleness. AWS
      # defaults this to true -- the cost model Serverless v1 exists for.
      auto_pause = optional(bool)

      # Minimum ACUs (whole units, engine-specific valid set). 0 keeps the
      # AWS default (1).
      min_capacity = optional(number, 0)

      # Maximum ACUs (whole units). 0 keeps the AWS default (16).
      max_capacity = optional(number, 0)

      # Seconds of idleness before pausing, 300-86400. 0 keeps the AWS
      # default (300).
      seconds_until_auto_pause = optional(number, 0)

      # Seconds a scaling operation waits for a safe scaling point before
      # timeout_action applies, 60-600. 0 keeps the AWS default (300).
      seconds_before_timeout = optional(number, 0)

      # What happens when scaling times out: "RollbackCapacityChange" (AWS
      # default -- keep current capacity) or "ForceApplyCapacityChange"
      # (scale anyway, dropping connections that block it).
      timeout_action = optional(string, "")
    }))

    # The instance class for a Multi-AZ RDS cluster (community
    # mysql/postgres engines), e.g. "db.m6gd.large". Setting it selects
    # the Multi-AZ cluster shape: AWS manages one writer and two readers
    # internally, `instances` stays empty, and allocated_storage_gb + iops
    # are required.
    db_cluster_instance_class = optional(string, "")

    # Provisioned storage in GiB for a Multi-AZ RDS cluster. Not
    # applicable to Aurora (Aurora storage grows automatically).
    allocated_storage_gb = optional(number, 0)

    # Provisioned IOPS for a Multi-AZ RDS cluster (required by AWS for
    # io1/io2/gp3 cluster storage). Not applicable to Aurora.
    iops = optional(number, 0)

    # The storage type. Aurora engines: "" (standard, billed per I/O) or
    # "aurora-iopt1" (I/O-Optimized, up to ~40% cheaper for I/O-heavy
    # workloads, switchable once per 30 days). Multi-AZ RDS clusters:
    # "io1", "io2", or "gp3".
    storage_type = optional(string, "")

    # The name of the initial database AWS creates in the cluster. Empty
    # creates no database (create one from SQL later). Create-time only.
    database_name = optional(string, "")

    # The master username. Required for a brand-new cluster -- AWS has no
    # default and rejects a blank value at CreateDBCluster. Only clusters
    # that inherit credentials from a source (snapshot restore,
    # point-in-time restore, a replication source, or joining an existing
    # global database as a secondary) leave it empty. Avoid the engine's
    # reserved names (e.g. "rdsadmin"). Create-time only -- changing it
    # replaces the cluster.
    master_username = optional(string, "")

    # Let AWS manage the master password in Secrets Manager: AWS
    # generates it, stores it, rotates it on schedule, and no secret ever
    # touches this manifest or the IaC state. The managed secret's ARN is
    # exported as the master_user_secret_arn output. Mutually exclusive
    # with master_password -- and the recommended posture.
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
    master_password = optional(string, "")

    # Encrypt cluster storage at rest. Strongly recommended -- and
    # create-time only: an unencrypted cluster cannot be encrypted later
    # (requires a snapshot-restore migration).
    storage_encrypted = optional(bool, false)

    # The KMS key for storage encryption when storage_encrypted is true.
    # Empty uses the AWS-managed aws/rds key. Reference an AwsKmsKey
    # key_arn output or pass a literal key ARN. Create-time only.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # Days automated backups are retained, 1-35. 0 keeps the AWS default
    # (1 day). Aurora backups are continuous -- this window bounds
    # point-in-time recovery, so production clusters typically want 7+.
    backup_retention_period = optional(number, 0)

    # The daily backup window in UTC, format "hh24:mi-hh24:mi" (e.g.
    # "04:00-05:00"). Empty lets AWS assign one. Must not overlap the
    # maintenance window.
    preferred_backup_window = optional(string, "")

    # The weekly maintenance window in UTC, format
    # "ddd:hh24:mi-ddd:hh24:mi" (e.g. "sun:05:00-sun:06:00"). Empty lets
    # AWS assign one.
    preferred_maintenance_window = optional(string, "")

    # Copy the cluster's tags onto automated and manual snapshots.
    copy_tags_to_snapshot = optional(bool, false)

    # Remove automated backups immediately when the cluster is deleted.
    # AWS defaults this to true; set false to retain the backups for the
    # remainder of their retention window after deletion -- the last line
    # of defense against a mistaken teardown.
    delete_automated_backups = optional(bool)

    # Skip the final snapshot when the cluster is deleted. When false
    # (the safe default), final_snapshot_identifier must be set -- AWS
    # refuses to delete without knowing the snapshot name.
    skip_final_snapshot = optional(bool, false)

    # The name for the final snapshot taken on deletion. Required when
    # skip_final_snapshot is false. Must start with a letter, contain
    # only letters, numbers, and hyphens, with no consecutive or
    # trailing hyphens -- AWS's snapshot-identifier rules.
    final_snapshot_identifier = optional(string, "")

    # Refuse deletion of the cluster while enabled. Turn this on for
    # anything holding data you cannot recreate -- deletion then requires
    # an explicit two-step (disable, delete).
    deletion_protection = optional(bool, false)

    # Aurora MySQL backtrack window in seconds, 0-259200 (72 hours).
    # Backtrack rewinds the cluster in place (no restore, no new
    # endpoint) -- the fastest "undo" for fat-fingered writes. 0 disables.
    # Aurora MySQL only; enabling on an existing cluster is not supported
    # by AWS.
    backtrack_window_seconds = optional(number, 0)

    # Map IAM identities to database users -- connect with short-lived
    # IAM auth tokens instead of passwords.
    iam_database_authentication_enabled = optional(bool, false)

    # IAM roles the cluster assumes for engine features that reach into
    # other AWS services (S3 import/export, Lambda invocation, Comprehend
    # /SageMaker for ML functions). Each entry associates one role,
    # optionally linked to a specific engine feature by feature_name --
    # the roles own their policies; this cluster only associates them.
    # Both IaC modules manage each entry as its own role-association
    # resource (never the cluster's inline role list, which cannot carry
    # feature names and conflicts with association resources), so roles
    # attach and detach without touching the cluster.
    iam_roles = optional(list(object({
      # The IAM role to associate. Reference an AwsIamRole role_arn output
      # or pass a literal role ARN. The role owns its policies; the
      # cluster only assumes it. The role's trust policy MUST allow
      # rds.amazonaws.com to assume it -- AWS validates that server-side at
      # association time and rejects the call with InvalidParameterValue
      # ("IAM role ARN value is invalid or does not include the required
      # permissions") otherwise; no plan-time check catches it. Required.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role = string

      # The engine feature the role is linked to (e.g. "s3Import",
      # "s3Export", "Lambda", "SageMaker", "Comprehend"). Empty associates
      # the role without a feature link; AWS requires the name whenever
      # the role powers a specific engine capability. Changing it replaces
      # the association (the cluster is untouched).
      feature_name = optional(string, "")
    })), [])

    # Enable the RDS Data API: SQL over HTTPS with IAM auth, no
    # persistent connections -- the natural fit for Lambda and other
    # connection-averse callers. Aurora PostgreSQL and Aurora MySQL
    # (Serverless v2 / provisioned), plus Serverless v1.
    enable_http_endpoint = optional(bool, false)

    # Database log types to export to CloudWatch Logs. MySQL family
    # ("aurora-mysql", "mysql"): "audit", "error", "general", "slowquery".
    # PostgreSQL family ("aurora-postgresql", "postgres"): "postgresql",
    # "upgrade". Both families also accept "iam-db-auth-error", and
    # Multi-AZ RDS clusters accept "instance".
    enabled_cloudwatch_logs_exports = optional(list(string), [])

    # Cluster-level Performance Insights (per-query performance
    # telemetry). On Aurora, per-instance settings in `instances` can
    # override this. Free at the default 7-day retention.
    performance_insights_enabled = optional(bool, false)

    # The KMS key encrypting Performance Insights data. Empty uses the
    # AWS default. Reference an AwsKmsKey key_arn output or pass a
    # literal key ARN. Cannot change after Performance Insights is first
    # enabled.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    performance_insights_kms_key_id = optional(string, "")

    # Days of Performance Insights history: 7 (free tier), 731 (2 years),
    # or any multiple of 31 in between. 0 keeps the AWS default (7).
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

    # Restore the cluster from an existing cluster snapshot (name or ARN)
    # at create time. Create-time only; mutually exclusive with
    # restore_to_point_in_time.
    snapshot_identifier = optional(string, "")

    # Clone or restore this cluster from another cluster's continuous
    # backup at create time -- point-in-time recovery as a first-class
    # create shape. Create-time only; mutually exclusive with
    # snapshot_identifier.
    restore_to_point_in_time = optional(object({
      # The source cluster identifier (name or ARN). Exactly one of
      # source_cluster_identifier or source_cluster_resource_id must be
      # set.
      source_cluster_identifier = optional(string, "")

      # The source cluster's immutable resource ID (cluster_resource_id
      # output) -- survives identifier renames and points at deleted
      # clusters' retained backups. Exactly one of the two source fields
      # must be set.
      source_cluster_resource_id = optional(string, "")

      # The UTC timestamp to restore to, RFC3339 (e.g.
      # "2026-07-01T09:45:00Z"). Mutually exclusive with
      # use_latest_restorable_time.
      restore_to_time = optional(string, "")

      # Restore to the most recent recoverable moment. Mutually exclusive
      # with restore_to_time.
      use_latest_restorable_time = optional(bool, false)

      # "full-copy" (independent storage, AWS default) or "copy-on-write"
      # (an Aurora fast clone -- shares storage with the source and only
      # pays for divergence; ideal for prod-data staging environments).
      restore_type = optional(string, "")
    }))

    # Make this cluster a cross-region (or cross-account) read replica of
    # the given source cluster ARN. Promote by clearing the field.
    replication_source_identifier = optional(string, "")

    # The region of the replication source, required by AWS when creating
    # an encrypted cross-region replica (it scopes the KMS re-encryption).
    # Create-time only.
    source_region = optional(string, "")

    # Join an Aurora Global Database: the identifier of the
    # aws_rds_global_cluster this cluster participates in. The first
    # cluster joined becomes the global writer; clusters added afterwards
    # become read-only secondaries.
    global_cluster_identifier = optional(string, "")

    # Let secondary-region endpoints accept writes and forward them to
    # the global writer (Aurora Global Database only). Apps in secondary
    # regions get a single connection string for reads AND writes.
    enable_global_write_forwarding = optional(bool, false)

    # Let reader instances in THIS cluster accept writes and forward them
    # to the writer -- one endpoint for the whole cluster without a
    # client-side split. Aurora MySQL 3.04+ / Aurora PostgreSQL 16.4+.
    enable_local_write_forwarding = optional(bool, false)

    # The name of an existing cluster parameter group to use. Mutually
    # exclusive with `parameters` -- either bring your own group or let
    # the module manage one from inline parameters.
    db_cluster_parameter_group_name = optional(string, "")

    # Cluster-level engine parameters, managed as a dedicated parameter
    # group owned by this cluster (the group is glue -- a named parameter
    # list -- so it stays folded). Mutually exclusive with
    # db_cluster_parameter_group_name.
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

    # The name of the DB (instance-level) parameter group applied to
    # cluster instances DURING a major engine version upgrade. Only
    # consulted when engine_version changes across a major version.
    db_instance_parameter_group_name = optional(string, "")

    # The CA certificate bundle for the cluster's instances (e.g.
    # "rds-ca-rsa2048-g1"). Empty keeps the AWS default bundle.
    ca_certificate_identifier = optional(string, "")

    # Apply modifications immediately instead of waiting for the next
    # maintenance window. Immediate changes can interrupt connections
    # (e.g. scaling); deferred changes wait quietly. AWS defaults to
    # deferred.
    apply_immediately = optional(bool, false)

    # Permit engine_version changes that cross a major version. Off (the
    # default) guards against an accidental major upgrade hidden in a
    # version bump.
    allow_major_version_upgrade = optional(bool, false)

    # Join the cluster to an AWS Managed Microsoft AD directory (d-...)
    # for Kerberos authentication (Aurora MySQL and Aurora PostgreSQL).
    # Pairs with domain_iam_role_name. (Self-managed AD is an
    # instance-kind shape -- clusters only support the managed
    # directory.)
    domain = optional(string, "")

    # The name of the IAM role RDS uses to join the managed directory
    # (needs the AmazonRDSDirectoryServiceAccess managed policy).
    # Required with domain.
    domain_iam_role_name = optional(string, "")

    # Apply minor engine version patches to the cluster automatically
    # during the maintenance window. AWS defaults this to true; disable
    # only when patch timing must be controlled manually. Per-instance
    # auto_minor_version_upgrade in `instances` overrides this for that
    # instance.
    auto_minor_version_upgrade = optional(bool)

    # Create the cluster by restoring a Percona XtraBackup stored in S3
    # -- the on-ramp for migrating a self-managed MySQL database into
    # Aurora MySQL without a logical dump. Create-time only; mutually
    # exclusive with snapshot_identifier and restore_to_point_in_time.
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
      # cluster S3 restores. Required.
      source_engine = string

      # The version of the source engine the backup was taken from (e.g.
      # "8.0"). Required.
      source_engine_version = string
    }))

    # Custom cluster endpoints -- stable DNS names scoped to a chosen
    # subset of the cluster's instances (e.g. an analytics endpoint over
    # the big readers). Each entry is managed as its own provider
    # resource keyed by `name`, so endpoints come and go without
    # touching the cluster.
    custom_endpoints = optional(list(object({
      # Endpoint name, unique within the cluster (lowercase letters,
      # digits, hyphens; starts with a letter). Becomes the DNS-visible
      # endpoint identifier and the key both IaC engines manage the
      # provider resource by -- renaming an entry replaces that endpoint
      # (the cluster is untouched). Required.
      name = string

      # Which instances the endpoint fronts: "READER" (reader instances
      # only) or "ANY" (all instances). Required.
      type = string

      # Pin the endpoint to exactly these instances, by their
      # spec.instances entry names. Mutually exclusive with
      # excluded_members. Both empty fronts every instance of the
      # endpoint's type.
      static_members = optional(list(string), [])

      # Front every instance of the endpoint's type EXCEPT these, by their
      # spec.instances entry names. Mutually exclusive with
      # static_members.
      excluded_members = optional(list(string), [])
    })), [])

    # Stream every audited database event to a dedicated Kinesis stream
    # (a Database Activity Stream), encrypted with the given KMS key --
    # the compliance-grade audit feed consumed by GuardDuty RDS
    # Protection and partner SIEMs. Aurora engines only. The stream's
    # Kinesis name is exported as the activity_stream_kinesis_stream_name
    # output.
    activity_stream = optional(object({
      # Delivery mode: "sync" (the database blocks until the event is
      # durably recorded -- audit-grade guarantees at a latency cost) or
      # "async" (the database never waits; rare event loss is possible).
      # Required.
      mode = string

      # The KMS key that encrypts the activity stream. Reference an
      # AwsKmsKey key_arn output or pass a literal key ARN. Required.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_id = string

      # Also capture the engine's native audit fields in the stream
      # (engine-version dependent; leave false unless the consumer needs
      # them).
      engine_native_audit_fields_included = optional(bool, false)
    }))
  })
}
