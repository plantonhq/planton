variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for the GCP Memorystore (Valkey) instance"
  type = object({
    # The GCP project for the instance. The CLI's tfvars converter
    # resolves StringValueOrRef fields to their literal string before the
    # module runs, so this arrives as a plain string.
    # If empty, the provider's default project is used (see locals.tf).
    project_id = optional(string, "")

    # Instance ID (GCP resource name). Immutable (ForceNew).
    instance_name = string

    # Region (e.g. us-central1). Immutable (ForceNew).
    location = string

    # Number of shards (>= 1). Mutable — resharding happens in place.
    shard_count = number

    # CLUSTER or CLUSTER_DISABLED. Empty leaves GCP's default.
    # Immutable (ForceNew).
    mode = optional(string, "")

    # Node type (SHARED_CORE_NANO / STANDARD_SMALL / HIGHMEM_MEDIUM /
    # HIGHMEM_XLARGE). Empty leaves GCP's default.
    node_type = optional(string, "")

    # Engine version (e.g. VALKEY_8_0). Empty leaves GCP's default.
    engine_version = optional(string, "")

    # Free-form engine parameters (e.g. maxmemory-policy).
    engine_configs = optional(map(string), {})

    # Read replicas per shard (0-5). 0 means no replicas.
    replica_count = optional(number, 0)

    # PSC auto-created endpoints. network arrives as the VPC's relative
    # resource path (projects/{p}/global/networks/{n}) — the only format
    # the API accepts. An empty per-entry project_id resolves to the
    # provider's effective project (see main.tf). Immutable (ForceNew).
    psc_auto_connections = optional(list(object({
      network    = string
      project_id = optional(string, "")
    })), [])

    # AUTH_DISABLED or IAM_AUTH. Immutable (ForceNew).
    authorization_mode = optional(string, "")

    # TRANSIT_ENCRYPTION_DISABLED or SERVER_AUTHENTICATION.
    # Immutable (ForceNew).
    transit_encryption_mode = optional(string, "")

    # CMEK key resource ID. Empty means Google-managed encryption.
    # Immutable (ForceNew).
    kms_key = optional(string, "")

    # RDB / AOF persistence.
    persistence_config = optional(object({
      mode = string
      rdb_config = optional(object({
        rdb_snapshot_period     = string
        rdb_snapshot_start_time = optional(string, "")
      }), null)
      aof_config = optional(object({
        append_fsync = string
      }), null)
    }), null)

    # MULTI_ZONE or SINGLE_ZONE placement. Immutable (ForceNew).
    zone_distribution_config = optional(object({
      mode = string
      zone = optional(string, "")
    }), null)

    # Weekly maintenance window (1h, UTC; starts on the hour — the API
    # supports no finer granularity).
    maintenance_policy = optional(object({
      weekly_maintenance_window = object({
        day  = string
        hour = number
      })
    }), null)

    # Daily automated backups with retention.
    automated_backup_config = optional(object({
      start_hour = number
      retention  = string
    }), null)

    # Cross-region DR: PRIMARY replicates to secondaries; SECONDARY
    # replicates from primary_instance. Instance references arrive as
    # full resource paths (the other instance's name output).
    cross_instance_replication_config = optional(object({
      instance_role = string
      primary_instance = optional(object({
        instance = string
      }), null)
      secondary_instances = optional(list(object({
        instance = string
      })), [])
    }), null)

    # Seed data from RDB files in GCS at creation. Mutually exclusive
    # with managed_backup_source. Immutable (ForceNew).
    gcs_source = optional(object({
      uris = list(string)
    }), null)

    # Seed data from a managed backup at creation. Mutually exclusive
    # with gcs_source. Immutable (ForceNew).
    managed_backup_source = optional(object({
      backup = string
    }), null)

    # User labels merged beneath Planton platform labels (platform keys
    # win on conflict).
    labels = optional(map(string), {})

    # Deletion protection. The spec defaults this to true (Planton
    # middleware materializes the default), and the module sends it
    # explicitly so destroy behavior is identical on both engines.
    deletion_protection_enabled = optional(bool, true)

    # Which CA signs the server certificate for the TLS-enabled instance.
    # Empty rides GCP's default (GOOGLE_MANAGED_PER_INSTANCE_CA).
    # Immutable (ForceNew).
    server_ca_mode = optional(string, "")

    # Certificate Authority Service CA pool, consumed only when
    # server_ca_mode is CUSTOMER_MANAGED_CAS_CA. Immutable (ForceNew).
    server_ca_pool = optional(string, "")

    # Self-service maintenance version — set to a newer available version
    # to apply maintenance on your schedule instead of GCP's rollout.
    # Update-only; downgrades are rejected.
    maintenance_version = optional(string, "")

    # Deletion policy: "", "DELETE" (default), "PREVENT" (destroy fails),
    # or "ABANDON" (remove from management, leave running in GCP).
    deletion_policy = optional(string, "")

    # Memorystore ACL policy to attach (full resource name
    # projects/{project}/locations/{region}/aclPolicies/{id}). Empty keeps
    # the instance's built-in default ACL. Mutable in place.
    acl_policy = optional(string, "")
  })

  validation {
    condition     = var.spec.acl_policy == "" || can(regex("^projects/[^/]+/locations/[^/]+/aclPolicies/[^/]+$", var.spec.acl_policy))
    error_message = "acl_policy must be empty or a full resource name of the form projects/{project}/locations/{region}/aclPolicies/{aclPolicyId}."
  }

  validation {
    condition     = contains(["", "DELETE", "PREVENT", "ABANDON"], var.spec.deletion_policy)
    error_message = "deletion_policy must be one of: DELETE, PREVENT, ABANDON."
  }

  validation {
    condition     = var.spec.instance_name != ""
    error_message = "instance_name is required."
  }

  validation {
    condition     = var.spec.location != ""
    error_message = "location is required."
  }

  validation {
    condition     = var.spec.shard_count >= 1
    error_message = "shard_count must be at least 1."
  }

  validation {
    condition     = !(var.spec.gcs_source != null && var.spec.managed_backup_source != null)
    error_message = "gcs_source and managed_backup_source are mutually exclusive — choose one seed source."
  }
}
