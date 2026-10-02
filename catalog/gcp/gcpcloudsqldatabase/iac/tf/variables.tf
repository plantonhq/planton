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
  description = "GcpCloudSqlDatabase specification"
  type = object({
    # The GCP project that owns the Cloud SQL instance.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Cloud SQL instance hosting this database. Accepts the instance name
    # or a reference to a GcpCloudSql resource. Immutable — a database cannot
    # move between instances.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    instance = string

    # Name of the database inside the instance. Immutable.
    # Example: "orders", "analytics_staging"
    database_name = string

    # Character set. MySQL: e.g. "utf8mb4" (recommended). PostgreSQL: must be
    # "UTF8" at creation. Ignored by SQL Server. If empty, the engine default
    # applies.
    charset = optional(string, "")

    # Collation. MySQL: e.g. "utf8mb4_0900_ai_ci". PostgreSQL: an OS locale
    # such as "en_US.UTF8". SQL Server: a SQL Server collation name. If
    # empty, the engine default applies.
    collation = optional(string, "")

    # Engine-side teardown behavior. "DELETE" (default) drops the database;
    # "PREVENT" fails any plan that would drop it; "ABANDON" removes it
    # from IaC management while leaving it in the instance. ABANDON is the
    # documented answer for PostgreSQL databases that cannot be dropped
    # while clients hold connections.
    deletion_policy = optional(string, "")
  })
}
