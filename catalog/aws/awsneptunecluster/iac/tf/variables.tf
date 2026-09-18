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
  description = "AwsNeptuneCluster specification"
  type = object({
    # The AWS region the cluster is created in. Must match the region of
    # the subnets, security groups, and KMS keys it references.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # Subnets for the cluster's Neptune subnet group. Provide at least
    # two subnets in DISTINCT availability zones -- AWS rejects a subnet
    # group that covers fewer than two AZs. Reference AwsSubnet subnet_id
    # outputs or pass literal subnet IDs. The module manages the subnet
    # group itself (pure glue: a named list of subnets); alternatively
    # point neptune_subnet_group_name at an existing group.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = optional(list(string), [])

    # Name of an existing Neptune subnet group to place the cluster in,
    # instead of providing subnet_ids. Changing the subnet group replaces
    # the cluster.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    neptune_subnet_group_name = optional(string, "")

    # Security groups attached to the cluster. Empty uses the VPC's
    # default security group (the AWS default). Reference AwsSecurityGroup
    # security_group_id outputs or pass literal SG IDs -- database ingress
    # rules belong on the referenced AwsSecurityGroup node, never inside
    # this cluster.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # Availability zones the cluster's storage is replicated across (at
    # most three). Leave empty and AWS picks the zones automatically --
    # the right call almost always, because this list is create-time-only
    # and a later change replaces the cluster.
    availability_zones = optional(list(string), [])

    # The port the cluster accepts connections on. 0 keeps the AWS
    # default (8182 -- the Neptune convention). Create-time only --
    # changing the port replaces the cluster. Both engines also pin the
    # cluster's port on every instance resource so instance state always
    # converges with the cluster (instances have no port of their own --
    # they listen on the cluster's).
    port = optional(number, 0)

    # The Neptune engine version, e.g. "1.4.5.1". Leave empty to let AWS
    # pick the current default version -- an empty pin never goes stale.
    # Minor upgrades apply in place; major upgrades additionally need
    # allow_major_version_upgrade and neptune_instance_parameter_group_name.
    engine_version = optional(string, "")

    # The storage I/O model: "" or "standard" (billed per I/O, the AWS
    # default) or "iopt1" (I/O-Optimized -- predictable pricing that pays
    # off for I/O-heavy workloads; requires engine version 1.3+ and is
    # switchable once per 30 days).
    storage_type = optional(string, "")

    # The DB instances that serve this cluster's queries -- one writer
    # (lowest promotion tier) plus any number of readers. Each entry is
    # managed as its own provider resource keyed by `name`, so scaling
    # readers in and out never touches the cluster. Empty is only valid
    # for headless shapes that attach compute later: a snapshot restore,
    # a cross-region replica, or a cluster joined to a global cluster --
    # a regular cluster with no instances stores data but cannot serve a
    # single query.
    instances = optional(list(object({
      # Instance name, unique within the cluster. Becomes part of the AWS
      # instance identifier and the key both IaC engines manage the
      # provider resource by -- renaming an entry replaces that instance
      # (the others are untouched). Required.
      name = string

      # The instance class: a provisioned class like "db.r6g.large", or
      # "db.serverless" for a Neptune Serverless instance that scales
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
      # for anything production-shaped. Create-time only.
      publicly_accessible = optional(bool, false)

      # The DB (instance-level) parameter group for this instance. Empty
      # keeps the engine default group.
      neptune_parameter_group_name = optional(string, "")

      # Apply minor engine version patches automatically during the
      # maintenance window. AWS defaults this to true; disable only when
      # patch timing must be controlled manually.
      auto_minor_version_upgrade = optional(bool)

      # The weekly maintenance window for THIS instance in UTC, format
      # "ddd:hh24:mi-ddd:hh24:mi". Empty inherits scheduling from AWS --
      # stagger per-instance windows so readers never patch simultaneously.
      preferred_maintenance_window = optional(string, "")
    })), [])

    # Neptune Serverless capacity bounds. When set, every instance in
    # `instances` must use class "db.serverless" -- each such instance
    # scales independently within these bounds.
    serverless_v2_scaling = optional(object({
      # Minimum NCUs, 1-128. The floor each serverless instance never
      # scales below -- also the idle cost floor (Neptune Serverless does
      # not pause to zero). Required.
      min_capacity = number

      # Maximum NCUs, 1-128. The hard spend/performance ceiling per
      # instance. Required.
      max_capacity = number
    }))

    # Encrypt cluster storage at rest. Defaults to TRUE (secure by
    # default) -- and create-time only: an unencrypted cluster cannot be
    # encrypted later (requires a snapshot-restore migration). Set an
    # explicit false only where encryption cannot apply: restoring an
    # UNENCRYPTED snapshot, or joining a global cluster whose primary is
    # unencrypted (encryption must match the source in both cases).
    storage_encrypted = optional(bool)

    # The KMS key for storage encryption when storage_encrypted is true.
    # Empty uses the AWS-managed aws/rds key. Reference an AwsKmsKey
    # key_arn output or pass a literal key ARN. Create-time only.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # Require SigV4-signed requests from IAM identities to query the
    # database -- Neptune's only credential mechanism (there is no master
    # username/password). Off relies purely on network reachability.
    # (The AWS default is off; an explicit false at create time is
    # indistinguishable from omitting it -- the provider only sends the
    # flag on later changes.)
    iam_database_authentication_enabled = optional(bool, false)

    # IAM roles the cluster assumes for engine features that reach into
    # other AWS services (the Neptune bulk loader reading from S3,
    # Neptune ML with SageMaker). Reference AwsIamRole role_arn outputs
    # or pass literal ARNs -- the roles own their policies; this cluster
    # only associates them.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    iam_roles = optional(list(string), [])

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

    # Copy the cluster's tags onto automated and manual snapshots.
    copy_tags_to_snapshot = optional(bool, false)

    # Skip the final snapshot when the cluster is deleted. When false
    # (the safe default), final_snapshot_identifier must be set -- AWS
    # refuses to delete without knowing the snapshot name. Also forwarded
    # to each cluster instance's delete (instance-level final snapshots
    # do not apply to cluster members, but the flag keeps teardown intent
    # consistent).
    skip_final_snapshot = optional(bool, false)

    # The name for the final snapshot taken on deletion. Required when
    # skip_final_snapshot is false. Letters, digits, and hyphens; no
    # double or trailing hyphen (the provider validates this at plan
    # time).
    final_snapshot_identifier = optional(string, "")

    # Refuse deletion of the cluster while enabled. Turn this on for
    # anything holding data you cannot recreate -- deletion then requires
    # an explicit two-step (disable, delete).
    deletion_protection = optional(bool, false)

    # Database log types to export to CloudWatch Logs: "audit"
    # (connection and query audit trail; requires the
    # neptune_enable_audit_log cluster parameter) and "slowquery"
    # (queries exceeding the slow-query threshold parameters).
    enabled_cloudwatch_logs_exports = optional(list(string), [])

    # Restore the cluster from an existing cluster snapshot (name or ARN)
    # at create time. Create-time only.
    snapshot_identifier = optional(string, "")

    # Make this cluster a read replica of the given source cluster ARN.
    # Promote by clearing the field.
    replication_source_identifier = optional(string, "")

    # Join a Neptune global database: the identifier of the global
    # cluster this cluster participates in. The first cluster joined
    # becomes the global writer; clusters added afterwards become
    # read-only secondaries. Letter-first; letters, digits, and hyphens;
    # no double or trailing hyphen.
    global_cluster_identifier = optional(string, "")

    # The name of an existing cluster parameter group to use. Mutually
    # exclusive with `parameters` -- either bring your own group or let
    # the module manage one from inline parameters.
    neptune_cluster_parameter_group_name = optional(string, "")

    # Cluster-level engine parameters (e.g. "neptune_enable_audit_log",
    # "neptune_query_timeout"), managed as a dedicated parameter group
    # owned by this cluster (the group is glue -- a named parameter list
    # -- so it stays folded). Mutually exclusive with
    # neptune_cluster_parameter_group_name.
    parameters = optional(list(object({
      # The parameter name (e.g. "neptune_enable_audit_log",
      # "neptune_query_timeout"). Required.
      name = string

      # The parameter value. Required.
      value = string

      # When the change lands: "immediate" (dynamic parameters apply now)
      # or "pending-reboot" (static parameters wait for the next instance
      # reboot). Empty defers to the provider default, which is
      # "pending-reboot" at the pinned provider version -- set "immediate"
      # explicitly when a dynamic parameter should land right away.
      apply_method = optional(string, "")
    })), [])

    # The name of the DB (instance-level) parameter group applied to
    # cluster instances DURING a major engine version upgrade. AWS
    # requires it when engine_version changes across a major version.
    neptune_instance_parameter_group_name = optional(string, "")

    # Apply modifications immediately instead of waiting for the next
    # maintenance window. Immediate changes can interrupt connections;
    # deferred changes wait quietly. AWS defaults to deferred. Applies to
    # cluster-scope changes AND to instance-scope changes (class resizes,
    # per-instance windows, parameter group switches) -- both engines
    # forward it to every cluster instance.
    apply_immediately = optional(bool, false)

    # Permit engine_version changes that cross a major version. Off (the
    # default) guards against an accidental major upgrade hidden in a
    # version bump.
    allow_major_version_upgrade = optional(bool, false)

    # Custom cluster endpoints: stable DNS names over chosen subsets of
    # the cluster's instances (e.g. an analytics endpoint fronting only
    # the big readers). Each entry is managed as its own provider
    # resource keyed by `name`, so endpoints come and go without touching
    # the cluster. Per-endpoint addresses are exported in the
    # custom_endpoint_addresses output keyed by name.
    custom_endpoints = optional(list(object({
      # Endpoint name, unique within the cluster (lowercase letters,
      # digits, hyphens; starts with a letter, no double or trailing
      # hyphen). Becomes the DNS-visible endpoint identifier and the key
      # both IaC engines manage the provider resource by -- renaming an
      # entry replaces that endpoint (the cluster is untouched). Required.
      name = string

      # Which instances the endpoint fronts: "READER" (reader instances
      # only), "WRITER" (the writer), or "ANY" (all instances). Required.
      endpoint_type = string

      # Pin the endpoint to exactly these instances, by their
      # spec.instances entry names. Mutually exclusive with
      # excluded_members. Both empty fronts every instance of the
      # endpoint's type.
      static_members = optional(list(string), [])

      # Front every instance of the endpoint's type EXCEPT these, by their
      # spec.instances entry names. Mutually exclusive with static_members.
      excluded_members = optional(list(string), [])
    })), [])

    # DB (instance-level) engine parameters, managed as a dedicated
    # instance parameter group owned by this cluster and applied to every
    # instance that does not bring its own group (the group is glue -- a
    # named parameter list -- so it stays folded, mirroring `parameters`).
    # Mutually exclusive with neptune_instance_parameter_group_name.
    instance_parameters = optional(list(object({
      # The parameter name (e.g. "neptune_enable_audit_log",
      # "neptune_query_timeout"). Required.
      name = string

      # The parameter value. Required.
      value = string

      # When the change lands: "immediate" (dynamic parameters apply now)
      # or "pending-reboot" (static parameters wait for the next instance
      # reboot). Empty defers to the provider default, which is
      # "pending-reboot" at the pinned provider version -- set "immediate"
      # explicitly when a dynamic parameter should land right away.
      apply_method = optional(string, "")
    })), [])
  })
}
