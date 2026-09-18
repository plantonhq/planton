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
  description = "GcpIamCustomRole specification"
  type = object({
    # The identifier for the role, unique within the project. Forms the full
    # role name: projects/<project>/roles/<role_id>.
    # 3-64 characters; letters, digits, underscores, and periods only —
    # hyphens are NOT allowed. Convention is camelCase (e.g. "logBucketWriter").
    # Immutable: changing it destroys and recreates the role, which breaks every
    # grant referencing the old role name.
    role_id = string

    # The GCP project that owns this custom role. The role can only be granted
    # on resources within this project.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing it destroys and recreates the role.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Human-readable title shown in the GCP console and IAM policy pickers
    # (max 100 characters). Mutable.
    title = string

    # Human-readable description of what this role is for and who should hold it.
    # Surfaces in the console next to the title — write it for the operator
    # auditing IAM policies later. Mutable.
    description = optional(string, "")

    # The IAM permissions this role grants, e.g. ["storage.objects.get",
    # "storage.objects.list"]. At least one is required. Mutable: adding or
    # removing permissions updates the role in place, and every existing grant
    # of the role immediately reflects the change — that is the point of
    # defining the bundle once.
    # Permission strings follow <service>.<resource>.<verb>; discover valid
    # values with `gcloud iam list-testable-permissions <resource>`.
    permissions = list(string)

    # The launch stage of the role, mirroring GCP's role lifecycle labels.
    # Purely informational — it does not change what the role can do, but
    # GA (the default) is right for production roles; use DISABLED to keep the
    # role defined while rejecting all of its grants (an IAM kill switch).
    stage = optional(string)

    # What destroying this resource does to the role. Custom-role deletion
    # is a SOFT delete in GCP — the role enters a deleted state (recoverable
    # by undelete for 7 days, purged after 37), so DELETE and ABANDON differ
    # less here than on hard-delete resources:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the role is soft-deleted; existing bindings to it stop
    #                granting access
    #   "PREVENT" -- destroy FAILS; protects a role that live bindings
    #                depend on
    #   "ABANDON" -- the role is removed from management but stays active,
    #                and every binding to it keeps working
    deletion_policy = optional(string, "")
  })
}
