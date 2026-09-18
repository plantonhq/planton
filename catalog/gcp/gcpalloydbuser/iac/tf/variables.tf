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
  description = "GcpAlloydbUser specification"
  type = object({
    # The GCP project that owns the AlloyDB cluster.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The AlloyDB cluster this user lives on. Accepts the full cluster resource
    # path or a reference to a GcpAlloydbCluster resource. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    cluster = string

    # The database role name of the user. Immutable.
    user_id = string

    # User authentication type. ALLOYDB_BUILT_IN (default) uses a password;
    # ALLOYDB_IAM_USER authenticates through IAM without a stored password.
    # Immutable.
    user_type = optional(string)

    # Password for ALLOYDB_BUILT_IN users. Mutable — updating rotates in place.
    # Never set for ALLOYDB_IAM_USER.
    password = optional(string, "")

    # Database roles granted to this user (e.g. "alloydbiamuser", "alloydbsuperuser").
    database_roles = optional(list(string), [])

    # What happens to the database user in GCP when this resource is destroyed.
    #   "DELETE"  -- (GCP's default when unset) the user is dropped from the
    #                cluster; objects it owns inside PostgreSQL keep their
    #                ownership rows, so reassign ownership first
    #   "PREVENT" -- destroy FAILS; protects a credential applications still
    #                authenticate with
    #   "ABANDON" -- the user is removed from management but keeps existing
    #                (and authenticating) on the cluster
    deletion_policy = optional(string, "")
  })
}
