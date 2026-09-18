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
  description = "AwsElasticacheUserGroup specification"
  type = object({
    # The AWS region where the user group is created. Must match the region
    # of every member user and of every cache the group attaches to.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Cache engine this group serves. Values: "redis", "valkey". A group only
    # accepts users whose engine matches, and only attaches to caches of the
    # same engine family. Create-time immutable — changing it destroys and
    # recreates the group.
    engine = string

    # The user ids that belong to this group. AWS refuses to create a group
    # without a user whose user NAME is "default" — include one (typically a
    # locked-down default user) alongside the per-application users.
    # Membership updates in place; the group is the single place an
    # application's cache access is granted or revoked.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    user_ids = list(string)
  })
}
