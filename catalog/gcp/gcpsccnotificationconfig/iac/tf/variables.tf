variable "metadata" {
  description = "Catalog object metadata"
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
  description = "GcpSccNotificationConfig specification"
  type = object({
    # Whose findings are streamed. Omit for the provider's default project.
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

    # The config's ID, unique within its parent: 1-128 letters, digits,
    # hyphens, or underscores. Immutable.
    config_id = string

    # What the config is for, up to 1024 characters.
    description = optional(string, "")

    # The topic findings are published to: a literal
    # projects/{project}/topics/{topic} or a GcpPubSubTopic reference.
    # Required on folder and organization configs; Google lets a project
    # config omit it. Grant the config's publisher on it with a
    # GcpPubSubTopicIamMember (role roles/pubsub.publisher, member
    # referencing the service_account_member output).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    pubsub_topic = optional(string, "")

    # Which finding create and update events are streamed (Google's
    # streaming_config.filter). Restrictions of the form
    # `<field> <operator> <value>`, combined with AND and OR (OR binds
    # tighter), negated with a leading "-":
    #   state = "ACTIVE" AND severity = "HIGH"
    #   category = "OPEN_FIREWALL" AND state = "ACTIVE"
    # Operators: = for every type; >, <, >=, <= for integers; : for substring
    # match. Empty streams every finding.
    filter = optional(string, "")

    # Where the config is stored. "global", the default, unless Security
    # Command Center data residency was set up at activation (then the
    # residency location, e.g. "eu" or "us").
    location = optional(string, "")

    # What destroying this block does:
    #   "" / "DELETE" -- the config is deleted and streaming stops
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the config leaves management and keeps streaming
    deletion_policy = optional(string, "")
  })
}
