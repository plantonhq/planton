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
  description = "GcpSccMuteConfig specification"
  type = object({
    # Whose findings the rule mutes. Omit for the provider's default project.
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

    # The rule's ID, unique within its parent: lowercase letters, digits,
    # and hyphens, starting with a letter and ending with a letter or digit,
    # at most 63 characters. Immutable.
    mute_config_id = string

    # Which findings are muted. Supported fields: severity, category,
    # resource.name, resource.project_name, resource.project_display_name,
    # resource.folders.resource_folder, resource.parent_name,
    # resource.parent_display_name, resource.type, findingClass,
    # indicator.ip_addresses, indicator.domains -- each with = or : (substring),
    # combined with AND and OR. Write it for the scope: a filter naming
    # project X on a rule scoped to project Y matches nothing. Example:
    #   category = "PUBLIC_BUCKET_ACL" AND resource.project_display_name = "sandbox"
    filter = string

    # How the rule mutes:
    #   "DYNAMIC" -- applies to existing and future matching findings, and
    #                stops applying when the rule changes or is deleted, or a
    #                finding stops matching (Google's recommendation)
    #   "STATIC"  -- sets a permanent mute on future matching findings only;
    #                later rule changes do not unmute them
    # Google treats the type as immutable after creation.
    type = string

    # Why the rule exists -- the accepted risk or the noise it removes.
    description = optional(string, "")

    # Where the rule is stored. "global", the default, unless Security
    # Command Center data residency was set up at activation (then the
    # residency location, e.g. "eu" or "us").
    location = optional(string, "")

    # What destroying this block does:
    #   "" / "DELETE" -- the rule is deleted (DYNAMIC mutes are lifted)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the rule leaves management and keeps muting
    deletion_policy = optional(string, "")
  })
}
