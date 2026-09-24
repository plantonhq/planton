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
  description = "GcpManagedKafkaAcl specification"
  type = object({
    # The GCP project the cluster lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The cluster's region, e.g. "us-central1". Immutable.
    location = string

    # The cluster the ACL applies to. A GcpManagedKafkaCluster reference
    # resolves to its full resource path (name output); a literal takes the
    # full path or the bare cluster ID. The modules derive the bare ID
    # Google's resource expects. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    cluster = string

    # The resource pattern, which is also the ACL's ID:
    #   cluster                          -- the cluster itself
    #   topic/{name}                     -- one topic
    #   consumerGroup/{name}             -- one consumer group
    #   transactionalId/{name}           -- one transactional ID
    #   topicPrefixed/{prefix}           -- every topic starting with prefix
    #   consumerGroupPrefixed/{prefix}   -- every consumer group with prefix
    #   transactionalIdPrefixed/{prefix} -- every transactional ID with prefix
    # {name} may be the literal wildcard "*" (every resource of the type).
    # One ACL exists per pattern on a cluster, so two manifests must never
    # declare the same one. Immutable.
    acl_id = string

    # The entries for this pattern -- at least one, at most 100.
    acl_entries = list(object({
      # Who the entry applies to, in Kafka's StandardAuthorizer form: "User:"
      # followed by a Google account, e.g.
      # "User:orders-api@my-project.iam.gserviceaccount.com", or "User:*" for
      # everyone. With mTLS, the principal is the certificate's mapped name
      # (see the cluster's ssl_principal_mapping_rules).
      principal = string

      # The Kafka operation:
      #   ALL, READ, WRITE, CREATE, DELETE, ALTER, DESCRIBE, CLUSTER_ACTION,
      #   DESCRIBE_CONFIGS, ALTER_CONFIGS, IDEMPOTENT_WRITE
      # Which operations are meaningful depends on the resource type (Kafka's
      # operations table): a producer needs WRITE (and DESCRIBE) on its topic;
      # a consumer needs READ on the topic and READ on its consumer group.
      # Google accepts any letter case; the spec takes the canonical upper
      # case.
      operation = string

      # ALLOW (Google's default when empty) or DENY. A DENY wins over any
      # ALLOW for the same principal and operation.
      permission_type = optional(string, "")

      # The client host the entry applies to. Managed Service for Apache Kafka
      # accepts only "*" (every host), which is also the default when empty.
      host = optional(string, "")
    }))

    # What happens to the ACL when this resource is destroyed:
    #   "" / "DELETE" -- deleted (the access it granted ends)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the ACL leaves management and stays on the cluster
    deletion_policy = optional(string, "")
  })
}
