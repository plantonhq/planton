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
  description = "GcpBigQueryReservationGroup specification"
  type = object({
    # The administration project the group lives in (the same one as its
    # reservations): a literal project ID or a GcpProject reference. If
    # omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The location of the group and its reservations: a multi-region (US,
    # EU) or a region. Empty means US (Google's default). Immutable.
    location = optional(string, "")

    # The group's name -- letters, digits, and dashes. Defaults to
    # metadata.name. Immutable.
    reservation_group_name = optional(string, "")

    # What happens to the group when this resource is destroyed:
    #   "" / "DELETE" -- deleted (remove its reservations from it first)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the group leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
