# Outputs — flattened onto KubernetesKafkaTopicOutputs by the
# platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Namespace the KafkaTopic resource lives in (the Kafka cluster's namespace)"
  value       = local.namespace
}

output "topic_name" {
  description = "The actual Kafka topic name (spec.topic_name when set, otherwise metadata.name)"
  value       = local.topic_name
}
