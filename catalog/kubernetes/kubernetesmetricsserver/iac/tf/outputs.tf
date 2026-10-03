# Outputs — flattened onto KubernetesMetricsServerOutputs by the
# platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace metrics-server was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (fixed \"metrics-server\" — one installation per cluster)"
  value       = local.release_name
}

output "service_name" {
  description = "Name of the metrics-server Service the APIService routes to (the chart fullname, pinned to the release name)"
  value       = local.release_name
}

output "api_service_name" {
  description = "Name of the registered APIService (\"v1beta1.metrics.k8s.io\"); empty when spec.api_service.create is false"
  value       = local.api_service_name
}
