# We dynamically create a KafkaTopic resource for each topic specified in var.spec.kafka_topics
# using the Strimzi KafkaTopic CRD.
resource "kubernetes_manifest" "kafka_topic" {
  # Convert the list of KafkaTopic objects in var.spec.kafka_topics into a map keyed by topic name.
  # This way, each resource is uniquely identified by its name.
  for_each = {
    for topic in try(var.spec.kafka_topics, []) : topic.name => topic
  }

  manifest = {
    apiVersion = "kafka.strimzi.io/v1beta2"
    kind       = "KafkaTopic"
    metadata = {
      name      = each.value.name
      namespace = local.namespace
      labels    = local.final_labels
    }
    spec = {
      topicName  = each.value.name
      partitions = try(each.value.partitions, 1)
      replicas   = try(each.value.replicas, 1)

      # Merge default config with any user-provided config
      config = merge(
        {
          "cleanup.policy"                      = "delete"
          "delete.retention.ms"                 = "86400000"
          "max.message.bytes"                   = "2097164"
          "message.timestamp.difference.max.ms" = "9223372036854775807"
          "message.timestamp.type"              = "CreateTime"
          "min.insync.replicas"                 = "1"
          "retention.bytes"                     = "-1"
          "retention.ms"                        = "604800000"
          "segment.bytes"                       = "1073741824"
          "segment.ms"                          = "604800000"
        },
        each.value.config != null ? each.value.config : {}
      )
    }
  }

  depends_on = [
    kubernetes_manifest.kafka_cluster
  ]
}
