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
  description = "GcpManagedKafkaConnector specification"
  type = object({
    # The GCP project the Connect cluster lives in: a literal project ID or
    # a GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Connect cluster's region, e.g. "us-central1". Immutable.
    location = string

    # The Connect cluster that runs the connector. A
    # GcpManagedKafkaConnectCluster reference resolves to its full resource
    # path (name output); a literal takes the full path or the bare Connect
    # cluster ID. The modules derive the bare ID Google's resource expects.
    # Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    connect_cluster = string

    # The connector's ID, the name Kafka Connect shows for it. Defaults to
    # metadata.name. Immutable.
    connector_id = optional(string, "")

    # The connector's Kafka Connect configuration, keyed by property:
    # "connector.class" picks the plugin (for example
    # com.google.pubsub.kafka.sink.CloudPubSubSinkConnector,
    # com.wepay.kafka.connect.bigquery.BigQuerySinkConnector,
    # io.aiven.kafka.connect.gcs.GcsSinkConnector,
    # org.apache.kafka.connect.mirror.MirrorSourceConnector), "tasks.max"
    # its parallelism, "topics" what a sink reads, plus the plugin's own
    # keys and converters. The values are stored in plain text on the
    # Connect cluster; keep credentials out of them where the plugin offers
    # Google IAM authentication instead.
    configs = optional(map(string), {})

    # Automatic restarts for failed tasks. Omit to leave failed tasks
    # stopped.
    task_restart_policy = optional(object({
      # The shortest wait before retrying a failed task -- a duration in
      # seconds with up to nine fractional digits and an "s" suffix, e.g.
      # "60s". Empty leaves Google's default.
      minimum_backoff = optional(string, "")

      # The longest wait between retries -- the backoff's upper bound, same
      # format, e.g. "1800s". Empty leaves Google's default.
      maximum_backoff = optional(string, "")
    }))

    # What happens to the connector when this resource is destroyed:
    #   "" / "DELETE" -- deleted (the pipeline stops)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- it leaves management and keeps running
    deletion_policy = optional(string, "")
  })
}
