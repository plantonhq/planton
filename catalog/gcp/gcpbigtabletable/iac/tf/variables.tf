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
  description = "GcpBigtableTable specification"
  type = object({
    # GCP project owning the Bigtable instance. Can be a literal project
    # ID or a reference to a GcpProject resource. If omitted, the
    # provider's default project is used.
    # Immutable: changing the project destroys and recreates the table.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Bigtable instance this table lives in — the instance's short
    # name (a GcpBigtableInstance reference resolves to it). Immutable
    # after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    instance = string

    # Name for the table (1-50 characters). If not specified, defaults to
    # metadata.name. Immutable after creation.
    table_name = optional(string, "")

    # Column families with their GC policies. Mutable: append families as
    # the application grows. At least one family is the practical minimum —
    # a table without families cannot store data — but the API allows an
    # empty table, so this is not enforced.
    column_families = optional(list(object({
      # Column family name (unique within the table).
      family = string

      # Value type for aggregate families. The shortcuts intsum, intmin,
      # intmax, and inthll declare server-side aggregate cells (e.g. a
      # counter incremented atomically at write time); a raw JSON Type
      # object expresses anything else. Leave empty for regular families.
      type = optional(string, "")

      # Garbage-collection policy for this family. Strongly recommended:
      # without one, every cell version accumulates forever. Managed as the
      # API's own per-family GC policy object, so policy changes never touch
      # the table or its data.
      gc_policy = optional(object({
        # Combination mode when both max_age and max_versions are set.
        # UNION: collect when either condition is met (most common).
        # INTERSECTION: collect only when both conditions are met.
        mode = optional(string, "")

        # Drop cells older than this duration (e.g. "720h" for 30 days,
        # "8760h" for a year). Duration string in Go format.
        max_age = optional(string, "")

        # Keep only the newest N versions of each cell. 0 means unset.
        max_versions = optional(number, 0)

        # Raw JSON GC rule tree for nested policies the typed fields cannot
        # express (e.g. a union of an intersection and an age rule). Mutually
        # exclusive with mode/max_age/max_versions. See the Bigtable Admin
        # API's GcRule JSON format.
        gc_rules = optional(string, "")

        # Allow a policy change that EXPANDS what is eligible for collection
        # on a replicated (multi-cluster) instance — Bigtable otherwise
        # rejects it as a safety measure against surprise data loss.
        ignore_warnings = optional(bool, false)
      }))
    })), [])

    # Row keys to pre-split the table at, so initial load distributes
    # across tablets instead of hammering one server (e.g. user prefixes
    # "user1", "user5", "user9"). Immutable: changing this REPLACES the
    # table and its data — set it right at creation, or manage splits
    # operationally.
    split_keys = optional(list(string), [])

    # Retain change stream data (a CDC feed consumable by Dataflow) for
    # this duration, between 1 and 7 days (e.g. "24h0m0s"). Empty disables
    # change streams; setting "0" on an existing table disables them.
    change_stream_retention = optional(string, "")

    # Built-in automated backups: how often to take them and how long to
    # keep them. Omit to leave automated backups off.
    automated_backup_policy = optional(object({
      # How long automated backups are retained (e.g. "72h" for 3 days).
      # Duration string in Go format.
      retention_period = string

      # How often automated backups are taken (e.g. "24h" for daily).
      # Duration string in Go format.
      frequency = string

      # Cloud Bigtable zones where automated backups are ALLOWED to be
      # created, each in the format "projects/{project}/locations/{zone}".
      # Empty means backups may be created in all zones of the instance.
      # Can only be set for tables in ENTERPRISE_PLUS instances (see the
      # GcpBigtableInstance edition field).
      locations = optional(list(string), [])
    }))

    # API-side deletion guard. PROTECTED (the default): the table cannot
    # be deleted by any client until this is set UNPROTECTED first — the
    # safety default for a data-bearing resource, sent explicitly by both
    # IaC engines so destroy behavior never depends on the engine.
    deletion_protection = optional(string)

    # Structured row key schema as the API's Type JSON (declares how row
    # keys decompose into typed fields for SQL queries and change
    # streams). In-place update is not supported by the API: to change an
    # existing schema, clear this field, apply, then set the new schema
    # and apply again. Byte delimiters must be base64-encoded.
    row_key_schema = optional(string, "")

    # Deletion policy — what a destroy does to BOTH objects this component
    # manages: the table and its per-family GC policies. Applies only once
    # deletion_protection (above) is UNPROTECTED:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the table (and its data) is deleted; GC policies go
    #                with it
    #   "PREVENT" -- destroy FAILS; a second wall for data-bearing tables
    #   "ABANDON" -- the table is removed from management but left intact
    #                in Bigtable — also the escape hatch when a GC-policy
    #                delete is rejected on a replicated instance
    deletion_policy = optional(string, "")
  })
}
