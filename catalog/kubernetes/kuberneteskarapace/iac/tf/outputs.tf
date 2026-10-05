# Outputs — must flatten onto KubernetesKarapaceOutputs
# (outputs.proto) identically to the Pulumi module's exports.

output "namespace" {
  description = "Namespace the registry runs in"
  value       = local.namespace
}

output "service_name" {
  description = "Name of the registry Service"
  value       = local.registry_name
}

output "endpoint" {
  description = "In-cluster registry endpoint (http(s)://<name>.<namespace>.svc.cluster.local:<port>) — the schema.registry.url value for clients"
  value       = local.endpoint
}

output "rest_proxy_endpoint" {
  description = "In-cluster REST-proxy endpoint (empty when the rest_proxy role is not enabled)"
  value       = local.rest_proxy_endpoint
}

output "schemas_topic" {
  description = "The Kafka topic storing the schemas"
  value       = local.topic_name
}
