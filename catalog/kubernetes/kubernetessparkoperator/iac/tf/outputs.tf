# Outputs — flattened onto KubernetesSparkOperatorOutputs by
# the platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace the operator was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name of the operator (metadata.name; the chart's fullname and every RBAC name are pinned to it)"
  value       = local.release_name
}

output "workload_service_account" {
  description = "Service account name Spark driver/executor pods run as in every workload namespace — SparkApplication declarations reference it"
  value       = local.workload_service_account
}

output "watched_namespaces" {
  description = "Namespaces the operator watches for Spark workloads (empty = cluster-wide)"
  value       = local.workload_namespaces
}
