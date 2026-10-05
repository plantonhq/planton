# Outputs — flattened onto KubernetesKafkaConnectOutputs by the
# platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace the Connect cluster deploys into"
  value       = local.namespace
}

output "connect_name" {
  description = "Connect cluster name (metadata.name) — the value KubernetesKafkaConnector resources bind to via their connect_cluster field (rendered as the strimzi.io/cluster label)"
  value       = local.connect_name
}

output "rest_api_service_name" {
  description = "Name of the Connect REST API Service (<name>-connect-api)"
  value       = local.rest_api_service_name
}

output "rest_api_endpoint" {
  description = "In-cluster Connect REST API endpoint (http://<name>-connect-api.<namespace>.svc.cluster.local:8083) — read-only inspection; connector management is declarative through KubernetesKafkaConnector"
  value       = local.rest_api_endpoint
}
