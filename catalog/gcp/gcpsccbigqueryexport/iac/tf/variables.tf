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
  description = "GcpSccBigQueryExport specification"
  type = object({
    # Whose findings are exported. Omit for the provider's default project.
    scope = optional(object({
      # A project: a literal project ID or a GcpProject reference.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      project_id = optional(string, "")

      # A folder: the folder's numeric ID -- a literal or a GcpFolder
      # reference. Covers every project beneath it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      folder_id = optional(string, "")

      # The organization: the numeric organization ID, without the
      # organizations/ prefix.
      organization_id = optional(string, "")
    }))

    # The export's ID, unique within its parent: lowercase letters, digits,
    # and hyphens, starting with a letter and ending with a letter or digit,
    # at most 63 characters. Immutable.
    big_query_export_id = string

    # The dataset findings are written to: a literal
    # projects/{project}/datasets/{dataset} or a GcpBigQueryDataset
    # reference (the module trims the dataset's self link to Google's form).
    # Dataset IDs use letters, digits, and underscores only. Required: an
    # export without a dataset writes nothing.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    dataset = string

    # Which finding create and update events are exported, in the same
    # syntax as notification filters:
    #   state = "ACTIVE" AND NOT mute = "MUTED"
    # Empty exports every finding.
    filter = optional(string, "")

    # What the export is for, up to 1024 characters.
    description = optional(string, "")

    # Where the export configuration is stored. "global", the default,
    # unless Security Command Center data residency was set up at activation
    # (then the residency location, e.g. "eu" or "us").
    location = optional(string, "")

    # What destroying this block does:
    #   "" / "DELETE" -- the export is deleted; the dataset and the rows
    #                    already written stay
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the export leaves management and keeps writing
    deletion_policy = optional(string, "")
  })
}
