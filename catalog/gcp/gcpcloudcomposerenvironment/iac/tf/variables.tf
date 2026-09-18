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
  description = "GcpCloudComposerEnvironment specification"
  type = object({
    # GCP project in which to create the Cloud Composer environment.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # GCP region for the Composer environment (e.g., "us-central1",
    # "europe-west12"). Immutable after creation.
    region = string

    # Name of the Composer environment resource in GCP.
    # Must be lowercase letters, numbers, and hyphens; start with a letter;
    # end with a letter or number; 1-64 characters.
    # Optional: when empty, defaults to metadata.name. Immutable after creation.
    environment_name = optional(string, "")

    # Networking and node configuration for the Composer environment.
    node_config = optional(object({
      # VPC network for the Composer environment.
      # Used for Composer 2.x with VPC peering networking.
      # Not applicable when composer_network_attachment is set (Composer 3 PSC).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      network = optional(string, "")

      # VPC subnetwork for the Composer environment.
      # Used for Composer 2.x with VPC peering networking.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnetwork = optional(string, "")

      # Service account the environment's workloads run as. Must hold
      # roles/composer.worker. Composer 3 REQUIRES this to be explicitly
      # specified (the API rejects environment creation without it); only
      # legacy Composer 2 environments fall back to the default Compute
      # Engine service account.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service_account = optional(string, "")

      # Network tags applied to Composer environment GKE nodes.
      # Used for firewall rule targeting.
      tags = optional(list(string), [])

      # PSC Network Attachment for Composer 3 networking.
      # Format: projects/{project}/regions/{region}/networkAttachments/{name}
      # Mutually exclusive with network/subnetwork (VPC peering).
      composer_network_attachment = optional(string, "")

      # IPv4 CIDR block for Composer 3 internal components.
      # Must be a /20 range. Only applicable when using Composer 3.
      composer_internal_ipv4_cidr_block = optional(string, "")

      # Deploy the ip-masq-agent DaemonSet on the environment's GKE nodes,
      # SNATing pod traffic to node IPs — needed when the pod CIDR is not
      # routable in the wider network.
      enable_ip_masq_agent = optional(bool, false)

      # VPC-native (alias IP) range assignment for the environment's GKE
      # pods and services.
      ip_allocation_policy = optional(object({
        # Name of the subnetwork's existing secondary range to use for pods.
        # Mutually exclusive with cluster_ipv4_cidr_block.
        cluster_secondary_range_name = optional(string, "")

        # CIDR block for the pod range when no named secondary range is used
        # (e.g. "10.4.0.0/14", or a netmask size like "/14").
        # Mutually exclusive with cluster_secondary_range_name.
        cluster_ipv4_cidr_block = optional(string, "")

        # Name of the subnetwork's existing secondary range to use for services.
        # Mutually exclusive with services_ipv4_cidr_block.
        services_secondary_range_name = optional(string, "")

        # CIDR block for the services range when no named secondary range is
        # used. Mutually exclusive with services_secondary_range_name.
        services_ipv4_cidr_block = optional(string, "")
      }))
    }))

    # Airflow software configuration including image version, packages,
    # and configuration overrides.
    software_config = optional(object({
      # Composer and Airflow image version.
      # Format: "composer-A.B.C-airflow-X.Y.Z" (e.g., "composer-2.9.7-airflow-2.9.3").
      # If empty, the latest stable version is used.
      image_version = optional(string, "")

      # Airflow configuration property overrides.
      # Keys are section-key pairs (e.g., "core-dags_are_paused_at_creation": "True").
      # Values that conflict with managed Composer settings are rejected.
      airflow_config_overrides = optional(map(string), {})

      # Custom PyPI packages to install in the environment.
      # Keys are package names, values are version specifiers or empty strings.
      # Example: {"numpy": ">=1.21", "requests": ""}
      pypi_packages = optional(map(string), {})

      # Additional environment variables available to all Airflow components.
      # Variable names starting with "AIRFLOW__" are reserved by Airflow and
      # should not be set here.
      env_variables = optional(map(string), {})

      # Web server plugins mode for Composer 3 environments.
      # When DISABLED, custom Airflow UI plugins are not loaded.
      # Only applicable to Composer 3.
      web_server_plugins_mode = optional(string, "")

      # Cloud Data Lineage integration: Airflow operators report dataset
      # lineage into Dataplex Data Lineage automatically.
      # Applies to Composer 2.1.2+.
      cloud_data_lineage_integration = optional(object({
        # Whether the integration is enabled.
        enabled = optional(bool, false)
      }))
    }))

    # Private networking configuration for Composer 2.x environments using
    # VPC peering or Private Service Connect. Not applicable to Composer 3
    # which uses enable_private_environment and composer_network_attachment instead.
    private_environment_config = optional(object({
      # Whether to deny access to the public Airflow web server endpoint.
      # When true, the web server is only accessible via private IP.
      enable_private_endpoint = optional(bool, false)

      # Connection type for the Composer 2.x private environment.
      connection_type = optional(string, "")

      # IP range for the GKE master network in CIDR notation.
      # Default: 172.16.0.0/28.
      master_ipv4_cidr_block = optional(string, "")

      # IP range for the Cloud SQL instance in CIDR notation.
      cloud_sql_ipv4_cidr_block = optional(string, "")

      # IP range for Cloud Composer internal components in CIDR notation.
      # Applies to Composer 2.x and newer.
      cloud_composer_network_ipv4_cidr_block = optional(string, "")

      # PSC connection subnetwork for Composer 2.x with Private Service Connect.
      cloud_composer_connection_subnetwork = optional(string, "")

      # Whether to allow public IPs from non-RFC1918 ranges for IP allocation
      # in the environment.
      enable_privately_used_public_ips = optional(bool, false)
    }))

    # Workload resource allocation for Airflow components (scheduler, worker,
    # web server, triggerer, DAG processor). Applies to Composer 2.x and 3.
    workloads_config = optional(object({
      # Resource allocation for the Airflow scheduler.
      # The scheduler parses DAGs, manages task scheduling, and triggers task instances.
      scheduler = optional(object({
        # CPU allocation in vCPUs (e.g., 0.5, 1.0, 2.0).
        cpu = optional(number, 0)

        # Memory allocation in GB (e.g., 1.0, 2.0, 4.0).
        memory_gb = optional(number, 0)

        # Storage allocation in GB (e.g., 1.0, 5.0, 10.0).
        storage_gb = optional(number, 0)

        # Number of replicas for this component.
        count = optional(number, 0)
      }))

      # Resource allocation for the Airflow web server (UI).
      web_server = optional(object({
        # CPU allocation in vCPUs.
        cpu = optional(number, 0)

        # Memory allocation in GB.
        memory_gb = optional(number, 0)

        # Storage allocation in GB.
        storage_gb = optional(number, 0)
      }))

      # Resource allocation for Airflow workers.
      # Workers execute the actual tasks defined in DAGs.
      worker = optional(object({
        # CPU allocation per worker in vCPUs.
        cpu = optional(number, 0)

        # Memory allocation per worker in GB.
        memory_gb = optional(number, 0)

        # Storage allocation per worker in GB.
        storage_gb = optional(number, 0)

        # Minimum number of workers. Must be >= 0.
        min_count = optional(number, 0)

        # Maximum number of workers. Must be >= min_count.
        max_count = optional(number, 0)
      }))

      # Resource allocation for the Airflow triggerer.
      # The triggerer monitors deferred tasks and resumes them when conditions are met.
      # Critical for deferrable operators in Airflow 2.x+.
      triggerer = optional(object({
        # CPU allocation per triggerer in vCPUs.
        cpu = optional(number, 0)

        # Memory allocation per triggerer in GB.
        memory_gb = optional(number, 0)

        # Number of triggerer replicas. Set to 0 to disable the triggerer.
        count = optional(number, 0)
      }))

      # Resource allocation for the DAG processor.
      # The DAG processor parses DAG files independently of the scheduler.
      # Only applicable to Composer 3 (replica count is capped at 3).
      dag_processor = optional(object({
        # CPU allocation in vCPUs (e.g., 0.5, 1.0, 2.0).
        cpu = optional(number, 0)

        # Memory allocation in GB (e.g., 1.0, 2.0, 4.0).
        memory_gb = optional(number, 0)

        # Storage allocation in GB (e.g., 1.0, 5.0, 10.0).
        storage_gb = optional(number, 0)

        # Number of replicas for this component.
        count = optional(number, 0)
      }))
    }))

    # Size of the Composer environment. Controls the managed infrastructure
    # capacity (GKE cluster and database sizing behind the scenes).
    environment_size = optional(string, "")

    # Resilience mode for the Composer environment. HIGH_RESILIENCE provides
    # multi-zone redundancy for increased availability. Applies to Composer 2.1.15+.
    resilience_mode = optional(string, "")

    # Customer-managed encryption key for the Composer environment.
    # All Composer-managed resources (GKE nodes, Cloud SQL, Cloud Storage) are
    # encrypted with this key. Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # Maintenance window configuration for the Composer environment.
    # Defines when GCP may perform maintenance operations on the environment.
    maintenance_window = optional(object({
      # Start time of the maintenance window in RFC3339 format.
      # Example: "2026-01-01T00:00:00Z"
      start_time = string

      # End time of the maintenance window in RFC3339 format.
      # Must be after start_time. The window duration must be at least 12 hours.
      end_time = string

      # Recurrence specification in RFC5545 RRULE format.
      # Examples: "FREQ=WEEKLY;BYDAY=TU,WE,TH" or "FREQ=DAILY".
      recurrence = string
    }))

    # Recovery configuration with scheduled snapshots for disaster recovery.
    recovery_config = optional(object({
      # Whether scheduled snapshots are enabled.
      enabled = optional(bool, false)

      # Cloud Storage location for snapshots (GCS bucket folder URI).
      # Example: "gs://my-bucket/composer-snapshots"
      snapshot_location = optional(string, "")

      # Cron schedule for snapshot creation in Unix-cron format.
      # Example: "0 4 * * *" (daily at 4 AM).
      snapshot_creation_schedule = optional(string, "")

      # Time zone for the cron schedule (e.g., "America/Los_Angeles", "UTC").
      time_zone = optional(string, "")
    }))

    # Network-level access restrictions for the Airflow web server UI.
    # When configured, only requests from the specified IP ranges are allowed.
    web_server_network_access_control = optional(object({
      # Allowed IP ranges that can access the Airflow web server.
      allowed_ip_ranges = optional(list(object({
        # IP address or CIDR range (e.g., "10.0.0.0/8", "203.0.113.0/24").
        value = string

        # Optional human-readable description of this IP range.
        description = optional(string, "")
      })), [])
    }))

    # IP-based access control for the environment's GKE cluster master.
    # Restricts which networks can reach the Kubernetes control plane that
    # runs the Airflow workloads.
    master_authorized_networks_config = optional(object({
      # Whether master authorized networks are enforced.
      enabled = optional(bool, false)

      # Networks allowed to reach the Kubernetes control plane.
      cidr_blocks = optional(list(object({
        # CIDR block allowed to access the cluster master (e.g., "10.0.0.0/8").
        cidr_block = string

        # Optional display name for the network.
        display_name = optional(string, "")
      })), [])
    }))

    # Retention policies for Airflow task logs and metadata database rows —
    # the levers that keep long-lived environments from accumulating
    # unbounded operational data.
    data_retention_config = optional(object({
      # Where Airflow task logs are stored.
      # CLOUD_LOGGING_ONLY: task logs go to Cloud Logging only.
      # CLOUD_LOGGING_AND_CLOUD_STORAGE: logs also land in the environment's
      # bucket. Applies to Composer 2.0.32+ (not Composer 3).
      task_logs_storage_mode = optional(string, "")

      # Whether Airflow metadata database retention is enforced.
      # Composer 3 only.
      airflow_metadata_retention_mode = optional(string, "")

      # Days of Airflow metadata (task history, XComs, logs metadata) to
      # retain when retention is enabled (30-730). Composer 3 only.
      airflow_metadata_retention_days = optional(number, 0)
    }))

    # Existing Cloud Storage bucket for the environment's DAGs, plugins, and
    # data (instead of the bucket Composer auto-creates). Resolves to the
    # bucket name. Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    storage_bucket = optional(string, "")

    # Enable private environment for Composer 3 environments. When true, the
    # environment does not have a public IP endpoint for the web server.
    # Only applicable to Composer 3.
    enable_private_environment = optional(bool, false)

    # Enable private builds only for Composer 3 environments. When true,
    # only builds using private connectivity are allowed for Python packages.
    # Only applicable to Composer 3.
    enable_private_builds_only = optional(bool, false)

    # User-defined labels to organize and track the environment. Merged
    # beneath Planton's platform attribution labels (platform keys win on
    # conflict).
    labels = optional(map(string), {})

    # Deletion policy for the environment — what happens when this
    # resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the environment is deleted (10-15 minutes: Composer
    #                tears down the managed GKE cluster and database).
    #                The DAG bucket Composer auto-created survives — GCP
    #                never deletes it with the environment
    #   "PREVENT" -- destroy FAILS; protects the environment whose DAGs
    #                a data platform runs on
    #   "ABANDON" -- the environment is removed from management but keeps
    #                running (and billing meaningfully — Composer bills
    #                for its infrastructure even when idle) in GCP
    deletion_policy = optional(string, "")
  })
}
