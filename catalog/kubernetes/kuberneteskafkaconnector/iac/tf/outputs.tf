# Outputs — flattened onto KubernetesKafkaConnectorOutputs by
# the platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Namespace the KafkaConnector resource lives in (the Connect cluster's namespace)"
  value       = local.namespace
}

output "connector_name" {
  description = "The connector's name inside the Connect cluster (metadata.name)"
  value       = var.metadata.name
}
