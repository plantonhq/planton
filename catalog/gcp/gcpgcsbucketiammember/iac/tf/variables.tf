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
  description = "GcpGcsBucketIamMember specification"
  type = object({
    # The bucket whose IAM policy receives this grant, by its globally unique
    # name (bucket names are not project-scoped, so there is no project
    # field). Reference a GcpGcsBucket -- its `bucket_id` output is exactly
    # this value.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    bucket = string

    # The role to grant on the bucket. A predefined role
    # ("roles/storage.objectCreator", "roles/storage.objectViewer",
    # "roles/storage.objectUser", "roles/storage.objectAdmin",
    # "roles/storage.legacyBucketReader", ...) or a custom role's full name
    # ("projects/<project>/roles/<role_id>" or
    # "organizations/<org>/roles/<role_id>"). Reference a GcpIamCustomRole to
    # grant a custom role -- its `name` output is exactly this value.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    role = string

    # The identity receiving the grant, in IAM member format:
    #   serviceAccount:<email>  -- a service account or Google service agent.
    #                              Reference GcpServiceAccount (`member`) or a
    #                              GcpLoggingSink (`writer_identity`, the
    #                              sink's writer, already in this format).
    #   user:<email>, group:<email>, domain:<domain>
    #   principal://... / principalSet://... -- workload identity federation
    #   allUsers / allAuthenticatedUsers     -- PUBLIC access to the objects
    #                                           the role covers; refused by
    #                                           buckets with public access
    #                                           prevention enforced
    # Grants to deleted principals ("deleted:...") are not supported.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    member = string

    # Optional IAM Condition restricting when this grant applies (an expiry,
    # or an object-name prefix such as
    # resource.name.startsWith("projects/_/buckets/<bucket>/objects/logs/")).
    # Conditions require the bucket to use uniform bucket-level access. The
    # condition is part of the grant's identity: the same role with and
    # without a condition are two independent grants.
    condition = optional(object({
      # Short human-readable title naming the condition's intent, e.g.
      # "logs-prefix-only".
      title = string

      # The CEL condition expression.
      expression = string

      # Optional longer explanation of what the condition does and why.
      description = optional(string, "")
    }))
  })
}
