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
  description = "AwsMemorydbAcl specification"
  type = object({
    # The AWS region where the ACL is created. Must match the region of every
    # member user and of every cluster the ACL attaches to.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The users that make up this ACL, referenced by user name (the value
    # AwsMemorydbUser exports as `status.outputs.user_name`). Membership
    # updates in place; the ACL is the single place an application's cluster
    # access is granted or revoked. May be empty — a cluster attached to an
    # empty ACL simply accepts no authenticated connections. Deleting a user
    # removes it from every ACL server-side automatically; the provider
    # reconciles that, so dropping an already-deleted member from this list
    # is a clean no-op, never an error.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    user_names = optional(list(string), [])
  })
}
