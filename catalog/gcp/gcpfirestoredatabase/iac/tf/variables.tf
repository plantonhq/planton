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
  description = "GcpFirestoreDatabase specification"
  type = object({
    # GCP project where the Firestore database will be created.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing the project destroys and recreates the database.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Location of the Firestore database. This determines where data is stored
    # and affects latency and availability. Immutable after creation.
    #
    # Multi-region locations: "nam5" (United States), "eur3" (Europe).
    # Single-region locations: any supported GCP region (e.g., "us-east1",
    # "europe-west1").
    #
    # Multi-region locations provide higher availability but at higher cost
    # and with slightly higher write latency.
    location_id = string

    # Name for the Firestore database. Must be 4-63 characters, start with a
    # lowercase letter, contain only lowercase letters, digits, and hyphens,
    # and end with a letter or digit.
    #
    # The special value "(default)" creates the project's default database.
    # Only one default database can exist per project. Client libraries connect
    # to the default database when no database ID is specified.
    #
    # Immutable after creation.
    database_name = string

    # Firestore database type. Determines the data model and API surface.
    #
    # FIRESTORE_NATIVE: Modern Firestore with real-time listeners, offline
    # support, and the Firestore client library API. Recommended for new
    # applications.
    #
    # DATASTORE_MODE: Legacy Datastore-compatible mode with the Datastore
    # client library API. Use for existing Datastore applications or workloads
    # that need Datastore's entity-group transactions.
    type = string

    # Concurrency control mode for the database. Determines how conflicts
    # between concurrent reads and writes are resolved.
    #
    # OPTIMISTIC: Uses optimistic concurrency control. Reads do not block
    # writes, and writes are validated at commit time. Default for
    # FIRESTORE_NATIVE databases.
    #
    # PESSIMISTIC: Uses pessimistic concurrency control. Reads block
    # concurrent writes to the same data. Default for DATASTORE_MODE
    # databases.
    #
    # OPTIMISTIC_WITH_ENTITY_GROUPS: Legacy Datastore mode using entity
    # group-based transactions. Only valid for DATASTORE_MODE databases.
    #
    # If not set, GCP applies the default for the chosen database type.
    concurrency_mode = optional(string, "")

    # Whether to enable point-in-time recovery (PITR) for this database.
    #
    # POINT_IN_TIME_RECOVERY_ENABLED: Retains 7 days of version history.
    # Reads can target any timestamp within the past hour or any 1-minute
    # snapshot within the past 7 days.
    #
    # POINT_IN_TIME_RECOVERY_DISABLED: Retains 1 hour of version history.
    # Reads can target any timestamp within the past hour.
    #
    # If not set, defaults to POINT_IN_TIME_RECOVERY_DISABLED.
    point_in_time_recovery_enablement = optional(string, "")

    # Delete protection for the database. When enabled, the database cannot
    # be deleted through any interface (Console, gcloud, API, IaC tools)
    # until protection is disabled.
    #
    # Defaults to DELETE_PROTECTION_DISABLED.
    delete_protection_state = optional(string)

    # Database edition. Determines the feature set and SLA tier.
    #
    # STANDARD: Default edition suitable for most workloads. Provides
    # standard Firestore features and SLA.
    #
    # ENTERPRISE: Enhanced edition with higher availability SLA, advanced
    # security features, and support for additional data access modes.
    # Requires type to be FIRESTORE_NATIVE.
    #
    # Immutable after creation. If not set, defaults to STANDARD.
    database_edition = optional(string, "")

    # Fully qualified name of the KMS key to use for customer-managed
    # encryption (CMEK). The key must exist in the same location as the
    # database. Immutable after creation.
    #
    # For multi-region databases: nam5 requires a Cloud KMS key in the
    # "us" multi-region; eur3 requires a key in the "europe" multi-region.
    #
    # Format: projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
    #
    # If not set, Google-managed encryption is used (default).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # App Engine integration mode. ENABLED couples the database's
    # lifecycle to the project's App Engine application (a legacy
    # coupling: disabling the App Engine app disables the database with
    # it). DISABLED keeps the database independent — the right choice for
    # everything that is not a legacy App Engine deployment.
    #
    # If not set, GCP applies its default.
    app_engine_integration_mode = optional(string, "")

    # Firestore API data access mode — whether the classic Firestore API
    # can read and write this database. ENTERPRISE edition only. Use
    # DATA_ACCESS_MODE_DISABLED on a database dedicated to the
    # MongoDB-compatible API to keep access single-protocol. At most one of
    # the two data-access modes can be ENABLED on a database (the API
    # rejects both-enabled at create), and realtime updates require THIS
    # mode to be explicitly ENABLED.
    firestore_data_access_mode = optional(string, "")

    # MongoDB-compatible API data access mode — whether MongoDB drivers
    # and tools can read and write this database. ENTERPRISE edition only.
    # Pair with MONGODB_COMPATIBLE_API-scoped GcpFirestoreIndex indexes
    # for query support. Mutually exclusive with an ENABLED
    # firestore_data_access_mode — a database is single-protocol, so a
    # MongoDB-dedicated database can never carry realtime updates.
    mongodb_compatible_data_access_mode = optional(string, "")

    # Realtime updates mode — whether clients can subscribe to live query
    # snapshots on this database. ENTERPRISE edition only. Enabling it
    # requires firestore_data_access_mode DATA_ACCESS_MODE_ENABLED
    # (realtime subscriptions ride the classic Firestore API; leaving the
    # access mode unset does not count as enabled).
    realtime_updates_mode = optional(string, "")

    # Resource Manager tags bound to the database for org-policy and IAM
    # conditions. Keys in the form "tagKeys/{id}", values "tagValues/{id}".
    # Create-time only: changing them later replaces the database.
    resource_manager_tags = optional(map(string), {})

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (the Planton default; note the raw
    #                provider defaults to ABANDON, which would leave the
    #                database running unmanaged after a destroy)
    #   "DELETE"  -- the database is deleted (delete protection, if
    #                enabled, still blocks it)
    #   "PREVENT" -- destroy FAILS; belt-and-suspenders beside
    #                delete_protection_state
    #   "ABANDON" -- the database is removed from management but keeps
    #                running in GCP
    deletion_policy = optional(string)
  })
}
