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
  description = "GcpAlloydbInstance specification"
  type = object({
    # The GCP project that owns the AlloyDB cluster.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The AlloyDB cluster this instance belongs to. Accepts the full cluster
    # resource path or a reference to a GcpAlloydbCluster resource. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    cluster = string

    # The instance ID within the cluster. Immutable.
    instance_id = string

    # Instance role: PRIMARY, READ_POOL, or SECONDARY. Immutable.
    # Presets default to READ_POOL for read scaling.
    instance_type = optional(string)

    # Number of CPUs. GCP selects the machine family. Mutually exclusive with
    # machine_type.
    cpu_count = optional(number, 0)

    # Explicit machine type (e.g. "n2-highmem-4"). Mutually exclusive with
    # cpu_count.
    machine_type = optional(string, "")

    # Read pool sizing. Required when instance_type is READ_POOL.
    read_pool_config = optional(object({
      # Read capacity — number of nodes in the read pool instance.
      node_count = optional(number, 0)
    }))

    # ZONAL or REGIONAL placement for PRIMARY/SECONDARY instances. GCP
    # defaults to REGIONAL when unset. Must stay empty on READ_POOL
    # instances: read-pool availability is DERIVED from node_count (1 node =
    # ZONAL, 2+ nodes = REGIONAL spread across zones) and the AlloyDB API
    # does not store a sent value — the stored object omits the field, so
    # any explicit value produces a perpetual re-plan diff (live-verified
    # against a single-node read pool).
    availability_type = optional(string, "")

    # PostgreSQL database flags as key-value pairs.
    database_flags = optional(map(string), {})

    # Human-readable display name.
    display_name = optional(string, "")

    # Query insights configuration.
    query_insights_config = optional(object({
      # Number of query execution plans captured per minute. Range: 0-20.
      query_plans_per_minute = optional(number, 0)

      # Maximum length of the query string stored in insights. Range: 256-4500.
      # 0 means unset — GCP applies its default (1024).
      query_string_length = optional(number, 0)

      # Whether to record application tags for queries.
      record_application_tags = optional(bool, false)

      # Whether to record the client IP address for each query.
      record_client_address = optional(bool, false)
    }))

    # When true, only AlloyDB Auth Proxy / Language Connectors may connect.
    require_connectors = optional(bool, false)

    # SSL mode: ENCRYPTED_ONLY or ALLOW_UNENCRYPTED_AND_ENCRYPTED.
    ssl_mode = optional(string, "")

    # Instance activation: ALWAYS keeps the instance running (the default
    # posture); NEVER stops it. Flipping ALWAYS→NEVER→ALWAYS is the
    # stop/start lever — a stopped instance keeps its configuration and
    # storage but serves nothing and stops billing for compute. Mind the
    # ordering restrictions (stop read pools before the primary).
    activation_policy = optional(string, "")

    # Enable a public IP on the instance.
    enable_public_ip = optional(bool, false)

    # Enable outbound public IP for the instance.
    enable_outbound_public_ip = optional(bool, false)

    # CIDR ranges allowed to reach the public IP. Requires enable_public_ip.
    authorized_external_networks = optional(list(object({
      cidr_range = string
    })), [])

    # Private Service Connect configuration.
    psc_instance_config = optional(object({
      # Consumer project numbers allowed to create PSC endpoints.
      allowed_consumer_projects = optional(list(string), [])

      # PSC service automation connections.
      psc_auto_connections = optional(list(object({
        # Consumer network, e.g. "projects/vpc-host/global/networks/default".
        consumer_network = optional(string, "")

        # Consumer project ID (not project number).
        consumer_project = optional(string, "")
      })), [])

      # PSC interfaces for outbound connectivity (0 or 1 supported by AlloyDB).
      psc_interface_configs = optional(list(object({
        # Network attachment resource in the consumer project.
        network_attachment_resource = optional(string, "")
      })), [])
    }))

    # User-defined labels on the instance (cost attribution, team ownership,
    # environment tagging). Merged with the platform's attribution labels;
    # on key conflicts the platform labels win. Mutable in place.
    labels = optional(map(string), {})

    # Unstructured metadata stored on the instance (annotations, not labels —
    # not used for billing filtering). Mutable in place.
    annotations = optional(map(string), {})

    # Pin a ZONAL instance to a specific Compute Engine zone (e.g.
    # "us-central1-a"). Only valid when availability_type is ZONAL — GCP
    # rejects it on REGIONAL instances; leave empty to let GCP pick a zone
    # with available capacity. Mutable: changing it live-migrates the
    # instance to the new zone.
    gce_zone = optional(string, "")

    # AlloyDB managed connection pooling (built-in pooler). Mutable in place.
    connection_pool_config = optional(object({
      # Turn managed connection pooling on or off. Mutable in place.
      enabled = optional(bool, false)

      # Pooler flags, keyed by flag name WITHOUT the "connection-pooling-"
      # prefix and with underscores instead of dashes (GCP's documented
      # convention for this provider surface): e.g. the flag
      # "connection-pooling-pool-mode" is set as key "pool_mode". Only
      # applied while enabled is true.
      flags = optional(map(string), {})
    }))

    # Draw this instance's private IPs from a specific Private Service
    # Access allocated range (RFC 1035 name, e.g.
    # "google-managed-services-default") instead of the range the parent
    # cluster uses. Immutable: changing it destroys and recreates the
    # instance.
    allocated_ip_range_override = optional(string, "")

    # What happens to the instance in GCP when this resource is destroyed.
    #   "DELETE"  -- (GCP's default when unset) the instance is deleted;
    #                the parent cluster and its data survive
    #   "PREVENT" -- destroy FAILS; protects serving capacity applications
    #                still connect to
    #   "ABANDON" -- the instance is removed from management but keeps
    #                running (and billing) in GCP
    deletion_policy = optional(string, "")
  })
}
