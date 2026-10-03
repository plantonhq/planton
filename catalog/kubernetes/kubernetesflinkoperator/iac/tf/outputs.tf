# Outputs — flattened onto KubernetesFlinkOperatorOutputs by
# the platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace the operator was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name of the operator (metadata.name; the chart's fullname is pinned to it)"
  value       = local.release_name
}

output "job_service_account" {
  description = "Service account name Flink job pods run as — FlinkDeployment declarations reference it"
  value       = local.job_service_account
}

output "watched_namespaces" {
  description = "Namespaces the operator watches for Flink CRs (empty = cluster-wide)"
  value       = try(var.spec.watch_namespaces, [])
}

output "webhook_service" {
  description = "Name of the operator's webhook Service (chart-fixed \"flink-operator-webhook-service\"); empty when the webhook is disabled"
  value       = local.webhook_service
}
