# Outputs — flattened onto KubernetesAltinityOperatorOutputs
# by the platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace the operator was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name of the operator (metadata.name)"
  value       = local.release_name
}

output "deployment_name" {
  description = "Name of the operator Deployment (= the chart fullname, pinned to metadata.name)"
  value       = local.deployment_name
}

output "credentials_secret_name" {
  description = "Name of the chart-managed Secret holding the operator's ClickHouse credentials (keys username/password)"
  value       = local.credentials_secret_name
}

output "metrics_endpoint" {
  description = "In-cluster Prometheus metrics endpoint for every managed cluster; empty when the metrics exporter is disabled"
  value       = local.metrics_endpoint
}
