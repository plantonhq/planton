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
  description = "Specification for the GCS bucket"
  type = object({
    # StringValueOrRef fields arrive from the proto→tfvars converter as
    # plain strings (already resolved), never as object({value}).
    # Empty falls back to the provider's default project.
    project_id = optional(string, "")

    # Globally unique bucket name. Immutable (ForceNew).
    bucket_name = string

    # Region, dual-region, or multi-region. Immutable (ForceNew).
    location = string

    # Default storage class for new objects: STANDARD (default when
    # empty), NEARLINE, COLDLINE, ARCHIVE. Mutable in place.
    storage_class = optional(string, "")

    # Delete all contained objects when destroying the bucket. Defaults to
    # false: destroying a non-empty bucket fails instead of erasing data.
    force_destroy = optional(bool, false)

    # IAM-only access control (UBLA). GCP's default is false.
    uniform_bucket_level_access_enabled = optional(bool, false)

    # "" / "inherited" / "enforced". Mutable in place.
    public_access_prevention = optional(string, "")

    # Keep noncurrent object versions on overwrite/delete. Mutable.
    versioning_enabled = optional(bool, false)

    # Automatic per-object storage-class management.
    # The API requires the switch inside the block; no literal default, so an
    # unset value can never be mistaken for an explicit false.
    autoclass = optional(object({
      enabled = optional(bool)
      # "" (GCP default NEARLINE) / "NEARLINE" / "ARCHIVE".
      terminal_storage_class = optional(string, "")
    }), null)

    # Lifecycle rules (max 100). Numeric conditions are nullable so an
    # explicit 0 (a meaningful value) is distinguishable from unset.
    lifecycle_rules = optional(list(object({
      action = object({
        # "Delete" / "SetStorageClass" / "AbortIncompleteMultipartUpload".
        type = string
        # Target class, only for SetStorageClass.
        storage_class = optional(string, "")
      })
      condition = object({
        age_days                   = optional(number)
        created_before             = optional(string, "")
        with_state                 = optional(string, "")
        matches_storage_class      = optional(list(string), [])
        matches_prefix             = optional(list(string), [])
        matches_suffix             = optional(list(string), [])
        num_newer_versions         = optional(number)
        days_since_noncurrent_time = optional(number)
        noncurrent_time_before     = optional(string, "")
        days_since_custom_time     = optional(number)
        custom_time_before         = optional(string, "")
        size_above_bytes           = optional(number)
        size_below_bytes           = optional(number)
      })
    })), [])

    # WORM retention. is_locked is IRREVERSIBLE once applied.
    retention_policy = optional(object({
      retention_period_seconds = number
      is_locked                = optional(bool, false)
    }), null)

    # Recovery window for deleted objects. GCP defaults to 7 days even
    # when the block is omitted; 0 disables soft delete.
    soft_delete_policy = optional(object({
      retention_duration_seconds = optional(number)
    }), null)

    # Default CMEK key path (resolved from a GcpKmsKey reference). Empty
    # means Google-managed encryption. Mutable in place.
    kms_key_name = optional(string, "")

    # Callers pay for access/egress. Mutable in place.
    requester_pays = optional(bool, false)

    # Auto-apply an event-based hold to every new object. Mutable.
    default_event_based_hold = optional(bool, false)

    # Allow per-object retention configuration. Create-time only
    # (ForceNew).
    enable_object_retention = optional(bool, false)

    # Static website serving.
    website = optional(object({
      main_page_suffix = optional(string, "")
      not_found_page   = optional(string, "")
    }), null)

    # Direct browser cross-origin access rules.
    cors_rules = optional(list(object({
      origins          = list(string)
      methods          = list(string)
      response_headers = optional(list(string), [])
      max_age_seconds  = optional(number, 0)
    })), [])

    # Access-log delivery to another bucket (resolved from a GcpGcsBucket
    # reference or a literal bucket name).
    logging = optional(object({
      log_bucket        = string
      log_object_prefix = optional(string, "")
    }), null)

    # Custom dual-region placement: exactly two regions. Immutable
    # (ForceNew).
    custom_placement_config = optional(object({
      data_locations = list(string)
    }), null)

    # "" / "DEFAULT" / "ASYNC_TURBO" (turbo replication; dual-region
    # only). Mutable in place.
    rpo = optional(string, "")

    # Real folder semantics (atomic renames). Requires UBLA; excludes
    # versioning. Create-time only (ForceNew).
    hierarchical_namespace_enabled = optional(bool, false)

    # User labels, merged beneath the platform attribution labels
    # (see locals.tf).
    labels = optional(map(string), {})

    # Additive IAM grants: one role to one member each, composing safely
    # with grants made elsewhere.
    iam_members = optional(list(object({
      role = string
      # Resolved from a GcpServiceAccount reference or a literal IAM
      # member string (serviceAccount:..., user:..., allUsers, ...).
      member = string
      condition = optional(object({
        title       = string
        expression  = string
        description = optional(string, "")
      }), null)
    })), [])

    # Network-layer IP filtering: which CIDR ranges / VPC networks may
    # reach the bucket, evaluated before IAM. Mutable in place.
    ip_filter = optional(object({
      # "Enabled" / "Disabled".
      mode                           = string
      allow_cross_org_vpcs           = optional(bool, false)
      allow_all_service_agent_access = optional(bool, false)
      public_network_source = optional(object({
        allowed_ip_cidr_ranges = list(string)
      }), null)
      vpc_network_sources = optional(list(object({
        # Resolved from a GcpVpcNetwork reference or a literal
        # projects/{project}/global/networks/{name} path.
        network                = string
        allowed_ip_cidr_ranges = list(string)
      })), [])
    }), null)

    # Encryption-type enforcement for NEW objects. Each mode is
    # "" (NotRestricted) / "NotRestricted" / "FullyRestricted".
    encryption_enforcement = optional(object({
      google_managed_restriction_mode    = optional(string, "")
      customer_managed_restriction_mode  = optional(string, "")
      customer_supplied_restriction_mode = optional(string, "")
    }), null)

    # Destroy-time guard: "" / "DELETE" deletes, "PREVENT" fails the
    # destroy, "ABANDON" unmanages without deleting. Also applied to
    # folders and managed folders (the notification config carries no
    # deletion policy — it is immutable and delete-only).
    deletion_policy = optional(string, "")

    # Folders — real directories, hierarchical-namespace buckets only.
    # The API never auto-creates parents: list every ancestor as its own
    # entry ("logs/2026/" needs "logs/" too). Max 5 nesting levels.
    folders = optional(list(object({
      # Path WITH trailing slash: "logs/", "logs/2026/".
      name = string
      # Sweep contained objects (client-side) on destroy. False fails
      # the destroy of a non-empty folder instead of erasing data.
      force_destroy = optional(bool, false)
    })), [])

    # Managed folders — prefix-scoped IAM anchors; require UBLA, not HNS.
    # Independent of each other (no parent requirement).
    managed_folders = optional(list(object({
      # Path WITH trailing slash: "reports/".
      name = string
      # Server-side allowNonEmpty on destroy: objects SURVIVE, they just
      # stop being covered by the managed folder's IAM.
      force_destroy = optional(bool, false)
    })), [])

    # Pub/Sub notification configs. IMMUTABLE end to end — every change
    # replaces the config. The GCS service agent needs
    # roles/pubsub.publisher on the topic BEFORE create (composed grant).
    notifications = optional(list(object({
      # Fully-qualified projects/{project}/topics/{name} (resolved from a
      # GcpPubSubTopic reference or a literal path).
      topic = string
      # "JSON_API_V1" / "NONE".
      payload_format = string
      # Empty means ALL event types.
      event_types        = optional(list(string), [])
      object_name_prefix = optional(string, "")
      custom_attributes  = optional(map(string), {})
    })), [])
  })

  validation {
    condition     = var.spec.bucket_name != ""
    error_message = "bucket_name is required."
  }

  validation {
    condition     = var.spec.location != ""
    error_message = "location is required."
  }

  validation {
    condition     = can(regex("^[a-z0-9]([a-z0-9-._]*[a-z0-9])?$", var.spec.bucket_name)) && length(var.spec.bucket_name) >= 3 && length(var.spec.bucket_name) <= 63
    error_message = "bucket_name must be 3-63 characters of lowercase letters, numbers, hyphens, or dots, starting and ending with a letter or number."
  }
}
