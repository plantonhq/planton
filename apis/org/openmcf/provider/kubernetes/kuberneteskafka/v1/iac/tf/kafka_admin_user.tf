# Create an admin KafkaUser in the same namespace.
# The Strimzi operator will generate a Secret for this user,
# holding credentials for SCRAM-SHA-512 authentication.
# The user name is prefixed with the resource name to avoid conflicts
# when multiple Kafka instances share the same namespace.
resource "kubernetes_manifest" "kafka_admin_user" {
  manifest = {
    apiVersion = "kafka.strimzi.io/v1beta2"
    kind       = "KafkaUser"
    metadata = {
      name      = local.admin_username
      namespace = local.namespace
      # Merge our final_labels with the label needed to associate this user with the Kafka cluster.
      labels = merge(local.final_labels, {
        "strimzi.io/cluster" = local.kafka_cluster_name
      })
    }
    spec = {
      authentication = {
        type = "scram-sha-512"
      }
    }
  }

  depends_on = [
    kubernetes_manifest.kafka_cluster
  ]
}
