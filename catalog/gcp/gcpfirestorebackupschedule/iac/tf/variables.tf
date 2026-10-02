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
  description = "GcpFirestoreBackupSchedule specification"
  type = object({
    # GCP project owning the database. Can be a literal project ID or a
    # reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Immutable: changing the project destroys and recreates the schedule.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Firestore database to back up — the database name (a
    # GcpFirestoreDatabase reference resolves to it). Immutable after
    # creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    database = string

    # How long each backup is kept, as a seconds duration string (e.g.
    # "604800s" for 7 days). Maximum 14 weeks ("8467200s"). The only
    # mutable field — extend or shorten protection in place.
    retention = string

    # Take a backup every day. Exactly one of daily or weekly_recurrence
    # must be set. Immutable after creation.
    daily = optional(bool, false)

    # Take a backup every week on the given day. Exactly one of daily or
    # weekly_recurrence must be set. Immutable after creation.
    weekly_recurrence = optional(object({
      # Day of the week the weekly backup runs.
      day = string
    }))

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the schedule is deleted (backups already taken
    #                outlive it either way, aging out per retention)
    #   "PREVENT" -- destroy FAILS; protects a compliance-mandated backup
    #                cadence from accidental teardown
    #   "ABANDON" -- the schedule is removed from management but keeps
    #                taking backups in GCP
    deletion_policy = optional(string, "")
  })
}
