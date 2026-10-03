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
  description = "GcpPubSubTopicIamMember specification"
  type = object({
    # The topic whose IAM policy receives this grant, by its full resource
    # name: projects/<project>/topics/<topic>. Reference a GcpPubSubTopic --
    # its `topic_id` output is exactly this value. The project is read from
    # the name, so there is no separate project field, and a topic in another
    # project is granted the same way.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    topic = string

    # The role to grant on the topic. A predefined role
    # ("roles/pubsub.publisher", "roles/pubsub.subscriber",
    # "roles/pubsub.viewer", "roles/pubsub.editor", "roles/pubsub.admin") or a
    # custom role's full name ("projects/<project>/roles/<role_id>" or
    # "organizations/<org>/roles/<role_id>"). Reference a GcpIamCustomRole to
    # grant a custom role -- its `name` output is exactly this value.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    role = string

    # The identity receiving the grant, in IAM member format:
    #   serviceAccount:<email>  -- a service account or Google service agent.
    #                              Reference GcpServiceAccount (`member`), a
    #                              GcpLoggingSink (`writer_identity`, the
    #                              sink's writer, already in this format), or
    #                              a GcpSccNotificationConfig
    #                              (`service_account_member`, the identity
    #                              that publishes its notifications).
    #   user:<email>, group:<email>, domain:<domain>
    #   principal://... / principalSet://... -- workload identity federation
    #   allUsers / allAuthenticatedUsers     -- public grants; avoid on topics
    # Grants to deleted principals ("deleted:...") are not supported.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    member = string
  })
}
