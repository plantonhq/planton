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
  description = "GcpManagedKafkaTopic specification"
  type = object({
    # The GCP project the cluster lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The cluster's region, e.g. "us-central1". Immutable.
    location = string

    # The cluster the topic lives on. A GcpManagedKafkaCluster reference
    # resolves to its full resource path (name output); a literal takes the
    # full path or the bare cluster ID. The modules derive the bare ID
    # Google's resource expects. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    cluster = string

    # The topic's name as producers and consumers see it -- Kafka's own
    # rule: letters, digits, dots, underscores, hyphens, at most 249
    # characters (avoid mixing dots and underscores; Kafka treats them as
    # colliding in metric names). Defaults to metadata.name. Immutable.
    topic_id = optional(string, "")

    # How many partitions the topic has -- the unit of consumer parallelism.
    # Can be raised in place, never lowered; raising it changes which
    # partition a keyed message lands on, so per-key ordering is only
    # guaranteed for messages written after the change. Empty leaves
    # Google's default.
    partition_count = optional(number, 0)

    # How many copies of each partition the cluster keeps. Google
    # recommends 3 for high availability (the brokers span three zones).
    # Immutable.
    replication_factor = optional(number, 0)

    # Topic-level overrides of the cluster's defaults, keyed by Kafka topic
    # property, e.g. "cleanup.policy" = "compact", "retention.ms" =
    # "604800000", "compression.type" = "producer". Keys absent here follow
    # the cluster default.
    configs = optional(map(string), {})

    # What happens to the topic when this resource is destroyed:
    #   "" / "DELETE" -- deleted, with its messages
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the topic leaves management and stays on the cluster
    deletion_policy = optional(string, "")
  })
}
