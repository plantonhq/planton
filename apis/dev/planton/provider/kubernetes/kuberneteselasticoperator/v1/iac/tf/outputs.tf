output "namespace" {
  description = "The Kubernetes namespace where ECK operator is deployed"
  value       = local.namespace
}

output "helm_release_name" {
  description = "The name of the Helm release for ECK operator"
  value       = local.helm_release_name
}

output "operator_version" {
  description = "The version of the ECK operator deployed"
  value       = local.helm_chart_version
}

