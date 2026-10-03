# Outputs — flattened onto KubernetesKafkaUserOutputs by the
# platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Namespace the KafkaUser resource lives in (the Kafka cluster's namespace)"
  value       = local.namespace
}

output "username" {
  description = "Kafka principal name (metadata.name)"
  value       = local.username
}

output "secret_name" {
  description = "Name of the operator-generated credentials Secret (empty for tls-external users — no Secret is generated)"
  value       = local.secret_name
}
