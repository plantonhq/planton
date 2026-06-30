output "namespace" {
  description = "The Kubernetes namespace where the Percona Operator for MongoDB is installed"
  value       = var.spec.create_namespace ? kubernetes_namespace.percona_operator[0].metadata[0].name : local.namespace
}

output "operator_version" {
  description = "The version of the Percona Operator for MongoDB Helm chart deployed"
  value       = local.helm_chart_version
}

output "operator_name" {
  description = "The name of the Percona Operator for MongoDB Helm release"
  value       = helm_release.percona_operator.name
}

output "helm_status" {
  description = "The status of the Helm release deployment"
  value       = helm_release.percona_operator.status
}

