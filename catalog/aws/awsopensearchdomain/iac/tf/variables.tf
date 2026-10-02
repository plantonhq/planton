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
  description = "AwsOpenSearchDomain specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # OpenSearch or Elasticsearch engine version. Format: "OpenSearch_X.Y" (e.g.,
    # "OpenSearch_2.11") or "Elasticsearch_X.Y" (e.g., "Elasticsearch_7.10").
    # Upgrades are applied in place; a version DOWNGRADE (or a change to an
    # incompatible version) forces domain recreation.
    engine_version = string

    # Cluster topology: data nodes, dedicated masters, coordinator node pools,
    # zone awareness, warm/cold storage.
    cluster_config = object({
      # Instance type for data nodes. Uses the `.search` suffix.
      # Examples: "t3.small.search" (dev), "r6g.large.search" (production),
      # "r6g.2xlarge.search" (high-memory workloads).
      instance_type = string

      # Number of data node instances. Default: 1.
      # For zone-aware deployments, use a multiple of the availability zone count.
      instance_count = optional(number)

      # Enable dedicated master nodes for cluster stability. Dedicated masters handle
      # cluster management tasks (shard allocation, index state management) without
      # competing with data node workloads. Recommended for production.
      dedicated_master_enabled = optional(bool, false)

      # Instance type for dedicated master nodes. Does not need EBS storage.
      # Example: "r6g.large.search". Only used when `dedicated_master_enabled` is true.
      dedicated_master_type = optional(string, "")

      # Number of dedicated master nodes. AWS recommends 3 for production (provides
      # quorum for split-brain protection). Only used when `dedicated_master_enabled` is true.
      dedicated_master_count = optional(number, 0)

      # Additional node pools beyond data and master nodes. Today AWS supports one
      # pool type: "coordinator" nodes, which take over request routing, query
      # fan-out, and response aggregation so data nodes spend their capacity on
      # indexing and searching. Valuable for high-concurrency dashboards and
      # aggregation-heavy workloads.
      node_options = optional(list(object({
        # The pool type. AWS currently supports "coordinator".
        node_type = string

        # Whether the pool is active. Set false to keep the pool definition while
        # removing its nodes.
        enabled = optional(bool, false)

        # Instance type for nodes in this pool (`.search` suffix).
        # Example: "m7g.large.search". Required when the pool is enabled.
        instance_type = optional(string, "")

        # Number of nodes in the pool (at least 1). Required when the pool is
        # enabled.
        count = optional(number, 0)
      })), [])

      # Enable zone awareness to distribute data nodes and replicas across multiple
      # Availability Zones for resilience against AZ-level failures.
      zone_awareness_enabled = optional(bool, false)

      # Number of Availability Zones. Must be 2 or 3.
      # Only used when `zone_awareness_enabled` is true.
      availability_zone_count = optional(number, 0)

      # Enable UltraWarm storage tier for infrequently accessed, read-only data.
      # UltraWarm uses S3-backed storage at lower cost per GB than hot storage.
      # Requires warm_type and warm_count.
      warm_enabled = optional(bool, false)

      # Instance type for UltraWarm nodes: "ultrawarm1.medium.search",
      # "ultrawarm1.large.search", or "ultrawarm1.xlarge.search". Required when
      # `warm_enabled` is true.
      warm_type = optional(string, "")

      # Number of UltraWarm nodes. Range: 2-150. Required when `warm_enabled`
      # is true.
      warm_count = optional(number, 0)

      # Enable cold storage backed by S3. Requires UltraWarm to be enabled.
      # Cold storage provides the lowest-cost tier for data that is rarely queried.
      cold_storage_enabled = optional(bool, false)

      # Enable Multi-AZ with Standby for 99.99% availability SLA. Deploys standby
      # nodes in a different AZ that take over automatically during AZ failures.
      # Requires 3 AZs and at least 3 data nodes.
      multi_az_with_standby_enabled = optional(bool, false)
    })

    # EBS volume configuration for data node storage. Required for most instance types
    # (all except certain storage-optimized types that use instance storage).
    ebs_options = object({
      # Whether EBS volumes are attached to data nodes. Required for most instance
      # types. Only storage-optimized instances (e.g., i3) use instance storage.
      ebs_enabled = optional(bool, false)

      # EBS volume type. "gp3" (recommended), "gp2", "io1", or "standard".
      # gp3 provides predictable performance with configurable IOPS and throughput.
      volume_type = optional(string, "")

      # Size of each EBS volume in GB. The total storage is volume_size * instance_count.
      volume_size = optional(number, 0)

      # Provisioned IOPS for the volume. Only valid for "gp3" and "io1" volume types.
      # gp3 baseline: 3000 IOPS. io1: specify based on workload.
      iops = optional(number, 0)

      # Provisioned throughput in MiB/s. Only valid for "gp3" volume type.
      # Minimum: 125 MiB/s. gp3 baseline: 125 MiB/s.
      throughput = optional(number, 0)
    })

    # Enable encryption at rest for indices and automated snapshots. Uses the
    # AWS-managed `aws/es` key unless `kms_key_id` is provided. Defaults to
    # TRUE (secure by default); set an explicit false only for legacy engine
    # versions that cannot encrypt.
    # One-way: encryption cannot be disabled once enabled (disabling forces
    # domain recreation), and enabling it on very old Elasticsearch versions
    # (< 6.7) also forces recreation.
    encrypt_at_rest_enabled = optional(bool)

    # Customer-managed KMS key ARN or ID for at-rest encryption. ForceNew — the
    # KMS key cannot be changed after domain creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # Enable TLS encryption for all traffic between nodes in the cluster.
    # Defaults to TRUE (secure by default). One-way: cannot be disabled once
    # enabled (disabling forces domain recreation).
    node_to_node_encryption_enabled = optional(bool)

    # VPC placement configuration. When provided, the domain is deployed into VPC
    # subnets and is not publicly accessible. ForceNew — adding or removing VPC
    # options destroys and recreates the domain.
    vpc_options = optional(object({
      # Subnet IDs where OpenSearch deploys ENIs. For zone-aware domains, provide
      # subnets in 2 or 3 AZs matching the cluster's availability_zone_count.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnet_ids = optional(list(string), [])

      # Security group IDs controlling inbound/outbound traffic to the domain.
      # Must allow HTTPS (port 443) from clients that need to access OpenSearch.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      security_group_ids = optional(list(string), [])
    }))

    # HTTPS enforcement, TLS policy, and custom endpoint configuration.
    domain_endpoint_options = optional(object({
      # Require HTTPS for all traffic to the domain endpoint. Default: true.
      # Strongly recommended for all environments.
      enforce_https = optional(bool)

      # TLS security policy for the HTTPS endpoint. Controls the minimum TLS version
      # and cipher suites. "Policy-Min-TLS-1-2-PFS-2023-10" is the recommended
      # baseline; the FIPS policy serves regulated workloads. Leave empty for the
      # provider default.
      tls_security_policy = optional(string, "")

      # Enable a custom domain endpoint (e.g., "search.example.com") instead of the
      # AWS-generated endpoint.
      custom_endpoint_enabled = optional(bool, false)

      # The fully qualified domain name for the custom endpoint.
      # Only used when `custom_endpoint_enabled` is true.
      custom_endpoint = optional(string, "")

      # ACM certificate ARN for the custom endpoint. Must be a valid certificate
      # covering the custom_endpoint FQDN.
      # Only used when `custom_endpoint_enabled` is true.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      custom_endpoint_certificate_arn = optional(string, "")
    }))

    # Advanced security options enable fine-grained access control: internal user
    # database, IAM-based authentication, JWT bearer authentication, and role-based
    # index-level permissions. Once enabled, FGAC cannot be disabled (ForceNew).
    advanced_security_options = optional(object({
      # Enable fine-grained access control. ForceNew if disabling (cannot disable
      # once enabled without domain recreation).
      enabled = optional(bool, false)

      # Enable the internal user database. When true, you can create users and roles
      # directly in OpenSearch Dashboards. When false, use IAM or SAML for authentication.
      internal_user_database_enabled = optional(bool, false)

      # Allow anonymous (unauthenticated) requests while FGAC is enabled -- roles
      # mapped to the anonymous backend decide what such requests may do. Can only
      # be enabled at the moment FGAC itself is first enabled on the domain.
      anonymous_auth_enabled = optional(bool, false)

      # IAM entity ARN (user or role) designated as the master user. The master user
      # has full access to the cluster, indices, and OpenSearch Dashboards.
      # Mutually exclusive with master_user_name/master_user_password.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      master_user_arn = optional(string, "")

      # Username for the internal user database master user.
      # Mutually exclusive with master_user_arn.
      master_user_name = optional(string, "")

      # Password for the internal user database master user. Must be at least 8
      # characters with uppercase, lowercase, digit, and special character.
      # Mutually exclusive with master_user_arn.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      master_user_password = optional(string, "")

      # JWT bearer-token authentication: clients present tokens signed by an
      # external identity provider (validated against jwks_url or public_key).
      # Requires OpenSearch 2.11+.
      jwt_options = optional(object({
        # Whether JWT authentication is active.
        enabled = optional(bool, false)

        # URL of the identity provider's JSON Web Key Set used to validate token
        # signatures (e.g., "https://idp.example.com/.well-known/jwks.json").
        # Provide jwks_url or public_key (or both).
        jwks_url = optional(string, "")

        # PEM-encoded public key used to validate token signatures. An alternative to
        # jwks_url for providers without a JWKS endpoint. The key material is public
        # by definition -- it verifies signatures, it cannot create them.
        public_key = optional(string, "")

        # The token claim that carries the user's roles/groups. Example: "roles".
        # Up to 64 characters.
        roles_key = optional(string, "")

        # The token claim that identifies the user. Example: "sub", "email".
        # Up to 64 characters.
        subject_key = optional(string, "")
      }))
    }))

    # Amazon Cognito authentication for OpenSearch Dashboards: users sign in
    # through a Cognito user pool and are authorized through a Cognito identity
    # pool. An alternative to FGAC's internal user database for the Dashboards
    # UI (the two can also be combined -- Cognito authenticates, FGAC authorizes).
    cognito_options = optional(object({
      # Whether Cognito authentication for Dashboards is active.
      enabled = optional(bool, false)

      # The Cognito user pool users sign in through.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      user_pool_id = optional(string, "")

      # The Cognito identity pool that exchanges user-pool sign-ins for AWS
      # credentials. Identity pools have no Planton kind; provide the raw ID
      # (format: "<region>:<uuid>").
      identity_pool_id = optional(string, "")

      # IAM role that allows OpenSearch Service to configure the user and identity
      # pools (typically carrying the AmazonOpenSearchServiceCognitoAccess policy).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = optional(string, "")
    }))

    # Publish domain logs to CloudWatch Logs for monitoring and troubleshooting.
    # Up to 4 configurations — one per log type (INDEX_SLOW_LOGS, SEARCH_SLOW_LOGS,
    # ES_APPLICATION_LOGS, AUDIT_LOGS).
    log_publishing_options = optional(list(object({
      # Type of log to publish. Values:
      # - "INDEX_SLOW_LOGS" — indexing operations exceeding the slow log threshold
      # - "SEARCH_SLOW_LOGS" — search queries exceeding the slow log threshold
      # - "ES_APPLICATION_LOGS" — OpenSearch application and error logs
      # - "AUDIT_LOGS" — fine-grained access control audit trail (requires FGAC enabled)
      log_type = string

      # CloudWatch Logs log group ARN where logs are published.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      cloudwatch_log_group_arn = string

      # Whether this log publishing option is active. Default: true.
      # Set to false to temporarily disable publishing without removing the configuration.
      enabled = optional(bool)
    })), [])

    # IAM-based access policy for the domain. Serialized to JSON by the IaC modules.
    # Controls who can perform actions on the domain and its indices.
    # For VPC domains, this works in conjunction with security groups.
    # For public domains, this is the primary access control mechanism (unless FGAC is enabled).
    access_policies = optional(any)

    # AWS Auto-Tune automatically optimizes JVM heap size, disk I/O, and other
    # performance settings based on cluster metrics. Configure maintenance windows
    # for changes that require a blue/green deployment. Not supported on t2/t3
    # (burstable) instance types.
    auto_tune_options = optional(object({
      # Auto-Tune state. "ENABLED" or "DISABLED".
      desired_state = string

      # Recurring windows during which Auto-Tune may apply blue/green optimizations.
      # Ignored when use_off_peak_window is true (the off-peak window is used
      # instead).
      maintenance_schedules = optional(list(object({
        # When the first window opens, as an RFC3339 timestamp.
        # Example: "2026-08-01T03:00:00Z".
        start_at = string

        # Length of each window, in hours (AWS's only supported duration unit).
        duration_hours = number

        # Recurrence as a cron expression. Example: "cron(0 3 ? * SUN *)" for
        # Sundays at 03:00 UTC.
        cron_expression_for_recurrence = string
      })), [])

      # What happens to Auto-Tune's applied changes when Auto-Tune is later
      # disabled: "NO_ROLLBACK" (keep the tuned settings) or "DEFAULT_ROLLBACK"
      # (revert to defaults -- requires a maintenance schedule to perform the
      # rollback in).
      rollback_on_disable = optional(string, "")

      # Schedule blue/green optimizations inside the domain's off-peak window
      # instead of explicit maintenance schedules.
      use_off_peak_window = optional(bool, false)
    }))

    # Hour of day (0-23, UTC) when the service takes an automated daily snapshot
    # of the domain's indices. Only relevant for domains running Elasticsearch
    # versions below 5.3 — newer versions snapshot hourly regardless — but the
    # setting remains configurable on all domains.
    automated_snapshot_start_hour = optional(number)

    # Daily 10-hour low-traffic window during which AWS schedules service software
    # updates and Auto-Tune blue/green optimizations. AWS defaults the window to
    # 10:00 PM local time when not configured.
    off_peak_window_options = optional(object({
      # Whether the off-peak window feature is active. Defaults to TRUE -- AWS
      # enables the window on every new domain and requires it enabled at create
      # time; false is only meaningful as a later update (AWS may still reject
      # disabling it while features depend on the window).
      enabled = optional(bool)

      # Hour (0-23, local time) when the 10-hour window opens. AWS defaults to
      # 22 (10:00 PM) when not configured.
      window_start_hour = optional(number)

      # Minute (0-59) past the hour when the window opens.
      window_start_minute = optional(number)
    }))

    # Enable automatic service software updates. When true, AWS applies mandatory
    # and optional service software updates during the off-peak window. Both
    # engines always send this setting explicitly, so false actively turns
    # auto-updates OFF (including on domains where they were previously on).
    auto_software_update_enabled = optional(bool, false)

    # Blue/green deployment strategy for configuration changes that require one.
    # "Default": AWS picks the standard strategy. "CapacityOptimized": AWS
    # provisions additional capacity more conservatively to reduce the performance
    # impact of the migration on capacity-sensitive clusters.
    deployment_strategy = optional(string, "")

    # IP address type for the domain. "ipv4" (default) or "dualstack" (IPv4 + IPv6).
    # One-way: changing from "dualstack" back to "ipv4" forces domain recreation.
    ip_address_type = optional(string, "")

    # Low-level key-value configuration options. Common options:
    # - "rest.action.multi.allow_explicit_index": "true" (default)
    # - "indices.fielddata.cache.size": percentage of heap
    # - "indices.query.bool.max_clause_count": max boolean clauses
    # Values must be strings.
    advanced_options = optional(map(string), {})

    # AI/ML capabilities on the domain: natural-language query generation in
    # Dashboards, the S3 vectors engine, and GPU-accelerated vector search.
    aiml_options = optional(object({
      # Natural-language query generation in OpenSearch Dashboards ("ENABLED" or
      # "DISABLED"): users describe a query in plain language and Dashboards
      # generates the DSL. Requires OpenSearch 2.13+.
      natural_language_query_generation_desired_state = optional(string, "")

      # Enable the S3 vectors engine: vector indexes stored on S3-backed storage
      # for large, cost-efficient vector search corpora.
      s3_vectors_engine_enabled = optional(bool, false)

      # Enable GPU-accelerated (serverless) vector search acceleration on the
      # domain.
      serverless_vector_acceleration_enabled = optional(bool, false)
    }))

    # AWS IAM Identity Center (successor to AWS SSO) integration for OpenSearch
    # Dashboards and API access — workforce users sign in with their Identity
    # Center identity instead of IAM credentials.
    identity_center_options = optional(object({
      # Whether Identity Center API access is active.
      enabled_api_access = optional(bool, false)

      # ARN of the IAM Identity Center instance to integrate with.
      identity_center_instance_arn = optional(string, "")

      # The Identity Center attribute that carries a user's groups for role
      # mapping: "GroupName" or "GroupId".
      roles_key = optional(string, "")

      # The Identity Center attribute that identifies a user: "UserName", "UserId",
      # or "Email".
      subject_key = optional(string, "")
    }))

    # SAML authentication for OpenSearch Dashboards: users sign in through an
    # external SAML 2.0 identity provider (Okta, Entra ID, ...). Managed as its
    # own provider resource alongside the domain; removing the block disables
    # SAML. Requires fine-grained access control (advanced_security_options).
    saml_options = optional(object({
      # The identity provider's entity ID (from the IdP metadata).
      # Example: "https://idp.example.com/saml/metadata".
      idp_entity_id = string

      # The IdP's SAML metadata document, as an XML string (typically downloaded
      # from the IdP's metadata endpoint).
      idp_metadata_content = string

      # The SAML backend role granted master (admin) access in Dashboards --
      # mutually exclusive with master_user_name. Empty grants no SAML
      # principal master access (map roles inside Dashboards instead).
      master_backend_role = optional(string, "")

      # The SAML username granted master (admin) access in Dashboards --
      # mutually exclusive with master_backend_role.
      master_user_name = optional(string, "")

      # The SAML assertion attribute that carries the user's roles/groups.
      # Empty uses the IdP's roles attribute conventions.
      roles_key = optional(string, "")

      # The SAML assertion attribute that identifies the user. Empty uses
      # NameID.
      subject_key = optional(string, "")

      # Dashboards session lifetime in minutes, 1-1440. 0 keeps the AWS
      # default (60).
      session_timeout_minutes = optional(number, 0)
    }))

    # AWS account IDs authorized to create OpenSearch-managed VPC endpoints
    # against this domain (the grantor side of cross-account private access).
    # Each entry is managed as its own provider resource keyed by account ID.
    authorized_vpc_endpoint_access_accounts = optional(list(string), [])
  })
}
