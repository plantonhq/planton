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
  description = "AwsRedshiftCluster specification"
  type = object({
    # The AWS region the cluster is created in. Must match the region of
    # the subnets, security groups, and KMS keys it references.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # Subnets for the cluster's Redshift subnet group. Provide at least
    # two subnets in DISTINCT availability zones so the cluster (and a
    # Multi-AZ standby, if enabled) has somewhere to land. Reference
    # AwsSubnet subnet_id outputs or pass literal subnet IDs. The module
    # manages the subnet group itself (pure glue: a named list of
    # subnets); alternatively point cluster_subnet_group_name at an
    # existing group. Changing the subnet group replaces the cluster.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = optional(list(string), [])

    # Name of an existing Redshift subnet group to place the cluster in,
    # instead of providing subnet_ids. Changing the subnet group replaces
    # the cluster.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    cluster_subnet_group_name = optional(string, "")

    # Security groups attached to the cluster (the cluster's
    # vpc_security_group_ids). Empty uses the VPC's default security
    # group (the AWS default). Reference AwsSecurityGroup
    # security_group_id outputs or pass literal SG IDs -- warehouse
    # ingress rules (e.g. port 5439 from your BI tooling) belong on the
    # referenced AwsSecurityGroup node, never inside this cluster.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # Pin the cluster to one availability zone. Empty lets AWS place it
    # -- preferred. Changing the zone on a live cluster requires
    # availability_zone_relocation_enabled; without relocation the pin is
    # create-time only.
    availability_zone = optional(string, "")

    # Allow the cluster to be relocated to another availability zone
    # during outages or on demand -- zero data loss, brief
    # unavailability. Requires RA3 node types and a cluster port in the
    # ranges 5431-5455 or 8191-8215 (an AWS relocation constraint).
    # Mutually exclusive with multi_az: relocation recovers by moving the
    # single cluster, Multi-AZ recovers by failing over to a standby.
    availability_zone_relocation_enabled = optional(bool, false)

    # Give the cluster a public IP so it can be reached from outside the
    # VPC. Off by default -- warehouses almost always stay private behind
    # VPC routing (and Query Editor / private BI reach them fine).
    publicly_accessible = optional(bool, false)

    # A static public IPv4 address for the cluster's leader node.
    # Requires publicly_accessible. Reference an AwsElasticIp public_ip
    # output or pass a literal Elastic IP ADDRESS (Redshift takes the IP
    # itself, not an allocation ID).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    elastic_ip = optional(string, "")

    # Force all COPY and UNLOAD traffic between the cluster and data
    # repositories (S3, DynamoDB, ...) through the VPC instead of the
    # public internet -- enabling VPC flow logs, endpoints, and other
    # network controls to see and govern warehouse data movement.
    enhanced_vpc_routing = optional(bool, false)

    # The port the cluster accepts connections on. 0 keeps the AWS
    # default (5439). Redshift accepts 1115-65535; if
    # availability_zone_relocation_enabled is on, AWS additionally
    # requires the port to be within 5431-5455 or 8191-8215.
    port = optional(number, 0)

    # The compute/storage class of every node. RA3 classes (ra3.large,
    # ra3.xlplus, ra3.4xlarge, ra3.16xlarge) decouple compute from
    # managed storage that tiers between SSD and S3 automatically -- the
    # right call for nearly all new clusters and the only family that
    # supports multi_az and availability-zone relocation. DC2 classes
    # (dc2.large, dc2.8xlarge) are the legacy dense-compute family with
    # node-local SSD only. Resizing to a different class is an in-place
    # (but access-interrupting) classic/elastic resize, never a replace.
    node_type = string

    # How many nodes the cluster runs. 0 keeps the AWS default (1, a
    # single-node cluster where leader and compute share one node).
    # 2+ creates a multi-node cluster with a dedicated leader --
    # required for production and for multi_az. Resize is in-place.
    number_of_nodes = optional(number, 0)

    # The Redshift engine version. Empty keeps the AWS default ("1.0" --
    # the only version family Redshift has ever shipped); actual engine
    # patches ride maintenance_track_name and allow_version_upgrade.
    cluster_version = optional(string, "")

    # The name of the first database created in the cluster. Empty keeps
    # the AWS default ("dev"). 1-64 characters: lowercase alphanumeric,
    # underscore, or dollar sign, starting with a letter or underscore
    # (AWS's CreateCluster contract).
    database_name = optional(string, "")

    # The admin username. Required for a brand-new cluster -- AWS has no
    # default and rejects a blank value at CreateCluster. 1-128
    # characters starting with a letter; letters, digits, and _.@+- are
    # legal. Only clusters restored from a snapshot leave it empty and
    # inherit the source's credentials. Create-time only -- changing it
    # replaces the cluster.
    master_username = optional(string, "")

    # Let AWS manage the admin password in Secrets Manager: AWS
    # generates it, stores it, rotates it on schedule, and no secret ever
    # touches this manifest or the IaC state. The managed secret's ARN is
    # exported as the master_password_secret_arn output. Mutually
    # exclusive with master_password -- and the recommended posture.
    manage_master_password = optional(bool, false)

    # The admin password, supplied directly (8-64 chars with at least one
    # uppercase letter, one lowercase letter, and one digit). Stored in
    # IaC state -- prefer manage_master_password, which keeps the secret
    # in Secrets Manager entirely. Mutually exclusive with
    # manage_master_password.
    master_password = optional(string, "")

    # The KMS key that encrypts the Secrets Manager secret holding the
    # managed admin password. Empty uses the AWS-managed
    # aws/secretsmanager key. Only meaningful with
    # manage_master_password. Reference an AwsKmsKey key_arn output or
    # pass a literal key ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    master_password_secret_kms_key_id = optional(string, "")

    # Encrypt cluster storage at rest. AWS defaults new clusters to
    # encrypted, and this spec keeps that default -- set false only for
    # a deliberate, exceptional reason. Toggling encryption on a live
    # cluster is an in-place but long-running migration.
    encrypted = optional(bool)

    # The KMS key for cluster storage encryption when encrypted is true.
    # Empty uses the AWS-managed Redshift service key. Reference an
    # AwsKmsKey key_arn output or pass a literal key ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # Run a Multi-AZ deployment: compute in two availability zones with
    # automatic failover and a single endpoint. Requires RA3 node types
    # and a multi-node cluster. Mutually exclusive with
    # availability_zone_relocation_enabled.
    multi_az = optional(bool, false)

    # IAM roles the cluster assumes to access other AWS services during
    # COPY, UNLOAD, CREATE EXTERNAL FUNCTION, and Redshift Spectrum
    # queries (S3, DynamoDB, Glue, Lambda, ...). AWS allows up to 10
    # associated roles. Reference AwsIamRole role_arn outputs or pass
    # literal role ARNs.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    iam_roles = optional(list(string), [])

    # The IAM role assumed when a SQL command does not name one
    # explicitly (e.g. COPY ... IAM_ROLE default). Must also be present
    # in iam_roles. Reference an AwsIamRole role_arn output or pass a
    # literal role ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    default_iam_role_arn = optional(string, "")

    # Days automated snapshots are retained, 0-35. 0 disables automated
    # snapshots entirely (not recommended); unset keeps the AWS default
    # (1 day). Production warehouses typically want 7+.
    automated_snapshot_retention_period = optional(number)

    # Days MANUAL snapshots are retained: 1-3653, or -1 to retain
    # indefinitely. 0 keeps the AWS default (-1, indefinite). Applies to
    # new manual snapshots taken after the change.
    manual_snapshot_retention_period = optional(number, 0)

    # The weekly maintenance window in UTC, format
    # "ddd:hh24:mi-ddd:hh24:mi" (e.g. "sat:03:00-sat:04:00"). Empty lets
    # AWS assign one.
    preferred_maintenance_window = optional(string, "")

    # The maintenance track the cluster follows. Empty keeps the AWS
    # default ("current" -- the latest approved release). "trailing"
    # stays one release behind; a snapshot restore can also land on a
    # named preview/source track inherited from its source cluster.
    maintenance_track_name = optional(string, "")

    # Permit engine version upgrades during the maintenance window. AWS
    # defaults this to true; disable only when upgrade timing must be
    # controlled manually.
    allow_version_upgrade = optional(bool)

    # Apply modifications immediately instead of waiting for the next
    # maintenance window. Immediate changes can interrupt connections;
    # deferred changes wait quietly. AWS defaults to deferred.
    apply_immediately = optional(bool, false)

    # Skip the final snapshot when the cluster is deleted. When false
    # (the safe default), final_snapshot_identifier must be set -- AWS
    # refuses to delete without knowing the snapshot name.
    skip_final_snapshot = optional(bool, false)

    # The name for the final snapshot taken on deletion. Required when
    # skip_final_snapshot is false.
    final_snapshot_identifier = optional(string, "")

    # Restore the cluster from an existing snapshot by NAME at create
    # time. Create-time only; mutually exclusive with snapshot_arn. The
    # restored cluster inherits the snapshot's credentials, so
    # master_username stays empty.
    snapshot_identifier = optional(string, "")

    # Restore the cluster from an existing snapshot by ARN at create
    # time -- the shape cross-account/cross-region snapshot shares use.
    # Create-time only; mutually exclusive with snapshot_identifier.
    snapshot_arn = optional(string, "")

    # The name of the cluster the source snapshot was taken from.
    # Required by AWS only when the snapshot name alone is ambiguous
    # (shared snapshots). Only meaningful alongside snapshot_identifier.
    snapshot_cluster_identifier = optional(string, "")

    # The AWS account that owns the source snapshot, for restoring from
    # a snapshot shared by another account. Only meaningful alongside a
    # restore source.
    owner_account = optional(string, "")

    # Audit logging for the cluster -- connection attempts, user
    # activity, and user changes delivered to S3 or CloudWatch Logs. A
    # cluster setting keyed by the cluster itself (folded, never a
    # standalone node).
    logging = optional(object({
      # Where audit logs are delivered: "s3" writes log files to an S3
      # bucket (the bucket needs a policy granting the Redshift service
      # write access); "cloudwatch" streams them to CloudWatch Logs --
      # the modern destination with retention, metric filters, and
      # subscriptions.
      log_destination_type = string

      # The S3 bucket audit logs are written to. Required when
      # log_destination_type is "s3".
      s3_bucket_name = optional(string, "")

      # An optional key prefix for log objects within the S3 bucket.
      s3_key_prefix = optional(string, "")

      # Which audit log types to export: "connectionlog" (connection
      # attempts), "useractivitylog" (every executed query -- requires the
      # enable_user_activity_logging cluster parameter), "userlog" (user
      # create/alter/drop events). Required for "cloudwatch"; ignored for
      # "s3" (S3 delivery always carries all three).
      log_exports = optional(list(string), [])
    }))

    # Cross-region disaster recovery: automatically copy this cluster's
    # snapshots to another region. A cluster setting keyed by the
    # cluster itself (folded, never a standalone node).
    snapshot_copy = optional(object({
      # The region snapshots are copied to. Must differ from the cluster's
      # own region. Required.
      destination_region = string

      # Days copied AUTOMATED snapshots are retained in the destination
      # region, 1-35. 0 keeps the AWS default (7).
      retention_period = optional(number, 0)

      # Days copied MANUAL snapshots are retained in the destination
      # region: 1-3653, or -1 to retain indefinitely. 0 keeps the AWS
      # default (-1, indefinite).
      manual_snapshot_retention_period = optional(number, 0)

      # The snapshot copy grant that lets Redshift encrypt copied
      # snapshots with a KMS key in the destination region. Required by
      # AWS when the cluster is KMS-encrypted; irrelevant otherwise.
      snapshot_copy_grant_name = optional(string, "")
    }))

    # The name of an existing cluster parameter group to use. Mutually
    # exclusive with `parameters` -- either bring your own group or let
    # the module manage one from inline parameters.
    cluster_parameter_group_name = optional(string, "")

    # Cluster-level engine parameters (e.g. "require_ssl",
    # "enable_user_activity_logging", "wlm_json_configuration"), managed
    # as a dedicated parameter group owned by this cluster (the group is
    # glue -- a named parameter list -- so it stays folded). Mutually
    # exclusive with cluster_parameter_group_name.
    parameters = optional(list(object({
      # The parameter name (e.g. "require_ssl",
      # "enable_user_activity_logging", "max_concurrency_scaling_clusters",
      # "wlm_json_configuration"). Required.
      name = string

      # The parameter value. Required.
      value = string
    })), [])

    # The parameter-group family for the managed group created from
    # `parameters`. Empty keeps "redshift-1.0" -- the long-standing
    # family AWS accepts on every cluster. AWS introduced "redshift-2.0"
    # with the Redshift patch 2.0 generation (new clusters' default
    # group is default.redshift-2.0); set it here when the managed group
    # should track that family. Only meaningful alongside `parameters`.
    parameter_group_family = optional(string, "")

    # The identifier of an existing snapshot schedule to associate with
    # this cluster (AWS allows exactly ONE schedule per cluster -- the
    # schedule replaces the default automated-snapshot cadence with
    # explicit cron/rate definitions). The schedule itself is an
    # account-scoped resource shared by many clusters and is not created
    # here. Empty keeps AWS's default snapshot cadence.
    snapshot_schedule_identifier = optional(string, "")

    # Usage limits capping what individual Redshift features may consume
    # on this cluster -- Spectrum data scanned, concurrency-scaling time,
    # cross-region datasharing transfer -- each with a breach action from
    # logging to hard disable. Cluster-scoped settings keyed by the
    # cluster itself (folded, never standalone nodes).
    usage_limits = optional(list(object({
      # The Redshift feature the limit applies to: "spectrum" (external S3
      # scans), "concurrency-scaling" (burst clusters),
      # "cross-region-datasharing" (data transferred to consumers in other
      # regions), or "extra-compute-for-automatic-optimization". Required;
      # changing it replaces the limit.
      feature_type = string

      # How the limit is measured: "time" (minutes of feature usage) or
      # "data-scanned" (terabytes). AWS's contract pairs spectrum and
      # cross-region-datasharing with "data-scanned", concurrency-scaling
      # and extra-compute-for-automatic-optimization with "time"
      # (CEL-enforced). Required; changing it replaces the limit.
      limit_type = string

      # The limit amount: minutes when limit_type is "time", terabytes
      # when it is "data-scanned". Must be positive.
      amount = optional(number, 0)

      # The period the amount applies to: "daily", "weekly", or "monthly".
      # Empty keeps the AWS default (monthly). A weekly period begins on
      # Sunday. Changing it replaces the limit.
      period = optional(string, "")

      # What Redshift does when the limit is breached: "log" writes an
      # event to the system table, "emit-metric" additionally publishes a
      # CloudWatch metric, "disable" turns the feature off until the
      # period resets (spectrum and concurrency-scaling only -- AWS
      # rejects disable for cross-region datasharing). Empty keeps the AWS
      # default (log).
      breach_action = optional(string, "")
    })), [])

    # Scheduled actions that pause, resume, or resize this cluster on a
    # cron/at schedule -- the standard nights-and-weekends cost lever for
    # non-production warehouses. Each entry keys one scheduled action
    # targeting this cluster.
    scheduled_actions = optional(list(object({
      # The scheduled action's name: 1-63 lowercase alphanumeric/hyphen
      # characters. NOTE: names are unique per AWS ACCOUNT (not per
      # cluster) -- include something cluster-specific to avoid collisions
      # across clusters. Changing it replaces the action.
      name = string

      # An optional description of what the action does and why.
      description = optional(string, "")

      # Suspend the action without deleting it. The zero value (false)
      # keeps the AWS default: the action is enabled and fires on
      # schedule.
      disabled = optional(bool, false)

      # When the action fires, in at() or cron() format -- e.g.
      # "cron(0 22 * * ? *)" (daily 22:00 UTC) or
      # "at(2026-09-01T03:00:00)". Cron fields are
      # Minutes Hours Day-of-month Month Day-of-week Year, in UTC.
      schedule = string

      # When the schedule becomes active (RFC 3339, e.g.
      # "2026-09-01T00:00:00Z"). Empty activates it immediately.
      start_time = optional(string, "")

      # When the schedule expires (RFC 3339). Empty keeps it active until
      # deleted.
      end_time = optional(string, "")

      # The IAM role Redshift assumes to perform the action. The role's
      # TRUST POLICY must allow the "scheduler.redshift.amazonaws.com"
      # service principal to sts:AssumeRole (AWS validates the trust at
      # create -- a role trusting only redshift.amazonaws.com or ec2 is
      # rejected; the provider retries briefly on trust-propagation
      # delays). Reference an AwsIamRole role_arn output or pass a literal
      # role ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      iam_role_arn = string

      # Pause the cluster (compute stops billing; storage persists).
      # Exactly one action arm must be set (CEL-enforced).
      pause_cluster = optional(bool, false)

      # Resume a paused cluster. Exactly one action arm must be set
      # (CEL-enforced).
      resume_cluster = optional(bool, false)

      # Resize the cluster to a new node count/type. Exactly one action
      # arm must be set (CEL-enforced).
      resize_cluster = optional(object({
        # Force a CLASSIC resize (full data redistribution -- hours, but
        # works between any topologies) instead of the default elastic
        # resize (minutes, but constrained to compatible node counts).
        classic = optional(bool, false)

        # The target cluster type. Empty keeps the current type; AWS derives
        # it from the node count otherwise.
        cluster_type = optional(string, "")

        # The target node class (e.g. "ra3.large"). Empty keeps the current
        # class.
        node_type = optional(string, "")

        # The target node count. 0 keeps the current count.
        number_of_nodes = optional(number, 0)
      }))
    })), [])

    # Redshift-managed VPC endpoints that expose this cluster inside
    # another subnet group (same-account cross-VPC access without
    # peering; requires RA3 node types). Each entry keys one managed
    # endpoint; its private address and port are exported per endpoint on
    # the outputs contract.
    endpoint_accesses = optional(list(object({
      # The endpoint's name: 1-30 lowercase alphanumeric/hyphen
      # characters, unique within the cluster. Changing it replaces the
      # endpoint.
      endpoint_name = string

      # The Redshift subnet group the endpoint lands in -- typically a
      # group in the CONSUMING VPC. Empty reuses the cluster's own subnet
      # group (module-managed or referenced), which yields an extra
      # endpoint in the cluster's own VPC. Changing it replaces the
      # endpoint.
      subnet_group_name = optional(string, "")

      # Security groups attached to the endpoint's network interfaces.
      # Empty uses the VPC's default security group. Reference
      # AwsSecurityGroup security_group_id outputs or pass literal SG IDs.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      vpc_security_group_ids = optional(list(string), [])
    })), [])

    # Grants that authorize OTHER AWS accounts to create managed VPC
    # endpoints to this cluster (the grantor side of cross-account
    # access; the grantee creates its endpoint in its own account). Each
    # entry keys one per-account authorization.
    endpoint_authorizations = optional(list(object({
      # The AWS account ID being authorized (12 digits). Changing it
      # replaces the authorization.
      account = string

      # Restrict the grant to specific VPCs in the grantee account. Empty
      # authorizes ALL of the account's VPCs. Reference AwsVpc vpc_id
      # outputs or pass literal VPC IDs.
      #
      # Containment-exempt: a grant admits another VPC's endpoints to the
      # cluster; the cluster never lives inside the VPC it authorizes.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      vpc_ids = optional(list(string), [])

      # Revoke the authorization on delete even if the grantee still has
      # live endpoints against the cluster (their endpoints are deleted
      # too). Off by default: delete fails while grantee endpoints exist.
      force_delete = optional(bool, false)
    })), [])
  })
}
