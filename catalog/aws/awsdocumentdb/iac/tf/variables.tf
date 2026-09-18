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
  description = "AwsDocumentDb specification"
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

    # Availability zones the cluster's storage is replicated across. Leave
    # empty and AWS picks three zones automatically -- the right call
    # almost always, because this list is create-time-only and a later
    # change replaces the cluster.
    availability_zones = optional(list(string), [])

    # The network stack of the cluster: "IPV4" (AWS default when unset) or
    # "DUAL" for dual-stack IPv4+IPv6. Requires subnets with IPv6 CIDRs
    # for "DUAL".
    network_type = optional(string, "")

    # The port the cluster accepts connections on. 0 keeps the AWS default
    # (27017 -- the MongoDB convention). DocumentDB rejects ports below
    # 1150. Create-time only -- changing the port replaces the cluster.
    port = optional(number, 0)

    # The DocumentDB engine version, e.g. "5.0.0". Leave empty to let AWS
    # pick the current default version -- an empty pin never goes stale.
    # Minor upgrades apply in place; major upgrades additionally need
    # allow_major_version_upgrade.
    engine_version = optional(string, "")

    # The storage I/O model: "" or "standard" (billed per I/O, the AWS
    # default) or "iopt1" (I/O-Optimized -- predictable pricing that pays
    # off for I/O-heavy workloads; switchable once per 30 days).
    storage_type = optional(string, "")

    # The DB instances that serve this cluster's queries -- one writer
    # (lowest promotion tier) plus any number of readers. Each entry is
    # managed as its own provider resource keyed by `name`, so scaling
    # readers in and out never touches the cluster. Empty is only valid
    # for headless shapes that attach compute later: a snapshot or
    # point-in-time restore, or a cluster joined to a global cluster --
    # a regular cluster with no instances stores data but cannot serve a
    # single query.
    instances = optional(list(object({
      # Instance name, unique within the cluster. Becomes part of the AWS
      # instance identifier and the key both IaC engines manage the
      # provider resource by -- renaming an entry replaces that instance
      # (the others are untouched). Required.
      name = string

      # The instance class: a provisioned class like "db.r6g.large", or
      # "db.serverless" for a DocumentDB Serverless instance that scales
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

      # Apply minor engine version patches automatically during the
      # maintenance window. AWS defaults this to true; disable only when
      # patch timing must be controlled manually.
      auto_minor_version_upgrade = optional(bool)

      # Per-instance Performance Insights (per-query performance
      # telemetry). Free at the default 7-day retention. DocumentDB scopes
      # this to the instance -- there is no cluster-level setting.
      performance_insights_enabled = optional(bool, false)

      # The KMS key encrypting this instance's Performance Insights data.
      # Empty uses the AWS default. Reference an AwsKmsKey key_arn output
      # or pass a literal key ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      performance_insights_kms_key_id = optional(string, "")

      # The weekly maintenance window for THIS instance in UTC, format
      # "ddd:hh24:mi-ddd:hh24:mi". Empty inherits scheduling from AWS --
      # stagger per-instance windows so readers never patch simultaneously.
      preferred_maintenance_window = optional(string, "")

      # The CA certificate bundle for this instance (e.g.
      # "rds-ca-rsa2048-g1"). Empty keeps the AWS default.
      ca_cert_identifier = optional(string, "")

      # Copy this instance's tags onto its snapshots.
      copy_tags_to_snapshot = optional(bool, false)

      # Restart the instance when its CA certificate rotates. Unset keeps
      # the AWS default (true -- rotation restarts the instance so the new
      # certificate is served immediately); false defers the restart to
      # the next maintenance action, for workloads that cannot absorb an
      # unscheduled restart. Tri-state on purpose: only an explicit value
      # is sent to AWS.
      certificate_rotation_restart = optional(bool)

      # Apply modifications to THIS instance immediately instead of
      # waiting for its next maintenance window. AWS defaults to deferred.
      apply_immediately = optional(bool, false)
    })), [])

    # DocumentDB Serverless capacity bounds. When set, every instance in
    # `instances` must use class "db.serverless" -- each such instance
    # scales independently within these bounds. Adding or modifying this
    # block is an in-place update; REMOVING it from a live cluster
    # replaces the cluster (AWS cannot switch a cluster off serverless).
    serverless_v2_scaling = optional(object({
      # Minimum DCUs, 0.5-256 in half-step multiples. The floor each
      # serverless instance never scales below -- also the idle cost floor
      # (DocumentDB Serverless does not pause to zero). Required.
      min_capacity = number

      # Maximum DCUs, 1-256 in half-step multiples. The hard
      # spend/performance ceiling per instance. Required.
      max_capacity = number
    }))

    # The master username. Required for a brand-new cluster -- AWS has no
    # default and rejects a blank value at CreateDBCluster. Only clusters
    # that inherit credentials from a source (snapshot restore,
    # point-in-time restore, or joining a global cluster) leave it empty.
    # Create-time only -- changing it replaces the cluster.
    master_username = optional(string, "")

    # Let AWS manage the master password in Secrets Manager: AWS
    # generates it, stores it, rotates it on schedule, and no secret ever
    # touches this manifest or the IaC state. The managed secret's ARN is
    # exported as the master_user_secret_arn output. Mutually exclusive
    # with master_password -- and the recommended posture.
    manage_master_user_password = optional(bool, false)

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
    # (1 day). Backups are continuous -- this window bounds point-in-time
    # recovery, so production clusters typically want 7+.
    backup_retention_period = optional(number, 0)

    # The daily backup window in UTC, format "hh24:mi-hh24:mi" (e.g.
    # "04:00-05:00"). Empty lets AWS assign one. Must not overlap the
    # maintenance window.
    preferred_backup_window = optional(string, "")

    # The weekly maintenance window in UTC, format
    # "ddd:hh24:mi-ddd:hh24:mi" (e.g. "sun:05:00-sun:06:00"). Empty lets
    # AWS assign one.
    preferred_maintenance_window = optional(string, "")

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

    # Database log types to export to CloudWatch Logs: "audit" (DDL and
    # authentication events; requires the audit_logs cluster parameter)
    # and "profiler" (slow-operation profiling; requires the profiler
    # cluster parameters).
    enabled_cloudwatch_logs_exports = optional(list(string), [])

    # Restore the cluster from an existing cluster snapshot (name or ARN)
    # at create time. Create-time only; mutually exclusive with
    # restore_to_point_in_time.
    snapshot_identifier = optional(string, "")

    # Clone or restore this cluster from another cluster's continuous
    # backup at create time -- point-in-time recovery as a first-class
    # create shape. Create-time only; mutually exclusive with
    # snapshot_identifier.
    restore_to_point_in_time = optional(object({
      # The source cluster identifier (name or ARN). Required.
      source_cluster_identifier = string

      # The UTC timestamp to restore to, RFC3339 (e.g.
      # "2026-07-01T09:45:00Z"). Mutually exclusive with
      # use_latest_restorable_time.
      restore_to_time = optional(string, "")

      # Restore to the most recent recoverable moment. Mutually exclusive
      # with restore_to_time.
      use_latest_restorable_time = optional(bool, false)

      # "full-copy" (independent storage, AWS default) or "copy-on-write"
      # (a fast clone -- shares storage with the source and only pays for
      # divergence; ideal for prod-data staging environments).
      restore_type = optional(string, "")
    }))

    # Join a DocumentDB global cluster: the identifier of the
    # aws_docdb_global_cluster this cluster participates in. The first
    # cluster joined becomes the global writer; clusters added afterwards
    # become read-only secondaries that inherit credentials from the
    # primary.
    global_cluster_identifier = optional(string, "")

    # The name of an existing cluster parameter group to use. Mutually
    # exclusive with `parameters` -- either bring your own group or let
    # the module manage one from inline parameters.
    db_cluster_parameter_group_name = optional(string, "")

    # Cluster-level engine parameters (e.g. "audit_logs", "tls",
    # "ttl_monitor"), managed as a dedicated parameter group owned by this
    # cluster (the group is glue -- a named parameter list -- so it stays
    # folded). Mutually exclusive with db_cluster_parameter_group_name.
    parameters = optional(list(object({
      # The parameter name (e.g. "audit_logs", "tls",
      # "ttl_monitor"). Required.
      name = string

      # The parameter value. Required.
      value = string

      # When the change lands: "immediate" (AWS default -- dynamic
      # parameters apply now) or "pending-reboot" (static parameters wait
      # for the next instance reboot).
      apply_method = optional(string, "")
    })), [])

    # Apply modifications immediately instead of waiting for the next
    # maintenance window. Immediate changes can interrupt connections;
    # deferred changes wait quietly. AWS defaults to deferred.
    apply_immediately = optional(bool, false)

    # Permit engine_version changes that cross a major version. Off (the
    # default) guards against an accidental major upgrade hidden in a
    # version bump.
    allow_major_version_upgrade = optional(bool, false)
  })
}
