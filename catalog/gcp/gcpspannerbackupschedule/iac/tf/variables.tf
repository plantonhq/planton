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
  description = "GcpSpannerBackupSchedule specification"
  type = object({
    # The GCP project that owns the parent Spanner instance. Accepts a
    # literal project ID or a reference to a GcpProject resource. If
    # omitted, the provider's default project is used.
    # Immutable: changing the project destroys and recreates the schedule.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Spanner instance hosting the database. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    instance = string

    # The database this schedule backs up. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    database = string

    # Unique name of the backup schedule within the database. Immutable.
    # If not specified, defaults to metadata.name. Must start with a
    # lowercase letter, contain lowercase letters, digits, and hyphens, and
    # end with a letter or digit.
    schedule_name = optional(string, "")

    # Crontab expression for when backups are created, evaluated in UTC
    # (maps to the provider's spec.cron_spec.text). Spanner accepts a
    # bounded set of frequencies — every 12 hours, daily, weekly, or
    # monthly. Examples:
    #   "0 2/12 * * *" — every 12 hours at 02:00 and 14:00 UTC
    #   "0 2 * * *"    — daily at 02:00 UTC
    #   "0 2 * * 0"    — weekly on Sunday at 02:00 UTC
    #   "0 2 8 * *"    — monthly on the 8th at 02:00 UTC
    # Mutable — cadence changes apply in place.
    cron = string

    # How long each backup is retained, as a seconds duration string ending
    # in 's' (e.g. "86400s" = 1 day, "2678400s" = 31 days). Maximum 366 days
    # ("31622400s"). Mutable — applies to backups created AFTER the change.
    retention_duration = string

    # The kind of backups the schedule creates. Immutable.
    #
    # FULL (default): every backup is a complete, self-contained copy.
    #
    # INCREMENTAL: backups form chains storing only changes since the
    # previous backup — significantly cheaper storage at the same restore
    # semantics. Requires the instance to be ENTERPRISE or ENTERPRISE_PLUS
    # edition.
    backup_type = optional(string)

    # How the backups are encrypted. If omitted, backups use
    # USE_DATABASE_ENCRYPTION (inherit the database's posture). Mutable.
    encryption_config = optional(object({
      # How backups are encrypted. Required when this message is present.
      encryption_type = string

      # Fully qualified KMS key for single-region backup CMEK.
      # Format: projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
      # Only valid with CUSTOMER_MANAGED_ENCRYPTION.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_name = optional(string, "")

      # Fully qualified KMS keys for multi-region backup CMEK — one key per
      # region of the instance's multi-region configuration.
      # Only valid with CUSTOMER_MANAGED_ENCRYPTION.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_names = optional(list(string), [])
    }))

    # Deletion policy for the schedule — what happens when this resource
    # is destroyed. The schedule is a control-plane object: none of these
    # values touches the BACKUPS it already created (those live until
    # their retention expires):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the schedule is deleted; no further backups are taken
    #   "PREVENT" -- destroy FAILS; protects the cadence a recovery
    #                objective depends on from riding along with a
    #                stack teardown
    #   "ABANDON" -- the schedule is removed from management but keeps
    #                running (and creating backups) in GCP
    deletion_policy = optional(string, "")
  })
}
