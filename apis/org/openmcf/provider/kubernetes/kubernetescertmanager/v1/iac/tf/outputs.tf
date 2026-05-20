output "namespace" {
  description = "Kubernetes namespace where cert-manager was deployed"
  value       = local.namespace_name
}

output "release_name" {
  description = "Helm release name"
  value       = local.helm_chart_name
}

output "service_account_name" {
  description = "Name of the cert-manager controller ServiceAccount"
  value       = local.ksa_name
}
