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
  description = "GcpSpannerDatabase specification"
  type = object({
    # The GCP project that owns the parent Spanner instance. Accepts a
    # literal project ID or a reference to a GcpProject resource. If
    # omitted, the provider's default project is used.
    # Immutable: changing the project destroys and recreates the database.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Spanner instance to create the database on. This determines the
    # compute capacity and geographic configuration available to the
    # database. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    instance = string

    # Unique name of the database within the instance. Immutable. If not
    # specified, defaults to metadata.name. Must be 2-30 characters: start
    # with a lowercase letter, contain lowercase letters, digits,
    # underscores, and hyphens, and end with a letter or digit.
    database_name = optional(string, "")

    # SQL dialect of the database. Immutable — this choice is permanent.
    #
    # GOOGLE_STANDARD_SQL (default): Google's SQL dialect with full Spanner
    # feature support, including interleaved tables and STRUCT types.
    #
    # POSTGRESQL: PostgreSQL-compatible interface for teams standardized on
    # PostgreSQL syntax and tooling. Some Spanner-specific features are not
    # available through this dialect.
    database_dialect = optional(string, "")

    # Retention period for database versions, enabling point-in-time
    # recovery. Between 1 hour and 7 days; accepts duration formats such as
    # "1h", "24h", "3d", "86400s". GCP defaults to "1h". Mutable — a longer
    # window widens point-in-time recovery at the cost of extra storage.
    version_retention_period = optional(string, "")

    # DDL statements executed when creating the database (tables, indexes,
    # views). Statements run atomically with creation — if any fails, the
    # database is not created.
    #
    # Lifecycle: APPEND-ONLY. New statements added later are applied via
    # UpdateDDL; modifying or removing an existing entry forces database
    # recreation. Use a migration tool for ongoing schema management.
    ddl = optional(list(string), [])

    # GCP API-side drop protection. While true, the database cannot be
    # deleted through ANY interface (Console, gcloud, API, Terraform,
    # Pulumi) and the parent Spanner instance cannot be deleted either.
    # Mutable. Defaults to false — prefer deletion_protection (below) for
    # day-to-day IaC safety and reserve this for compliance-grade locks.
    enable_drop_protection = optional(bool, false)

    # Customer-managed encryption (CMEK). Immutable. If omitted,
    # Google-managed encryption is used. Exactly one key shape inside:
    # kms_key_name for regional instances, kms_key_names (one per region)
    # for multi-region instances.
    encryption_config = optional(object({
      # Fully qualified name of the KMS key for single-region CMEK.
      # Format: projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_name = optional(string, "")

      # Fully qualified KMS key names for multi-region CMEK — one key per
      # region of the instance's multi-region configuration.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_names = optional(list(string), [])
    }))

    # Default time zone for the database, affecting time-zone-dependent SQL
    # functions. Must be a valid IANA Time Zone Database name (e.g.
    # "America/New_York", "UTC"). GCP defaults to "America/Los_Angeles".
    default_time_zone = optional(string, "")

    # IaC-side deletion guard. While true (the default), BOTH engines refuse
    # to destroy this resource — a plan that would delete the database fails
    # before touching GCP. Set false explicitly before an intentional
    # teardown. Unlike enable_drop_protection, GCP itself does not enforce
    # this: a console/gcloud delete would still succeed.
    deletion_protection = optional(bool)

    # Deletion policy — what a PERMITTED destroy does, once the two guards
    # above allow one (deletion_protection false, enable_drop_protection
    # off):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the database and its data are deleted (backups
    #                already taken survive until their retention expires)
    #   "PREVENT" -- destroy FAILS — a third, explicit wall for databases
    #                whose teardown must never ride along with a stack's
    #   "ABANDON" -- the database is removed from management but left
    #                intact in GCP, still billing for its storage
    deletion_policy = optional(string, "")
  })
}
